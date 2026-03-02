# Hospital API — Technical Runbook

## Purpose & Scope
This document is an execution-ready runbook for implementing and operating the Hospital API. It merges architecture, implementation notes, database schema and operational steps so an engineer can implement, test, and deploy the service.

Intended audience: Go backend engineers implementing handlers, services, repositories, and operations engineers running migrations and CI.


## 1. High-Level Architecture
- Layered structure: Handler → Service → Repository → Database. Handlers handle transport/validation, services orchestrate business logic and transactions, repositories perform SQL and map DB errors to domain errors.
- Keep strict import boundaries: handlers must not import repositories directly; they should depend on service interfaces.
- Transaction boundary: Service layer controls transactions. Repositories accept `*sqlx.DB` or `*sqlx.Tx` for operations within transactions.


## 2. Project Layout (canonical)

## Health Check Endpoint

A simple unauthenticated endpoint exposed at `/api/health` (within the API route group).
The handler should return both overall service status and component statuses for easier operational checks.
Example response structure:

```json
{
  "status": "healthy",
  "server": "up",
  "database": "up"
}
```

In case of database connectivity issues the HTTP status code should be 500 and the response `status` field set to `unhealthy`; the failure should be logged at error level.

The handler constructs responses via `internal/handler/HealthCheckResponse` and leverages the service layer (not direct DB calls).

- `cmd/` — application entrypoint(s)
- `internal/config` — env parsing and app config
- `internal/database/postgre` — DB connect, migrations helper
- `internal/logger` — structured logger initialization
- `internal/router` — Gin router setup
- `internal/handler` — HTTP handlers / DTOs
- `internal/service` — business logic, transactions
- `internal/repository` — SQL queries and error translation
- `internal/model` — domain models
- `internal/dto` — transport DTOs and validation
- `internal/middleware` — auth, logging middleware
- `migrations/` — SQL migration files
- `Dockerfile`, `docker-compose.yml`, `Makefile`


## 3. Environment variables (.env conventions)
Use a `.env` file in development. Canonical environment variables:

**Application:**
- `APP_NAME` — Application name (default: `hospital-api`)
- `APP_ENV` — Environment mode: `development` or `production` (default: `development`)
- `APP_PORT` — Server port (default: `8080`)

**Database (PostgreSQL):**
- `DB_HOST` — PostgreSQL host (default: `localhost`)
- `DB_PORT` — PostgreSQL port (default: `5432`)
- `DB_SSLMODE` — SSL mode (default: `disable`)
- `DB_MAX_OPEN_CONNS` — Max open connections (default: `25`)
- `DB_MAX_IDLE_CONNS` — Max idle connections (default: `10`)
- `DB_CONN_MAX_LIFETIME` — Connection max lifetime (default: `5m`)
- `POSTGRES_USER` — PostgreSQL username (default: `postgres`)
- `POSTGRES_PASSWORD` — PostgreSQL password
- `POSTGRES_DB` — PostgreSQL database name (default: `hospital_db`)
- `DATABASE_DSN` — Full PostgreSQL DSN for migrations, example: `postgres://user:pass@localhost:5432/hospital?sslmode=disable`

**JWT:**
- `JWT_SECRET` — JWT HMAC secret (base64 or raw string, required)
- `JWT_EXPIRATION_DAYS` — JWT token expiration in days (default: `30`)

**Logging:**
- `LOG_LEVEL` — Log level: `debug`, `info`, `warn`, `error` (default: `info`)
- `LOG_FORMAT` — Log format: `json` or `text` (default: `json`)

**Rate Limiting:**
- `RATE_LIMIT_ENABLED` — Enable rate limiting (default: `true`)
- `RATE_LIMIT_REQUESTS` — Requests per window (default: `100`)
- `RATE_LIMIT_WINDOW_SECONDS` — Rate limit window in seconds (default: `60`)


## 4. Database — Migrations & Schema
- Place SQL migration files in `migrations/` as sequential files, e.g. `001_create_hospitals.sql`, `002_create_staff.sql`, `003_create_patients.sql`.
- Required unique constraints (apply in migration SQL):
	- Patients: `national_id`, `passport_id`, `phone_number`, `email` (enforce according to locale; nullable columns should use partial-unique indexes as needed)
	- Staff: `username`
- Foreign keys should reference `hospital_id` where applicable.

Migrations are executed with `golang-migrate` (recommended) from a separate migration service. Example command:

```bash
migrate -path ./migrations -database "$DATABASE_DSN" up
```

Dev convenience (Makefile target):

```makefile
# Makefile snippet (add to Makefile)
migrate-up:
	docker run --rm -v "$(PWD)/migrations":/migrations --network host migrate/migrate \
	  -path=/migrations -database "$(DATABASE_DSN)" up

migrate-down:
	docker run --rm -v "$(PWD)/migrations":/migrations --network host migrate/migrate \
	  -path=/migrations -database "$(DATABASE_DSN)" down
```

Dev note: Prefer running the migrator as a separate Docker container in local development and CI. Optionally, enable `MIGRATE_AUTO=true` in dev to auto-apply pending migrations at startup (not recommended for production).


## 5. Database Connection & Wrapper (sqlx)
Use `github.com/jmoiron/sqlx` for DB access. Provide a small wrapper in `internal/database/postgre/sqlx.go` that exposes a `Connect` and `GetDB()` function.

Example:

```go
package postgre

import (
		"os"
		"time"

		"github.com/jmoiron/sqlx"
		_ "github.com/lib/pq"
)

var db *sqlx.DB

func Connect(dsn string) (*sqlx.DB, error) {
		d, err := sqlx.Open("postgres", dsn)
		if err != nil { return nil, err }
		d.SetMaxOpenConns(25)
		d.SetMaxIdleConns(5)
		d.SetConnMaxLifetime(5 * time.Minute)
		db = d
		return db, nil
}

func GetDB() *sqlx.DB { return db }
```

Call `postgre.Connect(os.Getenv("DATABASE_DSN"))` during app startup.


## 6. Authentication & Authorization
- Password hashing: use `golang.org/x/crypto/bcrypt`.
	- Registration: `hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)`; store `hash` in DB.
	- Login: `err := bcrypt.CompareHashAndPassword(storedHash, []byte(password))`
- JWT: Use `github.com/golang-jwt/jwt` with HS256 and a 30-day expiry. Include `staff_id` and `hospital_id` in claims.

Claims example and token generator:

```go
type Claims struct {
	StaffID   int `json:"staff_id"`
	HospitalID int `json:"hospital_id"`
	jwt.StandardClaims
}

func GenerateToken(staffID, hospitalID int, secret []byte) (string, error) {
	claims := Claims{
		StaffID: staffID,
		HospitalID: hospitalID,
		StandardClaims: jwt.StandardClaims{ExpiresAt: time.Now().Add(30*24*time.Hour).Unix()},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}
```

Middleware example (Gin) — place in `internal/middleware/auth.go`:

```go
func AuthMiddleware(secret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		hdr := c.GetHeader("Authorization")
		if !strings.HasPrefix(hdr, "Bearer ") { c.AbortWithStatusJSON(401, gin.H{"error":"missing token"}); return }
		tokenStr := strings.TrimPrefix(hdr, "Bearer ")
		token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) { return secret, nil })
		if err != nil || !token.Valid { c.AbortWithStatusJSON(401, gin.H{"error":"invalid token"}); return }
		claims := token.Claims.(*Claims)
		c.Set("hospital_id", claims.HospitalID)
		c.Set("staff_id", claims.StaffID)
		c.Next()
	}
}
```

Handlers must enforce that `hospital_id` from token matches resource hospital_id where applicable.


## 7. Handler & DTOs
- DTOs must have JSON tags and validation tags (`github.com/go-playground/validator/v10`).
- Validate transport-level constraints in handlers and business rules in services.

API response helpers (put in `internal/handler/response.go`):

```go
type ErrorResp struct { Code string `json:"code"` Message string `json:"message"` }
type APIResponse struct { Success bool `json:"success"` Data interface{} `json:"data,omitempty"` Error *ErrorResp `json:"error,omitempty"` }

func Success(c *gin.Context, data interface{}) { c.JSON(200, APIResponse{Success:true, Data:data}) }
func Fail(c *gin.Context, code int, err *ErrorResp) { c.JSON(code, APIResponse{Success:false, Error:err}) }
```


## 8. Repository & Query Patterns
- Use `sqlx` with raw SQL for clarity and performance.
- For partial name searches, use `ILIKE` to support case-insensitive matches:

```sql
SELECT id, first_name_en, last_name_en FROM patients
WHERE hospital_id = $1 AND (first_name_en ILIKE '%' || $2 || '%' OR last_name_en ILIKE '%' || $2 || '%')
LIMIT $3 OFFSET $4
```

- Translate DB-specific errors into domain-level errors in repository layer (e.g. `sql.ErrNoRows` -> `ErrNotFound`, unique constraint violation -> `ErrDuplicate`).


## 9. Transactions & Service patterns
- Service methods start transactions when operations span multiple repositories.
- Provide helpers: `func WithTx(db *sqlx.DB, fn func(tx *sqlx.Tx) error) error` to centralize tx commit/rollback semantics.


## 10. Error Mapping → HTTP Statuses
- NotFound -> 404
- Validation error / Bad request -> 400
- Business rule violation (duplicate) -> 409
- Unauthorized -> 401
- Forbidden -> 403
- Internal -> 500


## 11. Logging & Observability
- Use structured logs. Include `request_id`, `method`, `path`, `status`, `latency`.
- Never log raw passwords or full tokens. Redact sensitive fields in request/response logs.
- Future: add metrics and tracing hooks.


## 12. Security & Deployment Notes
- Run app behind a TLS-terminating reverse proxy (nginx) in production. Set `GIN_MODE=release` in prod.
- Manage `JWT_SECRET` securely (KMS or secret store in prod). Rotate secrets per policy.


## 12.5 Nginx & Docker Compose Infrastructure

### Architecture Overview
The application runs in a multi-container Docker setup orchestrated by `docker-compose.yml`:

```
┌─────────────┐
│   Nginx     │ (port 80:80) — Reverse proxy & API gateway
├─────────────┤
│     API     │ (internal :8080) — Go application
├─────────────┤
│  PostgreSQL │ (internal :5432) — Database
└─────────────┘
```

**Data Flow**: Client → Nginx (port 80) → API service (:8080) → PostgreSQL (:5432)

### Docker Compose Services

**1. PostgreSQL (postgres)**
- Image: `postgres:16`
- Health check: `pg_isready`
- Volume: `postgres_data` (persistent)
- Port: `:5432` (exposed for local dev debugging only)
- Env: `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB` from `.env`

**2. Database Migrator (migrate)**
- Image: `migrate/migrate`
- Runs migrations from `./migrations` directory
- Depends on postgres health check
- Execution: One-shot container (runs once, exits)

**3. API Service (api)**
- Build: Multi-stage Dockerfile (golang:1.25.0 → alpine:3.20)
- Env: Reads from `.env` file; sets `DB_HOST=postgres` for internal networking
- Depends on: postgres (healthy) + migrate (completed)
- Port: `:8080` internal (exposed only to nginx via docker-compose network)
- Health check: `wget http://localhost:8080/api/health`

**4. Nginx (nginx)**
- Image: `nginx:latest`
- Configuration: Mounted from `./nginx/nginx.conf`
- Port: `:80:80` (public-facing HTTP)
- Upstream: Routes to `api:8080` (docker-compose service DNS)
- Depends on: api (healthy)
- Health check: `wget http://localhost:80/api/health`

### Startup Sequence
```
1. Postgres starts (waits for health check ✓)
2. Migrate runs migrations (waits for postgres health check ✓)
3. API builds and starts (waits for postgres + migrate completion ✓)
4. Nginx starts and routes traffic (waits for api health check ✓)
```

### Nginx Configuration Details
- **Upstream block**: Resolves `api` to container DNS (docker-compose auto-manages)
- **Location blocks**:
  - `/api/health` — Health check endpoint (passthrough)
  - `/api/` — General API routing with X-Forwarded headers
  - `/` — Redirects to health endpoint
- **Headers preserved for API**:
  - `X-Real-IP` — Client IP
  - `X-Forwarded-For` — Proxy chain
  - `X-Forwarded-Proto` — Original scheme (future: HTTPS)
  - `X-Forwarded-Host` — Original host

### Development Workflow
```bash
# Start full stack: postgres → migrate → api → nginx
docker-compose up

# Access API
curl http://localhost/api/health

# View logs
docker-compose logs -f api
docker-compose logs -f nginx

# Rebuild API after code changes
docker-compose up --build api

# Stop everything
docker-compose down

# Reset database and restart
docker-compose down -v
docker-compose up
```

### Local Development Without Docker
For faster iteration, run the API locally:
```bash
# Start postgres & migrate only
docker-compose up postgres migrate

# In another terminal, run API locally
APP_PORT=8080 DB_HOST=localhost make run

# Nginx not needed for local dev; call API directly on :8080
curl http://localhost:8080/api/health
```

### Environment Variables
When running in Docker, ensure `.env` contains:
```bash
APP_PORT=8080                  # API internal port
DB_HOST=postgres              # Docker service name (not localhost!)
DB_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=<secure_password>
POSTGRES_DB=hospital_db
DATABASE_DSN=postgres://postgres:password@postgres:5432/hospital_db?sslmode=disable
JWT_SECRET=<secure_secret>
```

### Production Considerations
- Replace `nginx:latest` with versioned tag (e.g., `nginx:1.26-alpine`)
- Mount SSL certificates into nginx for HTTPS termination
- Use environment substitution (`.env` file) for secrets in docker-compose
- Set `APP_ENV=production` and `GIN_MODE=release` in prod `.env`
- Use a secrets manager (Docker Secrets, HashiCorp Vault) instead of `.env` in production
- Configure log aggregation (ELK, Datadog) for nginx and API logs
- Set resource limits (CPU, memory) per container


## 13. Testing Strategy
- Unit-test repository query builders, service logic, middleware and handlers.
- Use `httptest` for handler tests and interface-based mocks for repositories.
- CI: run `go test ./... -coverprofile=coverage.out` and fail if coverage is below threshold for critical packages.

Commands:
```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```


## 14. Operational Checklist (local dev)
1. Ensure `.env` contains `DATABASE_DSN`, `JWT_SECRET`, `SERVER_PORT`, `GIN_MODE=debug`, `MIGRATE_AUTO=false`.
2. Run migrations: `make migrate-up` (or `migrate -path ./migrations -database "$DATABASE_DSN" up`).
3. Start app: `go run ./cmd` (or via `docker-compose up`).
4. Create staff user, login, get token.
5. Call protected endpoints and validate behavior.


## 15. Appendix & References
- Libraries used: `github.com/jmoiron/sqlx`, `github.com/golang-jwt/jwt`, `golang.org/x/crypto/bcrypt`, `github.com/golang-migrate/migrate`.
- See `docs/chatgpt-techdoc.md` and `docs/techdoc.md` for original notes.

