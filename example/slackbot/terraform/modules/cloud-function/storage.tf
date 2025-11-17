resource "google_storage_bucket" "function_source_code" {
  name                        = "${var.environment}-${var.prefix}-function-src"
  location                    = "US"
  uniform_bucket_level_access = true
}

data "archive_file" "function_src" {
  type        = "zip"
  source_dir  = "${path.module}/../../../../../cmd/slackbot/handler"
  output_path = "${path.module}/function_src.zip"
  excludes = [
    ".git",
    ".github",
    ".serena",
    "bin",
    "example",
    "functions",
  ]
}

resource "google_storage_bucket_object" "function_src" {
  name   = "function_src-${data.archive_file.function_src.output_md5}.zip"
  bucket = google_storage_bucket.function_source_code.name
  source = data.archive_file.function_src.output_path
}
