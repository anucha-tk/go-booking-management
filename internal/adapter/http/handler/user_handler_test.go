package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestUserHandler_ListUsers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler()
	r := gin.New()
	r.GET("/admin/users", h.ListUsers)

	req := httptest.NewRequest(http.MethodGet, "/admin/users", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "success", resp["status"])
	data := resp["data"].(map[string]any)
	assert.Contains(t, data, "message")
	assert.Contains(t, data, "users")
}
