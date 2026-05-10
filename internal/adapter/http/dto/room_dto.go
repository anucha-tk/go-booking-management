package dto

import (
	"go-booking-management-init/internal/domain/room"
	"time"
)

type CreateRoomRequest struct {
	RoomNumber string `json:"roomNumber" binding:"required"`
	Type       string `json:"type" binding:"required"`
	Price      int64  `json:"price" binding:"required,gt=0"`
}

type UpdateRoomRequest struct {
	RoomNumber string `json:"roomNumber" binding:"required"`
	Type       string `json:"type" binding:"required"`
	Price      int64  `json:"price" binding:"required,gt=0"`
	Status     string `json:"status" binding:"required"`
}

type UpdateRoomStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

type ListRoomsQuery struct {
	Type     *string `form:"type"`
	MinPrice *int64  `form:"price_min"`
	MaxPrice *int64  `form:"price_max"`
	Limit    int32   `form:"limit" binding:"omitempty,gt=0,lte=100"`
	Offset   int32   `form:"offset" binding:"omitempty,gte=0"`
}

type AvailabilityQuery struct {
	StartDate time.Time `form:"start_date" binding:"required" time_format:"2006-01-02"`
	EndDate   time.Time `form:"end_date" binding:"required" time_format:"2006-01-02"`
	Type      *string   `form:"type"`
	MinPrice  *int64    `form:"price_min"`
	MaxPrice  *int64    `form:"price_max"`
	Limit     int32     `form:"limit" binding:"omitempty,gt=0,lte=100"`
	Offset    int32     `form:"offset" binding:"omitempty,gte=0"`
}

type RoomResponse struct {
	ID         int32  `json:"id"`
	RoomNumber string `json:"roomNumber"`
	Type       string `json:"type"`
	Price      int64  `json:"price"`
	Status     string `json:"status"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
}

type BookingResponse struct {
	ID        int32  `json:"id"`
	RoomID    int32  `json:"roomId"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
	Status    string `json:"status"`
}

type RoomDetailResponse struct {
	RoomResponse
	Bookings []BookingResponse `json:"bookings"`
}

func ToRoomResponse(rm *room.Room) *RoomResponse {
	if rm == nil {
		return nil
	}
	return &RoomResponse{
		ID:         rm.ID,
		RoomNumber: rm.RoomNumber,
		Type:       rm.Type,
		Price:      rm.Price,
		Status:     rm.Status,
		CreatedAt:  rm.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  rm.UpdatedAt.Format(time.RFC3339),
	}
}

func ToRoomResponseList(rooms []*room.Room) []*RoomResponse {
	resp := make([]*RoomResponse, len(rooms))
	for i, r := range rooms {
		resp[i] = ToRoomResponse(r)
	}
	return resp
}

func ToRoomDetailResponse(detail *room.Detail) *RoomDetailResponse {
	if detail == nil {
		return nil
	}
	bookings := make([]BookingResponse, len(detail.Bookings))
	for i, b := range detail.Bookings {
		bookings[i] = BookingResponse{
			ID:        b.ID,
			RoomID:    b.RoomID,
			StartDate: b.StartDate.Format("2006-01-02"),
			EndDate:   b.EndDate.Format("2006-01-02"),
			Status:    b.Status,
		}
	}

	return &RoomDetailResponse{
		RoomResponse: *ToRoomResponse(&detail.Room),
		Bookings:     bookings,
	}
}
