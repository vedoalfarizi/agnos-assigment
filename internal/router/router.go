package router

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/vedoalfarizi/hospital-api/internal/config"
	"github.com/vedoalfarizi/hospital-api/internal/handler"
	"github.com/vedoalfarizi/hospital-api/internal/middleware"
	"github.com/vedoalfarizi/hospital-api/internal/repository"
	"github.com/vedoalfarizi/hospital-api/internal/service"
)

func New(cfg *config.Config, db *sqlx.DB) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.Default()

	// Request context middleware - automatically adds request_id, client_ip, path, method to logs
	r.Use(middleware.RequestContextMiddleware())

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
		api.GET("/health", handler.HealthCheck(healthSvc))

		// staff create endpoint (public)
		api.POST("/staff/create", handler.CreateStaff(staffSvc))
		// staff login endpoint (public)
		api.POST("/staff/login", handler.LoginStaff(staffSvc))

		// Public patient search by ID endpoint
		api.GET("/patient/search/:id", handler.SearchPatientByID(patientSvc))

		// Protected patient endpoints (require auth)
		api.GET("/patient/search", authMiddleware, handler.SearchPatients(patientSvc))
	}

	return r
}
