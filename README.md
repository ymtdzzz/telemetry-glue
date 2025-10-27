# telemetry-glue

## Environment Variables

The following environment variables can be configured:

### Glue Configuration

- `GLUE_SPAN_BACKEND` - Span backend type (e.g., "newrelic")
- `GLUE_LOG_BACKEND` - Log backend type (e.g., "newrelic") Note that log backend is not supported yet.

#### New Relic Configuration

- `GLUE_NEW_RELIC_API_KEY` - New Relic API key
- `GLUE_NEW_RELIC_ACCOUNT_ID` - New Relic account ID

### Analyzer Configuration

- `ANALYZER_LANGUAGE` - Analysis language ("en" or "ja")

#### Ollama Configuration

- `ANALYZER_OLLAMA_MODEL_NAME` - Ollama model name

#### Gemini Configuration

- `ANALYZER_GEMINI_MODEL_NAME` - Gemini model name
- `ANALYZER_GEMINI_API_KEY` - Gemini API key

#### VertexAI Configuration

- `ANALYZER_VERTEX_AI_MODEL_NAME` - VertexAI model name
- `ANALYZER_VERTEX_AI_PROJECT_ID` - GCP project ID
- `ANALYZER_VERTEX_AI_LOCATION` - GCP location
