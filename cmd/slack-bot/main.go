package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/kelseyhightower/envconfig"
	"github.com/ymtdzzz/telemetry-glue/internal/slack"
)

func main() {
	// Determine mode from environment or path
	mode := os.Getenv("SERVICE_MODE")
	if mode == "" {
		mode = "bot" // Default to bot mode
	}

	switch mode {
	case "bot":
		runBotMode()
	case "worker":
		runWorkerMode()
	default:
		log.Fatalf("Unknown service mode: %s. Use 'bot' or 'worker'", mode)
	}
}

func runBotMode() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load configuration from environment variables
	var config slack.Config
	if err := envconfig.Process("", &config); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create Cloud Tasks client
	tasksClient, err := slack.NewTasksClient(ctx, &config)
	if err != nil {
		log.Fatalf("Failed to create tasks client: %v", err)
	}
	defer tasksClient.Close()

	// Create Slack Bot handler
	handler := slack.NewHandler(&config, tasksClient)

	// Signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start bot (asynchronously)
	go func() {
		log.Println("Starting Slack Bot...")
		if err := handler.Start(ctx); err != nil {
			log.Printf("Slack Bot error: %v", err)
			cancel()
		}
	}()

	// Wait for signal
	select {
	case sig := <-sigChan:
		log.Printf("Received signal: %v", sig)
		cancel()
	case <-ctx.Done():
		log.Println("Context cancelled")
	}

	log.Println("Slack Bot shutting down...")
}

func runWorkerMode() {
	// Port configuration
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Set up HTTP handlers
	http.HandleFunc("/", analyzeHandler)
	http.HandleFunc("/analyze", analyzeHandler) // Alternative endpoint

	log.Printf("Worker function listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func analyzeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Accept only POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Load configuration from environment variables
	var config slack.Config
	if err := envconfig.Process("", &config); err != nil {
		log.Printf("Failed to load configuration: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Read request body
	var req slack.AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed to decode request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create worker analyzer
	analyzer := slack.NewAnalyzer(&config)

	// Execute analysis
	response, err := analyzer.AnalyzeTrace(ctx, &req)
	if err != nil {
		log.Printf("Analysis failed for trace_id %s: %v", req.TraceID, err)

		// Post error to Slack
		errorResponse := &slack.AnalyzeResponse{
			TraceID: req.TraceID,
			Status:  "error",
			Error:   err.Error(),
		}

		// Post error to Slack
		if postErr := analyzer.PostErrorToSlack(&req, err.Error()); postErr != nil {
			log.Printf("Failed to post error to Slack: %v", postErr)
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	// Success response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	log.Printf("Analysis completed successfully for trace_id: %s", req.TraceID)
}
