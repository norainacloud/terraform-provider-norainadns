---
page_title: "norainadns_monitor Resource - NorainaDNS"
subcategory: ""
description: |-
  Manages a health monitor in NorainaDNS.
---

# norainadns_monitor (Resource)

Manages a health monitor in NorainaDNS.

Monitors check the health of your endpoints and are used to enable automatic DNS failover. NorainaDNS monitors health periodically.

## Example Usage

### Basic HTTPS Monitor

```hcl
resource "norainadns_monitor" "web" {
  name        = "Web Server"
  protocol    = "HTTPS"
  target_host = "www.example.com"
  target_path = "/health"
}
```

### HTTP Monitor with Custom Port

```hcl
resource "norainadns_monitor" "api" {
  name        = "API Server"
  protocol    = "HTTP"
  target_host = "api.example.com"
  target_port = 8080
  target_path = "/status"
}
```

### Origin Monitor (Bypass CDN/WAF)

When your application is behind a CDN or WAF, you may want to monitor the origin server directly:

```hcl
resource "norainadns_monitor" "origin" {
  name             = "Origin Server"
  protocol         = "HTTPS"
  target_host      = "cdn.example.com"    # Public hostname for Host header
  target_path      = "/health"
  force_ip_address = "192.0.2.50"         # Origin server IP
}
```

### Monitor with Custom Settings

```hcl
resource "norainadns_monitor" "critical" {
  name                   = "Critical Service"
  description            = "Main application health check"
  protocol               = "HTTPS"
  target_host            = "app.example.com"
  target_path            = "/healthz"
  check_interval_seconds = 60
  failure_threshold      = 2
  timeout_seconds        = 10
}
```

### Monitor with Custom Host Header and SNI

```hcl
resource "norainadns_monitor" "multi_tenant" {
  name               = "Tenant App"
  protocol           = "HTTPS"
  target_host        = "shared-server.example.com"
  target_path        = "/health"
  force_ip_address   = "192.0.2.100"
  custom_host_header = "tenant1.example.com"
  custom_sni         = "tenant1.example.com"
}
```

### Complete Failover Setup

```hcl
# Create the zone
resource "norainadns_zone" "example" {
  domain_name = "example.com"
}

# Create a monitor for the origin
resource "norainadns_monitor" "origin" {
  name             = "Origin Health"
  protocol         = "HTTPS"
  target_host      = "example.com"
  target_path      = "/health"
  force_ip_address = "192.0.2.20"  # Origin IP
}

# Create a failover record
resource "norainadns_record" "www" {
  zone_id        = norainadns_zone.example.id
  name           = "www"
  type           = "A"
  ttl            = 60
  is_dynamic     = true
  routing_policy = "failover"

  values {
    value      = "192.0.2.10"  # CDN IP
    priority   = 0
    monitor_id = norainadns_monitor.origin.id
  }

  values {
    value    = "192.0.2.20"  # Origin IP (fallback)
    priority = 1
  }
}
```

## Argument Reference

- `name` - (Required) The name of this monitor.
- `description` - (Optional) A description for this monitor.
- `protocol` - (Required) The protocol to use for health checks. Valid values: `HTTP`, `HTTPS`.
- `target_host` - (Required) The hostname to check. This is used for the Host header and SNI.
- `target_port` - (Optional) The port to check. Defaults to `443` for HTTPS, `80` for HTTP.
- `target_path` - (Optional) The path to check. Defaults to `/`.
- `check_interval_seconds` - (Optional) How often to check the endpoint, in seconds. Valid values: `60`, `300`, `600`, `900`, `1800`, `3600`. Defaults to `300`.
- `failure_threshold` - (Optional) Number of consecutive failures before marking unhealthy. Range: 1-10. Defaults to `3`.
- `timeout_seconds` - (Optional) Timeout for each health check, in seconds. Range: 5-60. Defaults to `10`.
- `custom_host_header` - (Optional) Custom Host header to send. Overrides `target_host` for the HTTP request.
- `custom_sni` - (Optional) Custom SNI (Server Name Indication) for TLS. Overrides `target_host` for TLS handshake.
- `force_ip_address` - (Optional) IP address to connect to directly, bypassing DNS. Use this to check origin servers behind CDN/WAF.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The unique identifier for this monitor.
- `is_active` - Whether this monitor is active.
- `current_status` - The current health status: `healthy`, `unhealthy`, `pending`, or `unknown`.

## Import

Monitors can be imported using the monitor ID:

```bash
terraform import norainadns_monitor.origin 3f2504e0-4f89-41d3-9a0c-0305e82c3301
```

## Notes

### CDN/WAF Bypass Pattern

When your application uses a CDN (like Cloudflare, CloudFront, or Fastly), health checks to the CDN endpoint may not accurately reflect origin server health. The CDN might:

- Cache responses (including error pages)
- Return its own health check responses
- Hide origin server failures

Use `force_ip_address` to check the origin directly:

```hcl
resource "norainadns_monitor" "origin" {
  name             = "Origin behind CDN"
  protocol         = "HTTPS"
  target_host      = "www.example.com"    # Used for Host header
  force_ip_address = "192.0.2.100"        # Origin's actual IP
  target_path      = "/health"
}
```

### Dependency Management

~> **Important:** You cannot delete a monitor that is referenced by a DNS record. First remove or update the record to remove the `monitor_id` reference, then delete the monitor.

Terraform handles creation order automatically:

```hcl
resource "norainadns_monitor" "web" {
  name        = "Web Monitor"
  # ...
}

resource "norainadns_record" "app" {
  # ...
  values {
    value      = "192.0.2.1"
    monitor_id = norainadns_monitor.web.id  # Creates implicit dependency
  }
}
```

When destroying resources, you may need to:
1. Remove `monitor_id` from the record
2. Apply to update the record
3. Then destroy the monitor

Or use `terraform destroy` which handles the ordering automatically.

### Health Check Status

The `current_status` attribute reflects the real-time health of the monitored endpoint:

| Status | Description |
|--------|-------------|
| `healthy` | Endpoint is responding correctly |
| `unhealthy` | Endpoint is failing health checks |
| `pending` | Monitor was just created, waiting for first check |
| `unknown` | Unable to determine status |

### Check Intervals

Choose an appropriate check interval based on your needs:

| Interval | Use Case |
|----------|----------|
| 60s | Critical services requiring fast failover |
| 300s | Standard production services |
| 600s | Less critical services |
| 1800s-3600s | Cost optimization for non-critical endpoints |
