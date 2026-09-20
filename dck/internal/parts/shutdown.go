package parts

import (
	"math"

	"go-secondreality/dck/internal/common"
	"go-secondreality/dck/internal/constants"
	"go-secondreality/dck/internal/driver"
	"go-secondreality/dck/internal/shim"
)

var (
	shutdownKuva     [65000]byte
	shutdownKuvaPal  [constants.PaletteByteCount]byte
	shutdownPal      [constants.PaletteByteCount]byte
	shutdownFadePals [64][constants.PaletteByteCount]byte

	shutdownOffset       int
	shutdownScrWidth     int = constants.VirtualScreenWidth
	shutdownScanDoubling bool
	shutdownVRAM         [constants.VirtualScreenWidth * 800]byte
)

func runShutdown() {
	clear(shutdownKuva[:])
	clear(shutdownKuvaPal[:])
	clear(shutdownPal[:])
	clear(shutdownFadePals[:])

	shutdownOffset = 0
	shutdownScrWidth = constants.VirtualScreenWidth
	shutdownScanDoubling = false
	clear(shutdownVRAM[:])

	for y := 0; y < constants.DoubleScreenHeight; y++ {
		src := y * constants.ScreenWidth
		dst := y*constants.VirtualScreenWidth + constants.ScreenWidth
		if src+constants.ScreenWidth > len(shim.VRAM) || dst+constants.ScreenWidth > len(shutdownVRAM) {
			break
		}
		copy(shutdownVRAM[dst:dst+constants.ScreenWidth], shim.VRAM[src:src+constants.ScreenWidth])
	}

	common.GetPalArea(shutdownKuvaPal[:], 0, constants.PaletteColorCount)

	shutdownScrWidth = constants.VirtualScreenWidth
	shutdownMain()
}

func shutdownMain() {
	for a := 0; a < constants.ScreenWidth; a++ {
		shutdownPutPixel(a, 0, 0)
	}

	for a := 0; a < 64; a++ {
		for b := 3; b < constants.PaletteByteCount; b++ {
			shutdownFadePals[a][b] = byte((a*63 + int(shutdownKuvaPal[b])*(64-a)) / 64)
		}
	}

	for y := 0; y < 100; y++ {
		for x := 0; x < constants.ScreenWidth; x++ {
			shutdownPutPixel(x, y+150, shutdownGetPixel(x+constants.ScreenWidth, y*4))
		}
	}

	shutdownSetStart(100 * 160)

	driver.Vsync(false)

	shutdownScanDoubling = true
	shutdownSetPalette(shutdownFadePals[3][:])

	shim.SetPal(0, 63, 63, 63)

	shutdownSetStart(0)

	driver.Vsync(false)

	driver.Blit()

	shutdownSetPalette(shutdownFadePals[20][:])
	shutdownScrWidth = constants.VirtualScreenWidth * 2

	driver.Vsync(false)
	driver.Blit()

	for a := 32; a > 2; a = a * 5 / 6 {
		driver.Vsync(false)
		shutdownSetPalette(shutdownFadePals[63-a][:])

		for b := a / 2; b <= a; b++ {
			shutdownCopyLine(0, (constants.ScreenHeight*160)-(b*constants.ScreenWidth), constants.PlanarWidth)
			shutdownCopyLine(0, (constants.ScreenHeight*160)+(b*constants.ScreenWidth), constants.PlanarWidth)
		}

		for b := 0; b < a; b++ {
			from := constants.PlanarWidth + (constants.DoubleScreenHeight*b/a)*160
			to := (constants.ScreenHeight * 160) + (b-a/2)*constants.ScreenWidth
			shutdownCopyLine(from, to, constants.PlanarWidth)
		}

		shutdownResolveVRAM()
		driver.Blit()
	}

	shutdownCopyLine(0, 202*160, constants.PlanarWidth)
	shutdownCopyLine(0, 198*160, constants.PlanarWidth)

	for x := 20; x <= 160; x += 3 {
		driver.Vsync(false)

		shutdownPutPixel(x, constants.ScreenHeight, 0)
		shutdownPutPixel(constants.ScreenWidth-x, constants.ScreenHeight, 0)
		shutdownPutPixel(x+1, constants.ScreenHeight, 0)
		shutdownPutPixel((constants.ScreenWidth-1)-x, constants.ScreenHeight, 0)
		shutdownPutPixel(x+2, constants.ScreenHeight, 0)
		shutdownPutPixel(318-x, constants.ScreenHeight, 0)
		shutdownPutPixel(x+3, constants.ScreenHeight, 0)
		shutdownPutPixel(317-x, constants.ScreenHeight, 0)

		shutdownResolveVRAM()
		driver.Blit()
	}

	shutdownPutPixel(160, 200, 1)

	for a := 0; a < 60; a++ {
		driver.Vsync(false)

		val := math.Cos(float64(a)/120.0*3.0*2.0*math.Pi)*31.0 + 32.0
		b := int(val)
		if b < 0 {
			b = 0
		} else if b > 63 {
			b = 63
		}

		shutdownSetRGBPalette(1, b, b, b)
		shutdownResolveVRAM()
		driver.Blit()
	}
}

func shutdownCopyLine(from, to, count int) {
	if count <= 0 {
		return
	}
	src := from * 4
	dst := to * 4
	bytes := count * 4
	if src < 0 || dst < 0 {
		return
	}
	if src+bytes > len(shutdownVRAM) || dst+bytes > len(shutdownVRAM) {
		return
	}
	copy(shutdownVRAM[dst:dst+bytes], shutdownVRAM[src:src+bytes])
}

func shutdownResolveVRAM() {
	src := shutdownOffset
	dst := 0
	height := constants.DoubleScreenHeight
	if shutdownScanDoubling {
		height = constants.ScreenHeight
	}

	for i := 0; i < height; i++ {
		if src+constants.ScreenWidth > len(shutdownVRAM) || dst+constants.ScreenWidth > len(shim.VRAM) {
			break
		}
		copy(shim.VRAM[dst:dst+constants.ScreenWidth], shutdownVRAM[src:src+constants.ScreenWidth])
		src += shutdownScrWidth
		dst += constants.ScreenWidth
		if shutdownScanDoubling {
			if dst+constants.ScreenWidth > len(shim.VRAM) {
				break
			}
			copy(shim.VRAM[dst:dst+constants.ScreenWidth], shim.VRAM[dst-constants.ScreenWidth:dst])
			dst += constants.ScreenWidth
		}
	}
}

func shutdownGetPixel(x, y int) byte {
	div := 1
	if shutdownScanDoubling {
		div = 2
	}
	row := y / div
	idx := x + row*shutdownScrWidth
	if idx < 0 || idx >= len(shutdownVRAM) {
		return 0
	}
	return shutdownVRAM[idx]
}

func shutdownPutPixel(x, y int, c byte) {
	div := 1
	if shutdownScanDoubling {
		div = 2
	}
	row := y / div
	idx := x + row*shutdownScrWidth
	if idx < 0 || idx >= len(shutdownVRAM) {
		return
	}
	shutdownVRAM[idx] = c
}

func shutdownSetPalette(p []byte) {
	common.SetPalArea(p, 0, constants.PaletteColorCount)
}

func shutdownSetRGBPalette(idx, r, g, b int) {
	shim.SetPal(idx, byte(r), byte(g), byte(b))
}

func shutdownSetStart(x int) {
	shutdownOffset = x * 4
}
