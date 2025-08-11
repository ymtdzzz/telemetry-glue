package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/kelseyhightower/envconfig"
	slackInternal "github.com/ymtdzzz/telemetry-glue/internal/slack"
)

func handleAnalyzeTrace(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Accept only POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Load configuration from environment variables
	var config slackInternal.Config
	if err := envconfig.Process("", &config); err != nil {
		log.Printf("Failed to load configuration: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Read request body
	var req slackInternal.AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed to decode request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create worker analyzer
	analyzer := slackInternal.NewAnalyzer(&config)

	// Execute analysis
	response, err := analyzer.AnalyzeTrace(ctx, &req)
	if err != nil {
		log.Printf("Analysis failed for trace_id %s: %v", req.TraceID, err)

		// Post error to Slack
		errorResponse := &slackInternal.AnalyzeResponse{
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
		if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
			log.Printf("Failed to encode error response: %v", err)
		}
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
