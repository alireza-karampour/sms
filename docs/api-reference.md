# API Reference

## Base URL

```
http://localhost:8080
```

## Authentication

Currently, the API does not implement authentication. All endpoints are publicly accessible.

## Endpoints

### SMS Operations

#### Send SMS

Send an SMS message to a phone number.

**Endpoint**: `POST /sms`

**Query Parameters**:
- `express` (boolean, optional): Set to `true` for express (high-priority) SMS delivery

**Request Body**:
```json
{
  "user_id": 1,
  "phone_number_id": 1,
  "to_phone_number": "+1234567890",
  "message": "Hello, this is a test SMS",
  "status": "pending"
}
```

**Request Body Schema**:
- `user_id` (integer, required): ID of the user sending the SMS
- `phone_number_id` (integer, required): ID of the phone number to use for sending
- `to_phone_number` (string, required): Destination phone number
- `message` (string, required): SMS message content
- `status` (string, optional): Initial status (defaults to "pending")

**Response**:
```json
{
  "msg": "OK"
}
```

**Status Codes**:
- `200 OK`: SMS queued successfully
- `400 Bad Request`: Invalid request data
- `403 Forbidden`: Insufficient balance
- `500 Internal Server Error`: Server error

**Example Requests**:

Normal SMS:
```bash
curl -X POST "http://localhost:8080/sms" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "phone_number_id": 1,
    "to_phone_number": "+1234567890",
    "message": "Hello World"
  }'
```

Express SMS:
```bash
curl -X POST "http://localhost:8080/sms?express=true" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "phone_number_id": 1,
    "to_phone_number": "+1234567890",
    "message": "Urgent message"
  }'
```

### User Operations

#### Create User

Create a new user account.

**Endpoint**: `POST /user`

**Request Body**:
```json
{
  "username": "john_doe",
  "balance": 100.00
}
```

**Response**:
```json
{
  "id": 1,
  "username": "john_doe",
  "balance": "100.00"
}
```

#### Get User Balance

Retrieve the current balance for a user.

**Endpoint**: `GET /user/{user_id}/balance`

**Path Parameters**:
- `user_id` (integer): User ID

**Response**:
```json
{
  "balance": "95.00"
}
```

#### Add Balance

Add funds to a user's account.

**Endpoint**: `POST /user/{username}/balance`

**Path Parameters**:
- `username` (string): Username

**Request Body**:
```json
{
  "balance": 50.00
}
```

**Response**:
```json
{
  "balance": "145.00"
}
```

### Phone Number Operations

#### Add Phone Number

Add a phone number to a user's account.

**Endpoint**: `POST /phone-number`

**Request Body**:
```json
{
  "user_id": 1,
  "phone_number": "+1987654321"
}
```

**Response**:
```json
{
  "id": 1,
  "user_id": 1,
  "phone_number": "+1987654321"
}
```

## Error Responses

### Standard Error Format

```json
{
  "error": "Error message description"
}
```

### Common Error Codes

- `400 Bad Request`: Invalid request format or missing required fields
- `403 Forbidden`: Insufficient balance for SMS operation
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Internal server error

### Example Error Responses

Insufficient Balance:
```json
{
  "error": "not enough balance"
}
```

Invalid Request:
```json
{
  "error": "invalid request data"
}
```

## Rate Limiting

Currently, no rate limiting is implemented. This is planned for future releases.

## SMS Cost

The cost per SMS is configurable and defaults to 5.0 units. This value is set in the configuration file (`SmsGW.yaml`):

```yaml
sms:
  cost: "5.0"
```

## Message Priority

The system supports two priority levels:

1. **Normal SMS**: Standard priority, processed in order
2. **Express SMS**: High priority, processed before normal SMS

Express SMS messages are processed with higher priority in the message queue system.

## Response Times

- **API Response**: Typically < 100ms for successful requests
- **SMS Processing**: Depends on queue depth and worker capacity
- **Express SMS**: Processed with higher priority than normal SMS

## Testing

### Test Data

The `curl/` directory contains example JSON files for testing:

- `new_user.json`: Sample user creation data
- `new_phone.json`: Sample phone number data
- `new_sms.json`: Sample SMS data

### Example Test Sequence

1. Create a user:
```bash
curl -X POST "http://localhost:8080/user" \
  -H "Content-Type: application/json" \
  -d @curl/new_user.json
```

2. Add a phone number:
```bash
curl -X POST "http://localhost:8080/phone-number" \
  -H "Content-Type: application/json" \
  -d @curl/new_phone.json
```

3. Send an SMS:
```bash
curl -X POST "http://localhost:8080/sms" \
  -H "Content-Type: application/json" \
  -d @curl/new_sms.json
```

## Future Enhancements

Planned API improvements include:

- **Authentication**: JWT-based authentication
- **Rate Limiting**: Per-user rate limiting
- **SMS Status**: Real-time SMS delivery status
- **Bulk SMS**: Send multiple SMS messages in one request
- **SMS Templates**: Predefined message templates
- **Webhooks**: Delivery status notifications
- **Analytics**: SMS usage analytics and reporting