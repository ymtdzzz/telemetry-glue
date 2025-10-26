resource "google_service_account" "http_function" {
  project      = var.project_id
  account_id   = "${var.environment}-${var.prefix}-http-fn"
  display_name = "HTTP Cloud Functions for Slack Bot (${var.environment})"
  description  = "Service account for Slack bot HTTP Cloud Functions in ${var.environment} environment"
}

resource "google_secret_manager_secret_iam_member" "http_slack_bot_token_accessor" {
  project   = var.project_id
  secret_id = var.secret_ids.slack_bot_token
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.http_function.email}"
}

resource "google_secret_manager_secret_iam_member" "http_slack_verification_token_accessor" {
  project   = var.project_id
  secret_id = var.secret_ids.slack_verification_token
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.http_function.email}"
}

resource "google_cloud_run_service_iam_member" "http_public_invoker" {
  location = var.region
  project  = var.project_id
  service  = google_cloudfunctions2_function.http_function.name

  role   = "roles/run.invoker"
  member = "allUsers"
}

resource "google_project_iam_member" "http_pubsub_publisher" {
  project = var.project_id
  role    = "roles/pubsub.publisher"
  member  = "serviceAccount:${google_service_account.http_function.email}"
}

resource "google_cloudfunctions2_function" "http_function" {
  name        = "${var.environment}-${var.prefix}-slack-trigger"
  location    = var.region
  description = "HTTP triggered Cloud Function for Slackbot"

  build_config {
    runtime     = "go124"
    entry_point = "HandleCommand"
    source {
      storage_source {
        bucket = google_storage_bucket.function_source_code.name
        object = google_storage_bucket_object.function_src.name
      }
    }
  }

  service_config {
    available_memory   = "256M"
    max_instance_count = 1
    min_instance_count = 0
    ingress_settings   = "ALLOW_ALL"

    environment_variables = {
      GCP_PROJECT_ID      = var.project_id
      GCP_PUBSUB_TOPIC_ID = google_pubsub_topic.slack_topic.id
    }

    secret_environment_variables {
      key        = "SLACK_BOT_TOKEN"
      project_id = var.project_id
      secret     = var.secret_ids.slack_bot_token
      version    = "latest"
    }

    secret_environment_variables {
      key        = "SLACK_VERIFICATION_TOKEN"
      project_id = var.project_id
      secret     = var.secret_ids.slack_verification_token
      version    = "latest"
    }

    service_account_email = google_service_account.http_function.email
  }
}
