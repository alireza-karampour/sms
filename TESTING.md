# SMS Gateway Testing Documentation

This document provides a comprehensive overview of the testing setup for the SMS Gateway project using Ginkgo and Gomega.

## Overview

The SMS Gateway now includes a complete testing framework with:
- **Unit Tests**: Fast, isolated tests for individual components
- **Integration Tests**: Tests that verify component interactions including SMS Worker functionality
- **Test Infrastructure**: Docker-based test dependencies and automated setup

## Test Architecture

### Test Structure
```
tests/
├── integration/          # Integration tests
│   ├── controllers_suite_test.go
│   ├── user_controller_test.go
│   ├── sms_controller_test.go
│   └── sms_worker_test.go
├── helpers/             # Test utilities
│   ├── test_suite.go    # Database and NATS setup
│   └── http_client.go   # HTTP testing utilities
├── fixtures/            # Test data
└── config/              # Test configuration
    └── test.yaml
```

### Test Dependencies
- **PostgreSQL**: Isolated test database with automatic schema migration
- **NATS**: Message queue for testing SMS workflows
- **Docker Compose**: Automated dependency management

## Quick Start

### 1. Setup Test Environment
```bash
# Setup test environment and dependencies
./scripts/test-runner.sh setup
```

### 2. Run Tests
```bash
# Run all tests
./scripts/test-runner.sh all

# Run specific test types
./scripts/test-runner.sh unit
./scripts/test-runner.sh integration

# Run with coverage
./scripts/test-runner.sh coverage
```

### 3. Cleanup
```bash
./scripts/test-runner.sh teardown
```

## Test Types

### Unit Tests
- **Location**: `pkg/` directory
- **Purpose**: Test individual functions and methods
- **Dependencies**: None (isolated)
- **Speed**: Fast
- **Example**: `pkg/utils/utils_test.go`

### Integration Tests
- **Location**: `tests/integration/`
- **Purpose**: Test component interactions
- **Dependencies**: Database, NATS
- **Speed**: Medium
- **Examples**:
  - User controller with database
  - SMS controller with NATS
  - SMS Worker with NATS and database
  - HTTP API endpoints

#### SMS Controller Integration Tests
- Test SMS API endpoints
- Verify NATS message publishing
- Test balance validation
- Test error handling scenarios

#### SMS Worker Integration Tests
- Test NATS message consumption
- Test database operations (SMS creation, balance deduction)
- Test rate limiting functionality
- Test error handling and retry logic
- Test concurrent message processing

## Test Helpers

### TestSuite
Provides isolated test environment:
```go
testSuite := helpers.SetupTestSuite()
defer testSuite.Cleanup()

// Access test database
db := testSuite.DB

// Access NATS connection
nats := testSuite.NATSConn

// Clean test data
testSuite.CleanupTestData()
```

### HTTPClient
Simplifies HTTP testing:
```go
client := helpers.NewHTTPClient("http://localhost:8080")

// Make requests
resp, err := client.Post("/user", helpers.RequestOptions{
    Body: userData,
})

// Assert responses
helpers.AssertResponseStatus(resp, http.StatusOK)
```

## Test Configuration

### Environment Variables
```bash
export TEST_POSTGRES_ADDRESS=127.0.0.1
export TEST_POSTGRES_PORT=5434
export TEST_POSTGRES_USERNAME=root
export TEST_POSTGRES_PASSWORD=1234
export TEST_NATS_ADDRESS=127.0.0.1:4223
export GIN_MODE=test
```

### Test Configuration File
`tests/config/test.yaml`:
```yaml
api:
  nats:
    address: "127.0.0.1:4223"
  listen: "127.0.0.1:8081"
  postgres:
    address: "127.0.0.1"
    port: 5434
    username: root
    password: 1234
```

## Writing Tests

### Integration Test Example
```go
var _ = Describe("User Controller Integration Tests", func() {
    var (
        testSuite *helpers.TestSuite
        router    *gin.Engine
        userCtrl  *controllers.User
        queries   *sqlc.Queries
    )

    BeforeEach(func() {
        testSuite = helpers.SetupTestSuite()
        queries = sqlc.New(testSuite.DB)
        
        gin.SetMode(gin.TestMode)
        router = gin.New()
        userCtrl = controllers.NewUser(router.Group("/"), testSuite.DB)
    })

    AfterEach(func() {
        testSuite.CleanupTestData()
        testSuite.Cleanup()
    })

    It("should create a new user successfully", func() {
        // Test implementation
        balance := pgtype.Numeric{Int: pgtype.Int{Int64: 10000, Valid: true}, Exp: -2}
        err := queries.AddUser(context.Background(), sqlc.AddUserParams{
            Username: "testuser",
            Balance:  balance,
        })
        Expect(err).NotTo(HaveOccurred())
    })
})
```

### SMS Worker Integration Test Example
```go
var _ = Describe("SMS Worker Integration Tests", func() {
    var (
        testSuite *helpers.TestSuite
        worker    *workers.Sms
        queries   *sqlc.Queries
    )

    BeforeEach(func() {
        testSuite = helpers.SetupTestSuite()
        queries = sqlc.New(testSuite.DB)
        
        // Create SMS worker
        worker, err = workers.NewSms(context.Background(), 
            testSuite.NATSConn.Address, testSuite.DB)
        Expect(err).NotTo(HaveOccurred())
    })

    It("should process normal SMS request successfully", func() {
        // Create SMS data
        smsData := sqlc.Sm{
            UserID:        userID,
            PhoneNumberID: phoneID,
            ToPhoneNumber: "+0987654321",
            Message:       "Test SMS message",
            Status:        "pending",
        }

        // Start worker and publish message
        // Verify SMS was added to database and balance deducted
    })
})
```

## Test Data Management

### Automatic Cleanup
- Each test gets a fresh database
- Test data is automatically cleaned up
- Database sequences are reset
- NATS streams are cleared

### Test Isolation
- No shared state between tests
- Parallel test execution supported
- Race detection enabled

## Continuous Integration

### GitHub Actions Example
```yaml
name: Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: '1.21'
      
      - name: Setup test environment
        run: ./scripts/test-runner.sh setup
      
      - name: Run tests
        run: ./scripts/test-runner.sh all
      
      - name: Generate coverage report
        run: ./scripts/test-runner.sh coverage
      
      - name: Cleanup
        if: always()
        run: ./scripts/test-runner.sh teardown
```

## Available Commands

### Test Runner Script
```bash
./scripts/test-runner.sh [COMMAND]

Commands:
  unit         Run unit tests only
  integration  Run integration tests only
  all          Run all tests
  coverage     Run tests with coverage report
  setup        Setup test dependencies
  teardown     Teardown test dependencies
  check        Check if dependencies are running
```

### Makefile Commands
```bash
make -f Makefile.test [TARGET]

Targets:
  setup-test-deps    Start test dependencies
  teardown-test-deps Stop test dependencies
  test-unit          Run unit tests
  test-integration   Run integration tests
  test-all           Run all tests
  test-coverage      Run tests with coverage
  clean-test         Clean test artifacts
```

## Troubleshooting

### Common Issues

1. **Port Conflicts**
   - Ensure ports 5434 and 4223 are available
   - Check for running PostgreSQL/NATS instances

2. **Docker Issues**
   - Verify Docker is running: `docker info`
   - Check Docker Compose version compatibility

3. **Database Connection**
   - Verify PostgreSQL container health
   - Check connection parameters in test config

4. **NATS Connection**
   - Verify NATS container is running
   - Check NATS health endpoint: `curl http://localhost:8223/healthz`

### Debug Mode
```bash
# Verbose test output
go test -v ./tests/... -race

# Debug specific test
go test -v ./tests/integration/user_controller_test.go -race

# Run with timeout
go test -timeout=30s ./tests/...
```

## Performance Considerations

- Tests use isolated databases for speed
- NATS streams configured for testing
- Parallel execution supported
- Race detection enabled by default
- Coverage reporting available

## Best Practices

1. **Test Isolation**: Each test should be independent
2. **Cleanup**: Always clean up test data
3. **Assertions**: Use descriptive assertion messages
4. **Error Handling**: Test both success and failure scenarios
5. **Performance**: Keep tests fast and focused
6. **Documentation**: Document complex test scenarios

## Contributing

When adding new tests:
1. Follow the existing test structure
2. Use the provided test helpers
3. Add appropriate test data cleanup
4. Include both positive and negative test cases
5. Update documentation for new test types

## Test Coverage

The test suite covers:
- ✅ User management (creation, retrieval, balance)
- ✅ SMS sending (normal and express)
- ✅ SMS Worker message processing
- ✅ Database operations
- ✅ NATS message publishing and consumption
- ✅ HTTP API endpoints
- ✅ Error handling scenarios
- ✅ Rate limiting functionality
- ✅ Concurrent message processing

## Future Enhancements

Potential improvements:
- Performance benchmarks
- Load testing
- Chaos engineering tests
- Security testing
- API contract testing