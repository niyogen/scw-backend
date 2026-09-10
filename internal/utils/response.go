package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// StandardResponse standardizes API response payload
type StandardResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// PaginatedResponse wraps list results with pagination metadata
type PaginatedResponse struct {
	Items      interface{} `json:"items"`
	TotalCount int64       `json:"total_count"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
}

func JSONSuccess(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, StandardResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func JSONError(c *gin.Context, statusCode int, message string, errDetail interface{}) {
	c.JSON(statusCode, StandardResponse{
		Success: false,
		Message: message,
		Error:   errDetail,
	})
}

func JSONBadRequest(c *gin.Context, message string, errDetail interface{}) {
	JSONError(c, http.StatusBadRequest, message, errDetail)
}

func JSONUnauthorized(c *gin.Context, message string) {
	JSONError(c, http.StatusUnauthorized, message, nil)
}

func JSONForbidden(c *gin.Context, message string) {
	JSONError(c, http.StatusForbidden, message, nil)
}

func JSONNotFound(c *gin.Context, message string) {
	JSONError(c, http.StatusNotFound, message, nil)
}

func JSONInternalServerError(c *gin.Context, message string, errDetail interface{}) {
	JSONError(c, http.StatusInternalServerError, message, errDetail)
}
