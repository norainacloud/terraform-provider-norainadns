# =============================================================================
# Variables for Multi-Environment Example
# =============================================================================

variable "norainadns_endpoint" {
  description = "The NorainaDNS API endpoint"
  type        = string
  default     = "https://dns.noraina.cloud"
}

variable "norainadns_api_key" {
  description = "The NorainaDNS API key"
  type        = string
  sensitive   = true
}

variable "domain_name" {
  description = "The base domain name"
  type        = string
  default     = "example.com"
}

variable "environment" {
  description = "The environment (dev, staging, prod)"
  type        = string
  default     = "dev"

  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "Environment must be dev, staging, or prod."
  }
}
