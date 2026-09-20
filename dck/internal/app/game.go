package app

import (
	"go-secondreality/dck/internal/constants"
	"go-secondreality/dck/internal/demo"
	"go-secondreality/dck/internal/driver"
	"go-secondreality/dck/internal/graphics"

	"github.com/hajimehoshi/ebiten/v2"
)

type Game struct {
	renderer  *graphics.Renderer
	demoCfg   demo.Config
	demoReady bool
}

func newGame(renderer *graphics.Renderer, demoCfg demo.Config) *Game {
	return &Game{renderer: renderer, demoCfg: demoCfg}
}

func (g *Game) Update() error {
	if !g.demoReady {
		g.demoReady = true
		go demo.Run(g.demoCfg)
	}

	if ebiten.IsKeyPressed(ebiten.KeyEscape) || ebiten.IsWindowBeingClosed() {
		driver.RequestQuit()
		return ebiten.Termination
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.renderer.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	// Use a fixed logical resolution so Ebiten always creates a window.
	return constants.VirtualScreenWidth, constants.VirtualScreenHeight
}
