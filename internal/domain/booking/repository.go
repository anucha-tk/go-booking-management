package booking

import (
	"context"
)

type Repository interface {
	Create(ctx context.Context, b *Booking) (*Booking, error)
	GetByID(ctx context.Context, id int32) (*Booking, error)
	ListByUser(ctx context.Context, userID int32) ([]*UserBookingInfo, error)
	ListAll(ctx context.Context) ([]*AdminBookingInfo, error)
	UpdateStatus(ctx context.Context, id int32, status string) (*Booking, error)
}
