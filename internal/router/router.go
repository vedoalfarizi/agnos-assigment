package router

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/vedoalfarizi/hospital-api/internal/config"
	"github.com/vedoalfarizi/hospital-api/internal/handler"
	"github.com/vedoalfarizi/hospital-api/internal/logger"
	"github.com/vedoalfarizi/hospital-api/internal/middleware"
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
	hospitalRepo := repository.NewHospitalRepo(db)
	// pass JWT secret and expiration from config
	staffSvc := service.NewStaffService(staffRepo, hospitalRepo, []byte(cfg.JWTSecret), cfg.JWTExpirationDays)

	// setup patient components (search)
	patientRepo := repository.NewPatientRepo(db)
	patientSvc := service.NewPatientService(patientRepo)

	// Auth middleware with JWT secret
	authMiddleware := middleware.AuthMiddleware([]byte(cfg.JWTSecret))

	// API routes group
	api := r.Group("/api")
	{
		// public health endpoint
		api.GET("/health", handler.HealthCheck(healthSvc, log.Logger))

		// staff create endpoint (public)
		api.POST("/staff/create", handler.CreateStaff(staffSvc, log.Logger))
		// staff login endpoint (public)
		api.POST("/staff/login", handler.LoginStaff(staffSvc, log.Logger))

		// Public patient search by ID endpoint
		api.GET("/patient/search/:id", handler.SearchPatientByID(patientSvc, log.Logger))

		// Protected patient endpoints (require auth)
		api.GET("/patient/search", authMiddleware, handler.SearchPatients(patientSvc, log.Logger))
	}

	return r
}
