.PHONY: help build run test clean lint fmt docker-build docker-run migrate-up migrate-down dev

# Default target
help:
	@echo "Available commands:"
	@echo "  build        Build the application"
	@echo "  run          Run the application"
	@echo "  dev          Run in development mode with hot reload"
	@echo "  test         Run tests"
	@echo "  test-cover   Run tests with coverage"
	@echo "  lint         Run linter"
	@echo "  fmt          Format code"
	@echo "  clean        Clean build artifacts"
	@echo "  docker-build Build Docker image"
	@echo "  docker-run   Run Docker container"
	@echo "  migrate-up   Run database migrations"
	@echo "  migrate-down Rollback database migrations"
	@echo "  deps         Download dependencies"
	@echo "  deps-update  Update dependencies"

# Build the application
build:
	@echo "Building application..."
	go build -o bin/todolist-api cmd/api/main.go

# Run the application
run:
	@echo "Running application..."
	go run cmd/api/main.go

# Run in development mode
dev:
	@echo "Running in development mode..."
	go run cmd/api/main.go

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-cover:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out coverage.html

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Update dependencies
deps-update:
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

# Docker commands
docker-build:
	@echo "Building Docker image..."
	docker build -t todolist-api:latest .

docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 todolist-api:latest

# Database migrations
migrate-up:
	@echo "Running database migrations..."
	migrate -path migrations -database "mysql://root:rootpassword@tcp(localhost:3306)/todolist_demo_dev" up

migrate-down:
	@echo "Rolling back database migrations..."
	migrate -path migrations -database "mysql://root:rootpassword@tcp(localhost:3306)/todolist_demo_dev" down

migrate-create:
	@echo "Creating new migration..."
	@if [ -z "$(NAME)" ]; then echo "Usage: make migrate-create NAME=migration_name"; exit 1; fi
	migrate create -ext sql -dir migrations -seq $(NAME)

# Development setup
setup-dev:
	@echo "Setting up development environment..."
	go mod download
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@echo "Development environment setup complete!"

# Production build
build-prod:
	@echo "Building for production..."
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/todolist-api-linux cmd/api/main.go

# Generate mocks
mocks:
	@echo "Generating mocks..."
	go generate ./...

# Swagger generation
swagger:
	@echo "Generating Swagger documentation..."
	swag init -g cmd/api/main.go -o docs

# Security scan
security:
	@echo "Running security scan..."
	gosec ./...

# All checks before commit
pre-commit: fmt lint test security
	@echo "All checks passed!"