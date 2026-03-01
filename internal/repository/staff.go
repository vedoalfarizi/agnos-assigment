package repository

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/vedoalfarizi/hospital-api/internal/model"
)

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
			return nil, ErrDuplicate
		}
		return nil, err
	}

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
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &s, nil
}

// TODO::move to hospital repository
// HospitalExists checks if a hospital with the given ID exists, returning ErrNotFound when missing.
func (r *StaffRepo) HospitalExists(hospitalID int) error {
	const query = `SELECT 1 FROM hospital WHERE id = $1`

	var exists int
	err := r.db.QueryRow(query, hospitalID).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		return err
	}

	return nil
}
