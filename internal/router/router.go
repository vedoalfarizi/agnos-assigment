package router

import (
	"github.com/gin-gonic/gin"

	"github.com/vedoalfarizi/hospital-api/internal/config"
	"github.com/vedoalfarizi/hospital-api/internal/handler"
	"github.com/vedoalfarizi/hospital-api/internal/logger"
	"github.com/vedoalfarizi/hospital-api/internal/repository"
	"github.com/vedoalfarizi/hospital-api/internal/service"
)

func New(log *logger.Logger, cfg *config.Config) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.Default()

	// setup health check components
	healthRepo := repository.NewHealthRepo()
	healthSvc := service.NewHealthService(healthRepo)

	// public health endpoint
	r.GET("/health", handler.HealthCheck(healthSvc, log.Logger))

	// example versioned endpoint (could be removed later)
	v1 := r.Group("/api/v1")
	{
		v1.GET("/status", handler.HealthCheck(healthSvc, log.Logger))
	}

	return r
}
