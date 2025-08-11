terraform {
  required_version = "~> 1.12.0"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 6.47.0"
    }
  }

  backend "gcs" {
    # Configuration will be provided via backend config file or CLI
    # Example:
    # bucket = "your-terraform-state-bucket"
    # prefix = "slack-functions"
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}
