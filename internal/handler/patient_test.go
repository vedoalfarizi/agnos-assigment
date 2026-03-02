package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/vedoalfarizi/hospital-api/internal/model"
	"github.com/vedoalfarizi/hospital-api/internal/service"
	"github.com/vedoalfarizi/hospital-api/mocks"
)

// Helper to create PatientService with mocked repository
func newPatientServiceWithMocks(mockPatientRepo *mocks.IPatientRepo) *service.PatientService {
	return service.NewPatientService(mockPatientRepo)
}

// TestSearchPatients_Success tests successful patient search with results
func TestSearchPatients_Success(t *testing.T) {
	mockPatientRepo := new(mocks.IPatientRepo)
	

	// Setup expectations
	firstName := "John"
	lastName := "Doe"
	nationalID := "1234567890123"
	dateOfBirth := time.Date(1990, 1, 15, 0, 0, 0, 0, time.UTC)

	patients := []model.Patient{
		{
			ID:          1,
			HospitalID:  5,
			FirstNameEn: &firstName,
			LastNameEn:  &lastName,
			NationalID:  &nationalID,
			DateOfBirth: &dateOfBirth,
		},
	}

	mockPatientRepo.On("SearchPatients", 5, mock.AnythingOfType("dto.PatientSearchRequest")).
		Return(patients, nil)

	svc := newPatientServiceWithMocks(mockPatientRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.GET("/patient/search", SearchPatients(svc))

	req := httptest.NewRequest("GET", "/patient/search?first_name=John&last_name=Doe", nil)
	// Manually set context values that would be set by auth middleware
	req.Header.Set("Content-Type", "application/json")

	// Create request context and set hospital_id
	c, _ := gin.CreateTestContext(recorder)
	c.Request = req
	c.Set("hospital_id", 5)
	c.Set("staff_id", 1)

	// Execute handler directly
	SearchPatients(svc)(c)

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
		t.Errorf("expected data to contain results")
	}

	mockPatientRepo.AssertExpectations(t)
}

// TestSearchPatients_NoResults tests successful search returning no results
func TestSearchPatients_NoResults(t *testing.T) {
	mockPatientRepo := new(mocks.IPatientRepo)
	

	// Setup expectations - return empty slice
	mockPatientRepo.On("SearchPatients", 5, mock.AnythingOfType("dto.PatientSearchRequest")).
		Return([]model.Patient{}, nil)

	svc := newPatientServiceWithMocks(mockPatientRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.GET("/patient/search", SearchPatients(svc))

	req := httptest.NewRequest("GET", "/patient/search?first_name=NotExist", nil)
	req.Header.Set("Content-Type", "application/json")

	c, _ := gin.CreateTestContext(recorder)
	c.Request = req
	c.Set("hospital_id", 5)
	c.Set("staff_id", 1)

	SearchPatients(svc)(c)

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

	// Verify empty results returned
	dataSlice, ok := resp.Data.([]interface{})
	if !ok {
		t.Fatalf("expected data to be slice, got %T", resp.Data)
	}

	if len(dataSlice) != 0 {
		t.Errorf("expected empty results, got %d items", len(dataSlice))
	}

	mockPatientRepo.AssertExpectations(t)
}

// TestSearchPatients_MissingHospitalID tests error when hospital_id not in context
func TestSearchPatients_MissingHospitalID(t *testing.T) {
	mockPatientRepo := new(mocks.IPatientRepo)
	

	svc := newPatientServiceWithMocks(mockPatientRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.GET("/patient/search", SearchPatients(svc))

	req := httptest.NewRequest("GET", "/patient/search?first_name=John", nil)
	req.Header.Set("Content-Type", "application/json")

	c, _ := gin.CreateTestContext(recorder)
	c.Request = req
	// Deliberately not setting hospital_id

	SearchPatients(svc)(c)

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
}

// TestSearchPatients_InvalidHospitalIDType tests error when hospital_id has wrong type
func TestSearchPatients_InvalidHospitalIDType(t *testing.T) {
	mockPatientRepo := new(mocks.IPatientRepo)
	

	svc := newPatientServiceWithMocks(mockPatientRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.GET("/patient/search", SearchPatients(svc))

	req := httptest.NewRequest("GET", "/patient/search?first_name=John", nil)
	req.Header.Set("Content-Type", "application/json")

	c, _ := gin.CreateTestContext(recorder)
	c.Request = req
	c.Set("hospital_id", "invalid-string") // Wrong type

	SearchPatients(svc)(c)

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
}

// TestSearchPatients_InvalidQueryParameters tests handling of unknown query parameters
func TestSearchPatients_InvalidQueryParameters(t *testing.T) {
	mockPatientRepo := new(mocks.IPatientRepo)
	

	// Setup expectations - unknown fields are ignored
	mockPatientRepo.On("SearchPatients", 5, mock.AnythingOfType("dto.PatientSearchRequest")).
		Return([]model.Patient{}, nil)

	svc := newPatientServiceWithMocks(mockPatientRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.GET("/patient/search", SearchPatients(svc))

	// Query parameters with unknown fields - they should be ignored
	req := httptest.NewRequest("GET", "/patient/search?unknown_field=value", nil)
	req.Header.Set("Content-Type", "application/json")

	c, _ := gin.CreateTestContext(recorder)
	c.Request = req
	c.Set("hospital_id", 5)
	c.Set("staff_id", 1)

	SearchPatients(svc)(c)

	// This request should succeed since unknown fields are just ignored
	if recorder.Code != http.StatusOK {
		t.Errorf("expected status 200 for unknown fields, got %d", recorder.Code)
	}

	mockPatientRepo.AssertExpectations(t)
}

// TestSearchPatients_ServiceError tests error handling for service failures
func TestSearchPatients_ServiceError(t *testing.T) {
	mockPatientRepo := new(mocks.IPatientRepo)
	

	// Setup expectations - return error
	mockPatientRepo.On("SearchPatients", 5, mock.AnythingOfType("dto.PatientSearchRequest")).
		Return(nil, errors.New("database connection error"))

	svc := newPatientServiceWithMocks(mockPatientRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.GET("/patient/search", SearchPatients(svc))

	req := httptest.NewRequest("GET", "/patient/search?first_name=John", nil)
	req.Header.Set("Content-Type", "application/json")

	c, _ := gin.CreateTestContext(recorder)
	c.Request = req
	c.Set("hospital_id", 5)
	c.Set("staff_id", 1)

	SearchPatients(svc)(c)

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

	mockPatientRepo.AssertExpectations(t)
}

// TestSearchPatients_MultipleResults tests search returning multiple patients
func TestSearchPatients_MultipleResults(t *testing.T) {
	mockPatientRepo := new(mocks.IPatientRepo)
	

	// Setup expectations
	firstName1 := "John"
	lastName1 := "Doe"
	firstName2 := "Jane"
	lastName2 := "Smith"

	patients := []model.Patient{
		{
			ID:          1,
			HospitalID:  5,
			FirstNameEn: &firstName1,
			LastNameEn:  &lastName1,
		},
		{
			ID:          2,
			HospitalID:  5,
			FirstNameEn: &firstName2,
			LastNameEn:  &lastName2,
		},
	}

	mockPatientRepo.On("SearchPatients", 5, mock.AnythingOfType("dto.PatientSearchRequest")).
		Return(patients, nil)

	svc := newPatientServiceWithMocks(mockPatientRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.GET("/patient/search", SearchPatients(svc))

	req := httptest.NewRequest("GET", "/patient/search?first_name=J", nil)
	req.Header.Set("Content-Type", "application/json")

	c, _ := gin.CreateTestContext(recorder)
	c.Request = req
	c.Set("hospital_id", 5)
	c.Set("staff_id", 1)

	SearchPatients(svc)(c)

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

	dataSlice, ok := resp.Data.([]interface{})
	if !ok {
		t.Fatalf("expected data to be slice, got %T", resp.Data)
	}

	if len(dataSlice) != 2 {
		t.Errorf("expected 2 results, got %d", len(dataSlice))
	}

	mockPatientRepo.AssertExpectations(t)
}

// TestSearchPatients_AllFieldsFilterCriteria tests search with all filter fields populated
func TestSearchPatients_AllFieldsFilterCriteria(t *testing.T) {
	mockPatientRepo := new(mocks.IPatientRepo)
	

	// Setup expectations
	firstName := "John"
	patients := []model.Patient{
		{
			ID:          1,
			HospitalID:  5,
			FirstNameEn: &firstName,
		},
	}

	mockPatientRepo.On("SearchPatients", 5, mock.AnythingOfType("dto.PatientSearchRequest")).
		Return(patients, nil)

	svc := newPatientServiceWithMocks(mockPatientRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.GET("/patient/search", SearchPatients(svc))

	// Query with multiple filter criteria
	req := httptest.NewRequest("GET",
		"/patient/search?first_name=John&last_name=Doe&national_id=1234567890&email=john@test.com&phone_number=0812345678", nil)
	req.Header.Set("Content-Type", "application/json")

	c, _ := gin.CreateTestContext(recorder)
	c.Request = req
	c.Set("hospital_id", 5)
	c.Set("staff_id", 1)

	SearchPatients(svc)(c)

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

	mockPatientRepo.AssertExpectations(t)
}

// TestSearchPatients_DifferentHospitalID tests isolation of results by hospital_id
func TestSearchPatients_DifferentHospitalID(t *testing.T) {
	mockPatientRepo := new(mocks.IPatientRepo)
	

	firstName := "John"
	patients := []model.Patient{
		{
			ID:          1,
			HospitalID:  10, // Different hospital
			FirstNameEn: &firstName,
		},
	}

	// Mock expects search for hospital 10, not 5
	mockPatientRepo.On("SearchPatients", 10, mock.AnythingOfType("dto.PatientSearchRequest")).
		Return(patients, nil)

	svc := newPatientServiceWithMocks(mockPatientRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.GET("/patient/search", SearchPatients(svc))

	req := httptest.NewRequest("GET", "/patient/search?first_name=John", nil)
	req.Header.Set("Content-Type", "application/json")

	c, _ := gin.CreateTestContext(recorder)
	c.Request = req
	c.Set("hospital_id", 10) // Different hospital_id
	c.Set("staff_id", 2)

	SearchPatients(svc)(c)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", recorder.Code)
	}

	mockPatientRepo.AssertExpectations(t)
}

// TestSearchPatients_NullableFields tests handling of nullable fields in response
func TestSearchPatients_NullableFields(t *testing.T) {
	mockPatientRepo := new(mocks.IPatientRepo)
	

	// Patient with some nullable fields as nil
	patients := []model.Patient{
		{
			ID:           1,
			HospitalID:   5,
			FirstNameEn:  nil, // Nullable field
			MiddleNameEn: nil,
			PhoneNumber:  nil,
		},
	}

	mockPatientRepo.On("SearchPatients", 5, mock.AnythingOfType("dto.PatientSearchRequest")).
		Return(patients, nil)

	svc := newPatientServiceWithMocks(mockPatientRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.GET("/patient/search", SearchPatients(svc))

	req := httptest.NewRequest("GET", "/patient/search", nil)
	req.Header.Set("Content-Type", "application/json")

	c, _ := gin.CreateTestContext(recorder)
	c.Request = req
	c.Set("hospital_id", 5)
	c.Set("staff_id", 1)

	SearchPatients(svc)(c)

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

	mockPatientRepo.AssertExpectations(t)
}

// ========== SearchPatientByID Tests ==========

// TestSearchPatientByID_Success tests successful patient lookup by ID
func TestSearchPatientByID_Success(t *testing.T) {
	mockPatientRepo := new(mocks.IPatientRepo)
	

	// Setup expectations
	firstName := "John"
	lastName := "Doe"
	nationalID := "1234567890123"
	dateOfBirth := time.Date(1990, 1, 15, 0, 0, 0, 0, time.UTC)

	patient := &model.Patient{
		ID:          1,
		HospitalID:  5,
		FirstNameEn: &firstName,
		LastNameEn:  &lastName,
		NationalID:  &nationalID,
		DateOfBirth: &dateOfBirth,
	}

	mockPatientRepo.On("GetPatientByID", "1234567890123").
		Return(patient, nil)

	svc := newPatientServiceWithMocks(mockPatientRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.GET("/patient/search/:id", SearchPatientByID(svc))

	req := httptest.NewRequest("GET", "/patient/search/1234567890123", nil)
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
		t.Errorf("expected data to contain patient")
	}

	mockPatientRepo.AssertExpectations(t)
}

// TestSearchPatientByID_PatientNotFound tests 404 when patient not found
func TestSearchPatientByID_PatientNotFound(t *testing.T) {
	mockPatientRepo := new(mocks.IPatientRepo)
	

	// Setup expectations - return nil for not found
	mockPatientRepo.On("GetPatientByID", "nonexistent").
		Return(nil, nil)

	svc := newPatientServiceWithMocks(mockPatientRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.GET("/patient/search/:id", SearchPatientByID(svc))

	req := httptest.NewRequest("GET", "/patient/search/nonexistent", nil)
	engine.ServeHTTP(recorder, req)

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

	if resp.Error == nil || resp.Error.Code != "PATIENT_NOT_FOUND" {
		t.Errorf("expected error code PATIENT_NOT_FOUND, got %v", resp.Error)
	}

	mockPatientRepo.AssertExpectations(t)
}

// TestSearchPatientByID_EmptyID tests error for missing ID parameter
func TestSearchPatientByID_EmptyID(t *testing.T) {
	mockPatientRepo := new(mocks.IPatientRepo)
	

	svc := newPatientServiceWithMocks(mockPatientRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.GET("/patient/search/:id", SearchPatientByID(svc))

	// Request without ID parameter - Gin will return empty string
	req := httptest.NewRequest("GET", "/patient/search/", nil)
	engine.ServeHTTP(recorder, req)

	// Gin doesn't match the route if no ID is provided, so it returns 404 from Gin
	if recorder.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", recorder.Code)
	}
}

// TestSearchPatientByID_DatabaseError tests error handling for database failures
func TestSearchPatientByID_DatabaseError(t *testing.T) {
	mockPatientRepo := new(mocks.IPatientRepo)
	

	// Setup expectations - return error
	mockPatientRepo.On("GetPatientByID", "1234567890123").
		Return(nil, errors.New("database connection error"))

	svc := newPatientServiceWithMocks(mockPatientRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.GET("/patient/search/:id", SearchPatientByID(svc))

	req := httptest.NewRequest("GET", "/patient/search/1234567890123", nil)
	engine.ServeHTTP(recorder, req)

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

	mockPatientRepo.AssertExpectations(t)
}

// TestSearchPatientByID_PassportID tests search by passport ID
func TestSearchPatientByID_PassportID(t *testing.T) {
	mockPatientRepo := new(mocks.IPatientRepo)
	

	// Setup expectations
	firstName := "Jane"
	passportID := "AB123456"

	patient := &model.Patient{
		ID:          2,
		HospitalID:  5,
		FirstNameEn: &firstName,
		PassportID:  &passportID,
	}

	mockPatientRepo.On("GetPatientByID", "AB123456").
		Return(patient, nil)

	svc := newPatientServiceWithMocks(mockPatientRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.GET("/patient/search/:id", SearchPatientByID(svc))

	req := httptest.NewRequest("GET", "/patient/search/AB123456", nil)
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

	mockPatientRepo.AssertExpectations(t)
}

// TestSearchPatientByID_WithHyphens tests ID with hyphens
func TestSearchPatientByID_WithHyphens(t *testing.T) {
	mockPatientRepo := new(mocks.IPatientRepo)
	

	// Setup expectations with hyphens in ID
	firstName := "John"
	hyphenID := "123-456-7890"

	patient := &model.Patient{
		ID:          3,
		HospitalID:  5,
		FirstNameEn: &firstName,
	}

	mockPatientRepo.On("GetPatientByID", hyphenID).
		Return(patient, nil)

	svc := newPatientServiceWithMocks(mockPatientRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.GET("/patient/search/:id", SearchPatientByID(svc))

	req := httptest.NewRequest("GET", "/patient/search/"+hyphenID, nil)
	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", recorder.Code)
	}

	mockPatientRepo.AssertExpectations(t)
}

// TestSearchPatientByID_WithNullableFields tests patient with nullable fields
func TestSearchPatientByID_WithNullableFields(t *testing.T) {
	mockPatientRepo := new(mocks.IPatientRepo)
	

	// Setup expectations - patient with nil fields
	patient := &model.Patient{
		ID:          1,
		HospitalID:  5,
		FirstNameEn: nil, // Nullable field
		LastNameEn:  nil,
		PhoneNumber: nil,
		Email:       nil,
		DateOfBirth: nil,
	}

	mockPatientRepo.On("GetPatientByID", "1234567890123").
		Return(patient, nil)

	svc := newPatientServiceWithMocks(mockPatientRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.GET("/patient/search/:id", SearchPatientByID(svc))

	req := httptest.NewRequest("GET", "/patient/search/1234567890123", nil)
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

	mockPatientRepo.AssertExpectations(t)
}

// TestSearchPatientByID_LongID tests search with a very long ID string
func TestSearchPatientByID_LongID(t *testing.T) {
	mockPatientRepo := new(mocks.IPatientRepo)
	

	// Create a long ID
	longID := "123456789012345678901234567890123456789012345678901234567890"
	firstName := "John"

	patient := &model.Patient{
		ID:          99,
		HospitalID:  5,
		FirstNameEn: &firstName,
	}

	mockPatientRepo.On("GetPatientByID", longID).
		Return(patient, nil)

	svc := newPatientServiceWithMocks(mockPatientRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.GET("/patient/search/:id", SearchPatientByID(svc))

	req := httptest.NewRequest("GET", "/patient/search/"+longID, nil)
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

	mockPatientRepo.AssertExpectations(t)
}

// TestSearchPatientByID_ResponseStructure tests complete response structure
func TestSearchPatientByID_ResponseStructure(t *testing.T) {
	mockPatientRepo := new(mocks.IPatientRepo)
	

	// Setup expectations with comprehensive patient data
	firstName := "John"
	lastName := "Doe"
	nationalID := "1234567890123"
	passportID := "AB123456"
	phoneNumber := "0812345678"
	email := "john@example.com"
	gender := "M"
	dateOfBirth := time.Date(1990, 1, 15, 0, 0, 0, 0, time.UTC)

	patient := &model.Patient{
		ID:          1,
		HospitalID:  5,
		FirstNameEn: &firstName,
		LastNameEn:  &lastName,
		NationalID:  &nationalID,
		PassportID:  &passportID,
		PhoneNumber: &phoneNumber,
		Email:       &email,
		Gender:      &gender,
		DateOfBirth: &dateOfBirth,
	}

	mockPatientRepo.On("GetPatientByID", "1234567890123").
		Return(patient, nil)

	svc := newPatientServiceWithMocks(mockPatientRepo)
	engine, _ := setupGinContext(t)
	recorder := httptest.NewRecorder()

	engine.GET("/patient/search/:id", SearchPatientByID(svc))

	req := httptest.NewRequest("GET", "/patient/search/1234567890123", nil)
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

	if resp.Error != nil {
		t.Errorf("unexpected error in response: %v", resp.Error)
	}

	// Verify response data is present (contains patient object)
	if resp.Data == nil {
		t.Errorf("expected data to contain patient information")
	}

	// Verify response structure has expected patient data
	respBytes, _ := json.Marshal(resp.Data)
	if len(respBytes) == 0 {
		t.Errorf("expected patient data in response")
	}

	mockPatientRepo.AssertExpectations(t)
}
