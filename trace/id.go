package trace

import (
	"fmt"
	"math"
	"math/rand/v2"

	"github.com/jschaf/observe/internal/hextbl"
)

// TraceID is the unique identifier for a trace.
//
//goland:noinspection GoNameStartsWithPackageName
type TraceID struct {
	n hextbl.Uint128
}

// IsValid checks whether the trace TraceID is valid.
func (t TraceID) IsValid() bool { return !t.n.IsZero() }

// Bytes return the hex string form of a TraceID as a byte array.
func (t TraceID) Bytes() [32]byte { return t.n.Bytes() }

// String returns the hex string representation form of a TraceID.
func (t TraceID) String() string {
	a := t.Bytes()
	return string(a[:])
}

// SpanID is the unique identifier for a span within a trace.
type SpanID struct {
	n uint64
}

// IsValid checks whether the SpanID is valid. A valid SpanID does not consist
// of zeros only.
func (s SpanID) IsValid() bool { return s.n != 0 }

// Bytes return the hex string form of a SpanID as a byte array.
func (s SpanID) Bytes() [16]byte { return hextbl.Uint64Bytes(s.n) }

// String returns the hex string form of a SpanID.
func (s SpanID) String() string {
	a := s.Bytes()
	return string(a[:])
}

func genTraceID() TraceID {
	return TraceID{n: hextbl.Uint128{Hi: rand.Uint64(), Lo: rand.Uint64()}} //nolint:gosec
}

func genSpanID() SpanID {
	return SpanID{n: rand.Uint64() & math.MaxInt64} //nolint:gosec
}

// ParseTraceID parses a 32-character hex string into a TraceID.
func ParseTraceID(s string) (TraceID, error) {
	n, ok := hextbl.ParseUint128(s)
	if !ok {
		return TraceID{}, fmt.Errorf("invalid hex trace ID: %s", s)
	}
	return TraceID{n: n}, nil
}

// ParseSpanID parses a 16-character hex string into a SpanID.
func ParseSpanID(s string) (SpanID, error) {
	n, ok := hextbl.ParseUint64(s)
	if !ok {
		return SpanID{}, fmt.Errorf("invalid hex span ID: %s", s)
	}
	return SpanID{n: n}, nil
}
