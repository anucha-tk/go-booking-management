package server

import (
	_ "go-booking-management-init/api" // Swagger documentation
	"go-booking-management-init/internal/adapter/http/middleware"
	"log/slog"
	"net/http"

	scalar "github.com/MarceloPetrucio/go-scalar-api-reference"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// @title Booking Management API
// @version 1.0
// @description This is a booking management server.
// @host localhost:8080
// @BasePath /
func (s *Server) RegisterRoutes() http.Handler {
	slog.Info("Registering routes...")
	r := gin.New()
	r.Use(middleware.Logger(), gin.Recovery())

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"}, // Add your frontend URL
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true, // Enable cookies/auth
	}))

	r.GET("/", s.HelloWorldHandler)

	r.GET("/health", s.healthHandler)

	r.GET("/swagger/*any", func(c *gin.Context) {
		path := c.Param("any")
		if path == "/doc.json" {
			c.File("./api/swagger.json")
			return
		}
		c.Redirect(http.StatusMovedPermanently, "/doc")
	})

	r.GET("/doc", func(c *gin.Context) {
		htmlContent, err := scalar.ApiReferenceHTML(&scalar.Options{
			SpecURL: "./api/swagger.json",
			CustomOptions: scalar.CustomOptions{
				PageTitle: "Booking Management API - Scalar",
			},
			DarkMode: true,
		})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(htmlContent))
	})

	return r
}

// HelloWorldHandler godoc
// @Summary Hello World
// @Description get string hello world
// @Tags root
// @Accept  json
// @Produce  json
// @Success 200 {object} map[string]string
// @Router / [get]
func (s *Server) HelloWorldHandler(c *gin.Context) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"

	c.JSON(http.StatusOK, resp)
}

// healthHandler godoc
// @Summary Health Check
// @Description get database health status
// @Tags root
// @Accept  json
// @Produce  json
// @Success 200 {object} map[string]string
// @Router /health [get]
func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s.db.Health())
}
