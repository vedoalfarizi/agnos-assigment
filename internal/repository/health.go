package repository

import (
	"errors"

	"github.com/vedoalfarizi/hospital-api/internal/database/postgre"
)

// HealthRepo provides low-level access to database health checks.
// It exists to separate persistence concerns from higher layers.

type HealthRepo struct{}

// NewHealthRepo constructs a HealthRepo instance.
func NewHealthRepo() *HealthRepo {
	return &HealthRepo{}
}

// Ping performs a simple ping against the shared *sqlx.DB. Returns an error
// if the connection pool is nil or the ping fails.
func (r *HealthRepo) Ping() error {
	db := postgre.GetDB()
	if db == nil {
		return errors.New("database not initialized")
	}
	return db.Ping()
}
