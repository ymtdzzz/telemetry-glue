resource "google_pubsub_topic" "slack_topic" {
  name = "${var.environment}-${var.prefix}-slack-topic"
}
