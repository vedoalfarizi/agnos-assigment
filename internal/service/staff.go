package service

import (
	"context"
	"fmt"

	"github.com/vedoalfarizi/hospital-api/internal/dto"
	"github.com/vedoalfarizi/hospital-api/internal/model"
	"github.com/vedoalfarizi/hospital-api/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// StaffService orchestrates staff-related business logic including validation,
// password hashing, and database operations.

type StaffService struct {
	repo *repository.StaffRepo
}

// NewStaffService builds a StaffService with the provided repository.
func NewStaffService(r *repository.StaffRepo) *StaffService {
	return &StaffService{repo: r}
}

// CreateStaff validates the request, checks hospital existence, hashes the password,
// and creates a new staff member. Returns a service-level response DTO.
func (s *StaffService) CreateStaff(ctx context.Context, req *dto.StaffCreateRequest) (*dto.StaffCreateResponse, error) {
	// Verify hospital exists via repository; repository returns ErrNotFound if missing
	err := s.repo.HospitalExists(req.HospitalID)
	if err != nil {
		return nil, err
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create staff model
	staff := &model.Staff{
		HospitalID: req.HospitalID,
		Username:   req.Username,
		Password:   string(hashedPassword),
	}

	// Create staff in database
	createdStaff, err := s.repo.CreateStaff(staff)
	if err != nil {
		return nil, err
	}

	// Return service-level response (data only)
	return &dto.StaffCreateResponse{
		ID:         createdStaff.ID,
		Username:   createdStaff.Username,
		HospitalID: createdStaff.HospitalID,
	}, nil
}
