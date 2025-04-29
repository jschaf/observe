//go:build !unix

package epoch

// Now returns the current time in nanoseconds since the epoch.
func Now() Nanos { return Nanos(time.Now().UnixNano()) }
