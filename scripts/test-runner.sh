#!/bin/bash

# Test Runner Script for SMS Gateway
# This script provides a convenient way to run different types of tests

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

# Function to check if dependencies are running
check_dependencies() {
  print_status "Checking test dependencies..."

  # Check PostgreSQL
  if ! docker compose -f docker-compose.test.yml exec postgres-test pg_isready -U root >/dev/null 2>&1; then
    print_error "PostgreSQL is not ready"
    return 1
  fi

  # Check NATS
  if ! curl -f http://localhost:8223/varz >/dev/null 2>&1; then
    print_error "NATS is not ready"
    return 1
  fi

  print_success "All dependencies are ready"
  return 0
}

# Function to setup test environment
setup_test_env() {
  print_status "Setting up test environment..."

  # Backup original config if it exists
  if [ -f "SmsGW.yaml" ]; then
    cp SmsGW.yaml SmsGW.yaml.backup
    print_status "Backed up original SmsGW.yaml"
  fi

  # Copy test config
  cp tests/config/test.yaml SmsGW.yaml

  # Set environment variables (only GIN_MODE is needed now)
  export GIN_MODE=test

  print_success "Test environment setup complete"
}

# Function to cleanup test environment
cleanup_test_env() {
  print_status "Cleaning up test environment..."

  # Restore original config if it exists
  if [ -f "SmsGW.yaml.backup" ]; then
    mv SmsGW.yaml.backup SmsGW.yaml
  else
    rm -f SmsGW.yaml
  fi

  print_success "Test environment cleanup complete"
}

# Function to run unit tests
run_unit_tests() {
  print_status "Running unit tests..."
  go test -v ./pkg/...
  print_success "Unit tests completed"
}

# Function to run integration tests
run_integration_tests() {
  print_status "Running integration tests..."
  ginkgo run -vv ./tests/integration/...
  print_success "Integration tests completed"
}

# Function to setup E2E environment
setup_e2e_env() {
  print_status "Setting up E2E environment..."
  
  # Check if Docker is running
  if ! docker info >/dev/null 2>&1; then
    print_error "Docker is not running. Please start Docker and try again."
    exit 1
  fi

  # Check if docker compose is available
  if ! docker compose version >/dev/null 2>&1; then
    print_error "docker compose is not available. Please install Docker Compose and try again."
    exit 1
  fi

  # Check if E2E docker-compose file exists
  if [ ! -f "tests/e2e/env/docker-compose.yml" ]; then
    print_error "E2E docker-compose.yml not found at tests/e2e/env/docker-compose.yml"
    exit 1
  fi

  # Check if E2E config exists
  if [ ! -f "tests/e2e/env/SmsGW.yaml" ]; then
    print_error "E2E config file not found at tests/e2e/env/SmsGW.yaml"
    exit 1
  fi

  # Change to E2E environment directory
  cd tests/e2e/env

  print_status "Starting E2E infrastructure..."
  docker compose up -d

  print_status "Waiting for services to be ready..."

  # Wait for PostgreSQL
  local pg_ready=false
  for i in {1..30}; do
    if docker compose exec postgres-e2e pg_isready -U root >/dev/null 2>&1; then
      pg_ready=true
      break
    fi
    sleep 2
  done

  if [ "$pg_ready" = false ]; then
    print_error "PostgreSQL failed to start within 60 seconds"
    docker compose logs postgres-e2e
    cd - >/dev/null
    exit 1
  fi

  # Wait for NATS
  local nats_ready=false
  for i in {1..30}; do
    if curl -f http://localhost:8222/healthz >/dev/null 2>&1; then
      nats_ready=true
      break
    fi
    sleep 2
  done

  if [ "$nats_ready" = false ]; then
    print_error "NATS failed to start within 60 seconds"
    docker compose logs nats-e2e
    cd - >/dev/null
    exit 1
  fi

  # Wait for API
  local api_ready=false
  for i in {1..30}; do
    if curl -f http://localhost:8080/health >/dev/null 2>&1; then
      api_ready=true
      break
    fi
    sleep 2
  done

  if [ "$api_ready" = false ]; then
    print_error "API failed to start within 60 seconds"
    docker compose logs api-e2e
    cd - >/dev/null
    exit 1
  fi

  # Return to original directory
  cd - >/dev/null

  print_success "E2E environment setup complete!"
  print_status "Services available at:"
  print_status "  - API: http://localhost:8080"
  print_status "  - PostgreSQL: localhost:5432"
  print_status "  - NATS: localhost:4222 (HTTP: localhost:8222)"
}

# Function to teardown E2E environment
teardown_e2e_env() {
  print_status "Tearing down E2E environment..."
  
  # Change to E2E environment directory
  cd tests/e2e/env
  
  docker compose down -v
  
  # Return to original directory
  cd - >/dev/null
  
  print_success "E2E environment stopped!"
}

# Function to run e2e tests
run_e2e_tests() {
  print_status "Running end-to-end tests..."
  
  # Check if Docker is running
  if ! docker info >/dev/null 2>&1; then
    print_error "Docker is not running. Please start Docker and try again."
    exit 1
  fi

  # Check if docker compose is available
  if ! docker compose version >/dev/null 2>&1; then
    print_error "docker compose is not available. Please install Docker Compose and try again."
    exit 1
  fi

  # Function to cleanup E2E infrastructure on exit
  cleanup_e2e() {
    print_status "Cleaning up E2E infrastructure..."
    docker compose -f docker-compose.e2e.yml down -v
    print_success "E2E infrastructure cleaned up"
  }

  # Set trap to cleanup on script exit
  trap cleanup_e2e EXIT

  print_status "Starting E2E test infrastructure..."

  # Check if API is responding
  print_status "Checking API health..."
  max_attempts=30
  attempt=0

  while [ $attempt -lt $max_attempts ]; do
    if curl -f http://localhost:8080/health >/dev/null 2>&1; then
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
  else
    print_error "E2E tests failed!"
    exit 1
  fi
}

# Function to run all tests
run_all_tests() {
  print_status "Running all tests..."
  go test -v ./...
  print_success "All tests completed"
}

# Function to run tests with coverage
run_coverage_tests() {
  print_status "Running tests with coverage..."
  go test -v ./... -coverprofile=coverage.out
  go tool cover -html=coverage.out -o coverage.html
  print_success "Coverage report generated: coverage.html"
}

# Function to show help
show_help() {
  echo "SMS Gateway Test Runner"
  echo ""
  echo "Usage: $0 [COMMAND]"
  echo ""
  echo "Commands:"
  echo "  unit         Run unit tests only"
  echo "  integration  Run integration tests only (requires setup)"
  echo "  e2e          Run end-to-end tests with full infrastructure"
  echo "  all          Run all tests (requires setup)"
  echo "  coverage     Run tests with coverage report (requires setup)"
  echo "  setup        Setup test environment and dependencies"
  echo "  teardown     Teardown test dependencies"
  echo "  e2e-setup    Setup E2E environment using tests/e2e/env/docker-compose.yml"
  echo "  e2e-teardown Teardown E2E environment"
  echo "  check        Check if dependencies are running"
  echo "  help         Show this help message"
  echo ""
  echo "Examples:"
  echo "  $0 unit              # Run unit tests"
  echo "  $0 setup             # Setup test environment"
  echo "  $0 integration      # Run integration tests"
  echo "  $0 e2e              # Run e2e tests (handles own infrastructure)"
  echo "  $0 e2e-setup         # Setup E2E environment manually"
  echo "  $0 e2e-teardown      # Teardown E2E environment"
  echo "  $0 all              # Run all tests"
  echo "  $0 coverage         # Run tests with coverage"
  echo ""
  echo "Note: E2E tests handle their own infrastructure setup and teardown."
  echo "      Other test types require running '$0 setup' first."
  echo "      Use 'e2e-setup' to manually setup E2E environment for development."
}

# Main script logic
main() {
  case "${1:-help}" in
  "unit")
    setup_test_env
    run_unit_tests
    cleanup_test_env
    ;;
  "integration")
    setup_test_env
    if ! check_dependencies; then
      print_error "Dependencies not ready. Run '$0 setup' first."
      exit 1
    fi
    run_integration_tests
    cleanup_test_env
    ;;
  "e2e")
    run_e2e_tests
    ;;
  "all")
    setup_test_env
    if ! check_dependencies; then
      print_error "Dependencies not ready. Run '$0 setup' first."
      exit 1
    fi
    run_all_tests
    cleanup_test_env
    ;;
  "coverage")
    setup_test_env
    if ! check_dependencies; then
      print_error "Dependencies not ready. Run '$0 setup' first."
      exit 1
    fi
    run_coverage_tests
    cleanup_test_env
    ;;
  "setup")
    print_status "Setting up test environment..."
    
    # Check if Docker is running
    if ! docker info >/dev/null 2>&1; then
      print_error "Docker is not running. Please start Docker and try again."
      exit 1
    fi

    # Setup test configuration
    setup_test_env
    
    # Start test dependencies
    print_status "Starting test dependencies..."
    docker compose -f docker-compose.test.yml up -d
    
    print_status "Waiting for services to be ready..."
    
    # Wait for PostgreSQL
    local pg_ready=false
    for i in {1..30}; do
      if docker compose -f docker-compose.test.yml exec postgres-test pg_isready -U root >/dev/null 2>&1; then
        pg_ready=true
        break
      fi
      sleep 2
    done

    if [ "$pg_ready" = false ]; then
      print_error "PostgreSQL failed to start within 60 seconds"
      exit 1
    fi

    # Wait for NATS
    local nats_ready=false
    for i in {1..30}; do
      if curl -f http://localhost:8223/varz >/dev/null 2>&1; then
        nats_ready=true
        break
      fi
      sleep 2
    done

    if [ "$nats_ready" = false ]; then
      print_error "NATS failed to start within 60 seconds"
      exit 1
    fi

    print_success "Test environment setup complete!"
    ;;
  "teardown")
    print_status "Tearing down test dependencies..."
    docker compose -f docker-compose.test.yml down -v
    print_success "Test dependencies stopped!"
    ;;
  "e2e-setup")
    setup_e2e_env
    ;;
  "e2e-teardown")
    teardown_e2e_env
    ;;
  "check")
    if check_dependencies; then
      print_success "Dependencies are running"
    else
      print_error "Dependencies are not ready"
      exit 1
    fi
    ;;
  "help" | *)
    show_help
    ;;
  esac
}

# Run main function with all arguments
main "$@"
