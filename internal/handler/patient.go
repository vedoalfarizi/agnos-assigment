package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/vedoalfarizi/hospital-api/internal/dto"
	"github.com/vedoalfarizi/hospital-api/internal/logger"
	"github.com/vedoalfarizi/hospital-api/internal/service"
)

// SearchPatientByID returns a Gin handler that searches for a single patient by national_id or passport_id.
// This is a public endpoint (no authentication required).
// Returns 404 if patient not found.
func SearchPatientByID(svc *service.PatientService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			logger.WarnfWithContext(c.Request.Context(), "invalid request: empty id parameter")
			Error(c, 400, "INVALID_REQUEST", "ID parameter is required")
			return
		}

		patient, err := svc.GetPatientByID(id)
		if err != nil {
			Error(c, 500, "INTERNAL_ERROR", "Failed to retrieve patient")
			return
		}

		if patient == nil {
			Error(c, 404, "PATIENT_NOT_FOUND", "Patient not found")
			return
		}

		Success(c, patient)
	}
}

// SearchPatients returns a Gin handler that searches for patients within a hospital.
// Requires JWT authentication. Extracts hospital_id from JWT claims to ensure
// staff can only search patients from their own hospital.
func SearchPatients(svc *service.PatientService) gin.HandlerFunc {
	return func(c *gin.Context) {
		hospitalIDInterface, exists := c.Get("hospital_id")
		if !exists {
			logger.ErrorfWithContext(c.Request.Context(), "internal error: hospital_id not found in context")
			Error(c, 500, "INTERNAL_ERROR", "Hospital ID not found in context")
			return
		}

		hospitalID, ok := hospitalIDInterface.(int)
		if !ok {
			logger.ErrorfWithContext(c.Request.Context(), "internal error: hospital_id is not an int: %T", hospitalIDInterface)
			Error(c, 500, "INTERNAL_ERROR", "Invalid hospital ID type")
			return
		}

		var query dto.PatientSearchRequest
		if err := c.ShouldBindQuery(&query); err != nil {
			logger.WarnfWithContext(c.Request.Context(), "patient search validation failed: hospital_id=%d, error=%v", hospitalID, err)
			Error(c, 400, "INVALID_REQUEST", "Invalid query parameters")
			return
		}

		// Note: db is not used directly in the current implementation but passed for consistency
		results, err := svc.SearchPatients(nil, hospitalID, query)
		if err != nil {
			Error(c, 500, "INTERNAL_ERROR", "Failed to search patients")
			return
		}

		Success(c, results)
	}
}
