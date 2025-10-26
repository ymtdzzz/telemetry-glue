variable "project_id" {
  description = "GCP project ID"
  type        = string
}

variable "region" {
  description = "The GCP region"
  type        = string
  default     = "asia-northeast1"
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

variable "secret_ids" {
  description = "List of Secret Manager secret IDs to create"
  type = object({
    slack_bot_token          = string
    slack_verification_token = string
  })
}
