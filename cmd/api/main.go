package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	adapterHttp "go-booking-management-init/internal/adapter/http"
	"go-booking-management-init/internal/server"
	"go-booking-management-init/pkg/api"
	"go-booking-management-init/pkg/logger"
	"log/slog"
	"os"
)

// @title Booking Management API
// @version 1.0
// @description This is a booking management server.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8088
// @BasePath /

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description MANDATORY: Type "Bearer " followed by a space and your JWT token. Example: "Bearer eyJhbGci..."

func gracefulShutdown(apiServer *http.Server, done chan bool) {
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Listen for the interrupt signal.
	<-ctx.Done()

	log.Println("shutting down gracefully, press Ctrl+C again to force")
	stop() // Allow Ctrl+C to force shutdown

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := apiServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown with error: %v", err)
	}

	log.Println("Server exiting")

	// Notify the main goroutine that the shutdown is complete
	done <- true
}

func main() {
	listRoutes := false
	exportPath := ""
	for i, arg := range os.Args {
		if arg == "--list-routes" {
			listRoutes = true
		}
		if arg == "--export-routes" && i+1 < len(os.Args) {
			exportPath = os.Args[i+1]
		}
	}

	logger.Init(os.Getenv("APP_ENV"))
	api.InitValidator()

	httpServer, router := server.NewServer()

	if listRoutes {
		adapterHttp.InspectRoutes(router.Engine().Routes(), "")
		os.Exit(0)
	}

	if exportPath != "" {
		err := adapterHttp.ExportRoutes(router.Engine().Routes(), exportPath)
		if err != nil {
			slog.Error("failed to export routes", "error", err)
			os.Exit(1)
		}
		fmt.Printf("✅ Routes exported to %s\n", exportPath)
		os.Exit(0)
	}

	slog.Info("Starting server", "port", os.Getenv("PORT"), "env", os.Getenv("APP_ENV"))

	// Create a done channel to signal when the shutdown is complete
	done := make(chan bool, 1)

	// Run graceful shutdown in a separate goroutine
	go gracefulShutdown(httpServer, done)

	err := httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		panic(fmt.Sprintf("http server error: %s", err))
	}

	// Wait for the graceful shutdown to complete
	<-done
	log.Println("Graceful shutdown complete.")
}
