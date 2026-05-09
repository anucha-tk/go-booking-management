package handler

import (
	"go-booking-management-init/pkg/api"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
}

func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

// ListUsers godoc
// @Summary List all users (Admin only)
// @Description get list of all registered users
// @Tags admin
// @Security Bearer
// @Produce  json
// @Success 200 {object} api.Response
// @Router /v1/admin/users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	// AC Requirement: Restricted endpoint example
	api.Success(c, gin.H{
		"message": "Welcome, Admin! User list will be here.",
		"users":   []interface{}{},
	})
}
