package repository

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/vedoalfarizi/hospital-api/internal/model"
)

// TestCreateStaff_TableDriven runs table-driven tests for CreateStaff method.
func TestCreateStaff_TableDriven(t *testing.T) {
	mockTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name        string
		staff       *model.Staff
		setupMock   func(m sqlmock.Sqlmock)
		expectedErr error
		expectedID  *int
	}{
		{
			name: "Create staff successfully",
			staff: &model.Staff{
				HospitalID: 1,
				Username:   "johndoe",
				Password:   "hashedpassword123",
			},
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`INSERT INTO staff \(hospital_id, username, password, created_at, updated_at\)`).
					WithArgs(1, "johndoe", "hashedpassword123").
					WillReturnRows(sqlmock.NewRows([]string{"id", "hospital_id", "username", "password", "created_at", "updated_at"}).
						AddRow(1, 1, "johndoe", "hashedpassword123", mockTime, mockTime))
			},
			expectedErr: nil,
			expectedID:  intPtr(1),
		},
		{
			name: "Create staff with different hospital",
			staff: &model.Staff{
				HospitalID: 2,
				Username:   "janedoe",
				Password:   "hashedpassword456",
			},
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`INSERT INTO staff \(hospital_id, username, password, created_at, updated_at\)`).
					WithArgs(2, "janedoe", "hashedpassword456").
					WillReturnRows(sqlmock.NewRows([]string{"id", "hospital_id", "username", "password", "created_at", "updated_at"}).
						AddRow(2, 2, "janedoe", "hashedpassword456", mockTime, mockTime))
			},
			expectedErr: nil,
			expectedID:  intPtr(2),
		},
		{
			name: "Duplicate username - returns ErrDuplicate",
			staff: &model.Staff{
				HospitalID: 1,
				Username:   "johndoe",
				Password:   "hashedpassword123",
			},
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`INSERT INTO staff \(hospital_id, username, password, created_at, updated_at\)`).
					WithArgs(1, "johndoe", "hashedpassword123").
					WillReturnError(&pq.Error{Code: "23505", Message: "duplicate key value violates unique constraint"})
			},
			expectedErr: ErrDuplicate,
			expectedID:  nil,
		},
		{
			name: "Database connection error",
			staff: &model.Staff{
				HospitalID: 1,
				Username:   "newuser",
				Password:   "hashedpassword",
			},
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`INSERT INTO staff \(hospital_id, username, password, created_at, updated_at\)`).
					WithArgs(1, "newuser", "hashedpassword").
					WillReturnError(sql.ErrConnDone)
			},
			expectedErr: sql.ErrConnDone,
			expectedID:  nil,
		},
		{
			name: "Other PostgreSQL error - not duplicate",
			staff: &model.Staff{
				HospitalID: 999,
				Username:   "user",
				Password:   "pass",
			},
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`INSERT INTO staff \(hospital_id, username, password, created_at, updated_at\)`).
					WithArgs(999, "user", "pass").
					WillReturnError(&pq.Error{Code: "23503", Message: "foreign key constraint violated"})
			},
			expectedErr: nil, // Will be a pq.Error but not ErrDuplicate, so any non-ErrDuplicate error
			expectedID:  nil,
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
			repo := NewStaffRepo(db)

			result, err := repo.CreateStaff(tt.staff)

			// Verify error
			if tt.expectedErr != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedErr)
				} else if !errors.Is(err, tt.expectedErr) {
					t.Errorf("expected error %v, got %v", tt.expectedErr, err)
				}
			} else {
				if tt.name == "Other PostgreSQL error - not duplicate" {
					// This case expects a pq.Error but not ErrDuplicate
					if err == nil {
						t.Errorf("expected a pq.Error, got nil")
					}
					if _, isPqErr := err.(*pq.Error); !isPqErr {
						t.Errorf("expected *pq.Error, got %T", err)
					}
				} else if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			}

			// Verify result
			if tt.expectedID != nil {
				if result == nil {
					t.Errorf("expected staff with ID %d, got nil", *tt.expectedID)
				} else if result.ID != *tt.expectedID {
					t.Errorf("expected staff ID %d, got %d", *tt.expectedID, result.ID)
				} else if result.Username != tt.staff.Username {
					t.Errorf("expected username %s, got %s", tt.staff.Username, result.Username)
				}
			} else {
				if result != nil {
					t.Errorf("expected nil result, got staff with ID %d", result.ID)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unmet sqlmock expectations: %v", err)
			}
		})
	}
}

// TestGetByUsername_TableDriven runs table-driven tests for GetByUsername method.
func TestGetByUsername_TableDriven(t *testing.T) {
	mockTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name        string
		username    string
		setupMock   func(m sqlmock.Sqlmock)
		expectedErr error
		expectedID  *int
	}{
		{
			name:     "User found",
			username: "johndoe",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT id, hospital_id, username, password, created_at, updated_at FROM staff WHERE username = \$1`).
					WithArgs("johndoe").
					WillReturnRows(sqlmock.NewRows([]string{"id", "hospital_id", "username", "password", "created_at", "updated_at"}).
						AddRow(1, 1, "johndoe", "hashedpassword123", mockTime, mockTime))
			},
			expectedErr: nil,
			expectedID:  intPtr(1),
		},
		{
			name:     "User not found - returns ErrNotFound",
			username: "nonexistentuser",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT id, hospital_id, username, password, created_at, updated_at FROM staff WHERE username = \$1`).
					WithArgs("nonexistentuser").
					WillReturnError(sql.ErrNoRows)
			},
			expectedErr: ErrNotFound,
			expectedID:  nil,
		},
		{
			name:     "Database connection error",
			username: "johndoe",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT id, hospital_id, username, password, created_at, updated_at FROM staff WHERE username = \$1`).
					WithArgs("johndoe").
					WillReturnError(sql.ErrConnDone)
			},
			expectedErr: sql.ErrConnDone,
			expectedID:  nil,
		},
		{
			name:     "Case-sensitive username lookup - different case not found",
			username: "JohnDoe",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT id, hospital_id, username, password, created_at, updated_at FROM staff WHERE username = \$1`).
					WithArgs("JohnDoe").
					WillReturnError(sql.ErrNoRows)
			},
			expectedErr: ErrNotFound,
			expectedID:  nil,
		},
		{
			name:     "Another user found",
			username: "janedoe",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT id, hospital_id, username, password, created_at, updated_at FROM staff WHERE username = \$1`).
					WithArgs("janedoe").
					WillReturnRows(sqlmock.NewRows([]string{"id", "hospital_id", "username", "password", "created_at", "updated_at"}).
						AddRow(2, 2, "janedoe", "hashedpassword456", mockTime, mockTime))
			},
			expectedErr: nil,
			expectedID:  intPtr(2),
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
			repo := NewStaffRepo(db)

			result, err := repo.GetByUsername(tt.username)

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
					t.Errorf("expected staff with ID %d, got nil", *tt.expectedID)
				} else if result.ID != *tt.expectedID {
					t.Errorf("expected staff ID %d, got %d", *tt.expectedID, result.ID)
				} else if result.Username != tt.username {
					t.Errorf("expected username %s, got %s", tt.username, result.Username)
				}
			} else {
				if result != nil {
					t.Errorf("expected nil result, got staff with ID %d", result.ID)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unmet sqlmock expectations: %v", err)
			}
		})
	}
}

// TestStaffRepo_CreateAndRetrieve tests the workflow of creating staff and then retrieving them.
func TestStaffRepo_CreateAndRetrieve(t *testing.T) {
	mockTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer mockDB.Close()

	// Setup expectations for CreateStaff
	mock.ExpectQuery(`INSERT INTO staff`).
		WithArgs(1, "newadmin", "hashedpass").
		WillReturnRows(sqlmock.NewRows([]string{"id", "hospital_id", "username", "password", "created_at", "updated_at"}).
			AddRow(5, 1, "newadmin", "hashedpass", mockTime, mockTime))

	// Setup expectations for GetByUsername
	mock.ExpectQuery(`SELECT id, hospital_id, username, password, created_at, updated_at FROM staff WHERE username = \$1`).
		WithArgs("newadmin").
		WillReturnRows(sqlmock.NewRows([]string{"id", "hospital_id", "username", "password", "created_at", "updated_at"}).
			AddRow(5, 1, "newadmin", "hashedpass", mockTime, mockTime))

	db := sqlx.NewDb(mockDB, "postgres")
	repo := NewStaffRepo(db)

	// Create staff
	newStaff := &model.Staff{
		HospitalID: 1,
		Username:   "newadmin",
		Password:   "hashedpass",
	}

	created, err := repo.CreateStaff(newStaff)
	if err != nil {
		t.Errorf("CreateStaff failed: %v", err)
	}
	if created == nil || created.ID != 5 {
		t.Errorf("expected created staff with ID 5, got %v", created)
	}

	// Retrieve the created staff
	retrieved, err := repo.GetByUsername("newadmin")
	if err != nil {
		t.Errorf("GetByUsername failed: %v", err)
	}
	if retrieved == nil || retrieved.ID != 5 {
		t.Errorf("expected retrieved staff with ID 5, got %v", retrieved)
	}
	if retrieved.Username != "newadmin" {
		t.Errorf("expected username newadmin, got %s", retrieved.Username)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

// TestStaffRepo_MultipleUsers tests retrieving multiple different users.
func TestStaffRepo_MultipleUsers(t *testing.T) {
	mockTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	usernames := []string{"admin1", "admin2", "doctor1", "nurse1"}
	userIDs := []int{1, 2, 3, 4}

	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer mockDB.Close()

	// Setup expectations for each user
	for i, username := range usernames {
		mock.ExpectQuery(`SELECT id, hospital_id, username, password, created_at, updated_at FROM staff WHERE username = \$1`).
			WithArgs(username).
			WillReturnRows(sqlmock.NewRows([]string{"id", "hospital_id", "username", "password", "created_at", "updated_at"}).
				AddRow(userIDs[i], 1, username, "hashedpass"+string(rune(i)), mockTime, mockTime))
	}

	db := sqlx.NewDb(mockDB, "postgres")
	repo := NewStaffRepo(db)

	// Retrieve each user
	for i, username := range usernames {
		result, err := repo.GetByUsername(username)
		if err != nil {
			t.Errorf("GetByUsername(%s) failed: %v", username, err)
		}
		if result == nil || result.ID != userIDs[i] {
			t.Errorf("expected user %s with ID %d, got %v", username, userIDs[i], result)
		}
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}
