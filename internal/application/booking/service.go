package booking

import (
	"context"
	"go-booking-management-init/internal/domain/booking"
	"go-booking-management-init/internal/domain/room"
	"time"
)

type Service struct {
	bookingRepo booking.Repository
	roomRepo    room.Repository
}

func NewService(bookingRepo booking.Repository, roomRepo room.Repository) *Service {
	return &Service{
		bookingRepo: bookingRepo,
		roomRepo:    roomRepo,
	}
}

func (s *Service) SubmitBooking(ctx context.Context, params booking.CreateParams) (*booking.Booking, error) {
	// 1. Validate dates - truncate to day for consistency
	start := params.StartDate.UTC().Truncate(24 * time.Hour)
	end := params.EndDate.UTC().Truncate(24 * time.Hour)

	if start.IsZero() || end.IsZero() {
		return nil, booking.ErrInvalidBookingDates
	}
	if !start.Before(end) {
		return nil, booking.ErrInvalidBookingDates
	}
	today := time.Now().UTC().Truncate(24 * time.Hour)
	if start.Before(today) {
		return nil, booking.ErrInvalidBookingDates
	}

	// 2. Check room exists and get price
	rm, err := s.roomRepo.GetByID(ctx, params.RoomID)
	if err != nil {
		return nil, err
	}
	if rm == nil {
		return nil, room.ErrRoomNotFound
	}

	// 3. Check availability
	availableRooms, err := s.roomRepo.ListAvailable(ctx, room.AvailabilityFilter{
		StartDate: start,
		EndDate:   end,
	})
	if err != nil {
		return nil, err
	}

	isAvailable := false
	for _, ar := range availableRooms {
		if ar.ID == params.RoomID {
			isAvailable = true
			break
		}
	}
	if !isAvailable {
		return nil, booking.ErrRoomNotAvailable
	}

	// 4. Calculate total price based on nights
	nights := int64(end.Sub(start).Hours() / 24)
	if nights <= 0 {
		nights = 1
	}
	totalPrice := rm.Price * nights

	// 5. Create booking
	b := &booking.Booking{
		UserID:     params.UserID,
		RoomID:     params.RoomID,
		StartDate:  start,
		EndDate:    end,
		TotalPrice: totalPrice,
		Status:     booking.StatusConfirmed,
	}

	return s.bookingRepo.Create(ctx, b)
}

func (s *Service) GetMyBookings(ctx context.Context, userID int32) ([]*booking.UserBookingInfo, error) {
	return s.bookingRepo.ListByUser(ctx, userID)
}

func (s *Service) GetAllBookings(ctx context.Context) ([]*booking.AdminBookingInfo, error) {
	return s.bookingRepo.ListAll(ctx)
}

func (s *Service) CancelBooking(ctx context.Context, userID int32, isAdmin bool, bookingID int32) (*booking.Booking, error) {
	b, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	if b == nil {
		return nil, booking.ErrBookingNotFound
	}

	// Check ownership or admin override
	if !isAdmin && b.UserID != userID {
		return nil, booking.ErrUnauthorized
	}

	if b.Status == booking.StatusCancelled {
		return b, nil // Already cancelled
	}

	return s.bookingRepo.UpdateStatus(ctx, bookingID, booking.StatusCancelled)
}
