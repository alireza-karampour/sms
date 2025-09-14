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

# Function to run e2e tests
run_e2e_tests() {
  print_status "Running end-to-end tests..."
  go test -v ./tests/e2e/...
  print_success "End-to-end tests completed"
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
  echo "  integration  Run integration tests only"
  echo "  e2e          Run end-to-end tests only"
  echo "  all          Run all tests"
  echo "  coverage     Run tests with coverage report"
  echo "  setup        Setup test dependencies"
  echo "  teardown     Teardown test dependencies"
  echo "  check        Check if dependencies are running"
  echo "  help         Show this help message"
  echo ""
  echo "Examples:"
  echo "  $0 unit              # Run unit tests"
  echo "  $0 integration      # Run integration tests"
  echo "  $0 e2e              # Run e2e tests"
  echo "  $0 all              # Run all tests"
  echo "  $0 coverage         # Run tests with coverage"
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
    setup_test_env
    if ! check_dependencies; then
      print_error "Dependencies not ready. Run '$0 setup' first."
      exit 1
    fi
    run_e2e_tests
    cleanup_test_env
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
    print_status "Setting up test dependencies..."
    docker compose -f docker-compose.test.yml up -d
    print_status "Waiting for services to be ready..."
    sleep 10
    if check_dependencies; then
      print_success "Test dependencies are ready!"
    else
      print_error "Failed to setup test dependencies"
      exit 1
    fi
    ;;
  "teardown")
    print_status "Tearing down test dependencies..."
    docker compose -f docker-compose.test.yml down -v
    print_success "Test dependencies stopped!"
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
