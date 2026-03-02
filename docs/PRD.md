# **Hospital Patient Search API - Product Requirements Document**

## **1. Executive Summary**

### Problem Statement
Hospital staff currently lack a centralized system to search patient data. Patient information is fragmented across multiple systems, forcing staff to manually gather data from various sources, resulting in delays and inefficiency during patient lookup operations.

### Proposed Solution
Build a simple, secure REST API using Gin Framework (Go) and PostgreSQL that allows hospital staff to search patient records integrated from other systems. The API will enforce role-based access control where each staff member can only access patients from their assigned hospital.

### Success Criteria
1. **Response Time**: Search queries return results within 200ms for 100+ patient datasets
2. **Data Availability**: Support searching across all 8 searchable patient fields (national_id, passport_id, first_name, middle_name, last_name, date_of_birth, phone_number, email)
3. **Security**: 100% of patient queries authenticated and filtered by hospital_id from JWT token
4. **User Management**: Staff self-registration and secure login operational on Day 1
5. **Multi-Tenancy**: Zero cross-hospital data leakage (staff can only view their hospital's patients)

---

## **2. User Experience & Functionality**

### User Personas
- **Admin Staff**: Hospital administrators who can register to the system and search patient records.

### User Stories

**Story 1: Staff Registration**
```
As a hospital staff,
I want to register my self,
so that I can access the patient search system.
```
**Acceptance Criteria:**
- Registration endpoint accepts username, password, and hospital_id
- Password validation enforces minimum 8 characters
- Duplicate username registration is rejected with 400 error
- Staff account is immediately usable after registration

**Story 2: Staff Login**
```
As a staff member,
I want to log in with my credentials,
so that I can access the patient search system securely.
```
**Acceptance Criteria:**
- Login endpoint accepts username and password
- Valid credentials return JWT token valid for 30 days
- JWT token contains staff ID, hospital_id
- Invalid credentials return 401 Unauthorized
- Token includes hospital_id for all subsequent authorization checks

**Story 3: Patient Search**
```
As a staff member,
I want to search patients by their details (national_id, passport_id, first_name, middle_name, last_name, date_of_birth, phone_number, email),
so that I can quickly find patient records without manual lookup.
```
**Acceptance Criteria:**
- GET `/api/patient/search` endpoint supports query parameters for all 8 searchable fields
- Supports partial matching (e.g., search "John" finds "John Smith") for first_name, middle_name, last_name, phone_number and email
- Single hospital_id filter applied automatically from JWT token (no cross-hospital access)
- Returns results in JSON format with patient ID, national_id, passport_id, first_name_th, middle_name_th, last_name_th, first_name_en, middle_name_en, last_name_en, date_of_birth, phone_number, email
- Response time < 200ms for query against 100+ patient dataset
- Unauthenticated requests return 401 Unauthorized
- Staff from Hospital A cannot search Hospital B's patients (enforced by hospital_id)

### Non-Goals
- Patient history or detailed medical records (scope limit)
- Patient data modification or deletion (read-only on Day 1)
- Advanced filtering (date ranges, complex boolean queries)
- Export/reporting features
- Appointment scheduling integration (deferred)
- Real-time sync from external systems (use scheduled API integration)

---

## **3. Technical Specifications**

### Architecture Overview
```
┌─────────────────────────────────────────────────────────┐
│  Gin REST API Server (Go)                              │
│  ├─ Authentication Service (JWT)                        │
│  ├─ Staff Registration & Login Routes                   │
│  ├─ Patient Search Service (Query Handler)              │
│  └─ Authorization Middleware (JWT + hospital_id check)  │
└──────────────────┬──────────────────────────────────────┘
                   │
                   │ SQL Queries
                   ▼
┌──────────────────────────────────────────────────────────┐
│  PostgreSQL Database                                     │
│  ├─ staff (id, hospital_id, username, password)          │
│  ├─ patients (id, national_id, passport_id, hospital_id, |
|  |           first_name_th, middle_name_th, last_name_th,│
|  |           first_name_en, middle_name_en, last_name_en,|
│  │           date_of_birth, phone_number, email, gender) │
│  └─ hospital (id, name)                                  │
└──────────────────────────────────────────────────────────┘

Data Flow (Search):
  Staff → Login → Get JWT + hospital_id
       → Search Request + JWT 
       → Middleware validates JWT & extracts hospital_id
       → Query Builder adds WHERE hospital_id = :hospital_id
       → PostgreSQL returns filtered results
       → Response (200ms max)
```

### Database Schema
**Staff Table**
```sql
CREATE TABLE staff (
  id SERIAL PRIMARY KEY,
  hospital_id INTEGER NOT NULL REFERENCES hospital(id),
  username VARCHAR(255) UNIQUE NOT NULL,
  password VARCHAR(255) NOT NULL,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_staff_hospital_id ON staff(hospital_id);
```

**Patients Table**
```sql
CREATE TABLE patients (
  id SERIAL PRIMARY KEY,
  national_id VARCHAR(50) UNIQUE,
  passport_id VARCHAR(50) UNIQUE,
  first_name_th VARCHAR(255),
  middle_name_th VARCHAR(255),
  last_name_th VARCHAR(255),
  first_name_en VARCHAR(255),
  middle_name_en VARCHAR(255),
  last_name_en VARCHAR(255),
  date_of_birth DATE,
  phone_number VARCHAR(20) UNIQUE,
  email VARCHAR(255) UNIQUE,
  gender VARCHAR(1),
  hospital_id INTEGER NOT NULL REFERENCES hospital(id),
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  CHECK (national_id IS NOT NULL OR passport_id IS NOT NULL)
);

CREATE INDEX idx_patients_hospital_id ON patients(hospital_id);
CREATE INDEX idx_patients_dob ON patients(date_of_birth);
```

**Hospital Table**
```sql
CREATE TABLE hospital (
  id SERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
);
```

### API Endpoints

**POST /api/staff/create**
- Request body: `{ username:string, password:string, hospital_id:integer }`
- Response: `{"success":boolean,"data":{"id":integer,"username":string,"hospital_id":integer,"created_at":string datetime}}`

**POST /api/staff/login**
- Request body: `{ username:string, password:string }`
- Response: `{"success":boolean,"data":{"access_token":string,"token_type":string,"refresh_token":string,"expires_in":integer}}`

**GET /api/patient/search**
- Query params: `national_id`, `passport_id`, `first_name`, `middle_name`, `last_name`, `date_of_birth`, `phone_number`, `email`
- Header: `Authorization: Bearer <jwt_token>`
- Response: `{"success":boolean,"data":[{"id":integer,"national_id":string,"passport_id":string,"first_name_th":string,"middle_name_th":string,"last_name_th":string,"first_name_en":string,"middle_name_en":string,"last_name_en":string,"date_of_birth":string,"phone_number":string,"email":string}]}`

**GET /api/patient/search/:id**
- Path param: `id` (patient ID)
- Response: Same as list endpoint but returns single patient object

### Integration Points
- **Authentication**: JWT-based with hospital_id claim for authorization
- **Database**: PostgreSQL connection pooling with optimized indexes for 1,000+ patient queries

### Security & Privacy
- **Authentication**: JWT tokens with 30-day expiration
- **Authorization**: Hospital-level isolation enforced by middleware on every request
- **Password Security**: bcrypt hashing with salt before storage
- **Data Transport**: All APIs should use HTTPS in production
- **Hospital Isolation**: Every query includes `WHERE hospital_id = ?` filter
- **No Data Exposure**: Error messages don't leak patient existence ("Patient not found" doesn't distinguish between "doesn't exist" vs "you don't have access")

---

## **4. Risks & Roadmap**

### Phase 1: MVP (Day 1)
- [x] Staff registration and login
- [x] Patient search API (all 8 fields)
- [x] Hospital-level access control via JWT
- [x] < 200ms response time

### Phase 2: v1.1 (Week 2)
- Add Dockerfile to setup the server (nginx) for the project
- Implement swaggo/gin-swagger for api documentation
- Add makefile
- Add readme (How to run, How to test, API Example curl)

### Phase 3: v2.0 (Month 2)
- Add graceful shutdown
- Add structured logging using logrus
- Add X-Request-ID on response header
- Audit logging for patient searches
- Rate limiting to prevent abuse

### Technical Risks
| Risk | Probability | Mitigation |
|------|-------------|-----------|
| **1-day deadline pressure** | High | Use Gin boilerplate, pre-built auth libraries, skip non-MVP features |
| **Search latency > 200ms** | Medium | Index all searchable fields, pagination (limit 50 results default), PostgreSQL query optimization |
| **Cross-hospital data leak** | High | Enforce hospital_id in middleware for ALL queries, comprehensive unit tests for auth logic |
| **Database performance at scale (1000+ patients)** | Medium | Use connection pooling, add table indexes, monitor slow queries |

---

## **5. Success Metrics & Testing**

### Evaluation Strategy
1. **Authentication Tests**: 
   - Verify JWT token includes hospital_id ✓
   - Test staff from Hospital A cannot query Hospital B ✓
   - Verify token expiration (30 days)

2. **Search Performance Tests**:
   - Run search against 1,000 patient records, measure response time (target: < 200ms)
   - Test all 8 fields individually and combined
   - Verify results are hospital-isolated

3. **Functional Tests**:
   - Register staff → Login → Search returns expected results
   - Invalid credentials return 401
   - Unauthenticated requests return 401

---

**Development Mode**: Day 1 focus on MVP. Quality over perfect compliance since "no compliance requirements" noted.
