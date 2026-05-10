package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"go-booking-management-init/internal/adapter/http/dto"
	"go-booking-management-init/internal/adapter/http/middleware"
	bookingApp "go-booking-management-init/internal/application/booking"
	"go-booking-management-init/internal/domain/booking"
	"go-booking-management-init/internal/domain/room"
	pkgAuth "go-booking-management-init/pkg/auth"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockBookingRepository struct {
	mock.Mock
}

func (m *mockBookingRepository) Create(ctx context.Context, b *booking.Booking) (*booking.Booking, error) {
	args := m.Called(ctx, b)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*booking.Booking), args.Error(1)
}

func (m *mockBookingRepository) GetByID(ctx context.Context, id int32) (*booking.Booking, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*booking.Booking), args.Error(1)
}

func (m *mockBookingRepository) ListByUser(ctx context.Context, userID int32) ([]*booking.UserBookingInfo, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*booking.UserBookingInfo), args.Error(1)
}

func (m *mockBookingRepository) ListAll(ctx context.Context) ([]*booking.AdminBookingInfo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*booking.AdminBookingInfo), args.Error(1)
}

func (m *mockBookingRepository) UpdateStatus(ctx context.Context, id int32, status string) (*booking.Booking, error) {
	args := m.Called(ctx, id, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*booking.Booking), args.Error(1)
}

func TestBookingHandler_CreateBooking(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockBRepo := new(mockBookingRepository)
	mockRRepo := new(mockRoomRepository)
	svc := bookingApp.NewService(mockBRepo, mockRRepo)
	h := NewBookingHandler(svc)

	r := gin.New()
	r.POST("/bookings", h.CreateBooking)

	startDate := time.Now().Add(24 * time.Hour).Truncate(24 * time.Hour)
	endDate := startDate.Add(48 * time.Hour)

	t.Run("success", func(t *testing.T) {
		r := gin.New()
		r.Use(func(c *gin.Context) { c.Set(middleware.UserContextKey, &pkgAuth.UserClaims{UserID: 1}); c.Next() })
		r.POST("/bookings", h.CreateBooking)

		reqBody := dto.CreateBookingRequest{
			RoomID:    101,
			StartDate: startDate,
			EndDate:   endDate,
		}

		mockRRepo.On("GetByID", mock.Anything, int32(101)).Return(&room.Room{ID: 101, Price: 1000}, nil).Once()
		mockRRepo.On("ListAvailable", mock.Anything, mock.Anything).Return([]*room.Room{{ID: 101}}, nil).Once()
		mockBRepo.On("Create", mock.Anything, mock.Anything).Return(&booking.Booking{
			ID: 1, UserID: 1, RoomID: 101, TotalPrice: 2000, Status: "confirmed",
			StartDate: startDate, EndDate: endDate, CreatedAt: time.Now(),
		}, nil).Once()

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/bookings", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("invalid json", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/bookings", bytes.NewBufferString("invalid"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing userId", func(t *testing.T) {
		reqBody := dto.CreateBookingRequest{RoomID: 101, StartDate: startDate, EndDate: endDate}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/bookings", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		r := gin.New()
		r.Use(func(c *gin.Context) { c.Set(middleware.UserContextKey, &pkgAuth.UserClaims{UserID: 1}); c.Next() })
		r.POST("/bookings", h.CreateBooking)

		mockRRepo.On("GetByID", mock.Anything, int32(101)).Return(nil, errors.New("db error")).Once()

		reqBody := dto.CreateBookingRequest{RoomID: 101, StartDate: startDate, EndDate: endDate}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/bookings", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestBookingHandler_GetMyBookings(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockBRepo := new(mockBookingRepository)
	svc := bookingApp.NewService(mockBRepo, nil)
	h := NewBookingHandler(svc)

	t.Run("success", func(t *testing.T) {
		r := gin.New()
		r.Use(func(c *gin.Context) { c.Set(middleware.UserContextKey, &pkgAuth.UserClaims{UserID: 1}); c.Next() })
		r.GET("/bookings/me", h.GetMyBookings)

		mockBRepo.On("ListByUser", mock.Anything, int32(1)).Return([]*booking.UserBookingInfo{{
			Booking:    booking.Booking{ID: 1, StartDate: time.Now(), EndDate: time.Now(), CreatedAt: time.Now()},
			RoomNumber: "101",
		}}, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/bookings/me", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("missing userId", func(t *testing.T) {
		r := gin.New()
		r.GET("/bookings/me", h.GetMyBookings)
		req := httptest.NewRequest(http.MethodGet, "/bookings/me", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestBookingHandler_GetAllBookings(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockBRepo := new(mockBookingRepository)
	svc := bookingApp.NewService(mockBRepo, nil)
	h := NewBookingHandler(svc)

	r := gin.New()
	r.GET("/bookings", h.GetAllBookings)

	t.Run("success", func(t *testing.T) {
		mockBRepo.On("ListAll", mock.Anything).Return([]*booking.AdminBookingInfo{{
			UserBookingInfo: booking.UserBookingInfo{
				Booking: booking.Booking{ID: 1, StartDate: time.Now(), EndDate: time.Now(), CreatedAt: time.Now()},
			},
			UserEmail: "test@test.com",
		}}, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/bookings", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		mockBRepo.On("ListAll", mock.Anything).Return([]*booking.AdminBookingInfo(nil), errors.New("error")).Once()
		req := httptest.NewRequest(http.MethodGet, "/bookings", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestBookingHandler_CancelBooking(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockBRepo := new(mockBookingRepository)
	svc := bookingApp.NewService(mockBRepo, nil)
	h := NewBookingHandler(svc)

	t.Run("success", func(t *testing.T) {
		r := gin.New()
		r.Use(func(c *gin.Context) { c.Set(middleware.UserContextKey, &pkgAuth.UserClaims{UserID: 1}); c.Next() })
		r.POST("/bookings/:id/cancel", h.CancelBooking)

		mockBRepo.On("GetByID", mock.Anything, int32(1)).Return(&booking.Booking{ID: 1, UserID: 1, Status: "confirmed"}, nil).Once()
		mockBRepo.On("UpdateStatus", mock.Anything, int32(1), "cancelled").Return(&booking.Booking{ID: 1, Status: "cancelled"}, nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/bookings/1/cancel", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		r := gin.New()
		r.POST("/bookings/:id/cancel", h.CancelBooking)
		req := httptest.NewRequest(http.MethodPost, "/bookings/abc/cancel", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
