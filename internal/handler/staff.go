package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/vedoalfarizi/hospital-api/internal/dto"
	"github.com/vedoalfarizi/hospital-api/internal/logger"
	"github.com/vedoalfarizi/hospital-api/internal/repository"
	"github.com/vedoalfarizi/hospital-api/internal/service"
)

// CreateStaff returns a Gin handler that creates a new staff member.
// It validates the request, calls the service layer, and returns appropriate responses.
func CreateStaff(svc *service.StaffService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.StaffCreateRequest

		// Parse JSON request body
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Warnf("invalid request body: %v", err)
			Error(c, 400, "INVALID_REQUEST", "Invalid request format")
			return
		}

		// Validate request
		validate := validator.New()
		if err := validate.Struct(&req); err != nil {
			logger.Warnf("validation failed: %v", err)
			Error(c, 400, "VALIDATION_ERROR", "Validation failed: "+err.Error())
			return
		}

		// Call service to create staff
		resp, err := svc.CreateStaff(c.Request.Context(), &req)
		if err != nil {
			// Handle specific domain errors
			if err == repository.ErrNotFound {
				logger.Warnf("hospital not found: %d", req.HospitalID)
				Error(c, 404, "HOSPITAL_NOT_FOUND", "Hospital not found")
				return
			}
			if err == repository.ErrDuplicate {
				logger.Warnf("duplicate username: %s", req.Username)
				Error(c, 409, "DUPLICATE_USERNAME", "Username already exists")
				return
			}
			logger.Errorf("failed to create staff: %v", err)
			Error(c, 500, "INTERNAL_ERROR", "Failed to create staff member")
			return
		}

		// Success response - handler adds human-readable message
		Success(c, gin.H{
			"id":          resp.ID,
			"username":    resp.Username,
			"hospital_id": resp.HospitalID,
		})
	}
}

// LoginStaff returns a Gin handler for staff authentication. It validates the
// credentials and returns a JWT token on success.
func LoginStaff(svc *service.StaffService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.StaffLoginRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Warnf("invalid login request: %v", err)
			Error(c, 400, "INVALID_REQUEST", "Invalid request format")
			return
		}

		validate := validator.New()
		if err := validate.Struct(&req); err != nil {
			logger.Warnf("validation failed: %v", err)
			Error(c, 400, "VALIDATION_ERROR", "Validation failed: "+err.Error())
			return
		}

		resp, err := svc.Login(c.Request.Context(), &req)
		if err != nil {
			if err == service.ErrInvalidCredentials {
				logger.Warnf("invalid credentials for user %s", req.Username)
				Error(c, 401, "UNAUTHORIZED", "Invalid username or password")
				return
			}
			// other errors bubble up
			logger.Errorf("login failed: %v", err)
			Error(c, 500, "INTERNAL_ERROR", "Login failed")
			return
		}

		Success(c, resp)
	}
}
