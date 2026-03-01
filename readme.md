# Hospital API — README

This repository contains the Hospital API backend (Go). The README contains quick operational guidance for local development, migrations, and testing.


## Quick local dev steps
1. Create an `.env` from `.env.example`.
2. Start a local Postgres instance (docker-compose or local DB).
3. Run `make migrate-up` to apply migrations.
4. Start the app: `go run ./cmd` or `docker-compose up`.
5. Run tests: `go test ./... -coverprofile=coverage.out`.

