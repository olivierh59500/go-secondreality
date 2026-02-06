package driver

import (
	"sync/atomic"
	"time"

	"go-secondreality/internal/common"
	"go-secondreality/internal/graphics"
	"go-secondreality/internal/shim"
)

const vblankHz = 70

var (
	renderer    *graphics.Renderer
	lastVblank  time.Time
	wantsToQuit atomic.Bool
	currentMode = graphics.Mode{Width: 320, Height: 200, JSSS: shim.DefaultJSSS}
)

func SetRenderer(r *graphics.Renderer) {
	renderer = r
}

func ChangeMode(width, height int, jsss float32) {
	currentMode = graphics.Mode{Width: width, Height: height, JSSS: jsss}
	common.Reset()
	if renderer != nil {
		renderer.Clear()
	}
}

func Blit() {
	if renderer == nil {
		return
	}
	renderer.Capture(currentMode, shim.VRAM, &shim.Palette, shim.StartPixel)
}

func WantsToQuit() bool {
	return wantsToQuit.Load()
}

func RequestQuit() {
	wantsToQuit.Store(true)
}

func Vsync(updateAudio bool) int {
	if updateAudio {
		// Audio update will be wired later.
	}

	cycle := time.Second / time.Duration(vblankHz)
	now := time.Now()
	if lastVblank.IsZero() {
		lastVblank = now
		return 1
	}

	elapsed := now.Sub(lastVblank)
	if elapsed < cycle {
		time.Sleep(cycle - elapsed)
		lastVblank = lastVblank.Add(cycle)
		return 1
	}

	ticks := int(elapsed / cycle)
	if ticks < 1 {
		ticks = 1
	}
	lastVblank = lastVblank.Add(time.Duration(ticks) * cycle)
	return ticks
}
