variable "project_id" {
  description = "GCP project ID"
  type        = string
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
}

variable "prefix" {
  description = "Prefix for resource names"
  type        = string
  default     = "telemetry-glue"
}
