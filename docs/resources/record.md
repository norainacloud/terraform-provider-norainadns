---
page_title: "norainadns_record Resource - NorainaDNS"
subcategory: ""
description: |-
  Manages a DNS record in NorainaDNS.
---

# norainadns_record (Resource)

Manages a DNS record in NorainaDNS.

Records are automatically synchronized to both AWS Route53 and GCP Cloud DNS. Supports simple, failover, weighted, and latency-based routing policies.

## Example Usage

### Simple A Record

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

### Root Domain Record

```hcl
resource "norainadns_record" "root" {
  zone_id = norainadns_zone.example.id
  name    = "@"  # @ represents the root domain
  type    = "A"
  ttl     = 300

  values {
    value = "192.0.2.1"
  }
}
```

### CNAME Record

```hcl
resource "norainadns_record" "blog" {
  zone_id = norainadns_zone.example.id
  name    = "blog"
  type    = "CNAME"
  ttl     = 300

  values {
    value = "myblog.wordpress.com"
  }
}
```

### MX Records

```hcl
resource "norainadns_record" "mail" {
  zone_id = norainadns_zone.example.id
  name    = "@"
  type    = "MX"
  ttl     = 3600

  values {
    value    = "10 mail1.example.com."
    priority = 0
  }

  values {
    value    = "20 mail2.example.com."
    priority = 1
  }
}
```

### TXT Record (SPF)

```hcl
resource "norainadns_record" "spf" {
  zone_id = norainadns_zone.example.id
  name    = "@"
  type    = "TXT"
  ttl     = 3600

  values {
    value = "v=spf1 include:_spf.google.com ~all"
  }
}
```

### Failover Record with Health Monitoring

```hcl
# First create a monitor for the origin server
resource "norainadns_monitor" "origin" {
  name             = "Origin Health Check"
  protocol         = "HTTPS"
  target_host      = "cdn.example.com"
  target_path      = "/health"
  force_ip_address = "192.0.2.20"  # Check origin directly
}

# Then create the failover record
resource "norainadns_record" "app" {
  zone_id        = norainadns_zone.example.id
  name           = "app"
  type           = "A"
  ttl            = 60
  is_dynamic     = true
  routing_policy = "failover"

  # Primary (CDN) - served when origin is healthy
  values {
    value      = "192.0.2.10"
    priority   = 0
    monitor_id = norainadns_monitor.origin.id
  }

  # Secondary (Origin) - served when primary is unhealthy
  values {
    value    = "192.0.2.20"
    priority = 1
  }
}
```

### Weighted Routing

```hcl
resource "norainadns_record" "weighted" {
  zone_id        = norainadns_zone.example.id
  name           = "api"
  type           = "A"
  ttl            = 60
  routing_policy = "weighted"

  values {
    value  = "192.0.2.1"
    weight = 70  # 70% of traffic
  }

  values {
    value  = "192.0.2.2"
    weight = 30  # 30% of traffic
  }
}
```

## Argument Reference

- `zone_id` - (Required, Forces new resource) The ID of the zone this record belongs to.
- `name` - (Required) The name of the record (subdomain). Use `@` for the root domain.
- `type` - (Required, Forces new resource) The record type. Supported types: `A`, `AAAA`, `CNAME`, `TXT`, `MX`, `NS`, `SRV`, `CAA`, `PTR`.
- `ttl` - (Optional) Time to live in seconds. Defaults to `300`. Valid range: 60-86400.
- `is_dynamic` - (Optional) Whether this is a dynamic record with health checking. Set to `true` for failover routing. Defaults to `false`.
- `routing_policy` - (Optional) The routing policy. Valid values: `simple`, `failover`, `weighted`, `latency`. Defaults to `simple`.
- `values` - (Required) One or more value blocks. See [Values](#values) below.

### Values

- `value` - (Required) The record value. For A records, this is an IP address. For CNAME records, this is a hostname.
- `priority` - (Optional) The priority of this value. Lower values have higher priority. For failover: `0` = primary, `1` = secondary. Defaults to `0`.
- `weight` - (Optional) The weight for weighted routing. Defaults to `0`.
- `monitor_id` - (Optional) The ID of the monitor for health-checked failover. Required for the primary value when using `failover` routing.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The unique identifier for this record.
- `sync_status` - The synchronization status (`synced`, `pending`, `error`).
- `fqdn` - The fully qualified domain name for this record.

## Import

Records can be imported using the format `zone_id/record_id`:

```bash
terraform import norainadns_record.www 550e8400-e29b-41d4-a716-446655440000/7c9e6679-7425-40de-944b-e07fc1f90ae7
```

## Notes

### Failover Routing

For automatic failover:

1. Create a monitor that checks your origin server's health
2. Set `is_dynamic = true` and `routing_policy = "failover"`
3. Define two values:
   - Primary (priority 0) with a `monitor_id`
   - Secondary (priority 1) as the fallback

When the monitored endpoint becomes unhealthy, DNS automatically switches to the secondary IP.

~> **Important:** The monitor must exist before the record. Terraform handles this automatically when you reference `norainadns_monitor.xxx.id`.

### CDN/WAF Bypass

A common pattern is to monitor the origin server directly while serving traffic through a CDN:

```hcl
resource "norainadns_monitor" "origin" {
  name             = "Origin"
  protocol         = "HTTPS"
  target_host      = "cdn.example.com"  # Public hostname
  force_ip_address = "192.0.2.20"       # Origin IP (bypasses CDN)
  custom_host_header = "origin.example.com"
}
```

This allows accurate health checking of the origin even when the CDN is caching responses.

### Record Types

| Type | Description | Example Value |
|------|-------------|---------------|
| A | IPv4 address | `192.0.2.1` |
| AAAA | IPv6 address | `2001:db8::1` |
| CNAME | Canonical name | `www.example.com` |
| TXT | Text record | `v=spf1 include:_spf.google.com ~all` |
| MX | Mail exchange | `10 mail.example.com.` |
| NS | Nameserver | `ns1.example.com.` |
| SRV | Service record | `10 5 5060 sipserver.example.com.` |
| CAA | Certificate Authority Authorization | `0 issue "letsencrypt.org"` |
| PTR | Pointer record | `host.example.com.` |
