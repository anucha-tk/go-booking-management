package room

import (
	"context"
	"go-booking-management-init/internal/domain/room"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockRoomRepo struct {
	mock.Mock
}

func (m *mockRoomRepo) Create(ctx context.Context, rm *room.Room) (*room.Room, error) {
	args := m.Called(ctx, rm)
	return args.Get(0).(*room.Room), args.Error(1)
}
func (m *mockRoomRepo) GetByID(ctx context.Context, id int32) (*room.Room, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*room.Room), args.Error(1)
}
func (m *mockRoomRepo) GetByNumber(ctx context.Context, roomNumber string) (*room.Room, error) {
	args := m.Called(ctx, roomNumber)
	return args.Get(0).(*room.Room), args.Error(1)
}
func (m *mockRoomRepo) List(ctx context.Context, filter room.Filter) ([]*room.Room, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*room.Room), args.Error(1)
}
func (m *mockRoomRepo) ListAvailable(ctx context.Context, filter room.AvailabilityFilter) ([]*room.Room, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*room.Room), args.Error(1)
}
func (m *mockRoomRepo) GetDetail(ctx context.Context, id int32) (*room.Detail, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*room.Detail), args.Error(1)
}
func (m *mockRoomRepo) Update(ctx context.Context, rm *room.Room) (*room.Room, error) {
	args := m.Called(ctx, rm)
	return args.Get(0).(*room.Room), args.Error(1)
}
func (m *mockRoomRepo) Delete(ctx context.Context, id int32) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockRoomRepo) UpdateStatus(ctx context.Context, id int32, status string) (*room.Room, error) {
	args := m.Called(ctx, id, status)
	return args.Get(0).(*room.Room), args.Error(1)
}

func TestDBSearchProvider(t *testing.T) {
	mockRepo := new(mockRoomRepo)
	p := NewDBSearchProvider(mockRepo)
	ctx := context.Background()

	assert.Equal(t, "Local Database", p.Name())

	filter := room.AvailabilityFilter{}
	mockRepo.On("ListAvailable", ctx, filter).Return([]*room.Room{{ID: 1}}, nil).Once()

	res, err := p.Search(ctx, filter)
	assert.NoError(t, err)
	assert.Len(t, res, 1)
	mockRepo.AssertExpectations(t)
}

func TestSimulatedSearchProvider(t *testing.T) {
	rooms := []*room.Room{{ID: 1}, {ID: 2}}
	p := NewSimulatedSearchProvider("Mock Provider", 10*time.Millisecond, rooms)
	ctx := context.Background()

	assert.Equal(t, "Mock Provider", p.Name())

	start := time.Now()
	res, err := p.Search(ctx, room.AvailabilityFilter{})
	assert.NoError(t, err)
	assert.Len(t, res, 2)
	assert.GreaterOrEqual(t, time.Since(start), 10*time.Millisecond)

	// Context timeout
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	res, err = p.Search(ctxTimeout, room.AvailabilityFilter{})
	assert.Error(t, err)
	assert.Nil(t, res)
}
