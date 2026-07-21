---
page_title: "norainadns_monitor Data Source - NorainaDNS"
subcategory: ""
description: |-
  Fetches information about a health monitor.
---

# norainadns_monitor (Data Source)

Fetches information about an existing health monitor in NorainaDNS.

Use this data source to reference a monitor created outside of Terraform or to check the current health status.

## Example Usage

### Lookup by ID

```hcl
data "norainadns_monitor" "origin" {
  id = "456"
}

output "monitor_status" {
  value = data.norainadns_monitor.origin.current_status
}
```

### Check Health Status

```hcl
data "norainadns_monitor" "web" {
  id = norainadns_monitor.web.id
}

output "is_healthy" {
  value = data.norainadns_monitor.web.current_status == "healthy"
}
```

### Use Existing Monitor in Record

```hcl
data "norainadns_monitor" "existing" {
  id = "123"
}

resource "norainadns_record" "app" {
  zone_id        = norainadns_zone.example.id
  name           = "app"
  type           = "A"
  ttl            = 60
  is_dynamic     = true
  routing_policy = "failover"

  values {
    value      = "192.0.2.10"
    priority   = 0
    monitor_id = data.norainadns_monitor.existing.id
  }

  values {
    value    = "192.0.2.20"
    priority = 1
  }
}
```

## Argument Reference

- `id` - (Required) The ID of the monitor to fetch.

## Attribute Reference

In addition to the argument above, the following attributes are exported:

- `name` - The name of this monitor.
- `description` - The description of this monitor.
- `protocol` - The protocol (HTTP, HTTPS).
- `target_host` - The hostname being monitored.
- `target_port` - The port being checked.
- `target_path` - The path being checked.
- `check_interval_seconds` - How often the endpoint is checked.
- `failure_threshold` - Number of failures before unhealthy.
- `timeout_seconds` - Timeout for each check.
- `custom_host_header` - Custom Host header (if set).
- `custom_sni` - Custom SNI (if set).
- `force_ip_address` - Forced IP address (if set).
- `is_active` - Whether this monitor is active.
- `current_status` - The current health status (`healthy`, `unhealthy`, `pending`, `unknown`).
