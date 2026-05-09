package auth

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type UserClaims struct {
	UserID int32  `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type TokenManager interface {
	GenerateToken(userID int32, email string, role string) (string, error)
	ValidateToken(tokenStr string) (*UserClaims, error)
}

type JWTManager struct {
	secretKey     []byte
	tokenDuration time.Duration
}

func NewJWTManager() *JWTManager {
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		// Log warning and use fallback for dev; in production this should be a fatal configuration error
		fmt.Println("WARNING: JWT_SECRET not set, using insecure default-secret")
		secretKey = "default-secret"
	}

	duration := time.Hour // Default 1 hour
	expiryStr := os.Getenv("JWT_EXPIRY")
	if expiryStr != "" {
		if d, err := time.ParseDuration(expiryStr); err == nil {
			duration = d
		} else {
			fmt.Printf("WARNING: Invalid JWT_EXPIRY '%s', defaulting to 1h\n", expiryStr)
		}
	}

	return &JWTManager{
		secretKey:     []byte(secretKey),
		tokenDuration: duration,
	}
}

func (m *JWTManager) GenerateToken(userID int32, email string, role string) (string, error) {
	claims := UserClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", userID),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secretKey)
}

func (m *JWTManager) ValidateToken(tokenStr string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
