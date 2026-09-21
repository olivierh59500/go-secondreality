package driver

// FrameClock lets the recorder advance the original 70 Hz vblank loop one
// step at a time, while the unchanged part code runs on its own goroutine.
type FrameClock struct {
	ready  chan struct{}
	resume chan struct{}
	stop   chan struct{}
	primed bool
}

var recordingClock *FrameClock

func NewFrameClock() *FrameClock {
	c := &FrameClock{ready: make(chan struct{}), resume: make(chan struct{}), stop: make(chan struct{})}
	recordingClock = c
	return c
}

func (c *FrameClock) vblank() int {
	select {
	case c.ready <- struct{}{}:
	case <-c.stop:
		return 1
	}
	select {
	case <-c.resume:
	case <-c.stop:
	}
	return 1
}

// Step returns only after the part has finished all work for the next vblank.
func (c *FrameClock) Step(done <-chan struct{}) bool {
	if !c.primed {
		select {
		case <-c.ready:
			c.primed = true
		case <-done:
			return false
		}
	}
	select {
	case c.resume <- struct{}{}:
	case <-done:
		return false
	}
	select {
	case <-c.ready:
		return true
	case <-done:
		return false
	}
}

func (c *FrameClock) Close() { RequestQuit(); close(c.stop) }

// FinishRecording stops after the final scroll has displayed a complete cycle.
// Interactive playback continues looping as before.
func FinishRecording() {
	if recordingClock != nil {
		RequestQuit()
	}
}
