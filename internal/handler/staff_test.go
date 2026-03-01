package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	"github.com/vedoalfarizi/hospital-api/internal/dto"
	"github.com/vedoalfarizi/hospital-api/internal/model"
	"github.com/vedoalfarizi/hospital-api/internal/repository"
	"github.com/vedoalfarizi/hospital-api/internal/service"
	"github.com/vedoalfarizi/hospital-api/mocks"
)

// Helper to create service with mocked repositories
func newStaffServiceWithMocks(mockStaffRepo *mocks.IStaffRepo, mockHospitalRepo *mocks.IHospitalRepo) *service.StaffService {
	jwtSecret := []byte("test-secret-key")
	expirationDays := 7
	return service.NewStaffService(mockStaffRepo, mockHospitalRepo, jwtSecret, expirationDays)
}

// Helper to create Gin context with test recorder
func setupGinContext(t *testing.T) (*gin.Engine, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	recorder := httptest.NewRecorder()
	return engine, recorder
}

// Helper to build request body
func buildCreateStaffRequest(username, password string, hospitalID int) []byte {
	req := dto.StaffCreateRequest{
		Username:   username,
		Password:   password,
		HospitalID: hospitalID,
	}
	body, _ := json.Marshal(req)
	return body
}

// TestCreateStaff_Success tests successful staff creation
func TestCreateStaff_Success(t *testing.T) {
	mockStaffRepo := new(mocks.IStaffRepo)
	mockHospitalRepo := new(mocks.IHospitalRepo)
	logger := logrus.New()

	// Setup expectations
	mockHospitalRepo.On("HospitalExists", 1).Return(nil)
	mockStaffRepo.On("CreateStaff", mock.AnythingOfType("*model.Staff")).
		Return(&model.Staff{
			ID:         1,
			HospitalID: 1,
			Username:   "john_doe",
		}, nil)

	svc := newStaffServiceWithMocks(mockStaffRepo, mockHospitalRepo)
	engine, recorder := setupGinContext(t)

	// Register route
	engine.POST("/staff/create", CreateStaff(svc, logger))

	// Build request
	body := buildCreateStaffRequest("john_doe", "securepass123", 1)
	req := httptest.NewRequest("POST", "/staff/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	engine.ServeHTTP(recorder, req)

	// Assertions
	if recorder.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", recorder.Code)
	}

	var resp APIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !resp.Success {
		t.Errorf("expected success=true, got false")
	}

	if resp.Data == nil {
		t.Errorf("expected data to be present in response")
	}

	// Verify mock expectations
	mockStaffRepo.AssertExpectations(t)
	mockHospitalRepo.AssertExpectations(t)
}

// TestCreateStaff_InvalidJSON tests malformed JSON request
func TestCreateStaff_InvalidJSON(t *testing.T) {
	mockStaffRepo := new(mocks.IStaffRepo)
	mockHospitalRepo := new(mocks.IHospitalRepo)
	logger := logrus.New()

	svc := newStaffServiceWithMocks(mockStaffRepo, mockHospitalRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.POST("/staff/create", CreateStaff(svc, logger))

	// Send malformed JSON
	req := httptest.NewRequest("POST", "/staff/create", bytes.NewReader([]byte(`{"invalid json`)))
	req.Header.Set("Content-Type", "application/json")

	engine.ServeHTTP(recorder, req)

	// Should return 400
	if recorder.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", recorder.Code)
	}

	var resp APIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Success {
		t.Errorf("expected success=false, got true")
	}

	if resp.Error == nil || resp.Error.Code != "INVALID_REQUEST" {
		t.Errorf("expected error code INVALID_REQUEST, got %v", resp.Error)
	}
}

// TestCreateStaff_MissingUsername tests validation error for missing username
func TestCreateStaff_MissingUsername(t *testing.T) {
	mockStaffRepo := new(mocks.IStaffRepo)
	mockHospitalRepo := new(mocks.IHospitalRepo)
	logger := logrus.New()

	svc := newStaffServiceWithMocks(mockStaffRepo, mockHospitalRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.POST("/staff/create", CreateStaff(svc, logger))

	// Missing username
	body := buildCreateStaffRequest("", "securepass123", 1)
	req := httptest.NewRequest("POST", "/staff/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", recorder.Code)
	}

	var resp APIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Success {
		t.Errorf("expected success=false, got true")
	}

	if resp.Error == nil || resp.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("expected error code VALIDATION_ERROR, got %v", resp.Error)
	}
}

// TestCreateStaff_InvalidUsernameFormat tests validation error for username too short
func TestCreateStaff_InvalidUsernameFormat(t *testing.T) {
	mockStaffRepo := new(mocks.IStaffRepo)
	mockHospitalRepo := new(mocks.IHospitalRepo)
	logger := logrus.New()

	svc := newStaffServiceWithMocks(mockStaffRepo, mockHospitalRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.POST("/staff/create", CreateStaff(svc, logger))

	// Username too short (min=3)
	body := buildCreateStaffRequest("ab", "securepass123", 1)
	req := httptest.NewRequest("POST", "/staff/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", recorder.Code)
	}

	var resp APIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Error == nil || resp.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("expected error code VALIDATION_ERROR")
	}
}

// TestCreateStaff_MissingPassword tests validation error for missing password
func TestCreateStaff_MissingPassword(t *testing.T) {
	mockStaffRepo := new(mocks.IStaffRepo)
	mockHospitalRepo := new(mocks.IHospitalRepo)
	logger := logrus.New()

	svc := newStaffServiceWithMocks(mockStaffRepo, mockHospitalRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.POST("/staff/create", CreateStaff(svc, logger))

	// Missing password
	body := buildCreateStaffRequest("john_doe", "", 1)
	req := httptest.NewRequest("POST", "/staff/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", recorder.Code)
	}

	var resp APIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Error == nil || resp.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("expected error code VALIDATION_ERROR")
	}
}

// TestCreateStaff_InvalidPasswordLength tests validation error for password too short
func TestCreateStaff_InvalidPasswordLength(t *testing.T) {
	mockStaffRepo := new(mocks.IStaffRepo)
	mockHospitalRepo := new(mocks.IHospitalRepo)
	logger := logrus.New()

	svc := newStaffServiceWithMocks(mockStaffRepo, mockHospitalRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.POST("/staff/create", CreateStaff(svc, logger))

	// Password too short (min=8)
	body := buildCreateStaffRequest("john_doe", "short01", 1)
	req := httptest.NewRequest("POST", "/staff/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", recorder.Code)
	}

	var resp APIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Error == nil || resp.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("expected error code VALIDATION_ERROR")
	}
}

// TestCreateStaff_MissingHospitalID tests validation error for missing hospital_id
func TestCreateStaff_MissingHospitalID(t *testing.T) {
	mockStaffRepo := new(mocks.IStaffRepo)
	mockHospitalRepo := new(mocks.IHospitalRepo)
	logger := logrus.New()

	svc := newStaffServiceWithMocks(mockStaffRepo, mockHospitalRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.POST("/staff/create", CreateStaff(svc, logger))

	// Invalid hospital_id (0)
	body := buildCreateStaffRequest("john_doe", "securepass123", 0)
	req := httptest.NewRequest("POST", "/staff/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", recorder.Code)
	}

	var resp APIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Error == nil || resp.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("expected error code VALIDATION_ERROR")
	}
}

// TestCreateStaff_HospitalNotFound tests service error when hospital doesn't exist
func TestCreateStaff_HospitalNotFound(t *testing.T) {
	mockStaffRepo := new(mocks.IStaffRepo)
	mockHospitalRepo := new(mocks.IHospitalRepo)
	logger := logrus.New()

	// Setup expectations: HospitalExists returns ErrNotFound
	mockHospitalRepo.On("HospitalExists", 999).Return(repository.ErrNotFound)

	svc := newStaffServiceWithMocks(mockStaffRepo, mockHospitalRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.POST("/staff/create", CreateStaff(svc, logger))

	body := buildCreateStaffRequest("john_doe", "securepass123", 999)
	req := httptest.NewRequest("POST", "/staff/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	engine.ServeHTTP(recorder, req)

	// Should return 404
	if recorder.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", recorder.Code)
	}

	var resp APIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Success {
		t.Errorf("expected success=false, got true")
	}

	if resp.Error == nil || resp.Error.Code != "HOSPITAL_NOT_FOUND" {
		t.Errorf("expected error code HOSPITAL_NOT_FOUND, got %v", resp.Error)
	}

	mockHospitalRepo.AssertExpectations(t)
}

// TestCreateStaff_DuplicateUsername tests service error when username already exists
func TestCreateStaff_DuplicateUsername(t *testing.T) {
	mockStaffRepo := new(mocks.IStaffRepo)
	mockHospitalRepo := new(mocks.IHospitalRepo)
	logger := logrus.New()

	// Setup expectations
	mockHospitalRepo.On("HospitalExists", 1).Return(nil)
	mockStaffRepo.On("CreateStaff", mock.AnythingOfType("*model.Staff")).
		Return(nil, repository.ErrDuplicate)

	svc := newStaffServiceWithMocks(mockStaffRepo, mockHospitalRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.POST("/staff/create", CreateStaff(svc, logger))

	body := buildCreateStaffRequest("john_doe", "securepass123", 1)
	req := httptest.NewRequest("POST", "/staff/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	engine.ServeHTTP(recorder, req)

	// Should return 409
	if recorder.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", recorder.Code)
	}

	var resp APIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Success {
		t.Errorf("expected success=false, got true")
	}

	if resp.Error == nil || resp.Error.Code != "DUPLICATE_USERNAME" {
		t.Errorf("expected error code DUPLICATE_USERNAME, got %v", resp.Error)
	}

	mockStaffRepo.AssertExpectations(t)
	mockHospitalRepo.AssertExpectations(t)
}

// TestCreateStaff_DatabaseError tests service error for generic database failures
func TestCreateStaff_DatabaseError(t *testing.T) {
	mockStaffRepo := new(mocks.IStaffRepo)
	mockHospitalRepo := new(mocks.IHospitalRepo)
	logger := logrus.New()

	// Setup expectations
	mockHospitalRepo.On("HospitalExists", 1).Return(nil)
	mockStaffRepo.On("CreateStaff", mock.AnythingOfType("*model.Staff")).
		Return(nil, errors.New("database connection error"))

	svc := newStaffServiceWithMocks(mockStaffRepo, mockHospitalRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.POST("/staff/create", CreateStaff(svc, logger))

	body := buildCreateStaffRequest("john_doe", "securepass123", 1)
	req := httptest.NewRequest("POST", "/staff/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	engine.ServeHTTP(recorder, req)

	// Should return 500
	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", recorder.Code)
	}

	var resp APIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Success {
		t.Errorf("expected success=false, got true")
	}

	if resp.Error == nil || resp.Error.Code != "INTERNAL_ERROR" {
		t.Errorf("expected error code INTERNAL_ERROR, got %v", resp.Error)
	}

	mockStaffRepo.AssertExpectations(t)
	mockHospitalRepo.AssertExpectations(t)
}

// TestCreateStaff_HospitalExistsError tests service error when HospitalExists call fails
func TestCreateStaff_HospitalExistsError(t *testing.T) {
	mockStaffRepo := new(mocks.IStaffRepo)
	mockHospitalRepo := new(mocks.IHospitalRepo)
	logger := logrus.New()

	// Setup expectations: HospitalExists returns generic error
	mockHospitalRepo.On("HospitalExists", 1).Return(errors.New("database error"))

	svc := newStaffServiceWithMocks(mockStaffRepo, mockHospitalRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.POST("/staff/create", CreateStaff(svc, logger))

	body := buildCreateStaffRequest("john_doe", "securepass123", 1)
	req := httptest.NewRequest("POST", "/staff/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	engine.ServeHTTP(recorder, req)

	// Should return 500
	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", recorder.Code)
	}

	var resp APIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Error == nil || resp.Error.Code != "INTERNAL_ERROR" {
		t.Errorf("expected error code INTERNAL_ERROR")
	}

	mockHospitalRepo.AssertExpectations(t)
}

// TestCreateStaff_NegativeHospitalID tests validation error for negative hospital_id
func TestCreateStaff_NegativeHospitalID(t *testing.T) {
	mockStaffRepo := new(mocks.IStaffRepo)
	mockHospitalRepo := new(mocks.IHospitalRepo)
	logger := logrus.New()

	svc := newStaffServiceWithMocks(mockStaffRepo, mockHospitalRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.POST("/staff/create", CreateStaff(svc, logger))

	// Negative hospital_id
	body := buildCreateStaffRequest("john_doe", "securepass123", -1)
	req := httptest.NewRequest("POST", "/staff/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", recorder.Code)
	}

	var resp APIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Error == nil || resp.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("expected error code VALIDATION_ERROR")
	}
}

// TestCreateStaff_EmptyRequest tests validation error for completely empty request
func TestCreateStaff_EmptyRequest(t *testing.T) {
	mockStaffRepo := new(mocks.IStaffRepo)
	mockHospitalRepo := new(mocks.IHospitalRepo)
	logger := logrus.New()

	svc := newStaffServiceWithMocks(mockStaffRepo, mockHospitalRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.POST("/staff/create", CreateStaff(svc, logger))

	// Empty object
	req := httptest.NewRequest("POST", "/staff/create", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")

	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", recorder.Code)
	}

	var resp APIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Error == nil || resp.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("expected error code VALIDATION_ERROR")
	}
}
