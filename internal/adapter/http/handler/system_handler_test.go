package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSystemHandler_Index(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	h := NewSystemHandler(engine)

	r := gin.New()
	r.GET("/", h.Index)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "success", resp["status"])
}

func TestSystemHandler_DebugRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.GET("/test", func(_ *gin.Context) {})

	h := NewSystemHandler(engine)
	r := gin.New()
	r.GET("/debug/routes", h.DebugRoutes)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/debug/routes", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var routes []map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &routes)
	assert.NoError(t, err)
	assert.NotEmpty(t, routes)
}
