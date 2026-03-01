package dto

import "time"

// PatientSearchRequest contains optional filter criteria for searching patients.
// All fields are optional - only non-empty fields will be used in the search query.
type PatientSearchRequest struct {
	NationalID  string `form:"national_id"`
	PassportID  string `form:"passport_id"`
	FirstName   string `form:"first_name"`
	MiddleName  string `form:"middle_name"`
	LastName    string `form:"last_name"`
	DateOfBirth string `form:"date_of_birth"` // YYYY-MM-DD format
	PhoneNumber string `form:"phone_number"`
	Email       string `form:"email"`
}

// PatientSearchResponse contains the patient info returned in search results.
type PatientSearchResponse struct {
	ID           int        `json:"id"`
	HospitalID   int        `json:"hospital_id"`
	FirstNameEn  *string    `json:"first_name_en"`
	MiddleNameEn *string    `json:"middle_name_en"`
	LastNameEn   *string    `json:"last_name_en"`
	FirstNameTh  *string    `json:"first_name_th"`
	MiddleNameTh *string    `json:"middle_name_th"`
	LastNameTh   *string    `json:"last_name_th"`
	NationalID   *string    `json:"national_id"`
	PassportID   *string    `json:"passport_id"`
	PhoneNumber  *string    `json:"phone_number"`
	Email        *string    `json:"email"`
	DateOfBirth  *time.Time `json:"date_of_birth"`
}

// PatientSearchByIDResponse contains patient info for single patient lookup (public endpoint).
type PatientSearchByIDResponse struct {
	FirstNameTh  *string `json:"first_name_th"`
	MiddleNameTh *string `json:"middle_name_th"`
	LastNameTh   *string `json:"last_name_th"`
	FirstNameEn  *string `json:"first_name_en"`
	MiddleNameEn *string `json:"middle_name_en"`
	LastNameEn   *string `json:"last_name_en"`
	DateOfBirth  *string `json:"date_of_birth"` // YYYY-MM-DD format
	PatientHN    *string `json:"patient_hn"`
	NationalID   *string `json:"national_id"`
	PassportID   *string `json:"passport_id"`
	PhoneNumber  *string `json:"phone_number"`
	Email        *string `json:"email"`
	Gender       *string `json:"gender"`
}
