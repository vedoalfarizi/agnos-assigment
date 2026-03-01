package handler

import (
	"github.com/gin-gonic/gin"
)

type HealthCheckResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func HealthCheck(c *gin.Context) {
	Success(c, HealthCheckResponse{
		Status:  "healthy",
		Message: "Server is running and ready to accept requests",
	})
}
