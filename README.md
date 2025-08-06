# telemetry-glue

A unified interface for querying telemetry data such as traces and logs across multiple observability backends like NewRelic, Datadog, and Google Cloud.

## Features

- **Unified API**: Query different observability backends using the same interface
- **CLI & Library**: Use as a command-line tool or Go library
- **Multiple Backends**: Support for NewRelic (with more backends coming)
- **Search Values**: Find unique attribute values with wildcard support
- **Future Features**: Top traces, span lists, log queries (coming soon)

## Setup

### 1. Environment Variables

Copy the example environment file and set your configuration:

```bash
cp .env.example .env
```

Edit `.env` with your actual values:

```bash
# New Relic API configuration
NEW_RELIC_API_KEY=your_api_key_here
NEW_RELIC_ACCOUNT_ID=your_account_id_here

# Environment (production/development/staging)
ENV=development
```

**Note**: In production environments, set these as actual environment variables instead of using a `.env` file.

### 2. Getting New Relic Credentials

1. **API Key**: Go to [New Relic API Keys](https://one.newrelic.com/launcher/api-keys-ui.api-keys-launcher) and create a User key
2. **Account ID**: Found in your New Relic URL: `https://one.newrelic.com/accounts/YOUR_ACCOUNT_ID/...`

## Usage

### CLI

```bash
# Build the CLI
go build ./cmd/telemetry-glue

# Search for attribute values
./telemetry-glue search-values --backend newrelic --attribute http.path --query "*user*" --since 1h

# Get help
./telemetry-glue --help
./telemetry-glue search-values --help
```

### Library

```go
package main

import (
    "fmt"
    "time"

    "github.com/ymtdzzz/telemetry-glue/pkg/telemetryglue"
    "github.com/ymtdzzz/telemetry-glue/internal/backend"
)

func main() {
    // Create client
    client, err := telemetryglue.NewClient("newrelic")
    if err != nil {
        panic(err)
    }

    // Search for values
    resp, err := client.QuickSearchValues(backend.SearchValuesRequest{
        Attribute: "http.path",
        Query:     "*user*",
        TimeRange: backend.TimeRange{
            Start: time.Now().Add(-1 * time.Hour),
            End:   time.Now(),
        },
    })
    if err != nil {
        panic(err)
    }

    fmt.Printf("Found %d values: %v\n", len(resp.Values), resp.Values)
    fmt.Printf("View in UI: %s\n", resp.WebLink)
}
```

## Development

### Project Structure

```
telemetry-glue/
├── cmd/telemetry-glue/         # CLI entry point
├── internal/core/              # Query integration layer (core logic)
├── internal/backend/           # Backend abstraction interfaces/implementations
├── pkg/telemetryglue/          # Public API for use as a library
├── .env.example                # Example environment variables
└── README.md
```

### Adding New Backends

1. Create a new directory under `internal/backend/`
2. Implement the `Backend` interface defined in `internal/backend/interface.go`
3. Add the backend to the factory functions

### Security

- Never commit `.env` files to version control
- Use environment variables in production
- Store API keys securely using secret management services

## License

[Add your license here]