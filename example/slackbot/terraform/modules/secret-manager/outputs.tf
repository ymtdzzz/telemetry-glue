output "slack_bot_token_secret_id" {
  description = "The secret ID for the Slack Bot"
  value       = google_secret_manager_secret.slack_bot_token.secret_id
}

output "slack_verification_token_secret_id" {
  description = "The signing secret ID for the Slack Bot"
  value       = google_secret_manager_secret.slack_verification_token.secret_id
}
