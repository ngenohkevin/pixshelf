.PHONY: all build run clean test migrate-up migrate-down migrate-create sqlc templ

# Application name
APP_NAME := pixshelf

# Go related
GO := go
GO_BUILD := $(GO) build
GO_TEST := $(GO) test
GO_CLEAN := $(GO) clean

# Database configuration
DB_URL := postgres://postgres:postgres@localhost:5432/pixshelf?sslmode=disable

# Build the application
build:
	$(GO_BUILD) -o $(APP_NAME) ./cmd/server

# Run the application
run:
	$(GO) run ./cmd/server

# Clean build artifacts
clean:
	$(GO_CLEAN)
	rm -f $(APP_NAME)

# Run tests
test:
	$(GO_TEST) -v ./...

# Create a new migration file
# Usage: make migrate-create name=create_users_table
migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "Error: migration name is required. Use 'make migrate-create name=migration_name'"; \
		exit 1; \
	fi
	migrate create -ext sql -dir migrations -seq $(name)
	@echo "Migration files created in migrations directory"

# Run migrations up
migrate-up:
	migrate -path migrations -database "$(DB_URL)" up

# Run migrations down
migrate-down:
	migrate -path migrations -database "$(DB_URL)" down 1

# Generate SQLC code
sqlc:
	sqlc generate

# Generate templ templates
templ:
	templ generate

# Set up the development environment
setup: migrate-up sqlc templ

# Start PostgreSQL container for development
db-start:
	docker-compose up -d db

# Stop PostgreSQL container
db-stop:
	docker-compose down db

# Start the entire application stack
docker-up:
	docker-compose up -d

# Stop the entire application stack
docker-down:
	docker-compose down

# Show logs for the containers
docker-logs:
	docker-compose logs -f

# Default target
all: build

# Generate everything needed for development
generate: sqlc templ
