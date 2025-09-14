#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

print_status "Testing Docker setup for SMS Gateway..."

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

print_success "Docker and Docker Compose are available"

# Check if required files exist
required_files=(
    "Dockerfile.api"
    "Dockerfile.worker"
    "docker-compose.yml"
    "docker-compose.e2e.yml"
    "Makefile"
    "scripts/run-e2e-tests.sh"
)

for file in "${required_files[@]}"; do
    if [ ! -f "$file" ]; then
        print_error "Required file $file is missing"
        exit 1
    fi
done

print_success "All required files are present"

# Test docker-compose syntax
print_status "Testing docker-compose.yml syntax..."
if docker compose -f docker-compose.yml config > /dev/null 2>&1; then
    print_success "docker-compose.yml syntax is valid"
else
    print_error "docker-compose.yml syntax is invalid"
    exit 1
fi

print_status "Testing docker-compose.e2e.yml syntax..."
if docker compose -f docker-compose.e2e.yml config > /dev/null 2>&1; then
    print_success "docker-compose.e2e.yml syntax is valid"
else
    print_error "docker-compose.e2e.yml syntax is invalid"
    exit 1
fi

# Test Makefile targets
print_status "Testing Makefile targets..."
make help > /dev/null 2>&1 && print_success "Makefile is valid"

# Test E2E script syntax
print_status "Testing E2E script syntax..."
if bash -n scripts/run-e2e-tests.sh; then
    print_success "E2E script syntax is valid"
else
    print_error "E2E script syntax is invalid"
    exit 1
fi

print_success "Docker setup validation completed successfully!"
print_status "You can now run:"
print_status "  make docker-build    # Build Docker images"
print_status "  make docker-up       # Start all services"
print_status "  make run-e2e-tests   # Run E2E tests with infrastructure"