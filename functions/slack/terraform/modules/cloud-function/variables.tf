variable "project_id" {
  description = "The GCP project ID"
  type        = string
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
}

variable "region" {
  description = "The GCP region"
  type        = string
  default     = "asia-northeast1"
}

variable "secret_ids" {
  description = "Secret Manager secret IDs from secret-manager module"
  type = object({
    slack_bot_token      = string
    slack_signing_secret = string
    newrelic_api_key     = string
    newrelic_account_id  = string
  })
}