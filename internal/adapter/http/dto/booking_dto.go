package dto

import (
	"go-booking-management-init/internal/domain/booking"
	"time"
)

type CreateBookingRequest struct {
	RoomID    int32     `json:"roomId" binding:"required"`
	StartDate time.Time `json:"startDate" binding:"required" time_format:"2006-01-02T15:04:05Z07:00"`
	EndDate   time.Time `json:"endDate" binding:"required" time_format:"2006-01-02T15:04:05Z07:00"`
}

type BookingResponse struct {
	ID         int32  `json:"id"`
	UserID     int32  `json:"userId"`
	RoomID     int32  `json:"roomId"`
	StartDate  string `json:"startDate"`
	EndDate    string `json:"endDate"`
	TotalPrice int64  `json:"totalPrice"`
	Status     string `json:"status"`
	CreatedAt  string `json:"createdAt"`
}

type UserBookingResponse struct {
	BookingResponse
	RoomNumber string `json:"roomNumber"`
	RoomType   string `json:"roomType"`
}

type AdminBookingResponse struct {
	UserBookingResponse
	UserEmail string `json:"userEmail"`
}

func ToBookingResponse(b *booking.Booking) BookingResponse {
	return BookingResponse{
		ID:         b.ID,
		UserID:     b.UserID,
		RoomID:     b.RoomID,
		StartDate:  b.StartDate.Format(time.RFC3339),
		EndDate:    b.EndDate.Format(time.RFC3339),
		TotalPrice: b.TotalPrice,
		Status:     b.Status,
		CreatedAt:  b.CreatedAt.Format(time.RFC3339),
	}
}

func ToUserBookingResponse(b *booking.UserBookingInfo) UserBookingResponse {
	return UserBookingResponse{
		BookingResponse: ToBookingResponse(&b.Booking),
		RoomNumber:      b.RoomNumber,
		RoomType:        b.RoomType,
	}
}

func ToAdminBookingResponse(b *booking.AdminBookingInfo) AdminBookingResponse {
	return AdminBookingResponse{
		UserBookingResponse: ToUserBookingResponse(&b.UserBookingInfo),
		UserEmail:           b.UserEmail,
	}
}

func ToUserBookingResponseList(list []*booking.UserBookingInfo) []UserBookingResponse {
	res := make([]UserBookingResponse, len(list))
	for i, v := range list {
		res[i] = ToUserBookingResponse(v)
	}
	return res
}

func ToAdminBookingResponseList(list []*booking.AdminBookingInfo) []AdminBookingResponse {
	res := make([]AdminBookingResponse, len(list))
	for i, v := range list {
		res[i] = ToAdminBookingResponse(v)
	}
	return res
}
