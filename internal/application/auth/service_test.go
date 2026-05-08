package auth

import (
	"context"
	domain "go-booking-management-init/internal/domain/auth"
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

func TestService_Register(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewService(mockRepo)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		email := "test@example.com"
		password := "password123"
		role := "guest"

		mockRepo.On("GetByEmail", ctx, email).Return(nil, domain.ErrUserNotFound).Once()
		mockRepo.On("Create", ctx, mock.Anything).Return(&domain.User{
			ID:    1,
			Email: email,
			Role:  role,
		}, nil).Once()

		user, err := service.Register(ctx, email, password, role)

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

		user, err := service.Register(ctx, email, password, role)

		assert.Error(t, err)
		assert.Equal(t, domain.ErrUserAlreadyExists, err)
		assert.Nil(t, user)
		mockRepo.AssertExpectations(t)
	})
	t.Run("invalid email", func(t *testing.T) {
		user, err := service.Register(ctx, "invalid-email", "password123", "guest")
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidEmail, err)
		assert.Nil(t, user)
	})

	t.Run("invalid role", func(t *testing.T) {
		user, err := service.Register(ctx, "test@example.com", "password123", "invalid-role")
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidRole, err)
		assert.Nil(t, user)
	})
}
