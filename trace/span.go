package trace

import (
	"time"

	"github.com/jschaf/observe/internal/epoch"
)

// Span is a single operation within a trace. Spans can nest to form a trace
// tree.
// https://opentelemetry.io/docs/specs/otel/trace/api/#span
type Span struct {
	tracer *Tracer     // immutable tracer that created this span
	name   string      // immutable display name of the span
	start  epoch.Nanos // immutable start time
	end    epoch.Nanos // set by first call to End; protected with atomics
}

// IsRecording returns true if the span currently records data. Returns false
// after calling [Span.End] or if the span is nil.
// https://opentelemetry.io/docs/specs/otel/trace/api/#isrecording
func (s *Span) IsRecording() bool {
	if s == nil {
		return false
	}
	return (&s.end).Load() == 0
}

// StartTime returns when the span started. If the span is nil, it returns the
// zero-value of time.Time.
func (s *Span) StartTime() time.Time {
	if s == nil {
		return time.Time{}
	}
	return s.start.ToTime()
}

// EndTime returns when the span ended. If the span is nil, it returns the
// zero-value of time.Time.
func (s *Span) EndTime() time.Time {
	if s == nil {
		return time.Time{}
	}
	return (&s.end).Load().ToTime()
}

type spanEndConfig struct {
	endTime epoch.Nanos
}

type SpanEndOption func(spanEndConfig) spanEndConfig

// WithEndTime sets the end time of the span.
func WithEndTime(t time.Time) SpanEndOption {
	return func(cfg spanEndConfig) spanEndConfig {
		cfg.endTime = epoch.NewNanos(t)
		return cfg
	}
}

// End signals the span has ended.
// https://opentelemetry.io/docs/specs/otel/trace/api/#end
func (s *Span) End(opts ...SpanEndOption) {
	if s == nil {
		return
	}
	cfg := spanEndConfig{}
	for _, opt := range opts {
		cfg = opt(cfg)
	}
	if cfg.endTime == 0 {
		cfg.endTime = epoch.NanosNow()
	}

	// The first call sets the end time. Ignore all later calls.
	if !(&s.end).SwapIfZero(cfg.endTime) {
		return
	}
}
