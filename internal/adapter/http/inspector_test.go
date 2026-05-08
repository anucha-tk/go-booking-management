package http

import (
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestInspectRoutes(_ *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/test", func(_ *gin.Context) {})

	// Should not panic
	InspectRoutes(r.Routes(), "")
	InspectRoutes(r.Routes(), "GET")
	InspectRoutes(r.Routes(), "nonexistent")
}

func TestExportRoutes_MarshalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/test", func(_ *gin.Context) {})

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "routes.json")

	// gin.RouteInfo contains HandlerFunc (func type) which is not JSON-serializable
	err := ExportRoutes(r.Routes(), filePath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "json")
}

func TestExportRoutes_WriteError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/test", func(_ *gin.Context) {})

	// Invalid path should produce a write error
	err := ExportRoutes(r.Routes(), "/nonexistent/deep/routes.json")
	assert.Error(t, err)
}
