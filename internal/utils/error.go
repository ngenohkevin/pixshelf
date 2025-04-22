package utils

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorResponse defines the structure for error responses
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// RespondWithError sends an error response to the client
func RespondWithError(c *gin.Context, code int, err error, message string) {
	c.JSON(code, ErrorResponse{
		Error:   err.Error(),
		Message: message,
		Code:    code,
	})
}

// NotFound responds with a 404 error
func NotFound(c *gin.Context, resource string, id interface{}) {
	message := fmt.Sprintf("%s with ID %v not found", resource, id)
	c.JSON(http.StatusNotFound, ErrorResponse{
		Error:   "not_found",
		Message: message,
		Code:    http.StatusNotFound,
	})
}

// BadRequest responds with a 400 error
func BadRequest(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, ErrorResponse{
		Error:   "bad_request",
		Message: err.Error(),
		Code:    http.StatusBadRequest,
	})
}

// InternalServerError responds with a 500 error
func InternalServerError(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Error:   "internal_server_error",
		Message: "An unexpected error occurred",
		Code:    http.StatusInternalServerError,
	})
}
