package handler

import (
	"go-booking-management-init/internal/adapter/http/dto"
	"go-booking-management-init/internal/adapter/http/middleware"
	bookingApp "go-booking-management-init/internal/application/booking"
	"go-booking-management-init/internal/domain/auth"
	"go-booking-management-init/internal/domain/booking"
	"go-booking-management-init/pkg/api"
	pkgAuth "go-booking-management-init/pkg/auth"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type BookingHandler struct {
	service *bookingApp.Service
}

func NewBookingHandler(service *bookingApp.Service) *BookingHandler {
	return &BookingHandler{service: service}
}

// CreateBooking godoc
// @Summary Submit a booking request
// @Description create a new booking for the authenticated user
// @Tags bookings
// @Accept  json
// @Produce  json
// @Security Bearer
// @Param   request body dto.CreateBookingRequest true "Booking details"
// @Success 201 {object} api.Response{data=dto.BookingResponse}
// @Failure 400 {object} api.Response
// @Failure 401 {object} api.Response
// @Failure 409 {object} api.Response
// @Failure 500 {object} api.Response
// @Router /v1/bookings [post]
func (h *BookingHandler) CreateBooking(c *gin.Context) {
	var req dto.CreateBookingRequest
	if !api.BindAndValidate(c, &req) {
		return
	}

	// Get UserID from context (set by AuthMiddleware)
	identityVal, exists := c.Get(middleware.UserContextKey)
	if !exists {
		api.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}
	claims := identityVal.(*pkgAuth.UserClaims)
	userID := claims.UserID

	params := booking.CreateParams{
		UserID:    userID,
		RoomID:    req.RoomID,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
	}

	created, err := h.service.SubmitBooking(c.Request.Context(), params)
	if err != nil {
		MapError(c, err)
		return
	}

	api.Created(c, dto.ToBookingResponse(created))
}

// GetMyBookings godoc
// @Summary Get current user's bookings
// @Description get a list of bookings for the authenticated user
// @Tags bookings
// @Produce  json
// @Security Bearer
// @Success 200 {object} api.Response{data=[]dto.UserBookingResponse}
// @Router /v1/bookings/me [get]
func (h *BookingHandler) GetMyBookings(c *gin.Context) {
	identityVal, exists := c.Get(middleware.UserContextKey)
	if !exists {
		api.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}
	claims := identityVal.(*pkgAuth.UserClaims)
	userID := claims.UserID

	bookings, err := h.service.GetMyBookings(c.Request.Context(), userID)
	if err != nil {
		MapError(c, err)
		return
	}

	api.Success(c, dto.ToUserBookingResponseList(bookings))
}

// GetAllBookings godoc
// @Summary List all bookings
// @Description get a complete list of all bookings in the system (Admin/Officer only)
// @Tags bookings
// @Produce  json
// @Security Bearer
// @Success 200 {object} api.Response{data=[]dto.AdminBookingResponse}
// @Router /v1/bookings [get]
func (h *BookingHandler) GetAllBookings(c *gin.Context) {
	bookings, err := h.service.GetAllBookings(c.Request.Context())
	if err != nil {
		MapError(c, err)
		return
	}

	api.Success(c, dto.ToAdminBookingResponseList(bookings))
}

// CancelBooking godoc
// @Summary Cancel a booking
// @Description cancel an existing booking for the authenticated user
// @Tags bookings
// @Produce  json
// @Security Bearer
// @Param   id path int true "Booking ID"
// @Success 200 {object} api.Response{data=dto.BookingResponse}
// @Router /v1/bookings/{id}/cancel [post]
func (h *BookingHandler) CancelBooking(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		api.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid booking ID")
		return
	}

	identityVal, exists := c.Get(middleware.UserContextKey)
	if !exists {
		api.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}
	claims := identityVal.(*pkgAuth.UserClaims)
	isAdmin := claims.Role == string(auth.RoleAdmin) || claims.Role == string(auth.RoleOfficer)

	cancelled, err := h.service.CancelBooking(c.Request.Context(), claims.UserID, isAdmin, int32(id))
	if err != nil {
		MapError(c, err)
		return
	}

	api.Success(c, dto.ToBookingResponse(cancelled))
}
