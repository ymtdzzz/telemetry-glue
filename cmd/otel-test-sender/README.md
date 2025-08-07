# OpenTelemetry Test Trace Sender

A standalone command for sending test OpenTelemetry traces to NewRelic for validation and testing purposes.

## Usage

### 1. Build

```bash
go build -o bin/otel-test-sender cmd/otel-test-sender/main.go
```

### 2. Environment Variables Setup

You can configure the tool using either of the following methods:

#### Option A: Using .env file
```bash
# Copy and edit the .env file from project root
cp .env.example .env
# Set the following in .env file:
# NEW_RELIC_API_KEY=your-newrelic-api-key
# NEW_RELIC_OTLP_ENDPOINT=otlp.nr-data.net:4318
```

#### Option B: Direct environment variables
```bash
export NEW_RELIC_API_KEY="your-newrelic-api-key"
export NEW_RELIC_OTLP_ENDPOINT="otlp.nr-data.net:4318"  # Optional
```

**Note**: When ENV is set to "production", .env files will not be loaded.

### 3. Execute

```bash
./bin/otel-test-sender
```

## Test Trace Structure

This command sends 3 test traces to NewRelic. Each trace contains the following structure:

### Parent Span: HTTP Request
- **Operation**: `http_request_N` (where N is 1-3)
- **Attributes**:
  - `http.method`: "GET"
  - `http.url`: "https://api.example.com/users/N"
  - `http.status_code`: 200
  - `user.id`: "user_N"

### Child Span 1: Database Query
- **Operation**: `database_query`
- **Attributes**:
  - `db.system`: "postgresql"
  - `db.name`: "userdb"
  - `db.statement`: "SELECT * FROM users WHERE id = N"
  - `db.operation`: "SELECT"
- **Duration**: 10-50ms (random)

### Child Span 2: External API Call
- **Operation**: `external_api_call`
- **Attributes**:
  - `http.method`: "POST"
  - `http.url`: "https://external-service.com/validate"
  - `http.status_code`: 200
  - `service.name`: "validation-service"
- **Duration**: 20-100ms (random)

### Child Span 3: Business Logic Processing
- **Operation**: `business_logic_processing`
- **Attributes**:
  - `operation`: "user_data_enrichment"
  - `processed_records`: 1-10 (random)
  - `cache_hit`: true/false (random)
- **Duration**: 5-30ms (random)

## Service Information

- **Service Name**: `test-service`
- **Service Version**: `1.0.0`
- **Environment**: `test`

## Viewing Results in NewRelic

After sending traces, you can view the results in NewRelic:

1. **APM & Services** > **test-service** > **Distributed tracing**
2. **Query your data** with NRQL queries like:
   ```sql
   SELECT * FROM Span WHERE service.name = 'test-service' SINCE 10 minutes ago
   ```

## Troubleshooting

### Error: "NEW_RELIC_API_KEY environment variable is required"
- NewRelic API key is not configured
- Set the API key in `.env` file or environment variables

### Traces not appearing in NewRelic
- Verify API key is correct
- Ensure OTLP is enabled in your NewRelic account
- Check network connectivity
- Verify endpoint URL (`NEW_RELIC_OTLP_ENDPOINT`) is correct

### 403 Forbidden Error
- Check API key permissions
- Verify OTLP is enabled for your NewRelic account
- Ensure you're using the correct regional endpoint (US vs EU)
- Check account ID if required by your NewRelic configuration

## Technical Details

- Uses OpenTelemetry OTLP HTTP exporter
- Sends traces to NewRelic's OTLP endpoint (default: `otlp.nr-data.net:4318`)
- Includes proper OpenTelemetry semantic conventions
- Simulates realistic application trace patterns