package trace

import (
	"sync"
	"testing"
	"time"

	"github.com/jschaf/observe/internal/epoch"
)

func TestLifecycle_Basic(t *testing.T) {
	t.Run("StopRecording", func(t *testing.T) {
		lc := newLifecycle()

		if !lc.isRecording() {
			t.Errorf("New lifecycle should be recording; got not recording")
		}

		if ok := lc.incTxStart(); !ok {
			t.Errorf("incTxStart should return true when recording; got false")
		}

		lc.stopRecording(epoch.NanosNow())

		if lc.isRecording() {
			t.Errorf("Lifecycle should not be recording after stopRecording; got recording")
		}

		if ok := lc.incTxStart(); ok {
			t.Errorf("incTxStart should return false after stopRecording; got true")
		}
	})

	t.Run("SpinWait", func(t *testing.T) {
		lc := newLifecycle()

		ok := lc.spinWaitTxFinish()
		if !ok {
			t.Errorf("spinWaitTxFinish should return true when no transactions are started; got false")
		}

		lc.incTxStart()

		// Wait a bit and then increment the finish counter.
		go func() {
			time.Sleep(1 * time.Microsecond)
			lc.incTxFinish()
		}()

		ok = lc.spinWaitTxFinish()
		if !ok {
			t.Fatalf("spinWaitTxFinish should return true when all transactions are finished; got false")
		}
	})
}

func TestLifecycle_Race(t *testing.T) {
	lc := newLifecycle()

	var wg sync.WaitGroup
	const count = 100
	const delay = 1 * time.Millisecond

	runMany := func(f func()) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range count {
				f()
				time.Sleep(delay)
			}
		}()
	}

	// Start transactions.
	runMany(func() { lc.incTxStart() })

	// Finish transactions.
	runMany(func() { lc.incTxFinish() })

	// Check end time.
	runMany(func() { lc.endTime() })

	// Check recording status.
	runMany(func() { lc.isRecording() })

	// Spin wait for transactions to finish.
	runMany(func() { lc.spinWaitTxFinish() })

	// Stop recording after a delay.
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(count / 2 * delay)
		for range count / 2 {
			lc.stopRecording(epoch.NanosNow())
			time.Sleep(delay)
		}
	}()

	wg.Wait()

	if lc.isRecording() {
		t.Errorf("Lifecycle should not be recording after stopRecording was called")
	}

	ok := lc.spinWaitTxFinish()
	if !ok {
		t.Errorf("spinWaitTxFinish should return true when all transactions are finished; got false")
	}
}

func TestLifecycle_SpinWaitRace(t *testing.T) {
	lc := newLifecycle()

	for range 10 {
		lc.incTxStart()
	}

	wg := sync.WaitGroup{}

	// Finish transactions in separate goroutines.
	wg.Add(10)
	for range 10 {
		go func() {
			defer wg.Done()
			time.Sleep(1 * time.Millisecond)
			lc.incTxFinish()
		}()
	}

	wg.Wait()
	ok := lc.spinWaitTxFinish()
	if !ok {
		t.Errorf("spinWaitTxFinish should return true when all transactions are finished; got false")
	}
}

func FuzzLifecycle(f *testing.F) {
	f.Add(uint64(0), uint64(0), false)
	f.Add(uint64(1), uint64(0), true)
	f.Add(uint64(10), uint64(5), true)
	f.Add(uint64(65535), uint64(65535), false)

	f.Fuzz(func(t *testing.T, startOps uint64, finishOps uint64, stopRecording bool) {
		// Limit the number of operations to avoid timeouts
		const maxOps = 10000
		startOps %= maxOps
		finishOps %= maxOps

		lc := newLifecycle()

		// Perform start operations
		for range startOps {
			lc.incTxStart()
		}

		// Perform finish operations
		for range finishOps {
			lc.incTxFinish()
		}

		// Optionally stop recording
		if stopRecording {
			lc.stopRecording(epoch.NanosNow())
		}

		// Verify state
		start, done := lc.loadTxCount()
		if start != startOps {
			t.Errorf("Expected start count to be %d, got %d", startOps, start)
		}
		if done != finishOps {
			t.Errorf("Expected done count to be %d, got %d", finishOps, done)
		}
		if lc.isRecording() == stopRecording {
			t.Errorf("Expected isRecording to be %v, got %v", !stopRecording, lc.isRecording())
		}

		// Test spinWaitTxFinish behavior
		if startOps == finishOps {
			ok := lc.spinWaitTxFinish()
			if !ok {
				t.Errorf("spinWaitTxFinish should return true when all transactions are finished; got false")
			}
		}
	})
}

func BenchmarkLifecycle(b *testing.B) {
	b.Run("three start and finish", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			lc := newLifecycle()
			lc.incTxStart()
			lc.incTxStart()
			lc.incTxStart()

			lc.incTxFinish()
			lc.incTxFinish()
			lc.incTxFinish()

			lc.spinWaitTxFinish()
		}
	})
	b.Run("three start and two finish", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			lc := newLifecycle()
			lc.incTxStart()
			lc.incTxStart()
			lc.incTxStart()

			lc.incTxFinish()
			lc.incTxFinish()

			lc.spinWaitTxFinish()
		}
	})
	b.Run("uncontended mutex three operations", func(b *testing.B) {
		b.ReportAllocs()
		n := 0
		for b.Loop() {
			mu := &sync.Mutex{}
			mu.Lock()
			n++
			mu.Unlock()

			mu.Lock()
			n++
			mu.Unlock()

			mu.Lock()
			n++
			mu.Unlock()
		}
	})
}
