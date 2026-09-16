package parts

import (
	"encoding/binary"
	"log"

	"go-secondreality/internal/constants"
	"go-secondreality/internal/driver"
	"go-secondreality/internal/music"
	"go-secondreality/internal/shim"
)

const (
	waterBackgroundOffset = 778
	waterFBW              = 158
	waterFBH              = 34
	waterFBSize           = waterFBW * waterFBH
)

func runWater() {
	if err := waterEnsureData(); err != nil {
		log.Printf("water: %v", err)
		return
	}

	pal := make([]byte, constants.PaletteByteCount)
	tmppal := make([]byte, constants.PaletteByteCount)
	font := make([]byte, 400*35)
	fbuf := make([]byte, waterFBSize+1)

	var co int
	var fadeout bool
	var quit bool
	var fp uint16
	var scp uint16
	var sss uint16

	if !shim.IsDemoFirstPart() {
		for !driver.WantsToQuit() && music.GetPlusFlags() <= 0 {
			driver.Vsync(false)
		}
		if driver.WantsToQuit() {
			return
		}
	}

	co = music.GetOrder()

	if 10 < len(waterMiekka) {
		end := 10 + constants.PaletteByteCount
		if end > len(waterMiekka) {
			end = len(waterMiekka)
		}
		copy(pal, waterMiekka[10:end])
	}

	if waterBackgroundOffset < len(waterMiekka) {
		end := waterBackgroundOffset + 400*34
		if end > len(waterMiekka) {
			end = len(waterMiekka)
		}
		copy(font, waterMiekka[waterBackgroundOffset:end])
	}

	for x := 0; x < constants.PaletteColorCount; x++ {
		shim.SetPal(x, 0, 0, 0)
	}

	if waterBackgroundOffset < len(waterTausta) {
		end := waterBackgroundOffset + constants.ScreenSize
		if end > len(waterTausta) {
			end = len(waterTausta)
		}
		copy(shim.VRAM[:end-waterBackgroundOffset], waterTausta[waterBackgroundOffset:end])
	}

	copy(tmppal, pal)
	clear(pal)

	bg := []byte(nil)
	if waterBackgroundOffset < len(waterTausta) {
		bg = waterTausta[waterBackgroundOffset:]
	}

	for y := 0; y < 63*2 && !driver.WantsToQuit(); y++ {
		driver.Vsync(false)

		if y&1 == 1 {
			for x := 0; x < constants.PaletteColorCount; x++ {
				idx := x * 3
				shim.SetPal(x, pal[idx], pal[idx+1], pal[idx+2])
				for pf := 0; pf < 3; pf++ {
					cidx := idx + pf
					if pal[cidx] < tmppal[cidx] {
						pal[cidx]++
					}
				}
			}
		}

		waterScr(sss, fbuf, bg)
		if sss == 2 {
			sss = 0
		} else {
			sss++
		}

		driver.Blit()
	}

	for !driver.WantsToQuit() {
		order, row := music.GetOrderRow()
		if !(order == co || row < 16) {
			break
		}
		driver.Vsync(false)
	}

	sss = 0
	scp = 0
	quit = false
	fp = 0
	fadeout = false
	clear(tmppal)

	for !driver.WantsToQuit() && !quit {
		driver.Vsync(false)

		if music.GetPlusFlags() == -11 {
			fadeout = true
		}

		if fadeout {
			if fp == 64 {
				quit = true
			} else {
				fp++
			}

			for x := 0; x < constants.PaletteColorCount; x++ {
				idx := x * 3
				shim.SetPal(x, pal[idx], pal[idx+1], pal[idx+2])
				for pf := 0; pf < 3; pf++ {
					cidx := idx + pf
					if pal[cidx] > tmppal[cidx] {
						pal[cidx]--
					}
				}
			}
		}

		waterScr(sss, fbuf, bg)

		if sss == 2 {
			sss = 0
			if waterFBSize > 1 {
				copy(fbuf[:waterFBSize-1], fbuf[1:waterFBSize])
			}
			for x := 0; x < waterFBH; x++ {
				pos := waterFBW + x*waterFBW
				if pos < 0 || pos >= len(fbuf) {
					continue
				}
				src := x*400 + int(scp)
				if src < 0 || src >= len(font) {
					continue
				}
				fbuf[pos] = font[src]
			}
			if scp < 390 {
				scp++
			}
		} else {
			sss++
		}

		driver.Blit()
	}
}

func waterScr(pos uint16, fbuf []byte, bg []byte) {
	switch pos {
	case 0:
		waterPutrouts1(waterWat1, fbuf, bg)
	case 1:
		waterPutrouts1(waterWat2, fbuf, bg)
	case 2:
		waterPutrouts1(waterWat3, fbuf, bg)
	}
}

func waterPutrouts1(src []byte, fbuf []byte, bg []byte) {
	if len(src) < 2 {
		return
	}
	idx := 0
	for index := 0; index < waterFBSize && idx+2 <= len(src); index++ {
		n := int(binary.LittleEndian.Uint16(src[idx:]))
		idx += 2
		if n == 0 {
			continue
		}
		v := byte(0)
		if index < len(fbuf) {
			v = fbuf[index]
		}
		for k := 0; k < n && idx+2 <= len(src); k++ {
			off := int(binary.LittleEndian.Uint16(src[idx:]))
			idx += 2
			if off < 0 || off >= len(shim.VRAM) {
				continue
			}
			if v != 0 {
				shim.VRAM[off] = v
				continue
			}
			if off >= 0 && off < len(bg) {
				shim.VRAM[off] = bg[off]
			}
		}
	}
}
