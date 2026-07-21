# =============================================================================
# Multi-Environment Example
# =============================================================================
# This example shows how to manage DNS for multiple environments (dev, staging,
# production) using a single Terraform configuration with workspaces or modules.

terraform {
  required_providers {
    norainadns = {
      source  = "norainacloud/norainadns"
      version = "~> 0.1"
    }
  }
}

provider "norainadns" {
  endpoint = var.norainadns_endpoint
  api_key  = var.norainadns_api_key
}

# =============================================================================
# Local Variables
# =============================================================================

locals {
  # Environment-specific configuration
  environments = {
    dev = {
      subdomain  = "dev"
      web_ip     = "10.0.1.10"
      api_ip     = "10.0.1.20"
      ttl        = 60
      monitoring = false
    }
    staging = {
      subdomain  = "staging"
      web_ip     = "10.0.2.10"
      api_ip     = "10.0.2.20"
      ttl        = 300
      monitoring = true
    }
    prod = {
      subdomain  = ""  # No subdomain for production
      web_ip     = "192.0.2.10"
      api_ip     = "192.0.2.20"
      ttl        = 300
      monitoring = true
    }
  }

  # Get current environment config
  env = local.environments[var.environment]
}

# =============================================================================
# Zone
# =============================================================================

resource "norainadns_zone" "main" {
  domain_name = var.domain_name
  description = "Main domain for ${var.environment}"
}

# =============================================================================
# Monitors (only for staging and prod)
# =============================================================================

resource "norainadns_monitor" "web" {
  count = local.env.monitoring ? 1 : 0

  name        = "${var.environment} Web Server"
  protocol    = "HTTPS"
  target_host = local.env.subdomain != "" ? "${local.env.subdomain}.${var.domain_name}" : var.domain_name
  target_path = "/health"

  check_interval_seconds = 60
  failure_threshold      = 3
}

resource "norainadns_monitor" "api" {
  count = local.env.monitoring ? 1 : 0

  name        = "${var.environment} API Server"
  protocol    = "HTTPS"
  target_host = local.env.subdomain != "" ? "api.${local.env.subdomain}.${var.domain_name}" : "api.${var.domain_name}"
  target_path = "/healthz"

  check_interval_seconds = 60
  failure_threshold      = 3
}

# =============================================================================
# DNS Records
# =============================================================================

# Web record
resource "norainadns_record" "web" {
  zone_id = norainadns_zone.main.id
  name    = local.env.subdomain != "" ? local.env.subdomain : "@"
  type    = "A"
  ttl     = local.env.ttl

  values {
    value = local.env.web_ip
  }
}

# WWW record (CNAME to main)
resource "norainadns_record" "www" {
  zone_id = norainadns_zone.main.id
  name    = local.env.subdomain != "" ? "www.${local.env.subdomain}" : "www"
  type    = "CNAME"
  ttl     = local.env.ttl

  values {
    value = local.env.subdomain != "" ? "${local.env.subdomain}.${var.domain_name}" : var.domain_name
  }
}

# API record
resource "norainadns_record" "api" {
  zone_id = norainadns_zone.main.id
  name    = local.env.subdomain != "" ? "api.${local.env.subdomain}" : "api"
  type    = "A"
  ttl     = local.env.ttl

  values {
    value = local.env.api_ip
  }
}

# =============================================================================
# Outputs
# =============================================================================

output "nameservers" {
  value = norainadns_zone.main.nameservers
}

output "web_fqdn" {
  value = norainadns_record.web.fqdn
}

output "api_fqdn" {
  value = norainadns_record.api.fqdn
}

output "environment" {
  value = var.environment
}
