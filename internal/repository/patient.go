package repository

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/vedoalfarizi/hospital-api/internal/dto"
	"github.com/vedoalfarizi/hospital-api/internal/model"
)

type PatientRepo struct {
	db *sqlx.DB
}

// NewPatientRepo constructs a PatientRepo instance with a provided db connection.
func NewPatientRepo(db *sqlx.DB) *PatientRepo {
	return &PatientRepo{db: db}
}

// SearchPatients dynamically builds a query based on provided search criteria.
// Always filters by hospital_id. For non-empty fields in the request:
// - IDs, dates, and contact info (national_id, passport_id, date_of_birth, phone_number, email) use exact match
// - Name fields (first_name, middle_name, last_name) use case-insensitive substring match across both Thai and English names
// Returns an empty slice (not an error) if no results are found.
func (r *PatientRepo) SearchPatients(hospitalID int, query dto.PatientSearchRequest) ([]model.Patient, error) {
	const baseQuery = `
		SELECT id, hospital_id, first_name_th, middle_name_th, last_name_th,
		       first_name_en, middle_name_en, last_name_en, national_id, passport_id,
		       date_of_birth, phone_number, email, gender, created_at, updated_at
		FROM patients
		WHERE hospital_id = $1
	`

	var whereConditions []string
	var args []interface{}
	args = append(args, hospitalID)

	paramIdx := 2

	// Exact match fields: national_id, passport_id, date_of_birth, phone_number, email
	if query.NationalID != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("national_id = $%d", paramIdx))
		args = append(args, query.NationalID)
		paramIdx++
	}

	if query.PassportID != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("passport_id = $%d", paramIdx))
		args = append(args, query.PassportID)
		paramIdx++
	}

	if query.DateOfBirth != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("date_of_birth = $%d", paramIdx))
		args = append(args, query.DateOfBirth)
		paramIdx++
	}

	if query.PhoneNumber != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("phone_number = $%d", paramIdx))
		args = append(args, query.PhoneNumber)
		paramIdx++
	}

	if query.Email != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("email = $%d", paramIdx))
		args = append(args, query.Email)
		paramIdx++
	}

	// Substring match for name fields: check all name fields (Thai & English) for each non-empty name
	nameSearches := []string{}
	if query.FirstName != "" {
		nameSearches = append(nameSearches, fmt.Sprintf("(first_name_en ILIKE '%%'||$%d||'%%' OR first_name_th ILIKE '%%'||$%d||'%%')", paramIdx, paramIdx))
		args = append(args, query.FirstName)
		paramIdx++
	}
	if query.MiddleName != "" {
		nameSearches = append(nameSearches, fmt.Sprintf("(middle_name_en ILIKE '%%'||$%d||'%%' OR middle_name_th ILIKE '%%'||$%d||'%%')", paramIdx, paramIdx))
		args = append(args, query.MiddleName)
		paramIdx++
	}
	if query.LastName != "" {
		nameSearches = append(nameSearches, fmt.Sprintf("(last_name_en ILIKE '%%'||$%d||'%%' OR last_name_th ILIKE '%%'||$%d||'%%')", paramIdx, paramIdx))
		args = append(args, query.LastName)
		paramIdx++
	}

	// If any name searches exist, combine them with OR
	if len(nameSearches) > 0 {
		whereConditions = append(whereConditions, "("+strings.Join(nameSearches, " OR ")+")")
	}

	// Build final query
	finalQuery := baseQuery
	if len(whereConditions) > 0 {
		finalQuery += " AND " + strings.Join(whereConditions, " AND ")
	}

	var patients []model.Patient
	err := r.db.Select(&patients, finalQuery, args...)
	if err != nil {
		return nil, err
	}

	// Return empty slice if no results (not an error)
	if len(patients) == 0 {
		return []model.Patient{}, nil
	}

	return patients, nil
}
