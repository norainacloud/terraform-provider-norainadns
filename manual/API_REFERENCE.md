# NorainaDNS API Reference

This document describes the NorainaDNS REST API that the Terraform provider uses.

## Authentication

All API requests must include the API key as a Bearer token in the `Authorization` header:

```http
Authorization: Bearer nordns_your_api_key_here
```

API keys can be generated from the NorainaDNS web interface under **Settings > API Keys**.

## Base URL

```
https://dns.noraina.cloud/api
```

## Endpoints

### Zones

#### List Zones

```http
GET /zones
```

**Response:**
```json
{
  "data": [
    {
      "id": 1,
      "domainName": "example.com",
      "description": "Production domain",
      "isActive": true,
      "syncStatus": "synced",
      "nameservers": ["ns1.noraina.cloud", "ns2.noraina.cloud"]
    }
  ]
}
```

#### Get Zone

```http
GET /zones/:id
```

#### Create Zone

```http
POST /zones
Content-Type: application/json

{
  "domainName": "example.com",
  "description": "My domain"
}
```

#### Update Zone

```http
PUT /zones/:id
Content-Type: application/json

{
  "description": "Updated description"
}
```

#### Delete Zone

```http
DELETE /zones/:id
```

---

### Records

#### List Records

```http
GET /zones/:zoneId/records
```

**Response:**
```json
{
  "data": [
    {
      "id": 1,
      "zoneId": 1,
      "name": "www",
      "recordType": "A",
      "ttl": 300,
      "isDynamic": false,
      "routingPolicy": "simple",
      "syncStatus": "synced",
      "values": [
        {
          "id": 1,
          "value": "192.0.2.1",
          "priority": 0,
          "weight": 0,
          "monitorId": null
        }
      ]
    }
  ]
}
```

#### Get Record

```http
GET /zones/:zoneId/records/:recordId
```

#### Create Record

```http
POST /zones/:zoneId/records
Content-Type: application/json

{
  "name": "www",
  "recordType": "A",
  "ttl": 300,
  "isDynamic": false,
  "routingPolicy": "simple",
  "values": [
    {
      "value": "192.0.2.1",
      "priority": 0,
      "weight": 0
    }
  ]
}
```

**Failover Record Example:**
```json
{
  "name": "app",
  "recordType": "A",
  "ttl": 60,
  "isDynamic": true,
  "routingPolicy": "failover",
  "values": [
    {
      "value": "192.0.2.10",
      "priority": 0,
      "monitorId": 123
    },
    {
      "value": "192.0.2.20",
      "priority": 1
    }
  ]
}
```

#### Update Record

```http
PUT /zones/:zoneId/records/:recordId
Content-Type: application/json

{
  "name": "www",
  "ttl": 600,
  "isDynamic": false,
  "routingPolicy": "simple",
  "values": [
    {
      "value": "192.0.2.2",
      "priority": 0
    }
  ]
}
```

#### Delete Record

```http
DELETE /zones/:zoneId/records/:recordId
```

---

### Monitors

#### List Monitors

```http
GET /monitors
```

**Response:**
```json
{
  "data": [
    {
      "id": 1,
      "name": "Origin Server",
      "description": "Health check for origin",
      "protocol": "HTTPS",
      "targetHost": "www.example.com",
      "targetPort": 443,
      "targetPath": "/health",
      "checkIntervalSeconds": 300,
      "failureThreshold": 3,
      "timeoutSeconds": 10,
      "customHostHeader": null,
      "customSni": null,
      "forceIpAddress": "192.0.2.100",
      "isActive": true,
      "currentStatus": "healthy"
    }
  ]
}
```

#### Get Monitor

```http
GET /monitors/:id
```

#### Create Monitor

```http
POST /monitors
Content-Type: application/json

{
  "name": "Origin Server",
  "description": "Health check for origin",
  "protocol": "HTTPS",
  "targetHost": "www.example.com",
  "targetPort": 443,
  "targetPath": "/health",
  "checkIntervalSeconds": 300,
  "failureThreshold": 3,
  "timeoutSeconds": 10,
  "forceIpAddress": "192.0.2.100"
}
```

#### Update Monitor

```http
PUT /monitors/:id
Content-Type: application/json

{
  "name": "Origin Server Updated",
  "targetPath": "/healthz",
  "timeoutSeconds": 15
}
```

#### Delete Monitor

```http
DELETE /monitors/:id
```

**Error Response (Monitor In Use):**
```json
{
  "error": {
    "code": "MONITOR_IN_USE",
    "message": "Monitor is currently in use by DNS records"
  }
}
```

---

## Error Responses

All error responses follow this format:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message",
    "details": {}
  }
}
```

### Common Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| UNAUTHORIZED | 401 | Invalid or missing API key |
| FORBIDDEN | 403 | API key doesn't have permission |
| ZONE_NOT_FOUND | 404 | Zone doesn't exist |
| RECORD_NOT_FOUND | 404 | Record doesn't exist |
| MONITOR_NOT_FOUND | 404 | Monitor doesn't exist |
| ZONE_ALREADY_EXISTS | 409 | Zone with this domain already exists |
| DUPLICATE_RECORD | 409 | Record with same name/type exists |
| MONITOR_IN_USE | 409 | Monitor is referenced by records |
| VALIDATION_ERROR | 422 | Invalid request data |
| INTERNAL_ERROR | 500 | Server error |

---

## Rate Limiting

The API has the following rate limits:

- **100 requests per minute** per API key
- **1000 requests per hour** per API key

Rate limit headers are included in responses:

```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1640000000
```

---

## Webhooks

NorainaDNS uses webhooks to receive health status updates from the monitors. These are configured automatically when monitors are created.

The webhook endpoints are:
- AWS Lambda: Triggered by API Gateway
- GCP Cloud Run: Direct HTTP endpoint

Both endpoints update DNS records in their respective cloud providers when health status changes.
