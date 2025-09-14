#!/bin/bash

# Test Environment Setup Script
# This script sets up the complete test environment for the SMS Gateway

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

# Check if required tools are installed
check_dependencies() {
  print_status "Checking dependencies..."

  local missing_deps=()

  if ! command -v docker &>/dev/null; then
    missing_deps+=("docker")
  fi

  if ! command -v docker compose &>/dev/null; then
    missing_deps+=("docker compose")
  fi

  if ! command -v go &>/dev/null; then
    missing_deps+=("go")
  fi

  if [ ${#missing_deps[@]} -ne 0 ]; then
    print_error "Missing dependencies: ${missing_deps[*]}"
    print_error "Please install the missing dependencies and try again."
    exit 1
  fi

  print_success "All dependencies are installed"
}

# Check if Docker is running
check_docker() {
  print_status "Checking Docker daemon..."

  if ! docker info &>/dev/null; then
    print_error "Docker daemon is not running"
    print_error "Please start Docker and try again."
    exit 1
  fi

  print_success "Docker daemon is running"
}

# Setup test configuration
setup_test_config() {
  print_status "Setting up test configuration..."

  # Backup original config if it exists
  if [ -f "SmsGW.yaml" ]; then
    cp SmsGW.yaml SmsGW.yaml.backup
    print_status "Backed up original SmsGW.yaml"
  fi

  # Copy test config
  cp tests/config/test.yaml SmsGW.yaml
  print_success "Test configuration setup complete"
}

# Start test dependencies
start_test_deps() {
  print_status "Starting test dependencies..."

  # Start PostgreSQL and NATS
  docker compose -f docker-compose.test.yml up -d

  print_status "Waiting for services to be ready..."

  # Wait for PostgreSQL
  local pg_ready=false
  for i in {1..30}; do
    if docker compose -f docker-compose.test.yml exec postgres-test pg_isready -U root &>/dev/null; then
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
    if curl -f http://localhost:8223/varz &>/dev/null; then
      nats_ready=true
      break
    fi
    sleep 2
  done

  if [ "$nats_ready" = false ]; then
    print_error "NATS failed to start within 60 seconds"
    exit 1
  fi

  print_success "All test dependencies are ready!"
}

# Verify test environment
verify_test_env() {
  print_status "Verifying test environment..."

  # Check PostgreSQL connection
  if ! docker compose -f docker-compose.test.yml exec postgres-test pg_isready -U root &>/dev/null; then
    print_error "PostgreSQL is not accessible"
    return 1
  fi

  # Check NATS connection
  if ! curl -f http://localhost:8223/varz &>/dev/null; then
    print_error "NATS is not accessible"
    return 1
  fi

  # Check if test config exists
  if [ ! -f "SmsGW.yaml" ]; then
    print_error "Test configuration file not found"
    return 1
  fi

  print_success "Test environment verification complete"
}

# Run a quick test to verify everything works
run_quick_test() {
  print_status "Running quick test to verify setup..."

  # Set test environment variables (only GIN_MODE is needed now)
  export GIN_MODE=test

  # Run a simple test
  if go test -v ./pkg/utils/... &>/dev/null; then
    print_success "Quick test passed - environment is working!"
  else
    print_warning "Quick test failed - there might be issues with the test environment"
  fi
}

# Show usage information
show_usage() {
  echo "SMS Gateway Test Environment Setup"
  echo ""
  echo "This script sets up the complete test environment including:"
  echo "  - Docker containers for PostgreSQL and NATS"
  echo "  - Test configuration files"
  echo "  - Environment variables"
  echo "  - Verification of the setup"
  echo ""
  echo "Usage: $0 [OPTIONS]"
  echo ""
  echo "Options:"
  echo "  --skip-deps    Skip dependency checks"
  echo "  --skip-docker  Skip Docker checks"
  echo "  --skip-test    Skip quick test run"
  echo "  --help         Show this help message"
  echo ""
  echo "After running this script, you can run tests with:"
  echo "  ./scripts/test-runner.sh all"
  echo "  make -f Makefile.test test-all"
}

# Main function
main() {
  local skip_deps=false
  local skip_docker=false
  local skip_test=false

  # Parse command line arguments
  while [[ $# -gt 0 ]]; do
    case $1 in
    --skip-deps)
      skip_deps=true
      shift
      ;;
    --skip-docker)
      skip_docker=true
      shift
      ;;
    --skip-test)
      skip_test=true
      shift
      ;;
    --help)
      show_usage
      exit 0
      ;;
    *)
      print_error "Unknown option: $1"
      show_usage
      exit 1
      ;;
    esac
  done

  print_status "Setting up SMS Gateway test environment..."

  # Check dependencies
  if [ "$skip_deps" = false ]; then
    check_dependencies
  fi

  # Check Docker
  if [ "$skip_docker" = false ]; then
    check_docker
  fi

  # Setup test configuration
  setup_test_config

  # Start test dependencies
  start_test_deps

  # Verify environment
  verify_test_env

  # Run quick test
  if [ "$skip_test" = false ]; then
    run_quick_test
  fi

  print_success "Test environment setup complete!"
  echo ""
  echo "You can now run tests using:"
  echo "  ./scripts/test-runner.sh all"
  echo "  make -f Makefile.test test-all"
  echo ""
  echo "To stop the test environment:"
  echo "  ./scripts/test-runner.sh teardown"
  echo "  make -f Makefile.test teardown-test-deps"
}

# Run main function with all arguments
main "$@"

