# Development Guide

## Overview

This guide provides comprehensive instructions for setting up a development environment for the SMS Gateway project, including code structure, testing, and contribution guidelines.

## Prerequisites

### Required Software

- **Go**: Version 1.25 or higher
- **PostgreSQL**: Version 13 or higher
- **NATS**: Version 2.8 or higher
- **Git**: Version control
- **Docker**: For containerized development (optional)

### Development Tools

- **IDE**: VS Code, GoLand, or any Go-compatible editor
- **SQLC**: For database code generation
- **Helm**: For Kubernetes development (optional)

## Project Structure

```
sms/
├── cmd/                    # Application entry points
│   ├── api/               # REST API server
│   ├── worker/            # Background worker
│   └── root.go            # CLI root command
├── internal/              # Private application code
│   ├── controllers/       # HTTP controllers
│   ├── streams/           # NATS stream constants
│   ├── subjects/          # NATS subject constants
│   └── workers/           # Background workers
├── pkg/                   # Public packages
│   ├── middlewares/       # HTTP middlewares
│   ├── nats/              # NATS utilities
│   └── utils/             # Common utilities
├── sqlc/                  # Generated database code
├── charts/                # Kubernetes Helm charts
├── curl/                  # Test data and examples
├── docs/                  # Documentation
├── go.mod                 # Go module definition
├── go.sum                 # Go module checksums
├── main.go                # Application entry point
├── schema.sql             # Database schema
├── queries.sql            # SQL queries
├── sqlc.yaml              # SQLC configuration
└── SmsGW.yaml             # Application configuration
```

## Development Setup

### 1. Clone Repository

```bash
git clone <repository-url>
cd sms
```

### 2. Install Dependencies

```bash
# Install Go dependencies
go mod download

# Install SQLC (if not already installed)
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

### 3. Database Setup

#### Local PostgreSQL

```bash
# Install PostgreSQL (Ubuntu/Debian)
sudo apt-get install postgresql postgresql-contrib

# Start PostgreSQL service
sudo systemctl start postgresql

# Create database and user
sudo -u postgres psql
```

```sql
-- Create database
CREATE DATABASE sms_db;

-- Create user
CREATE USER sms_user WITH PASSWORD 'dev_password';

-- Grant privileges
GRANT ALL PRIVILEGES ON DATABASE sms_db TO sms_user;

-- Exit PostgreSQL
\q
```

#### Docker PostgreSQL

```bash
# Run PostgreSQL in Docker
docker run --name postgres-dev \
  -e POSTGRES_DB=sms_db \
  -e POSTGRES_USER=sms_user \
  -e POSTGRES_PASSWORD=dev_password \
  -p 5432:5432 \
  -d postgres:13
```

### 4. NATS Setup

#### Local NATS

```bash
# Install NATS server
go install github.com/nats-io/nats-server/v2@latest

# Start NATS server
nats-server --jetstream
```

#### Docker NATS

```bash
# Run NATS in Docker
docker run --name nats-dev \
  -p 4222:4222 \
  -p 8222:8222 \
  -d nats:latest --jetstream
```

### 5. Configuration

Create development configuration:

```yaml
# SmsGW.yaml
api:
  nats:
    address: "127.0.0.1:4222"
  listen: "127.0.0.1:8080"
  postgres:
    address: "127.0.0.1"
    port: 5432
    username: "sms_user"
    password: "dev_password"

worker:
  nats:
    address: "127.0.0.1:4222"
  postgres:
    address: "127.0.0.1"
    port: 5432
    username: "sms_user"
    password: "dev_password"

sms:
  cost: "1.0"
  normal:
    ratelimit: 100
  express:
    ratelimit: 50
```

### 6. Database Schema

```bash
# Create database schema
psql -h localhost -U sms_user -d sms_db -f schema.sql
```

### 7. Generate Database Code

```bash
# Generate SQLC code
sqlc generate
```

## Running the Application

### Development Mode

#### Start API Server

```bash
# Run API server
go run main.go api

# Or with debug logging
LOG_LEVEL=debug go run main.go api
```

#### Start Worker

```bash
# Run worker in another terminal
go run main.go worker
```

#### Test the Application

```bash
# Test API endpoints
curl -X POST "http://localhost:8080/user" \
  -H "Content-Type: application/json" \
  -d '{"username": "test_user", "balance": 100.0}'

curl -X POST "http://localhost:8080/phone-number" \
  -H "Content-Type: application/json" \
  -d '{"user_id": 1, "phone_number": "+1234567890"}'

curl -X POST "http://localhost:8080/sms" \
  -H "Content-Type: application/json" \
  -d '{"user_id": 1, "phone_number_id": 1, "to_phone_number": "+0987654321", "message": "Hello World"}'
```

## Code Generation

### SQLC Code Generation

The project uses SQLC to generate type-safe Go code from SQL queries.

#### Configuration

```yaml
# sqlc.yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "queries.sql"
    schema: "schema.sql"
    gen:
      go:
        package: "sqlc"
        out: "sqlc"
        sql_package: "pgx/v5"
        emit_json_tags: true
        emit_prepared_queries: false
        emit_interface: true
        emit_exact_table_names: false
        emit_empty_slices: true
```

#### Generate Code

```bash
# Generate database code
sqlc generate

# Verify generated code
ls -la sqlc/
```

## Testing

### Unit Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Run specific test
go test -v ./pkg/utils
```

### Integration Tests

```bash
# Run integration tests (requires running services)
go test -tags=integration ./...

# Run tests with database
go test -v ./sqlc
```

### Test Data

The `curl/` directory contains test data:

```bash
# Test with sample data
curl -X POST "http://localhost:8080/user" \
  -H "Content-Type: application/json" \
  -d @curl/new_user.json

curl -X POST "http://localhost:8080/phone-number" \
  -H "Content-Type: application/json" \
  -d @curl/new_phone.json

curl -X POST "http://localhost:8080/sms" \
  -H "Content-Type: application/json" \
  -d @curl/new_sms.json
```

## Code Quality

### Linting

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
golangci-lint run

# Run linter with specific checks
golangci-lint run --enable=gofmt,goimports,vet,errcheck
```

### Formatting

```bash
# Format code
go fmt ./...

# Organize imports
goimports -w .
```

### Code Review Checklist

- [ ] Code follows Go conventions
- [ ] Functions have proper error handling
- [ ] Database operations use transactions
- [ ] NATS messages are properly acknowledged
- [ ] Configuration is externalized
- [ ] Logging is appropriate
- [ ] Tests cover new functionality
- [ ] Documentation is updated

## Debugging

### Debug Logging

```bash
# Enable debug logging
LOG_LEVEL=debug go run main.go api

# Enable debug logging for worker
LOG_LEVEL=debug go run main.go worker
```

### Database Debugging

```bash
# Connect to database
psql -h localhost -U sms_user -d sms_db

# Check tables
\dt

# Check data
SELECT * FROM users;
SELECT * FROM phone_numbers;
SELECT * FROM sms;
```

### NATS Debugging

```bash
# Install NATS CLI
go install github.com/nats-io/natscli/nats@latest

# Check NATS server info
nats server info

# List streams
nats stream list

# Check consumer status
nats consumer list

# Monitor messages
nats sub "sms.send.request"
```

## Performance Profiling

### CPU Profiling

```bash
# Enable CPU profiling
go run main.go api --cpuprofile=cpu.prof

# Analyze profile
go tool pprof cpu.prof
```

### Memory Profiling

```bash
# Enable memory profiling
go run main.go api --memprofile=mem.prof

# Analyze profile
go tool pprof mem.prof
```

### Benchmarking

```bash
# Run benchmarks
go test -bench=. ./...

# Run specific benchmark
go test -bench=BenchmarkSmsProcessing ./...
```

## Docker Development

### Dockerfile

```dockerfile
# Dockerfile
FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o sms main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/sms .
COPY --from=builder /app/SmsGW.yaml .

CMD ["./sms"]
```

### Docker Compose

```yaml
# docker-compose.yml
version: '3.8'

services:
  postgres:
    image: postgres:13
    environment:
      POSTGRES_DB: sms_db
      POSTGRES_USER: sms_user
      POSTGRES_PASSWORD: dev_password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./schema.sql:/docker-entrypoint-initdb.d/schema.sql

  nats:
    image: nats:latest
    command: ["--jetstream"]
    ports:
      - "4222:4222"
      - "8222:8222"

  sms-api:
    build: .
    command: ["./sms", "api"]
    ports:
      - "8080:8080"
    depends_on:
      - postgres
      - nats
    environment:
      - SMS_API_POSTGRES_ADDRESS=postgres
      - SMS_API_NATS_ADDRESS=nats:4222
      - SMS_SMS_NORMAL_RATELIMIT=100
      - SMS_SMS_EXPRESS_RATELIMIT=50

  sms-worker:
    build: .
    command: ["./sms", "worker"]
    depends_on:
      - postgres
      - nats
    environment:
      - SMS_WORKER_POSTGRES_ADDRESS=postgres
      - SMS_WORKER_NATS_ADDRESS=nats:4222
      - SMS_SMS_NORMAL_RATELIMIT=100
      - SMS_SMS_EXPRESS_RATELIMIT=50

volumes:
  postgres_data:
```

### Development with Docker

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f sms-api
docker-compose logs -f sms-worker

# Stop services
docker-compose down
```

## Contributing

### Git Workflow

1. **Fork Repository**: Fork the repository on GitHub
2. **Create Branch**: Create a feature branch
3. **Make Changes**: Implement your changes
4. **Write Tests**: Add tests for new functionality
5. **Run Tests**: Ensure all tests pass
6. **Submit PR**: Create a pull request

### Commit Messages

Follow conventional commit format:

```
feat: add user authentication
fix: resolve database connection issue
docs: update API documentation
test: add integration tests for SMS processing
```

### Pull Request Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed

## Checklist
- [ ] Code follows project conventions
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] Tests added/updated
```

## Troubleshooting

### Common Issues

1. **Database Connection Failed**
   ```
   Error: failed to connect to database
   ```
   **Solution**: Check PostgreSQL is running and credentials are correct

2. **NATS Connection Failed**
   ```
   Error: failed to connect to NATS
   ```
   **Solution**: Check NATS server is running and accessible

3. **SQLC Generation Failed**
   ```
   Error: sqlc generate failed
   ```
   **Solution**: Check SQL syntax and sqlc.yaml configuration

4. **Port Already in Use**
   ```
   Error: listen tcp :8080: bind: address already in use
   ```
   **Solution**: Change port in configuration or kill existing process

### Debug Commands

```bash
# Check Go version
go version

# Check module status
go mod tidy
go mod verify

# Check database connectivity
psql -h localhost -U sms_user -d sms_db -c "SELECT 1;"

# Check NATS connectivity
nats server info

# Check application status
curl http://localhost:8080/health
```

## Future Enhancements

### Planned Features

- **Authentication**: JWT-based authentication
- **Rate Limiting**: Per-user rate limiting
- **Metrics**: Prometheus metrics integration
- **Tracing**: Distributed tracing with Jaeger
- **Testing**: Comprehensive test suite
- **CI/CD**: Automated testing and deployment
- **Documentation**: API documentation with Swagger
- **Monitoring**: Health checks and monitoring