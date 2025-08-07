# telemetry-glue

## Overview

telemetry-glue is a library and CLI tool for querying telemetry data such as traces and logs across different observability backends like NewRelic, Google Cloud, Datadog, etc.

The tool provides backend-specific commands that handle the unique concepts and APIs of each vendor (e.g., entities in NewRelic, projects in Google Cloud) while maintaining consistent output formats for downstream processing.

## Usage

The CLI uses subcommands for each backend vendor to handle vendor-specific arguments and concepts. Examples below show the general structure, but specific arguments and options are subject to change:

```sh
# NewRelic examples (arguments are subject to change)
telemetry-glue newrelic search-values --entity my-app --attribute "http.method" --query "*users*" --range "2025-08-01T12:00:00+09:00,2025-08-02T00:00:00+09:00"
# output =>
# available values:
#   - GET
#   - POST
#   - ...

telemetry-glue newrelic top-traces --entity my-app --attribute "http.path" --value "/admin/users/new" --range "..."
# output =>
# result link: https://one.newrelic....
# TOP 5 (duration):
#   - 2025-08-01T12:01:12+09:00 {trace_id} -- 3.23s
#   - ... {trace_id} -- 2.8s
#   - ...

# Google Cloud examples (arguments are subject to change)
telemetry-glue googlecloud search-values --project my-project --attribute "http.method" --query "*users*" --range "..."

telemetry-glue googlecloud list-spans --project my-project --trace {trace_id} --range "..."
```

All commands output structured data (JSON, CSV, etc.) along with web links to the corresponding backend UI for further investigation. This design facilitates integration with LLMs, automation tools, and data processing pipelines.

The same functionality is available as a library for integration with bots, automation systems, and other tools.

## Design

### 1. Overall Architecture

- **Backend-specific implementations**: Each observability backend (NewRelic, Google Cloud, Datadog, etc.) has its own subcommand implementation that handles vendor-specific concepts and APIs.
- **Minimal abstraction**: Common functionality (output formatting, configuration management) is shared, but vendor-specific logic is encapsulated within each backend's implementation.
- **Library-first design**: Core logic is implemented as a library with CLI as a wrapper, enabling reuse in bots, automation tools, and other integrations.

### 2. Backend-specific Design

- **Vendor-specific subcommands**: Each backend has its own subcommand namespace (e.g., `newrelic`, `googlecloud`) that handles vendor-specific concepts and arguments.
- **Flexible argument handling**: Commands support both common arguments (e.g., `--attribute`, `--query`, `--range`) and vendor-specific arguments (e.g., `--entity` for NewRelic, `--project` for Google Cloud).
- **Consistent output**: All backends produce structured output (JSON, CSV) with web links, regardless of internal API differences.

### 3. Feature Categories

- **search-values**: Retrieves available values for specified attributes using wildcards or partial matches
- **top-traces**: Finds top traces by duration or other metrics with filtering capabilities
- **list-spans**: Retrieves detailed span information for specific traces
- **list-logs**: Retrieves log entries associated with specific traces

### 4. Extensibility & Maintainability

- Adding a new backend requires implementing vendor-specific commands and API integration within the new backend's namespace.
- Authentication and configuration are managed per backend through environment variables or configuration files.
- Each backend implementation can be tested independently with vendor-specific mocks.

### 5. CLI Design

- **Subcommand structure** (arguments subject to change):
  - `telemetry-glue newrelic search-values ...`
  - `telemetry-glue googlecloud top-traces ...`
  - `telemetry-glue newrelic list-spans ...`
  - etc.
- **Common arguments**: `--range`, `--output`, `--attribute`, `--query`
- **Vendor-specific arguments**: `--entity` (NewRelic), `--project` (Google Cloud), etc.

### 6. Library API Design

- Provides the same functionality as CLI commands through programmatic interfaces
- Each backend exposes its own API methods that handle vendor-specific concepts
- Return values are structured data with web links for UI navigation
- Designed for integration with automation tools, bots, and data processing pipelines

### 7. Security & Authentication

- API keys and authentication information for each backend are managed securely through environment variables or dedicated configuration files
- Each backend handles its own authentication requirements independently

### 8. Example Use Cases

- CLI-based telemetry data exploration and analysis
- Integration with Slack bots and other notification systems
- Data pipeline integration with LLMs and analytics tools
- Automated monitoring and alerting workflows

### 9. Directory Structure

The project structure is organized around backend-specific implementations while maintaining shared components:

```
telemetry-glue/
├── cmd/
│   └── telemetry-glue/         # CLI entry point
│       ├── main.go             # Main CLI application
│       ├── newrelic/           # NewRelic-specific commands
│       ├── googlecloud/        # Google Cloud-specific commands
│       └── common/             # Shared CLI utilities (flags, output formatting)
├── internal/
│   ├── backend/                # Backend-specific implementations
│   │   ├── newrelic/           # NewRelic API integration and logic
│   │   └── googlecloud/        # Google Cloud API integration and logic
│   ├── output/                 # Common output formatting (JSON, CSV, etc.)
│   └── config/                 # Configuration and authentication management
├── pkg/                        # Public API for library usage
├── test/                       # Tests and mocks
├── scripts/                    # Development and CI scripts
├── .env.example                # Example environment variables
├── go.mod
├── go.sum
├── README.md
└── design.md
```

**Directory roles:**

- `cmd/telemetry-glue/`: CLI application entry point and backend-specific command implementations
- `internal/backend/`: Backend-specific API integrations and business logic
- `internal/output/`: Shared output formatting utilities (JSON, CSV conversion, web link generation)
- `internal/config/`: Configuration loading and authentication management
- `pkg/`: Public library API for programmatic usage
- `test/`: Backend-specific tests and shared test utilities

This structure enables:

- Independent development and testing of each backend
- Shared utilities for common operations (output formatting, configuration)
- Clear separation between CLI commands and backend logic
- Easy addition of new backends without affecting existing implementations

### 10. Implementation Notes

- **Minimal shared interfaces**: Unlike traditional abstraction layers, this design minimizes shared interfaces to only essential common operations (output formatting, configuration management).
- **Backend autonomy**: Each backend implementation has full autonomy over its API integration, data structures, and vendor-specific logic.
- **Shared utilities**: Common functionality like JSON/CSV formatting, web link generation, and configuration loading is shared across backends.
- **Independent evolution**: Backends can evolve independently without affecting other implementations, allowing for vendor-specific optimizations and feature support.
