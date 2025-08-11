# Service Account for Cloud Functions
resource "google_service_account" "cloud_functions" {
  project      = var.project_id
  account_id   = "${var.environment}-slack-functions"
  display_name = "Cloud Functions for Slack Bot (${var.environment})"
  description  = "Service account for Slack bot Cloud Functions in ${var.environment} environment"
}

# IAM Role: Secret Manager Secret Accessor for all secrets
resource "google_secret_manager_secret_iam_member" "slack_bot_token_accessor" {
  project   = var.project_id
  secret_id = var.secret_ids.slack_bot_token
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.cloud_functions.email}"
}

resource "google_secret_manager_secret_iam_member" "slack_signing_secret_accessor" {
  project   = var.project_id
  secret_id = var.secret_ids.slack_signing_secret
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.cloud_functions.email}"
}

resource "google_secret_manager_secret_iam_member" "newrelic_api_key_accessor" {
  project   = var.project_id
  secret_id = var.secret_ids.newrelic_api_key
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.cloud_functions.email}"
}

resource "google_secret_manager_secret_iam_member" "newrelic_account_id_accessor" {
  project   = var.project_id
  secret_id = var.secret_ids.newrelic_account_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.cloud_functions.email}"
}

# IAM Role: Cloud Tasks Enqueuer
resource "google_project_iam_member" "cloud_tasks_enqueuer" {
  project = var.project_id
  role    = "roles/cloudtasks.enqueuer"
  member  = "serviceAccount:${google_service_account.cloud_functions.email}"
}

# IAM Role: Vertex AI User
resource "google_project_iam_member" "vertex_ai_user" {
  project = var.project_id
  role    = "roles/aiplatform.user"
  member  = "serviceAccount:${google_service_account.cloud_functions.email}"
}

# IAM Role: Logging Writer
resource "google_project_iam_member" "logging_writer" {
  project = var.project_id
  role    = "roles/logging.logWriter"
  member  = "serviceAccount:${google_service_account.cloud_functions.email}"
}
