package trace_test

import (
	"sync"
	"testing"
	"time"

	"github.com/jschaf/observe/internal/difftest"
	"github.com/jschaf/observe/trace"
)

func TestSpan_End(t *testing.T) {
	t.Run("WithEndTime", func(t *testing.T) {
		want := time.Now().Add(2 * time.Minute)
		span := startTestSpan(t)
		span.End(trace.WithEndTime(want))
		got := span.EndTime()
		difftest.AssertSame(t, "EndTime mismatch", want, got)
	})

	t.Run("IsRecording", func(t *testing.T) {
		tr := &trace.Tracer{}
		_, span := tr.Start(t.Context(), "test-span")
		if !span.IsRecording() {
			t.Errorf("Span.IsRecording() should return true before End() is called; got false")
		}
		span.End()
		if span.IsRecording() {
			t.Errorf("Span.IsRecording() should return false after End() is called; got true")
		}
	})
}

func TestSpan_End_Race(t *testing.T) {
	span := startTestSpan(t)

	var wg sync.WaitGroup

	// Readers
	readerCount := 50
	wg.Add(readerCount)
	for range readerCount {
		go func() {
			defer wg.Done()
			for range 5 {
				_ = span.IsRecording()
				_ = span.EndTime()
				time.Sleep(1 * time.Millisecond) // delay to allow interleaving
			}
		}()
	}

	// Writers
	writerCount := 50
	wg.Add(writerCount)
	for range writerCount {
		go func() {
			defer wg.Done()
			span.End(trace.WithEndTime(time.Now()))
			span.End()
		}()
	}

	wg.Wait()

	if span.EndTime().IsZero() || span.EndTime().UnixNano() == 0 {
		t.Errorf("Span end time should not be zero after End() was called")
	}
	if span.IsRecording() {
		t.Errorf("Span.IsRecording() should return false after End() was called")
	}
}

func BenchmarkStartEndSpan(b *testing.B) {
	tr := &trace.Tracer{}
	ctx := b.Context()
	b.ReportAllocs()
	for b.Loop() {
		_, span := tr.Start(ctx, "test-span")
		span.End()
	}
}
