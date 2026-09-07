// Package mobile exposes Second Reality to ebitenmobile.
package mobile

import (
	enginemobile "github.com/hajimehoshi/ebiten/v2/mobile"

	"go-secondreality/internal/app"
)

func init() {
	enginemobile.SetGame(app.NewGame(app.DefaultConfig()))
}

// Dummy forces gomobile to include this package in the Android binding.
func Dummy() {}
