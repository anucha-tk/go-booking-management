package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	appAuth "go-booking-management-init/internal/application/auth"
	domain "go-booking-management-init/internal/domain/auth"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestMapError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{"nil error", nil, http.StatusOK, ""},
		{"user already exists", domain.ErrUserAlreadyExists, http.StatusConflict, "USER_EXISTS"},
		{"invalid email", appAuth.ErrInvalidEmail, http.StatusUnprocessableEntity, "INVALID_EMAIL"},
		{"invalid role", appAuth.ErrInvalidRole, http.StatusBadRequest, "INVALID_ROLE"},
		{"password too long", appAuth.ErrPasswordTooLong, http.StatusBadRequest, "PASSWORD_TOO_LONG"},
		{"user not found", domain.ErrUserNotFound, http.StatusNotFound, "USER_NOT_FOUND"},
		{"unknown error", errors.New("something unexpected"), http.StatusInternalServerError, "SERVER_ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			MapError(c, tt.err)

			if tt.err == nil {
				assert.Equal(t, 0, w.Body.Len())
				return
			}

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.wantCode)
		})
	}
}

func TestMapError_ValidationErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	v := validator.New()
	type testReq struct {
		Name string `validate:"required"`
	}
	err := v.Struct(testReq{})

	MapError(c, err)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "VALIDATION_ERROR")
}
