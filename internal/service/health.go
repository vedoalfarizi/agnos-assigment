package service

import (
	"github.com/vedoalfarizi/hospital-api/internal/repository"
)

// HealthService orchestrates health-related business logic. It currently
// delegates to a repository layer but provides an abstraction for future
// enhancements (e.g., multi-check, caching, metrics).

type HealthService struct {
	repo *repository.HealthRepo
}

// NewHealthService builds a HealthService with the provided repository.
func NewHealthService(r *repository.HealthRepo) *HealthService {
	return &HealthService{repo: r}
}

// Check ensures the underlying database is reachable. Returns nil when healthy
// or an error otherwise.
func (s *HealthService) Check() error {
	return s.repo.Ping()
}
