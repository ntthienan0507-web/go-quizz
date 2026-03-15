package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response is a generic API response with typed data.
type Response[T any] struct {
	Status int `json:"status"`
	Data   T   `json:"data"`
}

// ErrorBody represents an error response payload.
type ErrorBody struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

// ListData is a generic wrapper for list responses with a total count.
type ListData[T any] struct {
	Items []T `json:"items"`
	Total int `json:"total"`
}

// MessageData is a simple message payload.
type MessageData struct {
	Message string `json:"message"`
}

// --- Success Helpers ---

// OK sends a 200 response with typed data.
func OK[T any](c *gin.Context, data T) {
	c.JSON(http.StatusOK, Response[T]{
		Status: http.StatusOK,
		Data:   data,
	})
}

// Created sends a 201 response with typed data.
func Created[T any](c *gin.Context, data T) {
	c.JSON(http.StatusCreated, Response[T]{
		Status: http.StatusCreated,
		Data:   data,
	})
}

// OKList sends a 200 response with a list and total count.
func OKList[T any](c *gin.Context, items []T, total int) {
	c.JSON(http.StatusOK, Response[ListData[T]]{
		Status: http.StatusOK,
		Data: ListData[T]{
			Items: items,
			Total: total,
		},
	})
}

// OKMessage sends a 200 response with a simple message.
func OKMessage(c *gin.Context, message string) {
	c.JSON(http.StatusOK, Response[MessageData]{
		Status: http.StatusOK,
		Data:   MessageData{Message: message},
	})
}

// --- Error Helpers ---

func errorJSON(c *gin.Context, status int, message string) {
	c.JSON(status, ErrorBody{
		Status:  status,
		Message: message,
	})
}

func BadRequest(c *gin.Context, message string) {
	errorJSON(c, http.StatusBadRequest, message)
}

func Unauthorized(c *gin.Context, message string) {
	errorJSON(c, http.StatusUnauthorized, message)
}

func Forbidden(c *gin.Context, message string) {
	errorJSON(c, http.StatusForbidden, message)
}

func NotFound(c *gin.Context, message string) {
	errorJSON(c, http.StatusNotFound, message)
}

func Conflict(c *gin.Context, message string) {
	errorJSON(c, http.StatusConflict, message)
}

func InternalError(c *gin.Context, message string) {
	errorJSON(c, http.StatusInternalServerError, message)
}
