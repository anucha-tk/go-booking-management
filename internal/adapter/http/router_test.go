package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-booking-management-init/internal/adapter/http/handler"

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

func TestRouter_HealthCheck(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	db := &mockDB{}
	healthHandler := handler.NewHealthHandler(db)
	router := NewRouter(Config{
		AllowOrigins: []string{"*"},
		SwaggerPath:  "./api/swagger.json",
	})
	h := router.RegisterRoutes(healthHandler)

	// Execute
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/health", nil)
	h.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response handler.HealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response.Status)
	assert.Equal(t, "OK", response.Data.Status)
}
