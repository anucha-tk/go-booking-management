package handler

import (
	"go-booking-management-init/internal/adapter/http/dto"
	roomApp "go-booking-management-init/internal/application/room"
	"go-booking-management-init/internal/domain/room"
	"go-booking-management-init/pkg/api"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type RoomHandler struct {
	service *roomApp.Service
}

func NewRoomHandler(service *roomApp.Service) *RoomHandler {
	return &RoomHandler{service: service}
}

// CreateRoom godoc
// @Summary Create a new room
// @Description add a new room to the system (Admin only)
// @Tags rooms
// @Accept  json
// @Produce  json
// @Security Bearer
// @Param   request body dto.CreateRoomRequest true "Room details"
// @Success 201 {object} api.Response{data=dto.RoomResponse}
// @Failure 400 {object} api.Response
// @Failure 403 {object} api.Response
// @Failure 500 {object} api.Response
// @Router /v1/rooms [post]
func (h *RoomHandler) CreateRoom(c *gin.Context) {
	var req dto.CreateRoomRequest
	if !api.BindAndValidate(c, &req) {
		return
	}

	rm := &room.Room{
		RoomNumber: req.RoomNumber,
		Type:       req.Type,
		Price:      req.Price,
		Status:     "available",
	}

	created, err := h.service.CreateRoom(c.Request.Context(), rm)
	if err != nil {
		MapError(c, err)
		return
	}

	api.Created(c, dto.ToRoomResponse(created))
}

// GetRoom godoc
// @Summary Get room details
// @Description get full details of a specific room including booking history
// @Tags rooms
// @Produce  json
// @Param   id path int true "Room ID"
// @Success 200 {object} api.Response{data=room.Detail}
// @Failure 404 {object} api.Response
// @Failure 500 {object} api.Response
// @Router /v1/rooms/{id} [get]
func (h *RoomHandler) GetRoom(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		api.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid room ID")
		return
	}

	detail, err := h.service.GetRoomDetail(c.Request.Context(), int32(id))
	if err != nil {
		MapError(c, err)
		return
	}

	api.Success(c, dto.ToRoomDetailResponse(detail))
}

// ListRooms godoc
// @Summary List all rooms
// @Description get a list of all rooms with optional filtering
// @Tags rooms
// @Produce  json
// @Param   type query string false "Filter by room type"
// @Param   price_min query int false "Filter by minimum price"
// @Param   price_max query int false "Filter by maximum price"
// @Success 200 {object} api.Response{data=[]dto.RoomResponse}
// @Failure 500 {object} api.Response
// @Router /v1/rooms [get]
func (h *RoomHandler) ListRooms(c *gin.Context) {
	var query dto.ListRoomsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		api.Error(c, http.StatusBadRequest, "INVALID_QUERY", err.Error())
		return
	}

	if query.Limit == 0 {
		query.Limit = 10
	}
	filter := room.Filter{
		Type:     query.Type,
		MinPrice: query.MinPrice,
		MaxPrice: query.MaxPrice,
		Limit:    query.Limit,
		Offset:   query.Offset,
	}

	rooms, err := h.service.ListRooms(c.Request.Context(), filter)
	if err != nil {
		MapError(c, err)
		return
	}

	api.Success(c, dto.ToRoomResponseList(rooms))
}

// UpdateRoom godoc
// @Summary Update room details
// @Description update an existing room (Admin only)
// @Tags rooms
// @Accept  json
// @Produce  json
// @Security Bearer
// @Param   id path int true "Room ID"
// @Param   request body dto.UpdateRoomRequest true "Room details"
// @Success 200 {object} api.Response{data=dto.RoomResponse}
// @Failure 400 {object} api.Response
// @Failure 403 {object} api.Response
// @Failure 404 {object} api.Response
// @Failure 500 {object} api.Response
// @Router /v1/rooms/{id} [put]
func (h *RoomHandler) UpdateRoom(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		api.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid room ID")
		return
	}

	var req dto.UpdateRoomRequest
	if !api.BindAndValidate(c, &req) {
		return
	}

	rm := &room.Room{
		ID:         int32(id),
		RoomNumber: req.RoomNumber,
		Type:       req.Type,
		Price:      req.Price,
		Status:     req.Status,
	}

	updated, err := h.service.UpdateRoom(c.Request.Context(), rm)
	if err != nil {
		MapError(c, err)
		return
	}

	api.Success(c, dto.ToRoomResponse(updated))
}

// DeleteRoom godoc
// @Summary Delete a room
// @Description remove a room from the system (Admin only)
// @Tags rooms
// @Produce  json
// @Security Bearer
// @Param   id path int true "Room ID"
// @Success 200 {object} api.Response
// @Failure 403 {object} api.Response
// @Failure 404 {object} api.Response
// @Failure 500 {object} api.Response
// @Router /v1/rooms/{id} [delete]
func (h *RoomHandler) DeleteRoom(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		api.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid room ID")
		return
	}

	err = h.service.DeleteRoom(c.Request.Context(), int32(id))
	if err != nil {
		MapError(c, err)
		return
	}

	api.Success(c, nil)
}

// UpdateRoomStatus godoc
// @Summary Update room status
// @Description update the status of a specific room (Admin/Officer)
// @Tags rooms
// @Accept  json
// @Produce  json
// @Security Bearer
// @Param   id path int true "Room ID"
// @Param   request body dto.UpdateRoomStatusRequest true "Status details"
// @Success 200 {object} api.Response{data=dto.RoomResponse}
// @Router /v1/rooms/{id}/status [patch]
func (h *RoomHandler) UpdateRoomStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		api.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid room ID")
		return
	}

	var req dto.UpdateRoomStatusRequest
	if !api.BindAndValidate(c, &req) {
		return
	}

	updated, err := h.service.UpdateRoomStatus(c.Request.Context(), int32(id), req.Status)
	if err != nil {
		MapError(c, err)
		return
	}

	api.Success(c, dto.ToRoomResponse(updated))
}

// SearchAvailability godoc
// @Summary Search available rooms
// @Description get a list of available rooms for specific dates
// @Tags availability
// @Produce  json
// @Param   start_date query string true "Start date (YYYY-MM-DD)"
// @Param   end_date query string true "End date (YYYY-MM-DD)"
// @Param   type query string false "Filter by room type"
// @Param   price_min query int false "Filter by minimum price"
// @Param   price_max query int false "Filter by maximum price"
// @Success 200 {object} api.Response{data=[]dto.RoomResponse}
// @Failure 400 {object} api.Response
// @Failure 500 {object} api.Response
// @Router /v1/availability [get]
func (h *RoomHandler) SearchAvailability(c *gin.Context) {
	var query dto.AvailabilityQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		api.Error(c, http.StatusBadRequest, "INVALID_QUERY", err.Error())
		return
	}

	if query.Limit == 0 {
		query.Limit = 10
	}
	filter := room.AvailabilityFilter{
		StartDate: query.StartDate,
		EndDate:   query.EndDate,
		Type:      query.Type,
		MinPrice:  query.MinPrice,
		MaxPrice:  query.MaxPrice,
		Limit:     query.Limit,
		Offset:    query.Offset,
	}

	rooms, err := h.service.ListAvailableRooms(c.Request.Context(), filter)
	if err != nil {
		MapError(c, err)
		return
	}

	api.Success(c, dto.ToRoomResponseList(rooms))
}
