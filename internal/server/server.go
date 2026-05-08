package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/joho/godotenv/autoload" // Load .env file

	adapterHttp "go-booking-management-init/internal/adapter/http"
	"go-booking-management-init/internal/adapter/http/handler"
	"go-booking-management-init/internal/database"
)

type Server struct {
	port int
	db   database.Service
}

func NewServer() *http.Server {
	portStr := os.Getenv("PORT")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Printf("invalid PORT %q, using default 8080", portStr)
		port = 8080
	}

	db := database.New()
	if db == nil {
		log.Fatal("database initialization returned nil service")
	}
	s := &Server{
		port: port,
		db:   db,
	}

	// Initialize handlers
	healthHandler := handler.NewHealthHandler(db)

	// Fetch router config from environment
	allowOrigins := strings.Split(os.Getenv("ALLOW_ORIGINS"), ",")
	if len(allowOrigins) == 0 || allowOrigins[0] == "" {
		allowOrigins = []string{"http://localhost:5173"} // Default
	}
	log.Printf("CORS allowed origins: %v", allowOrigins)

	swaggerPath := os.Getenv("SWAGGER_PATH")
	if swaggerPath == "" {
		swaggerPath = "./api/swagger.json" // Default
	}

	// Initialize router
	router := adapterHttp.NewRouter(adapterHttp.Config{
		AllowOrigins: allowOrigins,
		SwaggerPath:  swaggerPath,
	})

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      router.RegisterRoutes(healthHandler),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
