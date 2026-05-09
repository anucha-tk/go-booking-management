package handler

import (
	"go-booking-management-init/internal/adapter/http/dto"
	"go-booking-management-init/internal/application/auth"
	domain "go-booking-management-init/internal/domain/auth"
	"go-booking-management-init/pkg/api"
	"log/slog"
	"net/http"

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

// Login godoc
// @Summary Login a user
// @Description login with email and password to receive access and refresh tokens
// @Tags auth
// @Accept  json
// @Produce  json
// @Param   request body dto.LoginRequest true "Login credentials"
// @Success 200 {object} api.Response{data=dto.LoginResponse}
// @Failure 400 {object} api.Response
// @Failure 401 {object} api.UnauthorizedResponse
// @Failure 500 {object} api.Response
// @Router /v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if !api.BindAndValidate(c, &req) {
		return
	}

	accessToken, refreshToken, err := h.authService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		MapError(c, err)
		return
	}

	api.Success(c, dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

// Refresh godoc
// @Summary Refresh access token
// @Description get a new access token using a valid refresh token
// @Tags auth
// @Accept  json
// @Produce  json
// @Param   request body dto.RefreshRequest true "Refresh token"
// @Success 200 {object} api.Response{data=dto.LoginResponse}
// @Failure 400 {object} api.Response
// @Failure 401 {object} api.UnauthorizedResponse
// @Failure 500 {object} api.Response
// @Router /v1/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var input dto.RefreshRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		slog.WarnContext(c.Request.Context(), "invalid refresh request body", "err", err)
		api.Error(c, http.StatusBadRequest, "INVALID_REQUEST_BODY", "Invalid request body")
		return
	}

	accessToken, refreshToken, err := h.authService.RefreshToken(c.Request.Context(), input.RefreshToken)
	if err != nil {
		MapError(c, err)
		return
	}

	api.Success(c, dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}
