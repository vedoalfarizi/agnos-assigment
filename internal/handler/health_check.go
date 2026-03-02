package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/vedoalfarizi/hospital-api/internal/logger"
	"github.com/vedoalfarizi/hospital-api/internal/service"
)

type HealthCheckResponse struct {
	Status   string `json:"status"`   // "healthy" or "unhealthy"
	Server   string `json:"server"`   // "up"/"down"
	Database string `json:"database"` // "up"/"down"
}

// HealthCheck returns a Gin handler that uses the provided service.
// creating the handler via a constructor makes it easy to supply mocks during
// unit testing.
func HealthCheck(svc *service.HealthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		serverStatus := "up"
		databaseStatus := "up"
		overall := "healthy"

		if err := svc.Check(); err != nil {
			// database ping failed, mark unhealthy and log error for engineers
			databaseStatus = "down"
			overall = "unhealthy"
			logger.ErrorfWithContext(c.Request.Context(), "database health check failed: error=%v", err)
		}

		resp := HealthCheckResponse{
			Status:   overall,
			Server:   serverStatus,
			Database: databaseStatus,
		}
		if overall == "healthy" {
			c.JSON(200, resp)
		} else {
			// use 503 for unhealthy state and return component details directly
			c.JSON(503, resp)
		}
	}
}
