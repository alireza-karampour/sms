# Message Queue System

## Overview

The SMS Gateway uses NATS JetStream as its message queue system to provide reliable, scalable message processing with priority-based queuing.

## NATS JetStream Configuration

### Connection

The system connects to NATS using the following configuration:

```yaml
api:
  nats:
    address: "127.0.0.1:4222"

worker:
  nats:
    address: "127.0.0.1:4222"
```

### Connection Code

```go
func Connect(addr string) (*nats.Conn, error) {
    nc, err := nats.Connect(fmt.Sprintf("nats://%s", addr))
    if err != nil {
        return nil, err
    }
    return nc, nil
}
```

## Streams

The system defines two JetStream streams for different priority levels:

### 1. Normal SMS Stream (`Sms`)

**Configuration**:
```go
jetstream.StreamConfig{
    Name:        "Sms",
    Description: "work queue for handling sms with normal priority",
    Subjects: []string{
        "sms.send.request",
        "sms.send.status", 
        "sms.send.error",
    },
    Retention: jetstream.WorkQueuePolicy,
    Storage:   jetstream.FileStorage,
}
```

**Characteristics**:
- **Retention Policy**: Work Queue (messages are removed after acknowledgment)
- **Storage**: File Storage (persistent)
- **Subjects**: `sms.send.*`

### 2. Express SMS Stream (`SmsExpress`)

**Configuration**:
```go
jetstream.StreamConfig{
    Name:        "SmsExpress",
    Description: "work queue for handling sms with high priority",
    Subjects: []string{
        "sms.ex.send.request",
        "sms.ex.send.status",
        "sms.ex.send.error",
    },
    Retention: jetstream.WorkQueuePolicy,
    Storage:   jetstream.FileStorage,
}
```

**Characteristics**:
- **Retention Policy**: Work Queue (messages are removed after acknowledgment)
- **Storage**: File Storage (persistent)
- **Subjects**: `sms.ex.send.*`

## Subject Naming Convention

The system uses a hierarchical subject naming convention:

### Subject Structure

```
{service}.{priority}.{action}.{type}
```

### Subject Components

- **Service**: `sms` - Identifies the SMS service
- **Priority**: `send` (normal) or `ex` (express)
- **Action**: `send` - The action being performed
- **Type**: `request`, `status`, `error`

### Subject Examples

| Subject | Description |
|---------|-------------|
| `sms.send.request` | Normal SMS send request |
| `sms.send.status` | Normal SMS status update |
| `sms.send.error` | Normal SMS error |
| `sms.ex.send.request` | Express SMS send request |
| `sms.ex.send.status` | Express SMS status update |
| `sms.ex.send.error` | Express SMS error |

### Subject Generation

Subjects are generated using the `MakeSubject` utility function:

```go
func MakeSubject(s ...string) string {
    return strings.Join(s, ".")
}

// Examples:
MakeSubject("sms", "send", "request")     // "sms.send.request"
MakeSubject("sms", "ex", "send", "request") // "sms.ex.send.request"
```

## Message Publishing

### Publisher Configuration

The API server uses a simple publisher for sending messages:

```go
type Publisher struct {
    JetStream jetstream.JetStream
}

func NewSimplePublisher(nc *nats.Conn) (*Publisher, error) {
    js, err := nc.JetStream()
    if err != nil {
        return nil, err
    }
    return &Publisher{JetStream: js}, nil
}
```

### Publishing Messages

```go
// Publish SMS request
smsJson, err := json.Marshal(sms)
if err != nil {
    return err
}

_, err = publisher.JetStream.Publish(ctx, subject, smsJson)
if err != nil {
    return err
}
```

## Message Consumption

### Consumer Configuration

The worker uses consumers to process messages:

```go
type Consumer struct {
    JetStream jetstream.JetStream
}

func NewConsumer(nc *nats.Conn) (*Consumer, error) {
    js, err := nc.JetStream()
    if err != nil {
        return nil, err
    }
    return &Consumer{JetStream: js}, nil
}
```

### Consumer Setup

```go
// Normal SMS Consumer
normalSms := &StreamConsumersConfig{
    Stream: jetstream.StreamConfig{
        Name:        "Sms",
        Description: "work queue for handling sms with normal priority",
        Subjects: []string{
            "sms.send.request",
            "sms.send.status",
            "sms.send.error",
        },
        Retention:   jetstream.WorkQueuePolicy,
        Storage:     jetstream.FileStorage,
        AllowDirect: true,
    },
    Consumers: []jetstream.ConsumerConfig{
        {
            Name:        "Sms",
            Durable:     "Sms",
            Description: "consumes normal sms work queue",
        },
    },
}
```

### Message Processing

```go
func (s *Sms) handler(msg jetstream.Msg) {
    sub := Subject(msg.Subject())
    switch {
    case sub.Filter("sms", "send", "*"):
        s.handleNormalSms(msg)
    case sub.Filter("sms", "ex", "*", "*"):
        s.handleExpressSms(msg)
    }
}
```

### Rate Limiting

Both normal and express SMS handlers implement configurable rate limiting:

```go
// Normal SMS rate limiting
rate := sync.OnceValue(func() uint {
    return viper.GetUint("sms.normal.ratelimit")
})()

// Express SMS rate limiting  
rate := sync.OnceValue(func() uint {
    return viper.GetUint("sms.express.ratelimit")
})()
```

**Rate Limiting Features**:
- **Configurable**: Different rate limits for normal vs express SMS
- **Efficient**: Uses `sync.OnceValue` for optimal performance
- **Consistent**: Same pattern for both SMS types
- **Flexible**: Can be adjusted via configuration

## Message Acknowledgment

### Acknowledgment Types

1. **ACK**: Standard acknowledgment
2. **NAK**: Negative acknowledgment (retry)
3. **Double ACK**: Confirmation of successful processing
4. **Term**: Terminate message (no retry)

### Acknowledgment Flow

```go
// Successful processing
err = msg.DoubleAck(context.Background())
if err != nil {
    logrus.Errorf("failed to DoubleAck: %s", err.Error())
    return
}

// Error handling with retry
err = msg.NakWithDelay(time.Second)
if err != nil {
    logrus.Errorf("failed to NAK: %s\n", err.Error())
}

// Terminate message (no retry)
msg.TermWithReason(err.Error())
```

## Error Handling

### Message Processing Errors

```go
func (s *Sms) errHandler(ctx jetstream.ConsumeContext, err error) {
    logrus.Errorf("ConsumerError: %s\n", err)
}
```

### Retry Logic

- **NAK with Delay**: Retry message after delay
- **Terminate**: Stop processing message permanently
- **Logging**: Comprehensive error logging

### Transaction Safety

```go
tx, err := s.db.Begin(context.Background())
if err != nil {
    logrus.Errorf("failed to begin tx: %s\n", err.Error())
    err := msg.NakWithDelay(time.Second)
    return
}
defer tx.Rollback(context.Background())

// Process message
// ...

// Commit transaction
tx.Commit(context.Background())

// Acknowledge message
msg.DoubleAck(context.Background())
```

## Performance Considerations

### Message Batching

JetStream supports batch processing for improved performance:

```go
opts := []jetstream.PullConsumeOpt{
    jetstream.ConsumeErrHandler(s.errHandler),
    jetstream.MaxBatch(10), // Process up to 10 messages at once
}
```

### Connection Pooling

- Single NATS connection per service
- Connection reuse across operations
- Automatic reconnection on failure

### Memory Management

- File storage for persistence
- Work queue retention policy
- Automatic message cleanup after acknowledgment

## Monitoring and Observability

### Logging

```go
logrus.SetLevel(logrus.DebugLevel)
logrus.SetFormatter(&logrus.TextFormatter{
    ForceColors:            true,
    DisableLevelTruncation: true,
})
```

### Metrics (Future Enhancement)

Planned metrics include:
- Message throughput
- Queue depth
- Processing latency
- Error rates
- Consumer lag

## Deployment Considerations

### NATS Server Configuration

For production deployment, consider:

```yaml
# nats-server.conf
jetstream: {
    store_dir: "/data/jetstream"
    max_memory: 1GB
    max_file: 10GB
}

cluster: {
    name: "sms-cluster"
    routes: ["nats://nats-1:6222", "nats://nats-2:6222"]
}
```

### Kubernetes Deployment

The project includes Helm charts for NATS deployment:

```bash
# Deploy NATS cluster
helm install nats ./charts/nats

# Deploy PostgreSQL
helm install postgres ./charts/postgres
```

## Security Considerations

### Connection Security

- TLS encryption (future enhancement)
- Authentication tokens (future enhancement)
- Network isolation

### Message Security

- Message encryption (future enhancement)
- Access control (future enhancement)
- Audit logging (future enhancement)

## Troubleshooting

### Common Issues

1. **Connection Failures**
   - Check NATS server status
   - Verify network connectivity
   - Check configuration parameters

2. **Message Processing Errors**
   - Review error logs
   - Check database connectivity
   - Verify message format

3. **Performance Issues**
   - Monitor queue depth
   - Check consumer capacity
   - Review database performance

### Debug Commands

```bash
# Check NATS server status
nats server info

# List streams
nats stream list

# Check consumer status
nats consumer list

# Monitor messages
nats sub "sms.send.request"
```

## Future Enhancements

### Planned Features

- **Dead Letter Queue**: Handle failed messages
- **Message Encryption**: Secure message transmission
- **Priority Levels**: Multiple priority levels
- **Message Routing**: Intelligent message routing
- **Load Balancing**: Consumer load balancing
- **Metrics Integration**: Prometheus metrics
- **Alerting**: Automated alerting system