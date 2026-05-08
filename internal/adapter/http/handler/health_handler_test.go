package handler

import (
	"database/sql"
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

func (m *mockDB) DB() *sql.DB {
	return nil
}

type mockDBDown struct{}

func (m *mockDBDown) Health() map[string]string {
	return map[string]string{
		"status": "down",
	}
}

func (m *mockDBDown) Close() error {
	return nil
}

func (m *mockDBDown) DB() *sql.DB {
	return nil
}

func TestHealthHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("healthy database", func(t *testing.T) {
		r := gin.New()
		healthHandler := NewHealthHandler(&mockDB{})
		r.GET("/v1/health", healthHandler.HealthCheck)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/v1/health", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response HealthResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "success", response.Status)
		assert.Equal(t, "OK", response.Data.Status)
		assert.Equal(t, APIVersion, response.Data.Version)
		assert.Equal(t, "up", response.Data.Database["status"])
	})

	t.Run("unhealthy database", func(t *testing.T) {
		r := gin.New()
		healthHandler := NewHealthHandler(&mockDBDown{})
		r.GET("/v1/health", healthHandler.HealthCheck)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/v1/health", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
		assert.Contains(t, w.Body.String(), "SERVICE_UNAVAILABLE")
	})
}
