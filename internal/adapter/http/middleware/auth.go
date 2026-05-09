package middleware

import (
	"context"
	"net/http"
	"strings"

	"go-booking-management-init/internal/domain/auth"
	"go-booking-management-init/pkg/api"
	pkgAuth "go-booking-management-init/pkg/auth"

	"github.com/gin-gonic/gin"
)

// RevocationChecker is an interface for checking if a token has been revoked.
type RevocationChecker interface {
	IsTokenRevoked(ctx context.Context, accessToken string) (bool, error)
}

const (
	// UserContextKey is the key used to store user claims in the Gin context.
	UserContextKey = "identity.claims"
)

// AuthMiddleware extracts the JWT from the Authorization header and validates it.
func AuthMiddleware(tokenManager pkgAuth.TokenManager, checker RevocationChecker) gin.HandlerFunc {
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

		// Check for revocation
		revoked, err := checker.IsTokenRevoked(c.Request.Context(), tokenStr)
		if err != nil {
			// If we can't check revocation (DB down?), we might want to fail safe or fail closed.
			// Let's fail closed for security.
			api.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Failed to verify token status")
			c.Abort()
			return
		}

		if revoked {
			api.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Token has been revoked")
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
