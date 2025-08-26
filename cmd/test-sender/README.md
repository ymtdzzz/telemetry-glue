# Test Sender

A unified CLI tool for sending test telemetry data to various observability backends.

## Overview

Test Sender combines the functionality of the previous separate tools (`gcp-log-sender` and `otel-test-sender`) into a single, well-structured command using Cobra.

## Installation

```bash
# Build from source
go build -o test-sender ./cmd/test-sender

# Or build to specific location
go build -o bin/test-sender ./cmd/test-sender
```

## Usage

### Send Test Logs to GCP Cloud Logging

```bash
# Basic usage
./test-sender log gcp --project-id=my-project --trace-id=abc123def456

# Send multiple logs with custom message and severity
./test-sender log gcp \
  --project-id=my-project \
  --trace-id=abc123def456 \
  --message="Custom test message" \
  --count=5 \
  --severity=WARNING
```

**GCP Log Flags:**
- `--project-id` (required): GCP Project ID
- `--trace-id` (required): Trace ID to associate with logs
- `--message`: Log message content (default: "Test log message from test-sender")
- `--count`: Number of log entries to send (default: 1)
- `--severity`: Log severity level (default: INFO)

**Supported Severity Levels:**
DEFAULT, DEBUG, INFO, NOTICE, WARNING, ERROR, CRITICAL, ALERT, EMERGENCY

### Send Test Traces to New Relic

```bash
# Using environment variables
export NEW_RELIC_INGEST_API_KEY="your-api-key"
./test-sender trace newrelic

# Using command line flags
./test-sender trace newrelic --api-key=your-api-key --count=5

# Custom endpoint
./test-sender trace newrelic \
  --api-key=your-api-key \
  --endpoint=otlp.eu01.nr-data.net:4318 \
  --count=10
```

**New Relic Trace Flags:**
- `--api-key`: New Relic Ingest API Key (or set `NEW_RELIC_INGEST_API_KEY` env var)
- `--endpoint`: New Relic OTLP endpoint (default: "otlp.nr-data.net:4318")
- `--count`: Number of test traces to send (default: 3)

## Command Structure

```
test-sender
├── log
│   └── gcp          # Send logs to GCP Cloud Logging
└── trace
    └── newrelic     # Send traces to New Relic
```

## Features

- **Unified Interface**: Single command with logical subcommand structure
- **Cobra Integration**: Rich help system, auto-completion, and consistent UX
- **Backward Compatibility**: All original functionality preserved
- **Extensible**: Easy to add new backends (AWS, Datadog, etc.)
- **Environment Support**: Supports both flags and environment variables

## Test Trace Structure

New Relic traces include realistic test data:

### Parent Span: HTTP Request
- Operation: `http_request_N`
- Attributes: HTTP method, URL, status code, user ID

### Child Spans:
1. **Database Query** (10-50ms)
   - PostgreSQL SELECT operation
   - Database system and statement details

2. **External API Call** (20-100ms)
   - HTTP POST to validation service
   - Service name and response details

3. **Business Logic Processing** (5-30ms)
   - User data enrichment operation
   - Processing metrics and cache hit status

## Environment Variables

### New Relic Configuration
- `NEW_RELIC_INGEST_API_KEY`: Your New Relic Ingest API Key
- `NEW_RELIC_OTLP_ENDPOINT`: OTLP endpoint (optional)
- `ENV`: When set to "production", .env files won't be loaded

### GCP Authentication
- `GOOGLE_APPLICATION_CREDENTIALS`: Path to service account JSON
- Or use Application Default Credentials (ADC)

## Examples

```bash
# Send 5 warning logs to GCP with custom trace ID
./test-sender log gcp \
  --project-id=my-project \
  --trace-id=custom-trace-123 \
  --severity=WARNING \
  --count=5

# Send 10 test traces to New Relic EU
./test-sender trace newrelic \
  --api-key=your-key \
  --endpoint=otlp.eu01.nr-data.net:4318 \
  --count=10

# Quick test with environment variables
export NEW_RELIC_INGEST_API_KEY="your-key"
./test-sender trace newrelic --count=1
```

## Migration from Old Commands

### gcp-log-sender → test-sender log gcp
```bash
# Old
./gcp-log-sender --project-id=proj --trace-id=trace123 --count=3

# New  
./test-sender log gcp --project-id=proj --trace-id=trace123 --count=3
```

### otel-test-sender → test-sender trace newrelic
```bash
# Old
./otel-test-sender

# New
./test-sender trace newrelic
```

## Verification

### GCP Logs
- **Console**: Cloud Logging > Logs Explorer
- **Filter**: `trace="projects/PROJECT_ID/traces/TRACE_ID"`
- **Log Name**: `projects/PROJECT_ID/logs/test-sender`

### New Relic Traces
- **APM & Services** > **test-service** > **Distributed tracing**
- **NRQL Query**: `SELECT * FROM Span WHERE service.name = 'test-service' SINCE 10 minutes ago`