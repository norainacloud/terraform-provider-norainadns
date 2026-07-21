# =============================================================================
# Failover DNS Example
# =============================================================================
# This example shows how to configure automatic DNS failover using health
# monitors. When the primary (CDN) becomes unhealthy, traffic automatically
# switches to the secondary (origin) IP.

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
# Zone
# =============================================================================

resource "norainadns_zone" "app" {
  domain_name = "myapp.io"
  description = "Application domain with failover"
}

# =============================================================================
# Health Monitors
# =============================================================================

# Monitor the origin server directly (bypassing CDN)
# This ensures we detect origin failures even when CDN is caching
resource "norainadns_monitor" "origin" {
  name        = "Origin Server"
  description = "Health check for origin server behind CDN"
  protocol    = "HTTPS"
  target_host = "myapp.io"       # Used for Host header
  target_path = "/health"

  # Connect directly to origin IP, bypassing CDN
  force_ip_address = var.origin_ip

  # Check every minute for faster failover
  check_interval_seconds = 60

  # Mark unhealthy after 2 failures
  failure_threshold = 2

  # 10 second timeout
  timeout_seconds = 10
}

# Monitor for the API endpoint
resource "norainadns_monitor" "api_origin" {
  name        = "API Origin"
  description = "Health check for API origin"
  protocol    = "HTTPS"
  target_host = "api.myapp.io"
  target_path = "/healthz"

  force_ip_address = var.api_origin_ip

  check_interval_seconds = 60
  failure_threshold      = 2
  timeout_seconds        = 10
}

# =============================================================================
# DNS Records with Failover
# =============================================================================

# Main website - failover from CDN to origin
resource "norainadns_record" "www" {
  zone_id        = norainadns_zone.app.id
  name           = "www"
  type           = "A"
  ttl            = 60  # Low TTL for faster failover
  is_dynamic     = true
  routing_policy = "failover"

  # Primary: CDN IP (served when origin is healthy)
  values {
    value      = var.cdn_ip
    priority   = 0
    monitor_id = norainadns_monitor.origin.id
  }

  # Secondary: Origin IP (fallback when CDN/origin is unhealthy)
  values {
    value    = var.origin_ip
    priority = 1
  }
}

# Root domain - same failover setup
resource "norainadns_record" "root" {
  zone_id        = norainadns_zone.app.id
  name           = "@"
  type           = "A"
  ttl            = 60
  is_dynamic     = true
  routing_policy = "failover"

  values {
    value      = var.cdn_ip
    priority   = 0
    monitor_id = norainadns_monitor.origin.id
  }

  values {
    value    = var.origin_ip
    priority = 1
  }
}

# API endpoint - failover from CDN to API origin
resource "norainadns_record" "api" {
  zone_id        = norainadns_zone.app.id
  name           = "api"
  type           = "A"
  ttl            = 60
  is_dynamic     = true
  routing_policy = "failover"

  values {
    value      = var.api_cdn_ip
    priority   = 0
    monitor_id = norainadns_monitor.api_origin.id
  }

  values {
    value    = var.api_origin_ip
    priority = 1
  }
}

# =============================================================================
# Outputs
# =============================================================================

output "nameservers" {
  description = "Nameservers to configure at registrar"
  value       = norainadns_zone.app.nameservers
}

output "origin_monitor_status" {
  description = "Current origin health status"
  value       = norainadns_monitor.origin.current_status
}

output "api_monitor_status" {
  description = "Current API health status"
  value       = norainadns_monitor.api_origin.current_status
}
