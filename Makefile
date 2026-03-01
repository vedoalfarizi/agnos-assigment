.PHONY: help build run migrate-up migrate-down clean test dev db-up db-down db-logs db-reset

help:
	@echo "Hospital API - Available commands:"
	@echo ""
	@echo "  make build            - Build the application binary"
	@echo "  make run              - Run the application server"
	@echo "  make migrate-up       - Apply all pending database migrations"
	@echo "  make migrate-down     - Revert the last database migration"
	@echo "  make clean            - Remove build artifacts"
	@echo "  make test             - Run tests with coverage"
	@echo "  make dev              - Run server in development mode (with auto-reload)"
	@echo ""
	@echo "Database commands:"
	@echo "  make db-up            - Start PostgreSQL database (docker-compose)"
	@echo "  make db-down          - Stop PostgreSQL database"
	@echo "  make db-logs          - View database logs"
	@echo "  make db-reset         - Remove database and start fresh"
	@echo ""

build:
	@echo "Building application..."
	go build -o bin/api ./cmd/api

run: build
	@echo "Starting server..."
	./bin/api

migrate-up:
	@echo "Running migrations up..."
	@# load environment variables and run migration in same shell
	@set -a; [ -f .env ] && . .env; set +a; \
	if command -v migrate >/dev/null 2>&1; then \
		migrate -path ./migrations -database "$$DATABASE_DSN" up; \
	else \
		echo "golang-migrate not found. Using Docker..."; \
		docker-compose run --rm migrate -path=/migrations -database "$$DATABASE_DSN" up; \
	fi

migrate-down:
	@echo "Running migrations down..."
	@set -a; [ -f .env ] && . .env; set +a; \
	if command -v migrate >/dev/null 2>&1; then \
		migrate -path ./migrations -database "$$DATABASE_DSN" down; \
	else \
		echo "golang-migrate not found. Using Docker..."; \
		docker-compose run --rm migrate -path=/migrations -database "$$DATABASE_DSN" down; \
	fi

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/

test:
	@echo "Running tests..."
	go test ./... -v -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

dev:
	@e

db-up:
	@echo "Starting PostgreSQL database..."
	docker-compose up -d postgres
	@echo "Waiting for database to be ready..."
	@sleep 5
	@echo "Database is ready!"

db-down:
	@echo "Stopping PostgreSQL database..."
	docker-compose down

db-logs:
	@echo "Showing database logs..."
	docker-compose logs -f postgres

db-reset:
	@echo "Removing database and starting fresh..."
	docker-compose down -v
	docker-compose up -d postgres
	@echo "Waiting for database to be ready..."
	@sleep 5
	@echo "Database is ready!"cho "Starting server in development mode..."
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "air not found, running go run instead..."; \
		go run ./cmd/api; \
	fi
