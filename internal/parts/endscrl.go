package parts

import (
	"log"

	"go-secondreality/internal/blob"
	"go-secondreality/internal/driver"
	"go-secondreality/internal/shim"
)

const endscrlFonay = 25

var (
	endscrlFonaorder = []byte("ABCDEFGHIJKLMNOPQRSTUVWXabcdefghijklmnopqrstuvwxyz0123456789!?,.:" +
		"\x8f\x8f" +
		"()+-*='Z" +
		"\x84\x94" +
		"Y/&")

	endscrlText = make([]byte, 64000)

	endscrlFonap [256]int
	endscrlFonaw [256]int

	endscrlTptr   int
	endscrlTstart int
	endscrlChars  int

	endscrlTextline [100]byte
	endscrlScanbuf  [640]byte

	endscrlYscrl int
	endscrlLine  int
)

func endscrlSetStart(y int) {
	shim.SetStartPixel(y)
}

func endscrlSetRGBPalette(p, r, g, b int) {
	shim.SetPal(p, byte(r), byte(g), byte(b))
}

func endscrlFontPixel(x, y int) byte {
	if x < 0 || y < 0 || y >= endscrlFontRows || x >= endscrlFontStride {
		return 0
	}
	idx := y*endscrlFontStride + x
	if idx < 0 || idx >= len(endscrlFont) {
		return 0
	}
	return endscrlFont[idx]
}

func endscrlDoScroll() {
	if endscrlLine == 0 {
		endscrlTstart = 0
		endscrlChars = 0
		a := 0
		for endscrlTptr < len(endscrlText) {
			ch := endscrlText[endscrlTptr]
			if ch == '\n' || ch == 0 {
				break
			}
			if a < len(endscrlTextline) {
				endscrlTextline[a] = ch
			}
			endscrlTstart += endscrlFonaw[ch] + 2
			endscrlTptr++
			a++
			endscrlChars++
		}
		if endscrlTptr < len(endscrlText) {
			if a < len(endscrlTextline) {
				endscrlTextline[a] = endscrlText[endscrlTptr]
			}
			endscrlTptr++
		}
		endscrlTstart = (639 - endscrlTstart) / 2
		if endscrlTextline[0] == '[' {
			endscrlChars = 0
		}
	}

	for i := range endscrlScanbuf {
		endscrlScanbuf[i] = 0
	}

	x := endscrlTstart
	for a := 0; a < endscrlChars; a, x = a+1, x+2 {
		ch := endscrlTextline[a]
		w := endscrlFonaw[ch]
		start := endscrlFonap[ch]
		for b := 0; b < w; b++ {
			if x < 0 || x >= len(endscrlScanbuf) {
				x++
				continue
			}
			endscrlScanbuf[x] = endscrlFontPixel(start+b, endscrlLine)
			x++
		}
	}

	row := endscrlYscrl * 640
	row2 := (endscrlYscrl + 401) * 640
	if row >= 0 && row+640 <= len(shim.VRAM) {
		copy(shim.VRAM[row:row+640], endscrlScanbuf[:])
	}
	if row2 >= 0 && row2+640 <= len(shim.VRAM) {
		copy(shim.VRAM[row2:row2+640], endscrlScanbuf[:])
	}

	endscrlYscrl = (endscrlYscrl + 1) % 401

	if endscrlTextline[0] == '[' {
		height := int(endscrlTextline[1]-'0')*10 + int(endscrlTextline[2]-'0')
		if height <= 0 {
			height = endscrlFonay
		}
		endscrlLine = (endscrlLine + 1) % height
	} else {
		endscrlLine = (endscrlLine + 1) % endscrlFonay
	}

	endscrlSetStart(endscrlYscrl * 640)

	if endscrlTextline[0] == '%' {
		endscrlTptr = 0
	}
}

func endscrlInit() {
	data, err := blob.ReadAll("endscrol.txt")
	if err == nil {
		copy(endscrlText, data)
	}

	fonaIdx := 0
	for x := 0; x < endscrlFontStride && fonaIdx < len(endscrlFonaorder); {
		for x < endscrlFontStride {
			found := false
			for y := 0; y < endscrlFonay; y++ {
				if endscrlFontPixel(x, y) != 0 {
					found = true
					break
				}
			}
			if found {
				break
			}
			x++
		}
		b := x
		for x < endscrlFontStride {
			found := false
			for y := 0; y < endscrlFonay; y++ {
				if endscrlFontPixel(x, y) != 0 {
					found = true
					break
				}
			}
			if !found {
				break
			}
			x++
		}
		ch := endscrlFonaorder[fonaIdx]
		endscrlFonap[ch] = b
		endscrlFonaw[ch] = x - b
		fonaIdx++
	}

	endscrlFonap[32] = endscrlFontStride - 20
	endscrlFonaw[32] = 16
}

func runEndScrl() {
	if err := endscrlEnsureData(); err != nil {
		log.Printf("endscrl: %v", err)
		return
	}

	shim.ClearScreen()

	driver.Vsync(false)

	endscrlSetRGBPalette(1, 20, 20, 20)
	for i := 2; i <= 15; i++ {
		endscrlSetRGBPalette(i, 60, 60, 60)
	}

	endscrlInit()

	for !driver.WantsToQuit() {
		if driver.Vsync(false) == 1 {
			driver.Vsync(false)
		}
		endscrlDoScroll()
		driver.Blit()
	}
}
