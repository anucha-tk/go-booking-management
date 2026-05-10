package db

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	domain "go-booking-management-init/internal/domain/auth"

	"github.com/google/uuid"
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

func (m *mockQuerier) GetUserByID(ctx context.Context, id int32) (User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(User), args.Error(1)
}

func (m *mockQuerier) RevokeToken(ctx context.Context, arg RevokeTokenParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *mockQuerier) IsTokenRevoked(ctx context.Context, jti uuid.UUID) (bool, error) {
	args := m.Called(ctx, jti)
	return args.Bool(0), args.Error(1)
}

func (m *mockQuerier) CleanupExpiredTokens(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockQuerier) CreateRoom(ctx context.Context, arg CreateRoomParams) (Room, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(Room), args.Error(1)
}

func (m *mockQuerier) DeleteRoom(ctx context.Context, id int32) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockQuerier) GetRoom(ctx context.Context, id int32) (Room, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(Room), args.Error(1)
}

func (m *mockQuerier) GetRoomByNumber(ctx context.Context, roomNumber string) (Room, error) {
	args := m.Called(ctx, roomNumber)
	return args.Get(0).(Room), args.Error(1)
}

func (m *mockQuerier) ListAvailableRooms(ctx context.Context, arg ListAvailableRoomsParams) ([]Room, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Room), args.Error(1)
}

func (m *mockQuerier) ListRooms(ctx context.Context, arg ListRoomsParams) ([]Room, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Room), args.Error(1)
}

func (m *mockQuerier) UpdateRoom(ctx context.Context, arg UpdateRoomParams) (Room, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(Room), args.Error(1)
}

func (m *mockQuerier) UpdateRoomStatus(ctx context.Context, arg UpdateRoomStatusParams) (Room, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(Room), args.Error(1)
}

func (m *mockQuerier) ListBookingsByRoom(ctx context.Context, arg ListBookingsByRoomParams) ([]ListBookingsByRoomRow, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]ListBookingsByRoomRow), args.Error(1)
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
			CreatedAt:    now,
			UpdatedAt:    now,
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
			CreatedAt:    now,
			UpdatedAt:    now,
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

func TestSQLCAuthRepository_GetByID(t *testing.T) {
	mq := new(mockQuerier)
	repo := &SQLCAuthRepository{
		queries: mq,
	}
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		id := int32(1)
		now := time.Now()
		mq.On("GetUserByID", ctx, id).Return(User{
			ID:           id,
			Email:        "test@test.com",
			PasswordHash: "hash",
			Role:         "customer",
			CreatedAt:    now,
			UpdatedAt:    now,
		}, nil).Once()

		res, err := repo.GetByID(ctx, id)

		assert.NoError(t, err)
		assert.Equal(t, id, res.ID)
		mq.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mq.On("GetUserByID", ctx, int32(999)).Return(User{}, sql.ErrNoRows).Once()

		res, err := repo.GetByID(ctx, int32(999))

		assert.Error(t, err)
		assert.Equal(t, domain.ErrUserNotFound, err)
		assert.Nil(t, res)
	})
}

func TestSQLCAuthRepository_RevokeToken(t *testing.T) {
	mq := new(mockQuerier)
	repo := &SQLCAuthRepository{
		queries: mq,
	}
	ctx := context.Background()
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		jti := uuid.New().String()
		mq.On("RevokeToken", ctx, mock.Anything).Return(nil).Once()

		err := repo.RevokeToken(ctx, jti, now)

		assert.NoError(t, err)
		mq.AssertExpectations(t)
	})

	t.Run("invalid jti format", func(t *testing.T) {
		err := repo.RevokeToken(ctx, "not-a-uuid", now)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid jti format")
	})
}

func TestSQLCAuthRepository_IsTokenRevoked(t *testing.T) {
	mq := new(mockQuerier)
	repo := &SQLCAuthRepository{
		queries: mq,
	}
	ctx := context.Background()

	t.Run("token revoked", func(t *testing.T) {
		jti := uuid.New().String()
		parsedUUID := uuid.MustParse(jti)
		mq.On("IsTokenRevoked", ctx, parsedUUID).Return(true, nil).Once()

		revoked, err := repo.IsTokenRevoked(ctx, jti)

		assert.NoError(t, err)
		assert.True(t, revoked)
		mq.AssertExpectations(t)
	})

	t.Run("token not revoked", func(t *testing.T) {
		jti := uuid.New().String()
		parsedUUID := uuid.MustParse(jti)
		mq.On("IsTokenRevoked", ctx, parsedUUID).Return(false, nil).Once()

		revoked, err := repo.IsTokenRevoked(ctx, jti)

		assert.NoError(t, err)
		assert.False(t, revoked)
		mq.AssertExpectations(t)
	})

	t.Run("invalid jti format", func(t *testing.T) {
		revoked, err := repo.IsTokenRevoked(ctx, "not-a-uuid")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid jti format")
		assert.False(t, revoked)
	})
}

func TestSQLCAuthRepository_GetByID_GenericError(t *testing.T) {
	mq := new(mockQuerier)
	repo := &SQLCAuthRepository{
		queries: mq,
	}
	ctx := context.Background()

	mq.On("GetUserByID", ctx, int32(1)).Return(User{}, errors.New("db connection error")).Once()

	res, err := repo.GetByID(ctx, int32(1))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error in GetByID")
	assert.Nil(t, res)
	mq.AssertExpectations(t)
}
