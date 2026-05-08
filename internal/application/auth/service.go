package auth

import (
	"context"
	"errors"
	"fmt"
	domain "go-booking-management-init/internal/domain/auth"
	pkgAuth "go-booking-management-init/pkg/auth"
	"log/slog"
	"net/mail"
	"strings"
)

var (
	ErrInvalidEmail    = errors.New("invalid email address")
	ErrInvalidRole     = errors.New("invalid role")
	ErrPasswordTooLong = errors.New("password too long")
)

type Service interface {
	Register(ctx context.Context, email, password, role string) (*domain.User, error)
}

type service struct {
	userRepo domain.UserRepository
}

func NewService(userRepo domain.UserRepository) Service {
	return &service{
		userRepo: userRepo,
	}
}

func (s *service) Register(ctx context.Context, email, password, role string) (*domain.User, error) {
	// 1. Validate Email
	if _, err := mail.ParseAddress(email); err != nil {
		slog.WarnContext(ctx, "invalid email address during registration", "email", email, "err", err)
		return nil, ErrInvalidEmail
	}

	// 2. Validate Role
	role = strings.ToLower(role)
	validRoles := map[string]bool{"customer": true, "admin": true, "guest": true}
	if !validRoles[role] {
		slog.WarnContext(ctx, "invalid role during registration", "role", role)
		return nil, ErrInvalidRole
	}

	// 3. Normalize Email
	email = strings.ToLower(strings.TrimSpace(email))

	// 4. Validate Password Length (Prevent DoS)
	if len(password) > 72 {
		slog.WarnContext(ctx, "password too long during registration")
		return nil, ErrPasswordTooLong
	}

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
