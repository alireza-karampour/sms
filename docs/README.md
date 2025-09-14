# SMS Gateway Documentation

## Overview

This is a minimal SMS Gateway system built in Go that provides a REST API for sending SMS messages with a message queue-based architecture using NATS JetStream. The system supports both normal and express (high-priority) SMS delivery with user balance management.

## Architecture

The system follows a microservices architecture with two main components:

1. **API Server** (`cmd/api`) - REST API for SMS operations
2. **Worker** (`cmd/worker`) - Background worker for processing SMS messages

## Key Features

- **REST API** for SMS operations using Gin framework
- **Message Queue System** using NATS JetStream for reliable message processing
- **Priority Queues** - Normal and Express SMS handling
- **User Balance Management** - Automatic balance deduction for SMS costs
- **Database Integration** - PostgreSQL with SQLC for type-safe queries
- **Configuration Management** - Viper-based configuration system
- **Docker Support** - Kubernetes charts for deployment

## Technology Stack

- **Language**: Go 1.25
- **Web Framework**: Gin
- **Message Queue**: NATS JetStream
- **Database**: PostgreSQL
- **ORM**: SQLC (SQL code generation)
- **Configuration**: Viper
- **CLI**: Cobra
- **Logging**: Logrus
- **Testing**: Ginkgo/Gomega
- **Database Driver**: pgx/v5

## Quick Start

1. **Prerequisites**:
   - Go 1.25+
   - PostgreSQL
   - NATS Server

2. **Configuration**:
   - Copy `SmsGW.yaml` and adjust database/NATS settings
   - Set SMS cost in configuration

3. **Run the API Server**:
   ```bash
   go run main.go api
   ```

4. **Run the Worker**:
   ```bash
   go run main.go worker
   ```

## Documentation Structure

- [Architecture](architecture.md) - Detailed system architecture
- [API Reference](api-reference.md) - REST API documentation
- [Database Schema](database-schema.md) - Database structure and relationships
- [Message Queue](message-queue.md) - NATS JetStream configuration and usage
- [Configuration](configuration.md) - Configuration options and setup
- [Deployment](deployment.md) - Kubernetes deployment guide
- [Development](development.md) - Development setup and guidelines

## Project Structure

```
sms/
├── cmd/                    # Application entry points
│   ├── api/               # REST API server
│   ├── worker/            # Background worker
│   └── root.go            # CLI root command
├── internal/              # Private application code
│   ├── controllers/       # HTTP controllers (User, PhoneNumber, Sms)
│   ├── streams/           # NATS stream constants
│   ├── subjects/          # NATS subject constants
│   └── workers/           # Background workers (Sms)
├── pkg/                   # Public packages
│   ├── middlewares/       # HTTP middlewares
│   ├── nats/              # NATS utilities (Publisher, Consumer)
│   └── utils/             # Common utilities
├── sqlc/                  # Generated database code
├── charts/                # Kubernetes Helm charts
├── curl/                  # Test data and examples
├── tests/                 # Integration tests
├── docs/                  # Documentation
├── schema.sql             # Database schema
├── queries.sql            # SQL queries for SQLC
├── SmsGW.yaml            # Application configuration
└── main.go                # Application entry point
```

## Contributing

Please refer to the [Development Guide](development.md) for setup instructions and coding standards.