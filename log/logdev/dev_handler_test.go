package log

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/jschaf/observe/internal/difftest"
	"github.com/jschaf/observe/internal/tty"
)

func TestDevHandler_Sample(t *testing.T) {
	l := slog.New(NewDevHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	l.Info("info level message", slog.String("attribute1", "value1"), slog.Int("attribute2", 42))
	l.Debug("debug level message", slog.String("attribute3", "value3"))
	l.Warn("warning level with multiple attrs", slog.Float64("pi", 3.14), slog.Bool("isTest", true))
	l.Error("error level message", slog.Any("details", map[string]any{"key": "value"}))
	l.Info("info level message that is really, long", slog.String("attribute1", "value1"), slog.Int("attribute2", 42))
}

func TestDevHandler_Handle(t *testing.T) {
	ctx := t.Context()
	buf := &bytes.Buffer{}
	h := &DevHandler{w: buf}

	err := h.Handle(ctx, slog.Record{
		Time:    time.Date(2024, time.January, 1, 12, 0, 0, 0, time.UTC),
		Message: "msg",
	})
	if err != nil {
		t.Fatalf("handle record: %v", err)
	}
	got := buf.String()

	want := fmt.Sprintf("12:00:00.000\t%s\tmsg\n", tty.Blue.Add("info"))
	difftest.AssertSame(t, "DevHandler mismatch", want, got)
}

func BenchmarkDevHandler_Handle(b *testing.B) {
	ctx := b.Context()
	buf := &bytes.Buffer{}
	h := &DevHandler{w: buf}

	r := slog.Record{
		Time:    time.Date(2024, time.January, 1, 12, 0, 0, 0, time.UTC),
		Message: "msg",
	}
	r.AddAttrs(slog.String("foo", "bar"))
	r.AddAttrs(slog.Int64("int", 64))
	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		buf.Reset() // reset the buffer to avoid accumulation of data
		err := h.Handle(ctx, r)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}
