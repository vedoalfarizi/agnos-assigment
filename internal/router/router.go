package router

import (
	"github.com/gin-gonic/gin"

	"github.com/vedoalfarizi/hospital-api/internal/config"
	"github.com/vedoalfarizi/hospital-api/internal/handler"
	"github.com/vedoalfarizi/hospital-api/internal/logger"
)

func New(log *logger.Logger, cfg *config.Config) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.Default()

	v1 := r.Group("/api/v1")
	{
		v1.GET("/status", handler.HealthCheck)
	}

	return r
}
