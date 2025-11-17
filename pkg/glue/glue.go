package glue

import (
	"context"

	"github.com/ymtdzzz/telemetry-glue/pkg/app/config"
	"github.com/ymtdzzz/telemetry-glue/pkg/app/model"
	"github.com/ymtdzzz/telemetry-glue/pkg/glue/backend"
)

type Glue struct {
	spanBackend backend.GlueBackend
	logBackend  backend.GlueBackend
}

func NewGlue(cfg *config.GlueConfig) *Glue {
	glue := &Glue{}

	var nrBackend *backend.NewRelicBackend
	if cfg.NewRelic.HasAnyConfig() {
		nrBackend = backend.NewNewRelicBackend(&cfg.NewRelic)
	}

	if cfg.SpanBackend == config.BackendTypeNewRelic {
		glue.spanBackend = nrBackend
	}

	return glue
}

func (g *Glue) Execute(
	ctx context.Context,
	traceID string,
	spanReq *backend.SearchSpansRequest,
	logReq *backend.SearchLogsRequest,
) (*model.Telemetry, error) {
	var (
		spans model.Spans
		logs  model.Logs
		err   error
	)

	if g.spanBackend != nil {
		spans, err = g.spanBackend.SearchSpans(ctx, spanReq)
		if err != nil {
			return nil, err
		}
	}

	if g.logBackend != nil {
		logs, err = g.logBackend.SearchLogs(ctx, logReq)
		if err != nil {
			return nil, err
		}
	}

	return &model.Telemetry{
		Spans: spans,
		Logs:  logs,
	}, nil
}
