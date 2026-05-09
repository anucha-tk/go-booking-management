package http_test

import (
	"bytes"
	"encoding/json"
	"go-booking-management-init/internal/adapter/http/handler"
	"go-booking-management-init/internal/application/auth"
	"go-booking-management-init/internal/database"
	sqlcDB "go-booking-management-init/internal/infrastructure/db/sqlc"
	pkgAuth "go-booking-management-init/pkg/auth"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func TestLoginUser_Integration(t *testing.T) {
	_ = godotenv.Load("../../../.env")
	db := database.New()
	if db == nil || db.DB().Ping() != nil {
		t.Skip("Database not available")
	}
	userRepo := sqlcDB.NewSQLCAuthRepository(db.DB())
	tokenManager := pkgAuth.NewJWTManager()
	authService := auth.NewService(userRepo, tokenManager)
	authHandler := handler.NewAuthHandler(authService)

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/v1/auth/register", authHandler.Register)
	r.POST("/v1/auth/login", authHandler.Login)

	email := "login-test@example.com"
	password := "SecurePass123!"

	// Cleanup and Setup
	if _, err := db.DB().Exec("DELETE FROM users WHERE email = $1", email); err != nil {
		t.Fatalf("failed to cleanup test user: %v", err)
	}

	registerBody := map[string]string{
		"email":    email,
		"password": password,
		"role":     "customer",
	}
	jsonRegBody, _ := json.Marshal(registerBody)

	// Register first
	reqReg, _ := http.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewBuffer(jsonRegBody))
	reqReg.Header.Set("Content-Type", "application/json")
	wReg := httptest.NewRecorder()
	r.ServeHTTP(wReg, reqReg)
	assert.Equal(t, http.StatusCreated, wReg.Code)

	t.Run("success", func(t *testing.T) {
		loginBody := map[string]string{
			"email":    email,
			"password": password,
		}
		jsonLoginBody, _ := json.Marshal(loginBody)

		req, _ := http.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBuffer(jsonLoginBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "success", response["status"], "Response: %s", w.Body.String())

		data, ok := response["data"].(map[string]interface{})
		if assert.True(t, ok, "data should be a map, got: %v", response["data"]) {
			assert.NotEmpty(t, data["accessToken"])
			assert.NotEmpty(t, data["refreshToken"])
		}
	})

	t.Run("invalid password", func(t *testing.T) {
		loginBody := map[string]string{
			"email":    email,
			"password": "wrong-password",
		}
		jsonLoginBody, _ := json.Marshal(loginBody)

		req, _ := http.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBuffer(jsonLoginBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var response map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "error", response["status"])
		assert.Equal(t, "INVALID_CREDENTIALS", response["code"])
	})

	t.Run("user not found", func(t *testing.T) {
		loginBody := map[string]string{
			"email":    "nonexistent@example.com",
			"password": "any-password",
		}
		jsonLoginBody, _ := json.Marshal(loginBody)

		req, _ := http.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBuffer(jsonLoginBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestRefreshToken_Integration(t *testing.T) {
	_ = godotenv.Load("../../../.env")
	db := database.New()
	if db == nil || db.DB().Ping() != nil {
		t.Skip("Database not available")
	}
	userRepo := sqlcDB.NewSQLCAuthRepository(db.DB())
	tokenManager := pkgAuth.NewJWTManager()
	authService := auth.NewService(userRepo, tokenManager)
	authHandler := handler.NewAuthHandler(authService)

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/v1/auth/register", authHandler.Register)
	r.POST("/v1/auth/login", authHandler.Login)
	r.POST("/v1/auth/refresh", authHandler.Refresh)

	email := "refresh-test@example.com"
	password := "SecurePass123!"

	// Cleanup
	if _, err := db.DB().Exec("DELETE FROM users WHERE email = $1", email); err != nil {
		t.Fatalf("failed to cleanup test user: %v", err)
	}

	// Register
	registerBody := map[string]string{"email": email, "password": password, "role": "customer"}
	jsonRegBody, _ := json.Marshal(registerBody)
	reqReg, _ := http.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewBuffer(jsonRegBody))
	reqReg.Header.Set("Content-Type", "application/json")
	wReg := httptest.NewRecorder()
	r.ServeHTTP(wReg, reqReg)

	// Login to get tokens
	loginBody := map[string]string{"email": email, "password": password}
	jsonLoginBody, _ := json.Marshal(loginBody)
	reqLog, _ := http.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBuffer(jsonLoginBody))
	reqLog.Header.Set("Content-Type", "application/json")
	wLog := httptest.NewRecorder()
	r.ServeHTTP(wLog, reqLog)

	var loginResp map[string]interface{}
	_ = json.Unmarshal(wLog.Body.Bytes(), &loginResp)
	oldRefreshToken := loginResp["data"].(map[string]interface{})["refreshToken"].(string)

	t.Run("success", func(t *testing.T) {

		refreshBody := map[string]string{"refreshToken": oldRefreshToken}
		jsonRefreshBody, _ := json.Marshal(refreshBody)

		req, _ := http.NewRequest(http.MethodPost, "/v1/auth/refresh", bytes.NewBuffer(jsonRefreshBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		data := resp["data"].(map[string]interface{})
		assert.NotEmpty(t, data["accessToken"])
		assert.NotEmpty(t, data["refreshToken"])
		assert.NotEqual(t, oldRefreshToken, data["refreshToken"])
	})

	t.Run("invalid refresh token", func(t *testing.T) {
		refreshBody := map[string]string{"refreshToken": "invalid-token"}
		jsonRefreshBody, _ := json.Marshal(refreshBody)

		req, _ := http.NewRequest(http.MethodPost, "/v1/auth/refresh", bytes.NewBuffer(jsonRefreshBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "INVALID_REFRESH_TOKEN", resp["code"])
	})
}
