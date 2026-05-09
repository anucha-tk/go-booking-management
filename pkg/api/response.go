package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response represents a standard JSON response structure.
type Response struct {
	Status    string      `json:"status" example:"success"`
	Data      interface{} `json:"data,omitempty"`
	Message   string      `json:"message,omitempty"`
	Code      string      `json:"code,omitempty"`
	Errors    interface{} `json:"errors,omitempty"`
	RequestID string      `json:"requestId,omitempty"`
}

// UnauthorizedResponse represents a 401 error response.
// @Description Unauthorized error response
type UnauthorizedResponse struct {
	Status    string `json:"status" example:"error"`
	Message   string `json:"message" example:"Authorization header is required"`
	Code      string `json:"code" example:"UNAUTHORIZED"`
	RequestID string `json:"requestId" example:"66c6aa26-d246-44dd-8ecd-a39e15973d82"`
}

// UnauthorizedFormatResponse represents a 401 error response for invalid format.
// @Description Unauthorized error response (Invalid Format)
type UnauthorizedFormatResponse struct {
	Status    string `json:"status" example:"error"`
	Message   string `json:"message" example:"Authorization header format must be Bearer <token>"`
	Code      string `json:"code" example:"UNAUTHORIZED"`
	RequestID string `json:"requestId" example:"4af3b007-7f7b-44bc-800c-12df8ae15a47"`
}

// Success sends a success response with 200 OK.
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Status:    "success",
		Data:      data,
		RequestID: c.GetString("request_id"),
	})
}

// Created sends a success response with 201 Created.
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Status:    "success",
		Data:      data,
		RequestID: c.GetString("request_id"),
	})
}

// Error sends an error response with the given status, code and message.
func Error(c *gin.Context, status int, code, message string) {
	c.JSON(status, Response{
		Status:    "error",
		Code:      code,
		Message:   message,
		RequestID: c.GetString("request_id"),
	})
}

// ValidationError sends a 400 Bad Request response with validation details.
func ValidationError(c *gin.Context, errors interface{}) {
	c.JSON(http.StatusBadRequest, Response{
		Status:    "error",
		Code:      "VALIDATION_ERROR",
		Message:   "Validation failed",
		Errors:    errors,
		RequestID: c.GetString("request_id"),
	})
}
