resource "google_pubsub_topic" "slack_topic" {
  name = "${var.environment}-${var.prefix}-slack-topic"

  message_retention_duration = "3600s" # 1 hour
}
