package db

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	domain "go-booking-management-init/internal/domain/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockQuerier struct {
	mock.Mock
}

func (m *mockQuerier) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(User), args.Error(1)
}

func (m *mockQuerier) GetUserByEmail(ctx context.Context, email string) (User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(User), args.Error(1)
}

func TestSQLCAuthRepository_Create(t *testing.T) {
	mq := new(mockQuerier)
	repo := &SQLCAuthRepository{
		queries: mq,
	}
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		user := &domain.User{
			Email:        "test@example.com",
			PasswordHash: "hash",
			Role:         domain.RoleCustomer,
		}

		now := time.Now()
		mq.On("CreateUser", ctx, CreateUserParams{
			Email:        user.Email,
			PasswordHash: user.PasswordHash,
			Role:         string(user.Role),
		}).Return(User{
			ID:           1,
			Email:        user.Email,
			PasswordHash: user.PasswordHash,
			Role:         string(user.Role),
			CreatedAt:    sql.NullTime{Time: now, Valid: true},
			UpdatedAt:    sql.NullTime{Time: now, Valid: true},
		}, nil).Once()

		res, err := repo.Create(ctx, user)

		assert.NoError(t, err)
		assert.Equal(t, int32(1), res.ID)
		assert.Equal(t, user.Email, res.Email)
		mq.AssertExpectations(t)
	})

	t.Run("duplicate email", func(t *testing.T) {
		user := &domain.User{Email: "exists@test.com"}
		mq.On("CreateUser", ctx, mock.Anything).Return(User{}, errors.New("duplicate key value violates unique constraint")).Once()

		res, err := repo.Create(ctx, user)

		assert.Error(t, err)
		assert.Equal(t, domain.ErrUserAlreadyExists, err)
		assert.Nil(t, res)
	})

	t.Run("generic error", func(t *testing.T) {
		user := &domain.User{Email: "error@test.com"}
		mq.On("CreateUser", ctx, mock.Anything).Return(User{}, errors.New("db error")).Once()

		res, err := repo.Create(ctx, user)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
		assert.Nil(t, res)
	})
}

func TestSQLCAuthRepository_GetByEmail(t *testing.T) {
	mq := new(mockQuerier)
	repo := &SQLCAuthRepository{
		queries: mq,
	}
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		email := "test@example.com"
		now := time.Now()
		mq.On("GetUserByEmail", ctx, email).Return(User{
			ID:           1,
			Email:        email,
			PasswordHash: "hash",
			Role:         "customer",
			CreatedAt:    sql.NullTime{Time: now, Valid: true},
			UpdatedAt:    sql.NullTime{Time: now, Valid: true},
		}, nil).Once()

		res, err := repo.GetByEmail(ctx, email)

		assert.NoError(t, err)
		assert.Equal(t, email, res.Email)
		mq.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mq.On("GetUserByEmail", ctx, "notfound@test.com").Return(User{}, sql.ErrNoRows).Once()

		res, err := repo.GetByEmail(ctx, "notfound@test.com")

		assert.Error(t, err)
		assert.Equal(t, domain.ErrUserNotFound, err)
		assert.Nil(t, res)
	})

	t.Run("generic error", func(t *testing.T) {
		mq.On("GetUserByEmail", ctx, "error@test.com").Return(User{}, errors.New("db error")).Once()

		res, err := repo.GetByEmail(ctx, "error@test.com")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
		assert.Nil(t, res)
	})
}
