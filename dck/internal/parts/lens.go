package parts

import (
	"log"

	"github.com/olivierh59500/democonstructionkit/effects"
	"github.com/olivierh59500/democonstructionkit/indexed"

	"go-secondreality/dck/internal/common"
	"go-secondreality/dck/internal/constants"
	"go-secondreality/dck/internal/driver"
	"go-secondreality/dck/internal/music"
	"go-secondreality/dck/internal/shim"
)

const (
	lensZoomXW   = constants.ScreenWidth / 2
	lensZoomYW   = constants.ScreenHeight / 2
	lensExbSize  = 64784
	lensExbExtra = 4096
)

var (
	lensExbBuf [lensExbSize + lensExbExtra]byte

	lensPath1 []int16
	lensPath2 []int16

	lensBack   []byte
	lensRotPic []byte
	lensRot90  []byte

	lensW        int
	lensH        int
	lensMapper   *effects.IndexedLens
	lensRotozoom *indexed.Rotozoom256

	lensPalette [constants.PaletteByteCount]byte

	lensFade  = make([]byte, 16000)
	lensFade2 = make([]byte, 20000)

	lensZoomerPlanar [constants.ScreenWidth * constants.ScreenHeight]byte
	lensRotPicBuf    [16384 * 4]byte

	lensFirFade1  [200]int
	lensFirFade2  [200]int
	lensFirFade1a [200]int
	lensFirFade2a [200]int
)

func runLens() {
	if err := lensEnsureData(); err != nil {
		log.Printf("lens: %v", err)
		return
	}

	clear(lensExbBuf[:])
	lensPath1 = nil
	lensPath2 = nil
	lensBack = nil
	lensRotPic = nil
	lensRot90 = nil
	lensW = 0
	lensH = 0
	lensMapper = nil
	lensRotozoom = nil
	clear(lensPalette[:])
	clear(lensFade)
	clear(lensFade2)
	clear(lensZoomerPlanar[:])
	clear(lensRotPicBuf[:])
	clear(lensFirFade1[:])
	clear(lensFirFade2[:])
	clear(lensFirFade1a[:])
	clear(lensFirFade2a[:])

	copy(lensExbBuf[:], lensExb)

	shim.ClearScreen()
	lensRotPic = lensRotPicBuf[:]

	shim.Outp(0x3c8, 0)
	for i := 0; i < constants.PaletteByteCount; i++ {
		shim.Outp(0x3c9, 0)
	}

	if len(lensExp) >= 4 {
		count1 := int(int16(lensReadU16(lensExp[2:])))
		if count1 < 0 {
			count1 = 0
		}
		start := 4
		end := start + count1*2
		if end > len(lensExp) {
			end = len(lensExp)
		}
		lensPath1 = lensParseI16(lensExp[start:end])
		if end < len(lensExp) {
			lensPath2 = lensParseI16(lensExp[end:])
		} else {
			lensPath2 = nil
		}
	}

	if 16+constants.PaletteByteCount <= len(lensExbBuf) {
		copy(lensPalette[:], lensExbBuf[16:16+constants.PaletteByteCount])
	}

	if 16+constants.PaletteByteCount <= len(lensExbBuf) {
		lensBack = lensExbBuf[16+constants.PaletteByteCount:]
	}

	if len(lensEx0) >= 4 {
		lensW = int(int16(lensReadU16(lensEx0[0:])))
		lensH = int(int16(lensReadU16(lensEx0[2:])))
		cp := 4
		for i := 1; i < 4; i++ {
			if cp+2 >= len(lensEx0) {
				break
			}
			r := int(lensEx0[cp])
			g := int(lensEx0[cp+1])
			b := int(lensEx0[cp+2])
			cp += 3
			for a := 0; a < 64*3; a += 3 {
				c := r + int(lensPalette[a+0])
				if c > 63 {
					c = 63
				}
				lensPalette[a+i*64*3+0] = byte(c)
				c = g + int(lensPalette[a+1])
				if c > 63 {
					c = 63
				}
				lensPalette[a+i*64*3+1] = byte(c)
				c = b + int(lensPalette[a+2])
				if c > 63 {
					c = 63
				}
				lensPalette[a+i*64*3+2] = byte(c)
			}
		}
	}

	var err error
	lensMapper, err = lensCreateMapper()
	if err != nil {
		log.Printf("lens: %v", err)
		return
	}
	lensRotozoom, err = indexed.NewRotozoom256(indexed.Rotozoom256Config{Width: lensZoomXW, Height: lensZoomYW})
	if err != nil {
		log.Printf("lens: %v", err)
		return
	}

	lensBuildFades()
	lensBuildRotPic()

	driver.Vsync(false)
	common.SetPalArea(lensPalette[:], 0, constants.PaletteColorCount)

	if !driver.WantsToQuit() {
		lensPart1()
	}

	for !driver.WantsToQuit() && music.GetPlusFlags() < -20 {
		driver.Vsync(true)
	}

	if !driver.WantsToQuit() {
		lensPart2()
	}

	if !driver.WantsToQuit() {
		lensPart3()
	}
}

func lensBuildFades() {
	cp := 0
	for x := 0; x < 64; x++ {
		for y := 0; y < 64*3; y++ {
			a := (int(lensPalette[y])*(63-x) + x*63) / 63
			if cp < len(lensFade) {
				lensFade[cp] = byte(a)
			}
			cp++
		}
	}
	for x := 0; x < 16; x++ {
		for y := 0; y < 64*3; y++ {
			a := int(lensPalette[y]) + (15-x)*5
			if a > 63 {
				a = 63
			}
			if cp < len(lensFade) {
				lensFade[cp] = byte(a)
			}
			cp++
		}
	}

	cp = 0
	for x := 0; x < 32; x++ {
		for y := 64 * 3; y < constants.PaletteColorCount*3; y++ {
			aIdx := y % (64 * 3)
			a := int(lensPalette[y]) - int(lensPalette[aIdx])*(31-x)/31
			if cp < len(lensFade2) {
				lensFade2[cp] = byte(a)
			}
			cp++
		}
	}
}

func lensBuildRotPic() {
	if lensBack == nil {
		return
	}
	for x := 0; x < 256; x++ {
		for y := 0; y < 256; y++ {
			a := y*10/11 - 36/2
			if a < 0 || a > 199 {
				a = 0
			}
			src := x + 32 + a*constants.ScreenWidth
			dst := x + y*256
			if src >= 0 && src < len(lensBack) && dst < len(lensRotPic) {
				lensRotPic[dst] = lensBack[src]
			}
		}
	}
}

func lensPart1() {
	for b := 0; b < 200; b++ {
		a := b
		lensFirFade1a[b] = (19 + a/5 + 4) &^ 7
		lensFirFade2a[b] = (-(19 + (199-a)/5 + 4)) &^ 7
		lensFirFade1[b] = 170*64 + (100-b)*50
		lensFirFade2[b] = 170*64 + (100-b)*50
	}

	driver.Vsync(false)
	music.SetFrame(0)

	frame := 0
	for !driver.WantsToQuit() && frame < 300 {
		if frame < 80 {
			for c := 0; c < 6; c++ {
				cp := 0
				dp := 0
				for y := 0; y < constants.ScreenHeight; y++ {
					x := lensFirFade1[y] >> 6
					if x < 0 {
						x = 0
					} else if x > constants.ScreenWidth {
						x = constants.ScreenWidth
					}
					if dp+x >= 0 && dp+x < len(lensBack) && cp+x >= 0 && cp+x < len(shim.VRAM) {
						shim.VRAM[cp+x] = lensBack[dp+x]
					}
					x = lensFirFade2[y] >> 6
					if x < 0 {
						x = 0
					} else if x > constants.ScreenWidth {
						x = constants.ScreenWidth
					}
					if dp+x >= 0 && dp+x < len(lensBack) && cp+x >= 0 && cp+x < len(shim.VRAM) {
						shim.VRAM[cp+x] = lensBack[dp+x]
					}
					lensFirFade1[y] += lensFirFade1a[y]
					lensFirFade2[y] += lensFirFade2a[y]
					cp += constants.ScreenWidth
					dp += constants.ScreenWidth
				}
			}
		}

		driver.Blit()
		frame += driver.Vsync(false)
	}
}

func lensPart2() {
	frame := 0
	uframe := 0

	for !driver.WantsToQuit() && uframe < 715 {
		if uframe < 96 {
			a := (uframe - 32) / 2
			if a < 0 {
				a = 0
			}
			start := a * 3 * 192
			if start >= 0 && start+3*192 <= len(lensFade2) {
				common.SetPalArea(lensFade2[start:], 64, 192)
			}
		}

		if frame*2+1 < len(lensPath1) {
			x := int(lensPath1[frame*2+0])
			y := int(lensPath1[frame*2+1])
			lensDrawLens(x, y)
		}

		driver.Blit()
		a := driver.Vsync(false)
		uframe += a
		if a > 3 {
			a = 3
		}
		frame += a
	}

	for !driver.WantsToQuit() && uframe < 720 {
		uframe += driver.Vsync(true)
	}
}

func lensPart3() {
	lensRot90 = lensBack
	if lensRot90 != nil && len(lensRot90) >= 256*256 {
		for x := 0; x < 256; x++ {
			for y := 0; y < 256; y++ {
				dst := x + y*256
				src := y + (255-x)*256
				if dst < len(lensRot90) && src < len(lensRotPic) {
					lensRot90[dst] = lensRotPic[src]
				}
			}
		}
	}

	driver.Vsync(false)

	if 64*64*3 <= len(lensFade) {
		common.SetPalArea(lensFade[64*64*3:], 0, 64)
	}

	frame := 0
	for !driver.WantsToQuit() && frame < 2000 {
		if music.GetPlusFlags() > -4 {
			break
		}

		if frame*4+3 < len(lensPath2) {
			x := int(lensPath2[frame*4+0])
			y := int(lensPath2[frame*4+1])
			xa := int(lensPath2[frame*4+2])
			ya := int(lensPath2[frame*4+3])
			lensRotate(x, y, ya, xa)
		}

		lensZoomerToVRAM()

		frame += driver.Vsync(false)
		if frame > 2000-128 {
			a := frame - (2000 - 128)
			a /= 2
			if a > 63 {
				a = 63
			}
			start := a * 64 * 3
			if start >= 0 && start < len(lensFade) {
				common.SetPalArea(lensFade[start:], 0, 64)
			}
		}
		if frame < 16 {
			start := (64 + frame) * 64 * 3
			if start >= 0 && start < len(lensFade) {
				common.SetPalArea(lensFade[start:], 0, 64)
			}
		}
		driver.Blit()
	}

	for i := 0; i < constants.PaletteByteCount; i++ {
		lensPalette[i] = 63
	}
	common.SetPalArea(lensPalette[:], 0, constants.PaletteColorCount)
}

func lensZoomerToVRAM() {
	src := lensZoomerPlanar[:]
	dst := shim.VRAM
	for lineY := 0; lineY < 100; lineY++ {
		linesrc := src
		for i := 0; i < 40; i++ {
			dst[0] = linesrc[0]
			dst[1] = linesrc[0]
			dst[2] = linesrc[2]
			dst[3] = linesrc[2]
			dst[4] = linesrc[1]
			dst[5] = linesrc[1]
			dst[6] = linesrc[3]
			dst[7] = linesrc[3]
			dst = dst[8:]
			linesrc = linesrc[4:]
		}

		linesrc = src
		for i := 0; i < 40; i++ {
			dst[0] = linesrc[0]
			dst[1] = linesrc[0]
			dst[2] = linesrc[2]
			dst[3] = linesrc[2]
			dst[4] = linesrc[1]
			dst[5] = linesrc[1]
			dst[6] = linesrc[3]
			dst[7] = linesrc[3]
			dst = dst[8:]
			linesrc = linesrc[4:]
		}
		src = src[160:]
	}
}

// lensCreateMapper keeps the original displacement tables and palette banks,
// while sharing their validated, allocation-free renderer with other demos.
func lensCreateMapper() (*effects.IndexedLens, error) {
	return effects.NewIndexedLens(effects.IndexedLensConfig{
		Width: lensW, Height: lensH,
		CanvasWidth: constants.ScreenWidth, CanvasHeight: constants.ScreenHeight,
		Dense: lensEx1, Sparse: [2][]byte{lensEx2, lensEx3}, Restore: lensEx4,
		Masks: [3]byte{0x40, 0x80, 0xc0},
	})
}

func lensDrawLens(x, y int) {
	lensMapper.Draw(shim.VRAM, lensBack, x, y)
}

// lensRotate applies the indexed fixed-point rotozoom while preserving the
// source's palette banks and 16-bit wrap behavior.
func lensRotate(x, y, xa, ya int) {
	if lensRotozoom == nil {
		return
	}
	if err := lensRotozoom.Render(lensZoomerPlanar[:], lensRotPic, lensRot90, x, y, xa, ya); err != nil {
		log.Printf("lens rotozoom: %v", err)
	}
}

func lensParseI16(data []byte) []int16 {
	n := len(data) / 2
	out := make([]int16, n)
	for i := 0; i < n; i++ {
		out[i] = int16(lensReadU16(data[i*2:]))
	}
	return out
}

func lensReadU16(b []byte) uint16 {
	if len(b) < 2 {
		return 0
	}
	return uint16(b[0]) | uint16(b[1])<<8
}
