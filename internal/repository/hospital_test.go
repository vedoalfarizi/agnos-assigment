package repository

import (
	"database/sql"
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

// TestHospitalExists_TableDriven runs table-driven tests for the HospitalExists method.
func TestHospitalExists_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		hospitalID  int
		setupMock   func(m sqlmock.Sqlmock)
		expectedErr error
	}{
		{
			name:       "Hospital exists - returns nil",
			hospitalID: 1,
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT 1 FROM hospital WHERE id = \$1`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
			},
			expectedErr: nil,
		},
		{
			name:       "Hospital not found - returns ErrNotFound",
			hospitalID: 999,
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT 1 FROM hospital WHERE id = \$1`).
					WithArgs(999).
					WillReturnError(sql.ErrNoRows)
			},
			expectedErr: ErrNotFound,
		},
		{
			name:       "Database connection error",
			hospitalID: 2,
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT 1 FROM hospital WHERE id = \$1`).
					WithArgs(2).
					WillReturnError(sql.ErrConnDone)
			},
			expectedErr: sql.ErrConnDone,
		},
		{
			name:       "Row scan error - invalid data type",
			hospitalID: 3,
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT 1 FROM hospital WHERE id = \$1`).
					WithArgs(3).
					WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow("invalid_non_int"))
			},
			expectedErr: nil, // Placeholder - will be overridden in test logic since scan errors are non-nil
		},
		{
			name:       "Hospital exists - different ID",
			hospitalID: 42,
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT 1 FROM hospital WHERE id = \$1`).
					WithArgs(42).
					WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock database
			mockDB, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer mockDB.Close()

			// Setup expectations
			tt.setupMock(mock)

			// Wrap mock with sqlx
			db := sqlx.NewDb(mockDB, "postgres")
			repo := NewHospitalRepo(db)

			// Execute
			result := repo.HospitalExists(tt.hospitalID)

			// Verify - special handling for row scan error test
			if tt.name == "Row scan error - invalid data type" {
				// For scan errors, just verify there's an error
				if result == nil {
					t.Errorf("expected error for scan error case, got nil")
				}
			} else {
				if tt.expectedErr == nil {
					if result != nil {
						t.Errorf("expected no error, got %v", result)
					}
				} else {
					if result == nil {
						t.Errorf("expected error %v, got nil", tt.expectedErr)
					} else if !errors.Is(result, tt.expectedErr) {
						t.Errorf("expected error %v, got %v", tt.expectedErr, result)
					}
				}
			}

			// Ensure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unmet sqlmock expectations: %v", err)
			}
		})
	}
}
