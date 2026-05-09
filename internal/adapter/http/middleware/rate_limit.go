package middleware

import (
	"net/http"
	"sync"
	"time"

	"go-booking-management-init/pkg/api"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	clients = make(map[string]*client)
	mu      sync.Mutex
)

func init() {
	go cleanupLimiters()
}

func cleanupLimiters() {
	for {
		time.Sleep(time.Minute)
		mu.Lock()
		for ip, c := range clients {
			if time.Since(c.lastSeen) > 3*time.Minute {
				delete(clients, ip)
			}
		}
		mu.Unlock()
	}
}

func getLimiter(ip string, r rate.Limit, b int) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	c, exists := clients[ip]
	if !exists {
		c = &client{
			limiter: rate.NewLimiter(r, b),
		}
		clients[ip] = c
	}
	c.lastSeen = time.Now()

	return c.limiter
}

// ResetRateLimiters clears the internal clients map. Used for testing.
func ResetRateLimiters() {
	mu.Lock()
	defer mu.Unlock()
	clients = make(map[string]*client)
}

func RateLimit(r rate.Limit, b int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := getLimiter(ip, r, b)

		if !limiter.Allow() {
			api.Error(c, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Too many requests")
			c.Abort()
			return
		}

		c.Next()
	}
}
