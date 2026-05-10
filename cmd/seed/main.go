package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"go-booking-management-init/pkg/auth"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	database := os.Getenv("BLUEPRINT_DB_DATABASE")
	password := os.Getenv("BLUEPRINT_DB_PASSWORD")
	username := os.Getenv("BLUEPRINT_DB_USERNAME")
	port := os.Getenv("BLUEPRINT_DB_PORT")
	host := os.Getenv("BLUEPRINT_DB_HOST")
	schema := os.Getenv("BLUEPRINT_DB_SCHEMA")

	userInfo := url.UserPassword(username, password).String()
	connStr := fmt.Sprintf("postgres://%s@%s:%s/%s?sslmode=disable", userInfo, host, port, database)

	if schema != "" {
		connStr = fmt.Sprintf("%s&search_path=%s", connStr, url.QueryEscape(schema))
	}

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fmt.Println("Seeding database...")

	// 1. Truncate tables to ensure clean state
	_, err = db.Exec("TRUNCATE bookings, rooms, revoked_tokens, users RESTART IDENTITY CASCADE")
	if err != nil {
		log.Fatalf("failed to truncate tables: %v", err)
	}

	// 2. Insert Users
	// Roles: admin, officer, customer
	users := []struct {
		email string
		pass  string
		role  string
	}{
		{"admin@example.com", "admin1234", "admin"},
		{"officer@example.com", "officer1234", "officer"},
		{"member1@example.com", "password1234", "customer"},
		{"member2@example.com", "password1234", "customer"},
	}

	for _, u := range users {
		hash, err := auth.HashPassword(u.pass)
		if err != nil {
			log.Fatalf("failed to hash password for %s: %v", u.email, err)
		}
		_, err = db.Exec(`
			INSERT INTO users (email, password_hash, role) 
			VALUES ($1, $2, $3) 
			ON CONFLICT (email) DO UPDATE SET 
				password_hash = EXCLUDED.password_hash,
				role = EXCLUDED.role`,
			u.email, string(hash), u.role)
		if err != nil {
			log.Fatalf("failed to insert user %s: %v", u.email, err)
		}
	}
	fmt.Println("✓ Users seeded")

	// 3. Insert Rooms
	// Price in cents (e.g., 100000 = 1,000.00)
	rooms := []struct {
		number string
		rType  string
		price  int64
		status string
	}{
		{"101", "Single", 100000, "available"},
		{"102", "Double", 180000, "available"},
		{"201", "Suite", 350000, "maintenance"},
		{"202", "Deluxe", 250000, "available"},
		{"301", "Presidential", 1000000, "available"},
		{"302", "Deluxe", 250000, "cleaning"},
	}

	for _, r := range rooms {
		_, err = db.Exec(`
			INSERT INTO rooms (room_number, type, price, status) 
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (room_number) DO UPDATE SET
				type = EXCLUDED.type,
				price = EXCLUDED.price,
				status = EXCLUDED.status`,
			r.number, r.rType, r.price, r.status)
		if err != nil {
			log.Fatalf("failed to insert room %s: %v", r.number, err)
		}
	}
	fmt.Println("✓ Rooms seeded")

	// 4. Insert Bookings
	// Get IDs for some users and rooms
	var userID int32
	var roomID int32
	err = db.QueryRow("SELECT id FROM users WHERE email = 'member1@example.com'").Scan(&userID)
	if err != nil {
		log.Fatalf("failed to find user for booking: %v", err)
	}
	err = db.QueryRow("SELECT id FROM rooms WHERE room_number = '101'").Scan(&roomID)
	if err != nil {
		log.Fatalf("failed to find room for booking: %v", err)
	}

	now := time.Now()
	startDate := now.AddDate(0, 0, 1) // Tomorrow
	endDate := startDate.AddDate(0, 0, 3)

	_, err = db.Exec(`
		INSERT INTO bookings (user_id, room_id, start_date, end_date, total_price, status) 
		VALUES ($1, $2, $3, $4, $5, $6)`,
		userID, roomID, startDate, endDate, 300000, "confirmed")
	if err != nil {
		log.Fatalf("failed to insert booking: %v", err)
	}

	fmt.Println("✓ Bookings seeded")
	fmt.Println("\nDatabase seeded successfully!")
	fmt.Println("Credentials:")
	fmt.Println("  - Admin:   admin@example.com / admin1234")
	fmt.Println("  - Officer: officer@example.com / officer1234")
	fmt.Println("  - Member:  member1@example.com / password1234")
}
