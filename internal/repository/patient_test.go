package repository

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/vedoalfarizi/hospital-api/internal/dto"
)

// TestGetPatientByID_TableDriven runs table-driven tests for GetPatientByID method.
func TestGetPatientByID_TableDriven(t *testing.T) {
	mockTime := time.Date(1990, 1, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		patientID   string
		setupMock   func(m sqlmock.Sqlmock)
		expectedID  *int
		expectedErr error
	}{
		{
			name:      "Patient found by national_id",
			patientID: "1234567890123",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT p.id, p.hospital_id, p.first_name_th, p.middle_name_th, p.last_name_th,`).
					WithArgs("1234567890123").
					WillReturnRows(sqlmock.NewRows([]string{
						"id", "hospital_id", "first_name_th", "middle_name_th", "last_name_th",
						"first_name_en", "middle_name_en", "last_name_en", "national_id", "passport_id",
						"date_of_birth", "phone_number", "email", "gender", "hospital_name", "created_at", "updated_at",
					}).AddRow(
						1, 1, "สมชาย", nil, "ใจดี",
						"Somchai", nil, "Jaidee", "1234567890123", nil,
						mockTime, "0812345678", "somchai@example.com", "M", "General Hospital", mockTime, mockTime,
					))
			},
			expectedID:  intPtr(1),
			expectedErr: nil,
		},
		{
			name:      "Patient found by passport_id",
			patientID: "PP123456",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT p.id, p.hospital_id, p.first_name_th, p.middle_name_th, p.last_name_th,`).
					WithArgs("PP123456").
					WillReturnRows(sqlmock.NewRows([]string{
						"id", "hospital_id", "first_name_th", "middle_name_th", "last_name_th",
						"first_name_en", "middle_name_en", "last_name_en", "national_id", "passport_id",
						"date_of_birth", "phone_number", "email", "gender", "hospital_name", "created_at", "updated_at",
					}).AddRow(
						2, 2, nil, nil, "Smith",
						"John", nil, "Smith", nil, "PP123456",
						mockTime, "0887654321", "john@example.com", "M", "City Hospital", mockTime, mockTime,
					))
			},
			expectedID:  intPtr(2),
			expectedErr: nil,
		},
		{
			name:      "Patient not found",
			patientID: "NONEXISTENT",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT p.id, p.hospital_id, p.first_name_th, p.middle_name_th, p.last_name_th,`).
					WithArgs("NONEXISTENT").
					WillReturnError(sql.ErrNoRows)
			},
			expectedID:  nil,
			expectedErr: nil,
		},
		{
			name:      "Database error",
			patientID: "1234567890123",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT p.id, p.hospital_id, p.first_name_th, p.middle_name_th, p.last_name_th,`).
					WithArgs("1234567890123").
					WillReturnError(sql.ErrConnDone)
			},
			expectedID:  nil,
			expectedErr: sql.ErrConnDone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer mockDB.Close()

			tt.setupMock(mock)

			db := sqlx.NewDb(mockDB, "postgres")
			repo := NewPatientRepo(db)

			result, err := repo.GetPatientByID(tt.patientID)

			// Verify error
			if tt.expectedErr != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedErr)
				} else if !errors.Is(err, tt.expectedErr) {
					t.Errorf("expected error %v, got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			}

			// Verify result
			if tt.expectedID != nil {
				if result == nil {
					t.Errorf("expected patient with ID %d, got nil", *tt.expectedID)
				} else if result.ID != *tt.expectedID {
					t.Errorf("expected patient ID %d, got %d", *tt.expectedID, result.ID)
				}
			} else {
				if result != nil {
					t.Errorf("expected nil result, got patient with ID %d", result.ID)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unmet sqlmock expectations: %v", err)
			}
		})
	}
}

// TestSearchPatients_TableDriven runs table-driven tests for SearchPatients method.
func TestSearchPatients_TableDriven(t *testing.T) {
	mockTime := time.Date(1990, 1, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		hospitalID    int
		searchRequest dto.PatientSearchRequest
		setupMock     func(m sqlmock.Sqlmock, hospitalID int, request dto.PatientSearchRequest)
		expectedCount int
		expectedErr   error
	}{
		{
			name:          "Search with hospital_id only - returns all patients",
			hospitalID:    1,
			searchRequest: dto.PatientSearchRequest{},
			setupMock: func(m sqlmock.Sqlmock, hospitalID int, request dto.PatientSearchRequest) {
				m.ExpectQuery(`SELECT id, hospital_id, first_name_th, middle_name_th, last_name_th,`).
					WithArgs(hospitalID).
					WillReturnRows(sqlmock.NewRows([]string{
						"id", "hospital_id", "first_name_th", "middle_name_th", "last_name_th",
						"first_name_en", "middle_name_en", "last_name_en", "national_id", "passport_id",
						"date_of_birth", "phone_number", "email", "gender", "created_at", "updated_at",
					}).
						AddRow(1, 1, "สมชาย", nil, "ใจดี", "Somchai", nil, "Jaidee", "1234567890123", nil, mockTime, "0812345678", "somchai@example.com", "M", mockTime, mockTime).
						AddRow(2, 1, "สมหญิง", nil, "สวย", "Somying", nil, "Suwai", "1234567890124", nil, mockTime, "0887654321", "somying@example.com", "F", mockTime, mockTime))
			},
			expectedCount: 2,
			expectedErr:   nil,
		},
		{
			name:          "Search by national_id - exact match",
			hospitalID:    1,
			searchRequest: dto.PatientSearchRequest{NationalID: "1234567890123"},
			setupMock: func(m sqlmock.Sqlmock, hospitalID int, request dto.PatientSearchRequest) {
				m.ExpectQuery(`SELECT id, hospital_id, first_name_th, middle_name_th, last_name_th,`).
					WithArgs(hospitalID, request.NationalID).
					WillReturnRows(sqlmock.NewRows([]string{
						"id", "hospital_id", "first_name_th", "middle_name_th", "last_name_th",
						"first_name_en", "middle_name_en", "last_name_en", "national_id", "passport_id",
						"date_of_birth", "phone_number", "email", "gender", "created_at", "updated_at",
					}).AddRow(1, 1, "สมชาย", nil, "ใจดี", "Somchai", nil, "Jaidee", "1234567890123", nil, mockTime, "0812345678", "somchai@example.com", "M", mockTime, mockTime))
			},
			expectedCount: 1,
			expectedErr:   nil,
		},
		{
			name:          "Search by first_name - substring match",
			hospitalID:    1,
			searchRequest: dto.PatientSearchRequest{FirstName: "somchai"},
			setupMock: func(m sqlmock.Sqlmock, hospitalID int, request dto.PatientSearchRequest) {
				m.ExpectQuery(`SELECT id, hospital_id, first_name_th, middle_name_th, last_name_th,`).
					WithArgs(hospitalID, request.FirstName).
					WillReturnRows(sqlmock.NewRows([]string{
						"id", "hospital_id", "first_name_th", "middle_name_th", "last_name_th",
						"first_name_en", "middle_name_en", "last_name_en", "national_id", "passport_id",
						"date_of_birth", "phone_number", "email", "gender", "created_at", "updated_at",
					}).AddRow(1, 1, "สมชาย", nil, "ใจดี", "Somchai", nil, "Jaidee", "1234567890123", nil, mockTime, "0812345678", "somchai@example.com", "M", mockTime, mockTime))
			},
			expectedCount: 1,
			expectedErr:   nil,
		},
		{
			name:          "Search by date_of_birth - exact match",
			hospitalID:    1,
			searchRequest: dto.PatientSearchRequest{DateOfBirth: "1990-01-15"},
			setupMock: func(m sqlmock.Sqlmock, hospitalID int, request dto.PatientSearchRequest) {
				m.ExpectQuery(`SELECT id, hospital_id, first_name_th, middle_name_th, last_name_th,`).
					WithArgs(hospitalID, request.DateOfBirth).
					WillReturnRows(sqlmock.NewRows([]string{
						"id", "hospital_id", "first_name_th", "middle_name_th", "last_name_th",
						"first_name_en", "middle_name_en", "last_name_en", "national_id", "passport_id",
						"date_of_birth", "phone_number", "email", "gender", "created_at", "updated_at",
					}).AddRow(1, 1, "สมชาย", nil, "ใจดี", "Somchai", nil, "Jaidee", "1234567890123", nil, mockTime, "0812345678", "somchai@example.com", "M", mockTime, mockTime))
			},
			expectedCount: 1,
			expectedErr:   nil,
		},
		{
			name:          "Search by email - exact match",
			hospitalID:    1,
			searchRequest: dto.PatientSearchRequest{Email: "somchai@example.com"},
			setupMock: func(m sqlmock.Sqlmock, hospitalID int, request dto.PatientSearchRequest) {
				m.ExpectQuery(`SELECT id, hospital_id, first_name_th, middle_name_th, last_name_th,`).
					WithArgs(hospitalID, request.Email).
					WillReturnRows(sqlmock.NewRows([]string{
						"id", "hospital_id", "first_name_th", "middle_name_th", "last_name_th",
						"first_name_en", "middle_name_en", "last_name_en", "national_id", "passport_id",
						"date_of_birth", "phone_number", "email", "gender", "created_at", "updated_at",
					}).AddRow(1, 1, "สมชาย", nil, "ใจดี", "Somchai", nil, "Jaidee", "1234567890123", nil, mockTime, "0812345678", "somchai@example.com", "M", mockTime, mockTime))
			},
			expectedCount: 1,
			expectedErr:   nil,
		},
		{
			name:          "Search by passport_id - exact match",
			hospitalID:    1,
			searchRequest: dto.PatientSearchRequest{PassportID: "PP123456"},
			setupMock: func(m sqlmock.Sqlmock, hospitalID int, request dto.PatientSearchRequest) {
				m.ExpectQuery(`SELECT id, hospital_id, first_name_th, middle_name_th, last_name_th,`).
					WithArgs(hospitalID, request.PassportID).
					WillReturnRows(sqlmock.NewRows([]string{
						"id", "hospital_id", "first_name_th", "middle_name_th", "last_name_th",
						"first_name_en", "middle_name_en", "last_name_en", "national_id", "passport_id",
						"date_of_birth", "phone_number", "email", "gender", "created_at", "updated_at",
					}).AddRow(2, 1, nil, nil, "Smith", "John", nil, "Smith", nil, "PP123456", mockTime, "0887654321", "john@example.com", "M", mockTime, mockTime))
			},
			expectedCount: 1,
			expectedErr:   nil,
		},
		{
			name:          "Search with all name fields - combines with OR",
			hospitalID:    1,
			searchRequest: dto.PatientSearchRequest{FirstName: "somchai", MiddleName: "mid", LastName: "jaidee"},
			setupMock: func(m sqlmock.Sqlmock, hospitalID int, request dto.PatientSearchRequest) {
				m.ExpectQuery(`SELECT id, hospital_id, first_name_th, middle_name_th, last_name_th,`).
					WithArgs(hospitalID, request.FirstName, request.MiddleName, request.LastName).
					WillReturnRows(sqlmock.NewRows([]string{
						"id", "hospital_id", "first_name_th", "middle_name_th", "last_name_th",
						"first_name_en", "middle_name_en", "last_name_en", "national_id", "passport_id",
						"date_of_birth", "phone_number", "email", "gender", "created_at", "updated_at",
					}).AddRow(1, 1, "สมชาย", "มิด", "ใจดี", "Somchai", "Mid", "Jaidee", "1234567890123", nil, mockTime, "0812345678", "somchai@example.com", "M", mockTime, mockTime))
			},
			expectedCount: 1,
			expectedErr:   nil,
		},
		{
			name:          "Search with multiple filters - national_id and first_name",
			hospitalID:    1,
			searchRequest: dto.PatientSearchRequest{NationalID: "1234567890123", FirstName: "somchai"},
			setupMock: func(m sqlmock.Sqlmock, hospitalID int, request dto.PatientSearchRequest) {
				m.ExpectQuery(`SELECT id, hospital_id, first_name_th, middle_name_th, last_name_th,`).
					WithArgs(hospitalID, request.NationalID, request.FirstName).
					WillReturnRows(sqlmock.NewRows([]string{
						"id", "hospital_id", "first_name_th", "middle_name_th", "last_name_th",
						"first_name_en", "middle_name_en", "last_name_en", "national_id", "passport_id",
						"date_of_birth", "phone_number", "email", "gender", "created_at", "updated_at",
					}).AddRow(1, 1, "สมชาย", nil, "ใจดี", "Somchai", nil, "Jaidee", "1234567890123", nil, mockTime, "0812345678", "somchai@example.com", "M", mockTime, mockTime))
			},
			expectedCount: 1,
			expectedErr:   nil,
		},
		{
			name:          "Search returns multiple results",
			hospitalID:    1,
			searchRequest: dto.PatientSearchRequest{},
			setupMock: func(m sqlmock.Sqlmock, hospitalID int, request dto.PatientSearchRequest) {
				m.ExpectQuery(`SELECT id, hospital_id, first_name_th, middle_name_th, last_name_th,`).
					WithArgs(hospitalID).
					WillReturnRows(sqlmock.NewRows([]string{
						"id", "hospital_id", "first_name_th", "middle_name_th", "last_name_th",
						"first_name_en", "middle_name_en", "last_name_en", "national_id", "passport_id",
						"date_of_birth", "phone_number", "email", "gender", "created_at", "updated_at",
					}).
						AddRow(1, 1, "สมชาย", nil, "ใจดี", "Somchai", nil, "Jaidee", "1234567890123", nil, mockTime, "0812345678", "somchai@example.com", "M", mockTime, mockTime).
						AddRow(2, 1, "สมหญิง", nil, "สวย", "Somying", nil, "Suwai", "1234567890124", nil, mockTime, "0887654321", "somying@example.com", "F", mockTime, mockTime).
						AddRow(3, 1, "สมหนึ่ง", nil, "ใจใหญ่", "Somnung", nil, "Jaijai", "1234567890125", nil, mockTime, "0899999999", "somnung@example.com", "M", mockTime, mockTime))
			},
			expectedCount: 3,
			expectedErr:   nil,
		},
		{
			name:          "Search with no results - returns empty slice",
			hospitalID:    1,
			searchRequest: dto.PatientSearchRequest{NationalID: "NONEXISTENT"},
			setupMock: func(m sqlmock.Sqlmock, hospitalID int, request dto.PatientSearchRequest) {
				m.ExpectQuery(`SELECT id, hospital_id, first_name_th, middle_name_th, last_name_th,`).
					WithArgs(hospitalID, request.NationalID).
					WillReturnRows(sqlmock.NewRows([]string{
						"id", "hospital_id", "first_name_th", "middle_name_th", "last_name_th",
						"first_name_en", "middle_name_en", "last_name_en", "national_id", "passport_id",
						"date_of_birth", "phone_number", "email", "gender", "created_at", "updated_at",
					}))
			},
			expectedCount: 0,
			expectedErr:   nil,
		},
		{
			name:          "Database error during search",
			hospitalID:    1,
			searchRequest: dto.PatientSearchRequest{NationalID: "1234567890123"},
			setupMock: func(m sqlmock.Sqlmock, hospitalID int, request dto.PatientSearchRequest) {
				m.ExpectQuery(`SELECT id, hospital_id, first_name_th, middle_name_th, last_name_th,`).
					WithArgs(hospitalID, request.NationalID).
					WillReturnError(sql.ErrConnDone)
			},
			expectedCount: 0,
			expectedErr:   sql.ErrConnDone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer mockDB.Close()

			tt.setupMock(mock, tt.hospitalID, tt.searchRequest)

			db := sqlx.NewDb(mockDB, "postgres")
			repo := NewPatientRepo(db)

			results, err := repo.SearchPatients(tt.hospitalID, tt.searchRequest)

			// Verify error
			if tt.expectedErr != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedErr)
				} else if !errors.Is(err, tt.expectedErr) {
					t.Errorf("expected error %v, got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			}

			// Verify result count
			if len(results) != tt.expectedCount {
				t.Errorf("expected %d results, got %d", tt.expectedCount, len(results))
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unmet sqlmock expectations: %v", err)
			}
		})
	}
}

// TestSearchPatients_NameFields tests that name searches work across both Thai and English fields.
func TestSearchPatients_NameFields(t *testing.T) {
	mockTime := time.Date(1990, 1, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		hospitalID    int
		searchRequest dto.PatientSearchRequest
		setupMock     func(m sqlmock.Sqlmock, hospitalID int, request dto.PatientSearchRequest)
		expectedName  string
	}{
		{
			name:          "Search by middle_name - matches English",
			hospitalID:    1,
			searchRequest: dto.PatientSearchRequest{MiddleName: "Kumar"},
			setupMock: func(m sqlmock.Sqlmock, hospitalID int, request dto.PatientSearchRequest) {
				m.ExpectQuery(`SELECT id, hospital_id, first_name_th, middle_name_th, last_name_th,`).
					WithArgs(hospitalID, request.MiddleName).
					WillReturnRows(sqlmock.NewRows([]string{
						"id", "hospital_id", "first_name_th", "middle_name_th", "last_name_th",
						"first_name_en", "middle_name_en", "last_name_en", "national_id", "passport_id",
						"date_of_birth", "phone_number", "email", "gender", "created_at", "updated_at",
					}).AddRow(3, 1, nil, nil, nil, "Raj", "Kumar", "Patel", "1234567890125", nil, mockTime, "0899999999", "raj@example.com", "M", mockTime, mockTime))
			},
			expectedName: "Raj",
		},
		{
			name:          "Search by last_name - substring match",
			hospitalID:    1,
			searchRequest: dto.PatientSearchRequest{LastName: "smit"},
			setupMock: func(m sqlmock.Sqlmock, hospitalID int, request dto.PatientSearchRequest) {
				m.ExpectQuery(`SELECT id, hospital_id, first_name_th, middle_name_th, last_name_th,`).
					WithArgs(hospitalID, request.LastName).
					WillReturnRows(sqlmock.NewRows([]string{
						"id", "hospital_id", "first_name_th", "middle_name_th", "last_name_th",
						"first_name_en", "middle_name_en", "last_name_en", "national_id", "passport_id",
						"date_of_birth", "phone_number", "email", "gender", "created_at", "updated_at",
					}).AddRow(4, 1, nil, nil, nil, "John", nil, "Smith", nil, "PP123456", mockTime, "0811111111", "john@example.com", "M", mockTime, mockTime))
			},
			expectedName: "John",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer mockDB.Close()

			tt.setupMock(mock, tt.hospitalID, tt.searchRequest)

			db := sqlx.NewDb(mockDB, "postgres")
			repo := NewPatientRepo(db)

			results, err := repo.SearchPatients(tt.hospitalID, tt.searchRequest)

			if err != nil {
				t.Errorf("expected no error, got %v", err)
			}

			if len(results) == 0 {
				t.Errorf("expected at least 1 result, got 0")
			} else if results[0].FirstNameEn == nil || *results[0].FirstNameEn != tt.expectedName {
				expectedVal := ""
				if results[0].FirstNameEn != nil {
					expectedVal = *results[0].FirstNameEn
				}
				t.Errorf("expected first name %s, got %s", tt.expectedName, expectedVal)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unmet sqlmock expectations: %v", err)
			}
		})
	}
}

// Helper function to create int pointers
func intPtr(i int) *int {
	return &i
}
