package graphics

import (
	"github.com/olivierh59500/democonstructionkit/indexed"
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

type frameSnapshot struct {
	indices []byte
	palette [constants.PaletteColorCount]uint32
	mode    Mode
	valid   bool
}

type Renderer struct {
	mu      sync.Mutex
	pending *frameSnapshot
	display *frameSnapshot
	dirty   bool

	framePixels []byte
	frameImage  *ebiten.Image
	sourceImage *ebiten.Image
	displayMode Mode
	hasFrame    bool
}

func NewRenderer() *Renderer {
	const maxPixels = constants.VirtualScreenWidth * constants.VirtualScreenHeight
	return &Renderer{
		pending: &frameSnapshot{
			indices: make([]byte, maxPixels),
			mode:    Mode{Width: constants.ScreenWidth, Height: constants.ScreenHeight, JSSS: 0.80},
		},
		display: &frameSnapshot{
			indices: make([]byte, maxPixels),
			mode:    Mode{Width: constants.ScreenWidth, Height: constants.ScreenHeight, JSSS: 0.80},
		},
		framePixels: make([]byte, maxPixels*4),
		frameImage:  ebiten.NewImage(constants.VirtualScreenWidth, constants.VirtualScreenHeight),
	}
}

func (r *Renderer) Capture(mode Mode, vram []byte, palette *[constants.PaletteColorCount][4]byte, startPixel int) {
	pixelCount, valid := modePixelCount(mode)
	valid = valid && palette != nil && startPixel >= 0 && startPixel <= len(vram)-pixelCount

	r.mu.Lock()
	snapshot := r.pending
	snapshot.mode = mode
	snapshot.valid = valid
	if valid {
		copy(snapshot.indices[:pixelCount], vram[startPixel:startPixel+pixelCount])
		for i := 0; i < constants.PaletteColorCount; i++ {
			b := uint32(palette[i][0])
			g := uint32(palette[i][1])
			red := uint32(palette[i][2])
			snapshot.palette[i] = red | g<<8 | b<<16 | 0xFF<<24
		}
	}
	r.dirty = true
	r.mu.Unlock()
}

func (r *Renderer) Draw(screen *ebiten.Image) {
	if screen == nil {
		return
	}

	if r.consumePendingFrame() {
		r.uploadDisplayFrame()
	}
	if !r.hasFrame || r.sourceImage == nil {
		screen.Fill(color.Black)
		return
	}

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
	if destW != sw || destH != sh {
		screen.Fill(color.Black)
	}

	var opts ebiten.DrawImageOptions
	opts.GeoM.Scale(
		float64(destW)/float64(r.displayMode.Width),
		float64(destH)/float64(r.displayMode.Height),
	)
	opts.GeoM.Translate(float64(posX), float64(posY))
	opts.Filter = ebiten.FilterNearest

	screen.DrawImage(r.sourceImage, &opts)
}

func (r *Renderer) consumePendingFrame() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.dirty {
		return false
	}
	r.pending, r.display = r.display, r.pending
	r.dirty = false
	return true
}

func (r *Renderer) uploadDisplayFrame() {
	snapshot := r.display
	if !snapshot.valid {
		if r.hasFrame {
			r.refreshDisplayFrame()
			return
		}
		r.hasFrame = false
		r.sourceImage = nil
		return
	}

	pixelCount, valid := modePixelCount(snapshot.mode)
	if !valid {
		if r.hasFrame {
			r.refreshDisplayFrame()
			return
		}
		r.hasFrame = false
		r.sourceImage = nil
		return
	}
	if r.hasFrame && frameSnapshotIsEffectivelyBlack(snapshot, pixelCount) {
		r.refreshDisplayFrame()
		return
	}

	pixels := r.framePixels[:pixelCount*4]
	fillFramePixels(pixels, snapshot.indices[:pixelCount], &snapshot.palette)

	srcY := 0
	if snapshot.mode.Width == constants.VirtualScreenWidth && snapshot.mode.Height == 350 {
		srcY = (constants.VirtualScreenHeight - snapshot.mode.Height) / 2
	}
	srcRect := image.Rect(0, srcY, snapshot.mode.Width, srcY+snapshot.mode.Height)
	source := r.frameImage.SubImage(srcRect).(*ebiten.Image)
	source.WritePixels(pixels)

	r.sourceImage = source
	r.displayMode = snapshot.mode
	r.hasFrame = true
}

func (r *Renderer) refreshDisplayFrame() {
	if !r.hasFrame || r.sourceImage == nil {
		return
	}
	pixelCount, valid := modePixelCount(r.displayMode)
	if !valid {
		return
	}
	r.sourceImage.WritePixels(r.framePixels[:pixelCount*4])
}

func frameSnapshotIsEffectivelyBlack(snapshot *frameSnapshot, pixelCount int) bool {
	// This is an average RGB energy of roughly 0.01% of full scale. It catches
	// empty transition frames while preserving intentional very-dark fades.
	minimumColorEnergy := uint64(pixelCount) / 13
	if minimumColorEnergy == 0 {
		minimumColorEnergy = 1
	}
	var colorEnergy uint64
	for _, index := range snapshot.indices[:pixelCount] {
		color := snapshot.palette[index]
		colorEnergy += uint64(color&0xFF) + uint64((color>>8)&0xFF) + uint64((color>>16)&0xFF)
		if colorEnergy > minimumColorEnergy {
			return false
		}
	}
	return true
}

func modePixelCount(mode Mode) (int, bool) {
	if mode.Width <= 0 || mode.Width > constants.VirtualScreenWidth ||
		mode.Height <= 0 || mode.Height > constants.VirtualScreenHeight {
		return 0, false
	}
	return mode.Width * mode.Height, true
}

func fillFramePixels(dst, indices []byte, palette *[constants.PaletteColorCount]uint32) {
	_ = indexed.ExpandRGBA(dst, indices, palette)
}
