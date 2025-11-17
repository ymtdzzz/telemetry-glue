resource "google_secret_manager_secret" "slack_bot_token" {
  project   = var.project_id
  secret_id = "${var.environment}-${var.prefix}-slack-bot-token"

  labels = {
    environment = var.environment
    purpose     = "slack-bot"
  }

  replication {
    auto {}
  }
}

resource "google_secret_manager_secret" "slack_verification_token" {
  project   = var.project_id
  secret_id = "${var.environment}-${var.prefix}-slack-verification-token"

  labels = {
    environment = var.environment
    purpose     = "slack-bot"
  }

  replication {
    auto {}
  }
}

resource "google_secret_manager_secret" "new_relic_api_key" {
  project   = var.project_id
  secret_id = "${var.environment}-${var.prefix}-new-relic-api-key"

  labels = {
    environment = var.environment
    purpose     = "slack-bot"
  }

  replication {
    auto {}
  }
}

resource "google_secret_manager_secret" "new_relic_account_id" {
  project   = var.project_id
  secret_id = "${var.environment}-${var.prefix}-new-relic-account-id"

  labels = {
    environment = var.environment
    purpose     = "slack-bot"
  }

  replication {
    auto {}
  }
}
