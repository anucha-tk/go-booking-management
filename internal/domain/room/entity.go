package room

import (
	"context"
	"errors"
	"time"
)

var (
	ErrRoomNotFound      = errors.New("room not found")
	ErrRoomNumberExists  = errors.New("room number already exists")
	ErrInvalidRoomStatus = errors.New("invalid room status")
	ErrInvalidDateRange  = errors.New("invalid date range: start date must be before end date")
	ErrPastDate          = errors.New("start date must be in the future")
	ErrMissingDates      = errors.New("start date and end date are required")
)

const (
	StatusAvailable   = "available"
	StatusOccupied    = "occupied"
	StatusMaintenance = "maintenance"
	StatusCleaning    = "cleaning"
	StatusReserved    = "reserved"
)

func IsValidStatus(status string) bool {
	switch status {
	case StatusAvailable, StatusOccupied, StatusMaintenance, StatusCleaning, StatusReserved:
		return true
	}
	return false
}

type Room struct {
	ID         int32     `json:"id"`
	RoomNumber string    `json:"roomNumber"`
	Type       string    `json:"type"`
	Price      int64     `json:"price"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type Booking struct {
	ID        int32     `json:"id"`
	RoomID    int32     `json:"roomId"`
	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`
	Status    string    `json:"status"`
}

type Detail struct {
	Room
	Bookings []Booking `json:"bookings"`
}

type Filter struct {
	Type     *string
	MinPrice *int64
	MaxPrice *int64
	Limit    int32
	Offset   int32
}

type AvailabilityFilter struct {
	StartDate time.Time
	EndDate   time.Time
	Type      *string
	MinPrice  *int64
	MaxPrice  *int64
	Limit     int32
	Offset    int32
}

type Repository interface {
	Create(ctx context.Context, room *Room) (*Room, error)
	GetByID(ctx context.Context, id int32) (*Room, error)
	GetByNumber(ctx context.Context, roomNumber string) (*Room, error)
	List(ctx context.Context, filter Filter) ([]*Room, error)
	ListAvailable(ctx context.Context, filter AvailabilityFilter) ([]*Room, error)
	GetDetail(ctx context.Context, id int32) (*Detail, error)
	Update(ctx context.Context, room *Room) (*Room, error)
	Delete(ctx context.Context, id int32) error
	UpdateStatus(ctx context.Context, id int32, status string) (*Room, error)
}
