package trace

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/jschaf/observe/internal/hextbl"
)

func TestTraceID_String(t *testing.T) {
	tests := []struct {
		name string
		id   TraceID
		want string
	}{
		{name: "zero", id: newTraceID(0, 0), want: strings.Repeat("0", 32)},
		{name: "max", id: newTraceID(math.MaxUint64, math.MaxUint64), want: strings.Repeat("f", 32)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.id.String()
			if got != tt.want {
				t.Errorf("TraceID.String() = %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("random", func(t *testing.T) {
		for range 1000 {
			id := genTraceID()
			if !id.IsValid() {
				continue
			}
			got := id.String()
			if len(got) != 32 {
				t.Errorf("TraceID.String() = %v, want length 32", got)
			}
			if _, err := hex.DecodeString(got); err != nil {
				t.Errorf("TraceID.String() = %v, want valid hex string", got)
			}
		}
	})
}

func TestParseTraceID_Roundtrip(t *testing.T) {
	tests := []struct {
		name  string
		want  TraceID
		input string
	}{
		{
			name:  "Valid Full TraceID",
			input: "0123456789abcdef1ed2ba9876543210",
			want:  newTraceID(0x0123456789abcdef, 0x1ed2ba9876543210),
		},
		{
			name:  "Hi only",
			input: "dead1234cafe00010000000000000000",
			want:  newTraceID(0xdead1234cafe0001, 0),
		},
		{
			name:  "Lo only",
			input: "0000000000000000dead1234cafe0001",
			want:  newTraceID(0, 0xdead1234cafe0001),
		},
		{
			name:  "Leading zeros in hi",
			input: "0000000000000001fe3cba9876543210",
			want:  newTraceID(0x0000000000000001, 0xfe3cba9876543210),
		},
		{
			name:  "Leading zeros in lo",
			input: "fe3cba98765432100000000000000001",
			want:  newTraceID(0xfe3cba9876543210, 0x0000000000000001),
		},
		{
			name:  "Max values",
			input: strings.Repeat("f", 32),
			want:  newTraceID(math.MaxUint64, math.MaxUint64),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.want.String()
			if got != strings.ToLower(tc.input) {
				t.Errorf("TraceID.String():\n  Input TID: %+v\n  Expected: %q\n  Got:      %q", tc.want, tc.input, got)
			}

			parsed, err := ParseTraceID(tc.input)
			if err != nil {
				t.Fatalf("ParseTraceID(%q): Unexpected error: %v", tc.input, err)
			}

			if parsed != tc.want {
				t.Errorf("ParseTraceID result mismatch:\n  Input Str: %q\n  Expected TID: %+v\n  Got TID:      %+v", tc.input, tc.want, parsed)
			}

			roundTrip := parsed.String()
			if roundTrip != strings.ToLower(tc.input) {
				t.Errorf("Round trip String(ParseTraceID(%q)) failed:\n  Expected: %q\n  Got:      %q", tc.input, tc.input, roundTrip)
			}
		})
	}
}

func TestParseTraceID_Error(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{name: "Empty String", want: ""},
		{name: "Short String", want: "0123456789abcdef"},
		{name: "Long String", want: "0123456789abcdef1ed2ba9876543210abc"},
		{name: "Invalid Char (hi)", want: "0123456789abcdeX1ed3ba9876543210"},
		{name: "Invalid Char (lo)", want: "0123456789abcdef1ed3ba987654321X"},
		{name: "Invalid Char (space)", want: "0123456789abcdef 1ed3ba9876543210"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseTraceID(tc.want)
			if err == nil {
				t.Errorf("ParseTraceID(%q): Expected error, but got nil", tc.want)
			}
		})
	}
}

func TestSpanID_String(t *testing.T) {
	tests := []struct {
		name string
		id   SpanID
		want string
	}{
		{name: "zero", id: SpanID{}, want: "0000000000000000"},
		{name: "rand", id: SpanID{n: 0xcafeDead12345678}, want: "cafeDead12345678"},
		{name: "max", id: SpanID{n: math.MaxUint64}, want: strings.Repeat("f", 16)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.id.String()
			if got != strings.ToLower(tt.want) {
				t.Errorf("SpanID.String() = %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("random", func(t *testing.T) {
		for range 1000 {
			id := genSpanID()
			if !id.IsValid() {
				continue
			}
			got := id.String()
			if len(got) != 16 {
				t.Errorf("SpanID.String() = %v, want length 16", got)
			}
			if _, err := hex.DecodeString(got); err != nil {
				t.Errorf("TraceID.String() = %v, want valid hex string", got)
			}
		}
	})
}

func TestParseSpanID_Roundtrip(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  SpanID
	}{
		{
			name:  "Valid Full TraceID",
			input: "0123456789abcdef",
			want:  SpanID{n: 0x0123456789abcdef},
		},
		{
			name:  "Leading zeros",
			input: "0000000000000001",
			want:  SpanID{n: 0x0000000000000001},
		},
		{
			name:  "Max value",
			input: strings.Repeat("f", 16),
			want:  SpanID{n: math.MaxUint64},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.want.String()
			if got != strings.ToLower(tc.input) {
				t.Errorf("SpanID.String():\n  Input SID: %+v\n  Expected: %q\n  Got:      %q", tc.want, tc.input, got)
			}

			parsed, err := ParseSpanID(tc.input)
			if err != nil {
				t.Fatalf("ParseSpanID(%q): Unexpected error: %v", tc.input, err)
			}

			if parsed != tc.want {
				t.Errorf("ParseTraceID result mismatch:\n  Input Str: %q\n  Expected SID: %+v\n  Got SID:      %+v", tc.input, tc.want, parsed)
			}

			roundTrip := parsed.String()
			if roundTrip != strings.ToLower(tc.input) {
				t.Errorf("Round trip String(ParseSpanID(%q)) failed:\n  Expected: %q\n  Got:      %q", tc.input, tc.input, roundTrip)
			}
		})
	}
}

func TestParseSpanID_Error(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{name: "Empty String", want: ""},
		{name: "Short String", want: "0123456789abc"},
		{name: "Long String", want: "0123456789abcdef123"},
		{name: "Invalid Char (hi)", want: "0123456789abcdeX"},
		{name: "Invalid Char (space)", want: "012345678 abcdef"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseSpanID(tc.want)
			if err == nil {
				t.Errorf("ParseSpanID(%q): Expected error, but got nil", tc.want)
			}
		})
	}
}

// Benchmarks as of 2025-04-19:
//
//	BenchmarkTraceID_String/Trace.String	16.44 ns/op  32 B/op  1 allocs/op
//	BenchmarkTraceID_String/hex_encode		35.09 ns/op  64 B/op  2 allocs/op
//	BenchmarkTraceID_String/fmt_sprintf		113.1 ns/op  48 B/op  3 allocs/op
func BenchmarkTraceID_String(b *testing.B) {
	id := genTraceID()
	b.Run("Trace.String", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = id.String()
		}
	})
	b.Run("hex encode", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			var buf [16]byte
			binary.BigEndian.PutUint64(buf[0:8], id.n.Hi)
			binary.BigEndian.PutUint64(buf[8:16], id.n.Lo)
			hex.EncodeToString(buf[:])
		}
	})
	b.Run("fmt sprintf", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = fmt.Sprintf("%016x%016x", id.n.Hi, id.n.Lo)
		}
	})
}

// Benchmarks as of 2025-04-28:
//
//	BenchmarkSpanID_String/Span.String    13.32 ns/op	  16 B/op	  1 allocs/op
//	BenchmarkSpanID_String/FormatUint     20.36 ns/op	  16 B/op	  1 allocs/op
//	BenchmarkSpanID_String/hex_encode     35.95 ns/op	  64 B/op	  2 allocs/op
//	BenchmarkSpanID_String/fmt_sprintf    66.90 ns/op	  24 B/op	  2 allocs/op
func BenchmarkSpanID_String(b *testing.B) {
	id := genSpanID()
	b.Run("Span.String", func(b *testing.B) {
		for b.Loop() {
			b.ReportAllocs()
			_ = id.String()
		}
	})
	b.Run("FormatUint", func(b *testing.B) {
		for b.Loop() {
			b.ReportAllocs()
			strconv.FormatUint(id.n, 16)
		}
	})
	b.Run("hex encode", func(b *testing.B) {
		for b.Loop() {
			b.ReportAllocs()
			var buf [16]byte
			binary.BigEndian.PutUint64(buf[0:8], id.n)
			hex.EncodeToString(buf[:])
		}
	})
	b.Run("fmt sprintf", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = fmt.Sprintf("%016x", id.n)
		}
	})
}

// Benchmarks as of 2024-04-28:
//
//	BenchmarkParseTraceID/ParseTraceID      12.19 ns/op	   0 B/op	 0 allocs/op
//	BenchmarkParseTraceID/hex.DecodeString  26.79 ns/op	  16 B/op	 1 allocs/op
//	BenchmarkParseTraceID/strconv.ParseUint 44.54 ns/op	   0 B/op	 0 allocs/op
func BenchmarkParseTraceID(b *testing.B) {
	id := genTraceID()
	s := id.String()
	b.Run("ParseTraceID", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_, err := ParseTraceID(s)
			if err != nil {
				b.Fatalf("ParseTraceID failed: %v", err)
			}
		}
	})
	b.Run("hex.DecodeString", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			bytes, err := hex.DecodeString(s)
			if err != nil {
				b.Fatalf("hex.DecodeString failed: %v", err)
			}
			if len(bytes) != 16 {
				b.Fatalf("expected 16 bytes, got %d", len(bytes))
			}
			_ = newTraceID(
				binary.BigEndian.Uint64(bytes[0:8]),
				binary.BigEndian.Uint64(bytes[8:16]),
			)
		}
	})
	b.Run("strconv.ParseUint", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			hiStr, loStr := s[:16], s[16:]
			hi, err := strconv.ParseUint(hiStr, 16, 64)
			if err != nil {
				b.Fatalf("strconv.ParseUint failed for hi: %v", err)
			}
			lo, err := strconv.ParseUint(loStr, 16, 64)
			if err != nil {
				b.Fatalf("strconv.ParseUint failed for lo: %v", err)
			}
			_ = newTraceID(hi, lo)
		}
	})
}

// Benchmarks as of 2024-04-28:
//
//	BenchmarkParseSpanID/ParseSpanID        8.91 ns/op  0 B/op  0 allocs/op
//	BenchmarkParseSpanID/hex.DecodeString  17.12 ns/op   8 B/op  1 allocs/op
//	BenchmarkParseSpanID/strconv.ParseUint 21.31 ns/op   0 B/op  0 allocs/op
func BenchmarkParseSpanID(b *testing.B) {
	id := genSpanID()
	s := id.String()
	b.Run("ParseSpanID", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_, err := ParseSpanID(s)
			if err != nil {
				b.Fatalf("ParseSpanID failed: %v", err)
			}
		}
	})
	b.Run("hex.DecodeString", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			bytes, err := hex.DecodeString(s)
			if err != nil {
				b.Fatalf("hex.DecodeString failed: %v", err)
			}
			if len(bytes) != 8 {
				b.Fatalf("expected 8 bytes, got %d", len(bytes))
			}
			_ = SpanID{n: binary.BigEndian.Uint64(bytes)}
		}
	})
	b.Run("strconv.ParseUint", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			n, err := strconv.ParseUint(s, 16, 64)
			if err != nil {
				b.Fatalf("strconv.ParseUint failed for hi: %v", err)
			}
			_ = SpanID{n: n}
		}
	})
}

func newTraceID(hi, lo uint64) TraceID {
	return TraceID{n: hextbl.Uint128{Hi: hi, Lo: lo}}
}
