---
page_title: "NorainaDNS Provider"
subcategory: ""
description: |-
  The NorainaDNS provider allows you to manage DNS zones, records, and health monitors.
---

# NorainaDNS Provider

The NorainaDNS provider allows you to manage DNS zones, records, and health monitors in [NorainaDNS](https://noraina.cloud).

NorainaDNS is a multi-tenant DNS management system that synchronizes zones and records across multiple DNS servers, providing automatic failover support via health monitoring.

## Features

- **Multi-cloud DNS**: Records are automatically synchronized to both AWS Route53 and GCP Cloud DNS
- **Health-based failover**: Use monitors to automatically switch DNS to backup IPs when primary fails
- **CDN/WAF bypass**: Monitor origin servers directly, bypassing CDN/WAF for accurate health checks

## Example Usage

```hcl
terraform {
  required_providers {
    norainadns = {
      source  = "norainacloud/norainadns"
      version = "~> 0.1"
    }
  }
}

provider "norainadns" {
  endpoint = "https://dns.noraina.cloud"
  api_key  = var.norainadns_api_key
}

# Create a zone
resource "norainadns_zone" "example" {
  domain_name = "example.com"
  description = "Example domain"
}

# Create a simple A record
resource "norainadns_record" "www" {
  zone_id = norainadns_zone.example.id
  name    = "www"
  type    = "A"
  ttl     = 300

  values {
    value = "192.0.2.1"
  }
}

# Create a monitor for failover
resource "norainadns_monitor" "origin" {
  name             = "Origin Server"
  protocol         = "HTTPS"
  target_host      = "origin.example.com"
  target_path      = "/health"
  force_ip_address = "192.0.2.20"  # Monitor origin directly
}

# Create a failover record
resource "norainadns_record" "app" {
  zone_id        = norainadns_zone.example.id
  name           = "app"
  type           = "A"
  ttl            = 60
  is_dynamic     = true
  routing_policy = "failover"

  # Primary (CDN) - monitored
  values {
    value      = "192.0.2.10"
    priority   = 0
    monitor_id = norainadns_monitor.origin.id
  }

  # Secondary (Origin) - fallback
  values {
    value    = "192.0.2.20"
    priority = 1
  }
}

# Output the nameservers
output "nameservers" {
  value = norainadns_zone.example.nameservers
}
```

## Authentication

The provider requires an API key for authentication. You can generate an API key from the NorainaDNS web interface under **Settings > API Keys**.

### Configuration Options

The API key can be provided in two ways:

#### Via Provider Configuration

```hcl
provider "norainadns" {
  endpoint = "https://dns.noraina.cloud"
  api_key  = "nordns_your_api_key_here"
}
```

#### Via Environment Variables

```bash
export NORAINADNS_ENDPOINT="https://dns.noraina.cloud"
export NORAINADNS_API_KEY="nordns_your_api_key_here"
```

-> **Note:** We recommend using environment variables or a secrets manager for the API key to avoid storing secrets in Terraform configuration files.

## Argument Reference

- `endpoint` - (Required) The NorainaDNS API endpoint URL. Can also be set via `NORAINADNS_ENDPOINT` environment variable.
- `api_key` - (Required) The API key for authenticating with NorainaDNS. Can also be set via `NORAINADNS_API_KEY` environment variable.

## Resource Dependencies

When using failover routing, be aware of these dependency rules:

1. **Monitors before Records**: A monitor must exist before you can create a record that references it. Use `depends_on` or Terraform's implicit dependency (by referencing the monitor's ID) to ensure correct ordering.

2. **Records before Monitor Deletion**: You cannot delete a monitor that is referenced by a DNS record. First update the record to remove the monitor reference, then delete the monitor.

### Example with Correct Dependencies

```hcl
# Monitor is created first
resource "norainadns_monitor" "primary" {
  name        = "Primary Server"
  protocol    = "HTTPS"
  target_host = "primary.example.com"
}

# Record references monitor - Terraform handles the dependency automatically
resource "norainadns_record" "app" {
  zone_id        = norainadns_zone.example.id
  name           = "app"
  type           = "A"
  routing_policy = "failover"

  values {
    value      = "192.0.2.10"
    priority   = 0
    monitor_id = norainadns_monitor.primary.id  # Creates implicit dependency
  }

  values {
    value    = "192.0.2.20"
    priority = 1
  }
}
```
