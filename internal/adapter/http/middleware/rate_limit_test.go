package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-booking-management-init/internal/adapter/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should allow requests within limit", func(t *testing.T) {
		middleware.ResetRateLimiters()
		r := gin.New()
		// Limit to 10 req/s for testing
		r.Use(middleware.RateLimit(10, 10))
		r.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		for i := 0; i < 5; i++ {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			r.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
		}
	})

	t.Run("should reject requests exceeding limit", func(t *testing.T) {
		middleware.ResetRateLimiters()
		r := gin.New()
		// Limit to 1 req/s, burst 1
		r.Use(middleware.RateLimit(1, 1))
		r.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		// First request allowed
		w1 := httptest.NewRecorder()
		req1, _ := http.NewRequest(http.MethodGet, "/test", nil)
		r.ServeHTTP(w1, req1)
		assert.Equal(t, http.StatusOK, w1.Code)

		// Second request rejected
		w2 := httptest.NewRecorder()
		req2, _ := http.NewRequest(http.MethodGet, "/test", nil)
		r.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusTooManyRequests, w2.Code)
	})

	t.Run("should allow requests after waiting", func(t *testing.T) {
		middleware.ResetRateLimiters()
		r := gin.New()
		// Limit to 2 req/s, burst 1
		r.Use(middleware.RateLimit(2, 1))
		r.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		w1 := httptest.NewRecorder()
		req1, _ := http.NewRequest(http.MethodGet, "/test", nil)
		r.ServeHTTP(w1, req1)
		assert.Equal(t, http.StatusOK, w1.Code)

		// Wait for token to replenish (0.5s for 2 req/s)
		time.Sleep(510 * time.Millisecond)

		w2 := httptest.NewRecorder()
		req2, _ := http.NewRequest(http.MethodGet, "/test", nil)
		r.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusOK, w2.Code)
	})
}
