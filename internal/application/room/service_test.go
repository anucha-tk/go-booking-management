package room

import (
	"context"
	"testing"
	"time"

	"go-booking-management-init/internal/domain/room"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) Create(ctx context.Context, rm *room.Room) (*room.Room, error) {
	args := m.Called(ctx, rm)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*room.Room), args.Error(1)
}

func (m *mockRepository) GetByID(ctx context.Context, id int32) (*room.Room, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*room.Room), args.Error(1)
}

func (m *mockRepository) GetByNumber(ctx context.Context, roomNumber string) (*room.Room, error) {
	args := m.Called(ctx, roomNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*room.Room), args.Error(1)
}

func (m *mockRepository) List(ctx context.Context, filter room.Filter) ([]*room.Room, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*room.Room), args.Error(1)
}

func (m *mockRepository) ListAvailable(ctx context.Context, filter room.AvailabilityFilter) ([]*room.Room, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*room.Room), args.Error(1)
}

func (m *mockRepository) GetDetail(ctx context.Context, id int32) (*room.Detail, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*room.Detail), args.Error(1)
}

func (m *mockRepository) Update(ctx context.Context, rm *room.Room) (*room.Room, error) {
	args := m.Called(ctx, rm)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*room.Room), args.Error(1)
}

func (m *mockRepository) Delete(ctx context.Context, id int32) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockRepository) UpdateStatus(ctx context.Context, id int32, status string) (*room.Room, error) {
	args := m.Called(ctx, id, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*room.Room), args.Error(1)
}

func TestService_CreateRoom(t *testing.T) {
	mockRepo := new(mockRepository)
	svc := NewService(mockRepo, nil)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		rm := &room.Room{RoomNumber: "101", Type: "Deluxe", Price: 1000, Status: room.StatusAvailable}
		mockRepo.On("Create", ctx, rm).Return(&room.Room{ID: 1, RoomNumber: "101", Status: room.StatusAvailable}, nil).Once()

		created, err := svc.CreateRoom(ctx, rm)
		assert.NoError(t, err)
		assert.Equal(t, int32(1), created.ID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("duplicate room number", func(t *testing.T) {
		rm := &room.Room{RoomNumber: "101"}
		mockRepo.On("Create", ctx, rm).Return(nil, room.ErrRoomNumberExists).Once()

		created, err := svc.CreateRoom(ctx, rm)
		assert.ErrorIs(t, err, room.ErrRoomNumberExists)
		assert.Nil(t, created)
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid status", func(t *testing.T) {
		rm := &room.Room{RoomNumber: "101", Status: "invalid"}
		created, err := svc.CreateRoom(ctx, rm)
		assert.ErrorIs(t, err, room.ErrInvalidRoomStatus)
		assert.Nil(t, created)
	})
}

func TestService_GetRoom(t *testing.T) {
	mockRepo := new(mockRepository)
	svc := NewService(mockRepo, nil)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockRepo.On("GetByID", ctx, int32(1)).Return(&room.Room{ID: 1, RoomNumber: "101"}, nil).Once()
		rm, err := svc.GetRoom(ctx, 1)
		assert.NoError(t, err)
		assert.Equal(t, "101", rm.RoomNumber)
	})
}

func TestService_ListRooms(t *testing.T) {
	mockRepo := new(mockRepository)
	svc := NewService(mockRepo, nil)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		filter := room.Filter{}
		mockRepo.On("List", ctx, filter).Return([]*room.Room{{ID: 1}, {ID: 2}}, nil).Once()
		rooms, err := svc.ListRooms(ctx, filter)
		assert.NoError(t, err)
		assert.Len(t, rooms, 2)
	})
}

func TestService_UpdateRoom(t *testing.T) {
	mockRepo := new(mockRepository)
	svc := NewService(mockRepo, nil)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		rm := &room.Room{ID: 1, RoomNumber: "102", Status: room.StatusAvailable}
		mockRepo.On("Update", ctx, rm).Return(rm, nil).Once()
		updated, err := svc.UpdateRoom(ctx, rm)
		assert.NoError(t, err)
		assert.Equal(t, "102", updated.RoomNumber)
	})

	t.Run("invalid status", func(t *testing.T) {
		rm := &room.Room{ID: 1, Status: "invalid"}
		updated, err := svc.UpdateRoom(ctx, rm)
		assert.ErrorIs(t, err, room.ErrInvalidRoomStatus)
		assert.Nil(t, updated)
	})
}

func TestService_DeleteRoom(t *testing.T) {
	mockRepo := new(mockRepository)
	svc := NewService(mockRepo, nil)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockRepo.On("Delete", ctx, int32(1)).Return(nil).Once()
		err := svc.DeleteRoom(ctx, 1)
		assert.NoError(t, err)
	})
}

func TestService_UpdateRoomStatus(t *testing.T) {
	mockRepo := new(mockRepository)
	svc := NewService(mockRepo, nil)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockRepo.On("UpdateStatus", ctx, int32(1), "occupied").Return(&room.Room{ID: 1, Status: "occupied"}, nil).Once()
		updated, err := svc.UpdateRoomStatus(ctx, 1, "occupied")
		assert.NoError(t, err)
		assert.Equal(t, "occupied", updated.Status)
	})
	t.Run("invalid status", func(t *testing.T) {
		updated, err := svc.UpdateRoomStatus(ctx, 1, "invalid")
		assert.ErrorIs(t, err, room.ErrInvalidRoomStatus)
		assert.Nil(t, updated)
	})
}

func TestService_ListAvailableRooms(t *testing.T) {
	mockRepo := new(mockRepository)
	svc := NewService(mockRepo, nil)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		start := time.Now().UTC().Add(24 * time.Hour).Truncate(24 * time.Hour)
		end := start.Add(48 * time.Hour)
		filter := room.AvailabilityFilter{
			StartDate: start,
			EndDate:   end,
		}
		mockRepo.On("ListAvailable", mock.Anything, mock.MatchedBy(func(f room.AvailabilityFilter) bool {
			return f.StartDate.Equal(start) && f.EndDate.Equal(end)
		})).Return([]*room.Room{{ID: 1}}, nil).Once()
		rooms, err := svc.ListAvailableRooms(ctx, filter)
		assert.NoError(t, err)
		assert.Len(t, rooms, 1)
	})

	t.Run("missing dates", func(t *testing.T) {
		filter := room.AvailabilityFilter{}
		rooms, err := svc.ListAvailableRooms(ctx, filter)
		assert.ErrorIs(t, err, room.ErrMissingDates)
		assert.Nil(t, rooms)
	})

	t.Run("invalid range", func(t *testing.T) {
		start := time.Now().Add(48 * time.Hour)
		end := start.Add(-24 * time.Hour)
		filter := room.AvailabilityFilter{
			StartDate: start,
			EndDate:   end,
		}
		rooms, err := svc.ListAvailableRooms(ctx, filter)
		assert.ErrorIs(t, err, room.ErrInvalidDateRange)
		assert.Nil(t, rooms)
	})

	t.Run("past date", func(t *testing.T) {
		start := time.Now().Add(-24 * time.Hour)
		end := start.Add(48 * time.Hour)
		filter := room.AvailabilityFilter{
			StartDate: start,
			EndDate:   end,
		}
		rooms, err := svc.ListAvailableRooms(ctx, filter)
		assert.ErrorIs(t, err, room.ErrPastDate)
		assert.Nil(t, rooms)
	})
}

func TestService_GetRoomDetail(t *testing.T) {
	mockRepo := new(mockRepository)
	svc := NewService(mockRepo, nil)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		id := int32(1)
		expected := &room.Detail{
			Room: room.Room{ID: id, RoomNumber: "101"},
			Bookings: []room.Booking{
				{ID: 1, RoomID: id, Status: "confirmed"},
			},
		}

		mockRepo.On("GetDetail", ctx, id).Return(expected, nil).Once()

		result, err := svc.GetRoomDetail(ctx, id)

		assert.NoError(t, err)
		assert.Equal(t, expected, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("not_found", func(t *testing.T) {
		id := int32(99)
		mockRepo.On("GetDetail", ctx, id).Return(nil, nil).Once()

		result, err := svc.GetRoomDetail(ctx, id)

		assert.ErrorIs(t, err, room.ErrRoomNotFound)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}
