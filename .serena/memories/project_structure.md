# telemetry-glue Project Structure Details

## Directory Structure and Roles

### cmd/ - Entry Points

- **telemetry-glue/**: Main CLI implementation
  - `main.go`: CLI entry point
  - `cmd/root.go`: Cobra root command definition
  - `analyze/cmd.go`: Analysis commands
  - `gcp/cmd.go`: Google Cloud subcommands
  - `newrelic/cmd.go`: NewRelic subcommands
  - `common/flags.go`: Common flag definitions

- **otel-test-sender/**: OpenTelemetry test trace sender tool
- **gcp-log-sender/**: Google Cloud log sender tool

### internal/ - Internal Implementation

- **backend/**: Backend-specific implementations
  - `newrelic/client.go`: NewRelic API integration
  - `gcp/client.go`: Google Cloud API integration

- **analyzer/**: Analysis functionality
  - `aggregator.go`: Data aggregation processing
  - `prompt.go`: Prompt processing
  - `vertexai.go`: Vertex AI integration

- **output/**: Output formatting functionality
  - `formatter.go`: Output format processing

- **pipeline/**: Pipeline processing
  - `passthrough.go`: Passthrough processing

### Configuration Files

- `.env.example`: Environment variable template
- `go.mod`, `go.sum`: Go dependency management
- `.gitignore`: Git exclusion settings
- `design.md`: Design documentation
- `TODO.md`: TODO items
- `memory.md`: Memory-related documentation

## Architecture Characteristics

- **Backend independence**: Each observability backend has independent implementation
- **CLI-centric design**: CLI is primary interface, also usable as library
- **Plugin-style extension**: Easy to add new backends
- **Common output format**: Consistent output format regardless of backend

## Extension Points

1. Add new backend: Create new directory in `internal/backend/`
2. Add new command: Add subcommand in `cmd/telemetry-glue/`
3. Add new output format: Add functionality in `internal/output/`
4. Add new analysis feature: Add functionality in `internal/analyzer/`

## Key Design Principles

- Minimal abstraction between backends
- Vendor-specific logic encapsulated within each backend
- Shared utilities for common operations (formatting, configuration)
- Library-first design with CLI wrapper
- Independent evolution of backend implementations
