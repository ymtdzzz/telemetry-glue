locals {
  common_labels = {
    environment = var.environment
    project     = "telemetry-glue"
    component   = "slack-functions"
  }
}

# Storage bucket for Terraform state (if needed for other resources)
resource "google_storage_bucket" "terraform_state" {
  name     = "${var.project_id}-telemetry-glue-tfstate-${var.environment}"
  location = var.region

  uniform_bucket_level_access = true
  versioning {
    enabled = true
  }

  labels = local.common_labels

  lifecycle_rule {
    condition {
      age = 30
    }
    action {
      type = "Delete"
    }
  }
}

# Secret Manager for storing sensitive configuration
module "secret_manager" {
  source = "./modules/secret-manager"

  project_id  = var.project_id
  environment = var.environment
  region      = var.region
}

# Cloud Functions with IAM for Slack bot
module "cloud_function" {
  source = "./modules/cloud-function"

  project_id  = var.project_id
  environment = var.environment
  region      = var.region

  secret_ids = {
    slack_bot_token      = module.secret_manager.slack_bot_token_secret_id
    slack_signing_secret = module.secret_manager.slack_signing_secret_secret_id
    newrelic_api_key     = module.secret_manager.newrelic_api_key_secret_id
    newrelic_account_id  = module.secret_manager.newrelic_account_id_secret_id
  }
}
