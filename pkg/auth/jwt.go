package auth

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type UserClaims struct {
	UserID int32  `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type TokenManager interface {
	GenerateToken(userID int32, email string, role string) (string, error)
	GenerateRefreshToken(userID int32) (string, error)
	ValidateToken(tokenStr string) (*UserClaims, error)
	ValidateRefreshToken(tokenStr string) (int32, error)
}

type JWTManager struct {
	secretKey       []byte
	tokenDuration   time.Duration
	refreshDuration time.Duration
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

	refreshDuration := 7 * 24 * time.Hour // Default 7 days
	refreshExpiryStr := os.Getenv("JWT_REFRESH_EXPIRY")
	if refreshExpiryStr != "" {
		if d, err := time.ParseDuration(refreshExpiryStr); err == nil && d > 0 {
			refreshDuration = d
		} else {
			fmt.Printf("WARNING: Invalid or non-positive JWT_REFRESH_EXPIRY '%s', defaulting to 7d\n", refreshExpiryStr)
		}
	}

	return &JWTManager{
		secretKey:       []byte(secretKey),
		tokenDuration:   duration,
		refreshDuration: refreshDuration,
	}
}

func (m *JWTManager) GenerateToken(userID int32, email string, role string) (string, error) {
	claims := UserClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			Subject:   fmt.Sprintf("%d", userID),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secretKey)
}

func (m *JWTManager) GenerateRefreshToken(userID int32) (string, error) {
	claims := jwt.RegisteredClaims{
		ID:        uuid.New().String(),
		Subject:   fmt.Sprintf("%d", userID),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.refreshDuration)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
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

func (m *JWTManager) ValidateRefreshToken(tokenStr string) (int32, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.secretKey, nil
	})

	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		id, err := strconv.ParseInt(claims.Subject, 10, 32)
		if err != nil {
			return 0, errors.New("invalid subject in refresh token")
		}
		return int32(id), nil
	}

	return 0, errors.New("invalid refresh token")
}
