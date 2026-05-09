package http

import (
	"log/slog"
	"net/http"
	"os"

	"go-booking-management-init/internal/adapter/http/handler"
	"go-booking-management-init/internal/adapter/http/middleware"
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
	engine       *gin.Engine
	config       Config
	tokenManager pkgAuth.TokenManager
	userHandler  *handler.UserHandler
}

func NewRouter(config Config, tokenManager pkgAuth.TokenManager, userHandler *handler.UserHandler) *Router {
	// Set Gin mode based on environment or default to release to keep it clean
	env := os.Getenv("APP_ENV")
	if env != "development" && env != "local" && os.Getenv("GIN_MODE") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Default middlewares
	r.Use(middleware.Logger(), gin.Recovery())

	// CORS configuration
	r.Use(cors.New(cors.Config{
		AllowOrigins:     config.AllowOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	return &Router{
		engine:       r,
		config:       config,
		tokenManager: tokenManager,
		userHandler:  userHandler,
	}
}

func (r *Router) Engine() *gin.Engine {
	return r.engine
}

func (r *Router) RegisterRoutes(healthHandler *handler.HealthHandler, authHandler *handler.AuthHandler, systemHandler *handler.SystemHandler) http.Handler {
	v1 := r.engine.Group("/v1")
	{
		r.registerV1Routes(v1, healthHandler, authHandler, systemHandler, r.userHandler)
	}

	// Redirect root to /v1
	r.engine.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/v1")
	})

	return r.engine
}

func (r *Router) registerV1Routes(rg *gin.RouterGroup, healthHandler *handler.HealthHandler, authHandler *handler.AuthHandler, systemHandler *handler.SystemHandler, userHandler *handler.UserHandler) {
	// Health & Auth (Public)
	rg.GET("/health", healthHandler.HealthCheck)
	rg.POST("/auth/register", authHandler.Register)
	rg.POST("/auth/login", authHandler.Login)
	rg.POST("/auth/refresh", authHandler.Refresh)

	// Protected routes
	authRequired := rg.Group("/")
	authRequired.Use(middleware.AuthMiddleware(r.tokenManager))
	{
		// Add future protected routes here
		// Example: authRequired.GET("/profile", userHandler.Profile)

		// Admin only routes
		adminOnly := authRequired.Group("/admin")
		adminOnly.Use(middleware.RolesAllowed(auth.RoleAdmin))
		{
			adminOnly.GET("/users", userHandler.ListUsers)
		}
	}

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
