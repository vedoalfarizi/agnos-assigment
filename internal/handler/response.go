package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/vedoalfarizi/hospital-api/internal/logger"
)

type ErrorResp struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type APIResponse struct {
	RequestID string      `json:"request_id"`
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     *ErrorResp  `json:"error,omitempty"`
}

// extractRequestID retrieves request_id from context
func extractRequestID(c *gin.Context) string {
	if requestID, ok := c.Request.Context().Value(logger.RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(200, APIResponse{
		RequestID: extractRequestID(c),
		Success:   true,
		Data:      data,
	})
}

func SuccessWithStatus(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, APIResponse{
		RequestID: extractRequestID(c),
		Success:   true,
		Data:      data,
	})
}

func Error(c *gin.Context, statusCode int, code, message string) {
	c.JSON(statusCode, APIResponse{
		RequestID: extractRequestID(c),
		Success:   false,
		Error: &ErrorResp{
			Code:    code,
			Message: message,
		},
	})
}
