package auth

import (
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestJWTManager(t *testing.T) {
	secret := "test-secret"
	os.Setenv("JWT_SECRET", secret)
	os.Setenv("JWT_EXPIRY", "1h")

	manager := NewJWTManager()

	t.Run("Generate and Validate Token", func(t *testing.T) {
		userID := int32(1)
		email := "test@example.com"
		role := "customer"

		token, err := manager.GenerateToken(userID, email, role)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		claims, err := manager.ValidateToken(token)
		assert.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, email, claims.Email)
		assert.Equal(t, role, claims.Role)
	})

	t.Run("Invalid Token", func(t *testing.T) {
		_, err := manager.ValidateToken("invalid-token")
		assert.Error(t, err)
	})

	t.Run("Expired Token", func(t *testing.T) {
		// Set very short expiry
		os.Setenv("JWT_EXPIRY", "1ms")
		managerShort := NewJWTManager()

		token, err := managerShort.GenerateToken(1, "test@example.com", "customer")
		assert.NoError(t, err)

		time.Sleep(2 * time.Millisecond)

		_, err = managerShort.ValidateToken(token)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token is expired")
	})

	t.Run("Generate and Validate Refresh Token", func(t *testing.T) {
		userID := int32(1)

		token, err := manager.GenerateRefreshToken(userID)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		resID, err := manager.ValidateRefreshToken(token)
		assert.NoError(t, err)
		assert.Equal(t, userID, resID)
	})

	t.Run("Invalid Refresh Token", func(t *testing.T) {
		_, err := manager.ValidateRefreshToken("invalid-token")
		assert.Error(t, err)
	})
}

func TestNewJWTManager_Defaults(t *testing.T) {
	// Clear env
	origSecret := os.Getenv("JWT_SECRET")
	origExpiry := os.Getenv("JWT_EXPIRY")
	origRefreshExpiry := os.Getenv("JWT_REFRESH_EXPIRY")
	os.Unsetenv("JWT_SECRET")
	os.Unsetenv("JWT_EXPIRY")
	os.Unsetenv("JWT_REFRESH_EXPIRY")
	defer func() {
		os.Setenv("JWT_SECRET", origSecret)
		os.Setenv("JWT_EXPIRY", origExpiry)
		os.Setenv("JWT_REFRESH_EXPIRY", origRefreshExpiry)
	}()

	t.Run("defaults", func(t *testing.T) {
		manager := NewJWTManager()
		assert.NotNil(t, manager)
		assert.Equal(t, []byte("default-secret"), manager.secretKey)
		assert.Equal(t, time.Hour, manager.tokenDuration)
		assert.Equal(t, 7*24*time.Hour, manager.refreshDuration)
	})

	t.Run("invalid expiry", func(t *testing.T) {
		os.Setenv("JWT_EXPIRY", "invalid")
		os.Setenv("JWT_REFRESH_EXPIRY", "invalid")
		manager := NewJWTManager()
		assert.Equal(t, time.Hour, manager.tokenDuration)
		assert.Equal(t, 7*24*time.Hour, manager.refreshDuration)
	})
}

func TestJWTManager_SigningMethodErrors(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")
	manager := NewJWTManager()

	t.Run("ValidateToken wrong signing method", func(t *testing.T) {
		// Create token signed with RSA instead of HMAC
		token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.RegisteredClaims{
			Subject: "1",
		})
		tokenStr, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)

		_, err := manager.ValidateToken(tokenStr)
		assert.Error(t, err)
	})

	t.Run("ValidateRefreshToken wrong signing method", func(t *testing.T) {
		token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.RegisteredClaims{
			Subject: "1",
		})
		tokenStr, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)

		_, err := manager.ValidateRefreshToken(tokenStr)
		assert.Error(t, err)
	})

	t.Run("ValidateRefreshToken invalid subject", func(t *testing.T) {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
			Subject: "not-a-number",
		})
		tokenStr, _ := token.SignedString([]byte("test-secret"))

		_, err := manager.ValidateRefreshToken(tokenStr)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid subject")
	})
}
