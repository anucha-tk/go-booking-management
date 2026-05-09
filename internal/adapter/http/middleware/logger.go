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

		requestID := c.GetString("request_id")

		c.Next()

		status := c.Writer.Status()
		method := c.Request.Method
		ip := c.ClientIP()
		latency := time.Since(start)

		logAttrs := []any{
			"status", status,
			"method", method,
			"path", path,
			"query", query,
			"ip", ip,
			"latency", latency,
			"request_id", requestID,
		}

		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				slog.Error("request error", append(logAttrs, "error", e.Error())...)
			}
		} else {
			slog.Info("request", logAttrs...)
		}
	}
}
