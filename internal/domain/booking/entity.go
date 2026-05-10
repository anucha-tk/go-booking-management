package booking

import (
	"errors"
	"time"
)

var (
	ErrBookingNotFound     = errors.New("booking not found")
	ErrRoomNotAvailable    = errors.New("room is not available for the selected dates")
	ErrInvalidBookingDates = errors.New("invalid booking dates")
	ErrUnauthorized        = errors.New("unauthorized to access this booking")
	ErrBookingConflict     = errors.New("room is already booked for the selected dates")
)

const (
	StatusConfirmed = "confirmed"
	StatusCancelled = "cancelled"
	StatusPending   = "pending"
)

type Booking struct {
	ID         int32     `json:"id"`
	UserID     int32     `json:"userId"`
	RoomID     int32     `json:"roomId"`
	StartDate  time.Time `json:"startDate"`
	EndDate    time.Time `json:"endDate"`
	TotalPrice int64     `json:"totalPrice"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type CreateParams struct {
	UserID    int32     `json:"userId"`
	RoomID    int32     `json:"roomId"`
	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`
}

type UserBookingInfo struct {
	Booking
	RoomNumber string `json:"roomNumber"`
	RoomType   string `json:"roomType"`
}

type AdminBookingInfo struct {
	UserBookingInfo
	UserEmail string `json:"userEmail"`
}
