package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"net/http/httptest"
	"testing"
	"time"

	"errors"
	"go-booking-management-init/internal/application/auth"
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

func (m *mockAuthService) Login(ctx context.Context, email, password string) (string, string, error) {
	args := m.Called(ctx, email, password)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *mockAuthService) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	args := m.Called(ctx, refreshToken)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *mockAuthService) Logout(ctx context.Context, accessToken string) error {
	args := m.Called(ctx, accessToken)
	return args.Error(0)
}

func (m *mockAuthService) IsTokenRevoked(ctx context.Context, accessToken string) (bool, error) {
	args := m.Called(ctx, accessToken)
	return args.Bool(0), args.Error(1)
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

func TestAuthHandler_Login_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(mockAuthService)
	h := NewAuthHandler(mockSvc)

	r := gin.New()
	r.POST("/login", h.Login)

	mockSvc.On("Login", mock.Anything, "test@example.com", "password123").
		Return("access-token", "refresh-token", nil).Once()

	body := map[string]string{"email": "test@example.com", "password": "password123"}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "success", resp["status"])
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "access-token", data["accessToken"])
	assert.Equal(t, "refresh-token", data["refreshToken"])
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_Login_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(mockAuthService)
	h := NewAuthHandler(mockSvc)

	r := gin.New()
	r.POST("/login", h.Login)

	t.Run("invalid credentials", func(t *testing.T) {
		mockSvc.On("Login", mock.Anything, "wrong@test.com", "password123").
			Return("", "", auth.ErrInvalidCredentials).Once()

		body := map[string]string{"email": "wrong@test.com", "password": "password123"}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("validation error", func(t *testing.T) {
		body := map[string]string{"email": "not-email", "password": ""}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAuthHandler_Refresh_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(mockAuthService)
	h := NewAuthHandler(mockSvc)

	r := gin.New()
	r.POST("/refresh", h.Refresh)

	mockSvc.On("RefreshToken", mock.Anything, "old-refresh-token").
		Return("new-access-token", "new-refresh-token", nil).Once()

	body := map[string]string{"refreshToken": "old-refresh-token"}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "success", resp["status"])
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "new-access-token", data["accessToken"])
	assert.Equal(t, "new-refresh-token", data["refreshToken"])
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_Refresh_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(mockAuthService)
	h := NewAuthHandler(mockSvc)

	r := gin.New()
	r.POST("/refresh", h.Refresh)

	t.Run("invalid token", func(t *testing.T) {
		mockSvc.On("RefreshToken", mock.Anything, "invalid-token").
			Return("", "", auth.ErrInvalidRefreshToken).Once()

		body := map[string]string{"refreshToken": "invalid-token"}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestAuthHandler_Refresh_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuthHandler(nil)
	r := gin.New()
	r.POST("/refresh", h.Refresh)

	req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Logout(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		h := NewAuthHandler(mockSvc)
		r := gin.New()
		r.POST("/logout", h.Logout)

		mockSvc.On("Logout", mock.Anything, "valid-token").Return(nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/logout", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "success", resp["status"])
		mockSvc.AssertExpectations(t)
	})

	t.Run("missing authorization header", func(t *testing.T) {
		h := NewAuthHandler(nil)
		r := gin.New()
		r.POST("/logout", h.Logout)

		req := httptest.NewRequest(http.MethodPost, "/logout", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid token format", func(t *testing.T) {
		h := NewAuthHandler(nil)
		r := gin.New()
		r.POST("/logout", h.Logout)

		req := httptest.NewRequest(http.MethodPost, "/logout", nil)
		req.Header.Set("Authorization", "InvalidHeader")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		h := NewAuthHandler(mockSvc)
		r := gin.New()
		r.POST("/logout", h.Logout)

		mockSvc.On("Logout", mock.Anything, "bad-token").Return(errors.New("service error")).Once()

		req := httptest.NewRequest(http.MethodPost, "/logout", nil)
		req.Header.Set("Authorization", "Bearer bad-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
	})
}
