# Hospital API — README

A production-ready REST API for hospital staff to search patient records with hospital-level access control and authentication.

**Tech Stack**: Go 1.25.0 | PostgreSQL 16 | Gin Framework | JWT Auth | Docker | Nginx

---

## Project Overview

The Hospital API implements a multi-tenant architecture where:
- Each **hospital** is isolated (no cross-hospital data leakage)
- **Staff** users authenticate via JWT and can search patient records
- **Patients** are searchable by 8 fields (name, ID, contact, etc.)
- All queries execute via authenticated endpoints with <200ms response time (SLA)

See [docs/PRD.md](docs/PRD.md) for full requirements and [docs/techdoc.md](docs/techdoc.md) for architecture details.

---

## Quick Start

### Prerequisites
- Docker & Docker Compose (recommended for full stack)
- Go 1.25.0+ (if running API locally)
- `make` command-line tool

### Option 1: Full Stack with Docker Compose (Recommended)

```bash
# 1. Clone and setup environment
git clone <repo>
cd hospital-api
cp .env.example .env  # Edit with your configs

# 2. Start entire stack (postgres → migrate → api → nginx)
make docker-up

# 3. Verify health (via nginx)
curl http://localhost/api/health

# 4. Or access API directly
curl http://localhost:8080/api/health

# 5. Create staff user and test
curl -X POST http://localhost:8080/api/staff/create \
  -H "Content-Type: application/json" \
  -d '{"username": "doctor", "password": "secure_password", "hospital_id": 1}'

# 6. View logs
make docker-logs
```

**What runs**:
- Nginx reverse proxy (port 80) → API backend
- Go API (port 8080 - directly exposed)
- PostgreSQL database (port 5432 for local dev)
- Database migrations (auto-run on startup)

**Stop everything**:
```bash
make docker-down
# Also remove database volume (clean restart):
# docker-compose down -v
```

---

### Option 2: API Only (Local Development)

For faster iteration during development, run Postgres in Docker but the API locally:

```bash
# 1. Start database only
docker-compose up postgres migrate

# 2. In another terminal, run API locally
make run
# or: APP_PORT=8080 DB_HOST=localhost go run ./cmd/api

# 3. Access API directly (no nginx needed)
curl http://localhost:8080/api/health
```

---

### Option 3: Fully Local (No Docker)

Requires PostgreSQL installed locally:

```bash
# 1. Start PostgreSQL (assumes local install)
# On macOS: brew services start postgresql
# On Linux: sudo systemctl start postgresql

# 2. Create database and load environment
createdb hospital_db
source .env  # or: set -a; . .env; set +a

# 3. Run migrations
make migrate-up

# 4. Run API
make run

# 5. Access API
curl http://localhost:8080/api/health
```

---

## Common Commands

### Development & Testing
```bash
make build              # Compile binary to bin/api
make run               # Build and run locally
make test              # Run unit tests with coverage
make dev               # Run with auto-reload (requires 'air' installed)
```

### Database
```bash
make migrate-up        # Apply all pending migrations
make migrate-down      # Revert last migration
make db-up             # Start PostgreSQL container
make db-down           # Stop PostgreSQL container
make db-reset          # Stop + remove volume (clean slate)
make db-logs           # View database logs
```

### Docker Compose (Full Stack)
```bash
make docker-up                 # Start all services (postgres → migrate → api → nginx)
make docker-down               # Stop all services
make docker-logs               # View all services logs

# Or use docker-compose directly
docker-compose up              # Start all services
docker-compose up --build      # Rebuild images and start
docker-compose ps              # List running services
```

---

## API Endpoints

### Authentication
- `POST /api/staff/create` — Create new staff user
- `POST /api/staff/login` — Login and retrieve JWT token

### Patient Search
- `GET /api/patient/search/:id` — Search patient by ID (no auth required)
- `GET /api/patient/search` — Search patients with filters (requires auth)
  - Query params: `first_name_en`, `last_name_en`, `national_id`, `passport_id`, `phone_number`, `email`, `dob`, `gender`

### Health
- `GET /api/health` — Unauthenticated health check

Full API documentation available in [docs/techdoc.md](docs/techdoc.md#7-handler--dtos).

---

## Configuration

### Environment Variables

Create a `.env` file in the project root:

```bash
# Application
APP_NAME=hospital-api
APP_ENV=development
APP_PORT=8080

# Database (Docker: use 'postgres' as host; Local: use 'localhost')
DB_HOST=postgres
DB_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=<your_secure_password>
POSTGRES_DB=hospital_db
DATABASE_DSN=postgres://postgres:password@postgres:5432/hospital_db?sslmode=disable

# JWT
JWT_SECRET=<your_secure_jwt_secret>
JWT_EXPIRATION_DAYS=30

# Logging
LOG_LEVEL=info
LOG_FORMAT=json

# Rate Limiting
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW_SECONDS=60
```

For Docker Compose: Set `DB_HOST=postgres` (service name). For local dev: Set `DB_HOST=localhost`.

---

## Project Structure

```
hospital-api/
├── cmd/api/              # Application entrypoint
├── internal/
│   ├── config/          # Environment & configuration parsing
│   ├── database/        # Database connection, query helpers
│   ├── handler/         # HTTP handlers & DTOs
│   ├── service/         # Business logic & transactions
│   ├── repository/      # Data access & SQL
│   ├── model/           # Domain models
│   ├── router/          # Gin router setup
│   ├── middleware/      # Auth, logging middleware
│   └── logger/          # Structured logging
├── nginx/               # Nginx reverse proxy config
├── migrations/          # SQL migration files
├── docs/
│   ├── PRD.md          # Product Requirements Document
│   └── techdoc.md      # Technical architecture & runbook
├── Dockerfile          # Multi-stage build for API
├── docker-compose.yml  # Full stack orchestration
├── Makefile            # Development shortcuts
└── go.mod, go.sum      # Go dependencies
```

---

## Architecture

The application follows a **layered architecture** with strict separation of concerns:

```
HTTP Request
     ↓
[Nginx Reverse Proxy] ← port 80
     ↓
[Gin Handler] ← validates transport + JWT auth
     ↓
[Service] ← orchestrates business logic + transactions
     ↓
[Repository] ← SQL queries + error translation
     ↓
[PostgreSQL Database]
```

See [docs/techdoc.md](docs/techdoc.md) for complete architecture details.

---

## Testing

Run unit and integration tests:

```bash
# Run all tests
make test

# Run specific test file
go test -v ./internal/handler -run TestPatientSearch

# View coverage report
go tool cover -html=coverage.out
```

Tests use interface-based mocks for repository layer. See [internal/handler/patient_test.go](internal/handler/patient_test.go) for examples.

---

## Security Notes

- **Authentication**: JWT (HS256, 30-day expiry) with `staff_id` and `hospital_id` in claims
- **Password Hashing**: bcrypt with default cost
- **Multi-tenancy**: All queries enforce `hospital_id` from JWT claims — zero cross-hospital data leakage
- **Production**: Run behind TLS-terminating reverse proxy (nginx with SSL), use secure secret manager for `JWT_SECRET`

See [docs/techdoc.md#6-authentication--authorization](docs/techdoc.md#6-authentication--authorization) for details.

---

## Troubleshooting

### Docker Compose won't start API
```bash
# Check if Docker daemon is running
docker ps

# View full error logs
docker-compose logs api

# Rebuild images if dependencies changed
docker-compose up --build
```

### Database connection error
```bash
# Verify postgres is healthy
docker-compose exec postgres pg_isready -U postgres

# Check DATABASE_DSN in .env
# Docker: postgres://<user>:<pass>@postgres:5432/<db>?sslmode=disable
# Local:  postgres://<user>:<pass>@localhost:5432/<db>?sslmode=disable
```

### Port 80 already in use
```bash
# Use a different port for nginx
# Edit docker-compose.yml: ports: ["8000:80"]
# Then: curl http://localhost:8000/api/health
```

---

## Contributing

1. Check [docs/techdoc.md](docs/techdoc.md) for code patterns and conventions
2. Run tests before committing: `make test`
3. Ensure all endpoints are authenticated (except `/api/health`)
4. Add unit tests for new handlers/services

---

## References

- Product Requirements: [docs/PRD.md](docs/PRD.md)
- Technical Runbook: [docs/techdoc.md](docs/techdoc.md)
- Go Modules: See `go.mod` for dependencies

