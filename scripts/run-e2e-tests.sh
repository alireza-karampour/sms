#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to cleanup on exit
cleanup() {
    print_status "Cleaning up E2E infrastructure..."
    docker compose -f docker-compose.e2e.yml down -v
    print_success "E2E infrastructure cleaned up"
}

# Set trap to cleanup on script exit
trap cleanup EXIT

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    print_error "Docker is not running. Please start Docker and try again."
    exit 1
fi

# Check if docker compose is available
if ! docker compose version > /dev/null 2>&1; then
    print_error "docker compose is not available. Please install Docker Compose and try again."
    exit 1
fi

print_status "Starting E2E test infrastructure..."

# Build and start the E2E services
print_status "Building and starting services..."
docker compose -f docker-compose.e2e.yml up --build -d

# Wait for services to be healthy
print_status "Waiting for services to be healthy..."
timeout=300  # 5 minutes timeout
elapsed=0

while [ $elapsed -lt $timeout ]; do
    if docker compose -f docker-compose.e2e.yml ps | grep -q "unhealthy"; then
        print_warning "Some services are still starting up..."
        sleep 10
        elapsed=$((elapsed + 10))
    else
        print_success "All services are healthy!"
        break
    fi
done

if [ $elapsed -ge $timeout ]; then
    print_error "Services failed to become healthy within $timeout seconds"
    docker compose -f docker-compose.e2e.yml ps
    exit 1
fi

# Wait a bit more for services to fully initialize
print_status "Waiting for services to fully initialize..."
sleep 30

# Check if API is responding
print_status "Checking API health..."
max_attempts=30
attempt=0

while [ $attempt -lt $max_attempts ]; do
    if curl -f http://localhost:8080/health > /dev/null 2>&1; then
        print_success "API is responding!"
        break
    else
        print_warning "API not ready yet, waiting..."
        sleep 5
        attempt=$((attempt + 1))
    fi
done

if [ $attempt -ge $max_attempts ]; then
    print_error "API failed to respond after $max_attempts attempts"
    docker compose -f docker-compose.e2e.yml logs api-e2e
    exit 1
fi

# Run the E2E tests
print_status "Running E2E tests..."
if go test ./tests/e2e/... -v; then
    print_success "E2E tests passed!"
    exit 0
else
    print_error "E2E tests failed!"
    exit 1
fi