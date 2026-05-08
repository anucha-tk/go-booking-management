package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"go-booking-management-init/internal/domain/auth"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	// Try loading from multiple possible locations
	_ = godotenv.Load("../../../../.env")
	_ = godotenv.Load(".env")

	database := os.Getenv("BLUEPRINT_DB_DATABASE")
	password := os.Getenv("BLUEPRINT_DB_PASSWORD")
	username := os.Getenv("BLUEPRINT_DB_USERNAME")
	port := os.Getenv("BLUEPRINT_DB_PORT")
	host := os.Getenv("BLUEPRINT_DB_HOST")

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", username, password, host, port, database)
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		t.Fatalf("failed to connect to test db: %v", err)
	}

	return db, func() {
		_, _ = db.Exec("DELETE FROM users WHERE email = 'test-create@example.com' OR email = 'test-get@example.com'")
		db.Close()
	}
}

func TestUserRepository_CreateUser(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	repo := NewSQLCAuthRepository(db)
	ctx := context.Background()

	email := "test-create@example.com"
	user := &auth.User{
		Email:        email,
		PasswordHash: "hashed_pass",
		Role:         auth.RoleCustomer,
	}

	created, err := repo.Create(ctx, user)
	assert.NoError(t, err)
	assert.NotZero(t, created.ID)
	assert.Equal(t, email, created.Email)
}

func TestUserRepository_GetByEmail(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	repo := NewSQLCAuthRepository(db)
	ctx := context.Background()

	email := "test-get@example.com"
	_, _ = repo.Create(ctx, &auth.User{
		Email:        email,
		PasswordHash: "hashed_pass",
		Role:         auth.RoleCustomer,
	})

	found, err := repo.GetByEmail(ctx, email)
	assert.NoError(t, err)
	assert.Equal(t, email, found.Email)
}
