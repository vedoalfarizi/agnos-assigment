package repository

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// IHospitalRepo abstracts hospital-related data access for testability.
type IHospitalRepo interface {
	// HospitalExists checks if a hospital with the given ID exists, returning ErrNotFound when missing.
	HospitalExists(hospitalID int) error
}

// HospitalRepo provides access to hospital-related queries.
type HospitalRepo struct {
	db *sqlx.DB
}

// NewHospitalRepo constructs a HospitalRepo instance with a provided db connection.
func NewHospitalRepo(db *sqlx.DB) *HospitalRepo {
	return &HospitalRepo{db: db}
}

// HospitalExists checks if a hospital with the given ID exists, returning ErrNotFound when missing.
func (r *HospitalRepo) HospitalExists(hospitalID int) error {
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
