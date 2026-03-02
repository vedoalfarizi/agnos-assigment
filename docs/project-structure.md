# Hospital API — Project Structure

## Overview

Hospital API is a REST API built with **Go 1.25.0**, **Gin** framework, and **PostgreSQL**. It follows a **clean layered architecture** separating concerns across handlers, services, repositories, and models.

```
HTTP Request → Handler → Service → Repository → Database
```

---

## Directory Layout

### `/cmd/api` — Entry Point
Application startup and bootstrapping.

- `main.go` - Entry point, signal handling, graceful shutdown
- `bootstrap.go` - Initializes config, database, logger, and router

### `/internal/handler` — HTTP Layer
Handlers parse requests, call services, and format responses.

- `health_check.go` - Health status endpoint
- `patient.go` - Patient search endpoints
- `staff.go` - Staff authentication (create, login)
- `response.go` - Response utilities
- `*_test.go` - Handler tests

### `/internal/service` — Business Logic
Services orchestrate repositories and implement business rules.

- `health.go` - Database connectivity checks
- `patient.go` - Patient search logic
- `staff.go` - Staff authentication, JWT generation

### `/internal/repository` — Data Access
Repositories execute SQL queries and map results to models.

- `health.go` - Database health queries
- `hospital.go` - Hospital data queries
- `patient.go` - Patient queries and search
- `staff.go` - Staff user queries
- `*_test.go` - Repository tests (using `go-sqlmock`)

### `/internal/model` — Domain Models
Pure data structures directly mapped to database columns.

- `patient.go` - Patient entity with all DB fields
- `staff.go` - Staff user entity

**Example**: Uses pointer fields for nullable columns:
```go
type Patient struct {
    ID          int        `db:"id"`
    FirstNameTh *string    `db:"first_name_th"`
    DateOfBirth *time.Time `db:"date_of_birth"`
}
```

### `/internal/dto` — Response Objects
Data Transfer Objects shape API responses (different from models).

- `patient.go` - Patient response structures (client-facing)
- `staff.go` - Staff response structures

**Difference**: Models include all DB fields; DTOs include only fields sent to clients.

### `/internal/middleware` — HTTP Middleware
Cross-cutting concerns applied to all or specific routes.

- `auth.go` - JWT token validation
- `logging.go` - Request/response logging
- `request_context.go` - Adds request ID for tracing

### `/internal/logger` — Logging
Structured logging using `logrus`.

- `logger.go` - Logger initialization and utilities
- Functions: `Init()`, `Infof()`, `Errorf()`, `ErrorfWithContext()`

### `/internal/config` — Configuration
Loads environment variables into a Config struct at startup.

- `config.go` - Config struct and loading logic
- Fields: AppPort, DB credentials, JWT secret, LogLevel, etc.

### `/internal/database` — Database Connection
Manages PostgreSQL connectivity.

- `postgre/sqlx.go` - Connection pool setup with `sqlx`
- Functions: `Connect(dsn)`, `GetDB()`

### `/internal/router` — Route Registration
Sets up Gin routes and dependency injection.

- `router.go` - Creates Gin engine, registers all routes
- Injects services into handlers via constructor functions

### `/migrations` — Database Schema
SQL migration files for schema versioning.

```
1741900800_create_hospitals.up.sql     # Apply
1741900800_create_hospitals.down.sql   # Rollback
1741900860_create_staff.up.sql
1741900860_create_staff.down.sql
1741900920_create_patients.up.sql
1741900920_create_patients.down.sql
```

Each migration has an `up` (apply) and `down` (rollback) variant.

### `/mocks` — Test Mocks
Auto-generated mock implementations of repository interfaces.

- `IHospitalRepo.go` - Mock hospital repository
- `IPatientRepo.go` - Mock patient repository
- `IStaffRepo.go` - Mock staff repository

Generated via: `mockery --name=IPatientRepo --output=mocks`

### `/docs` — Documentation
- `openapi.yaml` - OpenAPI 3.0 specification (machine-readable API schema)
- `Agnos Health.postman_collection.json` - Postman collection (ready-to-import for testing)
- `PRD.md` - Product requirements and user stories
- `techdoc.md` - Technical architecture details
- `project-structure.md` - This file

### Root Level
- `go.mod` - Go dependencies
- `Makefile` - Build, test, run, migration commands
- `Dockerfile` - Container image
- `docker-compose.yml` - Multi-container setup (postgres, api, nginx)
- `nginx/nginx.conf` - Reverse proxy config
- `README.md` - Quick start guide
