package repository

import (
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/vedoalfarizi/hospital-api/internal/logger"
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
		logger.Errorf("database connection not initialized for health check")
		return errors.New("database not initialized")
	}
	err := r.db.Ping()
	if err != nil {
		logger.Errorf("database health check failed: error=%v", err)
		return err
	}
	logger.Debugf("database health check passed")
	return nil
}
