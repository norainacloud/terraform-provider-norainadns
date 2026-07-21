---
page_title: "norainadns_records Data Source - NorainaDNS"
subcategory: ""
description: |-
  Fetches all DNS records in a zone.
---

# norainadns_records (Data Source)

Fetches all DNS records in a specific zone.

Use this data source to list all records in a zone or to find records by name or type.

## Example Usage

### List All Records in a Zone

```hcl
data "norainadns_records" "all" {
  zone_id = norainadns_zone.example.id
}

output "all_records" {
  value = [for r in data.norainadns_records.all.records : {
    name = r.name
    type = r.type
    fqdn = r.fqdn
  }]
}
```

### Find Records by Type

```hcl
data "norainadns_records" "all" {
  zone_id = "1"
}

locals {
  a_records = [for r in data.norainadns_records.all.records : r if r.type == "A"]
  mx_records = [for r in data.norainadns_records.all.records : r if r.type == "MX"]
}

output "a_record_count" {
  value = length(local.a_records)
}
```

### Find Record by Name

```hcl
data "norainadns_records" "all" {
  zone_id = "1"
}

locals {
  www_record = [for r in data.norainadns_records.all.records : r if r.name == "www"][0]
}

output "www_ips" {
  value = local.www_record.values[*].value
}
```

## Argument Reference

- `zone_id` - (Required) The ID of the zone to fetch records from.

## Attribute Reference

The following attribute is exported:

- `records` - A list of records. Each record has:
  - `id` - The unique identifier.
  - `zone_id` - The zone ID.
  - `name` - The record name (subdomain).
  - `type` - The record type.
  - `ttl` - Time to live in seconds.
  - `is_dynamic` - Whether health checking is enabled.
  - `routing_policy` - The routing policy.
  - `sync_status` - The synchronization status.
  - `fqdn` - The fully qualified domain name.
  - `values` - List of values with `value`, `priority`, `weight`, and `monitor_id`.
