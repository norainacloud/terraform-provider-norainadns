---
page_title: "norainadns_zones Data Source - NorainaDNS"
subcategory: ""
description: |-
  Fetches all DNS zones.
---

# norainadns_zones (Data Source)

Fetches all DNS zones in your NorainaDNS account.

Use this data source to list all available zones or to find a zone by domain name.

## Example Usage

### List All Zones

```hcl
data "norainadns_zones" "all" {}

output "all_zones" {
  value = [for z in data.norainadns_zones.all.zones : {
    id     = z.id
    domain = z.domain_name
  }]
}
```

### Find Zone by Domain Name

```hcl
data "norainadns_zones" "all" {}

locals {
  example_zone = [for z in data.norainadns_zones.all.zones : z if z.domain_name == "example.com"][0]
}

resource "norainadns_record" "www" {
  zone_id = local.example_zone.id
  name    = "www"
  type    = "A"
  ttl     = 300

  values {
    value = "192.0.2.1"
  }
}
```

## Argument Reference

This data source has no required arguments.

## Attribute Reference

The following attribute is exported:

- `zones` - A list of zones. Each zone has the following attributes:
  - `id` - The unique identifier for this zone.
  - `domain_name` - The domain name of this zone.
  - `description` - The description of this zone.
  - `is_active` - Whether this zone is active.
  - `sync_status` - The synchronization status.
  - `nameservers` - List of nameservers.
