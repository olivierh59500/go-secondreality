package driver

import (
	"testing"
	"time"
)

func TestVsyncWaitsWhilePaused(t *testing.T) {
	SetPaused(true)
	defer SetPaused(false)

	done := make(chan struct{})
	go func() {
		Vsync(false)
		close(done)
	}()

	select {
	case <-done:
		t.Fatal("Vsync returned while the driver was paused")
	case <-time.After(25 * time.Millisecond):
	}

	SetPaused(false)
	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Vsync did not resume after the pause was released")
	}
}
