package service

import (
	"github.com/jmoiron/sqlx"

	"github.com/vedoalfarizi/hospital-api/internal/dto"
	"github.com/vedoalfarizi/hospital-api/internal/repository"
)

type PatientService struct {
	repo repository.IPatientRepo
}

// NewPatientService constructs a PatientService instance.
func NewPatientService(repo repository.IPatientRepo) *PatientService {
	return &PatientService{repo: repo}
}

// GetPatientByID retrieves a single patient by national_id or passport_id.
// Returns nil if patient not found, or an error if the database query fails.
func (s *PatientService) GetPatientByID(id string) (*dto.PatientSearchByIDResponse, error) {
	patient, err := s.repo.GetPatientByID(id)
	if err != nil {
		// Repository already logged this error
		return nil, err
	}

	if patient == nil {
		// Repository logs not found at debug level
		return nil, nil
	}

	// Format date as YYYY-MM-DD string
	var dateOfBirthStr *string
	if patient.DateOfBirth != nil {
		dateStr := patient.DateOfBirth.Format("2006-01-02")
		dateOfBirthStr = &dateStr
	}

	return &dto.PatientSearchByIDResponse{
		FirstNameTh:  patient.FirstNameTh,
		MiddleNameTh: patient.MiddleNameTh,
		LastNameTh:   patient.LastNameTh,
		FirstNameEn:  patient.FirstNameEn,
		MiddleNameEn: patient.MiddleNameEn,
		LastNameEn:   patient.LastNameEn,
		DateOfBirth:  dateOfBirthStr,
		PatientHN:    patient.HospitalName,
		NationalID:   patient.NationalID,
		PassportID:   patient.PassportID,
		PhoneNumber:  patient.PhoneNumber,
		Email:        patient.Email,
		Gender:       patient.Gender,
	}, nil
}

// SearchPatients searches for patients matching the provided criteria within a hospital.
// Returns a slice of PatientSearchResponse DTOs. Returns an empty slice if no matches are found.
func (s *PatientService) SearchPatients(db *sqlx.DB, hospitalID int, query dto.PatientSearchRequest) ([]dto.PatientSearchResponse, error) {
	// Note: db parameter is provided for consistency with service patterns,
	// though this implementation uses the repository's internal db connection

	patients, err := s.repo.SearchPatients(hospitalID, query)
	if err != nil {
		// Repository already logged this error
		return nil, err
	}

	// Map Patient models to DTOs
	responses := make([]dto.PatientSearchResponse, len(patients))
	for i, p := range patients {
		responses[i] = dto.PatientSearchResponse{
			ID:           p.ID,
			HospitalID:   p.HospitalID,
			FirstNameEn:  p.FirstNameEn,
			MiddleNameEn: p.MiddleNameEn,
			LastNameEn:   p.LastNameEn,
			FirstNameTh:  p.FirstNameTh,
			MiddleNameTh: p.MiddleNameTh,
			LastNameTh:   p.LastNameTh,
			NationalID:   p.NationalID,
			PassportID:   p.PassportID,
			PhoneNumber:  p.PhoneNumber,
			Email:        p.Email,
			DateOfBirth:  p.DateOfBirth,
		}
	}

	return responses, nil
}
