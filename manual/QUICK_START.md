# NorainaDNS Terraform Provider - Quick Start Guide

Get up and running with NorainaDNS in 5 minutes.

## Prerequisites

- Terraform >= 1.0 installed
- NorainaDNS API key (from Settings > API Keys in the web interface)

## Step 1: Set Up Authentication

```bash
export NORAINADNS_ENDPOINT="https://dash.norainadns.com"
export NORAINADNS_API_KEY="nordns_your_api_key_here"
```

## Step 2: Create Configuration

Create a file named `main.tf`:

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

# Create a zone
resource "norainadns_zone" "mysite" {
  domain_name = "mysite.com"
  description = "My website"
}

# Create www record
resource "norainadns_record" "www" {
  zone_id = norainadns_zone.mysite.id
  name    = "www"
  type    = "A"
  ttl     = 300

  values {
    value = "YOUR_SERVER_IP"
  }
}

# Output nameservers
output "nameservers" {
  value = norainadns_zone.mysite.nameservers
}
```

## Step 3: Initialize and Apply

```bash
# Initialize Terraform
terraform init

# Preview changes
terraform plan

# Apply changes
terraform apply
```

## Step 4: Configure Nameservers

After `terraform apply`, you'll see the nameservers in the output. Configure these at your domain registrar.

## Next Steps

- Read the [full documentation](README.md)
- See [failover example](../examples/failover/) for health-based DNS switching
- Explore [data sources](README.md#data-sources) to reference existing resources

## Common Tasks

### Add a CNAME record

```hcl
resource "norainadns_record" "blog" {
  zone_id = norainadns_zone.mysite.id
  name    = "blog"
  type    = "CNAME"
  ttl     = 300

  values {
    value = "blog.wordpress.com"
  }
}
```

### Add MX records for email

```hcl
resource "norainadns_record" "mail" {
  zone_id = norainadns_zone.mysite.id
  name    = "@"
  type    = "MX"
  ttl     = 3600

  values {
    value    = "10 mail.example.com."
    priority = 0
  }

  values {
    value    = "20 mail2.example.com."
    priority = 1
  }
}
```

### Configure failover

```hcl
# Monitor your origin server
resource "norainadns_monitor" "origin" {
  name             = "Origin"
  protocol         = "HTTPS"
  target_host      = "mysite.com"
  target_path      = "/health"
  force_ip_address = "YOUR_ORIGIN_IP"
}

# Failover record
resource "norainadns_record" "app" {
  zone_id        = norainadns_zone.mysite.id
  name           = "app"
  type           = "A"
  ttl            = 60
  is_dynamic     = true
  routing_policy = "failover"

  values {
    value      = "CDN_IP"
    priority   = 0
    monitor_id = norainadns_monitor.origin.id
  }

  values {
    value    = "YOUR_ORIGIN_IP"
    priority = 1
  }
}
```

## Need Help?

- [Full Documentation](README.md)
- [Examples](../examples/)
- [Troubleshooting](README.md#troubleshooting)
