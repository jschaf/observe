package propagate_test

import (
	"net/http"
	"testing"

	"github.com/jschaf/observe/internal/difftest"
	"github.com/jschaf/observe/trace/propagate"
)

const (
	headerTraceparent = "Traceparent"
	headerTracestate  = "Tracestate"
)

func TestHTTPHeader_ExtractContext(t *testing.T) {
	tests := []struct {
		name        string
		traceparent string
		tracestate  string
		want        http.Header
	}{
		// Invalid
		{
			name:        "invalid max version",
			traceparent: "ff-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
		},
		{
			name:        "invalid not hex",
			traceparent: "01-z0z01234z0z01234z0z01234z0z01234-00f067aa0ba902b7-01",
		},
		{
			name:        "invalid flag hex",
			traceparent: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-0z",
		},
		{
			name:        "invalid missing dashes",
			traceparent: "00 4bf92f3577b34da6a3ce929d0e0e4736 00f067aa0ba902b7 01",
		},
		{
			name:        "invalid missing flags",
			traceparent: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7",
		},
		{
			name:        "invalid short trace id",
			traceparent: "00-4bf92f3577b34da6a3ce929d0e0e47-00f067aa0ba902b7-01",
		},

		// Valid
		{
			name:        "valid lowercase",
			traceparent: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
			want: http.Header{
				headerTraceparent: []string{"00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"},
			},
		},
		{
			name:        "valid uppercase",
			traceparent: "00-4BF92F3577B34DA6A3CE929D0E0E4736-00F067AA0BA902B7-01",
			want: http.Header{
				headerTraceparent: []string{"00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"},
			},
		},
		{
			name:        "valid not sampled",
			traceparent: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-00",
			want: http.Header{
				headerTraceparent: []string{"00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-00"},
			},
		},
		{
			name:        "valid future version with extra data",
			traceparent: "04-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-00-extra-data",
			want: http.Header{
				headerTraceparent: []string{"00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-00"},
			},
		},
		{
			name:        "valid future version sampled extra flags",
			traceparent: "04-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-09",
			want: http.Header{
				headerTraceparent: []string{"00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"},
			},
		},
		{
			name:        "valid future version not sampled extra flags",
			traceparent: "04-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-08",
			want: http.Header{
				headerTraceparent: []string{"00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-00"},
			},
		},
		{
			name:        "valid with state",
			traceparent: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
			tracestate:  "key1=value1,  key2=value2",
			want: http.Header{
				headerTraceparent: []string{"00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"},
				headerTracestate:  []string{"key1=value1,key2=value2"},
			},
		},
		{
			name:        "ignores invalid state",
			traceparent: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
			tracestate:  "not valid",
			want: http.Header{
				headerTraceparent: []string{"00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := http.Header{}
			if tt.traceparent != "" {
				h.Set(headerTraceparent, tt.traceparent)
			}
			if tt.tracestate != "" {
				h.Set(headerTracestate, tt.tracestate)
			}
			got := propagate.HTTPHeader(h).ExtractContext()

			gotHdr := http.Header{}
			propagate.HTTPHeader(gotHdr).InjectContext(got)
			difftest.AssertSame(t, "InjectHeader mismatch", tt.want, gotHdr)
		})
	}
}

func BenchmarkExtractHTTPHeaderContext(b *testing.B) {
	h := http.Header{}
	h.Set(headerTraceparent, "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
	h.Set(headerTracestate, "key1=value1,key2=value2")

	b.ReportAllocs()
	for b.Loop() {
		p := propagate.HTTPHeader(h)
		p.ExtractContext()
	}
}
