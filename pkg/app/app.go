package app

import (
	"context"
	"fmt"

	"github.com/ymtdzzz/telemetry-glue/pkg/analyzer"
	"github.com/ymtdzzz/telemetry-glue/pkg/app/config"
	"github.com/ymtdzzz/telemetry-glue/pkg/app/logger"
	"github.com/ymtdzzz/telemetry-glue/pkg/app/model"
	"github.com/ymtdzzz/telemetry-glue/pkg/glue"
	"github.com/ymtdzzz/telemetry-glue/pkg/glue/backend"
)

// App struct that holds the application configuration, analyzer, and glue components
type App struct {
	config    *config.AppConfig
	logger    logger.Loggable
	analyzer  *analyzer.Analyzer
	glue      *glue.Glue
	traceID   string
	timeRange *backend.TimeRange
}

// NewApp creates a new App instance with the provided configuration
func NewApp(
	cfgPath string,
	logger logger.Loggable,
	traceID string,
	timeRange *backend.TimeRange,
) (*App, error) {
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		return nil, err
	}

	analyzer, err := analyzer.NewAnalyzer(&cfg.Analyzer)
	if err != nil {
		return nil, err
	}

	glue := glue.NewGlue(&cfg.Glue)

	return &App{
		config:    cfg,
		analyzer:  analyzer,
		glue:      glue,
		logger:    logger,
		traceID:   traceID,
		timeRange: timeRange,
	}, nil
}

func (a *App) RunDuration(ctx context.Context, queryOnly bool) error {
	telemetry, err := a.executeGlue(ctx)
	if err != nil {
		return err
	}

	if queryOnly {
		a.logger.Log("Query-only mode enabled; skipping analysis.")
		return nil
	}

	if len(telemetry.Spans) == 0 && len(telemetry.Logs) == 0 {
		a.logger.Log("No telemetry data found; skipping analysis.")
		return nil
	}

	report, err := a.analyzer.AnalyzeDuration(ctx, telemetry)
	if err != nil {
		a.logger.Log("Error during analysis: " + err.Error())
		return err
	}

	a.logger.Log("Generated Duration Analysis Report:")
	a.logger.Log(report)

	return nil
}

func (a *App) executeGlue(ctx context.Context) (*model.Telemetry, error) {
	a.logger.Log("Executing glue to fetch telemetry data...")

	spanReq := &backend.SearchSpansRequest{
		TraceID:   a.traceID,
		TimeRange: a.timeRange,
	}
	logReq := &backend.SearchLogsRequest{
		TraceID:   a.traceID,
		TimeRange: a.timeRange,
	}

	telemetry, err := a.glue.Execute(ctx, a.traceID, spanReq, logReq)
	if err != nil {
		a.logger.Log("Error executing glue: " + err.Error())
		return nil, err
	}
	tokenCount, err := telemetry.RoughTokenEstimate()
	if err != nil {
		a.logger.Log("Error estimating token count: " + err.Error())
		return nil, err
	}
	a.logger.Log(fmt.Sprintf("Fetched %d spans and %d logs! Roughly estimated token count: %d", len(telemetry.Spans), len(telemetry.Logs), tokenCount))

	return telemetry, nil
}
