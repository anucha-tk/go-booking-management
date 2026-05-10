package db

import (
	"context"
	"go-booking-management-init/internal/domain/auth"
	"go-booking-management-init/internal/domain/booking"
	"go-booking-management-init/internal/domain/room"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBookingRepository_Integration(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	repo := NewSQLCBookingRepository(db)
	roomRepo := NewSQLCRoomRepository(db)
	authRepo := NewSQLCAuthRepository(db)
	ctx := context.Background()

	// 1. Setup User and Room
	email := uniqueEmail(t)
	u, _ := authRepo.Create(ctx, &auth.User{Email: email, PasswordHash: "hash", Role: auth.RoleCustomer})
	defer func() { _, _ = db.Exec("DELETE FROM users WHERE id = $1", u.ID) }()

	roomNum := "B-IT-101"
	rm, _ := roomRepo.Create(ctx, &room.Room{RoomNumber: roomNum, Type: "Deluxe", Price: 1000, Status: room.StatusAvailable})
	defer func() { _ = roomRepo.Delete(ctx, rm.ID) }()

	startDate := time.Now().Add(24 * time.Hour).Truncate(24 * time.Hour)
	endDate := startDate.Add(48 * time.Hour)

	t.Run("Create and Get", func(t *testing.T) {
		b := &booking.Booking{
			UserID:     u.ID,
			RoomID:     rm.ID,
			StartDate:  startDate,
			EndDate:    endDate,
			TotalPrice: 2000,
			Status:     booking.StatusConfirmed,
		}

		created, err := repo.Create(ctx, b)
		assert.NoError(t, err)
		assert.NotZero(t, created.ID)
		defer func() { _, _ = db.Exec("DELETE FROM bookings WHERE id = $1", created.ID) }()

		// GetByID
		found, err := repo.GetByID(ctx, created.ID)
		assert.NoError(t, err)
		assert.Equal(t, created.ID, found.ID)

		// ListByUser
		list, err := repo.ListByUser(ctx, u.ID)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(list), 1)

		// ListAll
		all, err := repo.ListAll(ctx)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(all), 1)
	})

	t.Run("Booking Conflict", func(t *testing.T) {
		b1 := &booking.Booking{
			UserID:     u.ID,
			RoomID:     rm.ID,
			StartDate:  startDate,
			EndDate:    endDate,
			TotalPrice: 2000,
			Status:     booking.StatusConfirmed,
		}

		created1, err := repo.Create(ctx, b1)
		assert.NoError(t, err)
		defer func() { _, _ = db.Exec("DELETE FROM bookings WHERE id = $1", created1.ID) }()

		// Try overlapping booking
		b2 := &booking.Booking{
			UserID:     u.ID,
			RoomID:     rm.ID,
			StartDate:  startDate.Add(1 * time.Hour),
			EndDate:    endDate.Add(1 * time.Hour),
			TotalPrice: 2000,
			Status:     booking.StatusConfirmed,
		}

		res2, err := repo.Create(ctx, b2)
		assert.ErrorIs(t, err, booking.ErrBookingConflict)
		assert.Nil(t, res2)
	})

	t.Run("Update Status and Restore Availability", func(t *testing.T) {
		b := &booking.Booking{
			UserID:     u.ID,
			RoomID:     rm.ID,
			StartDate:  startDate.Add(400 * time.Hour),
			EndDate:    endDate.Add(400 * time.Hour),
			TotalPrice: 2000,
			Status:     booking.StatusConfirmed,
		}

		created, err := repo.Create(ctx, b)
		assert.NoError(t, err)
		defer func() { _, _ = db.Exec("DELETE FROM bookings WHERE id = $1", created.ID) }()

		// Cancel it
		updated, err := repo.UpdateStatus(ctx, created.ID, booking.StatusCancelled)
		assert.NoError(t, err)
		assert.Equal(t, booking.StatusCancelled, updated.Status)

		// Now try to book again for the same period
		b2 := &booking.Booking{
			UserID:     u.ID,
			RoomID:     rm.ID,
			StartDate:  b.StartDate,
			EndDate:    b.EndDate,
			TotalPrice: 2000,
			Status:     booking.StatusConfirmed,
		}
		created2, err := repo.Create(ctx, b2)
		assert.NoError(t, err)
		assert.NotNil(t, created2)
		defer func() { _, _ = db.Exec("DELETE FROM bookings WHERE id = $1", created2.ID) }()
	})
}
