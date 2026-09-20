package parts

import (
	"log"

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

	lensW  int
	lensH  int
	lensXs int
	lensYs int

	lens1 []byte
	lens2 []byte
	lens3 []byte
	lens4 []byte

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
	lensXs = 0
	lensYs = 0
	lens1 = nil
	lens2 = nil
	lens3 = nil
	lens4 = nil
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
		lensXs = lensW / 2
		lensYs = lensH / 2
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

	lens1 = lensEx1
	lens2 = lensEx2
	lens3 = lensEx3
	lens4 = lensEx4

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

func lensDrawLens(x0, y0 int) {
	if lens1 == nil || lens2 == nil || lens3 == nil || lens4 == nil {
		return
	}

	u1 := (x0 - lensXs) + (y0-lensYs)*constants.ScreenWidth
	u2 := (x0 - lensXs) + (y0+lensYs-1)*constants.ScreenWidth
	ys := lensH / 2
	ye := lensH - 1

	for y := 0; y < ys; y++ {
		if u1 >= 0 && u1 <= constants.ScreenSize {
			lensDoRow(lens1, u1, y, 0x40)
			lensDoRow2(lens2, u1, y, 0x80)
			lensDoRow2(lens3, u1, y, 0xC0)
			lensDoRow3(lens4, u1, y)
		}
		u1 += constants.ScreenWidth
		if u2 >= 0 && u2 <= constants.ScreenSize {
			row := ye - y
			lensDoRow(lens1, u2, row, 0x40)
			lensDoRow2(lens2, u2, row, 0x80)
			lensDoRow2(lens3, u2, row, 0xC0)
			lensDoRow3(lens4, u2, row)
		}
		u2 -= constants.ScreenWidth
	}
}

func lensDoRow(lens []byte, U, Y, M int) {
	if lens == nil {
		return
	}
	mask := uint16(M & 0xFF)
	mask |= mask << 8

	row := Y << 2
	if row+4 > len(lens) {
		return
	}
	count := int(int16(lensReadU16(lens[row+2:])))
	offset := int(lensReadU16(lens[row:]))
	if count < 4 {
		return
	}
	if offset+2 > len(lens) {
		return
	}
	displacement := int(int16(lensReadU16(lens[offset:])))

	ediBase := U + displacement
	ebpBase := U + displacement
	esi := offset + 2

	if (ebpBase & 1) != 0 {
		if esi+2 > len(lens) {
			return
		}
		idx := int(int16(lensReadU16(lens[esi:])))
		esi += 2
		dst := ebpBase
		src := ediBase + idx
		if dst >= 0 && dst < len(shim.VRAM) && src >= 0 && src < len(lensBack) {
			shim.VRAM[dst] = lensBack[src] | byte(M&0xFF)
		}
		ebpBase++
		count--
	}

	pairs := count >> 1
	esi -= constants.ScreenWidth
	ebpBase -= constants.ScreenWidth

	start := 64 - pairs
	if start < 0 {
		start = 0
	}

	for i := start; i < 64 && i < start+pairs; i++ {
		idx := 63 - i
		dstOff := constants.ScreenWidth + idx*2
		srcOff1 := constants.ScreenWidth + idx*4
		srcOff2 := srcOff1 + 2
		lensDoWord(lens, ebpBase, dstOff, ediBase, esi, srcOff1, srcOff2, mask)
	}

	if (count & 1) != 0 {
		ebpBase += count &^ 1
		esi += (count &^ 1) << 1
		off := esi + constants.ScreenWidth
		if off+2 <= len(lens) {
			idx := int(int16(lensReadU16(lens[off:])))
			dst := ebpBase + constants.ScreenWidth
			src := ediBase + idx
			if dst >= 0 && dst < len(shim.VRAM) && src >= 0 && src < len(lensBack) {
				shim.VRAM[dst] = lensBack[src] | byte(M&0xFF)
			}
		}
	}
}

func lensDoWord(lens []byte, dstBase, dstOff, srcBase, offBase, srcOff1, srcOff2 int, mask uint16) {
	if lens == nil {
		return
	}
	if lensBack == nil {
		return
	}
	if offBase+srcOff1+2 > len(lens) || offBase+srcOff2+2 > len(lens) {
		return
	}

	idx1 := int(int16(lensReadU16(lens[offBase+srcOff1:])))
	idx2 := int(int16(lensReadU16(lens[offBase+srcOff2:])))
	src1 := srcBase + idx1
	src2 := srcBase + idx2
	dst := dstBase + dstOff
	if src1 < 0 || src2 < 0 || dst < 0 {
		return
	}
	if src1 >= len(lensBack) || src2 >= len(lensBack) || dst+1 >= len(shim.VRAM) {
		return
	}
	ax := uint16(lensBack[src1]) | (uint16(lensBack[src2]) << 8) | mask
	shim.VRAM[dst] = byte(ax)
	shim.VRAM[dst+1] = byte(ax >> 8)
}

func lensDoRow2(lens []byte, U, Y, M int) {
	if lens == nil {
		return
	}
	mask := byte(M & 0xFF)
	row := Y << 2
	if row+4 > len(lens) {
		return
	}
	count := int(int16(lensReadU16(lens[row+2:])))
	offset := int(lensReadU16(lens[row:]))
	if count == 0 {
		return
	}
	if offset+2 > len(lens) {
		return
	}
	base := int(int16(lensReadU16(lens[offset:])))
	ediBase := U + base
	ebp := offset + 2
	for i := 0; i < count; i++ {
		if ebp+4 > len(lens) {
			break
		}
		dstIdx := int(int16(lensReadU16(lens[ebp:])))
		srcIdx := int(int16(lensReadU16(lens[ebp+2:])))
		dst := ediBase + dstIdx
		src := ediBase + srcIdx
		if dst >= 0 && dst < len(shim.VRAM) && src >= 0 && src < len(lensBack) {
			shim.VRAM[dst] = lensBack[src] | mask
		}
		ebp += 4
	}
}

func lensDoRow3(lens []byte, U, Y int) {
	if lens == nil {
		return
	}
	row := Y << 2
	if row+4 > len(lens) {
		return
	}
	count := int(lensReadU16(lens[row+2:]))
	if count == 0 {
		return
	}
	offs := int(lensReadU16(lens[row:]))
	if offs+2 > len(lens) {
		return
	}
	base := int(int16(lensReadU16(lens[offs:])))
	baseIndex := U + base
	offs += 2
	for i := 0; i < count; i++ {
		if offs+2 > len(lens) {
			break
		}
		rel := int(int16(lensReadU16(lens[offs:])))
		offs += 2
		idx := baseIndex + rel
		if idx >= 0 && idx < len(shim.VRAM) && idx < len(lensBack) {
			shim.VRAM[idx] = lensBack[idx]
		}
	}
}

func lensRotate(x, y, xa, ya int) {
	xpos := int32(x) << 16
	ypos := int32(y) << 16
	Xadd := int32(int16(ya)) << 6
	Yadd := int32(int16(xa)) << 6

	src := lensRotPic
	if abs32(Xadd) > abs32(Yadd) {
		src = lensRot90
		t := Xadd
		Xadd = -Yadd
		Yadd = t
		tx := xpos
		ty := ypos
		xpos = -ty
		ypos = tx
	}
	if src == nil || len(src) < 256*256 {
		return
	}

	var modaLo [lensZoomXW / 4]uint16
	var modaHi [lensZoomXW / 4]uint16
	var modbLo [lensZoomXW / 4]uint16
	var modbHi [lensZoomXW / 4]uint16

	{
		var si uint16
		var di uint16
		var al uint8
		var ah uint8

		cx := uint16(uint32(Yadd) & 0xFFFF)
		dx := uint16(uint32(Xadd) & 0xFFFF)
		bl := uint8(uint32(Yadd) >> 16)
		bh := uint8(uint32(Xadd) >> 16)

		bh = uint8(-int8(bh))
		olddx := dx
		dx = uint16(0 - dx)
		if olddx != 0 {
			bh--
		}

		for i := 0; i < lensZoomXW/4; i++ {
			t := uint32(si) + uint32(cx)
			si = uint16(t)
			c := uint8(t >> 16)
			al = uint8(uint16(al) + uint16(bl) + uint16(c))

			t = uint32(di) + uint32(dx)
			di = uint16(t)
			c = uint8(t >> 16)
			ah = uint8(uint16(ah) + uint16(bh) + uint16(c))
			modaLo[i] = uint16(al) | (uint16(ah) << 8)

			t = uint32(si) + uint32(cx)
			si = uint16(t)
			c = uint8(t >> 16)
			al = uint8(uint16(al) + uint16(bl) + uint16(c))

			t = uint32(di) + uint32(dx)
			di = uint16(t)
			c = uint8(t >> 16)
			ah = uint8(uint16(ah) + uint16(bh) + uint16(c))
			modbLo[i] = uint16(al) | (uint16(ah) << 8)

			t = uint32(si) + uint32(cx)
			si = uint16(t)
			c = uint8(t >> 16)
			al = uint8(uint16(al) + uint16(bl) + uint16(c))

			t = uint32(di) + uint32(dx)
			di = uint16(t)
			c = uint8(t >> 16)
			ah = uint8(uint16(ah) + uint16(bh) + uint16(c))
			modaHi[i] = uint16(al) | (uint16(ah) << 8)

			t = uint32(si) + uint32(cx)
			si = uint16(t)
			c = uint8(t >> 16)
			al = uint8(uint16(al) + uint16(bl) + uint16(c))

			t = uint32(di) + uint32(dx)
			di = uint16(t)
			c = uint8(t >> 16)
			ah = uint8(uint16(ah) + uint16(bh) + uint16(c))
			modbHi[i] = uint16(al) | (uint16(ah) << 8)
		}
	}

	Xadd = lensScale307(Xadd)
	Yadd = lensScale307(Yadd)

	dst := lensZoomerPlanar[:]
	for row := 0; row < lensZoomYW; row++ {
		ypos += Yadd
		xpos += Xadd
		base := uint16(((uint32(ypos) >> 8) & 0xFF00) | ((uint32(xpos) >> 16) & 0x00FF))
		for i := 0; i < lensZoomXW/4; i++ {
			offA0 := modaLo[i]
			offA1 := modaHi[i]
			offB0 := modbLo[i]
			offB1 := modbHi[i]

			loA := src[int(uint16(base+offA0))]
			hiA := src[int(uint16(base+offA1))]
			dst[0] = loA
			dst[1] = hiA
			loB := src[int(uint16(base+offB0))]
			hiB := src[int(uint16(base+offB1))]
			dst[2] = loB
			dst[3] = hiB
			dst = dst[4:]
		}
	}
}

func lensScale307(v int32) int32 {
	lo := uint32(v) * 307
	return int32(lo >> 8)
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

func abs32(v int32) int32 {
	if v < 0 {
		return -v
	}
	return v
}
