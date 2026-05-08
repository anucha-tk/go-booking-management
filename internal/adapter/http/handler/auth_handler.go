package handler

import (
	"errors"
	"go-booking-management-init/internal/adapter/http/dto"
	"go-booking-management-init/internal/application/auth"
	domainAuth "go-booking-management-init/internal/domain/auth"
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

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request payload",
			"code":    "INVALID_REQUEST",
		})
		return
	}

	user, err := h.authService.Register(c.Request.Context(), req.Email, req.Password, req.Role)
	if err != nil {
		switch {
		case errors.Is(err, domainAuth.ErrUserAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{
				"status":  "error",
				"message": "User already exists",
				"code":    "USER_EXISTS",
			})
		case errors.Is(err, auth.ErrInvalidEmail):
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"status":  "error",
				"message": "Invalid email address format",
				"code":    "INVALID_EMAIL",
			})
		case errors.Is(err, auth.ErrInvalidRole):
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "Invalid user role provided",
				"code":    "INVALID_ROLE",
			})
		case errors.Is(err, auth.ErrPasswordTooLong):
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "Password is too long (max 72 characters)",
				"code":    "PASSWORD_TOO_LONG",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Internal server error",
				"code":    "SERVER_ERROR",
			})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"data": dto.RegisterResponse{
			ID:        user.ID,
			Email:     user.Email,
			Role:      user.Role,
			CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
	})
}
