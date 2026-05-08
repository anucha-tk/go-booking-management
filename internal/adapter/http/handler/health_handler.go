package handler

import (
	"net/http"

	"go-booking-management-init/internal/database"

	"github.com/gin-gonic/gin"
)

const APIVersion = "1.0.0"

type HealthHandler struct {
	db database.Service
}

func NewHealthHandler(db database.Service) *HealthHandler {
	return &HealthHandler{db: db}
}

type HealthResponse struct {
	Status string     `json:"status" example:"success"`
	Data   HealthData `json:"data"`
}

type HealthData struct {
	Status   string            `json:"status" example:"OK"`
	Version  string            `json:"version" example:"1.0.0"`
	Database map[string]string `json:"database,omitempty"`
}

// HealthCheck godoc
// @Summary Health Check
// @Description get API health status including database connectivity
// @Tags health
// @Accept  json
// @Produce  json
// @Success 200 {object} HealthResponse
// @Success 503 {object} HealthResponse
// @Router /v1/health [get]
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	dbHealth := h.db.Health()
	status := http.StatusOK

	overallStatus := "OK"
	if dbHealth["status"] != "up" {
		overallStatus = "Service Unavailable"
		status = http.StatusServiceUnavailable
	}

	c.JSON(status, HealthResponse{
		Status: "success",
		Data: HealthData{
			Status:   overallStatus,
			Version:  APIVersion,
			Database: dbHealth,
		},
	})
}
