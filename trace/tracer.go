package trace

import (
	"context"
	"time"

	"github.com/jschaf/observe/internal/epoch"
)

type Tracer struct{}

type startConfig struct {
	startTime epoch.Nanos
}

type SpanStartOption func(startConfig) startConfig

func WithStartTime(t time.Time) SpanStartOption {
	return func(cfg startConfig) startConfig {
		cfg.startTime = epoch.NewNanos(t)
		return cfg
	}
}

// Start starts a Span and returns a new context containing the Span.
func (t *Tracer) Start(ctx context.Context, name string, opts ...SpanStartOption) (context.Context, Span) {
	cfg := startConfig{}
	for _, opt := range opts {
		cfg = opt(cfg)
	}
	if cfg.startTime == 0 {
		cfg.startTime = epoch.NanosNow()
	}

	span := Span{
		name:      name,
		tracer:    t,
		start:     cfg.startTime,
		lifecycle: newLifecycle(),
	}

	return ctx, span
}
