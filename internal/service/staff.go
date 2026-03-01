package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/vedoalfarizi/hospital-api/internal/dto"
	"github.com/vedoalfarizi/hospital-api/internal/model"
	"github.com/vedoalfarizi/hospital-api/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// StaffService orchestrates staff-related business logic including validation,
// password hashing, and database operations.

type StaffService struct {
	repo           repository.IStaffRepo
	hospitalRepo   repository.IHospitalRepo
	jwtSecret      []byte
	expirationDays int
}

// NewStaffService builds a StaffService with the provided repositories and
// authentication configuration (JWT secret and expiration days).
func NewStaffService(r repository.IStaffRepo, h repository.IHospitalRepo, jwtSecret []byte, expirationDays int) *StaffService {
	return &StaffService{repo: r, hospitalRepo: h, jwtSecret: jwtSecret, expirationDays: expirationDays}
}

// CreateStaff validates the request, checks hospital existence, hashes the password,
// and creates a new staff member. Returns a service-level response DTO.
func (s *StaffService) CreateStaff(ctx context.Context, req *dto.StaffCreateRequest) (*dto.StaffCreateResponse, error) {
	// Verify hospital exists via HospitalRepo; returns ErrNotFound if missing
	err := s.hospitalRepo.HospitalExists(req.HospitalID)
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

// ErrInvalidCredentials indicates the provided username/password pair was wrong.
var ErrInvalidCredentials = errors.New("invalid credentials")

// Login authenticates a staff member and returns a JWT access token if
// credentials are valid. It returns ErrInvalidCredentials for authentication
// failures, or other errors for unexpected issues.
func (s *StaffService) Login(ctx context.Context, req *dto.StaffLoginRequest) (*dto.StaffLoginResponse, error) {
	// lookup user by username
	staff, err := s.repo.GetByUsername(req.Username)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// compare password
	if bcrypt.CompareHashAndPassword([]byte(staff.Password), []byte(req.Password)) != nil {
		return nil, ErrInvalidCredentials
	}

	// build token
	expiresAt := time.Now().Add(time.Duration(s.expirationDays) * 24 * time.Hour).Unix()
	claims := jwt.MapClaims{
		"staff_id":    staff.ID,
		"hospital_id": staff.HospitalID,
		"exp":         expiresAt,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign token: %w", err)
	}

	return &dto.StaffLoginResponse{
		AccessToken: tokenString,
		TokenType:   "Bearer",
		ExpiresIn:   expiresAt - time.Now().Unix(),
	}, nil
}
