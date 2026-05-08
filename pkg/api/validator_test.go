package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatValidationError(t *testing.T) {
	v := validator.New()

	t.Run("required field", func(t *testing.T) {
		type req struct {
			Name string `validate:"required"`
		}
		err := v.Struct(req{})
		result := FormatValidationError(err)
		require.NotNil(t, result)
		assert.Contains(t, result["Name"], "required")
	})

	t.Run("email field", func(t *testing.T) {
		type req struct {
			Email string `validate:"email"`
		}
		err := v.Struct(req{Email: "not-an-email"})
		result := FormatValidationError(err)
		require.NotNil(t, result)
		assert.Contains(t, result["Email"], "email")
	})

	t.Run("min tag", func(t *testing.T) {
		type req struct {
			Name string `validate:"min=3"`
		}
		err := v.Struct(req{Name: "ab"})
		result := FormatValidationError(err)
		require.NotNil(t, result)
		assert.Contains(t, result["Name"], "characters")
	})

	t.Run("max tag", func(t *testing.T) {
		type req struct {
			Name string `validate:"max=5"`
		}
		err := v.Struct(req{Name: "toolong"})
		result := FormatValidationError(err)
		require.NotNil(t, result)
		assert.Contains(t, result["Name"], "characters")
	})

	t.Run("oneof tag", func(t *testing.T) {
		type req struct {
			Role string `validate:"oneof=admin user"`
		}
		err := v.Struct(req{Role: "invalid_role"})
		result := FormatValidationError(err)
		require.NotNil(t, result)
		assert.Contains(t, result["Role"], "Must be one of")
	})

	t.Run("multiple fields", func(t *testing.T) {
		type req struct {
			Email string `validate:"required,email"`
			Name  string `validate:"required,min=3,max=10"`
			Role  string `validate:"oneof=admin user"`
		}
		err := v.Struct(req{})
		result := FormatValidationError(err)
		require.NotNil(t, result)
		assert.Contains(t, result, "Email")
		assert.Contains(t, result, "Name")
	})

	t.Run("returns nil for non-validation error", func(t *testing.T) {
		result := FormatValidationError(errors.New("some error"))
		assert.Nil(t, result)
	})

	t.Run("returns nil for nil error", func(t *testing.T) {
		result := FormatValidationError(nil)
		assert.Nil(t, result)
	})
}

func TestValidate(t *testing.T) {
	type validStruct struct {
		Name string `validate:"required"`
	}

	t.Run("passes for valid struct", func(t *testing.T) {
		err := Validate(validStruct{Name: "hello"})
		assert.NoError(t, err)
	})

	t.Run("fails for invalid struct", func(t *testing.T) {
		err := Validate(validStruct{Name: ""})
		assert.Error(t, err)
	})
}

func TestGetValidator(t *testing.T) {
	v := GetValidator()
	assert.NotNil(t, v)
}

func TestInitValidator(_ *testing.T) {
	InitValidator()
	InitValidator()
}

func TestRegisterTagNameFunc(t *testing.T) {
	v := validator.New()
	RegisterTagNameFunc(v)

	assert.NotPanics(t, func() {
		RegisterTagNameFunc(v)
	})
}

func TestBindAndValidate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("valid request", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"email":"test@test.com","password":"password123","role":"customer"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		type testReq struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required,min=8"`
			Role     string `json:"role" binding:"required"`
		}

		var req testReq
		result := BindAndValidate(c, &req)
		assert.True(t, result)
	})

	t.Run("invalid request body", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`not-json`))
		c.Request.Header.Set("Content-Type", "application/json")

		type testReq struct {
			Email string `json:"email" binding:"required"`
		}

		var req testReq
		result := BindAndValidate(c, &req)
		assert.False(t, result)
	})

	t.Run("missing required field", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"email":"test@test.com"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		type testReq struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required"`
		}

		var req testReq
		result := BindAndValidate(c, &req)
		assert.False(t, result)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
