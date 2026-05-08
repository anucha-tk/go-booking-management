package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response represents a standard JSON response structure.
type Response struct {
	Status  string      `json:"status" example:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Code    string      `json:"code,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

// Success sends a success response with 200 OK.
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Status: "success",
		Data:   data,
	})
}

// Created sends a success response with 201 Created.
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Status: "success",
		Data:   data,
	})
}

// Error sends an error response with the given status, code and message.
func Error(c *gin.Context, status int, code, message string) {
	c.JSON(status, Response{
		Status:  "error",
		Code:    code,
		Message: message,
	})
}

// ValidationError sends a 400 Bad Request response with validation details.
func ValidationError(c *gin.Context, errors interface{}) {
	c.JSON(http.StatusBadRequest, Response{
		Status:  "error",
		Code:    "VALIDATION_ERROR",
		Message: "Validation failed",
		Errors:  errors,
	})
}
