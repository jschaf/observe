package log

import (
	"bytes"
	"context"
	"log/slog"
	"regexp"
	"strings"
	"testing"
	"time"
)

// textTimeRE is a regexp to match log timestamps for Text handler.
// This is RFC3339Nano with the fixed 3 digit sub-second precision.
const textTimeRE = `\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}(Z|[+-]\d{2}:\d{2})`

func TestLogTextHandler(t *testing.T) {
	ctx := context.Background()

	var buf bytes.Buffer
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))

	check := func(want string) {
		t.Helper()
		if want != "" {
			want = "time=" + textTimeRE + " " + want
		}
		checkLogOutput(t, buf.String(), want)
		buf.Reset()
	}

	Info(ctx, "msg", slog.Int("a", 1), slog.Int("b", 2))
	check(`level=INFO msg=msg a=1 b=2`)

	// By default, debug messages are not printed.
	Debug(ctx, "bg", slog.Int("a", 1), slog.Int("b", 2))
	check("")

	Warn(ctx, "w", slog.Duration("dur", 3*time.Second))
	check(`level=WARN msg=w dur=3s`)

	Error(ctx, "bad", slog.Int("a", 1))
	check(`level=ERROR msg=bad a=1`)

	Log(ctx, slog.LevelWarn+1, "w", slog.Int("a", 1), slog.String("b", "two"))
	check(`level=WARN\+1 msg=w a=1 b=two`)
}

func checkLogOutput(t *testing.T, got, wantRegexp string) {
	t.Helper()
	got = clean(got)
	wantRegexp = "^" + wantRegexp + "$"
	matched, err := regexp.MatchString(wantRegexp, got)
	if err != nil {
		t.Fatal(err)
	}
	if !matched {
		t.Errorf("\ngot  %s\nwant %s", got, wantRegexp)
	}
}

// clean prepares log output for comparison.
func clean(s string) string {
	if len(s) > 0 && s[len(s)-1] == '\n' {
		s = s[:len(s)-1]
	}
	return strings.ReplaceAll(s, "\n", "~")
}
