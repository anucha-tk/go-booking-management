package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-booking-management-init/internal/domain/auth"
	pkgAuth "go-booking-management-init/pkg/auth"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockTokenManager struct {
	mock.Mock
}

func (m *MockTokenManager) GenerateToken(userID int32, email string, role string) (string, error) {
	args := m.Called(userID, email, role)
	return args.String(0), args.Error(1)
}

func (m *MockTokenManager) GenerateRefreshToken(userID int32) (string, error) {
	args := m.Called(userID)
	return args.String(0), args.Error(1)
}

func (m *MockTokenManager) ValidateToken(tokenStr string) (*pkgAuth.UserClaims, error) {
	args := m.Called(tokenStr)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pkgAuth.UserClaims), args.Error(1)
}

func (m *MockTokenManager) ValidateRefreshToken(tokenStr string) (*jwt.RegisteredClaims, error) {
	args := m.Called(tokenStr)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*jwt.RegisteredClaims), args.Error(1)
}

type MockRevocationChecker struct {
	mock.Mock
}

func (m *MockRevocationChecker) IsTokenRevoked(ctx context.Context, accessToken string) (bool, error) {
	args := m.Called(ctx, accessToken)
	return args.Bool(0), args.Error(1)
}

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Valid Token", func(t *testing.T) {
		mockTM := new(MockTokenManager)
		mockRC := new(MockRevocationChecker)
		claims := &pkgAuth.UserClaims{
			UserID: 1,
			Email:  "test@example.com",
			Role:   string(auth.RoleAdmin),
		}
		mockTM.On("ValidateToken", "valid-token").Return(claims, nil)
		mockRC.On("IsTokenRevoked", mock.Anything, "valid-token").Return(false, nil)

		r := gin.New()
		r.Use(AuthMiddleware(mockTM, mockRC))
		r.GET("/test", func(c *gin.Context) {
			user, exists := c.Get(UserContextKey)
			assert.True(t, exists)
			assert.Equal(t, claims, user)
			c.Status(http.StatusOK)
		})

		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockTM.AssertExpectations(t)
		mockRC.AssertExpectations(t)
	})

	t.Run("Revoked Token", func(t *testing.T) {
		mockTM := new(MockTokenManager)
		mockRC := new(MockRevocationChecker)
		claims := &pkgAuth.UserClaims{
			UserID: 1,
			Email:  "test@example.com",
			Role:   string(auth.RoleAdmin),
		}
		mockTM.On("ValidateToken", "revoked-token").Return(claims, nil)
		mockRC.On("IsTokenRevoked", mock.Anything, "revoked-token").Return(true, nil)

		r := gin.New()
		r.Use(AuthMiddleware(mockTM, mockRC))
		r.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer revoked-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Token has been revoked")
	})

	t.Run("Missing Authorization Header", func(t *testing.T) {
		mockTM := new(MockTokenManager)
		mockRC := new(MockRevocationChecker)
		r := gin.New()
		r.Use(AuthMiddleware(mockTM, mockRC))
		r.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Invalid Authorization Format", func(t *testing.T) {
		mockTM := new(MockTokenManager)
		mockRC := new(MockRevocationChecker)
		r := gin.New()
		r.Use(AuthMiddleware(mockTM, mockRC))
		r.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Basic invalid")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Invalid Token", func(t *testing.T) {
		mockTM := new(MockTokenManager)
		mockRC := new(MockRevocationChecker)
		mockTM.On("ValidateToken", "invalid-token").Return(nil, assert.AnError)

		r := gin.New()
		r.Use(AuthMiddleware(mockTM, mockRC))
		r.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockTM.AssertExpectations(t)
	})

	t.Run("Valid Token with Multiple Spaces", func(t *testing.T) {
		mockTM := new(MockTokenManager)
		mockRC := new(MockRevocationChecker)
		claims := &pkgAuth.UserClaims{
			UserID: 1,
			Email:  "test@example.com",
			Role:   string(auth.RoleAdmin),
		}
		mockTM.On("ValidateToken", "valid-token").Return(claims, nil)

		r := gin.New()
		r.Use(AuthMiddleware(mockTM, mockRC))
		r.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer  valid-token") // Double space
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestRolesAllowed(t *testing.T) {
	gin.SetMode(gin.TestMode)

	setupRouter := func(roles ...auth.UserRole) *gin.Engine {
		r := gin.New()
		r.Use(func(c *gin.Context) {
			// Mocking AuthMiddleware setting the user
			if role := c.GetHeader("X-Role"); role != "" {
				c.Set(UserContextKey, &pkgAuth.UserClaims{Role: role})
			}
			c.Next()
		})
		r.GET("/admin", RolesAllowed(roles...), func(c *gin.Context) {
			c.Status(http.StatusOK)
		})
		return r
	}

	t.Run("Allowed Role", func(t *testing.T) {
		r := setupRouter(auth.RoleAdmin)
		req, _ := http.NewRequest(http.MethodGet, "/admin", nil)
		req.Header.Set("X-Role", string(auth.RoleAdmin))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Forbidden Role", func(t *testing.T) {
		r := setupRouter(auth.RoleAdmin)
		req, _ := http.NewRequest(http.MethodGet, "/admin", nil)
		req.Header.Set("X-Role", string(auth.RoleCustomer))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Multiple Allowed Roles", func(t *testing.T) {
		r := setupRouter(auth.RoleAdmin, auth.RoleOfficer)

		// Admin allowed
		req1, _ := http.NewRequest(http.MethodGet, "/admin", nil)
		req1.Header.Set("X-Role", string(auth.RoleAdmin))
		w1 := httptest.NewRecorder()
		r.ServeHTTP(w1, req1)
		assert.Equal(t, http.StatusOK, w1.Code)

		// Staff allowed
		req2, _ := http.NewRequest(http.MethodGet, "/admin", nil)
		req2.Header.Set("X-Role", string(auth.RoleOfficer))
		w2 := httptest.NewRecorder()
		r.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusOK, w2.Code)

		// Customer forbidden
		req3, _ := http.NewRequest(http.MethodGet, "/admin", nil)
		req3.Header.Set("X-Role", string(auth.RoleCustomer))
		w3 := httptest.NewRecorder()
		r.ServeHTTP(w3, req3)
		assert.Equal(t, http.StatusForbidden, w3.Code)
	})

	t.Run("No User in Context", func(t *testing.T) {
		r := gin.New()
		r.GET("/admin", RolesAllowed(auth.RoleAdmin), func(c *gin.Context) {
			c.Status(http.StatusOK)
		})
		req, _ := http.NewRequest(http.MethodGet, "/admin", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestAuthMiddleware_RevocationCheckError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockTM := new(MockTokenManager)
	mockRC := new(MockRevocationChecker)
	claims := &pkgAuth.UserClaims{
		UserID: 1,
		Email:  "test@example.com",
		Role:   string(auth.RoleAdmin),
	}
	mockTM.On("ValidateToken", "valid-token").Return(claims, nil)
	mockRC.On("IsTokenRevoked", mock.Anything, "valid-token").Return(false, assert.AnError)

	r := gin.New()
	r.Use(AuthMiddleware(mockTM, mockRC))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to verify token status")
}
