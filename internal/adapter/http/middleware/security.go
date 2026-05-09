package middleware

import "github.com/gin-gonic/gin"

// SecurityHeaders adds standard security headers to the response,
// including the Referrer-Policy requested by the user.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net https://fonts.googleapis.com https://fonts.scalar.com; font-src 'self' https://cdn.jsdelivr.net https://fonts.gstatic.com https://fonts.scalar.com; img-src 'self' data: https://cdn.jsdelivr.net; connect-src 'self' https://api.scalar.com https://cdn.jsdelivr.net;")
		c.Next()
	}
}
