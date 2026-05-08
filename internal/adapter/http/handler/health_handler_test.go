package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockDB struct{}

func (m *mockDB) Health() map[string]string {
	return map[string]string{
		"status": "up",
	}
}

func (m *mockDB) Close() error {
	return nil
}

func TestHealthHandler(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	r := gin.New()

	healthHandler := NewHealthHandler(&mockDB{})
	r.GET("/v1/health", healthHandler.HealthCheck)

	// Execute
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/health", nil)
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response HealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response.Status)
	assert.Equal(t, "OK", response.Data.Status)
	assert.Equal(t, APIVersion, response.Data.Version)
	assert.Equal(t, "up", response.Data.Database["status"])
}
