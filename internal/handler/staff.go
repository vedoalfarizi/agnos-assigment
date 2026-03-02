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
			logger.WarnfWithContext(c.Request.Context(), "staff creation request validation failed: error=%v", err)
			Error(c, 400, "INVALID_REQUEST", "Invalid request format")
			return
		}

		// Validate request
		validate := validator.New()
		if err := validate.Struct(&req); err != nil {
			logger.WarnfWithContext(c.Request.Context(), "staff creation field validation failed: username=%s, hospital_id=%d, error=%v", req.Username, req.HospitalID, err)
			Error(c, 400, "VALIDATION_ERROR", "Validation failed: "+err.Error())
			return
		}

		// Call service to create staff
		resp, err := svc.CreateStaff(c.Request.Context(), &req)
		if err != nil {
			// Handle specific domain errors - service/repo already logged these
			if err == repository.ErrNotFound {
				Error(c, 404, "HOSPITAL_NOT_FOUND", "Hospital not found")
				return
			}
			if err == repository.ErrDuplicate {
				Error(c, 409, "DUPLICATE_USERNAME", "Username already exists")
				return
			}
			// Other errors already logged by service/repo
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
			logger.WarnfWithContext(c.Request.Context(), "login request validation failed: error=%v", err)
			Error(c, 400, "INVALID_REQUEST", "Invalid request format")
			return
		}

		validate := validator.New()
		if err := validate.Struct(&req); err != nil {
			logger.WarnfWithContext(c.Request.Context(), "login field validation failed: username=%s, error=%v", req.Username, err)
			Error(c, 400, "VALIDATION_ERROR", "Validation failed: "+err.Error())
			return
		}

		resp, err := svc.Login(c.Request.Context(), &req)
		if err != nil {
			if err == service.ErrInvalidCredentials {
				// Service already logged invalid password, just return response
				Error(c, 401, "UNAUTHORIZED", "Invalid username or password")
				return
			}
			// Other errors already logged by service/repo
			Error(c, 500, "INTERNAL_ERROR", "Login failed")
			return
		}

		Success(c, resp)
	}
}
