package parts

import (
	"log"

	"go-secondreality/dck/internal/common"
	"go-secondreality/dck/internal/constants"
	"go-secondreality/dck/internal/driver"
	"go-secondreality/dck/internal/shim"
)

const creditsFonay = 32

var (
	creditsFonaorder = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789/?!:,.\"()+-")
	creditsFonap     [256]int
	creditsFonaw     [256]int

	creditsVramSplitTop    [constants.ScreenWidth * 100]byte
	creditsVramSplitBottom [constants.ScreenWidth * constants.ScreenHeight]byte
)

func creditsFontAt(x, y int) byte {
	if x < 0 || y < 0 || y >= creditsFontRows || x >= creditsFontStride {
		return 0
	}
	idx := y*creditsFontStride + x
	if idx < 0 || idx >= len(creditsFont) {
		return 0
	}
	return creditsFont[idx]
}

func creditsPutPixel(buf []byte, x, y int, c byte) {
	if x < 0 || y < 0 || x >= constants.ScreenWidth || y >= constants.ScreenHeight {
		return
	}
	idx := x + y*constants.ScreenWidth
	if idx < 0 || idx >= len(buf) {
		return
	}
	buf[idx] = c
}

func creditsReconstituteSplit(vertical, horizontal int) {
	clear(shim.VRAM)

	for i := 0; i < constants.ScreenHeight; i++ {
		rowTop := (i / 2) * constants.ScreenWidth
		if rowTop+constants.ScreenWidth > len(creditsVramSplitTop) {
			rowTop = len(creditsVramSplitTop) - constants.ScreenWidth
			if rowTop < 0 {
				rowTop = 0
			}
		}

		if horizontal < 0 {
			start := -horizontal
			if start < constants.ScreenWidth {
				count := constants.ScreenWidth + horizontal
				if count > 0 {
					if rowTop+start+count <= len(creditsVramSplitTop) {
						copy(shim.VRAM[i*constants.ScreenWidth:(i*constants.ScreenWidth)+count], creditsVramSplitTop[rowTop+start:rowTop+start+count])
					}
				}
			}
		} else {
			if horizontal < constants.ScreenWidth {
				count := constants.ScreenWidth - horizontal
				if count > 0 {
					if rowTop+count <= len(creditsVramSplitTop) {
						copy(shim.VRAM[i*constants.ScreenWidth+horizontal:(i*constants.ScreenWidth)+horizontal+count], creditsVramSplitTop[rowTop:rowTop+count])
					}
				}
			}
		}

		rowBottom := (i + constants.ScreenHeight + vertical)
		if rowBottom >= 0 && rowBottom < constants.DoubleScreenHeight {
			src := i * constants.ScreenWidth
			dst := rowBottom * constants.ScreenWidth
			if src+constants.ScreenWidth <= len(creditsVramSplitBottom) && dst+constants.ScreenWidth <= len(shim.VRAM) {
				copy(shim.VRAM[dst:dst+constants.ScreenWidth], creditsVramSplitBottom[src:src+constants.ScreenWidth])
			}
		}
	}
}

func creditsPrt(x, y int, txt []byte) {
	for _, ch := range txt {
		x2w := creditsFonaw[ch] + x
		sx := creditsFonap[ch]
		for x2 := x; x2 < x2w; x2++ {
			for y2 := y; y2 < y+creditsFonay; y2++ {
				d := creditsFontAt(sx, y2-y)
				creditsPutPixel(creditsVramSplitBottom[:], x2, y2, d)
			}
			sx++
		}
		x = x2w + 2
	}
}

func creditsPrtc(x, y int, text []byte, idx int) int {
	w := 0
	for i := idx; i < len(text) && text[i] != 0; i++ {
		w += creditsFonaw[text[i]] + 2
	}
	end := idx
	for end < len(text) && text[end] != 0 {
		end++
	}
	creditsPrt(x-w/2, y, text[idx:end])
	if end < len(text) {
		return end + 1
	}
	return end
}

func creditsScreenIn(pic []byte, text []byte) {
	clear(shim.VRAM)
	clear(creditsVramSplitBottom[:])
	clear(creditsVramSplitTop[:])

	driver.Vsync(false)

	if len(pic) >= 16+constants.PaletteByteCount {
		common.SetPalArea(pic[16:16+constants.PaletteByteCount], 0, constants.PaletteColorCount)
	}
	if len(pic) < 784 {
		return
	}
	img := pic[784:]

	y := 16
	idx := 0
	for idx < len(text) {
		idx = creditsPrtc(160, y, text, idx)
		if idx >= len(text) || text[idx] == 0 {
			break
		}
		y += creditsFonay + 10
	}

	for x := 0; x < 160; x++ {
		for y := 0; y < 100; y++ {
			off := y*160 + x
			if off >= len(img) {
				continue
			}
			px := img[off] + 16
			ix := 80 + x
			if ix < 0 || ix >= constants.ScreenWidth {
				continue
			}
			creditsVramSplitTop[ix+y*constants.ScreenWidth] = px
		}
	}

	for y := 200 * 128; y > 0 && !driver.WantsToQuit(); y = y * 12 / 13 {
		driver.Vsync(false)
		horizontal := y / 80
		vertical := y / 128
		creditsReconstituteSplit(vertical, horizontal)
		driver.Blit()
	}

	for a := 0; a < 200 && !driver.WantsToQuit(); a++ {
		driver.Vsync(false)
		driver.Blit()
	}

	v := 0
	for y := 0; y < 128*200 && !driver.WantsToQuit(); y = y + v {
		v += 15
		driver.Vsync(false)
		horizontal := -y / 80
		vertical := y / 128
		creditsReconstituteSplit(vertical, horizontal)
		driver.Blit()
	}
}

func creditsShiftPalette(pic []byte) {
	if len(pic) < 16+constants.PaletteByteCount {
		return
	}
	start := 16
	end := start + constants.PaletteByteCount - 16*3
	if end > len(pic) {
		end = len(pic)
	}
	copy(pic[start+16*3:], pic[start:end])
}

func creditsSetGraySteps(pic []byte, a int) {
	if len(pic) < 16+constants.PaletteByteCount {
		return
	}
	base := 16 + a*3
	if base+2 >= len(pic) {
		return
	}
	val := byte(7 * a)
	pic[base+0] = val
	pic[base+1] = val
	pic[base+2] = val
}

func creditsInit() {
	idx := 0
	for x := 0; x < creditsFontStride && idx < len(creditsFonaorder); {
		for x < creditsFontStride {
			found := false
			for y := 0; y < creditsFonay; y++ {
				if creditsFontAt(x, y) != 0 {
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
		for x < creditsFontStride {
			found := false
			for y := 0; y < creditsFonay; y++ {
				if creditsFontAt(x, y) != 0 {
					found = true
					break
				}
			}
			if !found {
				break
			}
			x++
		}
		ch := creditsFonaorder[idx]
		creditsFonap[ch] = b
		creditsFonaw[ch] = x - b
		idx++
	}

	creditsFonap[32] = creditsFontStride - 32
	creditsFonaw[32] = 8

	creditsShiftPalette(creditsPic1)
	creditsShiftPalette(creditsPic2)
	creditsShiftPalette(creditsPic3)
	creditsShiftPalette(creditsPic4)
	creditsShiftPalette(creditsPic5)
	creditsShiftPalette(creditsPic5b)
	creditsShiftPalette(creditsPic6)
	creditsShiftPalette(creditsPic7)
	creditsShiftPalette(creditsPic8)
	creditsShiftPalette(creditsPic9)
	creditsShiftPalette(creditsPic10)
	creditsShiftPalette(creditsPic10b)
	creditsShiftPalette(creditsPic11)
	creditsShiftPalette(creditsPic12)
	creditsShiftPalette(creditsPic13)
	creditsShiftPalette(creditsPic14)
	creditsShiftPalette(creditsPic14b)
	creditsShiftPalette(creditsPic15)
	creditsShiftPalette(creditsPic16)
	creditsShiftPalette(creditsPic17)
	creditsShiftPalette(creditsPic18)

	for a := 0; a < 10; a++ {
		creditsSetGraySteps(creditsPic1, a)
		creditsSetGraySteps(creditsPic2, a)
		creditsSetGraySteps(creditsPic3, a)
		creditsSetGraySteps(creditsPic4, a)
		creditsSetGraySteps(creditsPic5, a)
		creditsSetGraySteps(creditsPic5b, a)
		creditsSetGraySteps(creditsPic6, a)
		creditsSetGraySteps(creditsPic7, a)
		creditsSetGraySteps(creditsPic8, a)
		creditsSetGraySteps(creditsPic9, a)
		creditsSetGraySteps(creditsPic10, a)
		creditsSetGraySteps(creditsPic10b, a)
		creditsSetGraySteps(creditsPic11, a)
		creditsSetGraySteps(creditsPic12, a)
		creditsSetGraySteps(creditsPic13, a)
		creditsSetGraySteps(creditsPic14, a)
		creditsSetGraySteps(creditsPic14b, a)
		creditsSetGraySteps(creditsPic15, a)
		creditsSetGraySteps(creditsPic16, a)
		creditsSetGraySteps(creditsPic17, a)
		creditsSetGraySteps(creditsPic18, a)
	}
}

func runCredits() {
	if err := creditsEnsureData(); err != nil {
		log.Printf("credits: %v", err)
		return
	}

	shim.ClearScreen()
	creditsInit()

	if !driver.WantsToQuit() {
		creditsScreenIn(creditsPic1, []byte("GRAPHICS - MARVEL\x00MUSIC - SKAVEN\x00CODE - WILDFIRE\x00"))
	}
	if !driver.WantsToQuit() {
		creditsScreenIn(creditsPic2, []byte("GRAPHICS - MARVEL\x00MUSIC - SKAVEN\x00CODE - PSI\x00OBJECTS - WILDFIRE\x00"))
	}
	if !driver.WantsToQuit() {
		creditsScreenIn(creditsPic3, []byte("GRAPHICS - MARVEL\x00MUSIC - SKAVEN\x00CODE - WILDFIRE\x00ANIMATION - TRUG\x00"))
	}
	if !driver.WantsToQuit() {
		creditsScreenIn(creditsPic4, []byte("\x00GRAPHICS - PIXEL\x00"))
	}
	if !driver.WantsToQuit() {
		creditsScreenIn(creditsPic5, []byte("GRAPHICS - PIXEL\x00MUSIC - PURPLE MOTION\x00CODE - PSI\x00"))
	}
	if !driver.WantsToQuit() {
		creditsScreenIn(creditsPic5b, []byte("\x00MUSIC - PURPLE MOTION\x00CODE - TRUG\x00"))
	}
	if !driver.WantsToQuit() {
		creditsScreenIn(creditsPic6, []byte("\x00MUSIC - PURPLE MOTION\x00CODE - PSI\x00"))
	}
	if !driver.WantsToQuit() {
		creditsScreenIn(creditsPic7, []byte("\x00MUSIC - PURPLE MOTION\x00CODE - PSI\x00"))
	}
	if !driver.WantsToQuit() {
		creditsScreenIn(creditsPic8, []byte("\x00GRAPHICS - PIXEL\x00MUSIC - PURPLE MOTION\x00"))
	}
	if !driver.WantsToQuit() {
		creditsScreenIn(creditsPic9, []byte("GRAPHICS - PIXEL\x00MUSIC - PURPLE MOTION\x00CODE - TRUG\x00RENDERING - TRUG\x00"))
	}
	if !driver.WantsToQuit() {
		creditsScreenIn(creditsPic10, []byte("SKETCH - SKAVEN\x00GRAPHICS - PIXEL\x00MUSIC - PURPLE MOTION\x00CODE - PSI\x00"))
	}
	if !driver.WantsToQuit() {
		creditsScreenIn(creditsPic10b, []byte("SKETCH - SKAVEN\x00GRAPHICS - PIXEL\x00MUSIC - PURPLE MOTION\x00CODE - PSI\x00"))
	}
	if !driver.WantsToQuit() {
		creditsScreenIn(creditsPic11, []byte("\x00MUSIC - PURPLE MOTION\x00CODE - WILDFIRE\x00"))
	}
	if !driver.WantsToQuit() {
		creditsScreenIn(creditsPic12, []byte("\x00MUSIC - PURPLE MOTION\x00CODE - WILDFIRE\x00"))
	}
	if !driver.WantsToQuit() {
		creditsScreenIn(creditsPic13, []byte("\x00MUSIC - PURPLE MOTION\x00CODE - PSI\x00"))
	}
	if !driver.WantsToQuit() {
		creditsScreenIn(creditsPic14, []byte("GRAPHICS - PIXEL\x00MUSIC - PURPLE MOTION\x00CODE - TRUG\x00RENDERING - TRUG\x00"))
	}
	if !driver.WantsToQuit() {
		creditsScreenIn(creditsPic14b, []byte("\x00MUSIC - PURPLE MOTION\x00CODE - PSI\x00"))
	}
	if !driver.WantsToQuit() {
		creditsScreenIn(creditsPic15, []byte("GRAPHICS - MARVEL\x00MUSIC - PURPLE MOTION\x00CODE - PSI\x00"))
	}
	if !driver.WantsToQuit() {
		creditsScreenIn(creditsPic16, []byte("MUSIC - SKAVEN\x00CODE - PSI\x00WORLD - TRUG\x00"))
	}
	if !driver.WantsToQuit() {
		creditsScreenIn(creditsPic17, []byte("GRAPHICS - PIXEL\x00MUSIC - SKAVEN\x00"))
	}
	if !driver.WantsToQuit() {
		creditsScreenIn(creditsPic18, []byte("GRAPHICS - PIXEL\x00MUSIC - SKAVEN\x00CODE - WILDFIRE\x00"))
	}
}
