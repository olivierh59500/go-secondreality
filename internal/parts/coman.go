package parts

import (
	"encoding/binary"
	"log"

	"go-secondreality/internal/common"
	"go-secondreality/internal/constants"
	"go-secondreality/internal/driver"
	"go-secondreality/internal/music"
	"go-secondreality/internal/shim"
)

const comanLoopStride = constants.ScreenWidth / 2

const comanBgSize = 16 + constants.PaletteByteCount + constants.ScreenWidth*90 + 4*80

var (
	comanBg    [comanBgSize]byte
	comanBgUse [90 * 160]byte
	comanVbuf  [16384 * 4]byte

	comanCameraLevel int

	comanXsina1 int16
	comanYsina1 int16
	comanXsina2 int16
	comanYsina2 int16
	comanHeight int16

	comanPalette [constants.PaletteByteCount]byte
)

func comanReadI16(data []byte, off uint32) int16 {
	idx := int(off)
	if idx < 0 || idx+1 >= len(data) {
		return 0
	}
	return int16(binary.LittleEndian.Uint16(data[idx:]))
}

func comanAdd32CF(dst *uint32, src uint32) uint32 {
	a := *dst
	r := a + src
	*dst = r
	if r < a {
		return 1
	}
	return 0
}

func comanAdc32CF(dst *uint32, src uint32, cfIn uint32) uint32 {
	s := uint64(*dst) + uint64(src) + uint64(cfIn)
	*dst = uint32(s)
	return uint32(s >> 32)
}

func comanAdc16IntoLow(dst *uint32, imm int16, cfIn uint32) uint32 {
	lo := uint32(uint16(*dst))
	s := lo + uint32(uint16(imm)) + cfIn
	*dst = (*dst & 0xFFFF0000) | (s & 0xFFFF)
	if s > 0xFFFF {
		return 1
	}
	return 0
}

func comanTheLoop3(xw, yw, b int) {
	var eax uint32
	ecX := uint32(0xEC00FFFA)
	esi := uint32(xw) &^ 1
	edi := uint32(yw) &^ 1
	bp := b

	for _, blk := range comanBlocks {
		if blk.mode == 1 {
			si := uint16(esi)
			di := uint16(edi)
			si = si + uint16(comanXsina1)
			di = di + uint16(comanYsina1)
			esi = (esi & 0xFFFF0000) | uint32(si)
			edi = (edi & 0xFFFF0000) | uint32(di)
		} else {
			si := uint16(esi)
			di := uint16(edi)
			si = si + uint16(comanXsina2)
			di = di + uint16(comanYsina2)
			esi = (esi & 0xFFFF0000) | uint32(si)
			edi = (edi & 0xFFFF0000) | uint32(di)
		}

		bx := comanReadI16(comanW1dta, uint32(uint16(esi)))
		bx = int16(int32(bx) + int32(comanReadI16(comanW2dta, uint32(uint16(edi)))))
		bx = int16(int32(bx) + int32(blk.bxAdd))

		if int16(uint16(eax)) >= bx {
			if blk.mode == 1 {
				cf := comanAdd32CF(&eax, ecX)
				_ = comanAdc16IntoLow(&eax, -1, cf)
			} else {
				cf := comanAdd32CF(&eax, ecX)
				cf = comanAdc32CF(&eax, ecX, cf)
				_ = comanAdc16IntoLow(&eax, -1, cf)
			}
			continue
		}

		dl := uint8((uint16(uint16(bx) + uint16(blk.dxAdd))) & 0xFF)
		dl >>= 1

		for {
			cf := comanAdd32CF(&eax, blk.eaxInc)
			_ = comanAdc16IntoLow(&eax, 0, cf)

			if bp >= 0 && bp < len(comanVbuf) {
				comanVbuf[bp] = dl
			}

			axu := uint16(eax)
			if int16(axu) >= bx {
				cf1 := comanAdc32CF(&ecX, 0x00A00000<<4, 0)
				_ = comanAdc16IntoLow(&ecX, 0, cf1)
				bp -= comanLoopStride
				break
			}

			cf = comanAdd32CF(&eax, blk.eaxInc)
			_ = comanAdc16IntoLow(&eax, 0, cf)

			bp1 := bp - comanLoopStride
			if bp1 >= 0 && bp1 < len(comanVbuf) {
				comanVbuf[bp1] = dl
			}

			axu = uint16(eax)
			if int16(axu) >= bx {
				bp -= comanLoopStride

				cfAdd := comanAdd32CF(&ecX, 0x00A00000<<4)
				cf2 := comanAdc32CF(&ecX, 0x00A00000<<4, cfAdd)
				_ = comanAdc16IntoLow(&ecX, 0, cf2)
				bp -= comanLoopStride
				break
			}

			cf = comanAdd32CF(&eax, blk.eaxInc)
			_ = comanAdc16IntoLow(&eax, 0, cf)

			bp2 := bp - comanLoopStride*2
			if bp2 >= 0 && bp2 < len(comanVbuf) {
				comanVbuf[bp2] = dl
			}

			cfAdd := comanAdd32CF(&ecX, 0x01E00000<<4)
			_ = comanAdc16IntoLow(&ecX, 0, cfAdd)

			bp -= comanLoopStride * 3

			axu = uint16(eax)
			if int16(axu) >= bx {
				break
			}
		}

		if blk.mode == 1 {
			cf := comanAdd32CF(&eax, ecX)
			_ = comanAdc16IntoLow(&eax, -1, cf)
		} else {
			cf := comanAdd32CF(&eax, ecX)
			cf = comanAdc32CF(&eax, ecX, cf)
			_ = comanAdc16IntoLow(&eax, -1, cf)
		}
	}
}

func comanNewDoCol(xw, yw, xa, ya, b int) {
	comanHeight = int16(comanCameraLevel)

	comanXsina1 = int16(xa &^ 1)
	comanYsina1 = int16(ya &^ 1)

	comanXsina2 = comanXsina1 << 1
	comanYsina2 = comanYsina1 << 1

	comanTheLoop3(xw&0xFFFF, yw&0xFFFF, b)
}

func comanDoCopy(dest []byte, startrise int) {
	src0 := comanVbuf[60*160:]

	clearBase := 52 * 80 * 4
	clearCount := 18 * 80 * 4
	if clearBase+clearCount <= len(dest) {
		for i := 0; i < clearCount; i++ {
			dest[clearBase+i] = 0
		}
	}

	lines := 140 - startrise
	if lines <= 0 {
		return
	}

	dstLine := clearBase + 18*80*4 + startrise*constants.ScreenWidth
	for y := 0; y < lines; y++ {
		sy := y * 160
		dy := dstLine + y*constants.ScreenWidth
		if sy+160 > len(src0) || dy+constants.ScreenWidth > len(dest) {
			break
		}
		s := src0[sy:]
		d := dest[dy:]

		for z := 0; z < 80; z++ {
			a := s[z]
			d[4*z+0] = a
			d[4*z+1] = a
		}

		for z := 0; z < 80; z++ {
			b := s[z+80]
			d[4*z+2] = b
			d[4*z+3] = b
		}
	}

	clearStart := 68 * 160
	clearLen := 30 * 160
	if clearStart+clearLen <= len(comanVbuf) {
		for i := 0; i < clearLen; i++ {
			comanVbuf[clearStart+i] = 0
		}
	}
}

func comanDoIt() {
	startrise := 160
	frepeat := 1
	rot2 := 0
	rot := 0
	xwav := 0
	ywav := 0
	frame := 0
	comanCameraLevel = -270

	driver.Vsync(false)
	clear(comanVbuf[:])

	if !shim.IsDemoFirstPart() {
		for !driver.WantsToQuit() && music.GetPlusFlags() < 0 {
			driver.Vsync(false)
		}
	}

	for !driver.WantsToQuit() && frame < 4444 {
		a := music.GetPlusFlags()
		if a > -8 && a < 0 {
			break
		}

		for frepeat > 0 {
			frepeat--
			if a > -30 && a < 0 {
				if startrise < 160 {
					startrise++
				}
			}
			if frame < 400 && startrise > 0 {
				if frame == 4 {
					driver.Vsync(false)
					common.SetPalArea(comanPalette[:], 0, constants.PaletteColorCount)
				}
				if startrise > 0 {
					startrise--
				}
			}
		}

		rot2 += 4
		rot += int(common.Sin1024[rot2&1023]) / 15
		r := rot >> 3
		rsin := int(common.Sin1024[r&1023])
		rcos := int(common.Sin1024[(r+256)&1023])
		rsin2 := int(common.Sin1024[(r+177)&1023])
		rcos2 := int(common.Sin1024[(r+177+256)&1023])
		xw := xwav
		yw := ywav

		for i := 0; i < 160; i++ {
			x := i - 80
			y := 160
			xa := int((int32(x)*int32(rcos) + int32(y)*int32(rsin)) / 256)
			ya := int((int32(y)*int32(rcos2) - int32(x)*int32(rsin2)) / 256)
			b := (i&1)*80 + (i >> 1) + 199*160

			comanNewDoCol(xw&0xFFFF, yw&0xFFFF, xa, ya, b)

			if i == 80 {
				xwav += xa * 4
				ywav += ya * 4
			}
		}

		frepeat = driver.Vsync(false)
		frame += frepeat

		if startrise < 140 {
			comanDoCopy(shim.VRAM, startrise)
		}

		driver.Blit()
	}
}

func runComan() {
	if err := comanEnsureData(); err != nil {
		log.Printf("coman: %v", err)
		return
	}

	clear(comanBg[:])
	clear(comanBgUse[:])
	clear(comanVbuf[:])

	comanCameraLevel = 0
	comanXsina1 = 0
	comanYsina1 = 0
	comanXsina2 = 0
	comanYsina2 = 0
	comanHeight = 0
	clear(comanPalette[:])

	shim.ClearScreen()

	for a := 0; a < constants.PaletteColorCount; a++ {
		uc := (223 - a*22/26) * 3
		b := (230 - a) / 4
		b += int(common.Sin1024[(a*4)&1023]) / 32
		if b < 0 {
			b = 0
		}
		if b > 63 {
			b = 63
		}
		if uc+2 < len(comanPalette) {
			comanPalette[uc+1] = byte(b)
			b = (255 - a) / 3
			if b > 63 {
				b = 63
			}
			comanPalette[uc+2] = byte(b)

			b = a - 220
			if b < 0 {
				b = -b
			}
			if b > 40 {
				b = 40
			}
			b = 40 - b
			comanPalette[uc+0] = byte(b / 3)
		}
	}

	for a := 0; a < constants.PaletteByteCount-16*3; a++ {
		b := int(comanPalette[a])
		b = b * 9 / 6
		if b > 63 {
			b = 63
		}
		comanPalette[a] = byte(b)
	}

	for a := 0; a < 24; a++ {
		uc := (255 - a) * 3
		b := a - 4
		if b < 0 {
			b = 0
		}
		if uc+2 < len(comanPalette) {
			comanPalette[uc+0] = byte(b / 2)
			comanPalette[uc+1] = 0
			comanPalette[uc+2] = 0
		}
	}

	comanPalette[0] = 0
	comanPalette[1] = 0
	comanPalette[2] = 0

	for x := (constants.PaletteColorCount - 16) * 3; x < constants.PaletteByteCount; x++ {
		idx := 16 + x
		if idx < len(comanBg) {
			comanPalette[x] = comanBg[idx]
		}
	}

	for y := 0; y < 90; y++ {
		for x := 0; x < 80; x++ {
			src := x*4 + y*constants.ScreenWidth + constants.PaletteByteCount + 16
			if src >= 0 && src < len(comanBg) {
				comanBgUse[x+y*160] = comanBg[src]
			}
		}
		for x := 0; x < 80; x++ {
			src := x*4 + y*constants.ScreenWidth + 2 + constants.PaletteByteCount + 16
			if src >= 0 && src < len(comanBg) {
				comanBgUse[x+80+y*160] = comanBg[src]
			}
		}
	}

	driver.Vsync(false)
	common.SetPalArea(comanPalette[:], 0, constants.PaletteColorCount)

	clear(shim.VRAM)

	driver.Vsync(false)
	common.SetPalArea(comanPalette[:], 0, constants.PaletteColorCount)

	comanDoIt()
}
