package middleware

import (
	"net/http"
	"strings"

	"go-booking-management-init/internal/domain/auth"
	"go-booking-management-init/pkg/api"
	pkgAuth "go-booking-management-init/pkg/auth"

	"github.com/gin-gonic/gin"
)

const (
	// UserContextKey is the key used to store user claims in the Gin context.
	UserContextKey = "identity.claims"
)

// AuthMiddleware extracts the JWT from the Authorization header and validates it.
func AuthMiddleware(tokenManager pkgAuth.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			api.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authorization header is required")
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			api.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authorization header format must be Bearer <token>")
			c.Abort()
			return
		}

		tokenStr := parts[1]
		claims, err := tokenManager.ValidateToken(tokenStr)
		if err != nil {
			api.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired token")
			c.Abort()
			return
		}

		c.Set(UserContextKey, claims)
		c.Next()
	}
}

// RolesAllowed restricts access to users with specific roles.
func RolesAllowed(roles ...auth.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get(UserContextKey)
		if !exists || user == nil {
			api.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
			c.Abort()
			return
		}

		claims, ok := user.(*pkgAuth.UserClaims)
		if !ok {
			api.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user context")
			c.Abort()
			return
		}

		allowed := false
		for _, role := range roles {
			if string(role) == claims.Role {
				allowed = true
				break
			}
		}

		if !allowed {
			api.Error(c, http.StatusForbidden, "FORBIDDEN", "You do not have permission to access this resource")
			c.Abort()
			return
		}

		c.Next()
	}
}
