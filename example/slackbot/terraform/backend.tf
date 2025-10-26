terraform {
  required_version = "~> 1.13.0"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 7.8.0"
    }
  }

  backend "gcs" {
    bucket = "telemetry-glue-tfstate"
    prefix = "terraform/state"
  }
}
