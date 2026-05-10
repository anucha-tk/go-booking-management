package dto

import (
	"go-booking-management-init/internal/domain/booking"
	"go-booking-management-init/internal/domain/room"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRoomDTO_Mappers(t *testing.T) {
	now := time.Now()
	rm := &room.Room{
		ID:         1,
		RoomNumber: "101",
		Type:       "Deluxe",
		Price:      1000,
		Status:     "available",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	resp := ToRoomResponse(rm)
	assert.Equal(t, int32(1), resp.ID)
	assert.Equal(t, "101", resp.RoomNumber)
	assert.Equal(t, now.Format(time.RFC3339), resp.CreatedAt)

	list := ToRoomResponseList([]*room.Room{rm})
	assert.Len(t, list, 1)

	detail := &room.Detail{
		Room: *rm,
		Bookings: []room.Booking{
			{ID: 1, StartDate: now, EndDate: now.Add(24 * time.Hour), Status: "confirmed"},
		},
	}
	detailResp := ToRoomDetailResponse(detail)
	assert.Equal(t, "101", detailResp.RoomNumber)
	assert.Len(t, detailResp.Bookings, 1)

	assert.Nil(t, ToRoomResponse(nil))
	assert.Nil(t, ToRoomDetailResponse(nil))
}

func TestBookingDTO_Mappers(t *testing.T) {
	now := time.Now()
	b := &booking.Booking{
		ID:         1,
		UserID:     1,
		RoomID:     101,
		StartDate:  now,
		EndDate:    now.Add(24 * time.Hour),
		TotalPrice: 1000,
		Status:     "confirmed",
		CreatedAt:  now,
	}

	resp := ToBookingResponse(b)
	assert.Equal(t, int32(1), resp.ID)
	assert.Equal(t, now.Format(time.RFC3339), resp.StartDate)

	ub := &booking.UserBookingInfo{
		Booking:    *b,
		RoomNumber: "101",
	}
	ubResp := ToUserBookingResponse(ub)
	assert.Equal(t, "101", ubResp.RoomNumber)

	ab := &booking.AdminBookingInfo{
		UserBookingInfo: *ub,
		UserEmail:       "test@test.com",
	}
	abResp := ToAdminBookingResponse(ab)
	assert.Equal(t, "test@test.com", abResp.UserEmail)

	ubList := ToUserBookingResponseList([]*booking.UserBookingInfo{ub})
	assert.Len(t, ubList, 1)

	abList := ToAdminBookingResponseList([]*booking.AdminBookingInfo{ab})
	assert.Len(t, abList, 1)
}
