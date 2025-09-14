# Configuration Guide

## Overview

The SMS Gateway uses Viper for configuration management, supporting YAML configuration files with environment variable overrides.

## Configuration File

The main configuration file is `SmsGW.yaml` located in the project root:

```yaml
api:
  nats:
    address: "127.0.0.1:4222"
  listen: "127.0.0.1:8080"
  postgres:
    address: "127.0.0.1"
    port: 5433
    username: root
    password: 1234

worker:
  nats:
    address: "127.0.0.1:4222"
  postgres:
    address: "127.0.0.1"
    port: 5433
    username: root
    password: 1234

sms:
  cost: "5.0"
```

## Configuration Sections

### API Server Configuration

```yaml
api:
  nats:
    address: "127.0.0.1:4222"  # NATS server address
  listen: "127.0.0.1:8080"     # API server listen address
  postgres:
    address: "127.0.0.1"       # PostgreSQL server address
    port: 5433                 # PostgreSQL port
    username: root             # Database username
    password: 1234             # Database password
```

**Parameters**:
- `api.nats.address`: NATS server connection address
- `api.listen`: HTTP server listen address and port
- `api.postgres.address`: PostgreSQL server address
- `api.postgres.port`: PostgreSQL server port
- `api.postgres.username`: Database username
- `api.postgres.password`: Database password

### Worker Configuration

```yaml
worker:
  nats:
    address: "127.0.0.1:4222"  # NATS server address
  postgres:
    address: "127.0.0.1"       # PostgreSQL server address
    port: 5433                 # PostgreSQL port
    username: root             # Database username
    password: 1234             # Database password
```

**Parameters**:
- `worker.nats.address`: NATS server connection address
- `worker.postgres.address`: PostgreSQL server address
- `worker.postgres.port`: PostgreSQL server port
- `worker.postgres.username`: Database username
- `worker.postgres.password`: Database password

### SMS Configuration

```yaml
sms:
  cost: "5.0"  # Cost per SMS in decimal format
  normal:
    ratelimit: 100  # Rate limit for normal SMS in milliseconds
  express:
    ratelimit: 50   # Rate limit for express SMS in milliseconds
```

**Parameters**:
- `sms.cost`: Cost per SMS message (decimal string)
- `sms.normal.ratelimit`: Rate limit for normal SMS messages in milliseconds
- `sms.express.ratelimit`: Rate limit for express SMS messages in milliseconds

## Configuration Loading

### Viper Configuration

The application uses Viper for configuration management:

```go
func init() {
    viper.SetConfigName("SmsGW")
    viper.AddConfigPath(".")
    viper.AddConfigPath("$HOME/.config")
    err := viper.ReadInConfig()
    if err != nil {
        logrus.Errorf("viper failed to read config: %s", err)
        os.Exit(1)
    }
    logrus.Info("config file read")
}
```

### Configuration Paths

Viper searches for configuration files in the following order:

1. Current directory (`.`)
2. User's config directory (`$HOME/.config`)
3. Environment variables (override file values)

### Default Values

Some configuration values have defaults set in code:

```go
viper.SetDefault("api.sms.cost", 5)
```

## Environment Variables

You can override configuration values using environment variables:

### Environment Variable Format

```
SMS_API_NATS_ADDRESS=127.0.0.1:4222
SMS_API_LISTEN=127.0.0.1:8080
SMS_API_POSTGRES_ADDRESS=127.0.0.1
SMS_API_POSTGRES_PORT=5433
SMS_API_POSTGRES_USERNAME=root
SMS_API_POSTGRES_PASSWORD=1234

SMS_WORKER_NATS_ADDRESS=127.0.0.1:4222
SMS_WORKER_POSTGRES_ADDRESS=127.0.0.1
SMS_WORKER_POSTGRES_PORT=5433
SMS_WORKER_POSTGRES_USERNAME=root
SMS_WORKER_POSTGRES_PASSWORD=1234

SMS_SMS_COST=5.0
SMS_SMS_NORMAL_RATELIMIT=100
SMS_SMS_EXPRESS_RATELIMIT=50
```

### Environment Variable Naming

Environment variables follow the pattern:
```
SMS_{SECTION}_{SUBSECTION}_{PARAMETER}
```

Where:
- `SMS` is the prefix
- `SECTION` is the configuration section (API, WORKER, SMS)
- `SUBSECTION` is the subsection (NATS, POSTGRES)
- `PARAMETER` is the parameter name

## Configuration Validation

### Required Parameters

All configuration parameters are required. The application will exit if any required parameter is missing.

### Parameter Types

- **String**: Addresses, usernames, passwords
- **Integer**: Ports
- **Decimal**: SMS cost (stored as string for precision)

### Validation Examples

```go
// Validate NATS connection
natsConn, err := nats.Connect(viper.GetString("api.nats.address"))
if err != nil {
    return err
}

// Validate PostgreSQL connection
pool, err := pgxpool.New(context.Background(), 
    fmt.Sprintf("postgresql://%s:%s@%s:%d",
        viper.GetString("api.postgres.username"),
        viper.GetString("api.postgres.password"),
        viper.GetString("api.postgres.address"),
        viper.GetInt("api.postgres.port"),
    ))
```

## Configuration Examples

### Development Configuration

```yaml
api:
  nats:
    address: "localhost:4222"
  listen: "localhost:8080"
  postgres:
    address: "localhost"
    port: 5432
    username: "sms_user"
    password: "dev_password"

worker:
  nats:
    address: "localhost:4222"
  postgres:
    address: "localhost"
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

### Production Configuration

```yaml
api:
  nats:
    address: "nats-cluster:4222"
  listen: "0.0.0.0:8080"
  postgres:
    address: "postgres-cluster"
    port: 5432
    username: "sms_prod_user"
    password: "secure_production_password"

worker:
  nats:
    address: "nats-cluster:4222"
  postgres:
    address: "postgres-cluster"
    port: 5432
    username: "sms_prod_user"
    password: "secure_production_password"

sms:
  cost: "5.0"
  normal:
    ratelimit: 200
  express:
    ratelimit: 100
```

### Docker Configuration

```yaml
api:
  nats:
    address: "nats:4222"
  listen: "0.0.0.0:8080"
  postgres:
    address: "postgres"
    port: 5432
    username: "sms_user"
    password: "docker_password"

worker:
  nats:
    address: "nats:4222"
  postgres:
    address: "postgres"
    port: 5432
    username: "sms_user"
    password: "docker_password"

sms:
  cost: "5.0"
  normal:
    ratelimit: 150
  express:
    ratelimit: 75
```

## Configuration Management

### Configuration Updates

The application reads configuration at startup. To apply configuration changes:

1. Update the configuration file
2. Restart the application
3. Configuration is reloaded automatically

### Configuration Backup

```bash
# Backup configuration
cp SmsGW.yaml SmsGW.yaml.backup

# Restore configuration
cp SmsGW.yaml.backup SmsGW.yaml
```

### Configuration Versioning

Consider versioning your configuration files:

```bash
# Versioned configuration
SmsGW-dev.yaml
SmsGW-staging.yaml
SmsGW-prod.yaml
```

## Security Considerations

### Sensitive Data

**Never commit sensitive data to version control**:

```bash
# Add to .gitignore
SmsGW.yaml
*.env
secrets/
```

### Environment Variables

Use environment variables for sensitive data:

```bash
# Set sensitive environment variables
export SMS_API_POSTGRES_PASSWORD="secure_password"
export SMS_WORKER_POSTGRES_PASSWORD="secure_password"
```

### Configuration Encryption

For production environments, consider:

- Encrypting configuration files
- Using secret management systems
- Implementing configuration rotation

## Troubleshooting

### Common Configuration Issues

1. **Configuration File Not Found**
   ```
   Error: viper failed to read config: open SmsGW.yaml: no such file or directory
   ```
   **Solution**: Ensure `SmsGW.yaml` exists in the current directory

2. **Invalid Configuration Values**
   ```
   Error: invalid configuration value
   ```
   **Solution**: Check parameter types and values

3. **Connection Failures**
   ```
   Error: failed to connect to database
   ```
   **Solution**: Verify database connection parameters

### Configuration Validation

```bash
# Validate configuration syntax
yamllint SmsGW.yaml

# Test configuration loading
go run main.go --help
```

### Debug Configuration

Enable debug logging to see configuration values:

```go
logrus.SetLevel(logrus.DebugLevel)
logrus.Debugf("Configuration loaded: %+v", viper.AllSettings())
```

## Future Enhancements

### Planned Features

- **Configuration Validation**: Schema validation for configuration files
- **Hot Reloading**: Reload configuration without restart
- **Configuration Templates**: Template-based configuration generation
- **Secret Management**: Integration with secret management systems
- **Configuration UI**: Web-based configuration management
- **Configuration Monitoring**: Monitor configuration changes and drift