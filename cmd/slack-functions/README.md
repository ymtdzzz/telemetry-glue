# Slack Functions (Unified)

This directory contains the unified Slack Bot functionality deployed as Google Cloud Functions.

## Overview

- **Purpose**: Unified implementation of Slack Bot and trace analysis worker
- **Framework**: Google Cloud Functions Framework for Go
- **Deployment Target**: Cloud Functions (unified implementation)

## Endpoints

### `/SlackEvent`

- **Role**: Slack Events API request handling
- **Features**: App Mention, Message Event processing
- **Usage**: Start trace analysis with `@bot analyze <trace_id>`

### `/AnalyzeTrace`

- **Role**: Trace analysis worker
- **Features**: Trace analysis using New Relic API
- **Usage**: Analysis processing called from Cloud Tasks

## Architecture

```
Slack → Events API → /SlackEvent → Cloud Tasks → /AnalyzeTrace → New Relic API
                                                      ↓
                                               Vertex AI (Analysis)
                                                      ↓
                                                Slack (Results)
```

## Local Development

```bash
cd cmd/slack-functions
go run *.go
```

The server starts on port 8080.

## Testing

### Slack Event Testing

```bash
curl -X POST http://localhost:8080/SlackEvent \
  -H "Content-Type: application/json" \
  -d '{
    "type": "event_callback",
    "event": {
      "type": "app_mention",
      "channel": "C1234567890",
      "user": "U0987654321",
      "text": "<@U1111111111> analyze abc123def456",
      "ts": "1234567890.123456"
    }
  }'
```

### Trace Analysis Testing

```bash
curl -X POST http://localhost:8080/AnalyzeTrace \
  -H "Content-Type: application/json" \
  -d '{
    "trace_id": "test-trace-id",
    "channel_id": "test-channel",
    "thread_ts": "test-thread"
  }'
```

## Cloud Functions Deployment

### Deploy Both Endpoints

```bash
# Slack Event handler
gcloud functions deploy slack-event \
  --runtime go123 \
  --trigger-http \
  --allow-unauthenticated \
  --source cmd/slack-functions \
  --entry-point SlackEvent

# Analyze handler
gcloud functions deploy analyze-trace \
  --runtime go123 \
  --trigger-http \
  --allow-unauthenticated \
  --source cmd/slack-functions \
  --entry-point AnalyzeTrace
```

### Cloud Run Deployment (Recommended)

```bash
gcloud run deploy slack-functions \
  --source cmd/slack-functions \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated
```

## Environment Variables

The following environment variables are required:

- `NEWRELIC_API_KEY`: New Relic API key
- `NEWRELIC_ACCOUNT_ID`: New Relic account ID
- `SLACK_BOT_TOKEN`: Slack Bot token
- `VERTEXAI_PROJECT_ID`: Vertex AI project ID
- `VERTEXAI_LOCATION`: Vertex AI location
- `GCP_PROJECT_ID`: Google Cloud project ID
- `GCP_LOCATION`: Cloud Tasks location
- `GCP_QUEUE_NAME`: Cloud Tasks queue name

## Slack App Configuration

### Events API

- **Request URL**: `https://your-function-url/SlackEvent`
- **Subscribe to bot events**:
  - `app_mention`
  - `message.channels` (optional)

### OAuth & Permissions

- **Bot Token Scopes**:
  - `chat:write`
  - `app_mentions:read`
  - `channels:read`

## Improvements from Previous Implementation

- **Unified Deployment**: Single deployment for both functions
- **Code Sharing**: Shared configuration and utilities
- **Simplification**: Cleaner implementation with Functions Framework
- **Scalable**: Auto-scaling with Cloud Functions
- **Maintainable**: Clear structure with separation of concerns

## File Structure

```
cmd/slack-functions/
├── main.go              # Entry point, register both endpoints
├── event_handler.go     # Slack Events API processing
├── analyze_handler.go   # Trace analysis processing
└── README.md           # This file
```
