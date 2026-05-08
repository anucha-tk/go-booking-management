package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		status := c.Writer.Status()
		method := c.Request.Method
		ip := c.ClientIP()
		latency := time.Since(start)

		if len(c.Errors) > 0 {
			for _, e := range c.Errors.Errors() {
				slog.Error(e)
			}
		} else {
			slog.Info("request",
				"status", status,
				"method", method,
				"path", path,
				"query", query,
				"ip", ip,
				"latency", latency,
			)
		}
	}
}
