package parts

import (
	"log"

	"go-secondreality/internal/common"
	"go-secondreality/internal/constants"
	"go-secondreality/internal/driver"
	"go-secondreality/internal/music"
	"go-secondreality/internal/shim"
)

const (
	jpSrcWidth  = 184
	jpRowStride = 186
)

var (
	jpRowdata1 [200][jpRowStride]byte
	jpRowdata2 [200][jpRowStride]byte
	jpRow      [400][]byte

	jpPal2    [constants.PaletteByteCount]byte
	jpPalette [constants.PaletteByteCount]byte
	jpRowbuf  [640]byte

	jpFrameY1  [201]int
	jpFrameY2  [201]int
	jpFrameY1T [800]int
	jpFrameY2T [800]int
	jpLastY    [400]int
	jpLastS    [400]int
)

func jpScrollY(y int) {
	shim.SetStartPixel(y * constants.ScreenWidth)
}

func jpLineZoom(dst []byte, src []byte, zoom int) {
	if len(dst) < constants.ScreenWidth {
		return
	}
	if zoom > 318 {
		zoom = 318
	}
	if zoom < 0 {
		zoom = 0
	}

	bg := byte(64)
	for i := 0; i < constants.ScreenWidth; i++ {
		dst[i] = bg
	}
	if src == nil || zoom == 0 {
		return
	}

	srcW := jpSrcWidth
	if len(src) < srcW {
		srcW = len(src)
	}
	if srcW <= 0 {
		return
	}

	if zoom > constants.ScreenWidth {
		zoom = constants.ScreenWidth
	}
	left := (constants.ScreenWidth - zoom) / 2
	if left < 0 {
		left = 0
	}

	for x := 0; x < zoom; x++ {
		sx := x * srcW / zoom
		if sx < 0 {
			sx = 0
		} else if sx >= srcW {
			sx = srcW - 1
		}
		dx := left + x
		if dx >= 0 && dx < constants.ScreenWidth {
			dst[dx] = src[sx]
		}
	}
}

func jpDoIt() {
	frame := 0
	y1 := 0
	y2 := 0

	if !shim.IsDemoFirstPart() {
		for !driver.WantsToQuit() && music.GetPlusFlags() < 4 {
			driver.Vsync(false)
		}
	}

	for !driver.WantsToQuit() && frame < 700 {
		c := driver.Vsync(false)
		frame += c
		if frame > 511 {
			_ = 400
		} else {
			y1 = jpFrameY1T[frame] / 16
			y2 = jpFrameY2T[frame] / 16
		}

		xsc := (400 - (y2 - y1)) / 8

		for y := 0; y < 400; y++ {
			dst := shim.VRAM[y*constants.ScreenWidth:]
			if y < y1 || y >= y2 {
				jpLineZoom(dst, nil, 0)
				continue
			}

			den := y2 - y1
			if den <= 0 {
				jpLineZoom(dst, nil, 0)
				continue
			}
			b := (y - y1) * 400 / den
			a := 184 + (int(common.Sin1024[(b*32/25)&1023])*xsc+32)/64
			a &^= 1

			if jpLastY[y] != b || jpLastS[y] != a {
				jpLineZoom(dst, jpRow[b], a)
				jpLastY[y] = b
				jpLastS[y] = a
			}
		}

		driver.Blit()
	}
}

func jpReset() {
	clear(jpRowdata1[:])
	clear(jpRowdata2[:])
	clear(jpRow[:])
	clear(jpPal2[:])
	clear(jpPalette[:])
	clear(jpRowbuf[:])
	clear(jpFrameY1[:])
	clear(jpFrameY2[:])
	clear(jpFrameY1T[:])
	clear(jpFrameY2T[:])
	for i := range jpLastY {
		jpLastY[i] = -1
		jpLastS[i] = -1
	}
}

func runJPLogo() {
	if err := jpEnsureData(); err != nil {
		log.Printf("jplogo: %v", err)
		return
	}

	jpReset()

	shim.ClearScreen()

	for a := 0; a < 200; a++ {
		jpRow[a] = jpRowdata1[a][:]
	}
	for a := 0; a < 200; a++ {
		jpRow[a+200] = jpRowdata2[a][:]
	}

	frame := 0
	halt := false
	y1 := 0
	y1a := 500
	y2 := 399 * 16
	y2a := 500
	mika := 1
	a := 0

	for frame = 0; frame < 200; frame++ {
		if !halt {
			y1 += y1a
			y2 += y2a
			y2a += 16
			if y2 > 400*16 {
				y2 -= y2a
				y2a = -y2a * mika / 8
				if mika < 4 {
					mika += 3
				}
			}
			y1a += 16

			la := a
			a = (y2 - y1) - 400*16
			if (a^la)&0x8000 != 0 {
				y1a = y1a * 7 / 8
			}
			y1a += a / 8
			y2a -= a / 8
		}

		if frame > 90 {
			if y2 >= 399*16 {
				y2 = 400 * 16
				halt = true
			} else {
				y2a = 8
			}
			y1 = y2 - 400*16
		}

		jpFrameY1[frame] = y1
		jpFrameY2[frame] = y2
	}

	jpFrameY1[200] = jpFrameY1[199]
	jpFrameY2[200] = jpFrameY2[199]

	for i := 0; i < 800; i++ {
		b := i / 4
		c := i & 3
		d := 3 - c
		jpFrameY1T[i] = (jpFrameY1[b]*d + jpFrameY1[b+1]*c) / 3
		jpFrameY2T[i] = (jpFrameY2[b]*d + jpFrameY2[b+1]*c) / 3
	}

	common.Readp(jpPalette[:], -1, jpPic)
	jpPalette[64*3+0] = 0
	jpPalette[64*3+1] = 0
	jpPalette[64*3+2] = 0

	for y := 0; y < 400; y++ {
		common.Readp(jpRowbuf[:], y, jpPic)
		copy(jpRow[y][:], jpRowbuf[70:70+jpSrcWidth])
		jpRow[y][jpSrcWidth] = 65
	}

	common.SetPalArea(jpPalette[:], 0, constants.PaletteColorCount)

	for a := 0; a < 400; a++ {
		for b := 0; b < jpSrcWidth; b++ {
			if jpRow[a][b] == 0 {
				jpRow[a][b] = 64
			}
		}
	}

	for y := 0; y < 400; y++ {
		jpLastY[y] = -1
		jpLastS[y] = -1
	}

	driver.Vsync(false)
	jpScrollY(400)
	driver.Vsync(false)

	for y := 0; y < 400; y++ {
		dst := shim.VRAM[y*constants.ScreenWidth:]
		jpLineZoom(dst, jpRow[y], jpSrcWidth)
	}

	a = 64
	y := 400 * 64
	for y > 0 && !driver.WantsToQuit() {
		y -= a
		a += 6
		if y < 0 {
			y = 0
		}
		jpScrollY(y / 64)
		driver.Vsync(false)
		driver.Blit()
	}

	driver.Vsync(false)
	jpDoIt()
	jpScrollY(0)
}
