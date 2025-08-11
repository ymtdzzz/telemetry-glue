# Secret Manager Secret for Slack Bot Token
resource "google_secret_manager_secret" "slack_bot_token" {
  project   = var.project_id
  secret_id = "${var.environment}-slack-bot-token"

  labels = {
    environment = var.environment
    purpose     = "slack-bot"
  }

  replication {
    auto {}
  }
}

# Secret Manager Secret for Slack Signing Secret
resource "google_secret_manager_secret" "slack_signing_secret" {
  project   = var.project_id
  secret_id = "${var.environment}-slack-signing-secret"

  labels = {
    environment = var.environment
    purpose     = "slack-bot"
  }

  replication {
    auto {}
  }
}

# Secret Manager Secret for NewRelic API Key
resource "google_secret_manager_secret" "newrelic_api_key" {
  project   = var.project_id
  secret_id = "${var.environment}-newrelic-api-key"

  labels = {
    environment = var.environment
    purpose     = "newrelic"
  }

  replication {
    auto {}
  }
}

# Secret Manager Secret for NewRelic Account ID
resource "google_secret_manager_secret" "newrelic_account_id" {
  project   = var.project_id
  secret_id = "${var.environment}-newrelic-account-id"

  labels = {
    environment = var.environment
    purpose     = "newrelic"
  }

  replication {
    auto {}
  }
}
