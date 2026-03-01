package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/vedoalfarizi/hospital-api/internal/dto"
	"github.com/vedoalfarizi/hospital-api/internal/service"
)

// SearchPatients returns a Gin handler that searches for patients within a hospital.
// Requires JWT authentication. Extracts hospital_id from JWT claims to ensure
// staff can only search patients from their own hospital.
func SearchPatients(svc *service.PatientService, log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract hospital_id and staff_id from context (set by auth middleware)
		hospitalIDInterface, exists := c.Get("hospital_id")
		if !exists {
			log.Errorf("hospital_id not found in context")
			Error(c, 500, "INTERNAL_ERROR", "Hospital ID not found in context")
			return
		}

		hospitalID, ok := hospitalIDInterface.(int)
		if !ok {
			log.Errorf("hospital_id is not an int: %T", hospitalIDInterface)
			Error(c, 500, "INTERNAL_ERROR", "Invalid hospital ID type")
			return
		}

		// Bind query parameters
		var query dto.PatientSearchRequest
		if err := c.ShouldBindQuery(&query); err != nil {
			log.Warnf("invalid query parameters: %v", err)
			Error(c, 400, "INVALID_REQUEST", "Invalid query parameters")
			return
		}

		// Call service to search patients
		// Note: db is not used directly in the current implementation but passed for consistency
		results, err := svc.SearchPatients(nil, hospitalID, query)
		if err != nil {
			log.Errorf("failed to search patients: %v", err)
			Error(c, 500, "INTERNAL_ERROR", "Failed to search patients")
			return
		}

		// Return results (empty slice if no matches)
		Success(c, results)
	}
}
