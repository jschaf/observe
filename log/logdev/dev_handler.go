package log

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/jschaf/observe/internal/humanize"
	"github.com/jschaf/observe/internal/tty"
)

type DevHandler struct {
	w    io.Writer
	opts slog.HandlerOptions
}

func NewDevHandler(w io.Writer, opts *slog.HandlerOptions) *DevHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	return &DevHandler{w: w, opts: *opts}
}

func (h *DevHandler) Enabled(_ context.Context, l slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}
	return l >= minLevel
}

const (
	align       = 40 // make logs easier to scan by aligning the first attr
	alignStr    = "                                        "
	readyPrefix = "ready: "
)

func (h *DevHandler) Handle(_ context.Context, r slog.Record) error {
	buf := NewBuffer()
	defer buf.Free()

	// Time
	appendTime(buf, r.Time)

	// Level
	_ = buf.WriteByte('\t')
	appendLevel(buf, r)

	// Message
	r.Message = strings.TrimPrefix(r.Message, readyPrefix)
	_ = buf.WriteByte('\t')
	_, _ = buf.WriteString(r.Message)

	// Attrs
	if r.NumAttrs() > 0 {
		padCount := max(align-len(r.Message), 2)
		pad := alignStr[:padCount]
		_, _ = buf.WriteString(pad)
		r.Attrs(func(attr slog.Attr) bool {
			appendAttr(buf, attr)
			return true
		})
	}

	// Newline
	_ = buf.WriteByte('\n')

	_, err := h.w.Write(*buf)
	if err != nil {
		return fmt.Errorf("write record: %w", err)
	}
	return nil
}

func (h *DevHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	panic("not implemented")
}

func (h *DevHandler) WithGroup(_ string) slog.Handler {
	panic("not implemented")
}

func appendTime(buf *Buffer, t time.Time) {
	h, m, s := t.Clock()

	// Hours
	_ = buf.WriteByte('0' + byte(h/10))
	_ = buf.WriteByte('0' + byte(h%10))

	// Minutes
	_ = buf.WriteByte(':')
	_ = buf.WriteByte('0' + byte(m/10))
	_ = buf.WriteByte('0' + byte(m%10))

	// Seconds
	_ = buf.WriteByte(':')
	_ = buf.WriteByte('0' + byte(s/10))
	_ = buf.WriteByte('0' + byte(s%10))

	// Milliseconds
	_ = buf.WriteByte('.')
	ms := t.Nanosecond() / 1e6
	lo := ms % 10
	ms /= 10
	mid := ms % 10
	ms /= 10
	hi := ms
	_ = buf.WriteByte('0' + byte(hi))
	_ = buf.WriteByte('0' + byte(mid))
	_ = buf.WriteByte('0' + byte(lo))
}

func appendLevel(buf *Buffer, r slog.Record) {
	switch {
	case r.Level < slog.LevelInfo:
		_, _ = buf.WriteString(tty.Magenta.Code())
		_, _ = buf.WriteString("debug")
		_, _ = buf.WriteString(tty.Reset.Code())
	case r.Level < slog.LevelWarn:
		if strings.HasPrefix(r.Message, readyPrefix) {
			_, _ = buf.WriteString(tty.Green.Code())
			_, _ = buf.WriteString("ready")
		} else {
			_, _ = buf.WriteString(tty.Blue.Code())
			_, _ = buf.WriteString("info")
		}
		_, _ = buf.WriteString(tty.Reset.Code())
	case r.Level < slog.LevelError:
		_, _ = buf.WriteString(tty.Yellow.Code())
		_, _ = buf.WriteString("warn")
		_, _ = buf.WriteString(tty.Reset.Code())
	default:
		_, _ = buf.WriteString(tty.Red.Code())
		_, _ = buf.WriteString("error")
		_, _ = buf.WriteString(tty.Reset.Code())
	}
}

func appendAttr(buf *Buffer, attr slog.Attr) {
	_ = buf.WriteByte(' ')
	switch attr.Key {
	case "url":
		appendValue(buf, attr.Value)
	default:
		_, _ = buf.WriteString(attr.Key)
		_ = buf.WriteByte('=')
		appendValue(buf, attr.Value)
	}
}

func appendValue(buf *Buffer, v slog.Value) {
	switch v.Kind() {
	case slog.KindString:
		_, _ = buf.WriteString(v.String())
	case slog.KindInt64:
		*buf = strconv.AppendInt(*buf, v.Int64(), 10)
	case slog.KindUint64:
		*buf = strconv.AppendUint(*buf, v.Uint64(), 10)
	case slog.KindFloat64:
		*buf = strconv.AppendFloat(*buf, v.Float64(), 'g', -1, 64)
	case slog.KindBool:
		*buf = strconv.AppendBool(*buf, v.Bool())
	case slog.KindDuration:
		_, _ = buf.WriteString(humanize.Duration(v.Duration()))
	case slog.KindTime:
		*buf = appendRFC3339Millis(*buf, v.Time())
	case slog.KindAny:
		a := v.Any()
		switch a := a.(type) {
		case error:
			_, _ = buf.WriteString(a.Error())
		default:
			_, _ = fmt.Fprint(buf, a)
		}
	default:
		panic(fmt.Sprintf("bad kind: %s", v.Kind()))
	}
}

func appendRFC3339Millis(b []byte, t time.Time) []byte {
	// Format according to time.RFC3339Nano since it is highly optimized,
	// but truncate it to use millisecond resolution.
	// Unfortunately, that format trims trailing 0s, so add 1/10 millisecond
	// to guarantee that there are exactly 4 digits after the period.
	const prefixLen = len("2006-01-02T15:04:05.000")
	n := len(b)
	t = t.Truncate(time.Millisecond).Add(time.Millisecond / 10)
	b = t.AppendFormat(b, time.RFC3339Nano)
	b = append(b[:n+prefixLen], b[n+prefixLen+1:]...) // drop the 4th digit
	return b
}
