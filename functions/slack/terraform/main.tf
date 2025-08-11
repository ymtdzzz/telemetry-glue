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
