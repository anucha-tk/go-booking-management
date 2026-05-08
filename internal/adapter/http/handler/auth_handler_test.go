package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	domain "go-booking-management-init/internal/domain/auth"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockAuthService struct {
	mock.Mock
}

func (m *mockAuthService) Register(ctx context.Context, email, password string, role domain.UserRole) (*domain.User, error) {
	args := m.Called(ctx, email, password, role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func TestAuthHandler_Register_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(mockAuthService)
	h := NewAuthHandler(mockSvc)

	r := gin.New()
	r.POST("/register", h.Register)

	now := time.Now()
	mockSvc.On("Register", mock.Anything, "test@example.com", "password123", domain.RoleCustomer).
		Return(&domain.User{
			ID: 1, Email: "test@example.com", Role: domain.RoleCustomer,
			CreatedAt: now, UpdatedAt: now,
		}, nil).Once()

	body := map[string]string{"email": "test@example.com", "password": "password123", "role": "customer"}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "success", resp["status"])
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_Register_ValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuthHandler(nil)
	r := gin.New()
	r.POST("/register", h.Register)

	tests := []struct {
		name string
		body map[string]string
	}{
		{"missing email", map[string]string{"password": "password123", "role": "customer"}},
		{"missing password", map[string]string{"email": "test@test.com", "role": "customer"}},
		{"short password", map[string]string{"email": "test@test.com", "password": "short", "role": "customer"}},
		{"invalid email", map[string]string{"email": "not-email", "password": "password123", "role": "customer"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			jsonBody, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestAuthHandler_Register_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(mockAuthService)
	h := NewAuthHandler(mockSvc)

	r := gin.New()
	r.POST("/register", h.Register)

	mockSvc.On("Register", mock.Anything, "exists@test.com", "password123", domain.RoleCustomer).
		Return(nil, domain.ErrUserAlreadyExists).Once()

	body := map[string]string{"email": "exists@test.com", "password": "password123", "role": "customer"}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	mockSvc.AssertExpectations(t)
}
