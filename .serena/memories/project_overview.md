# telemetry-glue Project Overview

## Purpose
telemetry-glue is a unified interface library and CLI tool for querying telemetry data such as traces and logs across multiple observability backends like NewRelic, Google Cloud, Datadog, etc.

## Key Features
- **Unified API**: Query different observability backends using the same interface
- **CLI & Library**: Use as a command-line tool or Go library
- **Multiple Backends**: Support for NewRelic (with more backends coming)
- **Search Values**: Find unique attribute values with wildcard support
- **Top Traces**: Find top traces by duration for specific attribute values
- **Spans**: Get all spans within a specific trace for detailed analysis
- **Test Tools**: OpenTelemetry test trace sender for validation

## Architecture
- **Backend-specific implementations**: Each observability backend (NewRelic, Google Cloud, etc.) has its own subcommand implementation
- **Minimal abstraction**: Common functionality (output formatting, configuration management) is shared, while vendor-specific logic is encapsulated
- **Library-first design**: Core logic implemented as library, CLI serves as wrapper

## Usage Examples
```bash
# Build main CLI
go build ./cmd/telemetry-glue

# Search attribute values with NewRelic
./telemetry-glue newrelic search-values --entity "your-app" --attribute http.path --query "*user*" --since 1h

# Find top traces for specific attribute value
./telemetry-glue newrelic top-traces --entity "your-app" --attribute http.path --value "/api/users" --since 1h

# Get spans for specific trace
./telemetry-glue newrelic spans --trace-id "d8a60536187fa0927e45911f8c0dd64b" --since 1h
```

## Tech Stack
- **Language**: Go 1.24.5
- **CLI Framework**: Cobra
- **Key Dependencies**:
  - github.com/spf13/cobra - CLI construction
  - github.com/newrelic/newrelic-client-go/v2 - NewRelic API
  - go.opentelemetry.io/otel - OpenTelemetry
  - google.golang.org/api - Google Cloud API
  - github.com/joho/godotenv - Environment variable management

## Directory Structure
```
cmd/                        # Entry points
├── telemetry-glue/         # Main CLI
├── otel-test-sender/       # OpenTelemetry test sender
└── gcp-log-sender/         # GCP log sender
internal/                   # Internal implementation
├── backend/               # Backend-specific implementations
│   ├── newrelic/          # NewRelic integration
│   └── gcp/               # Google Cloud integration
├── analyzer/              # Analysis functionality
├── output/                # Output formatting
└── pipeline/              # Pipeline processing
docs/                      # Documentation
```