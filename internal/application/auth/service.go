package auth

import (
	"context"
	"errors"
	"fmt"
	domain "go-booking-management-init/internal/domain/auth"
	"go-booking-management-init/pkg/api"
	pkgAuth "go-booking-management-init/pkg/auth"
	"log/slog"
	"strings"

	"github.com/go-playground/validator/v10"
)

var (
	ErrInvalidEmail       = errors.New("invalid email address")
	ErrInvalidRole        = errors.New("invalid role")
	ErrPasswordTooShort   = errors.New("password is too short (min 8 characters)")
	ErrPasswordTooLong    = errors.New("password too long")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type Service interface {
	Register(ctx context.Context, email, password string, role domain.UserRole) (*domain.User, error)
	Login(ctx context.Context, email, password string) (string, error)
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

func (s *service) Login(ctx context.Context, email, password string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return "", domain.ErrUserNotFound
		}
		slog.ErrorContext(ctx, "failed to get user by email", "email", email, "err", err)
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	if err := pkgAuth.ComparePassword(user.PasswordHash, password); err != nil {
		slog.WarnContext(ctx, "invalid password login attempt", "email", email)
		return "", ErrInvalidCredentials
	}

	token, err := s.tokenManager.GenerateToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		slog.ErrorContext(ctx, "failed to generate token", "user_id", user.ID, "err", err)
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	slog.InfoContext(ctx, "user logged in successfully", "user_id", user.ID)
	return token, nil
}
