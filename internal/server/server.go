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
	"go-booking-management-init/internal/application/auth"
	bookingApp "go-booking-management-init/internal/application/booking"
	roomApp "go-booking-management-init/internal/application/room"
	"go-booking-management-init/internal/database"
	"go-booking-management-init/internal/domain/room"
	sqlcDB "go-booking-management-init/internal/infrastructure/db/sqlc"
	pkgAuth "go-booking-management-init/pkg/auth"
)

type Server struct {
	port int
	db   database.Service
}

func NewServer() (*http.Server, *adapterHttp.Router) {
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

	// Initialize repositories
	userRepo := sqlcDB.NewSQLCAuthRepository(db.DB())
	roomRepo := sqlcDB.NewSQLCRoomRepository(db.DB())
	bookingRepo := sqlcDB.NewSQLCBookingRepository(db.DB())

	// Initialize token manager
	tokenManager := pkgAuth.NewJWTManager()

	// Initialize services
	authService := auth.NewService(userRepo, tokenManager)
	roomService := roomApp.NewService(roomRepo, []roomApp.SearchProvider{
		roomApp.NewDBSearchProvider(roomRepo),
		roomApp.NewSimulatedSearchProvider("Partner A", 50*time.Millisecond, []*room.Room{
			{ID: 1001, RoomNumber: "PA-101", Type: "Deluxe", Price: 1500, Status: room.StatusAvailable},
		}),
		roomApp.NewSimulatedSearchProvider("Partner B", 80*time.Millisecond, []*room.Room{
			{ID: 2001, RoomNumber: "PB-201", Type: "Suite", Price: 3000, Status: room.StatusAvailable},
		}),
	})
	bookingService := bookingApp.NewService(bookingRepo, roomRepo)

	// Initialize handlers
	healthHandler := handler.NewHealthHandler(db)
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler()
	roomHandler := handler.NewRoomHandler(roomService)
	bookingHandler := handler.NewBookingHandler(bookingService)

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
	}, tokenManager, authService, userHandler, roomHandler, bookingHandler)

	systemHandler := handler.NewSystemHandler(router.Engine())

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      router.RegisterRoutes(healthHandler, authHandler, systemHandler),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server, router
}
