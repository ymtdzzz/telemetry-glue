package slack_fn

import (
	"log"
	"net/http"
	"os"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/ymtdzzz/telemetry-glue/internal/slackfn"
)

func init() {
	// Register Slack event handler
	functions.HTTP("SlackEvent", slackfn.HandleSlackEvent)

	// Register trace analysis handler
	functions.HTTP("AnalyzeTrace", slackfn.HandleAnalyzeTrace)
}

// main is only used for local development
func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
