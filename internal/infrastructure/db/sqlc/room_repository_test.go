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

func TestRoomRepository_Integration(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	repo := NewSQLCRoomRepository(db)
	ctx := context.Background()

	t.Run("Create and Get", func(t *testing.T) {
		roomNum := "INTEG-101"
		rm := &room.Room{
			RoomNumber: roomNum,
			Type:       "Deluxe",
			Price:      1500,
			Status:     room.StatusAvailable,
		}

		created, err := repo.Create(ctx, rm)
		assert.NoError(t, err)
		assert.NotZero(t, created.ID)
		assert.Equal(t, roomNum, created.RoomNumber)

		// GetByID
		found, err := repo.GetByID(ctx, created.ID)
		assert.NoError(t, err)
		assert.Equal(t, created.ID, found.ID)
		assert.Equal(t, roomNum, found.RoomNumber)

		// GetByNumber
		foundByNum, err := repo.GetByNumber(ctx, roomNum)
		assert.NoError(t, err)
		assert.Equal(t, created.ID, foundByNum.ID)

		// Cleanup
		_ = repo.Delete(ctx, created.ID)
	})

	t.Run("Update and Delete", func(t *testing.T) {
		rm := &room.Room{
			RoomNumber: "INTEG-102",
			Type:       "Standard",
			Price:      800,
			Status:     room.StatusAvailable,
		}

		created, _ := repo.Create(ctx, rm)
		// Update
		created.Type = "Suite"
		created.Price = 2500
		updated, err := repo.Update(ctx, created)
		assert.NoError(t, err)
		assert.Equal(t, "Suite", updated.Type)
		assert.Equal(t, int64(2500), updated.Price)

		// UpdateStatus
		updatedStatus, err := repo.UpdateStatus(ctx, created.ID, room.StatusOccupied)
		assert.NoError(t, err)
		assert.Equal(t, room.StatusOccupied, updatedStatus.Status)

		// Delete
		err = repo.Delete(ctx, created.ID)
		assert.NoError(t, err)

		found, _ := repo.GetByID(ctx, created.ID)
		assert.Nil(t, found)
	})

	t.Run("List and Filter", func(t *testing.T) {
		r1, _ := repo.Create(ctx, &room.Room{RoomNumber: "L-1", Type: "T1", Price: 100, Status: room.StatusAvailable})
		r2, _ := repo.Create(ctx, &room.Room{RoomNumber: "L-2", Type: "T2", Price: 200, Status: room.StatusAvailable})
		defer func() {
			_ = repo.Delete(ctx, r1.ID)
			_ = repo.Delete(ctx, r2.ID)
		}()

		// List all
		list, err := repo.List(ctx, room.Filter{})
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(list), 2)

		// Filter by type
		t1 := "T1"
		listT1, _ := repo.List(ctx, room.Filter{Type: &t1})
		assert.Len(t, listT1, 1)
		assert.Equal(t, "L-1", listT1[0].RoomNumber)

		// Filter by price
		minPrice := int64(150)
		listPrice, _ := repo.List(ctx, room.Filter{MinPrice: &minPrice})
		for _, r := range listPrice {
			assert.GreaterOrEqual(t, r.Price, minPrice)
		}
	})

	t.Run("Availability and Detail", func(t *testing.T) {
		r1, _ := repo.Create(ctx, &room.Room{RoomNumber: "A-1", Type: "Deluxe", Price: 1000, Status: room.StatusAvailable})
		defer func() { _ = repo.Delete(ctx, r1.ID) }()

		// Availability (no bookings)
		start := time.Now().Add(24 * time.Hour)
		end := start.Add(48 * time.Hour)
		available, err := repo.ListAvailable(ctx, room.AvailabilityFilter{
			StartDate: start,
			EndDate:   end,
		})
		assert.NoError(t, err)
		found := false
		for _, r := range available {
			if r.ID == r1.ID {
				found = true
				break
			}
		}
		assert.True(t, found)

		// GetDetail
		detail, err := repo.GetDetail(ctx, r1.ID)
		assert.NoError(t, err)
		assert.Equal(t, "A-1", detail.Room.RoomNumber)
		assert.Empty(t, detail.Bookings)

		// GetDetail with bookings
		bookingRepo := NewSQLCBookingRepository(db)
		authRepo := NewSQLCAuthRepository(db)
		u, _ := authRepo.Create(ctx, &auth.User{Email: "detail@example.com", PasswordHash: "h", Role: "customer"})
		defer func() { _, _ = db.Exec("DELETE FROM users WHERE id = $1", u.ID) }()

		b, _ := bookingRepo.Create(ctx, &booking.Booking{
			UserID:     u.ID,
			RoomID:     r1.ID,
			StartDate:  time.Now().Add(2000 * time.Hour),
			EndDate:    time.Now().Add(2001 * time.Hour),
			TotalPrice: 100,
			Status:     booking.StatusConfirmed,
		})
		defer func() { _, _ = db.Exec("DELETE FROM bookings WHERE id = $1", b.ID) }()

		detailWithB, err := repo.GetDetail(ctx, r1.ID)
		assert.NoError(t, err)
		assert.NotEmpty(t, detailWithB.Bookings)
		assert.Equal(t, b.ID, detailWithB.Bookings[0].ID)
	})
}
