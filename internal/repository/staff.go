package repository

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/vedoalfarizi/hospital-api/internal/logger"
	"github.com/vedoalfarizi/hospital-api/internal/model"
)

// IStaffRepo abstracts staff-related data access for testability.
type IStaffRepo interface {
	// CreateStaff inserts a new staff member and returns the created staff with ID populated.
	CreateStaff(staff *model.Staff) (*model.Staff, error)

	// GetByUsername retrieves a staff record by username. It returns ErrNotFound if no such user exists.
	GetByUsername(username string) (*model.Staff, error)
}

// Domain-level errors
var (
	ErrDuplicate = errors.New("duplicate entry")
	ErrNotFound  = errors.New("not found")
)

type StaffRepo struct {
	db *sqlx.DB
}

// NewStaffRepo constructs a StaffRepo instance with a provided db connection.
func NewStaffRepo(db *sqlx.DB) *StaffRepo {
	return &StaffRepo{db: db}
}

// CreateStaff inserts a new staff member and returns the created staff with ID populated.
func (r *StaffRepo) CreateStaff(staff *model.Staff) (*model.Staff, error) {
	const query = `
		INSERT INTO staff (hospital_id, username, password, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id, hospital_id, username, password, created_at, updated_at
	`

	var result model.Staff
	err := r.db.QueryRowx(query, staff.HospitalID, staff.Username, staff.Password).StructScan(&result)
	if err != nil {
		// Handle unique constraint violation (username already exists)
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			logger.Warnf("duplicate staff username: username=%s, hospital_id=%d", staff.Username, staff.HospitalID)
			return nil, ErrDuplicate
		}
		// Log other database errors with context
		logger.Errorf("failed to create staff member: username=%s, hospital_id=%d, error=%v", staff.Username, staff.HospitalID, err)
		return nil, err
	}

	logger.Infof("staff member created successfully: id=%d, username=%s, hospital_id=%d", result.ID, result.Username, result.HospitalID)
	return &result, nil
}

// GetByUsername retrieves a staff record by username. It returns ErrNotFound
// if no such user exists.
func (r *StaffRepo) GetByUsername(username string) (*model.Staff, error) {
	const query = `
		SELECT id, hospital_id, username, password, created_at, updated_at
		FROM staff
		WHERE username = $1
	`

	var s model.Staff
	err := r.db.Get(&s, query, username)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Debugf("staff member not found: username=%s", username)
			return nil, ErrNotFound
		}
		logger.Errorf("failed to retrieve staff member: username=%s, error=%v", username, err)
		return nil, err
	}
	logger.Debugf("staff member found: username=%s, id=%d, hospital_id=%d", username, s.ID, s.HospitalID)
	return &s, nil
}
