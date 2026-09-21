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
	recording *driver.FrameClock
	done      chan struct{}
}

func newGame(renderer *graphics.Renderer, demoCfg demo.Config) *Game {
	return &Game{renderer: renderer, demoCfg: demoCfg}
}

func (g *Game) Update() error {
	if !g.demoReady {
		g.demoReady = true
		g.done = make(chan struct{})
		go func() { defer close(g.done); demo.Run(g.demoCfg) }()
	}
	if g.recording != nil && !g.recording.Step(g.done) {
		return ebiten.Termination
	}

	if ebiten.IsKeyPressed(ebiten.KeyEscape) || ebiten.IsWindowBeingClosed() {
		driver.RequestQuit()
		return ebiten.Termination
	}

	return nil
}

// NewRecordingGame keeps the original music-driven sequence at 70 Hz.
func NewRecordingGame(cfg Config) *Game {
	g := NewGame(cfg)
	g.recording = driver.NewFrameClock()
	return g
}

func (g *Game) Close() {
	if g.recording != nil {
		g.recording.Close()
		if g.done != nil {
			<-g.done
		}
	}
}

func (g *Game) RecordingChapter() string { return demo.CurrentPart() }

func (g *Game) Draw(screen *ebiten.Image) {
	g.renderer.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	// Use a fixed logical resolution so Ebiten always creates a window.
	return constants.VirtualScreenWidth, constants.VirtualScreenHeight
}
