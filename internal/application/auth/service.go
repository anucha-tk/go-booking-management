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
)

var (
	ErrInvalidEmail    = errors.New("invalid email address")
	ErrInvalidRole     = errors.New("invalid role")
	ErrPasswordTooLong = errors.New("password too long")
)

type Service interface {
	Register(ctx context.Context, email, password string, role domain.UserRole) (*domain.User, error)
}

type service struct {
	userRepo domain.UserRepository
}

func NewService(userRepo domain.UserRepository) Service {
	return &service{
		userRepo: userRepo,
	}
}

func (s *service) Register(ctx context.Context, email, password string, role domain.UserRole) (*domain.User, error) {
	// 1. Validate Input using Service Layer Validator (validate tag)
	type registerInput struct {
		Email    string `validate:"required,email"`
		Password string `validate:"required,max=72"`
		Role     string `validate:"required,oneof=customer admin guest"`
	}

	input := registerInput{
		Email:    strings.ToLower(strings.TrimSpace(email)),
		Password: password,
		Role:     string(role),
	}

	if err := api.Validate(input); err != nil {
		slog.WarnContext(ctx, "validation failed in service layer", "err", err)
		// Map back to specific domain errors if needed, or return the validation error
		// For now, let's keep the legacy mapping but make it slightly better
		errMsg := err.Error()
		if strings.Contains(errMsg, "'Email'") {
			return nil, ErrInvalidEmail
		}
		if strings.Contains(errMsg, "'Password'") {
			return nil, ErrPasswordTooLong
		}
		return nil, ErrInvalidRole
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
