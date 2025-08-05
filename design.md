# telemetry-glue

## Overview

telemetry-glue is a library and CLI tool for querying telemetry data such as traces and logs without being dependent on specific observability backends like NewRelic, Datadog, or Google Cloud.

The entry point is a trace. The tool retrieves information such as logs that match a given trace ID across multiple backends.

## Usage

For example, the CLI can be used as follows:

```sh
# Retrieve a list of valid paths by partial match (wildcard fuzzy search)
telemetry-glue search values "http.path=*user*" --from new_relic --range "2025-08-01T12:00:00+09:00,2025-08-02T00:00:00+09:00"
# output =>
# available values:
#   - /admin/users/new
#   - /user/:id
#   - ...

# Query based on the path list (exact value specified)
# Output is the top 5 by duration, and also returns a link to display the search results in the Web UI, allowing the user to continue searching on the web
telemetry-glue top traces "http.path=/admin/users/new" --from new_relic --range "2025-08-01T12:00:00+09:00,2025-08-02T00:00:00+09:00"
# output =>
# result link: https://one.newrelic....
# TOP 5 (duration):
#   - 2025-08-01T12:01:12+09:00 {trace_id} -- 3.23s
#   - ... {trace_id} -- 2.8s
#   - ...

# Query detailed span information based on a specified trace ID (output: CSV, JSON, etc. and a web link to the relevant trace)
telemetry-glue list spans --trace {trace_id} --from new_relic --range "2025-08-01T12:00:00+09:00,2025-08-02T00:00:00+09:00"

# Query log information in another backend based on a specified trace ID (output: CSV, JSON, etc. and a web link to the relevant log)
telemetry-glue list logs --trace {trace_id} --from cloud_logging --range "2025-08-01T12:00:00+09:00,2025-08-02T00:00:00+09:00"
```

For subcommands under the `list` command, the output is data such as CSV or JSON, except for web links. This is intended for input to LLMs or further processing by programs. If you want to check the information in a more user-friendly format, you can access it via the provided web link.

Only CLI usage is shown as an example, but the same functionality is available as a library. By using it as a library, you can, for example, interactively explore traces via a Slack bot and send search results to an LLM for analysis.

## Design

### 1. Overall Architecture

- An interface layer abstracts each observability backend (NewRelic, Datadog, Google Cloud, etc.), allowing pluggable backend-specific implementations.
- A query integration layer receives user queries (trace ID, path, etc.), queries multiple backends in parallel, and integrates the results.
- The core logic is implemented as a library, and the CLI is a wrapper, making it reusable from other tools (such as Slack bots).

### 2. Feature-specific Design

- **search values**  
  Retrieves a list of values for a specified attribute (e.g., http.path) using wildcards or partial matches. Absorbs differences in search APIs for each backend.
- **top traces**  
  Filters by specified path, etc., and retrieves the top N traces by duration. Also attaches a Web UI link to the search results.
- **list spans / list logs**  
  Retrieves spans or logs based on a trace ID. Output format can be selected as CSV/JSON. Also attaches a web link.

### 3. Extensibility & Maintainability

- To add a new backend, simply implement the abstract interface.
- Backend authentication information and endpoints are managed via configuration files/environment variables.
- The backend layer is designed to be mockable, making unit and integration testing easy.

### 4. CLI Design

- Example command structure:
  - telemetry-glue search values ...
  - telemetry-glue top traces ...
  - telemetry-glue list spans ...
  - telemetry-glue list logs ...
- Common options:  
  --from (specify backend), --range (specify time range), --output (specify output format), --trace (specify trace ID), etc.

### 5. Library API Design

- Provides the same functionality as the CLI as functions/classes.
- Return values are structured data (lists, dictionaries, etc.) plus web links.

### 6. Security & Authentication

- API keys and authentication information for each backend are managed securely (using .env files or secret management services is recommended).

### 7. Example Use Cases

- CLI usage
- Usage from Slack bots, etc.
- Data integration with LLMs

### 8. Directory Structure

The project follows a modular and extensible directory structure, in line with Go best practices, to maximize maintainability, testability, and clarity:

```
telemetry-glue/
├── cmd/
│   └── telemetry-glue/         # CLI entry point (main.go)
├── internal/
│   ├── core/                   # Query integration layer (core logic)
│   ├── backend/                # Backend abstraction interfaces and implementations
│   │   ├── newrelic/           # NewRelic backend implementation
│   │   ├── datadog/            # Datadog backend implementation
│   │   └── gcp/                # Google Cloud backend implementation
│   ├── config/                 # Configuration and authentication management
│   └── util/                   # Utility functions and types
├── pkg/                        # Public API for use as a library
├── test/                       # Integration and end-to-end tests, test mocks
├── scripts/                    # Development/build/CI scripts
├── .env.example                # Example for authentication/configuration
├── go.mod
├── go.sum
├── README.md
└── design.md
```

**Directory roles:**
- `cmd/telemetry-glue/`: The CLI application's entry point and command routing.
- `internal/core/`: Core logic for integrating queries and aggregating results from multiple backends.
- `internal/backend/`: Abstract interfaces for observability backends and their concrete implementations (e.g., NewRelic, Datadog, GCP).
- `internal/config/`: Handles configuration loading and authentication management.
- `internal/util/`: Shared utility code.
- `pkg/`: Public API for use as a Go library (for bots, LLM integration, etc.).
- `test/`: Integration/E2E tests and test mocks.
- `scripts/`: Helper scripts for development and CI.
- `.env.example`: Example environment variables for backend authentication.

This structure ensures that:
- Adding a new backend only requires implementing the interface in `internal/backend/`.
- The CLI and library share the same core logic.
- Configuration and authentication are managed centrally.
- The codebase is easy to test and extend.


### 9. Backend Interface Design

The backend interface abstracts the interaction with each observability backend (such as NewRelic, Datadog, Google Cloud, etc.) and defines a common contract for all supported operations. This enables easy extensibility and consistent integration across different providers.

Key points:
- All time ranges are represented using Go's `time.Time` type, and ISO8601 strings are parsed at instantiation.
- Each backend must implement the following interface:

```go
import "time"

type TimeRange struct {
    Start time.Time // Start of the range
    End   time.Time // End of the range
}

type SearchValuesRequest struct {
    Attribute string    // e.g. "http.path"
    Query     string    // e.g. "*user*"
    TimeRange TimeRange
}

type SearchValuesResponse struct {
    Values  []string
    WebLink string // Link to the relevant search result in the backend UI
}

type TopTracesRequest struct {
    Attribute string    // e.g. "http.path"
    Value     string    // e.g. "/admin/users/new"
    TimeRange TimeRange
    Limit     int
}

type TopTracesResponse struct {
    Traces  []TraceSummary
    WebLink string // Link to the search result in the backend UI
}

type ListSpansRequest struct {
    TraceID   string
    TimeRange TimeRange
}

type ListSpansResponse struct {
    Spans   []Span
    WebLink string // Link to the trace in the backend UI
}

type ListLogsRequest struct {
    TraceID   string
    TimeRange TimeRange
}

type ListLogsResponse struct {
    Logs    []LogEntry
    WebLink string // Link to the logs in the backend UI
}

type TraceSummary struct {
    TraceID    string
    StartTime  time.Time
    Duration   float64 // seconds
    Attributes map[string]interface{}
}

type Span struct {
    SpanID     string
    TraceID    string
    Name       string
    StartTime  time.Time
    EndTime    time.Time
    Attributes map[string]interface{}
}

type LogEntry struct {
    Timestamp  time.Time
    TraceID    string
    SpanID     string
    Message    string
    Attributes map[string]interface{}
}

type Backend interface {
    Name() string

    SearchValues(req SearchValuesRequest) (SearchValuesResponse, error)
    TopTraces(req TopTracesRequest) (TopTracesResponse, error)
    ListSpans(req ListSpansRequest) (ListSpansResponse, error)
    ListLogs(req ListLogsRequest) (ListLogsResponse, error)
}
```

This design ensures that adding a new backend only requires implementing the `Backend` interface, and all core operations (searching values, retrieving top traces, listing spans/logs) are handled in a consistent and type-safe manner.

