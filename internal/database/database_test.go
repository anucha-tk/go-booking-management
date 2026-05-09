package database

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	testcontainers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func mustStartPostgresContainer() (func(context.Context, ...testcontainers.TerminateOption) error, error) {
	var (
		dbName = "database"
		dbPwd  = "password"
		dbUser = "user"
	)

	dbContainer, err := postgres.Run(
		context.Background(),
		"postgres:latest",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPwd),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		return nil, err
	}

	os.Setenv("BLUEPRINT_DB_DATABASE", dbName)
	os.Setenv("BLUEPRINT_DB_PASSWORD", dbPwd)
	os.Setenv("BLUEPRINT_DB_USERNAME", dbUser)

	dbHost, err := dbContainer.Host(context.Background())
	if err != nil {
		return dbContainer.Terminate, err
	}

	dbPort, err := dbContainer.MappedPort(context.Background(), "5432/tcp")
	if err != nil {
		return dbContainer.Terminate, err
	}

	os.Setenv("BLUEPRINT_DB_HOST", dbHost)
	os.Setenv("BLUEPRINT_DB_PORT", dbPort.Port())

	return dbContainer.Terminate, err
}

func TestMain(m *testing.M) {
	teardown, err := mustStartPostgresContainer()
	if err != nil {
		log.Printf("Skipping database tests because postgres container could not be started: %s", err)
		os.Exit(0)
	}

	code := m.Run()

	if teardown != nil && teardown(context.Background()) != nil {
		log.Printf("could not teardown postgres container: %s", err)
	}
	os.Exit(code)
}

func TestNew(t *testing.T) {
	srv := New()
	if srv == nil {
		t.Fatal("New() returned nil")
	}
}

func TestReuseConnection(t *testing.T) {
	srv1 := New()
	srv2 := New()
	if srv1 != srv2 {
		t.Fatal("New() should return the same instance")
	}
}

func TestHealth(t *testing.T) {
	srv := New()

	stats := srv.Health()

	if stats["status"] != "up" {
		t.Fatalf("expected status to be up, got %s", stats["status"])
	}

	if _, ok := stats["error"]; ok {
		t.Fatalf("expected error not to be present")
	}

	if stats["message"] != "It's healthy" {
		t.Fatalf("expected message to be 'It's healthy', got %s", stats["message"])
	}
}

func TestClose(t *testing.T) {
	srv := New()

	if srv.Close() != nil {
		t.Fatalf("expected Close() to return nil")
	}
}

func TestDB(t *testing.T) {
	srv := New()
	db := srv.DB()
	if db == nil {
		t.Fatal("DB() returned nil")
	}
}
