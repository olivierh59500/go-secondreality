package graphics

import (
	"image"
	"image/color"
	"sync"

	"go-secondreality/internal/constants"

	"github.com/hajimehoshi/ebiten/v2"
)

type Mode struct {
	Width  int
	Height int
	JSSS   float32
}

type Renderer struct {
	mu            sync.Mutex
	renderVram    []byte
	renderPalette [constants.PaletteColorCount][4]byte
	startPixel    int
	mode          Mode
	framePixels   []byte
	frameImage    *ebiten.Image
}

func NewRenderer() *Renderer {
	return &Renderer{
		renderVram:  make([]byte, constants.VRAMX*constants.VRAMY),
		framePixels: make([]byte, constants.VirtualScreenWidth*constants.VirtualScreenHeight*4),
		frameImage:  ebiten.NewImage(constants.VirtualScreenWidth, constants.VirtualScreenHeight),
		mode: Mode{
			Width:  constants.ScreenWidth,
			Height: constants.ScreenHeight,
			JSSS:   0.80,
		},
	}
}

func (r *Renderer) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	clear(r.framePixels)
}

func (r *Renderer) Capture(mode Mode, vram []byte, palette *[constants.PaletteColorCount][4]byte, startPixel int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.mode = mode
	r.startPixel = startPixel
	copy(r.renderVram, vram)

	for i := 0; i < constants.PaletteColorCount; i++ {
		b := palette[i][0]
		g := palette[i][1]
		rc := palette[i][2]
		r.renderPalette[i][0] = rc
		r.renderPalette[i][1] = g
		r.renderPalette[i][2] = b
		r.renderPalette[i][3] = 0xFF
	}
}

func (r *Renderer) Draw(screen *ebiten.Image) {
	if screen == nil {
		return
	}

	screen.Fill(color.Black)

	r.mu.Lock()
	r.fillFrameLocked()
	r.frameImage.WritePixels(r.framePixels)
	mode := r.mode
	r.mu.Unlock()

	sw := screen.Bounds().Dx()
	sh := screen.Bounds().Dy()

	ratioX := sw / constants.ScreenWidth
	ratioY := sh / constants.ScreenHeight
	ratio := ratioX
	if ratioY < ratio {
		ratio = ratioY
	}
	if ratio < 1 {
		ratio = 1
	}

	destW := ratio * constants.ScreenWidth
	destH := ratio * constants.ScreenHeight
	posX := (sw - destW) / 2
	posY := (sh - destH) / 2

	srcX := 0
	srcY := 0
	if mode.Width == constants.VirtualScreenWidth && mode.Height == 350 {
		srcY = (constants.VirtualScreenHeight - mode.Height) / 2
	}

	srcRect := image.Rect(srcX, srcY, srcX+mode.Width, srcY+mode.Height)
	sub := r.frameImage.SubImage(srcRect).(*ebiten.Image)

	scaleX := float64(destW) / float64(mode.Width)
	scaleY := float64(destH) / float64(mode.Height)

	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Scale(scaleX, scaleY)
	opts.GeoM.Translate(float64(posX), float64(posY))
	opts.Filter = ebiten.FilterNearest

	screen.DrawImage(sub, opts)
}

func (r *Renderer) fillFrameLocked() {
	clear(r.framePixels)

	mode := r.mode
	if mode.Width == 0 || mode.Height == 0 {
		return
	}

	srcIndex := r.startPixel
	dstYStart := 0
	if mode.Width == constants.VirtualScreenWidth && mode.Height == 350 {
		dstYStart = (constants.VirtualScreenHeight - mode.Height) / 2
	}

	for y := 0; y < mode.Height; y++ {
		dstRow := (dstYStart + y) * constants.VirtualScreenWidth
		for x := 0; x < mode.Width; x++ {
			if srcIndex < 0 || srcIndex >= len(r.renderVram) {
				return
			}
			color := r.renderPalette[r.renderVram[srcIndex]]
			di := (dstRow + x) * 4
			r.framePixels[di] = color[0]
			r.framePixels[di+1] = color[1]
			r.framePixels[di+2] = color[2]
			r.framePixels[di+3] = color[3]
			srcIndex++
		}
	}
}
