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

func TestRegisterUser_Integration(t *testing.T) {
	// Load .env from root if it exists, otherwise rely on environment variables
	_ = godotenv.Load("../../../.env")
	_ = godotenv.Load(".env")
	// Setup real dependencies
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

	email := "integration-test@example.com"
	// Cleanup before test
	if err := db.DB().Ping(); err == nil {
		_, _ = db.DB().Exec("DELETE FROM users WHERE email = $1", email)
	}

	body := map[string]string{
		"email":    email,
		"password": "SecurePass123!",
		"role":     "customer",
	}
	jsonBody, _ := json.Marshal(body)

	// Execute
	req, _ := http.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Assert
	if !assert.Equal(t, http.StatusCreated, w.Code) {
		t.Logf("Response body: %s", w.Body.String())
		t.FailNow()
	}
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])

	data, ok := response["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("data is not a map: %v", response["data"])
	}
	assert.Equal(t, email, data["email"])

	// Test Duplicate
	req2, _ := http.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewBuffer(jsonBody))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusConflict, w2.Code)
}

func TestRegisterUser_PasswordTooLong(t *testing.T) {
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

	body := map[string]string{
		"email":    "long-password@example.com",
		"password": "this-is-a-very-long-password-that-should-exceed-the-seventy-two-character-limit-that-we-just-added",
		"role":     "customer",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "PASSWORD_TOO_LONG", response["code"])
}
