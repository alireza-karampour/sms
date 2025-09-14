# SMS Gateway Testing Guide

This directory contains comprehensive test suites for the SMS Gateway project, including unit tests and integration tests using Ginkgo and Gomega.

## Test Structure

```
tests/
├── integration/          # Integration tests
│   ├── controllers_suite_test.go
│   ├── user_controller_test.go
│   ├── sms_controller_test.go
│   └── sms_worker_test.go
├── helpers/             # Test helpers and utilities
│   ├── test_suite.go
│   └── http_client.go
├── fixtures/            # Test data fixtures
└── config/              # Test configuration
    └── test.yaml
```

## Test Types

### Unit Tests
- Located in `pkg/` directory
- Test individual functions and methods in isolation
- Fast execution, no external dependencies
- Run with: `go test ./pkg/...`

### Integration Tests
- Located in `tests/integration/`
- Test interactions between components
- Use real database and NATS connections
- Test HTTP API endpoints and SMS Worker functionality
- Run with: `go test ./tests/integration/...`

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

## Test Dependencies

The tests require the following services:
- **PostgreSQL**: Database for storing users, phone numbers, and SMS records
- **NATS**: Message queue for SMS processing

These are automatically managed using Docker Compose.

## Quick Start

### 1. Setup Test Environment

```bash
# Start test dependencies
./scripts/test-runner.sh setup

# Or using Makefile
make -f Makefile.test setup-test-deps
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
# Stop test dependencies
./scripts/test-runner.sh teardown

# Or using Makefile
make -f Makefile.test teardown-test-deps
```

## Test Configuration

Test configuration is managed through environment variables and the `tests/config/test.yaml` file:

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

## Test Helpers

### TestSuite
The `TestSuite` helper provides:
- Database connection with isolated test database
- NATS connection
- Automatic cleanup after tests
- Schema migration

### HTTPClient
The `HTTPClient` helper provides:
- Convenient HTTP request methods
- JSON request/response handling
- Response assertion helpers

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
        // ... test implementation
    })
})
```

## Test Data Management

### Automatic Cleanup
- Each test gets a fresh database
- Test data is automatically cleaned up after each test
- Database sequences are reset

### Test Fixtures
- Common test data is provided by `helpers.GetTestData()`
- Custom test data can be created as needed

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
      
      - name: Setup test dependencies
        run: ./scripts/test-runner.sh setup
      
      - name: Run tests
        run: ./scripts/test-runner.sh all
      
      - name: Cleanup
        if: always()
        run: ./scripts/test-runner.sh teardown
```

## Troubleshooting

### Common Issues

1. **Port conflicts**: Ensure ports 5434 and 4223 are available
2. **Docker issues**: Make sure Docker is running and accessible
3. **Database connection**: Check if PostgreSQL container is healthy
4. **NATS connection**: Verify NATS container is running

### Debug Mode

Run tests with verbose output:
```bash
go test -v ./tests/... -race
```

### Test Isolation

Each test runs in isolation:
- Fresh database per test
- Clean NATS streams
- No shared state between tests

## Performance Considerations

- Tests use in-memory databases when possible
- NATS streams are configured for testing
- Parallel test execution is supported
- Race detection is enabled by default

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
5. Update this README if adding new test types or helpers