# NorainaDNS Terraform Provider - User Manual

This manual provides comprehensive documentation for using the NorainaDNS Terraform Provider to manage DNS zones, records, and health monitors.

## Table of Contents

1. [Introduction](#introduction)
2. [Installation](#installation)
3. [Authentication](#authentication)
4. [Quick Start](#quick-start)
5. [Resources](#resources)
6. [Data Sources](#data-sources)
7. [Failover Configuration](#failover-configuration)
8. [Best Practices](#best-practices)
9. [Troubleshooting](#troubleshooting)
10. [Examples](#examples)

---

## Introduction

### What is NorainaDNS?

NorainaDNS is a multi-tenant DNS management system that synchronizes DNS zones and records across both AWS Route53 and GCP Cloud DNS. It provides:

- **Multi-cloud redundancy**: DNS is automatically replicated to both AWS and GCP
- **Health-based failover**: Automatic DNS switching when endpoints become unhealthy
- **CDN/WAF bypass monitoring**: Monitor origin servers directly, bypassing CDN caching

### What is the Terraform Provider?

The NorainaDNS Terraform Provider allows you to manage your DNS infrastructure as code using HashiCorp Terraform. You can:

- Create and manage DNS zones
- Create and manage DNS records (A, AAAA, CNAME, TXT, MX, etc.)
- Configure health monitors for automatic failover
- Import existing resources into Terraform state

---

## Installation

### Requirements

- Terraform >= 1.0
- Go >= 1.21 (only for building from source)
- A NorainaDNS account with API access

### Installing from Terraform Registry

Add the provider to your Terraform configuration:

```hcl
terraform {
  required_providers {
    norainadns = {
      source  = "norainacloud/norainadns"
      version = "~> 0.1"
    }
  }
}
```

Then run:

```bash
terraform init
```

### Building from Source

```bash
git clone https://github.com/norainacloud/terraform-provider-norainadns.git
cd terraform-provider-norainadns
make install
```

---

## Authentication

The provider requires an API key for authentication. You can obtain an API key from the NorainaDNS web interface under **Settings > API Keys**.

### Option 1: Environment Variables (Recommended)

```bash
export NORAINADNS_ENDPOINT="https://dns.noraina.cloud"
export NORAINADNS_API_KEY="nordns_your_api_key_here"
```

Then configure the provider without explicit credentials:

```hcl
provider "norainadns" {}
```

### Option 2: Provider Configuration

```hcl
provider "norainadns" {
  endpoint = "https://dns.noraina.cloud"
  api_key  = var.norainadns_api_key  # Use a variable, never hardcode!
}
```

### Security Best Practices

- **Never commit API keys to version control**
- Use environment variables or a secrets manager (HashiCorp Vault, AWS Secrets Manager, etc.)
- If using Terraform Cloud/Enterprise, store the API key as a sensitive variable
- Rotate API keys periodically

---

## Quick Start

### Step 1: Configure the Provider

Create a `main.tf` file:

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
  # Uses NORAINADNS_ENDPOINT and NORAINADNS_API_KEY environment variables
}
```

### Step 2: Create a Zone

```hcl
resource "norainadns_zone" "example" {
  domain_name = "example.com"
  description = "My production domain"
}

output "nameservers" {
  value = norainadns_zone.example.nameservers
}
```

### Step 3: Add Records

```hcl
resource "norainadns_record" "www" {
  zone_id = norainadns_zone.example.id
  name    = "www"
  type    = "A"
  ttl     = 300

  values {
    value = "192.0.2.1"
  }
}
```

### Step 4: Apply Configuration

```bash
terraform init
terraform plan
terraform apply
```

### Step 5: Configure Nameservers

After creating the zone, configure the nameservers shown in the output at your domain registrar.

---

## Resources

### norainadns_zone

Manages a DNS zone. Zones are automatically synchronized to both AWS Route53 and GCP Cloud DNS.

```hcl
resource "norainadns_zone" "example" {
  domain_name = "example.com"
  description = "Production domain"
}
```

#### Arguments

| Argument | Type | Required | Description |
|----------|------|----------|-------------|
| domain_name | string | Yes | The domain name (e.g., "example.com") |
| description | string | No | A description for this zone |

#### Attributes

| Attribute | Description |
|-----------|-------------|
| id | Unique identifier |
| is_active | Whether the zone is active |
| nameservers | List of nameservers to configure at registrar |
| sync_status | Synchronization status (synced, pending, error) |

---

### norainadns_record

Manages a DNS record within a zone.

```hcl
resource "norainadns_record" "www" {
  zone_id = norainadns_zone.example.id
  name    = "www"
  type    = "A"
  ttl     = 300

  values {
    value = "192.0.2.1"
  }
}
```

#### Arguments

| Argument | Type | Required | Description |
|----------|------|----------|-------------|
| zone_id | string | Yes | The zone ID |
| name | string | Yes | Record name (use "@" for root) |
| type | string | Yes | Record type (A, AAAA, CNAME, TXT, MX, NS, SRV, CAA, PTR) |
| ttl | number | No | Time to live (default: 300, range: 60-86400) |
| is_dynamic | bool | No | Enable health checking (default: false) |
| routing_policy | string | No | Routing policy (simple, failover, weighted, latency) |
| values | block | Yes | One or more value blocks |

#### Values Block

| Argument | Type | Required | Description |
|----------|------|----------|-------------|
| value | string | Yes | The record value |
| priority | number | No | Priority (0 = primary, 1 = secondary) |
| weight | number | No | Weight for weighted routing |
| monitor_id | string | No | Monitor ID for health checking |

---

### norainadns_monitor

Manages a health monitor for endpoint health checking.

```hcl
resource "norainadns_monitor" "origin" {
  name        = "Origin Server"
  protocol    = "HTTPS"
  target_host = "www.example.com"
  target_path = "/health"
}
```

#### Arguments

| Argument | Type | Required | Description |
|----------|------|----------|-------------|
| name | string | Yes | Monitor name |
| description | string | No | Description |
| protocol | string | Yes | HTTP or HTTPS |
| target_host | string | Yes | Hostname to check |
| target_port | number | No | Port (default: 443/80) |
| target_path | string | No | Path to check (default: /) |
| check_interval_seconds | number | No | Check frequency (default: 300) |
| failure_threshold | number | No | Failures before unhealthy (default: 3) |
| timeout_seconds | number | No | Check timeout (default: 10) |
| custom_host_header | string | No | Custom Host header |
| custom_sni | string | No | Custom SNI |
| force_ip_address | string | No | IP to connect to directly (bypasses DNS) |

#### Attributes

| Attribute | Description |
|-----------|-------------|
| id | Unique identifier |
| is_active | Whether the monitor is active |
| current_status | Health status (healthy, unhealthy, pending, unknown) |

---

## Data Sources

### norainadns_zone / norainadns_zones

Fetch information about existing zones.

```hcl
# Single zone
data "norainadns_zone" "example" {
  id = "123"
}

# All zones
data "norainadns_zones" "all" {}
```

### norainadns_record / norainadns_records

Fetch information about existing records.

```hcl
# Single record
data "norainadns_record" "www" {
  zone_id = "1"
  id      = "42"
}

# All records in a zone
data "norainadns_records" "all" {
  zone_id = "1"
}
```

### norainadns_monitor / norainadns_monitors

Fetch information about existing monitors.

```hcl
# Single monitor
data "norainadns_monitor" "origin" {
  id = "456"
}

# All monitors
data "norainadns_monitors" "all" {}
```

---

## Failover Configuration

Failover allows automatic DNS switching when your primary endpoint becomes unhealthy.

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Normal Operation                         │
│                                                             │
│   User Request → DNS (Primary IP) → CDN → Origin Server    │
│                                                             │
│   Health Check (bypasses CDN) ────────→ Origin Server      │
│                                    └── Returns HEALTHY      │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                     Failover Active                          │
│                                                             │
│   User Request → DNS (Secondary IP) → Origin Server        │
│                                                             │
│   Health Check (bypasses CDN) ────────→ Origin Server      │
│                                    └── Returns UNHEALTHY    │
└─────────────────────────────────────────────────────────────┘
```

### Configuration Steps

1. **Create a health monitor** that checks your origin server:

```hcl
resource "norainadns_monitor" "origin" {
  name             = "Origin Health"
  protocol         = "HTTPS"
  target_host      = "www.example.com"  # For Host header
  target_path      = "/health"
  force_ip_address = "192.0.2.20"       # Origin IP (bypasses CDN)

  check_interval_seconds = 60
  failure_threshold      = 2
  timeout_seconds        = 10
}
```

2. **Create a failover record** with primary and secondary values:

```hcl
resource "norainadns_record" "www" {
  zone_id        = norainadns_zone.example.id
  name           = "www"
  type           = "A"
  ttl            = 60              # Low TTL for fast failover
  is_dynamic     = true
  routing_policy = "failover"

  # Primary: CDN IP (served when origin is healthy)
  values {
    value      = "192.0.2.10"      # CDN IP
    priority   = 0
    monitor_id = norainadns_monitor.origin.id
  }

  # Secondary: Origin IP (served when origin is unhealthy)
  values {
    value    = "192.0.2.20"        # Origin IP
    priority = 1
  }
}
```

### CDN/WAF Bypass

When monitoring origins behind a CDN or WAF, use `force_ip_address` to check the origin directly:

```hcl
resource "norainadns_monitor" "origin" {
  name             = "Origin"
  protocol         = "HTTPS"
  target_host      = "www.example.com"    # Used for Host header
  force_ip_address = "192.0.2.100"        # Origin's actual IP
  target_path      = "/health"
}
```

This ensures health checks reach your origin server, not the CDN's cached responses.

---

## Best Practices

### 1. Use Low TTLs for Failover Records

```hcl
resource "norainadns_record" "app" {
  ttl = 60  # 1 minute for fast failover
  # ...
}
```

### 2. Create Monitors Before Records

Terraform handles dependencies automatically when you reference monitor IDs:

```hcl
resource "norainadns_monitor" "origin" {
  # ...
}

resource "norainadns_record" "app" {
  # ...
  values {
    monitor_id = norainadns_monitor.origin.id  # Creates dependency
  }
}
```

### 3. Use Variables for IPs

```hcl
variable "cdn_ip" {
  type = string
}

variable "origin_ip" {
  type = string
}

resource "norainadns_record" "www" {
  # ...
  values {
    value = var.cdn_ip
  }
  values {
    value = var.origin_ip
  }
}
```

### 4. Use Workspaces for Environments

```bash
terraform workspace new production
terraform workspace new staging
terraform workspace select production
```

### 5. Store State Remotely

```hcl
terraform {
  backend "s3" {
    bucket = "my-terraform-state"
    key    = "norainadns/terraform.tfstate"
    region = "eu-west-1"
  }
}
```

---

## Troubleshooting

### Error: Monitor In Use

```
Error: Could not delete monitor: Monitor is currently in use by DNS records
```

**Solution**: Remove or update records that reference this monitor before deleting it.

### Error: Zone Already Exists

```
Error: Zone example.com already exists
```

**Solution**: Import the existing zone into Terraform:

```bash
terraform import norainadns_zone.example 550e8400-e29b-41d4-a716-446655440000
```

### Error: Invalid API Key

```
Error: Authentication failed: Invalid API key
```

**Solution**: Check that your API key is correct and hasn't expired.

### Slow Failover

**Problem**: DNS takes too long to switch during failover.

**Solutions**:
- Lower the TTL on failover records (minimum 60 seconds)
- Reduce `failure_threshold` on monitors (e.g., 2 instead of 3)
- Reduce `check_interval_seconds` (e.g., 60 instead of 300)

### Health Check Not Reaching Origin

**Problem**: Health checks pass even when origin is down.

**Solution**: Use `force_ip_address` to bypass CDN:

```hcl
resource "norainadns_monitor" "origin" {
  force_ip_address = "your-origin-ip"
}
```

---

## Examples

See the `examples/` directory for complete, working examples:

- **basic/**: Simple zone and record configuration
- **failover/**: Health monitoring with automatic failover
- **multi-environment/**: Managing dev, staging, and production

---

## Support

For issues with the Terraform provider, please open an issue on the GitHub repository.

For NorainaDNS service questions, contact support@noraina.cloud.
