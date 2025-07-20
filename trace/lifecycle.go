package trace

import (
	"runtime"
	"sync/atomic"

	"github.com/jschaf/observe/internal/epoch"
)

// lifecycle tracks the lifecycle of a span, specifically the number of
// transactions that have started and finished, and whether the span is still
// recording transactions. It uses an atomic counter to enable wait-free,
// thread-safe operations.
type lifecycle struct {
	// v is a 64-bit atomic counter with the following bit layout:
	//   - 0-15:  transaction start count
	//   - 16-31: transaction finish count
	//   - 32-62: unused
	//   -    63: whether the lifecycle ended (1) or is still recording (0)

	v   uint64
	end epoch.Nanos
}

func newLifecycle() *lifecycle {
	return &lifecycle{}
}

const (
	txBits  = 16
	endBits = 63
)

// incTxStart increments the transaction start count. Returns true if the
// lifecycle is still recording transactions, false if it has been stopped.
func (l *lifecycle) incTxStart() bool {
	n := atomic.AddUint64(&l.v, 1)
	return isRecordingBit(n)
}

// incTxFinish increments the transaction finish count.
func (l *lifecycle) incTxFinish() {
	atomic.AddUint64(&l.v, 1<<txBits)
}

// stopRecording marks the lifecycle as ended, indicating no more transactions
// will be recorded. Returns true if this call successfully stopped
// recording, false if it was already stopped.
func (l *lifecycle) stopRecording(end epoch.Nanos) bool {
	n := atomic.OrUint64(&l.v, 1<<endBits)
	isFirst := isRecordingBit(n)
	if !isFirst {
		return false // already stopped
	}
	l.end.Store(end)
	return true
}

func (l *lifecycle) endTime() epoch.Nanos {
	return l.end.Load()
}

// loadTxCount returns the current transaction start and finish counts.
func (l *lifecycle) loadTxCount() (start, done uint64) {
	n := atomic.LoadUint64(&l.v)
	mask := (uint64(1) << txBits) - 1
	start = n & mask
	done = (n >> txBits) & mask
	return start, done
}

// isRecording checks if the lifecycle is still recording transactions.
func (l *lifecycle) isRecording() bool {
	n := atomic.LoadUint64(&l.v)
	return isRecordingBit(n)
}

// isRecordingBits checks if the given bits indicate that the lifecycle is
// still recording transactions.
func isRecordingBit(v uint64) bool {
	return v&(1<<endBits) == 0
}

// spinWaitTxFinish waits until all transactions that have started have also
// finished.
func (l *lifecycle) spinWaitTxFinish() bool {
	const maxSpinIterations = 1 << 20
	const goschedInterval = 1024
	for i := range maxSpinIterations {
		started, finished := l.loadTxCount()
		if started == finished {
			return true
		}
		if i%goschedInterval == 0 {
			runtime.Gosched()
		}
	}
	return false
}
