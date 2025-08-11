# telemetry-glue Suggested Commands

## Development Commands

### Build
```bash
# Build main CLI
go build ./cmd/telemetry-glue

# Build specific tools
go build -o bin/otel-test-sender cmd/otel-test-sender/main.go
go build -o bin/gcp-log-sender cmd/gcp-log-sender/main.go
```

### Testing
```bash
# Run all tests
go test ./...

# Test specific package
go test ./internal/backend/newrelic

# Test with coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Formatting & Linting
```bash
# Format code
go fmt ./...

# Linting with golangci-lint
golangci-lint run

# Linting with fix mode
golangci-lint run --fix

# Static analysis
go vet ./...

# Tidy modules
go mod tidy
```

### Execution
```bash
# Setup environment variables (development)
cp .env.example .env
# Edit .env file

# CLI help
./telemetry-glue --help
./telemetry-glue newrelic --help

# NewRelic operation examples
./telemetry-glue newrelic search-values --entity "your-app" --attribute http.path --query "*user*" --since 1h
./telemetry-glue newrelic top-traces --entity "your-app" --attribute http.path --value "/api/users" --since 1h
./telemetry-glue newrelic spans --trace-id "d8a60536187fa0927e45911f8c0dd64b" --since 1h

# Test senders
./bin/otel-test-sender
./bin/gcp-log-sender
```

## System Commands (Linux)
```bash
# File operations
ls -la
find . -name "*.go"
grep -r "pattern" .

# Git operations
git status
git diff
git log --oneline -10

# Process management
ps aux | grep telemetry
kill -9 <pid>
```

## Dependency Management
```bash
# Add dependency
go get <package>

# Update dependencies
go get -u ./...

# Remove unused dependencies
go mod tidy

# Check dependencies
go mod graph
go list -m all
```