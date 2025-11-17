resource "google_service_account" "pubsub_function" {
  project      = var.project_id
  account_id   = "${var.environment}-${var.prefix}-pubsub-fn"
  display_name = "HTTP Cloud Functions for Slack Bot (${var.environment})"
  description  = "Service account for Slack bot HTTP Cloud Functions in ${var.environment} environment"
}

resource "google_secret_manager_secret_iam_member" "pubsub_slack_bot_token_accessor" {
  project   = var.project_id
  secret_id = var.secret_ids.slack_bot_token
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.pubsub_function.email}"
}

resource "google_secret_manager_secret_iam_member" "pubsub_new_relic_api_key_accessor" {
  project   = var.project_id
  secret_id = var.secret_ids.new_relic_api_key
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.pubsub_function.email}"
}

resource "google_secret_manager_secret_iam_member" "pubsub_new_relic_account_id_accessor" {
  project   = var.project_id
  secret_id = var.secret_ids.new_relic_account_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.pubsub_function.email}"
}

resource "google_project_iam_member" "pubsub_vertexai_user" {
  project = var.project_id
  role    = "roles/aiplatform.user"
  member  = "serviceAccount:${google_service_account.pubsub_function.email}"
}

resource "google_cloudfunctions2_function" "pubsub_function" {
  name        = "${var.environment}-${var.prefix}-slack-analyze"
  location    = var.region
  description = "Pubsub triggered Cloud Function for Slackbot"

  build_config {
    runtime     = "go124"
    entry_point = "HandlePubsub"
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
    ingress_settings   = "ALLOW_INTERNAL_ONLY"

    environment_variables = {
      GLUE_SPAN_BACKEND             = "newrelic"
      ANALYZER_LANGUAGE             = "ja"
      ANALYZER_VERTEX_AI_MODEL_NAME = "gemini-2.5-flash-lite"
      ANALYZER_VERTEX_AI_PROJECT_ID = var.project_id
      ANALYZER_VERTEX_AI_LOCATION   = var.region
    }

    secret_environment_variables {
      key        = "GLUE_NEW_RELIC_API_KEY"
      project_id = var.project_id
      secret     = var.secret_ids.new_relic_api_key
      version    = "latest"
    }

    secret_environment_variables {
      key        = "GLUE_NEW_RELIC_ACCOUNT_ID"
      project_id = var.project_id
      secret     = var.secret_ids.new_relic_account_id
      version    = "latest"
    }

    secret_environment_variables {
      key        = "SLACK_BOT_TOKEN"
      project_id = var.project_id
      secret     = var.secret_ids.slack_bot_token
      version    = "latest"
    }

    service_account_email = google_service_account.pubsub_function.email
  }

  event_trigger {
    trigger_region = var.region
    event_type     = "google.cloud.pubsub.topic.v1.messagePublished"
    pubsub_topic   = google_pubsub_topic.slack_topic.id
    retry_policy   = "RETRY_POLICY_RETRY"
  }
}
