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
https://dash.norainadns.com/api/v1
```

## Endpoints

Paths below are relative to the base URL. All identifiers (`id`, `zoneId`, `monitorId`) are UUID strings.

Successful responses are wrapped as `{ "success": true, "data": ... }`.

### Zones

#### List Zones

```http
GET /zones
```

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "3f2b1c9e-8a4d-4e6f-9b1a-2c7d5e8f0a13",
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
  "success": true,
  "data": [
    {
      "id": "b7e4a2d1-5c3f-4a8b-9e6d-1f0c2a3b4d5e",
      "zoneId": "3f2b1c9e-8a4d-4e6f-9b1a-2c7d5e8f0a13",
      "name": "www",
      "recordType": "A",
      "ttl": 300,
      "isDynamic": false,
      "routingPolicy": "simple",
      "syncStatus": "synced",
      "values": [
        {
          "id": "c1d2e3f4-a5b6-4c7d-8e9f-0a1b2c3d4e5f",
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
      "monitorId": "9a8b7c6d-5e4f-4a3b-8c2d-1e0f9a8b7c6d"
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
  "success": true,
  "data": [
    {
      "id": "9a8b7c6d-5e4f-4a3b-8c2d-1e0f9a8b7c6d",
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
  "success": false,
  "error": {
    "code": "CONFLICT",
    "message": "Cannot delete a monitor that is in use by records"
  }
}
```

---

## Error Responses

All error responses follow this format:

```json
{
  "success": false,
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
| VALIDATION_ERROR | 400 | Invalid request data (`details.errors` lists the failing fields) |
| INVALID_DOMAIN | 400 | Invalid domain name |
| AUTH_FAILED | 401 | Missing or invalid credentials |
| INVALID_API_KEY | 401 | Invalid or expired API key |
| ACCESS_DENIED | 403 | Not allowed to perform this action |
| INSUFFICIENT_PERMISSIONS | 403 | API key lacks the required permission |
| ZONE_ACCESS_DENIED | 403 | API key is not scoped to this zone |
| QUOTA_EXCEEDED | 403 | Tenant zone or monitor quota reached |
| ZONE_NOT_FOUND | 404 | Zone doesn't exist |
| RECORD_NOT_FOUND | 404 | Record doesn't exist |
| MONITOR_NOT_FOUND | 404 | Monitor doesn't exist |
| DUPLICATE_ZONE | 409 | Zone with this domain already exists |
| DUPLICATE_RECORD | 409 | Record with same name/type exists |
| CONFLICT | 409 | Other conflicts, e.g. deleting a monitor still used by records |
| RATE_LIMIT_EXCEEDED | 429 | Too many requests |
| PROVIDER_ERROR | 502 | Upstream DNS provider error |
| PROVIDER_SYNC_ERROR | 502 | Change could not be applied to all DNS providers |
| INTERNAL_ERROR | 500 | Server error |

---

## Rate Limiting

The API has the following rate limits:

- **100 requests per minute** per API key
- **1000 requests per hour** per API key

Exceeding a limit returns `429` with error code `RATE_LIMIT_EXCEEDED`. Responses include the standard `RateLimit-Policy`, `RateLimit-Limit`, `RateLimit-Remaining` and `RateLimit-Reset` headers (`RateLimit-Reset` is the number of seconds until the window resets). Because both limits apply, these headers report the hourly window:

```http
RateLimit-Policy: 1000;w=3600
RateLimit-Limit: 1000
RateLimit-Remaining: 995
RateLimit-Reset: 1800
```

---

## Webhooks

NorainaDNS uses webhooks to receive health status updates from the monitors. These are configured automatically when monitors are created.

The webhook endpoints are:
- AWS Lambda: Triggered by API Gateway
- GCP Cloud Run: Direct HTTP endpoint

Both endpoints update DNS records in their respective cloud providers when health status changes.
