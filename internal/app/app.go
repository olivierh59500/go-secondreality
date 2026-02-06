package app

import (
	"go-secondreality/internal/demo"
	"go-secondreality/internal/driver"
	"go-secondreality/internal/graphics"

	"github.com/hajimehoshi/ebiten/v2"
)

func Run(cfg Config) error {
	renderer := graphics.NewRenderer()
	driver.SetRenderer(renderer)

	ebiten.SetWindowSize(960, 600)
	ebiten.SetWindowResizable(true)
	ebiten.SetWindowTitle("Second Reality")
	ebiten.SetTPS(70)
	if !cfg.Windowed {
		ebiten.SetFullscreen(true)
	}

	game := NewGame(renderer)
	demoCfg := demo.Config{StartPart: cfg.StartPart, Loop: cfg.Loop}
	go demo.Run(demoCfg)

	return ebiten.RunGame(game)
}
