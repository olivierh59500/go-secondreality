package app

import (
	"go-secondreality/internal/demo"
	"go-secondreality/internal/driver"
	"go-secondreality/internal/graphics"

	"github.com/hajimehoshi/ebiten/v2"
)

// NewGame creates the shared game implementation used by desktop and mobile.
// Android-dependent services such as audio are deliberately started from the
// first Update rather than while the native library is loading.
func NewGame(cfg Config) *Game {
	renderer := graphics.NewRenderer()
	driver.SetRenderer(renderer)
	demoCfg := demo.Config{StartPart: cfg.StartPart, Loop: cfg.Loop}
	return newGame(renderer, demoCfg)
}

func Run(cfg Config) error {
	ebiten.SetWindowSize(960, 600)
	ebiten.SetWindowResizable(true)
	ebiten.SetWindowTitle("Second Reality")
	ebiten.SetTPS(70)
	if !cfg.Windowed {
		ebiten.SetFullscreen(true)
	}

	return ebiten.RunGame(NewGame(cfg))
}
