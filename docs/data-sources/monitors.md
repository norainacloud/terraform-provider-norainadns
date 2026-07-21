---
page_title: "norainadns_monitors Data Source - NorainaDNS"
subcategory: ""
description: |-
  Fetches all health monitors.
---

# norainadns_monitors (Data Source)

Fetches all health monitors in your NorainaDNS account.

Use this data source to list all monitors or find a monitor by name.

## Example Usage

### List All Monitors

```hcl
data "norainadns_monitors" "all" {}

output "all_monitors" {
  value = [for m in data.norainadns_monitors.all.monitors : {
    id     = m.id
    name   = m.name
    status = m.current_status
  }]
}
```

### Find Monitor by Name

```hcl
data "norainadns_monitors" "all" {}

locals {
  origin_monitor = [for m in data.norainadns_monitors.all.monitors : m if m.name == "Origin Server"][0]
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
    monitor_id = local.origin_monitor.id
  }

  values {
    value    = "192.0.2.20"
    priority = 1
  }
}
```

### Check All Monitor Statuses

```hcl
data "norainadns_monitors" "all" {}

output "unhealthy_monitors" {
  value = [for m in data.norainadns_monitors.all.monitors : m.name if m.current_status == "unhealthy"]
}
```

## Argument Reference

This data source has no required arguments.

## Attribute Reference

The following attribute is exported:

- `monitors` - A list of monitors. Each monitor has:
  - `id` - The unique identifier.
  - `name` - The monitor name.
  - `description` - The description.
  - `protocol` - The protocol (HTTP, HTTPS).
  - `target_host` - The hostname being monitored.
  - `target_port` - The port being checked.
  - `target_path` - The path being checked.
  - `check_interval_seconds` - Check frequency in seconds.
  - `failure_threshold` - Failures before unhealthy.
  - `timeout_seconds` - Timeout for each check.
  - `custom_host_header` - Custom Host header.
  - `custom_sni` - Custom SNI.
  - `force_ip_address` - Forced IP address.
  - `is_active` - Whether active.
  - `current_status` - Health status (`healthy`, `unhealthy`, `pending`, `unknown`).
