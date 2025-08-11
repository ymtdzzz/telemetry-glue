output "service_account_email" {
  description = "Email of the Cloud Functions service account"
  value       = google_service_account.cloud_functions.email
}

output "service_account_name" {
  description = "Name of the Cloud Functions service account"
  value       = google_service_account.cloud_functions.name
}