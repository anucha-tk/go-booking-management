package http

import (
	"log/slog"
	"net/http"
	"os"

	"go-booking-management-init/internal/adapter/http/handler"
	"go-booking-management-init/internal/adapter/http/middleware"
	"go-booking-management-init/pkg/api"

	scalar "github.com/MarceloPetrucio/go-scalar-api-reference"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Config struct {
	AllowOrigins []string
	SwaggerPath  string
}

type Router struct {
	engine *gin.Engine
	config Config
}

func NewRouter(config Config) *Router {
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
		engine: r,
		config: config,
	}
}

func (r *Router) RegisterRoutes(healthHandler *handler.HealthHandler, authHandler *handler.AuthHandler) http.Handler {
	v1 := r.engine.Group("/v1")
	{
		v1.GET("/health", healthHandler.HealthCheck)
		v1.POST("/auth/register", authHandler.Register)

		// Documentation routes
		v1.GET("/swagger/*any", func(c *gin.Context) {
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

		v1.GET("/doc", func(c *gin.Context) {
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

		// Root route
		v1.GET("/", func(c *gin.Context) {
			api.Success(c, gin.H{
				"message": "Booking Management API",
			})
		})
	}

	// Redirect root to /v1
	r.engine.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/v1")
	})

	return r.engine
}
