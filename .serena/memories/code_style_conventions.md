# telemetry-glue Code Style & Conventions

## Naming Conventions
- **Package names**: lowercase, concise (e.g., `newrelic`, `gcp`, `analyzer`)
- **Function names**: camelCase, exported functions start with uppercase
- **Variable names**: camelCase, concise and meaningful names
- **Struct names**: PascalCase, exported structs start with uppercase
- **Constants**: camelCase or UPPER_SNAKE_CASE (for error constants)

## Error Handling
- Custom error constant definitions:
```go
var (
    ErrMissingAPIKey    = errors.New("NEW_RELIC_API_KEY is required")
    ErrMissingAccountID = errors.New("NEW_RELIC_ACCOUNT_ID is required")
)
```

## File Organization
- **Main functions**: Keep concise, delegate actual processing to other functions
- **init functions**: Initialize environment settings (e.g., loading .env files)
- **Struct definitions**: Clearly define structs for requests and responses

## Import Order
1. Standard library
2. Third-party libraries
3. Internal project packages

## Environment Variable Management
- Auto-load .env files in non-production environments
- Use actual environment variables in production
- Implement validation for required environment variables

## Comments
- Provide proper documentation comments for exported functions and structs
- Add explanatory comments for complex logic
- Use TODO, FIXME comments appropriately

## File Naming
- Split files by functionality (e.g., `client.go`, `spans.go`, `search_values.go`)
- Test files follow `*_test.go` format (currently no test files exist)

## Design Patterns
- **Backend-specific implementations**: Each backend has independent implementation
- **Minimal common interfaces**: Only essential commonalities are shared
- **Error handling**: Explicit error handling, avoid panics