package db

import (
	"context"
	"database/sql"
	"errors"
	"go-booking-management-init/internal/domain/room"

	"github.com/jackc/pgx/v5/pgconn"
)

type SQLCRoomRepository struct {
	queries Querier
	db      *sql.DB
}

func NewSQLCRoomRepository(db *sql.DB) *SQLCRoomRepository {
	return &SQLCRoomRepository{
		queries: New(db),
		db:      db,
	}
}

func (r *SQLCRoomRepository) Create(ctx context.Context, rm *room.Room) (*room.Room, error) {
	params := CreateRoomParams{
		RoomNumber: rm.RoomNumber,
		Type:       rm.Type,
		Price:      rm.Price,
		Status:     rm.Status,
	}

	res, err := r.queries.CreateRoom(ctx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, room.ErrRoomNumberExists
		}
		return nil, err
	}

	return mapToDomainRoom(res), nil
}

func (r *SQLCRoomRepository) GetByID(ctx context.Context, id int32) (*room.Room, error) {
	res, err := r.queries.GetRoom(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return mapToDomainRoom(res), nil
}

func (r *SQLCRoomRepository) GetByNumber(ctx context.Context, roomNumber string) (*room.Room, error) {
	res, err := r.queries.GetRoomByNumber(ctx, roomNumber)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return mapToDomainRoom(res), nil
}

func (r *SQLCRoomRepository) List(ctx context.Context, filter room.Filter) ([]*room.Room, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 100 // Default limit for repository level
	}

	params := ListRoomsParams{
		Limit:  limit,
		Offset: filter.Offset,
	}
	if filter.Type != nil {
		params.Type = sql.NullString{String: *filter.Type, Valid: true}
	}
	if filter.MinPrice != nil {
		params.MinPrice = sql.NullInt64{Int64: *filter.MinPrice, Valid: true}
	}
	if filter.MaxPrice != nil {
		params.MaxPrice = sql.NullInt64{Int64: *filter.MaxPrice, Valid: true}
	}

	res, err := r.queries.ListRooms(ctx, params)
	if err != nil {
		return nil, err
	}

	rooms := make([]*room.Room, len(res))
	for i, v := range res {
		rooms[i] = mapToDomainRoom(v)
	}
	return rooms, nil
}

func (r *SQLCRoomRepository) ListAvailable(ctx context.Context, filter room.AvailabilityFilter) ([]*room.Room, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 100 // Default limit for repository level
	}

	params := ListAvailableRoomsParams{
		StartDate:   filter.StartDate,
		EndDate:     filter.EndDate,
		LimitCount:  limit,
		OffsetCount: filter.Offset,
	}
	if filter.Type != nil {
		params.Type = sql.NullString{String: *filter.Type, Valid: true}
	}
	if filter.MinPrice != nil {
		params.MinPrice = sql.NullInt64{Int64: *filter.MinPrice, Valid: true}
	}
	if filter.MaxPrice != nil {
		params.MaxPrice = sql.NullInt64{Int64: *filter.MaxPrice, Valid: true}
	}

	res, err := r.queries.ListAvailableRooms(ctx, params)
	if err != nil {
		return nil, err
	}

	rooms := make([]*room.Room, len(res))
	for i, v := range res {
		rooms[i] = mapToDomainRoom(v)
	}
	return rooms, nil
}

func (r *SQLCRoomRepository) Update(ctx context.Context, rm *room.Room) (*room.Room, error) {
	params := UpdateRoomParams{
		ID:         rm.ID,
		RoomNumber: rm.RoomNumber,
		Type:       rm.Type,
		Price:      rm.Price,
		Status:     rm.Status,
	}

	res, err := r.queries.UpdateRoom(ctx, params)
	if err != nil {
		return nil, err
	}

	return mapToDomainRoom(res), nil
}

func (r *SQLCRoomRepository) Delete(ctx context.Context, id int32) error {
	return r.queries.DeleteRoom(ctx, id)
}

func (r *SQLCRoomRepository) GetDetail(ctx context.Context, id int32) (*room.Detail, error) {
	rm, err := r.GetByID(ctx, id)
	if err != nil || rm == nil {
		return nil, err
	}

	bookings, err := r.queries.ListBookingsByRoom(ctx, ListBookingsByRoomParams{
		RoomID: id,
		Limit:  10, // Limit to last 10 bookings to prevent unbounded data
	})
	if err != nil {
		return nil, err
	}

	domainBookings := make([]room.Booking, len(bookings))
	for i, b := range bookings {
		domainBookings[i] = room.Booking{
			ID:        b.ID,
			RoomID:    b.RoomID,
			StartDate: b.StartDate,
			EndDate:   b.EndDate,
			Status:    b.Status,
		}
	}

	return &room.Detail{
		Room:     *rm,
		Bookings: domainBookings,
	}, nil
}

func (r *SQLCRoomRepository) UpdateStatus(ctx context.Context, id int32, status string) (*room.Room, error) {
	params := UpdateRoomStatusParams{
		ID:     id,
		Status: status,
	}

	res, err := r.queries.UpdateRoomStatus(ctx, params)
	if err != nil {
		return nil, err
	}

	return mapToDomainRoom(res), nil
}

func mapToDomainRoom(rm Room) *room.Room {
	return &room.Room{
		ID:         rm.ID,
		RoomNumber: rm.RoomNumber,
		Type:       rm.Type,
		Price:      rm.Price,
		Status:     rm.Status,
		CreatedAt:  rm.CreatedAt,
		UpdatedAt:  rm.UpdatedAt,
	}
}
