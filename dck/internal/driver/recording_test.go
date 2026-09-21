package driver

import (
	"sync/atomic"
	"testing"
)

func TestFrameClockWaitsForPartWork(t *testing.T) {
	c := NewFrameClock()
	t.Cleanup(func() { recordingClock = nil; wantsToQuit.Store(false) })
	var frames atomic.Int32
	done := make(chan struct{})
	go func() {
		defer close(done)
		for range 10 {
			c.vblank()
			frames.Add(1)
		}
	}()
	for i := int32(1); i <= 10; i++ {
		more := c.Step(done)
		if frames.Load() != i {
			t.Fatalf("part is not synchronized at tick %d: %d", i, frames.Load())
		}
		if more != (i < 10) {
			t.Fatalf("completion at tick %d: %v", i, more)
		}
	}
	c.Close()
}
