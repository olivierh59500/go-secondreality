package app

import (
	"go-secondreality/internal/driver"
	"go-secondreality/internal/graphics"
	"go-secondreality/internal/constants"

	"github.com/hajimehoshi/ebiten/v2"
)

type Game struct {
	renderer *graphics.Renderer
}

func NewGame(renderer *graphics.Renderer) *Game {
	return &Game{renderer: renderer}
}

func (g *Game) Update() error {
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
