package main

import (
	"log"
	"net/http"
	"os"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

func init() {
	// Register Slack event handler
	functions.HTTP("SlackEvent", handleSlackEvent)

	// Register trace analysis handler
	functions.HTTP("AnalyzeTrace", handleAnalyzeTrace)
}

func main() {
	// Use PORT environment variable, or default to 8080.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Slack Functions listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
