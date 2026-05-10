package http

import (
	"log/slog"
	"net/http"
	"os"

	"go-booking-management-init/internal/adapter/http/handler"
	"go-booking-management-init/internal/adapter/http/middleware"
	applicationAuth "go-booking-management-init/internal/application/auth"
	"go-booking-management-init/internal/domain/auth"
	"go-booking-management-init/pkg/api"
	pkgAuth "go-booking-management-init/pkg/auth"

	scalar "github.com/MarceloPetrucio/go-scalar-api-reference"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Config struct {
	AllowOrigins []string
	SwaggerPath  string
}

type Router struct {
	engine         *gin.Engine
	config         Config
	tokenManager   pkgAuth.TokenManager
	authService    applicationAuth.Service
	userHandler    *handler.UserHandler
	roomHandler    *handler.RoomHandler
	bookingHandler *handler.BookingHandler
}

func NewRouter(config Config, tokenManager pkgAuth.TokenManager, authService applicationAuth.Service, userHandler *handler.UserHandler, roomHandler *handler.RoomHandler, bookingHandler *handler.BookingHandler) *Router {
	// Set Gin mode based on environment or default to release to keep it clean
	env := os.Getenv("APP_ENV")
	if env != "development" && env != "local" && os.Getenv("GIN_MODE") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Global middlewares
	r.Use(
		middleware.RequestID(),
		middleware.Logger(),
		middleware.SecurityHeaders(),
		middleware.RateLimit(50, 100), // 50 req/s, 100 burst
		gin.Recovery(),
	)

	// CORS configuration
	r.Use(cors.New(cors.Config{
		AllowOrigins:     config.AllowOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	return &Router{
		engine:         r,
		config:         config,
		tokenManager:   tokenManager,
		authService:    authService,
		userHandler:    userHandler,
		roomHandler:    roomHandler,
		bookingHandler: bookingHandler,
	}
}

func (r *Router) Engine() *gin.Engine {
	return r.engine
}

func (r *Router) RegisterRoutes(healthHandler *handler.HealthHandler, authHandler *handler.AuthHandler, systemHandler *handler.SystemHandler) http.Handler {
	v1 := r.engine.Group("/v1")
	{
		r.registerV1Routes(v1, healthHandler, authHandler, systemHandler, r.userHandler, r.roomHandler, r.bookingHandler)
	}

	// Redirect root to /v1
	r.engine.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/v1")
	})

	return r.engine
}

func (r *Router) registerV1Routes(rg *gin.RouterGroup, healthHandler *handler.HealthHandler, authHandler *handler.AuthHandler, systemHandler *handler.SystemHandler, userHandler *handler.UserHandler, roomHandler *handler.RoomHandler, bookingHandler *handler.BookingHandler) {
	// Health & Auth (Public)
	rg.GET("/health", healthHandler.HealthCheck)
	rg.POST("/auth/register", authHandler.Register)
	rg.POST("/auth/login", authHandler.Login)
	rg.POST("/auth/refresh", authHandler.Refresh)
	rg.POST("/auth/logout", authHandler.Logout)

	// Protected routes
	authRequired := rg.Group("/")
	authRequired.Use(middleware.AuthMiddleware(r.tokenManager, r.authService))
	{
		// Booking routes
		authRequired.POST("/bookings", bookingHandler.CreateBooking)
		authRequired.GET("/bookings/me", bookingHandler.GetMyBookings)
		authRequired.POST("/bookings/:id/cancel", bookingHandler.CancelBooking)

		// Admin only routes
		adminOnly := authRequired.Group("/admin")
		adminOnly.Use(middleware.RolesAllowed(auth.RoleAdmin))
		{
			adminOnly.GET("/users", userHandler.ListUsers)
			adminOnly.POST("/rooms", roomHandler.CreateRoom)
			adminOnly.PUT("/rooms/:id", roomHandler.UpdateRoom)
			adminOnly.DELETE("/rooms/:id", roomHandler.DeleteRoom)
		}

		// Admin & Staff routes
		staffRoutes := authRequired.Group("/")
		staffRoutes.Use(middleware.RolesAllowed(auth.RoleAdmin, auth.RoleOfficer))
		{
			staffRoutes.PATCH("/rooms/:id/status", roomHandler.UpdateRoomStatus)
			staffRoutes.GET("/bookings", bookingHandler.GetAllBookings)
		}
	}

	// Public Room Routes
	rg.GET("/rooms", roomHandler.ListRooms)
	rg.GET("/rooms/:id", roomHandler.GetRoom)
	rg.GET("/availability", roomHandler.SearchAvailability)

	// Debug routes (Dev only)
	env := os.Getenv("APP_ENV")
	if env == "development" || env == "local" {
		rg.GET("/debug/routes", systemHandler.DebugRoutes)
	}

	// Documentation
	r.registerDocRoutes(rg)

	// Root
	rg.GET("/", systemHandler.Index)
}

func (r *Router) registerDocRoutes(rg *gin.RouterGroup) {
	rg.GET("/swagger/*any", func(c *gin.Context) {
		path := c.Param("any")
		if path == "/doc.json" {
			if _, err := os.Stat(r.config.SwaggerPath); os.IsNotExist(err) {
				slog.Error("swagger file not found", "path", r.config.SwaggerPath)
				api.Error(c, http.StatusNotFound, "NOT_FOUND", "Swagger documentation file not found")
				return
			}
			c.File(r.config.SwaggerPath)
			return
		}
		c.Redirect(http.StatusMovedPermanently, "/v1/doc")
	})

	rg.GET("/doc", func(c *gin.Context) {
		if _, err := os.Stat(r.config.SwaggerPath); os.IsNotExist(err) {
			slog.Error("swagger file not found for scalar", "path", r.config.SwaggerPath)
			api.Error(c, http.StatusNotFound, "NOT_FOUND", "API documentation source not found")
			return
		}
		htmlContent, err := scalar.ApiReferenceHTML(&scalar.Options{
			SpecURL: r.config.SwaggerPath,
			CustomOptions: scalar.CustomOptions{
				PageTitle: "Booking Management API - Scalar",
			},
			DarkMode: true,
		})
		if err != nil {
			slog.Error("failed to generate scalar api reference", "error", err)
			api.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Failed to generate API documentation")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(htmlContent))
	})
}
