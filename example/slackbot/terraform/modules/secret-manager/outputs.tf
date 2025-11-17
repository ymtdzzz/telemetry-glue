output "slack_bot_token_secret_id" {
  description = "The secret ID for the Slack Bot token"
  value       = google_secret_manager_secret.slack_bot_token.secret_id
}

output "slack_verification_token_secret_id" {
  description = "The secret ID for the Slack Verification token"
  value       = google_secret_manager_secret.slack_verification_token.secret_id
}

output "new_relic_api_key_secret_id" {
  description = "The secret ID for the New Relic API Key"
  value       = google_secret_manager_secret.new_relic_api_key.secret_id
}

output "new_relic_account_id_secret_id" {
  description = "The secret ID for the New Relic Account ID"
  value       = google_secret_manager_secret.new_relic_account_id.secret_id
}
