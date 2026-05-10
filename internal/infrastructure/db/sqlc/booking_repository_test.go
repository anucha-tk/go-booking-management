package db

import (
	"context"
	"database/sql"
	"go-booking-management-init/internal/domain/booking"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSQLCBookingRepository_Create(t *testing.T) {
	mq := new(mockQuerier)
	repo := &SQLCBookingRepository{
		queries: mq,
	}
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		now := time.Now()
		b := &booking.Booking{
			UserID:     1,
			RoomID:     101,
			StartDate:  now.Add(24 * time.Hour),
			EndDate:    now.Add(48 * time.Hour),
			TotalPrice: 1000,
			Status:     "confirmed",
		}

		mq.On("CreateBookingSafe", ctx, mock.Anything).Return(Booking{
			ID:         1,
			UserID:     b.UserID,
			RoomID:     b.RoomID,
			StartDate:  b.StartDate,
			EndDate:    b.EndDate,
			TotalPrice: b.TotalPrice,
			Status:     b.Status,
			CreatedAt:  now,
			UpdatedAt:  now,
		}, nil).Once()

		res, err := repo.Create(ctx, b)

		assert.NoError(t, err)
		assert.Equal(t, int32(1), res.ID)
		mq.AssertExpectations(t)
	})

	t.Run("conflict", func(t *testing.T) {
		mq.On("CreateBookingSafe", ctx, mock.Anything).Return(Booking{}, sql.ErrNoRows).Once()

		res, err := repo.Create(ctx, &booking.Booking{})

		assert.ErrorIs(t, err, booking.ErrBookingConflict)
		assert.Nil(t, res)
		mq.AssertExpectations(t)
	})
}

func TestSQLCBookingRepository_GetByID(t *testing.T) {
	mq := new(mockQuerier)
	repo := &SQLCBookingRepository{
		queries: mq,
	}
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mq.On("GetBooking", ctx, int32(1)).Return(Booking{ID: 1}, nil).Once()

		res, err := repo.GetByID(ctx, 1)

		assert.NoError(t, err)
		assert.Equal(t, int32(1), res.ID)
		mq.AssertExpectations(t)
	})
}
