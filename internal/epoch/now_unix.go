//go:build unix

package epoch

import (
	"syscall"
	"time"
)

func Now() Nanos {
	var tv syscall.Timeval
	err := syscall.Gettimeofday(&tv)
	if err != nil {
		return Nanos(time.Now().UnixNano())
	}
	return Nanos(tv.Nano())
}
