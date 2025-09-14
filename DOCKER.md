# Docker Setup for SMS Gateway

This document describes how to run the SMS Gateway using Docker containers.

## Overview

The SMS Gateway consists of two main services:
- **API Service**: REST API server for handling SMS requests
- **Worker Service**: Background worker for processing SMS messages

Both services are dockerized and can be run together with their dependencies (PostgreSQL and NATS) using Docker Compose.

## Files Created

### Dockerfiles
- `Dockerfile.api`: Docker image for the API service
- `Dockerfile.worker`: Docker image for the worker service

### Docker Compose Files
- `docker-compose.yml`: Production setup with all services
- `docker-compose.e2e.yml`: E2E testing setup with isolated services
- `docker-compose.test.yml`: Development testing setup (existing)

### Scripts
- `scripts/run-e2e-tests.sh`: Automated E2E test runner with infrastructure setup

### Configuration
- `Makefile`: Convenient commands for building and running services
- `.dockerignore`: Optimizes Docker build by excluding unnecessary files

## Quick Start

### 1. Build Docker Images
```bash
make docker-build
```

### 2. Start All Services
```bash
make docker-up
```

### 3. Check Service Status
```bash
docker compose ps
```

### 4. View Logs
```bash
make docker-logs
```

### 5. Stop Services
```bash
make docker-down
```

## Running E2E Tests

The E2E tests can be run with automatic infrastructure setup:

```bash
make run-e2e-tests
```

This command will:
1. Build Docker images for API and worker services
2. Start PostgreSQL, NATS, API, and worker services
3. Wait for all services to be healthy
4. Run the E2E tests
5. Clean up all resources when done

## Service Architecture

### API Service
- **Port**: 8080
- **Health Check**: `GET /health`
- **Dependencies**: PostgreSQL, NATS
- **Configuration**: Uses `SmsGW.yaml`

### Worker Service
- **Dependencies**: PostgreSQL, NATS
- **Configuration**: Uses `SmsGW.yaml`
- **Health Check**: Process-based (checks if worker process is running)

### PostgreSQL
- **Port**: 5432 (production), 5434 (E2E tests)
- **Database**: postgres
- **User**: root
- **Password**: 1234
- **Schema**: Auto-initialized from `schema.sql`

### NATS
- **Port**: 4222 (production), 4223 (E2E tests)
- **HTTP Port**: 8222 (production), 8223 (E2E tests)
- **Features**: JetStream enabled for message persistence

## Development Workflow

### Local Development
1. Start infrastructure services:
   ```bash
   make dev-setup
   ```

2. Run API locally:
   ```bash
   make run-api
   ```

3. Run worker locally:
   ```bash
   make run-worker
   ```

4. Clean up:
   ```bash
   make dev-cleanup
   ```

### Testing
- **Integration Tests**: `make integration-test`
- **E2E Tests**: `make run-e2e-tests`
- **All Tests**: `make test`

## Configuration

The services use the `SmsGW.yaml` configuration file. Key settings:

```yaml
api:
  listen: "127.0.0.1:8080"
  postgres:
    address: "postgres"
    port: 5432
    username: "root"
    password: "1234"
  nats:
    address: "nats:4222"

worker:
  postgres:
    address: "postgres"
    port: 5432
    username: "root"
    password: "1234"
  nats:
    address: "nats:4222"
```

## Health Checks

All services include health checks:

- **API**: HTTP endpoint at `/health`
- **Worker**: Process-based check
- **PostgreSQL**: `pg_isready` command
- **NATS**: HTTP endpoint at `/healthz`

## Troubleshooting

### Services Not Starting
1. Check Docker is running: `docker info`
2. Check service logs: `make docker-logs`
3. Check service status: `docker compose ps`

### E2E Tests Failing
1. Ensure all services are healthy before tests run
2. Check API is responding: `curl http://localhost:8080/health`
3. Check logs for specific errors

### Network Issues
- Ensure ports 8080, 5432, and 4222 are available
- For E2E tests, ports 8080, 5434, and 4223 are used

## Security Notes

- Services run as non-root users inside containers
- Database credentials are hardcoded for development (change for production)
- No TLS/SSL configured (add for production use)
- Health check endpoints are public (consider authentication for production)

## Production Considerations

For production deployment:
1. Use environment variables for sensitive configuration
2. Enable TLS/SSL for all services
3. Use proper secrets management
4. Configure proper logging and monitoring
5. Set up proper backup strategies for PostgreSQL and NATS
6. Use container orchestration (Kubernetes, Docker Swarm)
7. Implement proper health checks and monitoring
8. Configure resource limits and scaling policies