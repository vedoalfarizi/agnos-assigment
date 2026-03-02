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
	r.Use(middleware.RequestContextMiddleware())
	r.Use(middleware.LoggingMiddleware())

	healthRepo := repository.NewHealthRepo(db)
	healthSvc := service.NewHealthService(healthRepo)

	staffRepo := repository.NewStaffRepo(db)
	hospitalRepo := repository.NewHospitalRepo(db)
	staffSvc := service.NewStaffService(staffRepo, hospitalRepo, []byte(cfg.JWTSecret), cfg.JWTExpirationDays)

	patientRepo := repository.NewPatientRepo(db)
	patientSvc := service.NewPatientService(patientRepo)

	authMiddleware := middleware.AuthMiddleware([]byte(cfg.JWTSecret))

	api := r.Group("/api")
	{
		api.GET("/health", handler.HealthCheck(healthSvc))

		api.POST("/staff/create", handler.CreateStaff(staffSvc))
		api.POST("/staff/login", handler.LoginStaff(staffSvc))

		api.GET("/patient/search/:id", handler.SearchPatientByID(patientSvc))

		api.GET("/patient/search", authMiddleware, handler.SearchPatients(patientSvc))
	}

	return r
}
