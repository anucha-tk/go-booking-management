package booking

import (
	"context"
	"go-booking-management-init/internal/domain/booking"
	"go-booking-management-init/internal/domain/room"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockBookingRepo struct {
	mock.Mock
}

func (m *mockBookingRepo) Create(ctx context.Context, b *booking.Booking) (*booking.Booking, error) {
	args := m.Called(ctx, b)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*booking.Booking), args.Error(1)
}

func (m *mockBookingRepo) GetByID(ctx context.Context, id int32) (*booking.Booking, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*booking.Booking), args.Error(1)
}

func (m *mockBookingRepo) ListByUser(ctx context.Context, userID int32) ([]*booking.UserBookingInfo, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*booking.UserBookingInfo), args.Error(1)
}

func (m *mockBookingRepo) ListAll(ctx context.Context) ([]*booking.AdminBookingInfo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*booking.AdminBookingInfo), args.Error(1)
}

func (m *mockBookingRepo) UpdateStatus(ctx context.Context, id int32, status string) (*booking.Booking, error) {
	args := m.Called(ctx, id, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*booking.Booking), args.Error(1)
}

type mockRoomRepo struct {
	mock.Mock
}

func (m *mockRoomRepo) Create(ctx context.Context, rm *room.Room) (*room.Room, error) {
	args := m.Called(ctx, rm)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*room.Room), args.Error(1)
}

func (m *mockRoomRepo) GetByID(ctx context.Context, id int32) (*room.Room, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*room.Room), args.Error(1)
}

func (m *mockRoomRepo) GetByNumber(ctx context.Context, roomNumber string) (*room.Room, error) {
	args := m.Called(ctx, roomNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*room.Room), args.Error(1)
}

func (m *mockRoomRepo) List(ctx context.Context, filter room.Filter) ([]*room.Room, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*room.Room), args.Error(1)
}

func (m *mockRoomRepo) ListAvailable(ctx context.Context, filter room.AvailabilityFilter) ([]*room.Room, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*room.Room), args.Error(1)
}

func (m *mockRoomRepo) GetDetail(ctx context.Context, id int32) (*room.Detail, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*room.Detail), args.Error(1)
}

func (m *mockRoomRepo) Update(ctx context.Context, rm *room.Room) (*room.Room, error) {
	args := m.Called(ctx, rm)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*room.Room), args.Error(1)
}

func (m *mockRoomRepo) Delete(ctx context.Context, id int32) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockRoomRepo) UpdateStatus(ctx context.Context, id int32, status string) (*room.Room, error) {
	args := m.Called(ctx, id, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*room.Room), args.Error(1)
}

func TestService_SubmitBooking(t *testing.T) {
	bRepo := new(mockBookingRepo)
	rRepo := new(mockRoomRepo)
	svc := NewService(bRepo, rRepo)
	ctx := context.Background()

	startDate := time.Now().Add(24 * time.Hour).Truncate(24 * time.Hour)
	endDate := startDate.Add(48 * time.Hour)

	t.Run("success", func(t *testing.T) {
		params := booking.CreateParams{
			UserID:    1,
			RoomID:    101,
			StartDate: startDate,
			EndDate:   endDate,
		}

		rRepo.On("GetByID", ctx, int32(101)).Return(&room.Room{ID: 101, Price: 1000}, nil).Once()
		rRepo.On("ListAvailable", ctx, mock.Anything).Return([]*room.Room{{ID: 101}}, nil).Once()

		bRepo.On("Create", ctx, mock.MatchedBy(func(b *booking.Booking) bool {
			return b.RoomID == 101 && b.TotalPrice == 2000 && b.Status == booking.StatusConfirmed
		})).Return(&booking.Booking{ID: 1, RoomID: 101}, nil).Once()

		res, err := svc.SubmitBooking(ctx, params)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, int32(1), res.ID)
		bRepo.AssertExpectations(t)
		rRepo.AssertExpectations(t)
	})

	t.Run("invalid dates - past", func(t *testing.T) {
		params := booking.CreateParams{
			UserID:    1,
			RoomID:    101,
			StartDate: time.Now().Add(-24 * time.Hour),
			EndDate:   time.Now(),
		}

		res, err := svc.SubmitBooking(ctx, params)

		assert.ErrorIs(t, err, booking.ErrInvalidBookingDates)
		assert.Nil(t, res)
	})

	t.Run("room not available", func(t *testing.T) {
		params := booking.CreateParams{
			UserID:    1,
			RoomID:    101,
			StartDate: startDate,
			EndDate:   endDate,
		}

		rRepo.On("GetByID", ctx, int32(101)).Return(&room.Room{ID: 101, Price: 1000}, nil).Once()
		rRepo.On("ListAvailable", ctx, mock.Anything).Return([]*room.Room{}, nil).Once()

		res, err := svc.SubmitBooking(ctx, params)

		assert.ErrorIs(t, err, booking.ErrRoomNotAvailable)
		assert.Nil(t, res)
	})
}

func TestService_GetMyBookings(t *testing.T) {
	bRepo := new(mockBookingRepo)
	svc := NewService(bRepo, nil)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		expected := []*booking.UserBookingInfo{{Booking: booking.Booking{ID: 1}, RoomNumber: "101"}}
		bRepo.On("ListByUser", ctx, int32(1)).Return(expected, nil).Once()

		res, err := svc.GetMyBookings(ctx, 1)

		assert.NoError(t, err)
		assert.Len(t, res, 1)
		bRepo.AssertExpectations(t)
	})
}

func TestService_GetAllBookings(t *testing.T) {
	bRepo := new(mockBookingRepo)
	svc := NewService(bRepo, nil)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		expected := []*booking.AdminBookingInfo{{UserBookingInfo: booking.UserBookingInfo{Booking: booking.Booking{ID: 1}}, UserEmail: "test@test.com"}}
		bRepo.On("ListAll", ctx).Return(expected, nil).Once()

		res, err := svc.GetAllBookings(ctx)

		assert.NoError(t, err)
		assert.Len(t, res, 1)
		bRepo.AssertExpectations(t)
	})
}

func TestService_CancelBooking(t *testing.T) {
	bRepo := new(mockBookingRepo)
	svc := NewService(bRepo, nil)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		bRepo.On("GetByID", ctx, int32(1)).Return(&booking.Booking{ID: 1, UserID: 1, Status: booking.StatusConfirmed}, nil).Once()
		bRepo.On("UpdateStatus", ctx, int32(1), booking.StatusCancelled).Return(&booking.Booking{ID: 1, Status: booking.StatusCancelled}, nil).Once()

		res, err := svc.CancelBooking(ctx, 1, false, 1)

		assert.NoError(t, err)
		assert.Equal(t, booking.StatusCancelled, res.Status)
		bRepo.AssertExpectations(t)
	})

	t.Run("success admin override", func(t *testing.T) {
		bRepo.On("GetByID", ctx, int32(1)).Return(&booking.Booking{ID: 1, UserID: 2, Status: booking.StatusConfirmed}, nil).Once()
		bRepo.On("UpdateStatus", ctx, int32(1), booking.StatusCancelled).Return(&booking.Booking{ID: 1, Status: booking.StatusCancelled}, nil).Once()

		res, err := svc.CancelBooking(ctx, 1, true, 1)

		assert.NoError(t, err)
		assert.Equal(t, booking.StatusCancelled, res.Status)
		bRepo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		bRepo.On("GetByID", ctx, int32(1)).Return((*booking.Booking)(nil), nil).Once()

		res, err := svc.CancelBooking(ctx, 99, false, 1)

		assert.ErrorIs(t, err, booking.ErrBookingNotFound)
		assert.Nil(t, res)
	})

	t.Run("unauthorized", func(t *testing.T) {
		bRepo.On("GetByID", ctx, int32(1)).Return(&booking.Booking{ID: 1, UserID: 2}, nil).Once()

		res, err := svc.CancelBooking(ctx, 1, false, 1)

		assert.ErrorIs(t, err, booking.ErrUnauthorized)
		assert.Nil(t, res)
	})
}
