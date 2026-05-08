package handler

import (
	"go-booking-management-init/pkg/api"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SystemHandler struct {
	engine *gin.Engine
}

func NewSystemHandler(engine *gin.Engine) *SystemHandler {
	return &SystemHandler{
		engine: engine,
	}
}

// Index godoc
// @Summary API Root
// @Description get API welcome message
// @Tags system
// @Produce  json
// @Success 200 {object} api.Response
// @Router /v1/ [get]
func (h *SystemHandler) Index(c *gin.Context) {
	api.Success(c, gin.H{
		"message": "Booking Management API",
	})
}

// DebugRoutes godoc
// @Summary Debug API Routes
// @Description get all registered routes (Dev only)
// @Tags debug
// @Produce  json
// @Success 200 {array} interface{}
// @Router /v1/debug/routes [get]
func (h *SystemHandler) DebugRoutes(c *gin.Context) {
	routes := h.engine.Routes()
	type routeDisplay struct {
		Method  string `json:"method"`
		Path    string `json:"path"`
		Handler string `json:"handler"`
	}

	display := make([]routeDisplay, len(routes))
	for i, r := range routes {
		display[i] = routeDisplay{
			Method:  r.Method,
			Path:    r.Path,
			Handler: r.Handler,
		}
	}

	c.JSON(http.StatusOK, display)
}
