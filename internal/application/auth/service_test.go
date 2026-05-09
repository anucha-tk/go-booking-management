package auth

import (
	"context"
	"errors"
	domain "go-booking-management-init/internal/domain/auth"
	pkgAuth "go-booking-management-init/pkg/auth"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

type MockTokenManager struct {
	mock.Mock
}

func (m *MockTokenManager) GenerateToken(userID int32, email string, role string) (string, error) {
	args := m.Called(userID, email, role)
	return args.String(0), args.Error(1)
}

func (m *MockTokenManager) ValidateToken(tokenStr string) (*pkgAuth.UserClaims, error) {
	args := m.Called(tokenStr)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pkgAuth.UserClaims), args.Error(1)
}

func TestService_Register(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockToken := new(MockTokenManager)
	service := NewService(mockRepo, mockToken)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		email := "test@example.com"
		password := "password123"
		role := "guest"

		mockRepo.On("GetByEmail", ctx, email).Return(nil, domain.ErrUserNotFound).Once()
		mockRepo.On("Create", ctx, mock.Anything).Return(&domain.User{
			ID:    1,
			Email: email,
			Role:  domain.UserRole(role),
		}, nil).Once()

		user, err := service.Register(ctx, email, password, domain.UserRole(role))

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, email, user.Email)
		mockRepo.AssertExpectations(t)
	})

	t.Run("user already exists", func(t *testing.T) {
		email := "existing@example.com"
		password := "password123"
		role := "guest"

		mockRepo.On("GetByEmail", ctx, email).Return(&domain.User{ID: 1, Email: email}, nil).Once()

		user, err := service.Register(ctx, email, password, domain.UserRole(role))

		assert.Error(t, err)
		assert.Equal(t, domain.ErrUserAlreadyExists, err)
		assert.Nil(t, user)
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid email", func(t *testing.T) {
		user, err := service.Register(ctx, "invalid-email", "password123", domain.RoleGuest)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidEmail, err)
		assert.Nil(t, user)
	})

	t.Run("invalid role", func(t *testing.T) {
		user, err := service.Register(ctx, "test@example.com", "password123", domain.UserRole("invalid-role"))
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidRole, err)
		assert.Nil(t, user)
	})

	t.Run("password too long", func(t *testing.T) {
		longPass := string(make([]byte, 73))
		user, err := service.Register(ctx, "test@example.com", longPass, domain.RoleGuest)
		assert.Error(t, err)
		assert.Equal(t, ErrPasswordTooLong, err)
		assert.Nil(t, user)
	})

	t.Run("password too short", func(t *testing.T) {
		shortPass := "short"
		user, err := service.Register(ctx, "test@example.com", shortPass, domain.RoleGuest)
		assert.Error(t, err)
		assert.Equal(t, ErrPasswordTooShort, err)
		assert.Nil(t, user)
	})

	t.Run("get by email returns unexpected error", func(t *testing.T) {
		email := "dberror@example.com"
		password := "password123"

		mockRepo.On("GetByEmail", ctx, email).Return(nil, errors.New("connection refused")).Once()

		user, err := service.Register(ctx, email, password, domain.RoleGuest)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to check if user exists")
		assert.Nil(t, user)
		mockRepo.AssertExpectations(t)
	})

	t.Run("create user fails", func(t *testing.T) {
		email := "createfail@example.com"
		password := "password123"

		mockRepo.On("GetByEmail", ctx, email).Return(nil, domain.ErrUserNotFound).Once()
		mockRepo.On("Create", ctx, mock.Anything).Return(nil, errors.New("insert failed")).Once()

		user, err := service.Register(ctx, email, password, domain.RoleGuest)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create user")
		assert.Nil(t, user)
		mockRepo.AssertExpectations(t)
	})
}

func TestService_Login(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockToken := new(MockTokenManager)
	service := NewService(mockRepo, mockToken)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		email := "test@example.com"
		password := "password123"
		hashedPassword, _ := pkgAuth.HashPassword(password)
		user := &domain.User{ID: 1, Email: email, PasswordHash: hashedPassword, Role: domain.RoleCustomer}

		mockRepo.On("GetByEmail", ctx, email).Return(user, nil).Once()
		mockToken.On("GenerateToken", user.ID, user.Email, string(user.Role)).Return("valid-token", nil).Once()

		token, err := service.Login(ctx, email, password)

		assert.NoError(t, err)
		assert.Equal(t, "valid-token", token)
		mockRepo.AssertExpectations(t)
		mockToken.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		email := "nonexistent@example.com"
		mockRepo.On("GetByEmail", ctx, email).Return(nil, domain.ErrUserNotFound).Once()

		token, err := service.Login(ctx, email, "any-password")

		assert.Error(t, err)
		assert.Equal(t, domain.ErrUserNotFound, err)
		assert.Empty(t, token)
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid password", func(t *testing.T) {
		email := "test@example.com"
		hashedPassword, _ := pkgAuth.HashPassword("correct-password")
		user := &domain.User{ID: 1, Email: email, PasswordHash: hashedPassword}

		mockRepo.On("GetByEmail", ctx, email).Return(user, nil).Once()

		token, err := service.Login(ctx, email, "wrong-password")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid credentials")
		assert.Empty(t, token)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		email := "dberror@example.com"
		mockRepo.On("GetByEmail", ctx, email).Return(nil, errors.New("db error")).Once()

		token, err := service.Login(ctx, email, "password123")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get user")
		assert.Empty(t, token)
		mockRepo.AssertExpectations(t)
	})

	t.Run("token generation error", func(t *testing.T) {
		email := "test@example.com"
		password := "password123"
		hashedPassword, _ := pkgAuth.HashPassword(password)
		user := &domain.User{ID: 1, Email: email, PasswordHash: hashedPassword, Role: domain.RoleCustomer}

		mockRepo.On("GetByEmail", ctx, email).Return(user, nil).Once()
		mockToken.On("GenerateToken", user.ID, user.Email, string(user.Role)).Return("", errors.New("token error")).Once()

		token, err := service.Login(ctx, email, password)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to generate token")
		assert.Empty(t, token)
		mockRepo.AssertExpectations(t)
		mockToken.AssertExpectations(t)
	})
}
