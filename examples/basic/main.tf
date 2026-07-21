# =============================================================================
# Basic NorainaDNS Example
# =============================================================================
# This example shows how to create a zone with simple DNS records.

terraform {
  required_providers {
    norainadns = {
      source  = "norainacloud/norainadns"
      version = "~> 0.1"
    }
  }
}

# Configure the provider
provider "norainadns" {
  endpoint = var.norainadns_endpoint
  api_key  = var.norainadns_api_key
}

# Create a DNS zone
resource "norainadns_zone" "example" {
  domain_name = "example.com"
  description = "Example production domain"
}

# Create an A record for www
resource "norainadns_record" "www" {
  zone_id = norainadns_zone.example.id
  name    = "www"
  type    = "A"
  ttl     = 300

  values {
    value = "192.0.2.1"
  }
}

# Create an A record for the root domain
resource "norainadns_record" "root" {
  zone_id = norainadns_zone.example.id
  name    = "@"
  type    = "A"
  ttl     = 300

  values {
    value = "192.0.2.1"
  }
}

# Create a CNAME for the blog
resource "norainadns_record" "blog" {
  zone_id = norainadns_zone.example.id
  name    = "blog"
  type    = "CNAME"
  ttl     = 300

  values {
    value = "blog.wordpress.com"
  }
}

# Create MX records for email
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

# Create SPF record
resource "norainadns_record" "spf" {
  zone_id = norainadns_zone.example.id
  name    = "@"
  type    = "TXT"
  ttl     = 3600

  values {
    value = "v=spf1 include:_spf.google.com ~all"
  }
}

# Output the nameservers
output "nameservers" {
  description = "Configure these nameservers at your domain registrar"
  value       = norainadns_zone.example.nameservers
}

output "www_fqdn" {
  description = "The FQDN for the www record"
  value       = norainadns_record.www.fqdn
}
