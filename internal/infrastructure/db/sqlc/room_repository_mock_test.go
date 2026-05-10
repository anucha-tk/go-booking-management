package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"go-booking-management-init/internal/domain/room"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSQLCRoomRepository_Create(t *testing.T) {
	mq := new(mockQuerier)
	repo := &SQLCRoomRepository{
		queries: mq,
	}
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		rm := &room.Room{
			RoomNumber: "101",
			Type:       "Deluxe",
			Price:      1000,
			Status:     "available",
		}

		now := time.Now()
		mq.On("CreateRoom", ctx, CreateRoomParams{
			RoomNumber: rm.RoomNumber,
			Type:       rm.Type,
			Price:      rm.Price,
			Status:     rm.Status,
		}).Return(Room{
			ID:         1,
			RoomNumber: rm.RoomNumber,
			Type:       rm.Type,
			Price:      rm.Price,
			Status:     rm.Status,
			CreatedAt:  now,
			UpdatedAt:  now,
		}, nil).Once()

		res, err := repo.Create(ctx, rm)

		assert.NoError(t, err)
		assert.Equal(t, int32(1), res.ID)
		assert.Equal(t, rm.RoomNumber, res.RoomNumber)
		mq.AssertExpectations(t)
	})
}

func TestSQLCRoomRepository_GetByID(t *testing.T) {
	mq := new(mockQuerier)
	repo := &SQLCRoomRepository{
		queries: mq,
	}
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		id := int32(1)
		now := time.Now()
		mq.On("GetRoom", ctx, id).Return(Room{
			ID:         id,
			RoomNumber: "101",
			CreatedAt:  now,
			UpdatedAt:  now,
		}, nil).Once()

		res, err := repo.GetByID(ctx, id)

		assert.NoError(t, err)
		assert.Equal(t, id, res.ID)
		mq.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mq.On("GetRoom", ctx, int32(999)).Return(Room{}, sql.ErrNoRows).Once()

		res, err := repo.GetByID(ctx, int32(999))

		assert.NoError(t, err)
		assert.Nil(t, res)
	})
}

func TestSQLCRoomRepository_GetByNumber(t *testing.T) {
	mq := new(mockQuerier)
	repo := &SQLCRoomRepository{
		queries: mq,
	}
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		num := "101"
		mq.On("GetRoomByNumber", ctx, num).Return(Room{
			ID:         1,
			RoomNumber: num,
		}, nil).Once()

		res, err := repo.GetByNumber(ctx, num)

		assert.NoError(t, err)
		assert.Equal(t, num, res.RoomNumber)
		mq.AssertExpectations(t)
	})
}

func TestSQLCRoomRepository_List(t *testing.T) {
	mq := new(mockQuerier)
	repo := &SQLCRoomRepository{
		queries: mq,
	}
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		filter := room.Filter{}
		mq.On("ListRooms", ctx, ListRoomsParams{Limit: 100, Offset: 0}).Return([]Room{{ID: 1}, {ID: 2}}, nil).Once()

		res, err := repo.List(ctx, filter)

		assert.NoError(t, err)
		assert.Len(t, res, 2)
		mq.AssertExpectations(t)
	})

	t.Run("with filters", func(t *testing.T) {
		roomType := "Deluxe"
		minPrice := int64(500)
		maxPrice := int64(1500)
		filter := room.Filter{
			Type:     &roomType,
			MinPrice: &minPrice,
			MaxPrice: &maxPrice,
		}

		mq.On("ListRooms", ctx, mock.MatchedBy(func(p ListRoomsParams) bool {
			return p.Type.String == roomType && p.MinPrice.Int64 == minPrice && p.MaxPrice.Int64 == maxPrice
		})).Return([]Room{{ID: 1, Type: roomType, Price: 1000}}, nil).Once()

		res, err := repo.List(ctx, filter)

		assert.NoError(t, err)
		assert.Len(t, res, 1)
		assert.Equal(t, roomType, res[0].Type)
		mq.AssertExpectations(t)
	})
}

func TestSQLCRoomRepository_Update(t *testing.T) {
	mq := new(mockQuerier)
	repo := &SQLCRoomRepository{
		queries: mq,
	}
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		rm := &room.Room{ID: 1, RoomNumber: "102"}
		mq.On("UpdateRoom", ctx, mock.Anything).Return(Room{ID: 1, RoomNumber: "102"}, nil).Once()

		res, err := repo.Update(ctx, rm)

		assert.NoError(t, err)
		assert.Equal(t, "102", res.RoomNumber)
		mq.AssertExpectations(t)
	})
}

func TestSQLCRoomRepository_Delete(t *testing.T) {
	mq := new(mockQuerier)
	repo := &SQLCRoomRepository{
		queries: mq,
	}
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mq.On("DeleteRoom", ctx, int32(1)).Return(nil).Once()

		err := repo.Delete(ctx, 1)

		assert.NoError(t, err)
		mq.AssertExpectations(t)
	})
}

func TestSQLCRoomRepository_UpdateStatus(t *testing.T) {
	mq := new(mockQuerier)
	repo := &SQLCRoomRepository{
		queries: mq,
	}
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mq.On("UpdateRoomStatus", ctx, UpdateRoomStatusParams{ID: 1, Status: "occupied"}).
			Return(Room{ID: 1, Status: "occupied"}, nil).Once()

		res, err := repo.UpdateStatus(ctx, 1, "occupied")

		assert.NoError(t, err)
		assert.Equal(t, "occupied", res.Status)
		mq.AssertExpectations(t)
	})
}
