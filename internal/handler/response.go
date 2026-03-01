package handler

import (
	"github.com/gin-gonic/gin"
)

type ErrorResp struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorResp  `json:"error,omitempty"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(200, APIResponse{
		Success: true,
		Data:    data,
	})
}

func SuccessWithStatus(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, APIResponse{
		Success: true,
		Data:    data,
	})
}

func Error(c *gin.Context, statusCode int, code, message string) {
	c.JSON(statusCode, APIResponse{
		Success: false,
		Error: &ErrorResp{
			Code:    code,
			Message: message,
		},
	})
}
