output "terraform_state_bucket" {
  description = "Name of the Terraform state bucket"
  value       = google_storage_bucket.terraform_state.name
}

# Secret Manager outputs
output "slack_bot_token_secret_id" {
  description = "The secret ID for the Slack Bot Token"
  value       = module.secret_manager.slack_bot_token_secret_id
}

output "slack_signing_secret_secret_id" {
  description = "The secret ID for the Slack Signing Secret"
  value       = module.secret_manager.slack_signing_secret_secret_id
}

output "newrelic_api_key_secret_id" {
  description = "The secret ID for the NewRelic API Key"
  value       = module.secret_manager.newrelic_api_key_secret_id
}

output "newrelic_account_id_secret_id" {
  description = "The secret ID for the NewRelic Account ID"
  value       = module.secret_manager.newrelic_account_id_secret_id
}

# Cloud Function outputs
output "cloud_functions_service_account_email" {
  description = "Email of the Cloud Functions service account"
  value       = module.cloud_function.service_account_email
}

output "cloud_functions_service_account_name" {
  description = "Name of the Cloud Functions service account"
  value       = module.cloud_function.service_account_name
}
