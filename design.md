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
