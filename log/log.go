package log

import (
	"context"
	"log/slog"
	"runtime"
	"time"
)

// Debug logs at [slog.LevelDebug].
func Debug(ctx context.Context, msg string, attrs ...slog.Attr) {
	log(ctx, slog.LevelDebug, msg, attrs)
}

// Info logs at [slog.LevelInfo].
func Info(ctx context.Context, msg string, attrs ...slog.Attr) {
	log(ctx, slog.LevelInfo, msg, attrs)
}

// Warn logs at [slog.LevelWarn].
func Warn(ctx context.Context, msg string, attrs ...slog.Attr) {
	log(ctx, slog.LevelWarn, msg, attrs)
}

// Error logs at [slog.LevelError].
func Error(ctx context.Context, msg string, attrs ...slog.Attr) {
	log(ctx, slog.LevelError, msg, attrs)
}

// Log emits a log record with the current time and the given level and message.
func Log(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	log(ctx, level, msg, attrs)
}

func log(ctx context.Context, level slog.Level, msg string, attrs []slog.Attr) {
	l := slog.Default()
	if !l.Enabled(ctx, level) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(3, pcs[:]) // skip [Callers, log, caller]
	r := slog.NewRecord(time.Now(), level, msg, pcs[0])
	r.AddAttrs(attrs...)
	_ = l.Handler().Handle(ctx, r)
}
