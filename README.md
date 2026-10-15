# NorainaDNS Terraform Provider

The NorainaDNS Terraform Provider allows you to manage DNS zones, records, and health monitors in [NorainaDNS](https://noraina.cloud).

## Features

- **Multi-cloud DNS**: Records are automatically synchronized to both AWS Route53 and GCP Cloud DNS
- **Health-based failover**: Use monitors to automatically switch DNS to backup IPs when primary fails
- **CDN/WAF bypass**: Monitor origin servers directly, bypassing CDN for accurate health checks

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21 (to build the provider plugin)

## Installation

### From Terraform Registry

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

### Building from Source

```bash
git clone https://github.com/norainacloud/terraform-provider-norainadns.git
cd terraform-provider-norainadns
make install
```

## Quick Start

1. **Set up authentication:**

```bash
export NORAINADNS_ENDPOINT="https://dash.norainadns.com"
export NORAINADNS_API_KEY="nordns_your_api_key_here"
```

2. **Create your Terraform configuration:**

```hcl
terraform {
  required_providers {
    norainadns = {
      source  = "norainacloud/norainadns"
      version = "~> 0.1"
    }
  }
}

provider "norainadns" {}

resource "norainadns_zone" "example" {
  domain_name = "example.com"
  description = "My domain"
}

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

3. **Apply the configuration:**

```bash
terraform init
terraform apply
```

## Documentation

- **Provider documentation**: [docs/](docs/)
- **User manual**: [manual/](manual/)
- **Examples**: [examples/](examples/)

### Resources

- [norainadns_zone](docs/resources/zone.md) - Manage DNS zones
- [norainadns_record](docs/resources/record.md) - Manage DNS records
- [norainadns_monitor](docs/resources/monitor.md) - Manage health monitors

### Data Sources

- [norainadns_zone](docs/data-sources/zone.md) - Fetch zone information
- [norainadns_zones](docs/data-sources/zones.md) - List all zones
- [norainadns_record](docs/data-sources/record.md) - Fetch record information
- [norainadns_records](docs/data-sources/records.md) - List zone records
- [norainadns_monitor](docs/data-sources/monitor.md) - Fetch monitor information
- [norainadns_monitors](docs/data-sources/monitors.md) - List all monitors

## Development

### Building

```bash
make build
```

### Testing

```bash
# Unit tests
make test

# Acceptance tests (requires API credentials)
export NORAINADNS_ENDPOINT="https://dash.norainadns.com"
export NORAINADNS_API_KEY="your_api_key"
make testacc
```

### Installing Locally

```bash
make install
```

## License

This project is licensed under the MPL-2.0 License - see the LICENSE file for details.
