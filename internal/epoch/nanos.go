package epoch

import (
	"sync/atomic"
	"time"
)

// Nanos is the nanoseconds since the Unix epoch.
type Nanos int64

// NewNanos converts a time.Time to Nanos. If the time is zero, it
// returns zero.
func NewNanos(t time.Time) Nanos {
	if t.IsZero() {
		return 0
	}
	return Nanos(t.UnixNano())
}

// NanosNow returns the current time in nanoseconds since the epoch.
func NanosNow() Nanos { return Nanos(time.Now().UnixNano()) }

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

// SwapIfZero atomically swaps the value of u with new if u is zero.
// Returns true if the value was swapped.
//
//goland:noinspection GoMixedReceiverTypes
func (u *Nanos) SwapIfZero(new Nanos) bool {
	p := (*int64)(u)
	return atomic.CompareAndSwapInt64(p, 0, int64(new))
}

//goland:noinspection GoMixedReceiverTypes
func (u *Nanos) Load() Nanos {
	return Nanos(atomic.LoadInt64((*int64)(u)))
}
