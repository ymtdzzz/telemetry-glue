output "slack_bot_token_secret_id" {
  description = "The secret ID for the Slack Bot Token"
  value       = google_secret_manager_secret.slack_bot_token.secret_id
}

output "slack_signing_secret_secret_id" {
  description = "The secret ID for the Slack Signing Secret"
  value       = google_secret_manager_secret.slack_signing_secret.secret_id
}

output "newrelic_api_key_secret_id" {
  description = "The secret ID for the NewRelic API Key"
  value       = google_secret_manager_secret.newrelic_api_key.secret_id
}

output "newrelic_account_id_secret_id" {
  description = "The secret ID for the NewRelic Account ID"
  value       = google_secret_manager_secret.newrelic_account_id.secret_id
}
