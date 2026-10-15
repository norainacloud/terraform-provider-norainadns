# =============================================================================
# Variables for Failover Example
# =============================================================================

variable "norainadns_endpoint" {
  description = "The NorainaDNS API endpoint"
  type        = string
  default     = "https://dash.norainadns.com"
}

variable "norainadns_api_key" {
  description = "The NorainaDNS API key"
  type        = string
  sensitive   = true
}

# =============================================================================
# IP Addresses
# =============================================================================

variable "cdn_ip" {
  description = "IP address of the CDN for main website"
  type        = string
  default     = "192.0.2.10"
}

variable "origin_ip" {
  description = "IP address of the origin server for main website"
  type        = string
  default     = "192.0.2.20"
}

variable "api_cdn_ip" {
  description = "IP address of the CDN for API"
  type        = string
  default     = "192.0.2.30"
}

variable "api_origin_ip" {
  description = "IP address of the API origin server"
  type        = string
  default     = "192.0.2.40"
}
