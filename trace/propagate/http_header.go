package propagate

import (
	"net/http"

	"github.com/jschaf/observe/internal/hextbl"
	"github.com/jschaf/observe/trace"
)

const (
	headerTraceparent = "Traceparent" // canonical Go casing; avoids allocation
	headerTracestate  = "Tracestate"
)

type HTTPHeader http.Header

// ExtractContext returns a Context from the w3c HTTP headers,
// traceparent and tracestate. Returns a zero-valued Context if the headers are
// not present or invalid.
func (h HTTPHeader) ExtractContext() trace.Context {
	m := map[string][]string(h) // using the underlying map is about 2x faster
	traceparent := m[headerTraceparent]
	if len(traceparent) == 0 {
		return trace.Context{}
	}
	sc := ParseTraceParent(traceparent[0])
	if !sc.IsValid() {
		return trace.Context{}
	}

	tracestate := m[headerTracestate]
	if len(tracestate) > 0 {
		state, err := trace.ParseState(tracestate[0])
		if err != nil {
			// Ignore the error. A tracestate parse error must not affect parsing
			// the traceparent according to the spec.
			return sc
		}
		sc.State = state
	}

	return sc
}

// InjectContext adds the traceparent and tracestate headers to the
// provided http.Header from Context.
func (h HTTPHeader) InjectContext(sc trace.Context) {
	if !sc.IsValid() {
		return
	}
	a := [55]byte{
		'0', '0', '-', // version
	}
	t := sc.TraceID.Bytes()
	copy(a[3:], t[:])
	a[35] = '-'
	s := sc.SpanID.Bytes()
	copy(a[36:], s[:])
	a[52] = '-'
	a[53] = hextbl.Lookup[sc.Flags>>4]
	a[54] = hextbl.Lookup[sc.Flags&0xf]
	http.Header(h).Set(headerTraceparent, string(a[:]))

	state := sc.State.String()
	if len(state) > 0 {
		http.Header(h).Set(headerTracestate, state)
	}
}

// ParseTraceParent parses the W3C trace and span ID from an HTTP header value.
// Returns a zero-valued Context if the value is empty or invalid.
// https://www.w3.org/TR/trace-context/#traceparent-header
//
//	<version>-<trace-id>-<span-id>-<trace-flags>
//	00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01
//
//	- version is 1-byte representing an 8-bit unsigned integer in hex
//	- trace-id is 16-bytes in hex
//	- span-id is 8-bytes in hex
//	- trace-flags is an 8-bit flag in hex
func ParseTraceParent(bs string) trace.Context {
	const minHeaderLen = 3 + 32 + 1 + 16 + 1 + 2 // 55
	if len(bs) < minHeaderLen {
		return trace.Context{}
	}

	// Version
	const maxVersion = 254
	version := hextbl.Reverse[bs[0]]<<4 | hextbl.Reverse[bs[1]]
	if version > maxVersion || bs[2] != '-' {
		return trace.Context{}
	}
	if version == 0 && len(bs) != minHeaderLen {
		return trace.Context{}
	}
	bs = bs[3:] // consume the version and '-'

	// Trace ID
	traceID, err := trace.ParseTraceID(bs[:32])
	if err != nil || bs[32] != '-' {
		return trace.Context{}
	}
	bs = bs[33:] // consume the trace ID and '-'

	// Span ID
	spanID, err := trace.ParseSpanID(bs[:16])
	if err != nil || bs[16] != '-' {
		return trace.Context{}
	}
	bs = bs[17:] // consume the span ID and '-'

	// Flags
	flag, ok := hextbl.ReadByte(bs[0], bs[1])
	if !ok {
		return trace.Context{}
	}
	flags := trace.Flags(flag) & trace.FlagsSampled

	return trace.Context{
		TraceID: traceID,
		SpanID:  spanID,
		Flags:   flags,
		Remote:  true,
	}
}
