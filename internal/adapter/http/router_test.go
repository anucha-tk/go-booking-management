package http

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
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

func (m *mockDB) DB() *sql.DB {
	return nil
}

func TestRouter_HealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := &mockDB{}
	healthHandler := handler.NewHealthHandler(db)
	router := NewRouter(Config{
		AllowOrigins: []string{"*"},
		SwaggerPath:  "./api/swagger.json",
	}, nil, nil, nil, nil)
	systemHandler := handler.NewSystemHandler(router.Engine())
	h := router.RegisterRoutes(healthHandler, nil, systemHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/health", nil)
	h.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response handler.HealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response.Status)
	assert.Equal(t, "OK", response.Data.Status)
}

func TestRouter_RootRedirect(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := NewRouter(Config{
		AllowOrigins: []string{"*"},
		SwaggerPath:  "./api/swagger.json",
	}, nil, nil, nil, nil)
	systemHandler := handler.NewSystemHandler(router.Engine())
	h := router.RegisterRoutes(nil, nil, systemHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMovedPermanently, w.Code)
}

func TestRouter_API_Index(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := NewRouter(Config{
		AllowOrigins: []string{"*"},
		SwaggerPath:  "./api/swagger.json",
	}, nil, nil, nil, nil)
	systemHandler := handler.NewSystemHandler(router.Engine())
	h := router.RegisterRoutes(nil, nil, systemHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/", nil)
	h.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRouter_DebugRoutes_DevMode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	os.Setenv("APP_ENV", "development")
	defer os.Unsetenv("APP_ENV")

	router := NewRouter(Config{
		AllowOrigins: []string{"*"},
		SwaggerPath:  "./api/swagger.json",
	}, nil, nil, nil, nil)
	systemHandler := handler.NewSystemHandler(router.Engine())
	h := router.RegisterRoutes(nil, nil, systemHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/debug/routes", nil)
	h.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRouter_DebugRoutes_ProdMode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	os.Setenv("APP_ENV", "production")
	defer os.Unsetenv("APP_ENV")

	router := NewRouter(Config{
		AllowOrigins: []string{"*"},
		SwaggerPath:  "./api/swagger.json",
	}, nil, nil, nil, nil)
	systemHandler := handler.NewSystemHandler(router.Engine())
	h := router.RegisterRoutes(nil, nil, systemHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/debug/routes", nil)
	h.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestRouter_SwaggerDoc_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := NewRouter(Config{
		AllowOrigins: []string{"*"},
		SwaggerPath:  "./nonexistent-swagger.json",
	}, nil, nil, nil, nil)
	systemHandler := handler.NewSystemHandler(router.Engine())
	h := router.RegisterRoutes(nil, nil, systemHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/swagger/doc.json", nil)
	h.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestRouter_SwaggerDoc_Redirect(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := NewRouter(Config{
		AllowOrigins: []string{"*"},
		SwaggerPath:  "./api/swagger.json",
	}, nil, nil, nil, nil)
	systemHandler := handler.NewSystemHandler(router.Engine())
	h := router.RegisterRoutes(nil, nil, systemHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/swagger/anything", nil)
	h.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMovedPermanently, w.Code)
}

func TestRouter_SwaggerDoc_Found(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create temp swagger file
	tmpFile, err := os.CreateTemp("", "swagger-*.json")
	assert.NoError(t, err)
	_, err = tmpFile.WriteString(`{"swagger":"2.0"}`)
	assert.NoError(t, err)
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	router := NewRouter(Config{
		AllowOrigins: []string{"*"},
		SwaggerPath:  tmpFile.Name(),
	}, nil, nil, nil, nil)
	systemHandler := handler.NewSystemHandler(router.Engine())
	h := router.RegisterRoutes(nil, nil, systemHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/swagger/doc.json", nil)
	h.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRouter_ScalarDoc(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tmpFile, err := os.CreateTemp("", "swagger-*.json")
	assert.NoError(t, err)
	_, err = tmpFile.WriteString(`{"swagger":"2.0"}`)
	assert.NoError(t, err)
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	router := NewRouter(Config{
		AllowOrigins: []string{"*"},
		SwaggerPath:  tmpFile.Name(),
	}, nil, nil, nil, nil)
	systemHandler := handler.NewSystemHandler(router.Engine())
	h := router.RegisterRoutes(nil, nil, systemHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/doc", nil)
	h.ServeHTTP(w, req)

	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError,
		"expected 200 or 500, got %d: %s", w.Code, w.Body.String())
}

func TestRouter_ScalarDoc_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := NewRouter(Config{
		AllowOrigins: []string{"*"},
		SwaggerPath:  "./nonexistent-scalar.json",
	}, nil, nil, nil, nil)
	systemHandler := handler.NewSystemHandler(router.Engine())
	h := router.RegisterRoutes(nil, nil, systemHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/doc", nil)
	h.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
