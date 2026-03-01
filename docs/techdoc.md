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

A simple unauthenticated endpoint exposed at `/health` (and optionally versioned paths like `/api/v1/status`).
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

