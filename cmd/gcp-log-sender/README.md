# GCP Log Sender

A CLI tool to send test logs to Google Cloud Logging with specified trace IDs.

## Purpose

This tool is designed to create test data for developing and testing trace ID-based log retrieval functionality from GCP Cloud Logging API.

## Prerequisites

1. **Google Cloud Project** with Cloud Logging API enabled
2. **Authentication** via one of:
   - Service Account JSON file with `logging.writer` permissions
   - Application Default Credentials (ADC)
   - Set `GOOGLE_APPLICATION_CREDENTIALS` environment variable

## Usage

```bash
# Build the tool
go build -o gcp-log-sender ./cmd/gcp-log-sender

# Send a single log entry
./gcp-log-sender --project-id=YOUR_PROJECT_ID --trace-id=abc123def456

# Send multiple log entries with custom message
./gcp-log-sender \
  --project-id=YOUR_PROJECT_ID \
  --trace-id=abc123def456 \
  --message="Custom test message" \
  --count=5

# Send with different severity level
./gcp-log-sender \
  --project-id=YOUR_PROJECT_ID \
  --trace-id=abc123def456 \
  --severity=WARNING
```

## CLI Flags

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--project-id` | Yes | - | GCP Project ID |
| `--trace-id` | Yes | - | Trace ID to associate with logs |
| `--message` | No | "Test log message from gcp-log-sender" | Log message content |
| `--count` | No | 1 | Number of log entries to send |
| `--severity` | No | INFO | Log severity level |

## Severity Levels

- DEFAULT
- DEBUG  
- INFO
- NOTICE
- WARNING
- ERROR
- CRITICAL
- ALERT
- EMERGENCY

## Example Output

```
Successfully sent log entry 1/3
  Trace: projects/my-project/traces/abc123def456
  Message: Test log message from gcp-log-sender (entry 1/3)
  Response: &{...}

Successfully sent log entry 2/3
  Trace: projects/my-project/traces/abc123def456
  Message: Test log message from gcp-log-sender (entry 2/3)
  Response: &{...}

Successfully sent log entry 3/3
  Trace: projects/my-project/traces/abc123def456
  Message: Test log message from gcp-log-sender (entry 3/3)
  Response: &{...}

Completed sending 3 log entries with trace ID: abc123def456
You can now test log retrieval using this trace ID in GCP Cloud Logging
```

## Verification

After sending logs, you can verify them in:

1. **GCP Console**: Cloud Logging > Logs Explorer
2. **Filter by trace ID**: `trace="projects/YOUR_PROJECT_ID/traces/YOUR_TRACE_ID"`
3. **CLI tool** (once implemented): Use the trace ID with log retrieval functionality

## Notes

- Logs are sent to log name: `projects/PROJECT_ID/logs/gcp-log-sender`
- Each entry includes a "sender" label for easy identification
- Small delay (100ms) between multiple entries to ensure proper ordering
- Uses `global` monitored resource type for simplicity