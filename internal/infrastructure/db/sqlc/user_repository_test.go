package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"testing"

	"go-booking-management-init/internal/domain/auth"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func uniqueEmail(t *testing.T) string {
	return fmt.Sprintf("%s.%d@example.com", t.Name(), os.Getpid())
}

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	_ = godotenv.Load("../../../../.env")
	_ = godotenv.Load(".env")

	database := os.Getenv("BLUEPRINT_DB_DATABASE")
	if database == "" {
		t.Skip("BLUEPRINT_DB_DATABASE not set")
	}

	password := os.Getenv("BLUEPRINT_DB_PASSWORD")
	username := os.Getenv("BLUEPRINT_DB_USERNAME")
	port := os.Getenv("BLUEPRINT_DB_PORT")
	host := os.Getenv("BLUEPRINT_DB_HOST")

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", username, password, host, port, database)
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		t.Fatalf("failed to connect to test db: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Skipf("database not reachable: %v", err)
	}

	return db, func() {
		db.Close()
	}
}

func TestUserRepository_CreateUser(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	repo := NewSQLCAuthRepository(db)
	ctx := context.Background()

	email := uniqueEmail(t)
	user := &auth.User{
		Email:        email,
		PasswordHash: "hashed_pass",
		Role:         auth.RoleCustomer,
	}

	created, err := repo.Create(ctx, user)
	assert.NoError(t, err)
	assert.NotZero(t, created.ID)
	assert.Equal(t, email, created.Email)

	// Cleanup
	_, _ = db.Exec("DELETE FROM users WHERE email = $1", email)
}

func TestUserRepository_GetByEmail(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	repo := NewSQLCAuthRepository(db)
	ctx := context.Background()

	email := uniqueEmail(t)
	_, _ = repo.Create(ctx, &auth.User{
		Email:        email,
		PasswordHash: "hashed_pass",
		Role:         auth.RoleCustomer,
	})

	found, err := repo.GetByEmail(ctx, email)
	assert.NoError(t, err)
	assert.Equal(t, email, found.Email)

	// Cleanup
	_, _ = db.Exec("DELETE FROM users WHERE email = $1", email)
}

func TestUserRepository_GetByEmail_NotFound(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	repo := NewSQLCAuthRepository(db)
	ctx := context.Background()

	user, err := repo.GetByEmail(ctx, uniqueEmail(t))
	assert.Error(t, err)
	assert.True(t, errors.Is(err, auth.ErrUserNotFound))
	assert.Nil(t, user)
}

func TestUserRepository_DuplicateEmail(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	repo := NewSQLCAuthRepository(db)
	ctx := context.Background()

	email := uniqueEmail(t)

	// First create should succeed
	_, err := repo.Create(ctx, &auth.User{
		Email:        email,
		PasswordHash: "hash1",
		Role:         auth.RoleCustomer,
	})
	assert.NoError(t, err)

	// Second create with same email should fail
	_, err = repo.Create(ctx, &auth.User{
		Email:        email,
		PasswordHash: "hash2",
		Role:         auth.RoleCustomer,
	})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, auth.ErrUserAlreadyExists))

	// Cleanup
	_, _ = db.Exec("DELETE FROM users WHERE email = $1", email)
}

func TestWithTx(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	assert.NoError(t, err)

	queries := New(db)
	txQueries := queries.WithTx(tx)
	assert.NotNil(t, txQueries)

	email := uniqueEmail(t)
	params := CreateUserParams{
		Email:        email,
		PasswordHash: "tx_hash",
		Role:         "customer",
	}
	user, err := txQueries.CreateUser(ctx, params)
	assert.NoError(t, err)
	assert.Equal(t, email, user.Email)
	assert.NotZero(t, user.ID)

	// Rollback should remove the user from the database
	err = tx.Rollback()
	assert.NoError(t, err)

	repo := NewSQLCAuthRepository(db)
	_, err = repo.GetByEmail(ctx, email)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, auth.ErrUserNotFound))
}
