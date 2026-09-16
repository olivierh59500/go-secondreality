// Package mobile exposes Second Reality to ebitenmobile.
package mobile

import (
	enginemobile "github.com/hajimehoshi/ebiten/v2/mobile"

	"go-secondreality/internal/app"
	"go-secondreality/internal/driver"
)

func init() {
	enginemobile.SetGame(app.NewGame(app.DefaultConfig()))
}

// Pause stops the independently paced demo goroutine while Android is paused.
func Pause() {
	driver.SetPaused(true)
}

// Resume restarts demo pacing without trying to catch up for time spent paused.
func Resume() {
	driver.SetPaused(false)
}

// Dummy forces gomobile to include this package in the Android binding.
func Dummy() {}
