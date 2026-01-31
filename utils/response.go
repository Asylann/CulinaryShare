package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// SuccessResponse represents a standardized success response
type SuccessResponse struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Status     int         `json:"status"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	TotalCount int64       `json:"totalCount"`
	TotalPages int         `json:"totalPages"`
}

// RespondWithError sends a standardized error response
func RespondWithError(c *gin.Context, status int, message string, err error) {
	response := ErrorResponse{
		Status:  status,
		Message: message,
	}
	if err != nil {
		response.Error = err.Error()
	}
	c.JSON(status, response)
}

// RespondWithSuccess sends a standardized success response
func RespondWithSuccess(c *gin.Context, status int, message string, data interface{}) {
	response := SuccessResponse{
		Status:  status,
		Message: message,
		Data:    data,
	}
	c.JSON(status, response)
}

// RespondWithPagination sends a paginated response
func RespondWithPagination(c *gin.Context, data interface{}, page, limit int, totalCount int64) {
	totalPages := int(totalCount) / limit
	if int(totalCount)%limit != 0 {
		totalPages++
	}

	response := PaginatedResponse{
		Status:     http.StatusOK,
		Message:    "Success",
		Data:       data,
		Page:       page,
		Limit:      limit,
		TotalCount: totalCount,
		TotalPages: totalPages,
	}
	c.JSON(http.StatusOK, response)
}

// Common error responses
func BadRequest(c *gin.Context, message string, err error) {
	RespondWithError(c, http.StatusBadRequest, message, err)
}

func Unauthorized(c *gin.Context, message string) {
	RespondWithError(c, http.StatusUnauthorized, message, nil)
}

func Forbidden(c *gin.Context, message string) {
	RespondWithError(c, http.StatusForbidden, message, nil)
}

func NotFound(c *gin.Context, message string) {
	RespondWithError(c, http.StatusNotFound, message, nil)
}

func InternalServerError(c *gin.Context, message string, err error) {
	RespondWithError(c, http.StatusInternalServerError, message, err)
}

func Conflict(c *gin.Context, message string) {
	RespondWithError(c, http.StatusConflict, message, nil)
}

func Created(c *gin.Context, message string, data interface{}) {
	RespondWithSuccess(c, http.StatusCreated, message, data)
}

func OK(c *gin.Context, message string, data interface{}) {
	RespondWithSuccess(c, http.StatusOK, message, data)
}
