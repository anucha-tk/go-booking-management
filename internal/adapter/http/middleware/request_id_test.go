package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go-booking-management-init/internal/adapter/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should generate new request id if not provided", func(t *testing.T) {
		r := gin.New()
		r.Use(middleware.RequestID())
		r.GET("/test", func(c *gin.Context) {
			id := c.GetString("request_id")
			assert.NotEmpty(t, id)
			c.Status(http.StatusOK)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
	})

	t.Run("should use provided request id", func(t *testing.T) {
		r := gin.New()
		r.Use(middleware.RequestID())
		providedID := "test-request-id"
		r.GET("/test", func(c *gin.Context) {
			id := c.GetString("request_id")
			assert.Equal(t, providedID, id)
			c.Status(http.StatusOK)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-Request-ID", providedID)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, providedID, w.Header().Get("X-Request-ID"))
	})
}
