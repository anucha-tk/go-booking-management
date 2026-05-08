package dto

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegisterRequest_JSON(t *testing.T) {
	t.Run("valid JSON", func(t *testing.T) {
		data := `{"email":"test@test.com","password":"password123","role":"customer"}`
		var req RegisterRequest
		err := json.Unmarshal([]byte(data), &req)
		assert.NoError(t, err)
		assert.Equal(t, "test@test.com", req.Email)
		assert.Equal(t, "password123", req.Password)
		assert.Equal(t, "customer", req.Role)
	})

	t.Run("empty JSON", func(t *testing.T) {
		data := `{}`
		var req RegisterRequest
		err := json.Unmarshal([]byte(data), &req)
		assert.NoError(t, err)
		assert.Empty(t, req.Email)
		assert.Empty(t, req.Password)
		assert.Empty(t, req.Role)
	})
}

func TestRegisterResponse_JSON(t *testing.T) {
	resp := RegisterResponse{
		ID:        1,
		Email:     "test@test.com",
		Role:      "admin",
		CreatedAt: "2024-01-01T00:00:00Z",
		UpdatedAt: "2024-01-01T00:00:00Z",
	}

	data, err := json.Marshal(resp)
	assert.NoError(t, err)

	var decoded RegisterResponse
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, resp, decoded)
}
