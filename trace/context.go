package trace

// Context identifies a span in a trace.
type Context struct {
	TraceID TraceID
	SpanID  SpanID
	State   State
	Flags   Flags
	Remote  bool
}

func (sc Context) IsValid() bool { return sc.TraceID.IsValid() && sc.SpanID.IsValid() }

// IsSampled returns if Flags has the sampled bit set.
func (sc Context) IsSampled() bool { return sc.Flags.IsSampled() }
