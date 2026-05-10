package db

import (
	"context"
	"database/sql"
	"errors"
	"go-booking-management-init/internal/domain/booking"
)

type SQLCBookingRepository struct {
	queries Querier
	db      *sql.DB
}

func NewSQLCBookingRepository(db *sql.DB) *SQLCBookingRepository {
	return &SQLCBookingRepository{
		queries: New(db),
		db:      db,
	}
}

func (r *SQLCBookingRepository) Create(ctx context.Context, b *booking.Booking) (*booking.Booking, error) {
	params := CreateBookingSafeParams{
		UserID:     b.UserID,
		RoomID:     b.RoomID,
		StartDate:  b.StartDate,
		EndDate:    b.EndDate,
		TotalPrice: b.TotalPrice,
		Status:     b.Status,
	}

	res, err := r.queries.CreateBookingSafe(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, booking.ErrBookingConflict
		}
		return nil, err
	}

	return mapToDomainBooking(res), nil
}

func (r *SQLCBookingRepository) GetByID(ctx context.Context, id int32) (*booking.Booking, error) {
	res, err := r.queries.GetBooking(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, booking.ErrBookingNotFound
		}
		return nil, err
	}

	return mapToDomainBooking(res), nil
}
func (r *SQLCBookingRepository) ListByUser(ctx context.Context, userID int32) ([]*booking.UserBookingInfo, error) {
	res, err := r.queries.ListBookingsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	bookings := make([]*booking.UserBookingInfo, len(res))
	for i, v := range res {
		bookings[i] = &booking.UserBookingInfo{
			Booking: *mapToDomainBooking(v),
		}
	}
	return bookings, nil
}

func (r *SQLCBookingRepository) ListByRoom(ctx context.Context, roomID int32) ([]*booking.Booking, error) {
	res, err := r.queries.ListBookingsByRoom(ctx, ListBookingsByRoomParams{
		RoomID: roomID,
		Limit:  100,
	})
	if err != nil {
		return nil, err
	}

	bookings := make([]*booking.Booking, len(res))
	for i, v := range res {
		bookings[i] = &booking.Booking{
			ID:        v.ID,
			RoomID:    v.RoomID,
			StartDate: v.StartDate,
			EndDate:   v.EndDate,
			Status:    v.Status,
		}
	}
	return bookings, nil
}

func (r *SQLCBookingRepository) ListAll(ctx context.Context) ([]*booking.AdminBookingInfo, error) {
	res, err := r.queries.ListAllBookings(ctx)
	if err != nil {
		return nil, err
	}

	bookings := make([]*booking.AdminBookingInfo, len(res))
	for i, v := range res {
		bookings[i] = &booking.AdminBookingInfo{
			UserBookingInfo: booking.UserBookingInfo{
				Booking:    *mapToDomainBookingFromAll(v),
				RoomNumber: v.RoomNumber,
				RoomType:   v.RoomType,
			},
			UserEmail: v.UserEmail,
		}
	}
	return bookings, nil
}

func (r *SQLCBookingRepository) UpdateStatus(ctx context.Context, id int32, status string) (*booking.Booking, error) {
	params := UpdateBookingStatusParams{
		ID:     id,
		Status: status,
	}

	res, err := r.queries.UpdateBookingStatus(ctx, params)
	if err != nil {
		return nil, err
	}

	return mapToDomainBooking(res), nil
}

func mapToDomainBooking(b Booking) *booking.Booking {
	return &booking.Booking{
		ID:         b.ID,
		UserID:     b.UserID,
		RoomID:     b.RoomID,
		StartDate:  b.StartDate,
		EndDate:    b.EndDate,
		TotalPrice: b.TotalPrice,
		Status:     b.Status,
		CreatedAt:  b.CreatedAt,
		UpdatedAt:  b.UpdatedAt,
	}
}

func mapToDomainBookingFromAll(v ListAllBookingsRow) *booking.Booking {
	return &booking.Booking{
		ID:         v.ID,
		UserID:     v.UserID,
		RoomID:     v.RoomID,
		StartDate:  v.StartDate,
		EndDate:    v.EndDate,
		TotalPrice: v.TotalPrice,
		Status:     v.Status,
		CreatedAt:  v.CreatedAt,
		UpdatedAt:  v.UpdatedAt,
	}
}
