package router

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/vedoalfarizi/hospital-api/internal/config"
	"github.com/vedoalfarizi/hospital-api/internal/handler"
	"github.com/vedoalfarizi/hospital-api/internal/logger"
	"github.com/vedoalfarizi/hospital-api/internal/repository"
	"github.com/vedoalfarizi/hospital-api/internal/service"
)

func New(log *logger.Logger, cfg *config.Config, db *sqlx.DB) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.Default()

	// setup health check components
	healthRepo := repository.NewHealthRepo(db)
	healthSvc := service.NewHealthService(healthRepo)

	// setup staff components (registration + login)
	staffRepo := repository.NewStaffRepo(db)
	// pass JWT secret and expiration from config
	staffSvc := service.NewStaffService(staffRepo, []byte(cfg.JWTSecret), cfg.JWTExpirationDays)

	// public health endpoint
	r.GET("/health", handler.HealthCheck(healthSvc, log.Logger))

	// staff create endpoint (public)
	r.POST("/staff/create", handler.CreateStaff(staffSvc, log.Logger))
	// staff login endpoint (public)
	r.POST("/staff/login", handler.LoginStaff(staffSvc, log.Logger))

	// example versioned endpoint (could be removed later)
	v1 := r.Group("/api/v1")
	{
		v1.GET("/status", handler.HealthCheck(healthSvc, log.Logger))
	}

	return r
}
