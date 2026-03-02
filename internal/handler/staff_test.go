package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/vedoalfarizi/hospital-api/internal/dto"
	"github.com/vedoalfarizi/hospital-api/internal/model"
	"github.com/vedoalfarizi/hospital-api/internal/repository"
	"github.com/vedoalfarizi/hospital-api/internal/service"
	"github.com/vedoalfarizi/hospital-api/mocks"
	"golang.org/x/crypto/bcrypt"
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
	engine.POST("/staff/create", CreateStaff(svc))

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
	

	svc := newStaffServiceWithMocks(mockStaffRepo, mockHospitalRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.POST("/staff/create", CreateStaff(svc))

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
	

	svc := newStaffServiceWithMocks(mockStaffRepo, mockHospitalRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.POST("/staff/create", CreateStaff(svc))

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
	

	svc := newStaffServiceWithMocks(mockStaffRepo, mockHospitalRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.POST("/staff/create", CreateStaff(svc))

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
	

	svc := newStaffServiceWithMocks(mockStaffRepo, mockHospitalRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.POST("/staff/create", CreateStaff(svc))

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
	

	svc := newStaffServiceWithMocks(mockStaffRepo, mockHospitalRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.POST("/staff/create", CreateStaff(svc))

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
	

	svc := newStaffServiceWithMocks(mockStaffRepo, mockHospitalRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.POST("/staff/create", CreateStaff(svc))

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
	

	// Setup expectations: HospitalExists returns ErrNotFound
	mockHospitalRepo.On("HospitalExists", 999).Return(repository.ErrNotFound)

	svc := newStaffServiceWithMocks(mockStaffRepo, mockHospitalRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.POST("/staff/create", CreateStaff(svc))

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
	

	// Setup expectations
	mockHospitalRepo.On("HospitalExists", 1).Return(nil)
	mockStaffRepo.On("CreateStaff", mock.AnythingOfType("*model.Staff")).
		Return(nil, repository.ErrDuplicate)

	svc := newStaffServiceWithMocks(mockStaffRepo, mockHospitalRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.POST("/staff/create", CreateStaff(svc))

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
	

	// Setup expectations
	mockHospitalRepo.On("HospitalExists", 1).Return(nil)
	mockStaffRepo.On("CreateStaff", mock.AnythingOfType("*model.Staff")).
		Return(nil, errors.New("database connection error"))

	svc := newStaffServiceWithMocks(mockStaffRepo, mockHospitalRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.POST("/staff/create", CreateStaff(svc))

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
	

	// Setup expectations: HospitalExists returns generic error
	mockHospitalRepo.On("HospitalExists", 1).Return(errors.New("database error"))

	svc := newStaffServiceWithMocks(mockStaffRepo, mockHospitalRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.POST("/staff/create", CreateStaff(svc))

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
	

	svc := newStaffServiceWithMocks(mockStaffRepo, mockHospitalRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.POST("/staff/create", CreateStaff(svc))

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
	

	svc := newStaffServiceWithMocks(mockStaffRepo, mockHospitalRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.POST("/staff/create", CreateStaff(svc))

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

// ========== LoginStaff Tests ==========

// Helper to build login request body
func buildLoginRequest(username, password string) []byte {
	req := dto.StaffLoginRequest{
		Username: username,
		Password: password,
	}
	body, _ := json.Marshal(req)
	return body
}

// TestLoginStaff_Success tests successful staff login
func TestLoginStaff_Success(t *testing.T) {
	mockStaffRepo := new(mocks.IStaffRepo)
	mockHospitalRepo := new(mocks.IHospitalRepo)
	

	// Setup expectations - staff found with correct password hash
	password := "securepass123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	mockStaffRepo.On("GetByUsername", "john_doe").
		Return(&model.Staff{
			ID:         1,
			HospitalID: 1,
			Username:   "john_doe",
			Password:   string(hashedPassword), // pre-hashed
		}, nil)

	svc := newStaffServiceWithMocks(mockStaffRepo, mockHospitalRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.POST("/staff/login", LoginStaff(svc))

	body := buildLoginRequest("john_doe", "securepass123")
	req := httptest.NewRequest("POST", "/staff/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	engine.ServeHTTP(recorder, req)

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
		t.Errorf("expected data to contain JWT token")
	}

	// Verify token structure in response
	dataMap, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data to be map, got %T", resp.Data)
	}

	if token, exists := dataMap["access_token"]; !exists || token == "" {
		t.Errorf("expected access_token in response")
	}

	if tokenType, exists := dataMap["token_type"]; !exists || tokenType != "Bearer" {
		t.Errorf("expected token_type=Bearer")
	}

	if expiresIn, exists := dataMap["expires_in"]; !exists || expiresIn == nil {
		t.Errorf("expected expires_in in response")
	}

	mockStaffRepo.AssertExpectations(t)
}

// TestLoginStaff_InvalidJSON tests malformed JSON request
func TestLoginStaff_InvalidJSON(t *testing.T) {
	mockStaffRepo := new(mocks.IStaffRepo)
	mockHospitalRepo := new(mocks.IHospitalRepo)
	

	svc := newStaffServiceWithMocks(mockStaffRepo, mockHospitalRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.POST("/staff/login", LoginStaff(svc))

	req := httptest.NewRequest("POST", "/staff/login", bytes.NewReader([]byte(`{"invalid json`)))
	req.Header.Set("Content-Type", "application/json")

	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", recorder.Code)
	}

	var resp APIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Error == nil || resp.Error.Code != "INVALID_REQUEST" {
		t.Errorf("expected error code INVALID_REQUEST")
	}
}

// TestLoginStaff_MissingUsername tests validation error for missing username
func TestLoginStaff_MissingUsername(t *testing.T) {
	mockStaffRepo := new(mocks.IStaffRepo)
	mockHospitalRepo := new(mocks.IHospitalRepo)
	

	svc := newStaffServiceWithMocks(mockStaffRepo, mockHospitalRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.POST("/staff/login", LoginStaff(svc))

	body := buildLoginRequest("", "securepass123")
	req := httptest.NewRequest("POST", "/staff/login", bytes.NewReader(body))
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

// TestLoginStaff_InvalidUsernameLength tests validation error for username too short
func TestLoginStaff_InvalidUsernameLength(t *testing.T) {
	mockStaffRepo := new(mocks.IStaffRepo)
	mockHospitalRepo := new(mocks.IHospitalRepo)
	

	svc := newStaffServiceWithMocks(mockStaffRepo, mockHospitalRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.POST("/staff/login", LoginStaff(svc))

	// Username too short (min=3)
	body := buildLoginRequest("ab", "securepass123")
	req := httptest.NewRequest("POST", "/staff/login", bytes.NewReader(body))
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

// TestLoginStaff_MissingPassword tests validation error for missing password
func TestLoginStaff_MissingPassword(t *testing.T) {
	mockStaffRepo := new(mocks.IStaffRepo)
	mockHospitalRepo := new(mocks.IHospitalRepo)
	

	svc := newStaffServiceWithMocks(mockStaffRepo, mockHospitalRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.POST("/staff/login", LoginStaff(svc))

	body := buildLoginRequest("john_doe", "")
	req := httptest.NewRequest("POST", "/staff/login", bytes.NewReader(body))
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

// TestLoginStaff_InvalidPasswordLength tests validation error for password too short
func TestLoginStaff_InvalidPasswordLength(t *testing.T) {
	mockStaffRepo := new(mocks.IStaffRepo)
	mockHospitalRepo := new(mocks.IHospitalRepo)
	

	svc := newStaffServiceWithMocks(mockStaffRepo, mockHospitalRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.POST("/staff/login", LoginStaff(svc))

	// Password too short (min=8)
	body := buildLoginRequest("john_doe", "short01")
	req := httptest.NewRequest("POST", "/staff/login", bytes.NewReader(body))
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

// TestLoginStaff_EmptyRequest tests validation error for all empty fields
func TestLoginStaff_EmptyRequest(t *testing.T) {
	mockStaffRepo := new(mocks.IStaffRepo)
	mockHospitalRepo := new(mocks.IHospitalRepo)
	

	svc := newStaffServiceWithMocks(mockStaffRepo, mockHospitalRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.POST("/staff/login", LoginStaff(svc))

	req := httptest.NewRequest("POST", "/staff/login", bytes.NewReader([]byte(`{}`)))
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

// TestLoginStaff_UserNotFound tests invalid credentials error when user doesn't exist
func TestLoginStaff_UserNotFound(t *testing.T) {
	mockStaffRepo := new(mocks.IStaffRepo)
	mockHospitalRepo := new(mocks.IHospitalRepo)
	

	// Setup expectations - user not found
	mockStaffRepo.On("GetByUsername", "nonexistent").
		Return(nil, repository.ErrNotFound)

	svc := newStaffServiceWithMocks(mockStaffRepo, mockHospitalRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.POST("/staff/login", LoginStaff(svc))

	body := buildLoginRequest("nonexistent", "securepass123")
	req := httptest.NewRequest("POST", "/staff/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	engine.ServeHTTP(recorder, req)

	// Should return 401 for invalid credentials
	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", recorder.Code)
	}

	var resp APIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Success {
		t.Errorf("expected success=false, got true")
	}

	if resp.Error == nil || resp.Error.Code != "UNAUTHORIZED" {
		t.Errorf("expected error code UNAUTHORIZED, got %v", resp.Error)
	}

	mockStaffRepo.AssertExpectations(t)
}

// TestLoginStaff_InvalidPassword tests invalid credentials error for wrong password
func TestLoginStaff_InvalidPassword(t *testing.T) {
	mockStaffRepo := new(mocks.IStaffRepo)
	mockHospitalRepo := new(mocks.IHospitalRepo)
	

	// Setup expectations - user found but password doesn't match
	password := "securepass123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	mockStaffRepo.On("GetByUsername", "john_doe").
		Return(&model.Staff{
			ID:         1,
			HospitalID: 1,
			Username:   "john_doe",
			Password:   string(hashedPassword),
		}, nil)

	svc := newStaffServiceWithMocks(mockStaffRepo, mockHospitalRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.POST("/staff/login", LoginStaff(svc))

	// Wrong password
	body := buildLoginRequest("john_doe", "wrongpassword1")
	req := httptest.NewRequest("POST", "/staff/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	engine.ServeHTTP(recorder, req)

	// Should return 401 for invalid credentials
	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", recorder.Code)
	}

	var resp APIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Error == nil || resp.Error.Code != "UNAUTHORIZED" {
		t.Errorf("expected error code UNAUTHORIZED")
	}

	mockStaffRepo.AssertExpectations(t)
}

// TestLoginStaff_DatabaseError tests service error for database failures
func TestLoginStaff_DatabaseError(t *testing.T) {
	mockStaffRepo := new(mocks.IStaffRepo)
	mockHospitalRepo := new(mocks.IHospitalRepo)
	

	// Setup expectations - database error
	mockStaffRepo.On("GetByUsername", "john_doe").
		Return(nil, errors.New("database connection error"))

	svc := newStaffServiceWithMocks(mockStaffRepo, mockHospitalRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.POST("/staff/login", LoginStaff(svc))

	body := buildLoginRequest("john_doe", "securepass123")
	req := httptest.NewRequest("POST", "/staff/login", bytes.NewReader(body))
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
		t.Errorf("expected error code INTERNAL_ERROR")
	}

	mockStaffRepo.AssertExpectations(t)
}
