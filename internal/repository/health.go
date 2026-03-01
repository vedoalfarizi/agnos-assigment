package repository

import (
	"errors"

	"github.com/jmoiron/sqlx"
)

// IHealthRepo abstracts health-related data access for testability.
type IHealthRepo interface {
	// Ping performs a simple ping against the provided *sqlx.DB.
	Ping() error
}

// HealthRepo provides low-level access to database health checks.
// It exists to separate persistence concerns from higher layers.

type HealthRepo struct {
	db *sqlx.DB
}

// NewHealthRepo constructs a HealthRepo with an explicit database connection.
func NewHealthRepo(db *sqlx.DB) *HealthRepo {
	return &HealthRepo{db: db}
}

// Ping performs a simple ping against the provided *sqlx.DB.
func (r *HealthRepo) Ping() error {
	if r.db == nil {
		return errors.New("database not initialized")
	}
	return r.db.Ping()
}
