package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/vedoalfarizi/hospital-api/internal/dto"
	"github.com/vedoalfarizi/hospital-api/internal/repository"
	"github.com/vedoalfarizi/hospital-api/internal/service"
)

// CreateStaff returns a Gin handler that creates a new staff member.
// It validates the request, calls the service layer, and returns appropriate responses.
func CreateStaff(svc *service.StaffService, log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.StaffCreateRequest

		// Parse JSON request body
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Warnf("invalid request body: %v", err)
			Error(c, 400, "INVALID_REQUEST", "Invalid request format")
			return
		}

		// Validate request
		validate := validator.New()
		if err := validate.Struct(&req); err != nil {
			log.Warnf("validation failed: %v", err)
			Error(c, 400, "VALIDATION_ERROR", "Validation failed: "+err.Error())
			return
		}

		// Call service to create staff
		resp, err := svc.CreateStaff(c.Request.Context(), &req)
		if err != nil {
			// Handle specific domain errors
			if err == repository.ErrNotFound {
				log.Warnf("hospital not found: %d", req.HospitalID)
				Error(c, 404, "HOSPITAL_NOT_FOUND", "Hospital not found")
				return
			}
			if err == repository.ErrDuplicate {
				log.Warnf("duplicate username: %s", req.Username)
				Error(c, 409, "DUPLICATE_USERNAME", "Username already exists")
				return
			}
			log.Errorf("failed to create staff: %v", err)
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
