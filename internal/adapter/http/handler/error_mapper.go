package handler

import (
	"errors"
	"net/http"

	appAuth "go-booking-management-init/internal/application/auth"
	domainAuth "go-booking-management-init/internal/domain/auth"
	domainRoom "go-booking-management-init/internal/domain/room"
	"go-booking-management-init/pkg/api"

	"github.com/gin-gonic/gin"
)

// MapError maps common application/domain errors to HTTP responses using pkg/api.
func MapError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	// Handle validation errors that might bubble up from the service layer
	verrs := api.FormatValidationError(err)
	if verrs != nil {
		api.ValidationError(c, verrs)
		return
	}

	switch {
	case errors.Is(err, domainAuth.ErrUserAlreadyExists):
		api.Error(c, http.StatusConflict, "USER_EXISTS", "User already exists")
	case errors.Is(err, appAuth.ErrInvalidEmail):
		api.Error(c, http.StatusUnprocessableEntity, "INVALID_EMAIL", "Invalid email address format")
	case errors.Is(err, appAuth.ErrInvalidRole):
		api.Error(c, http.StatusBadRequest, "INVALID_ROLE", "Invalid user role provided")
	case errors.Is(err, appAuth.ErrPasswordTooLong):
		api.Error(c, http.StatusBadRequest, "PASSWORD_TOO_LONG", "Password is too long (max 72 characters)")
	case errors.Is(err, domainAuth.ErrUserNotFound):
		api.Error(c, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
	case errors.Is(err, appAuth.ErrInvalidCredentials):
		api.Error(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password")
	case errors.Is(err, appAuth.ErrInvalidRefreshToken):
		api.Error(c, http.StatusUnauthorized, "INVALID_REFRESH_TOKEN", "Invalid or expired refresh token")
	case errors.Is(err, domainRoom.ErrRoomNotFound):
		api.Error(c, http.StatusNotFound, "ROOM_NOT_FOUND", "Room not found")
	case errors.Is(err, domainRoom.ErrRoomNumberExists):
		api.Error(c, http.StatusConflict, "ROOM_EXISTS", "Room number already exists")
	case errors.Is(err, domainRoom.ErrInvalidRoomStatus),
		errors.Is(err, domainRoom.ErrInvalidDateRange),
		errors.Is(err, domainRoom.ErrPastDate),
		errors.Is(err, domainRoom.ErrMissingDates):
		api.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	default:

		api.Error(c, http.StatusInternalServerError, "SERVER_ERROR", "Internal server error")
	}
}
