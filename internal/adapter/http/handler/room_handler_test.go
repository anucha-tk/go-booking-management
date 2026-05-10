package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-booking-management-init/internal/adapter/http/dto"
	roomApp "go-booking-management-init/internal/application/room"
	"go-booking-management-init/internal/domain/room"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockRoomRepository struct {
	mock.Mock
}

func (m *mockRoomRepository) Create(ctx context.Context, rm *room.Room) (*room.Room, error) {
	args := m.Called(ctx, rm)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*room.Room), args.Error(1)
}

func (m *mockRoomRepository) GetByID(ctx context.Context, id int32) (*room.Room, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*room.Room), args.Error(1)
}

func (m *mockRoomRepository) GetByNumber(ctx context.Context, roomNumber string) (*room.Room, error) {
	args := m.Called(ctx, roomNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*room.Room), args.Error(1)
}

func (m *mockRoomRepository) List(ctx context.Context, filter room.Filter) ([]*room.Room, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*room.Room), args.Error(1)
}

func (m *mockRoomRepository) ListAvailable(ctx context.Context, filter room.AvailabilityFilter) ([]*room.Room, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*room.Room), args.Error(1)
}

func (m *mockRoomRepository) GetDetail(ctx context.Context, id int32) (*room.Detail, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*room.Detail), args.Error(1)
}

func (m *mockRoomRepository) Update(ctx context.Context, rm *room.Room) (*room.Room, error) {
	args := m.Called(ctx, rm)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*room.Room), args.Error(1)
}

func (m *mockRoomRepository) Delete(ctx context.Context, id int32) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockRoomRepository) UpdateStatus(ctx context.Context, id int32, status string) (*room.Room, error) {
	args := m.Called(ctx, id, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*room.Room), args.Error(1)
}

func TestRoomHandler_CreateRoom(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRepo := new(mockRoomRepository)
	svc := roomApp.NewService(mockRepo, nil)
	h := NewRoomHandler(svc)

	r := gin.New()
	r.POST("/rooms", h.CreateRoom)

	t.Run("success", func(t *testing.T) {
		reqBody := dto.CreateRoomRequest{
			RoomNumber: "101",
			Type:       "Deluxe",
			Price:      1000,
		}
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(&room.Room{
			ID: 1, RoomNumber: "101", Type: "Deluxe", Price: 1000, Status: "available",
			CreatedAt: time.Now(), UpdatedAt: time.Now(),
		}, nil).Once()

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/rooms", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "success", resp["status"])
		mockRepo.AssertExpectations(t)
	})

	t.Run("duplicate room number", func(t *testing.T) {
		reqBody := dto.CreateRoomRequest{
			RoomNumber: "101",
			Type:       "Deluxe",
			Price:      1000,
		}
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil, room.ErrRoomNumberExists).Once()

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/rooms", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
	})
}

func TestRoomHandler_GetRoom(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRepo := new(mockRoomRepository)
	svc := roomApp.NewService(mockRepo, nil)
	h := NewRoomHandler(svc)

	r := gin.New()
	r.GET("/rooms/:id", h.GetRoom)

	t.Run("success", func(t *testing.T) {
		id := int32(1)
		expected := &room.Detail{
			Room: room.Room{ID: id, RoomNumber: "101", Type: "Deluxe", Status: "available"},
			Bookings: []room.Booking{
				{ID: 1, RoomID: id, Status: "confirmed"},
			},
		}

		mockRepo.On("GetDetail", mock.Anything, id).Return(expected, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/rooms/1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		data := resp["data"].(map[string]interface{})
		assert.Equal(t, "101", data["roomNumber"])
		bookings := data["bookings"].([]interface{})
		assert.Len(t, bookings, 1)
		mockRepo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockRepo.On("GetDetail", mock.Anything, int32(99)).Return(nil, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/rooms/99", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/rooms/abc", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestRoomHandler_ListRooms(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRepo := new(mockRoomRepository)
	svc := roomApp.NewService(mockRepo, nil)
	h := NewRoomHandler(svc)

	r := gin.New()
	r.GET("/rooms", h.ListRooms)

	t.Run("success", func(t *testing.T) {
		mockRepo.On("List", mock.Anything, mock.Anything).Return([]*room.Room{{ID: 1}, {ID: 2}}, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/rooms?type=Deluxe&price_min=500", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("binding error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/rooms?limit=abc", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestRoomHandler_UpdateRoom(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRepo := new(mockRoomRepository)
	svc := roomApp.NewService(mockRepo, nil)
	h := NewRoomHandler(svc)

	r := gin.New()
	r.PUT("/rooms/:id", h.UpdateRoom)

	t.Run("success", func(t *testing.T) {
		reqBody := dto.UpdateRoomRequest{
			RoomNumber: "102",
			Type:       "Suite",
			Price:      2000,
			Status:     "available",
		}
		mockRepo.On("Update", mock.Anything, mock.Anything).Return(&room.Room{
			ID: 1, RoomNumber: "102", Type: "Suite", Price: 2000, Status: "available",
			CreatedAt: time.Now(), UpdatedAt: time.Now(),
		}, nil).Once()

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/rooms/1", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/rooms/abc", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("validation error", func(t *testing.T) {
		reqBody := dto.UpdateRoomRequest{
			RoomNumber: "", // Required
		}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/rooms/1", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("not found", func(t *testing.T) {
		mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil, room.ErrRoomNotFound).Once()
		reqBody := dto.UpdateRoomRequest{
			RoomNumber: "102",
			Type:       "Suite",
			Price:      2000,
			Status:     "available",
		}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/rooms/99", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestRoomHandler_DeleteRoom(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRepo := new(mockRoomRepository)
	svc := roomApp.NewService(mockRepo, nil)
	h := NewRoomHandler(svc)

	r := gin.New()
	r.DELETE("/rooms/:id", h.DeleteRoom)

	t.Run("success", func(t *testing.T) {
		mockRepo.On("Delete", mock.Anything, int32(1)).Return(nil).Once()

		req := httptest.NewRequest(http.MethodDelete, "/rooms/1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/rooms/abc", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("not found", func(t *testing.T) {
		mockRepo.On("Delete", mock.Anything, int32(99)).Return(room.ErrRoomNotFound).Once()

		req := httptest.NewRequest(http.MethodDelete, "/rooms/99", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestRoomHandler_UpdateRoomStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRepo := new(mockRoomRepository)
	svc := roomApp.NewService(mockRepo, nil)
	h := NewRoomHandler(svc)

	r := gin.New()
	r.PATCH("/rooms/:id/status", h.UpdateRoomStatus)

	t.Run("success", func(t *testing.T) {
		reqBody := dto.UpdateRoomStatusRequest{
			Status: "occupied",
		}
		mockRepo.On("UpdateStatus", mock.Anything, int32(1), "occupied").Return(&room.Room{
			ID: 1, Status: "occupied",
			CreatedAt: time.Now(), UpdatedAt: time.Now(),
		}, nil).Once()

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPatch, "/rooms/1/status", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid status", func(t *testing.T) {
		reqBody := dto.UpdateRoomStatusRequest{
			Status: "invalid",
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPatch, "/rooms/1/status", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "VALIDATION_ERROR", resp["code"])
	})
}

func TestRoomHandler_SearchAvailability(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRepo := new(mockRoomRepository)
	svc := roomApp.NewService(mockRepo, nil)
	h := NewRoomHandler(svc)

	r := gin.New()
	r.GET("/availability", h.SearchAvailability)

	t.Run("success", func(t *testing.T) {
		start := time.Now().UTC().Add(24 * time.Hour).Truncate(24 * time.Hour)
		end := start.Add(48 * time.Hour)
		mockRepo.On("ListAvailable", mock.Anything, mock.Anything).Return([]*room.Room{{ID: 1}, {ID: 2}, {ID: 3}}, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/availability?start_date="+start.Format("2006-01-02")+"&end_date="+end.Format("2006-01-02"), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		data := resp["data"].([]interface{})
		assert.Len(t, data, 3)
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid query", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/availability?start_date=invalid", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
