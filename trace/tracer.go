package trace

import (
	"context"
	"time"

	"github.com/jschaf/observe/internal/epoch"
)

type Tracer struct{}

type startConfig struct {
	start epoch.Nanos
}

type SpanStartOption func(*startConfig)

func WithStartTime(t time.Time) SpanStartOption {
	return func(cfg *startConfig) { cfg.start = epoch.NewNanos(t) }
}

// Start starts a Span and returns a new context containing the Span.
func (t *Tracer) Start(ctx context.Context, name string, opts ...SpanStartOption) (context.Context, *Span) {
	cfg := startConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}
	if cfg.start == 0 {
		cfg.start = epoch.Now()
	}

	span := &Span{
		name:   name,
		tracer: t,
		start:  cfg.start,
	}

	return ctx, span
}
