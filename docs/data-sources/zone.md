---
page_title: "norainadns_zone Data Source - NorainaDNS"
subcategory: ""
description: |-
  Fetches information about a DNS zone.
---

# norainadns_zone (Data Source)

Fetches information about an existing DNS zone in NorainaDNS.

Use this data source to reference an existing zone that was created outside of Terraform or in a different Terraform configuration.

## Example Usage

### Lookup by ID

```hcl
data "norainadns_zone" "example" {
  id = "123"
}

output "zone_nameservers" {
  value = data.norainadns_zone.example.nameservers
}
```

### Use in Record Creation

```hcl
data "norainadns_zone" "production" {
  id = "1"
}

resource "norainadns_record" "www" {
  zone_id = data.norainadns_zone.production.id
  name    = "www"
  type    = "A"
  ttl     = 300

  values {
    value = "192.0.2.1"
  }
}
```

## Argument Reference

- `id` - (Required) The ID of the zone to fetch.

## Attribute Reference

In addition to the argument above, the following attributes are exported:

- `domain_name` - The domain name of this zone.
- `description` - The description of this zone.
- `is_active` - Whether this zone is active.
- `sync_status` - The synchronization status of this zone.
- `nameservers` - List of nameservers for this zone.
