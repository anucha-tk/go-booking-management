package handler

import (
	"go-booking-management-init/internal/adapter/http/dto"
	"go-booking-management-init/internal/application/auth"
	"go-booking-management-init/pkg/api"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService auth.Service
}

func NewAuthHandler(authService auth.Service) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if !api.BindAndValidate(c, &req) {
		return
	}

	user, err := h.authService.Register(c.Request.Context(), req.Email, req.Password, req.Role)
	if err != nil {
		MapError(c, err)
		return
	}

	api.Created(c, dto.RegisterResponse{
		ID:        user.ID,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}
