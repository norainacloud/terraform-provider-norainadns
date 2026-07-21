---
page_title: "norainadns_zone Resource - NorainaDNS"
subcategory: ""
description: |-
  Manages a DNS zone in NorainaDNS.
---

# norainadns_zone (Resource)

Manages a DNS zone in NorainaDNS.

When a zone is created, it is automatically synchronized to both AWS Route53 and GCP Cloud DNS. The provider returns branded nameservers that you should configure at your domain registrar.

## Example Usage

### Basic Zone

```hcl
resource "norainadns_zone" "example" {
  domain_name = "example.com"
  description = "Production domain"
}

output "nameservers" {
  description = "Configure these at your registrar"
  value       = norainadns_zone.example.nameservers
}
```

### Zone with Records

```hcl
resource "norainadns_zone" "myapp" {
  domain_name = "myapp.io"
  description = "Application domain"
}

resource "norainadns_record" "www" {
  zone_id = norainadns_zone.myapp.id
  name    = "www"
  type    = "A"
  ttl     = 300

  values {
    value = "192.0.2.1"
  }
}

resource "norainadns_record" "root" {
  zone_id = norainadns_zone.myapp.id
  name    = "@"
  type    = "A"
  ttl     = 300

  values {
    value = "192.0.2.1"
  }
}
```

## Argument Reference

- `domain_name` - (Required, Forces new resource) The domain name for this zone. Must be a valid domain name (e.g., `example.com`).
- `description` - (Optional) A description for this zone.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The unique identifier for this zone.
- `is_active` - Whether this zone is active.
- `sync_status` - The synchronization status of this zone (`synced`, `pending`, `error`).
- `nameservers` - List of nameservers for this zone. Configure these at your domain registrar.

## Import

Zones can be imported using the zone ID:

```bash
terraform import norainadns_zone.example 550e8400-e29b-41d4-a716-446655440000
```

## Notes

### Multi-Cloud Synchronization

NorainaDNS automatically synchronizes zones to both AWS Route53 and GCP Cloud DNS. This provides:

- **Redundancy**: If one cloud provider has issues, DNS can still be served
- **Performance**: DNS queries are served from the geographically closest provider
- **Consistency**: Both providers always have the same records

### Nameserver Configuration

After creating a zone, you must configure the returned nameservers at your domain registrar. The nameservers are Noraina-branded and automatically route to both AWS and GCP.

Example nameservers:
- `ns1.noraina.cloud`
- `ns2.noraina.cloud`
- `ns3.noraina.cloud`
- `ns4.noraina.cloud`
