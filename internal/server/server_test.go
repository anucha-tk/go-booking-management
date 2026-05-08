package server

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	// Set env vars to known dummy values
	os.Setenv("PORT", "8080")
	os.Setenv("BLUEPRINT_DB_DATABASE", "testdb")
	os.Setenv("BLUEPRINT_DB_PASSWORD", "testpass")
	os.Setenv("BLUEPRINT_DB_USERNAME", "testuser")
	os.Setenv("BLUEPRINT_DB_PORT", "5432")
	os.Setenv("BLUEPRINT_DB_HOST", "localhost")
	os.Setenv("ALLOW_ORIGINS", "http://localhost:5173")
	os.Setenv("SWAGGER_PATH", "./api/swagger.json")

	// If database is not actually running, sql.Open may still succeed (lazy)
	// but pinging will fail. We just verify NewServer creates non-nil objects.
	srv, router := NewServer()
	assert.NotNil(t, srv)
	assert.NotNil(t, router)
	assert.Equal(t, ":8080", srv.Addr)
}
