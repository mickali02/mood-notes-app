## Filename Makefile

# Include environment variables from .envrc if it exists
# Use '-include' to ignore error if .envrc is missing
-include .envrc

## Go Tools

.PHONY: fmt
fmt:
	@echo 'Formatting code...'
	@go fmt ./...

.PHONY: vet
vet: fmt
	@echo 'Vetting code...'
	@go vet ./...

.PHONY: test
test: vet
	@echo 'Running tests...'
	@go test -v -race ./... # Add -race flag for data race detection

## Application Execution

.PHONY: run
run: vet
	@echo 'Running application...'
	@go run ./cmd/web -dsn=${MOODNOTES_DB_DSN} # Use CORRECT DSN variable from .envrc

## Database Operations

.PHONY: db/psql
db/psql:
	@echo 'Connecting to database...'
	@psql ${MOODNOTES_DB_DSN} # Use CORRECT DSN variable

# db/migrations/new: create a new database migration file pair
# Usage: make name=your_migration_name db/migrations/new
.PHONY: db/migrations/new
db/migrations/new:
ifndef name
	$(error "Error: name is undefined. Usage: make name=create_users db/migrations/new")
endif
	@echo 'Creating migration files for ${name}...'
	@mkdir -p ./migrations # Ensure migrations directory exists
	@migrate create -seq -ext=.sql -dir=./migrations ${name}

# db/migrations/up: apply all pending UP database migrations
.PHONY: db/migrations/up
db/migrations/up:
	@echo 'Running up migrations...'
	@migrate -path ./migrations -database ${MOODNOTES_DB_DSN} up # Use CORRECT DSN variable

# db/migrations/down: apply all DOWN database migrations (Use with caution!)
.PHONY: db/migrations/down
db/migrations/down:
	@echo 'Running down migrations...'
	@migrate -path ./migrations -database ${MOODNOTES_DB_DSN} down -all # Use CORRECT DSN variable

## Help Command

.PHONY: help
help:
	@echo '--------------------'
	@echo 'Available Commands:'
	@echo '--------------------'
	@echo 'make fmt                          - Format Go code'
	@echo 'make vet                          - Run Go vet checks (includes fmt)'
	@echo 'make test                         - Run Go tests (includes vet, fmt)'
	@echo 'make run                          - Build and run the web application (requires MOODNOTES_DB_DSN)'
	@echo 'make db/psql                      - Connect to the database via psql (requires MOODNOTES_DB_DSN)'
	@echo 'make name=<name> db/migrations/new - Create new migration files'
	@echo 'make db/migrations/up             - Apply all UP migrations (requires MOODNOTES_DB_DSN)'
	@echo 'make db/migrations/down           - Apply all DOWN migrations (requires MOODNOTES_DB_DSN)'
	@echo 'make help                         - Show this help message'
	@echo '--------------------'

# Default target when running 'make' without arguments
.DEFAULT_GOAL := help
