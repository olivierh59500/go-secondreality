package parts

import (
	"bytes"
	"fmt"
	"log"
	"sync"

	srdata "go-secondreality"
	"go-secondreality/internal/common"
	"go-secondreality/internal/constants"
	"go-secondreality/internal/driver"
	"go-secondreality/internal/music"
	"go-secondreality/internal/shim"
)

const (
	alkuFontStride   = 1500
	alkuScrlf        = 9
	alkuPlanarWidth  = 352
	alkuPlanarHeight = 500
	alkuTbufHeight   = 186
)

const alkuFonaOrder = "ABCDEFGHIJKLMNOPQRSTUVWXabcdefghijklmnopqrstuvwxyz0123456789!?,.:\x8f\x8f()+-*='\x8f\x99"

var (
	alkuDataOnce sync.Once
	alkuDataErr  error
	alkuHzpic    []byte
	alkuFontData []byte
)

type alkuState struct {
	font     []byte
	planar   []byte
	palette  [constants.PaletteByteCount]byte
	palette2 [constants.PaletteByteCount]byte
	fuckpal  [constants.PaletteByteCount]byte
	fade1    [constants.PaletteByteCount]byte
	fade2    [constants.PaletteByteCount]byte
	picin    [constants.PaletteByteCount]int16
	textin   [constants.PaletteByteCount]int16
	textout  [constants.PaletteByteCount]int16
	cfpal    [constants.PaletteByteCount * 2]byte
	fonap    [256]int
	fonaw    [256]int
	dtau     []uint16
	tbuf     [alkuTbufHeight][alkuPlanarWidth]byte
	a        int
	p        int
	tptr     int
}

func runAlku() {
	if err := alkuEnsureData(); err != nil {
		log.Printf("alku: %v", err)
		return
	}

	st := &alkuState{
		font:   make([]byte, alkuFontStride*31+1500*5*2),
		planar: make([]byte, alkuPlanarWidth*alkuPlanarHeight),
		dtau:   make([]uint16, 30000),
	}
	st.run()
}

func alkuEnsureData() error {
	alkuDataOnce.Do(func() {
		data := srdata.AlkuData()
		var err error
		alkuHzpic, err = alkuExtractArray(data, "hzpic")
		if err != nil {
			alkuDataErr = err
			return
		}
		alkuFontData, err = alkuExtractArray(data, "alkuFont")
		if err != nil {
			alkuDataErr = err
			return
		}
	})
	return alkuDataErr
}

func alkuExtractArray(data []byte, name string) ([]byte, error) {
	idx := bytes.Index(data, []byte(name))
	if idx == -1 {
		return nil, fmt.Errorf("array %q not found", name)
	}
	brace := bytes.IndexByte(data[idx:], '{')
	if brace == -1 {
		return nil, fmt.Errorf("array %q missing '{'", name)
	}
	i := idx + brace + 1
	out := make([]byte, 0, 1024)
	for i < len(data) {
		c := data[i]
		if c == '}' {
			return out, nil
		}
		if c == '-' || (c >= '0' && c <= '9') {
			val, n := alkuParseNumber(data[i:])
			if n > 0 {
				out = append(out, byte(uint8(val)))
				i += n
				continue
			}
		}
		i++
	}
	return nil, fmt.Errorf("array %q unterminated", name)
}

func alkuParseNumber(data []byte) (int, int) {
	if len(data) == 0 {
		return 0, 0
	}
	i := 0
	sign := 1
	if data[i] == '-' {
		sign = -1
		i++
		if i >= len(data) {
			return 0, 0
		}
	}
	if data[i] == '0' && i+1 < len(data) && (data[i+1] == 'x' || data[i+1] == 'X') {
		i += 2
		val := 0
		for i < len(data) {
			d := data[i]
			var hv int
			switch {
			case d >= '0' && d <= '9':
				hv = int(d - '0')
			case d >= 'a' && d <= 'f':
				hv = int(d-'a') + 10
			case d >= 'A' && d <= 'F':
				hv = int(d-'A') + 10
			default:
				return sign * val, i
			}
			val = val*16 + hv
			i++
		}
		return sign * val, i
	}
	if data[i] < '0' || data[i] > '9' {
		return 0, 0
	}
	val := 0
	for i < len(data) && data[i] >= '0' && data[i] <= '9' {
		val = val*10 + int(data[i]-'0')
		i++
	}
	return sign * val, i
}

func (s *alkuState) run() {
	s.a = 0
	s.p = 0
	s.tptr = 0

	s.alkuInit()

	for music.Sync() < 1 && !alkuDemoWantsToQuit() {
		alkuDemoVsync()
		alkuDemoBlit()
	}
	if alkuDemoWantsToQuit() {
		return
	}

	s.prtc(160, 120, "A")
	s.prtc(160, 160, "Future Crew")
	s.prtc(160, 200, "Production")
	s.dofade(s.fade1[:], s.fade2[:])
	s.wait(300)
	s.dofade(s.fade2[:], s.fade1[:])
	s.fonapois()

	for music.Sync() < 2 && !alkuDemoWantsToQuit() {
		alkuDemoVsync()
		alkuDemoBlit()
	}
	if alkuDemoWantsToQuit() {
		return
	}

	s.prtc(160, 160, "First Presented")
	s.prtc(160, 200, "at Assembly 93")
	s.dofade(s.fade1[:], s.fade2[:])
	s.wait(300)
	s.dofade(s.fade2[:], s.fade1[:])
	s.fonapois()

	for music.Sync() < 3 && !alkuDemoWantsToQuit() {
		alkuDemoVsync()
		alkuDemoBlit()
	}
	if alkuDemoWantsToQuit() {
		return
	}

	s.prtc(160, 120, "in")
	s.prtc(160, 160, string([]byte{0x8f}))
	s.prtc(160, 179, string([]byte{0x99}))
	s.dofade(s.fade1[:], s.fade2[:])
	s.wait(300)
	s.dofade(s.fade2[:], s.fade1[:])
	s.fonapois()

	for music.Sync() < 4 && !alkuDemoWantsToQuit() {
		alkuDemoVsync()
		alkuDemoBlit()
	}
	if alkuDemoWantsToQuit() {
		return
	}

	copy(common.FadePal[:], s.fade1[:])
	common.CopFadePal = s.picin[:]
	common.CopDoFade = 128

	common.FrameCount = 0
	s.a = 1
	s.p = 1

	for common.CopDoFade != 0 && !alkuDemoWantsToQuit() {
		common.Copper2()
		common.Copper3()

		s.alkuDoScroll(2)

		alkuDemoVsync()
		alkuDemoBlit()
	}

	f := 60
	for s.a < constants.ScreenWidth && !alkuDemoWantsToQuit() {
		if f == 0 {
			common.CopFadePal = s.textin[:]
			common.CopDoFade = 64
			f += 20
		} else if f == 50 {
			common.CopFadePal = s.textout[:]
			common.CopDoFade = 64
			f++
		} else if f > 50 && common.CopDoFade == 0 {
			common.CopPal = s.palette[:]
			common.DoPal = 1
			s.clearTbuf()

			switch s.tptr {
			case 0:
				s.addtext(160, 50, "Graphics")
				s.addtext(160, 90, "Marvel")
				s.addtext(160, 130, "Pixel")
				s.ffonapois()
			case 1:
				s.faddtext(160, 50, "Music")
				s.faddtext(160, 90, "Purple Motion")
				s.faddtext(160, 130, "Skaven")
				s.ffonapois()
			case 2:
				s.faddtext(160, 30, "Code")
				s.faddtext(160, 70, "Psi")
				s.faddtext(160, 110, "Trug")
				s.faddtext(160, 148, "Wildfire")
				s.ffonapois()
			case 3:
				s.faddtext(160, 50, "Additional Design")
				s.faddtext(160, 90, "Abyss")
				s.faddtext(160, 130, "Gore")
				s.ffonapois()
			case 4:
				s.ffonapois()
			default:
				s.faddtext(160, 80, "BUG BUG BUG")
				s.faddtext(160, 130, "Timing error")
				s.ffonapois()
			}
			s.tptr++

			for ((s.a&1) != 0 || music.Sync() < 4+s.tptr) && !alkuDemoWantsToQuit() && s.a < 319 {
				common.Copper2()
				common.Copper3()

				s.alkuDoScroll(0)

				alkuDemoVsync()
				alkuDemoBlit()
			}

			aa := s.a
			if aa < constants.ScreenWidth-12 {
				s.fmaketext(aa + 16)
			}
			f = 0
		} else {
			f++
		}

		common.Copper2()
		common.Copper3()

		s.alkuDoScroll(1)
		alkuDemoVsync()
		alkuDemoBlit()
	}

	s.alkuDoScroll(1)
	alkuDemoBlit()

	if f > 63/alkuScrlf {
		s.dofade(s.palette2[:], s.palette[:])
	}

	s.fonapois()
}

func (s *alkuState) alkuInit() {
	shim.ClearScreen()
	common.SetPalArea(s.fade1[:], 0, constants.PaletteColorCount)

	if len(alkuHzpic) >= 16+constants.PaletteByteCount {
		copy(s.palette[:], alkuHzpic[16:16+constants.PaletteByteCount])
	}

	for a := 0; a < 88; a++ {
		s.outline(a*4+784, 4*(a+176*25))
		s.outline(a*4+784, 4*(a+176*25+88))
	}

	s.simulateScroll()

	for i := range s.font {
		s.font[i] = 0
	}
	if len(alkuFontData) > 0 {
		copyLen := len(alkuFontData)
		if copyLen > 45001 {
			copyLen = 45001
		}
		copy(s.font, alkuFontData[:copyLen])
	}

	for y := 0; y < 32; y++ {
		row := y * alkuFontStride
		for a := 0; a < alkuFontStride; a++ {
			switch s.font[row+a] & 3 {
			case 0x1:
				s.font[row+a] = 0x40
			case 0x2:
				s.font[row+a] = 0x80
			case 0x3:
				s.font[row+a] = 0xC0
			default:
				s.font[row+a] = 0
			}
		}
	}

	for y := 0; y < constants.PaletteByteCount; y += 3 {
		if y < 64*3 {
			s.palette2[y+0] = s.palette[y+0]
			s.palette2[y+1] = s.palette[y+1]
			s.palette2[y+2] = s.palette[y+2]
		} else if y < 128*3 {
			baseR := int(s.palette[0x1*3+0])
			baseG := int(s.palette[0x1*3+1])
			baseB := int(s.palette[0x1*3+2])
			s.fade2[y+0] = byte(baseR)
			s.fade2[y+1] = byte(baseG)
			s.fade2[y+2] = byte(baseB)
			s.palette2[y+0] = byte((baseR*63 + int(s.palette[y%(64*3)+0])*(63-baseR)) >> 6)
			s.palette2[y+1] = byte((baseG*63 + int(s.palette[y%(64*3)+1])*(63-baseG)) >> 6)
			s.palette2[y+2] = byte((baseB*63 + int(s.palette[y%(64*3)+2])*(63-baseB)) >> 6)
		} else if y < 192*3 {
			baseR := int(s.palette[0x2*3+0])
			baseG := int(s.palette[0x2*3+1])
			baseB := int(s.palette[0x2*3+2])
			s.fade2[y+0] = byte(baseR)
			s.fade2[y+1] = byte(baseG)
			s.fade2[y+2] = byte(baseB)
			s.palette2[y+0] = byte((baseR*63 + int(s.palette[y%(64*3)+0])*(63-baseR)) >> 6)
			s.palette2[y+1] = byte((baseG*63 + int(s.palette[y%(64*3)+1])*(63-baseG)) >> 6)
			s.palette2[y+2] = byte((baseB*63 + int(s.palette[y%(64*3)+2])*(63-baseB)) >> 6)
		} else {
			baseR := int(s.palette[0x3*3+0])
			baseG := int(s.palette[0x3*3+1])
			baseB := int(s.palette[0x3*3+2])
			s.fade2[y+0] = byte(baseR)
			s.fade2[y+1] = byte(baseG)
			s.fade2[y+2] = byte(baseB)
			s.palette2[y+0] = byte((baseR*63 + int(s.palette[y%(64*3)+0])*(63-baseR)) >> 6)
			s.palette2[y+1] = byte((baseG*63 + int(s.palette[y%(64*3)+1])*(63-baseG)) >> 6)
			s.palette2[y+2] = byte((baseB*63 + int(s.palette[y%(64*3)+2])*(63-baseB)) >> 6)
		}
	}

	for a := 192; a < constants.PaletteByteCount; a++ {
		s.palette[a] = s.palette[a-192]
	}

	order := []byte(alkuFonaOrder)
	x := 0
	oi := 0
	for x < alkuFontStride && oi < len(order) {
		for x < alkuFontStride {
			found := false
			for y := 0; y < 32; y++ {
				if s.font[y*alkuFontStride+x] != 0 {
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

		for x < alkuFontStride {
			empty := true
			for y := 0; y < 32; y++ {
				if s.font[y*alkuFontStride+x] != 0 {
					empty = false
					break
				}
			}
			if empty {
				break
			}
			x++
		}

		ch := order[oi]
		s.fonap[ch] = b
		s.fonaw[ch] = x - b
		oi++
	}

	s.fonap[32] = alkuFontStride - 20
	s.fonaw[32] = 16

	for a := 0; a < constants.PaletteByteCount; a++ {
		s.textin[a] = int16((int(s.palette2[a]) - int(s.palette[a])) * 256 / 64)
		s.textout[a] = int16((int(s.palette[a]) - int(s.palette2[a])) * 256 / 64)
		s.picin[a] = int16((int(s.palette[a]) - int(s.fade1[a])) * 256 / 128)
	}
}

func (s *alkuState) ascrolltext(scrl uint16, text []uint16) {
	const base = 100 * alkuPlanarWidth
	i := 0
	for {
		for j := 0; j < 20; j++ {
			if i+1 >= len(text) {
				return
			}
			idx := text[i]
			if idx == 0xFFFF {
				return
			}
			val := text[i+1]
			offset := base + int(scrl) + int(idx)
			if offset > 0 && offset-1 < len(s.planar) {
				s.planar[offset-1] ^= byte(val & 0xFF)
			}
			i += 2
		}
	}
}

func (s *alkuState) outline(srcOffset, dstOffset int) {
	const (
		srcStride  = 640
		dstStride  = alkuPlanarWidth * 2
		lines      = 75
		blockShift = lines * srcStride
		dstBlock   = lines * dstStride
	)
	maxSrc := srcOffset + 4 + blockShift + (lines-1)*srcStride
	maxDst := dstOffset + 4 + dstBlock + (lines-1)*dstStride
	if maxSrc >= len(alkuHzpic) || maxDst >= len(s.planar) {
		return
	}

	for off := 4; off >= 1; off-- {
		curSrc := srcOffset + off
		curDst := dstOffset + off
		for ccc := 0; ccc < lines; ccc++ {
			s.planar[curDst+ccc*dstStride] = alkuHzpic[curSrc+ccc*srcStride]
		}
		curSrc += blockShift
		for ccc := 0; ccc < lines; ccc++ {
			s.planar[curDst+dstBlock+ccc*dstStride] = alkuHzpic[curSrc+ccc*srcStride]
		}
	}
}

func (s *alkuState) simulateScroll() {
	for y := 0; y < constants.DoubleScreenHeight; y++ {
		src := common.CopStart*4 + common.CopScrl + y*alkuPlanarWidth
		dst := y * constants.ScreenWidth
		if src < 0 || src+constants.ScreenWidth > len(s.planar) {
			continue
		}
		if dst < 0 || dst+constants.ScreenWidth > len(shim.VRAM) {
			continue
		}
		copy(shim.VRAM[dst:dst+constants.ScreenWidth], s.planar[src:src+constants.ScreenWidth])
	}
}

func (s *alkuState) wait(t int) {
	for i := 0; i < t; i++ {
		if alkuDemoWantsToQuit() {
			break
		}
		alkuDemoVsync()
		alkuDemoBlit()
	}
}

func (s *alkuState) fonapois() {
	start := constants.ScreenWidth * 64
	end := constants.ScreenWidth * (64 + 256)
	if start < 0 {
		start = 0
	}
	if end > len(shim.VRAM) {
		end = len(shim.VRAM)
	}
	for i := start; i < end; i++ {
		shim.VRAM[i] &= 63
	}
}

func (s *alkuState) prt(x, y int, txt string) {
	x2 := 0
	y2w := y + 32
	for i := 0; i < len(txt); i++ {
		ch := txt[i]
		x2w := s.fonaw[ch] + x
		sx := s.fonap[ch]
		for x2 = x; x2 < x2w; x2++ {
			for y2 := y; y2 < y2w; y2++ {
				if y2 < 0 || y2 >= constants.ScreenHeight || x2 < 0 || x2 >= constants.ScreenWidth {
					continue
				}
				d := s.font[(y2-y)*alkuFontStride+sx]
				shim.VRAM[x2+y2*constants.ScreenWidth] |= d
			}
			sx++
		}
		x = x2 + 2
	}
}

func (s *alkuState) prtc(x, y int, txt string) {
	w := 0
	for i := 0; i < len(txt); i++ {
		w += s.fonaw[txt[i]] + 2
	}
	s.prt(x-w/2, y, txt)
}

func (s *alkuState) dofade(pal1, pal2 []byte) {
	var pal [constants.PaletteByteCount]byte
	for index := 0; index < 64 && !alkuDemoWantsToQuit(); index++ {
		for b := 0; b < constants.PaletteByteCount; b++ {
			pal[b] = byte((int(pal1[b])*(64-index) + int(pal2[b])*index) >> 6)
		}
		common.CopPal = pal[:]
		common.DoPal = 1
		common.Copper2()
		common.Copper3()
		alkuDemoVsync()
		alkuDemoBlit()
	}
}

func (s *alkuState) fdofade(pal1, pal2 []byte, lerpValue int) int {
	if lerpValue < 0 || lerpValue > 64 {
		return 0
	}
	for b := 0; b < constants.PaletteByteCount; b++ {
		s.fuckpal[b] = byte((int(pal1[b])*(64-lerpValue) + int(pal2[b])*lerpValue) >> 6)
	}
	common.CopPal = s.fuckpal[:]
	common.DoPal = 1
	return 0
}

func (s *alkuState) addtext(tx, ty int, txt string) {
	w := 0
	for i := 0; i < len(txt); i++ {
		w += s.fonaw[txt[i]] + 2
	}
	w /= 2
	for i := 0; i < len(txt); i++ {
		ch := txt[i]
		for x := 0; x < s.fonaw[ch]; x++ {
			for y := 0; y < 31; y++ {
				yy := y + ty
				xx := tx + x - w
				if yy < 0 || yy >= alkuTbufHeight || xx < 0 || xx >= alkuPlanarWidth {
					continue
				}
				s.tbuf[yy][xx] = s.font[y*alkuFontStride+s.fonap[ch]+x]
			}
		}
		tx += s.fonaw[ch] + 2
	}
}

func (s *alkuState) clearTbuf() {
	for y := 0; y < alkuTbufHeight; y++ {
		for x := 0; x < constants.ScreenWidth; x++ {
			s.tbuf[y][x] = 0
		}
	}
}

func (s *alkuState) alkuDoScroll(mode int) int {
	if mode != 0 {
		for common.FrameCount < alkuScrlf {
			common.Copper2()
			common.Copper3()
			alkuDemoVsync()
			alkuDemoBlit()
			if alkuDemoWantsToQuit() {
				return 0
			}
		}
	}
	if common.FrameCount < alkuScrlf {
		return 0
	}
	common.FrameCount -= alkuScrlf

	if mode == 1 {
		s.ascrolltext(uint16(s.a), s.dtau)
	}

	common.CopStart = s.a / 4
	common.CopScrl = s.a & 3

	if (s.a & 3) == 0 {
		src := (s.a/4+86)*4 + 784
		dst := 4 * ((s.a/4 + 86) + 176*25)
		s.outline(src, dst)
		s.outline(src, 4*((s.a/4+86)+176*25+88))
	}

	s.simulateScroll()
	s.a++
	s.p ^= 1
	return 1
}

func (s *alkuState) faddtext(tx, ty int, txt string) {
	w := 0
	for i := 0; i < len(txt); i++ {
		w += s.fonaw[txt[i]] + 2
	}
	w /= 2
	for i := 0; i < len(txt); i++ {
		ch := txt[i]
		for x := 0; x < s.fonaw[ch]; x++ {
			for y := 0; y < 32; y++ {
				yy := y + ty
				xx := tx + x - w
				if yy < 0 || yy >= alkuTbufHeight || xx < 0 || xx >= alkuPlanarWidth {
					continue
				}
				s.tbuf[yy][xx] = s.font[y*alkuFontStride+s.fonap[ch]+x]
			}
		}
		s.alkuDoScroll(0)
		tx += s.fonaw[ch] + 2
	}
}

func (s *alkuState) fmaketext(scrl int) {
	p := 0
	for y := 1; y < 184; y++ {
		for x := constants.ScreenWidth; x > 0; x-- {
			if s.tbuf[y][x] != s.tbuf[y][x-1] {
				if p+1 >= len(s.dtau) {
					break
				}
				s.dtau[p] = uint16(x + y*alkuPlanarWidth)
				s.dtau[p+1] = uint16(s.tbuf[y][x] ^ s.tbuf[y][x-1])
				p += 2
			}
		}
	}
	if p+1 < len(s.dtau) {
		s.dtau[p] = 0xFFFF
		s.dtau[p+1] = 0xFFFF
	}

	for x := 0; x < constants.ScreenWidth; x++ {
		for y := 1; y < 184; y++ {
			idx := y*alkuPlanarWidth + 352*100 + (x + scrl)
			if idx >= 0 && idx < len(s.planar) {
				s.planar[idx] ^= s.tbuf[y][x]
			}
		}
	}

	for s.a <= scrl && !alkuDemoWantsToQuit() {
		common.Copper2()
		common.Copper3()
		s.alkuDoScroll(0)
		alkuDemoVsync()
		alkuDemoBlit()
	}
}

func (s *alkuState) ffonapois() {
	start := 80 * 64 * 4
	end := 80 * (64 + 256 + 10) * 4
	if start < 0 {
		start = 0
	}
	if end > len(s.planar) {
		end = len(s.planar)
	}
	for i := start; i < end; i++ {
		s.planar[i] &= 0x3f
	}
	s.alkuDoScroll(0)
}

func (s *alkuState) fffade(pal1, pal2 []byte, frames int) {
	for i := 0; i < constants.PaletteByteCount; i++ {
		s.cfpal[i] = pal1[i]
		s.cfpal[i+constants.PaletteByteCount] = byte((int(pal2[i]) - int(pal1[i])) * 256 / frames)
	}
}

func alkuDemoBlit() {
	driver.Blit()
}

func alkuDemoWantsToQuit() bool {
	return driver.WantsToQuit()
}

func alkuDemoVsync() int {
	return driver.Vsync(false)
}
