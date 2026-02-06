package parts

import (
	"go-secondreality/internal/constants"
	"go-secondreality/internal/driver"
	"go-secondreality/internal/shim"
)

const (
	ddStarsCount = 1024
)

type ddStar struct {
	age   uint8
	xseed int16
	yseed int16
}

var (
	ddStars     [ddStarsCount]ddStar
	ddMulDivX   [256]int16
	ddMulDivY   [256]int16
	ddSeed      uint32
	ddStarFrame int
	ddStarLimit int
	ddPalFade   uint8
)

func ddRng16() uint16 {
	prod := uint64(0x343FD) * uint64(ddSeed)
	ddSeed = uint32(prod) + 0x269EC3
	return uint16((prod >> 32) & 0xFFFF)
}

func ddSatLshift3Inc(in uint8) uint8 {
	s := uint16(in+1) << 3
	if s > 255 {
		return 255
	}
	return uint8(s)
}

func ddStarsPaletteRamp() {
	if ddPalFade > 32 {
		return
	}
	bl := ddSatLshift3Inc(ddPalFade)
	ddPalFade++

	shim.Outp(0x3c8, 0)
	shim.Outp(0x3c9, 0)
	shim.Outp(0x3c9, 0)
	shim.Outp(0x3c9, 0)

	shim.Outp(0x3c9, uint32(((25*70/100)*int(bl))>>8))
	shim.Outp(0x3c9, uint32(((31*70/100)*int(bl))>>8))
	shim.Outp(0x3c9, uint32(((38*70/100)*int(bl))>>8))

	shim.Outp(0x3c9, uint32(((45*56/100)*int(bl))>>8))
	shim.Outp(0x3c9, uint32(((58*56/100)*int(bl))>>8))
	shim.Outp(0x3c9, uint32(((69*56/100)*int(bl))>>8))

	shim.Outp(0x3c9, uint32(((67*64/100)*int(bl))>>8))
	shim.Outp(0x3c9, uint32(((84*64/100)*int(bl))>>8))
	shim.Outp(0x3c9, uint32(((99*64/100)*int(bl))>>8))

	shim.Outp(0x3c9, 0)
	shim.Outp(0x3c9, 0)
	shim.Outp(0x3c9, 0)
	shim.Outp(0x3c9, 10)
	shim.Outp(0x3c9, 20)
	shim.Outp(0x3c9, 35)
	shim.Outp(0x3c9, 20)
	shim.Outp(0x3c9, 30)
	shim.Outp(0x3c9, 45)
	shim.Outp(0x3c9, 30)
	shim.Outp(0x3c9, 40)
	shim.Outp(0x3c9, 60)
}

func ddInitStars() {
	ddSeed = 0
	for i := 0; i < ddStarsCount; i++ {
		age := uint8(0)
		if i&0xFF != 0 {
			age = uint8((i & 0xFF) - 1)
		}
		ddStars[i].age = age
		ddStars[i].xseed = int16(int(ddRng16()&1023) - 512)
		ddStars[i].yseed = int16(int(ddRng16()&1023) - 512)
	}

	t := 150
	for k := 0; k < 256; k++ {
		ddMulDivY[k] = int16(((uint32(108) << 16) / uint32(t)) >> 1)
		ddMulDivX[k] = int16(((uint32(144) << 16) / uint32(t)) >> 1)
		t += 4
	}

	ddStarLimit = 512
	ddStarFrame = 0
	ddPalFade = 0
}

func ddDrawStars() {
	limit := ddStarLimit
	if limit > ddStarsCount {
		limit = ddStarsCount
	}
	for i := 0; i < limit; i++ {
		s := &ddStars[i]
		age := s.age - 2
		s.age = age
		if age > 0xF0 {
			s.xseed = int16(int(ddRng16()&1023) - 512)
			s.yseed = int16(int(ddRng16()&1023) - 512)
			continue
		}

		idx := age
		y := (int32(s.yseed)*int32(ddMulDivY[idx])>>14 + 100)
		if y < 0 || y > 199 {
			continue
		}
		x := (int32(s.xseed)*int32(ddMulDivX[idx])>>14 + 160)
		if x < 0 || x > 319 {
			continue
		}

		col := byte(1)
		if age >= 180 {
			col = 3
		} else if age >= 110 {
			col = 2
		}

		off := int(y)*constants.ScreenWidth + int(x)
		if off >= 0 && off < len(shim.VRAM) {
			shim.VRAM[off] = col
		}
		off2 := off + constants.ScreenWidth*constants.ScreenHeight
		if off2 >= 0 && off2 < len(shim.VRAM) {
			shim.VRAM[off2] = col
		}
	}
}

func runDDStars() {
	shim.ClearScreen()
	ddInitStars()

	for !driver.WantsToQuit() {
		driver.Vsync(false)
		ddStarsPaletteRamp()
		ddStarFrame++
		if ddStarFrame == 1500 {
			ddStarLimit = ddStarsCount
		}

		clear(shim.VRAM[:constants.ScreenWidth*constants.VirtualScreenHeight])
		ddDrawStars()
		driver.Blit()
	}
}
