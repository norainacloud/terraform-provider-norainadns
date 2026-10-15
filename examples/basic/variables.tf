# =============================================================================
# Variables for Basic Example
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
