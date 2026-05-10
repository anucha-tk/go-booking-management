package auth

import (
	"context"
	"errors"
	"fmt"
	domain "go-booking-management-init/internal/domain/auth"
	"go-booking-management-init/pkg/api"
	pkgAuth "go-booking-management-init/pkg/auth"
	"log/slog"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
)

var (
	ErrInvalidEmail        = errors.New("invalid email address")
	ErrInvalidRole         = errors.New("invalid role")
	ErrPasswordTooShort    = errors.New("password is too short (min 8 characters)")
	ErrPasswordTooLong     = errors.New("password too long")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
)

type Service interface {
	Register(ctx context.Context, email, password string, role domain.UserRole) (*domain.User, error)
	Login(ctx context.Context, email, password string) (string, string, error)
	RefreshToken(ctx context.Context, refreshToken string) (string, string, error)
	Logout(ctx context.Context, accessToken string) error
	IsTokenRevoked(ctx context.Context, accessToken string) (bool, error)
}

type service struct {
	userRepo     domain.UserRepository
	tokenManager pkgAuth.TokenManager
}

func NewService(userRepo domain.UserRepository, tokenManager pkgAuth.TokenManager) Service {
	return &service{
		userRepo:     userRepo,
		tokenManager: tokenManager,
	}
}

func (s *service) Register(ctx context.Context, email, password string, role domain.UserRole) (*domain.User, error) {
	// 1. Validate Input using Service Layer Validator (validate tag)
	type registerInput struct {
		Email    string `validate:"required,email"`
		Password string `validate:"required,min=8,max=72"`
		Role     string `validate:"required,oneof=customer admin guest"`
	}

	input := registerInput{
		Email:    strings.ToLower(strings.TrimSpace(email)),
		Password: password,
		Role:     string(role),
	}

	if err := api.Validate(input); err != nil {
		slog.WarnContext(ctx, "validation failed in service layer", "err", err)
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			for _, fe := range ve {
				switch fe.Field() {
				case "Email":
					return nil, ErrInvalidEmail
				case "Password":
					if fe.Tag() == "min" {
						return nil, ErrPasswordTooShort
					}
					return nil, ErrPasswordTooLong
				case "Role":
					return nil, ErrInvalidRole
				}
			}
		}
		return nil, err
	}

	// 2. Use normalized values from input
	email = input.Email
	role = domain.UserRole(input.Role)

	// 5. Check if user already exists
	existing, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		slog.ErrorContext(ctx, "failed to check if user exists", "email", email, "err", err)
		return nil, fmt.Errorf("failed to check if user exists: %w", err)
	}
	if existing != nil {
		return nil, domain.ErrUserAlreadyExists
	}

	// 4. Hash password
	hashedPassword, err := pkgAuth.HashPassword(password)
	if err != nil {
		slog.ErrorContext(ctx, "failed to hash password", "err", err)
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &domain.User{
		Email:        email,
		PasswordHash: hashedPassword,
		Role:         role,
	}

	createdUser, err := s.userRepo.Create(ctx, user)
	if err != nil {
		slog.ErrorContext(ctx, "failed to create user", "email", email, "err", err)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	slog.InfoContext(ctx, "user registered successfully", "user_id", createdUser.ID, "email", createdUser.Email)
	return createdUser, nil
}

func (s *service) Login(ctx context.Context, email, password string) (string, string, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return "", "", domain.ErrUserNotFound
		}
		slog.ErrorContext(ctx, "failed to get user by email", "email", email, "err", err)
		return "", "", fmt.Errorf("failed to get user: %w", err)
	}

	if err := pkgAuth.ComparePassword(user.PasswordHash, password); err != nil {
		slog.WarnContext(ctx, "invalid password login attempt", "email", email)
		return "", "", ErrInvalidCredentials
	}

	accessToken, err := s.tokenManager.GenerateToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		slog.ErrorContext(ctx, "failed to generate access token", "user_id", user.ID, "err", err)
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.tokenManager.GenerateRefreshToken(user.ID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to generate refresh token", "user_id", user.ID, "err", err)
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	slog.InfoContext(ctx, "user logged in successfully", "user_id", user.ID)
	return accessToken, refreshToken, nil
}

func (s *service) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	claims, err := s.tokenManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		slog.WarnContext(ctx, "invalid refresh token attempt", "err", err)
		return "", "", ErrInvalidRefreshToken
	}

	// Check if refresh token is revoked
	revoked, err := s.userRepo.IsTokenRevoked(ctx, claims.ID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check if refresh token is revoked", "jti", claims.ID, "err", err)
		return "", "", fmt.Errorf("failed to check revocation: %w", err)
	}
	if revoked {
		slog.WarnContext(ctx, "attempted use of revoked refresh token", "jti", claims.ID)
		return "", "", ErrInvalidRefreshToken
	}

	userID64, err := strconv.ParseInt(claims.Subject, 10, 32)
	if err != nil {
		return "", "", ErrInvalidRefreshToken
	}
	userID := int32(userID64)

	// Fetch user to get current email and role
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			slog.WarnContext(ctx, "refresh token user not found", "user_id", userID)
			return "", "", ErrInvalidRefreshToken
		}
		slog.ErrorContext(ctx, "failed to get user for token refresh", "user_id", userID, "err", err)
		return "", "", fmt.Errorf("failed to get user: %w", err)
	}

	newAccessToken, err := s.tokenManager.GenerateToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		slog.ErrorContext(ctx, "failed to generate new access token", "user_id", user.ID, "err", err)
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := s.tokenManager.GenerateRefreshToken(user.ID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to generate new refresh token", "user_id", user.ID, "err", err)
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Revoke the old refresh token to prevent reuse (rolling refresh strategy)
	if err := s.userRepo.RevokeToken(ctx, claims.ID, claims.ExpiresAt.Time); err != nil {
		slog.ErrorContext(ctx, "failed to revoke old refresh token during refresh", "jti", claims.ID, "err", err)
		// We don't return error here because the new tokens are already generated and valid
	}

	slog.InfoContext(ctx, "token refreshed successfully", "user_id", user.ID, "old_jti", claims.ID)
	return newAccessToken, newRefreshToken, nil
}

func (s *service) Logout(ctx context.Context, accessToken string) error {
	claims, err := s.tokenManager.ValidateToken(accessToken)
	if err != nil {
		slog.WarnContext(ctx, "failed to validate token for logout", "err", err)
		return err
	}

	if claims.ID == "" {
		return errors.New("token missing jti")
	}

	err = s.userRepo.RevokeToken(ctx, claims.ID, claims.ExpiresAt.Time)
	if err != nil {
		slog.ErrorContext(ctx, "failed to revoke token", "jti", claims.ID, "err", err)
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	slog.InfoContext(ctx, "token revoked successfully", "jti", claims.ID, "user_id", claims.UserID)
	return nil
}

func (s *service) IsTokenRevoked(ctx context.Context, accessToken string) (bool, error) {
	claims, err := s.tokenManager.ValidateToken(accessToken)
	if err != nil {
		return false, err
	}

	if claims.ID == "" {
		return false, errors.New("token missing jti")
	}

	return s.userRepo.IsTokenRevoked(ctx, claims.ID)
}
