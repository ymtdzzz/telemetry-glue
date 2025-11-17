package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/araddon/dateparse"
	"github.com/spf13/cobra"
	"github.com/ymtdzzz/telemetry-glue/pkg/app"
	"github.com/ymtdzzz/telemetry-glue/pkg/app/logger"
	"github.com/ymtdzzz/telemetry-glue/pkg/glue/backend"
)

// flags holds flags for analyze command
type flags struct {
	analysisType string
	configPath   string
	queryOnly    bool
	startTime    string
	duration     time.Duration
}

// analyzeCmd creates the analyze subcommand
func analyzeCmd() *cobra.Command {
	flags := &flags{}

	cmd := &cobra.Command{
		Use:   "analyze <trace-id>",
		Short: "Analyze telemetry data using LLM",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAnalyze(flags, args)
		},
	}

	cmd.Flags().StringVarP(&flags.analysisType, "type", "t", "", "[required] Analysis type (duration, error)")
	cmd.Flags().StringVarP(&flags.configPath, "config", "c", "", "[required] Config path")
	cmd.Flags().StringVarP(&flags.startTime, "start-time", "s", "", "[required] Start time for telemetry data (e.g., '2025-01-12 12:00:00)")
	cmd.Flags().DurationVarP(&flags.duration, "duration", "d", 30*time.Minute, "[required] Duration from start time for telemetry data")
	cmd.Flags().BoolVarP(&flags.queryOnly, "query-only", "q", false, "Only display the fetched telemetry without executing LLM analysis")

	if err := cmd.MarkFlagRequired("type"); err != nil {
		panic(fmt.Sprintf("Failed to mark type flag as required: %v", err))
	}
	if err := cmd.MarkFlagRequired("config"); err != nil {
		panic(fmt.Sprintf("Failed to mark config flag as required: %v", err))
	}
	if err := cmd.MarkFlagRequired("start-time"); err != nil {
		panic(fmt.Sprintf("Failed to mark start-time flag as required: %v", err))
	}
	if err := cmd.MarkFlagRequired("duration"); err != nil {
		panic(fmt.Sprintf("Failed to mark duration flag as required: %v", err))
	}

	return cmd
}

func runAnalyze(flags *flags, args []string) error {
	if flags.analysisType != "duration" && flags.analysisType != "error" {
		return fmt.Errorf("unsupported analysis type: %s (supported: duration, error)", flags.analysisType)
	}
	startTime, err := dateparse.ParseAny(flags.startTime)
	if err != nil {
		return fmt.Errorf("failed to parse start time: %w", err)
	}
	endTime := startTime.Add(flags.duration)

	traceID := args[0]

	l := logger.NewStdoutLogger()

	app, err := app.NewApp(flags.configPath, l, traceID, &backend.TimeRange{
		Start: startTime,
		End:   endTime,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize app: %w", err)
	}

	ctx := context.Background()

	switch flags.analysisType {
	case "duration":
		return app.RunDuration(ctx, flags.queryOnly)
	case "error":
		return errors.New("error analysis is not yet implemented")
	}

	return nil
}
