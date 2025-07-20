package epoch

import (
	"sync/atomic"
	"time"
	_ "unsafe"
)

// Nanos is the nanoseconds since the Unix epoch.
//
//nolint:recvcheck
type Nanos int64

// NewNanos converts a time.Time to Nanos.
func NewNanos(t time.Time) Nanos { return Nanos(t.UnixNano()) }

// NanosNow returns the current time in nanoseconds since the epoch.
//
//go:nosplit
func NanosNow() Nanos { return Nanos(nanotime()) }

// ToTime converts the Nanos to a time.Time. If the value is zero, it
// returns the zero-value of time.Time.
//
//goland:noinspection GoMixedReceiverTypes
func (u Nanos) ToTime() time.Time {
	if u == 0 {
		return time.Time{}
	}
	return time.Unix(0, int64(u))
}

//goland:noinspection GoMixedReceiverTypes
func (u *Nanos) Load() Nanos {
	return Nanos(atomic.LoadInt64((*int64)(u)))
}

//goland:noinspection GoMixedReceiverTypes
func (u *Nanos) Store(ns Nanos) {
	atomic.StoreInt64((*int64)(u), int64(ns))
}

//go:linkname nanotime runtime.nanotime
//go:noescape
func nanotime() int64
