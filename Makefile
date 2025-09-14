# SMS Gateway Makefile

.PHONY: help build run-api run-worker run-e2e-tests clean docker-build docker-up docker-down docker-logs test integration-test e2e-test

# Default target
help:
	@echo "Available commands:"
	@echo "  build              - Build the Go application"
	@echo "  run-api            - Run API service locally"
	@echo "  run-worker         - Run worker service locally"
	@echo "  run-e2e-tests      - Run E2E tests with dockerized services"
	@echo "  clean              - Clean build artifacts"
	@echo "  docker-build       - Build Docker images"
	@echo "  docker-up          - Start all services with Docker Compose"
	@echo "  docker-down        - Stop all services"
	@echo "  docker-logs        - Show logs from all services"
	@echo "  test               - Run all tests"
	@echo "  integration-test   - Run integration tests"
	@echo "  e2e-test           - Run E2E tests (requires running services)"

# Build the application
build:
	@echo "Building SMS Gateway..."
	go build -o sms ./main.go

# Run API service locally
run-api: build
	@echo "Starting API service..."
	./sms api

# Run worker service locally
run-worker: build
	@echo "Starting worker service..."
	./sms worker

# Run E2E tests with automatic infrastructure setup
run-e2e-tests:
	@echo "Running E2E tests with dockerized services..."
	./scripts/test-runner.sh e2e

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f sms

# Build Docker images
docker-build:
	@echo "Building Docker images..."
	docker compose build

# Start all services with Docker Compose
docker-up:
	@echo "Starting all services..."
	docker compose up -d

# Stop all services
docker-down:
	@echo "Stopping all services..."
	docker compose down

# Show logs from all services
docker-logs:
	@echo "Showing logs from all services..."
	docker compose logs -f

# Run all tests
test:
	@echo "Running all tests..."
	go test ./...

# Run integration tests
integration-test:
	@echo "Running integration tests..."
	go test ./tests/integration/... -v

# Run E2E tests (requires running services)
e2e-test:
	@echo "Running E2E tests..."
	go test ./tests/e2e/... -v

# Development setup
dev-setup:
	@echo "Setting up development environment..."
	docker compose -f docker-compose.test.yml up -d
	@echo "Waiting for services to be ready..."
	sleep 10
	@echo "Development environment ready!"

# Stop development environment
dev-cleanup:
	@echo "Cleaning up development environment..."
	docker compose -f docker-compose.test.yml down -v