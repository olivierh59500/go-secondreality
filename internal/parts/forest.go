package parts

import (
	"log"

	"go-secondreality/internal/constants"
	"go-secondreality/internal/driver"
	"go-secondreality/internal/music"
	"go-secondreality/internal/shim"
)

func runForest() {
	if err := forestEnsureData(); err != nil {
		log.Printf("forest: %v", err)
		return
	}

	fbuf := make([]byte, constants.VirtualScreenWidth*31)
	font := make([]byte, 237*31)
	pal := make([]byte, constants.PaletteByteCount)
	fpal := make([]byte, constants.PaletteByteCount)
	tmppal := make([]byte, constants.PaletteByteCount)

	var scp uint16
	var sss uint16
	const veke = 2800
	frame := 0
	quit := false
	fp := 0
	fadeout := false

	for row := 0; row < 31; row++ {
		src := 778 + row*constants.VirtualScreenWidth
		dst := row * constants.VirtualScreenWidth
		if src >= len(forestO2) || dst >= len(fbuf) {
			break
		}
		end := src + constants.VirtualScreenWidth
		if end > len(forestO2) {
			end = len(forestO2)
		}
		copy(fbuf[dst:dst+(end-src)], forestO2[src:end])
	}

	for i := 0; i < len(fbuf); i++ {
		if fbuf[i] > 0 {
			fbuf[i] = fbuf[i] + 128
		}
	}

	if 10 < len(forestHBack) {
		end := 10 + constants.PaletteByteCount
		if end > len(forestHBack) {
			end = len(forestHBack)
		}
		copy(pal, forestHBack[10:end])
	}

	if 778 < len(forestHBack) {
		end := 778 + constants.ScreenSize
		if end > len(forestHBack) {
			end = len(forestHBack)
		}
		copy(shim.VRAM[:end-778], forestHBack[778:end])
	}

	for i := 0; i < constants.PaletteColorCount; i++ {
		shim.SetPal(i, 0, 0, 0)
	}

	copy(tmppal, pal)
	clear(tmppal[:32*3])
	clear(tmppal[128*3 : 128*3+32*3])
	clear(fpal)

	for y := 0; y < 64; y++ {
		driver.Vsync(false)
		for x := 0; x < constants.PaletteColorCount; x++ {
			idx := x * 3
			shim.SetPal(x, fpal[idx], fpal[idx+1], fpal[idx+2])
		}
		for i := 0; i < constants.PaletteByteCount; i++ {
			if fpal[i] < tmppal[i] {
				fpal[i]++
			}
		}
		driver.Blit()
	}

	copy(tmppal, pal)
	copy(fpal, pal)
	clear(fpal[:32*3])
	clear(fpal[128*3 : 128*3+32*3])

	for x := 0; x < constants.PaletteColorCount; x++ {
		idx := x * 3
		shim.SetPal(x, fpal[idx], fpal[idx+1], fpal[idx+2])
	}

	for row := 0; row < 31; row++ {
		dst := row*237 + 104
		src := row * constants.VirtualScreenWidth
		if dst+133 > len(font) || src+133 > len(fbuf) {
			break
		}
		copy(font[dst:dst+133], fbuf[src:src+133])
	}

	scp = 133

	for !driver.WantsToQuit() && music.GetPlusFlags() > 0 {
		driver.Blit()
	}

	sss = 0
	for y := 0; y < 63*2 && !driver.WantsToQuit(); y++ {
		driver.Vsync(false)
		if y&1 == 1 {
			for x := 0; x <= 176 && x < constants.PaletteColorCount; x++ {
				idx := x * 3
				shim.SetPal(x, fpal[idx], fpal[idx+1], fpal[idx+2])
			}
			for i := 0; i <= 176 && i < constants.PaletteColorCount; i++ {
				idx := i * 3
				if fpal[idx] < tmppal[idx] {
					fpal[idx]++
				}
				if fpal[idx+1] < tmppal[idx+1] {
					fpal[idx+1]++
				}
				if fpal[idx+2] < tmppal[idx+2] {
					fpal[idx+2]++
				}
			}
		}

		forestStep(&scp, &sss, font, fbuf)
		driver.Blit()
	}

	clear(tmppal)
	frame = 0
	quit = false
	fp = 0
	fadeout = false

	for !driver.WantsToQuit() && frame != veke && !quit {
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
				shim.SetPal(x, fpal[idx], fpal[idx+1], fpal[idx+2])
				if fpal[idx] > tmppal[idx] {
					fpal[idx]--
				}
				if fpal[idx+1] > tmppal[idx+1] {
					fpal[idx+1]--
				}
				if fpal[idx+2] > tmppal[idx+2] {
					fpal[idx+2]--
				}
			}
		}

		forestStep(&scp, &sss, font, fbuf)
		frame++
		driver.Blit()
	}
}

func forestStep(scp *uint16, sss *uint16, font []byte, fbuf []byte) {
	switch *sss {
	case 0:
		forestPutroutsFromStream(forestPosi1, font)
	case 1:
		forestPutroutsFromStream(forestPosi2, font)
	default:
		forestPutroutsFromStream(forestPosi3, font)
		forestScrollFont(font, fbuf, scp)
	}
	if *sss == 2 {
		*sss = 0
	} else {
		*sss++
	}
}

func forestScrollFont(font []byte, fbuf []byte, scp *uint16) {
	if len(font) > 1 {
		copy(font[:len(font)-1], font[1:])
	}
	for row := 0; row < 31; row++ {
		dst := row*237 + 236
		src := row*constants.VirtualScreenWidth + int(*scp)
		if dst >= len(font) || src >= len(fbuf) {
			continue
		}
		font[dst] = fbuf[src]
	}
	if *scp < constants.VirtualScreenWidth-1 {
		*scp++
	}
}

func forestPutroutsFromStream(pos []byte, font []byte) {
	if len(forestHBack) < 778+constants.ScreenSize {
		return
	}
	bg := forestHBack[778:]
	p := 0
	fontIdx := 0
	blocks := 237 * 31
	for b := 0; b < blocks && fontIdx < len(font); b++ {
		if p+2 > len(pos) {
			return
		}
		count := int(pos[p]) | int(pos[p+1])<<8
		p += 2
		if count > 0 {
			if p+2*count > len(pos) {
				count = (len(pos) - p) / 2
			}
			for i := 0; i < count; i++ {
				dest := int(pos[p]) | int(pos[p+1])<<8
				p += 2
				if dest >= 0 && dest < constants.ScreenSize && dest < len(bg) && dest < len(shim.VRAM) {
					shim.VRAM[dest] = bg[dest] + font[fontIdx]
				}
			}
		}
		fontIdx++
	}
}
