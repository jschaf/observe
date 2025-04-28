package trace_test

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/jschaf/observe/internal/difftest"
	"github.com/jschaf/observe/trace"
)

func TestTracer_Start_Race(t *testing.T) {
	tracer := &trace.Tracer{}
	ctx := t.Context()

	var wg sync.WaitGroup
	count := 100
	wg.Add(count)
	for i := range count {
		go func() {
			defer wg.Done()
			_, span := tracer.Start(ctx, "test-span-"+strconv.Itoa(i))
			time.Sleep(time.Duration(i%5) * time.Millisecond)
			span.End()
		}()
	}

	wg.Wait()
}

func TestTracer_Start(t *testing.T) {
	t.Run("WithStartTime", func(t *testing.T) {
		want := time.Now().Add(-2 * time.Minute)
		span := startTestSpan(t, trace.WithStartTime(want))
		span.End()
		got := span.StartTime()
		difftest.AssertSame(t, "StartTime mismatch", want, got)
	})
}

func startTestSpan(t *testing.T, opts ...trace.SpanStartOption) *trace.Span {
	t.Helper()
	tr := &trace.Tracer{}
	ctx := t.Context()
	_, span := tr.Start(ctx, "test-span: "+t.Name(), opts...)
	return span
}
