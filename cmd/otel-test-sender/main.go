package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	serviceName    = "test-service"
	serviceVersion = "1.0.0"
)

func main() {
	ctx := context.Background()

	// Load .env file (only when ENV is not "production")
	if env := os.Getenv("ENV"); env != "production" {
		if err := godotenv.Load(); err != nil {
			log.Printf("Warning: Could not load .env file: %v", err)
		}
	}

	// Get NewRelic configuration from environment variables
	apiKey := os.Getenv("NEW_RELIC_API_KEY")
	if apiKey == "" {
		log.Fatal("NEW_RELIC_API_KEY environment variable is required")
	}

	endpoint := os.Getenv("NEW_RELIC_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "otlp.nr-data.net:4318"
	}

	// Initialize OpenTelemetry
	tp, err := initTracer(ctx, apiKey, endpoint)
	if err != nil {
		log.Fatalf("Failed to initialize tracer: %v", err)
	}
	defer func() {
		// Nothing to do here - shutdown is handled in main()
	}()

	tracer := otel.Tracer("test-tracer")

	// Send test traces to NewRelic
	fmt.Println("Sending test traces to NewRelic...")

	for i := 0; i < 3; i++ {
		if err := sendTestTrace(ctx, tracer, i+1); err != nil {
			log.Printf("Failed to send trace %d: %v", i+1, err)
		} else {
			fmt.Printf("Successfully sent test trace %d\n", i+1)
		}
		time.Sleep(1 * time.Second)
	}

	// Wait for TracerProvider shutdown
	fmt.Println("Flushing traces...")
	if err := tp.Shutdown(ctx); err != nil {
		log.Printf("Error shutting down tracer provider: %v", err)
	} else {
		fmt.Println("All test traces sent. Check NewRelic dashboard for results.")
	}
}

func initTracer(ctx context.Context, apiKey, endpoint string) (*sdktrace.TracerProvider, error) {
	headers := map[string]string{
		"Api-Key": apiKey,
	}

	// Configure OTLP HTTP exporter
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithHeaders(headers),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
	}

	// Configure resource
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(serviceVersion),
			attribute.String("environment", "test"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Configure TracerProvider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)

	return tp, nil
}

func sendTestTrace(ctx context.Context, tracer trace.Tracer, traceID int) error {
	// Parent span: HTTP request simulation
	ctx, parentSpan := tracer.Start(ctx, fmt.Sprintf("http_request_%d", traceID))
	defer parentSpan.End()

	parentSpan.SetAttributes(
		attribute.String("http.method", "GET"),
		attribute.String("http.url", fmt.Sprintf("https://api.example.com/users/%d", traceID)),
		attribute.Int("http.status_code", 200),
		attribute.String("user.id", fmt.Sprintf("user_%d", traceID)),
	)

	// Child span 1: Database query
	_, dbSpan := tracer.Start(ctx, "database_query")
	dbSpan.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.name", "userdb"),
		attribute.String("db.statement", fmt.Sprintf("SELECT * FROM users WHERE id = %d", traceID)),
		attribute.String("db.operation", "SELECT"),
	)
	simulateWork(10, 50) // 10-50ms (random)
	dbSpan.End()

	// Child span 2: External API call
	_, apiSpan := tracer.Start(ctx, "external_api_call")
	apiSpan.SetAttributes(
		attribute.String("http.method", "POST"),
		attribute.String("http.url", "https://external-service.com/validate"),
		attribute.Int("http.status_code", 200),
		attribute.String("service.name", "validation-service"),
	)
	simulateWork(20, 100) // 20-100ms (random)
	apiSpan.End()

	// Child span 3: Business logic processing
	_, processSpan := tracer.Start(ctx, "business_logic_processing")
	processSpan.SetAttributes(
		attribute.String("operation", "user_data_enrichment"),
		attribute.Int("processed_records", rand.Intn(10)+1),
		attribute.Bool("cache_hit", rand.Float64() > 0.5),
	)
	simulateWork(5, 30) // 5-30ms (random)
	processSpan.End()

	return nil
}

func simulateWork(minMs, maxMs int) {
	duration := time.Duration(rand.Intn(maxMs-minMs)+minMs) * time.Millisecond
	time.Sleep(duration)
}
