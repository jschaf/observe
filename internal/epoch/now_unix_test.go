package epoch

import (
	"testing"
	"time"
)

// Benchmarks as of 2025-04-28:
// BenchmarkNow/unix_epoch.Now   72574975    14.64 ns/op
// BenchmarkNow/time_Now         38530129	   31.12 ns/op
func BenchmarkNow_unix(b *testing.B) {
	b.Run("unix epoch.Now", func(b *testing.B) {
		for b.Loop() {
			Now()
		}
	})
	b.Run("time Now", func(b *testing.B) {
		for b.Loop() {
			time.Now().UnixNano()
		}
	})
}
