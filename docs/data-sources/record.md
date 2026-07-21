---
page_title: "norainadns_record Data Source - NorainaDNS"
subcategory: ""
description: |-
  Fetches information about a DNS record.
---

# norainadns_record (Data Source)

Fetches information about an existing DNS record in NorainaDNS.

Use this data source to reference a record created outside of Terraform or to read the current state of a record.

## Example Usage

### Lookup by ID

```hcl
data "norainadns_record" "www" {
  zone_id = "1"
  id      = "42"
}

output "record_fqdn" {
  value = data.norainadns_record.www.fqdn
}

output "record_values" {
  value = data.norainadns_record.www.values
}
```

### Check Record Status

```hcl
data "norainadns_record" "app" {
  zone_id = norainadns_zone.example.id
  id      = norainadns_record.app.id
}

output "sync_status" {
  value = data.norainadns_record.app.sync_status
}
```

## Argument Reference

- `zone_id` - (Required) The ID of the zone containing this record.
- `id` - (Required) The ID of the record to fetch.

## Attribute Reference

In addition to the arguments above, the following attributes are exported:

- `name` - The name of the record (subdomain).
- `type` - The record type (A, AAAA, CNAME, etc.).
- `ttl` - Time to live in seconds.
- `is_dynamic` - Whether this is a dynamic record with health checking.
- `routing_policy` - The routing policy (simple, failover, weighted, latency).
- `sync_status` - The synchronization status.
- `fqdn` - The fully qualified domain name.
- `values` - List of record values. Each value has:
  - `value` - The record value.
  - `priority` - The priority of this value.
  - `weight` - The weight for weighted routing.
  - `monitor_id` - The ID of the associated monitor (if any).
