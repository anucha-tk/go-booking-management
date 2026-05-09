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

func (m *MockUserRepository) GetByID(ctx context.Context, id int32) (*domain.User, error) {
	args := m.Called(ctx, id)
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

func (m *MockTokenManager) GenerateRefreshToken(userID int32) (string, error) {
	args := m.Called(userID)
	return args.String(0), args.Error(1)
}

func (m *MockTokenManager) ValidateToken(tokenStr string) (*pkgAuth.UserClaims, error) {
	args := m.Called(tokenStr)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pkgAuth.UserClaims), args.Error(1)
}

func (m *MockTokenManager) ValidateRefreshToken(tokenStr string) (int32, error) {
	args := m.Called(tokenStr)
	return int32(args.Int(0)), args.Error(1)
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
		mockToken.On("GenerateToken", user.ID, user.Email, string(user.Role)).Return("access-token", nil).Once()
		mockToken.On("GenerateRefreshToken", user.ID).Return("refresh-token", nil).Once()

		accessToken, refreshToken, err := service.Login(ctx, email, password)

		assert.NoError(t, err)
		assert.Equal(t, "access-token", accessToken)
		assert.Equal(t, "refresh-token", refreshToken)
		mockRepo.AssertExpectations(t)
		mockToken.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		email := "nonexistent@example.com"
		mockRepo.On("GetByEmail", ctx, email).Return(nil, domain.ErrUserNotFound).Once()

		accessToken, refreshToken, err := service.Login(ctx, email, "any-password")

		assert.Error(t, err)
		assert.Equal(t, domain.ErrUserNotFound, err)
		assert.Empty(t, accessToken)
		assert.Empty(t, refreshToken)
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid password", func(t *testing.T) {
		email := "test@example.com"
		hashedPassword, _ := pkgAuth.HashPassword("correct-password")
		user := &domain.User{ID: 1, Email: email, PasswordHash: hashedPassword}

		mockRepo.On("GetByEmail", ctx, email).Return(user, nil).Once()

		accessToken, refreshToken, err := service.Login(ctx, email, "wrong-password")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid credentials")
		assert.Empty(t, accessToken)
		assert.Empty(t, refreshToken)
		mockRepo.AssertExpectations(t)
	})

	t.Run("token generation error", func(t *testing.T) {
		email := "test@example.com"
		password := "password123"
		hashedPassword, _ := pkgAuth.HashPassword(password)
		user := &domain.User{ID: 1, Email: email, PasswordHash: hashedPassword, Role: domain.RoleCustomer}

		mockRepo.On("GetByEmail", ctx, email).Return(user, nil).Once()
		mockToken.On("GenerateToken", user.ID, user.Email, string(user.Role)).Return("", errors.New("token error")).Once()

		accessToken, refreshToken, err := service.Login(ctx, email, password)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to generate access token")
		assert.Empty(t, accessToken)
		assert.Empty(t, refreshToken)
		mockRepo.AssertExpectations(t)
		mockToken.AssertExpectations(t)
	})
}

func TestService_RefreshToken(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockToken := new(MockTokenManager)
	service := NewService(mockRepo, mockToken)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		refreshToken := "valid-refresh-token"
		userID := int32(1)
		user := &domain.User{ID: userID, Email: "test@example.com", Role: domain.RoleCustomer}

		mockToken.On("ValidateRefreshToken", refreshToken).Return(int(userID), nil).Once()
		mockRepo.On("GetByID", ctx, userID).Return(user, nil).Once()
		mockToken.On("GenerateToken", user.ID, user.Email, string(user.Role)).Return("new-access-token", nil).Once()
		mockToken.On("GenerateRefreshToken", user.ID).Return("new-refresh-token", nil).Once()

		newAccessToken, newRefreshToken, err := service.RefreshToken(ctx, refreshToken)

		assert.NoError(t, err)
		assert.Equal(t, "new-access-token", newAccessToken)
		assert.Equal(t, "new-refresh-token", newRefreshToken)
		mockRepo.AssertExpectations(t)
		mockToken.AssertExpectations(t)
	})

	t.Run("invalid refresh token", func(t *testing.T) {
		refreshToken := "invalid-token"
		mockToken.On("ValidateRefreshToken", refreshToken).Return(0, errors.New("invalid")).Once()

		newAccessToken, newRefreshToken, err := service.RefreshToken(ctx, refreshToken)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid refresh token")
		assert.Empty(t, newAccessToken)
		assert.Empty(t, newRefreshToken)
		mockToken.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		refreshToken := "valid-token"
		userID := int32(1)
		mockToken.On("ValidateRefreshToken", refreshToken).Return(int(userID), nil).Once()
		mockRepo.On("GetByID", ctx, userID).Return(nil, domain.ErrUserNotFound).Once()

		newAccessToken, newRefreshToken, err := service.RefreshToken(ctx, refreshToken)

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidRefreshToken, err)

		assert.Empty(t, newAccessToken)
		assert.Empty(t, newRefreshToken)
		mockRepo.AssertExpectations(t)
		mockToken.AssertExpectations(t)
	})
}
