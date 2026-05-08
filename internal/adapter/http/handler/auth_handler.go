package handler

import (
	"go-booking-management-init/internal/adapter/http/dto"
	"go-booking-management-init/internal/application/auth"
	domain "go-booking-management-init/internal/domain/auth"
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

// Register godoc
// @Summary Register a new user
// @Description create a new user account with email, password, and role
// @Tags auth
// @Accept  json
// @Produce  json
// @Param   request body dto.RegisterRequest true "Registration details"
// @Success 201 {object} api.Response{data=dto.RegisterResponse}
// @Failure 400 {object} api.Response
// @Failure 409 {object} api.Response
// @Failure 500 {object} api.Response
// @Router /v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {

	var req dto.RegisterRequest
	if !api.BindAndValidate(c, &req) {
		return
	}

	user, err := h.authService.Register(c.Request.Context(), req.Email, req.Password, domain.UserRole(req.Role))
	if err != nil {
		MapError(c, err)
		return
	}

	api.Created(c, dto.RegisterResponse{
		ID:        user.ID,
		Email:     user.Email,
		Role:      string(user.Role),
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}
