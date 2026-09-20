package parts

import (
	"log"
	"math"

	"go-secondreality/dck/internal/common"
	"go-secondreality/dck/internal/constants"
	"go-secondreality/dck/internal/driver"
	"go-secondreality/dck/internal/music"
	"go-secondreality/dck/internal/shim"
)

const (
	plzMaxY       = 280
	plzYAdd       = 0
	plzSize       = 84
	plzVmemWidth  = 640
	plzVmemHeight = 134
	plzVramPad    = 1024

	plzKuvaSize    = 64 * 256
	plzDistSize    = 128 * 256
	plzSinxSize    = 128
	plzSinySize    = 128
	plzPalSize     = constants.PaletteByteCount
	plzClrTauSize  = 8 * 256 * 2 * 2
	plzClrPtrSize  = 4
	plzVidMemSize  = 2 * plzSize * plzMaxY
	plzPalsSize    = 6 * constants.PaletteByteCount * 2
	plzFadePalSize = 2 * constants.PaletteByteCount

	plzFromOffsetKuva1 = 0
	plzFromOffsetKuva2 = plzFromOffsetKuva1 + plzKuvaSize
	plzFromOffsetKuva3 = plzFromOffsetKuva2 + plzKuvaSize
	plzFromOffsetDist1 = plzFromOffsetKuva3 + plzKuvaSize
	plzFromOffsetSinx  = plzFromOffsetDist1 + plzDistSize
	plzFromOffsetSiny  = plzFromOffsetSinx + plzSinxSize
	plzFromOffsetPal   = plzFromOffsetSiny + plzSinySize
	plzFromOffsetTau   = plzFromOffsetPal + plzPalSize
	plzFromOffsetClr   = plzFromOffsetTau + plzClrTauSize
	plzFromOffsetVid   = plzFromOffsetClr + plzClrPtrSize
	plzFromOffsetPals  = plzFromOffsetVid + plzVidMemSize
	plzFromOffsetFade  = plzFromOffsetPals + plzPalsSize
	plzFromMemSize     = plzFromOffsetFade + plzFadePalSize
)

type plzPolygonsToDraw struct {
	p   int
	dis int
}

type plzPoint struct {
	x   int
	y   int
	z   int
	xx  int
	yy  int
	zz  int
	xxx int
	yyy int
}

type plzPolygon struct {
	p1    int
	p2    int
	p3    int
	p4    int
	p5    int
	p6    int
	n     int
	color int
}

type plzLineDef struct {
	p1  int
	p2  int
	n   int
	col int
}

type plzVert struct {
	x  int
	y  int
	xf int32
	u  int32
	v  int32
}

type plzObject struct {
	name  string
	pnts  int
	point [256]plzPoint
	faces int
	pg    [256]plzPolygon
	lines int
	lin   [256]plzLineDef
}

var (
	plzTimetable = []int{
		64*6*2 - 45,
		64*6*4 - 45,
		64*6*5 - 45,
		64*6*6 - 45,
		64*6*7 + 90,
		0,
	}

	plzFrom     []byte
	plzFromMem  []byte
	plzFromBase int
	plzDsegBase int

	plzYY       int
	plzAX1      int32
	plzAX2      int32
	plzXX1      int32
	plzXX2      int32
	plzTxx1     int32
	plzTxy1     int32
	plzTax1     int32
	plzTay1     int32
	plzTxx2     int32
	plzTxy2     int32
	plzTax2     int32
	plzTay2     int32
	plzSini     [2000]int
	plzVramMem  []byte
	plzVramHalf []byte
	plzVramBase int
	plzKuva1    [64 * 256]byte
	plzKuva2    [64 * 256]byte
	plzKuva3    [64 * 256]byte
	plzDist1    [128 * 256]byte
	plzPal      [constants.PaletteByteCount]byte
	plzClrTau   [8][256][2]int16
	plzClrPtr   int
	plzVidMem   [2][plzSize * plzMaxY]byte
	plzPals     [6][constants.PaletteByteCount]int16
	plzCurPal   int
	plzTtptr    int
	plzL1       int
	plzL2       int
	plzL3       int
	plzL4       int
	plzK1       int
	plzK2       int
	plzK3       int
	plzK4       int
	plzIL1      int
	plzIL2      int
	plzIL3      int
	plzIL4      int
	plzIK1      int
	plzIK2      int
	plzIK3      int
	plzIK4      int
	plzFadePal  [2 * constants.PaletteByteCount]byte
	plzDropY    int
	plzLC1      [plzSize]int
	plzLC2      [plzSize]int
	plzLC3      [plzSize]int
	plzLC4      [plzSize]int
	plzFpal     [constants.PaletteByteCount]byte
	plzPolys    int
	plzLightSrc [6]int
	plzLLS      [6]int
	plzCxx      int32
	plzCxy      int32
	plzCxz      int32
	plzCyx      int32
	plzCyy      int32
	plzCyz      int32
	plzCzx      int32
	plzCzy      int32
	plzCzz      int32
	plzKx       int
	plzKy       int
	plzKz       int
	plzDis      int
	plzLtx      int
	plzTy       int
	plzLsKx     int
	plzLsKy     int
	plzLsKz     int
	plzLsX      int
	plzLsY      int
	plzLsZ      int
	plzPage     int
	plzFrames   int
	plzPtodraw  [256]plzPolygonsToDraw
	plzCube     = plzObject{
		name: "Cube",
		pnts: 8,
		point: [256]plzPoint{
			{x: 125, y: 125, z: 125},
			{x: 125, y: -125, z: 125},
			{x: -125, y: -125, z: 125},
			{x: -125, y: 125, z: 125},
			{x: 125, y: 125, z: -125},
			{x: 125, y: -125, z: -125},
			{x: -125, y: -125, z: -125},
			{x: -125, y: 125, z: -125},
		},
		faces: 6,
		pg: [256]plzPolygon{
			{p1: 1, p2: 2, p3: 3, p4: 0, color: 0},
			{p1: 7, p2: 6, p3: 5, p4: 4, color: 0},
			{p1: 0, p2: 4, p3: 5, p4: 1, color: 1},
			{p1: 1, p2: 5, p3: 6, p4: 2, color: 2},
			{p1: 2, p2: 6, p3: 7, p4: 3, color: 1},
			{p1: 3, p2: 7, p3: 4, p4: 0, color: 2},
		},
	}
)

func runPLZ() {
	if err := plzEnsureData(); err != nil {
		log.Printf("plz: %v", err)
		return
	}

	plzReset()

	shim.ClearScreen()

	plzInitVect()
	plzMain()
	plzVect()

	driver.ChangeMode(constants.ScreenWidth, constants.DoubleScreenHeight, shim.DefaultJSSS)
}

func plzSwapPage() {
	plzPage = (plzPage + 1) % 6
	switch plzPage {
	case 0:
		common.CopStart = 0xaa00 + 40
		common.CopScrl = 4
	case 1:
		common.CopStart = 0x0000 + 40
		common.CopScrl = 0
	case 2:
		common.CopStart = 0x5500 + 40
		common.CopScrl = 4
	case 3:
		common.CopStart = 0xaa00 + 40
		common.CopScrl = 0
	case 4:
		common.CopStart = 0x0000 + 40
		common.CopScrl = 4
	case 5:
		common.CopStart = 0x5500 + 40
		common.CopScrl = 0
	}
}

func plzInitParams() {
	plzL1 = plzIL1
	plzL2 = plzIL2
	plzL3 = plzIL3
	plzL4 = plzIL4
	plzK1 = plzIK1
	plzK2 = plzIK2
	plzK3 = plzIK3
	plzK4 = plzIK4
}

func plzMove() {
	plzK1 = (plzK1 - 3) & 0x0FFF
	plzK2 = (plzK2 - 2) & 0x0FFF
	plzK3 = (plzK3 + 1) & 0x0FFF
	plzK4 = (plzK4 + 2) & 0x0FFF

	plzL1 = (plzL1 - 1) & 0x0FFF
	plzL2 = (plzL2 - 2) & 0x0FFF
	plzL3 = (plzL3 + 2) & 0x0FFF
	plzL4 = (plzL4 + 3) & 0x0FFF
}

func plzDoDrop() {
	common.CopDrop = (common.CopDrop + 1) & 0xFFFF
	if common.CopDrop <= 64 {
		if len(plzDtau) > 0 {
			off := (common.CopDrop << 1) & 0xFFFF
			if off+1 < len(plzDtau)*2 {
				if int(off/2) < len(plzDtau) {
					plzDropY = int(plzDtau[off/2])
				}
			}
		}
		return
	}

	if common.CopDrop >= 256 {
		common.CopDrop = 0
		return
	}

	if common.CopDrop < 128 {
		if common.CopDrop > (64 + 32) {
			common.CopDrop = 0
			return
		}
	}

	common.CopPal = plzFadePal[:constants.PaletteByteCount]
	common.DoPal = 1

	if common.CopDrop == 65 {
		if len(plzDtau) > 0 {
			plzDropY = int(plzDtau[0])
		}
		plzInitParams()
		return
	}

	if len(plzDtau) > 0 {
		plzDropY = int(plzDtau[0])
	}

	if common.CopFadePal == nil || len(common.CopFadePal) < constants.PaletteByteCount {
		return
	}
	accHi := plzFadePal[:constants.PaletteByteCount]
	accLo := plzFadePal[constants.PaletteByteCount:]
	for i := 0; i < constants.PaletteByteCount; i++ {
		ax := uint16(common.CopFadePal[i])
		al := byte(ax & 0xFF)
		ah := byte(ax >> 8)

		lo := uint16(accLo[i]) + uint16(al)
		accLo[i] = byte(lo & 0xFF)

		hi := uint16(accHi[i]) + uint16(ah) + uint16(lo>>8)
		accHi[i] = byte(hi & 0xFF)
	}
}

func plzRd16(buf []byte, offset int) uint16 {
	idx := offset & (plzSinBufferSize - 1)
	idx &^= 1
	if idx+1 >= len(buf) {
		return 0
	}
	return uint16(buf[idx]) | uint16(buf[idx+1])<<8
}

func plzRd8(buf []byte, offset int) byte {
	idx := offset & (plzSinBufferSize - 1)
	if idx < 0 || idx >= len(buf) {
		return 0
	}
	return buf[idx]
}

func plzCCCFromK(k int) int {
	return (k &^ 3) + (3 - (k & 3))
}

func plzSetParas(c1, c2, c3, c4 int) {
	C1 := uint32(uint16(c1))
	C2 := uint32(uint16(c2))
	C3 := uint32(uint16(c3))
	C4 := uint32(uint16(c4))
	for ccc := 0; ccc < plzSize; ccc++ {
		plzLC1[ccc] = int(C1 + uint32(8*ccc))
		plzLC2[ccc] = int((C2 << 1) + uint32(80*8-8*ccc))
		plzLC3[ccc] = int(C3 + uint32(80*4-4*ccc))
		plzLC4[ccc] = int((C4 << 1) + uint32(32*ccc))
	}
}

func plzLine(y int, vseg []byte) {
	y2 := uint32(uint16(y)) << 1
	var out [plzSize]byte
	for k := 0; k < plzSize; k++ {
		ccc := plzCCCFromK(k)
		bx1 := plzRd16(plzLsini16, plzLC2[ccc]+int(y2))
		s1 := plzRd8(plzPSini, plzLC1[ccc]+int(bx1))
		bx2 := plzRd16(plzLsini4, plzLC4[ccc]+int(y2))
		s2 := plzRd8(plzPSini, plzLC3[ccc]+int(bx2)+int(y2))
		out[ccc] = byte(uint16(s1) + uint16(s2))
	}
	if len(vseg) >= plzSize {
		copy(vseg[:plzSize], out[:])
	}
}

func plzInitPlz() {
	common.CopStart = 96 * (682 - 400)

	for a := 0; a < constants.PaletteColorCount; a++ {
		shim.SetPal(a, 63, 63, 63)
	}

	pptr := 3
	for a := 1; a < 64; a++ {
		plzPals[0][pptr+0] = int16(plzPTauVal(a))
		plzPals[0][pptr+1] = int16(plzPTauVal(0))
		plzPals[0][pptr+2] = int16(plzPTauVal(0))
		pptr += 3
	}
	for a := 0; a < 64; a++ {
		plzPals[0][pptr+0] = int16(plzPTauVal(63 - a))
		plzPals[0][pptr+1] = int16(plzPTauVal(0))
		plzPals[0][pptr+2] = int16(plzPTauVal(0))
		pptr += 3
	}
	for a := 0; a < 64; a++ {
		plzPals[0][pptr+0] = int16(plzPTauVal(0))
		plzPals[0][pptr+1] = int16(plzPTauVal(0))
		plzPals[0][pptr+2] = int16(plzPTauVal(a))
		pptr += 3
	}
	for a := 0; a < 64; a++ {
		plzPals[0][pptr+0] = int16(plzPTauVal(a))
		plzPals[0][pptr+1] = int16(plzPTauVal(0))
		plzPals[0][pptr+2] = int16(plzPTauVal(63 - a))
		pptr += 3
	}

	pptr = 3
	for a := 1; a < 64; a++ {
		plzPals[1][pptr+0] = int16(plzPTauVal(a))
		plzPals[1][pptr+1] = int16(plzPTauVal(0))
		plzPals[1][pptr+2] = int16(plzPTauVal(0))
		pptr += 3
	}
	for a := 0; a < 64; a++ {
		plzPals[1][pptr+0] = int16(plzPTauVal(63 - a))
		plzPals[1][pptr+1] = int16(plzPTauVal(0))
		plzPals[1][pptr+2] = int16(plzPTauVal(a))
		pptr += 3
	}
	for a := 0; a < 64; a++ {
		plzPals[1][pptr+0] = int16(plzPTauVal(0))
		plzPals[1][pptr+1] = int16(plzPTauVal(a))
		plzPals[1][pptr+2] = int16(plzPTauVal(63 - a))
		pptr += 3
	}
	for a := 0; a < 64; a++ {
		plzPals[1][pptr+0] = int16(plzPTauVal(a))
		plzPals[1][pptr+1] = int16(plzPTauVal(63))
		plzPals[1][pptr+2] = int16(plzPTauVal(a))
		pptr += 3
	}

	pptr = 3
	for a := 1; a < 64; a++ {
		plzPals[3][pptr+0] = int16(plzPTauVal(a))
		plzPals[3][pptr+1] = int16(plzPTauVal(0))
		plzPals[3][pptr+2] = int16(plzPTauVal(0))
		pptr += 3
	}
	for a := 0; a < 64; a++ {
		plzPals[3][pptr+0] = int16(plzPTauVal(63))
		plzPals[3][pptr+1] = int16(plzPTauVal(a))
		plzPals[3][pptr+2] = int16(plzPTauVal(a))
		pptr += 3
	}
	for a := 0; a < 64; a++ {
		plzPals[3][pptr+0] = int16(plzPTauVal(63 - a))
		plzPals[3][pptr+1] = int16(plzPTauVal(63 - a))
		plzPals[3][pptr+2] = int16(plzPTauVal(63))
		pptr += 3
	}
	for a := 0; a < 64; a++ {
		plzPals[3][pptr+0] = int16(plzPTauVal(0))
		plzPals[3][pptr+1] = int16(plzPTauVal(0))
		plzPals[3][pptr+2] = int16(plzPTauVal(63))
		pptr += 3
	}

	pptr = 3
	for a := 1; a < 64; a++ {
		plzPals[2][pptr+0] = int16(plzPTauVal(0) / 2)
		plzPals[2][pptr+1] = int16(plzPTauVal(0) / 2)
		plzPals[2][pptr+2] = int16(plzPTauVal(0) / 2)
		pptr += 3
	}
	for a := 0; a < 64; a++ {
		plzPals[2][pptr+0] = int16(plzPTauVal(a) / 2)
		plzPals[2][pptr+1] = int16(plzPTauVal(a) / 2)
		plzPals[2][pptr+2] = int16(plzPTauVal(a) / 2)
		pptr += 3
	}
	for a := 0; a < 64; a++ {
		plzPals[2][pptr+0] = int16(plzPTauVal(63-a) / 2)
		plzPals[2][pptr+1] = int16(plzPTauVal(63-a) / 2)
		plzPals[2][pptr+2] = int16(plzPTauVal(63-a) / 2)
		pptr += 3
	}
	for a := 0; a < 64; a++ {
		plzPals[2][pptr+0] = int16(plzPTauVal(0) / 2)
		plzPals[2][pptr+1] = int16(plzPTauVal(0) / 2)
		plzPals[2][pptr+2] = int16(plzPTauVal(0) / 2)
		pptr += 3
	}

	pptr = 3
	for a := 1; a < 75; a++ {
		v := plzPTauVal(63 - a*64/75)
		plzPals[4][pptr+0] = int16(v)
		plzPals[4][pptr+1] = int16(v)
		plzPals[4][pptr+2] = int16(v)
		pptr += 3
	}
	for a := 0; a < 106; a++ {
		plzPals[4][pptr+0] = 0
		plzPals[4][pptr+1] = 0
		plzPals[4][pptr+2] = 0
		pptr += 3
	}
	for a := 0; a < 75; a++ {
		v := plzPTauVal(a * 64 / 75)
		plzPals[4][pptr+0] = int16(v * 8 / 10)
		plzPals[4][pptr+1] = int16(v * 9 / 10)
		plzPals[4][pptr+2] = int16(v)
		pptr += 3
	}

	for i := 0; i < constants.PaletteByteCount; i++ {
		v := int32(plzPals[0][i])
		v = (v - 63) * 2
		plzPals[0][i] = int16(v)
	}
	for pal := 1; pal < 5; pal++ {
		for i := 0; i < constants.PaletteByteCount; i++ {
			v := int32(plzPals[pal][i])
			plzPals[pal][i] = int16(v * 8)
		}
	}
}

func plzMain() {
	driver.ChangeMode(constants.ScreenWidth, constants.DoubleScreenHeight, 0)

	if !shim.IsDemoFirstPart() {
		for !driver.WantsToQuit() && music.GetPlusFlags() < 0 {
			driver.Vsync(true)
		}
	}

	music.SetFrame(0)

	plzInitPlz()
	common.CopDrop = 128
	clear(common.FadePal[:])
	clear(common.FadePalShort[:])
	if plzCurPal < len(plzPals) {
		common.CopFadePal = plzPals[plzCurPal][:]
		plzCurPal++
	}
	if len(plzDtau) > 0 {
		plzDropY = int(plzDtau[0])
	}
	common.FrameCount = 0
	plzInitParams()

	for !driver.WantsToQuit() {
		common.FrameCount = 0

		if plzTtptr < len(plzTimetable) && music.GetFrame() > plzTimetable[plzTtptr] {
			clear(plzFadePal[:constants.PaletteByteCount])
			common.CopDrop = 1
			if plzCurPal < len(plzPals) {
				common.CopFadePal = plzPals[plzCurPal][:]
				plzCurPal++
			}
			plzTtptr++
			if plzTtptr < len(plzInitTable) {
				plzIL1 = plzInitTable[plzTtptr][0]
				plzIL2 = plzInitTable[plzTtptr][1]
				plzIL3 = plzInitTable[plzTtptr][2]
				plzIL4 = plzInitTable[plzTtptr][3]
				plzIK1 = plzInitTable[plzTtptr][4]
				plzIK2 = plzInitTable[plzTtptr][5]
				plzIK3 = plzInitTable[plzTtptr][6]
				plzIK4 = plzInitTable[plzTtptr][7]
			}
		}

		if plzCurPal == 4 && common.CopDrop > 64 {
			break
		}

		plzSetParas(plzK1, plzK2, plzK3, plzK4)
		for y := 0; y < plzMaxY; y += 2 {
			offset := y*plzSize + plzYAdd*6
			plzLine(y, plzVidMem[0][offset:])
		}

		plzSetParas(plzL1, plzL2, plzL3, plzL4)
		for y := 1; y < plzMaxY; y += 2 {
			offset := y*plzSize + plzYAdd*6
			plzLine(y, plzVidMem[0][offset:])
		}

		plzSetParas(plzK1, plzK2, plzK3, plzK4)
		for y := 1; y < plzMaxY; y += 2 {
			offset := y*plzSize + plzYAdd*6
			plzLine(y, plzVidMem[1][offset:])
		}

		plzSetParas(plzL1, plzL2, plzL3, plzL4)
		for y := 0; y < plzMaxY; y += 2 {
			offset := y*plzSize + plzYAdd*6
			plzLine(y, plzVidMem[1][offset:])
		}

		if common.CopDrop != 0 {
			plzDoDrop()
		}

		common.Copper2()
		plzMove()

		clear(shim.VRAM[:constants.ScreenWidth*constants.DoubleScreenHeight])

		for row := 0; row < plzMaxY; row++ {
			if row+plzDropY >= constants.DoubleScreenHeight {
				break
			}
			dst := (row + plzDropY) * constants.ScreenWidth
			src1 := row*plzSize + 2
			src2 := row*plzSize + 2
			for x := 0; x < 80; x++ {
				s2 := plzVidMem[1][src2]
				s1 := plzVidMem[0][src1]
				shim.VRAM[dst+0] = s2
				shim.VRAM[dst+1] = s1
				shim.VRAM[dst+2] = s2
				shim.VRAM[dst+3] = s1
				dst += 4
				src1++
				src2++
			}
		}

		driver.Blit()
		common.FrameCount += driver.Vsync(false)
	}
}

func plzShadePal(destPal []byte, srcPal []byte, shade int) {
	dl := byte(shade)
	for i := 0; i < 192 && i < len(destPal) && i < len(srcPal); i++ {
		ax := uint16(srcPal[i]) * uint16(dl)
		ax >>= 6
		destPal[i] = byte(ax)
	}
}

func plzGetSpl(where int) {
	if len(plzSplineCoef) < 1024 || len(plzBuu) == 0 {
		return
	}
	idx4 := (where >> 8) & 0xFFFF
	t := where & 0xFF
	wofs := t

	w0 := plzSplineCoef[wofs+0:]
	w1 := plzSplineCoef[wofs+256:]
	w2 := plzSplineCoef[wofs+512:]
	w3 := plzSplineCoef[wofs+768:]

	var out [8]int16

	for comp := 0; comp < 8; comp++ {
		i0 := idx4 + 0
		i1 := idx4 + 1
		i2 := idx4 + 2
		i3 := idx4 + 3
		if i3 >= len(plzBuu) {
			break
		}
		a3 := int32(plzBuu[i3][comp])
		a2 := int32(plzBuu[i2][comp])
		a1 := int32(plzBuu[i1][comp])
		a0 := int32(plzBuu[i0][comp])

		s := int32(0)
		s += a3 * int32(w0[0])
		s += a2 * int32(w1[0])
		s += a1 * int32(w2[0])
		s += a0 * int32(w3[0])
		if s < 0 {
			s += (1 << 15) - 1
		}
		out[comp] = int16(s >> 15)
	}

	plzLtx = int(out[0])
	plzTy = int(out[1])
	plzDis = int(out[2])
	plzKx = int(out[3])
	plzKy = int(out[4])
	plzKz = int(out[5])
	plzLsKx = int(out[6])
	plzLsKy = int(out[7])
}

func plzCountConst() {
	sx := plzSinitAt(plzKx)
	sy := plzSinitAt(plzKy)
	sz := plzSinitAt(plzKz)
	cx := plzKosinitAt(plzKx)
	cy := plzKosinitAt(plzKy)
	cz := plzKosinitAt(plzKz)

	plzCxx = (cy * cz) >> 22
	plzCxy = (cy * sz) >> 22
	plzCxz = -(sy) >> 7

	plzCyx = ((((sx*cz + 16384) >> 15) * sy) - (cx * sz)) >> 22
	plzCyy = ((((sx*sy + 16384) >> 15) * sz) + (cx * cz)) >> 22
	plzCyz = (cy * sx) >> 22

	plzCzx = ((((cx*cz + 16384) >> 15) * sy) + (sx * sz)) >> 22
	plzCzy = ((((cx*sy + 16384) >> 15) * sz) - (sx * cz)) >> 22
	plzCzz = (cy * cx) >> 22
}

func plzRotate() {
	for a := 0; a < plzCube.pnts; a++ {
		x := plzCube.point[a].x
		y := plzCube.point[a].y
		z := plzCube.point[a].z

		xx := int((((int32(x)*plzCxx)>>1 + (int32(y)*plzCxy)>>1 + (int32(z)*plzCxz)>>1) >> 7)) + plzLtx
		yy := int((((int32(x)*plzCyx)>>1 + (int32(y)*plzCyy)>>1 + (int32(z)*plzCyz)>>1) >> 7)) + plzTy
		zz := int((((int32(x)*plzCzx)>>1 + (int32(y)*plzCzy)>>1 + (int32(z)*plzCzz)>>1) >> 7)) + plzDis

		plzCube.point[a].xx = xx
		plzCube.point[a].yy = yy
		plzCube.point[a].zz = zz
		if zz == 0 {
			zz = 1
		}
		plzCube.point[a].xxx = (xx*256)/zz + 160 + 160
		plzCube.point[a].yyy = (yy*142)/zz + 66
	}
}

func plzSortFaces() {
	a := 0
	p := 0
	for a < plzCube.faces {
		x := int32(plzCube.point[plzCube.pg[a].p1].xx)
		y := int32(plzCube.point[plzCube.pg[a].p1].yy)
		z := int32(plzCube.point[plzCube.pg[a].p1].zz)

		ax := int32(plzCube.point[plzCube.pg[a].p2].xx) - x
		ay := int32(plzCube.point[plzCube.pg[a].p2].yy) - y
		az := int32(plzCube.point[plzCube.pg[a].p2].zz) - z

		bx := int32(plzCube.point[plzCube.pg[a].p3].xx) - x
		by := int32(plzCube.point[plzCube.pg[a].p3].yy) - y
		bz := int32(plzCube.point[plzCube.pg[a].p3].zz) - z

		nx := ay*bz - az*by
		ny := az*bx - ax*bz
		nz := ax*by - ay*bx

		lkx := -x
		lky := -y
		lkz := -z

		s := lkx*nx + lky*ny + lkz*nz
		if s > 0 {
			a++
			continue
		}

		s = (int32(plzLsX)*nx+int32(plzLsY)*ny+int32(plzLsZ)*nz)/250000 + 32
		plzLightSrc[p] = int(s)
		c := plzCube.pg[a].color
		if plzLLS[p] != plzLightSrc[p] {
			start := c * 64 * 3
			plzShadePal(plzFpal[start:], plzPal[start:], plzLightSrc[p])
			plzLLS[p] = plzLightSrc[p]
		}

		plzPtodraw[p].p = a
		p++
		a++
	}
	plzPolys = p
}

func plzStartMask4(m uint) uint8 {
	switch m & 3 {
	case 0:
		return 0x0F
	case 1:
		return 0x0E
	case 2:
		return 0x0C
	default:
		return 0x08
	}
}

func plzEndMask4(m uint) uint8 {
	switch m & 3 {
	case 0:
		return 0x01
	case 1:
		return 0x03
	case 2:
		return 0x07
	default:
		return 0x0F
	}
}

func plzPoke4Full(dst []byte, c byte) {
	if len(dst) < 4 {
		return
	}
	dst[0] = c
	dst[1] = c
	dst[2] = c
	dst[3] = c
}

func plzPoke4Mask(dst []byte, c byte, m uint8) {
	if len(dst) < 4 {
		return
	}
	if m&1 != 0 {
		dst[0] = c
	}
	if m&2 != 0 {
		dst[1] = c
	}
	if m&4 != 0 {
		dst[2] = c
	}
	if m&8 != 0 {
		dst[3] = c
	}
}

func plzPoke4MaskClip(lineBase int, x int, c byte, m uint8) {
	if m == 0 {
		return
	}
	for i := 0; i < 4; i++ {
		if m&(1<<uint(i)) == 0 {
			continue
		}
		idx := plzVramBase + lineBase + x + i
		if idx >= 0 && idx < len(plzVramMem) {
			plzVramMem[idx] = c
		}
	}
}

func plzPoke4FullClip(lineBase int, x int, c byte) {
	plzPoke4MaskClip(lineBase, x, c, 0x0F)
}

func plzBxFromTxy(tX, tY uint32) uint16 {
	return uint16((((tY >> 16) & 0xFF) << 8) | ((tX >> 16) & 0xFF))
}

func plzDrawSpan(y int, xStart, xEnd int, uStart, vStart, uStep, vStep int32) {
	if y < 0 || y >= plzVmemHeight {
		return
	}
	if xStart > xEnd {
		return
	}
	ctau := &plzClrTau[plzClrPtr][y]
	if int16(xStart) <= ctau[0] {
		ctau[0] = int16(xStart)
	}
	if int16(xEnd) >= ctau[1] {
		ctau[1] = int16(xEnd)
	}

	if xStart < 0 {
		delta := -xStart
		uStart += uStep * int32(delta)
		vStart += vStep * int32(delta)
		xStart = 0
	}
	if xEnd >= plzVmemWidth {
		xEnd = plzVmemWidth - 1
	}
	if xStart > xEnd {
		return
	}
	lineBase := plzVramBase + y*plzVmemWidth
	u := uStart
	v := vStart
	for x := xStart; x <= xEnd; x++ {
		c := plzSampleBX(plzBxFromTxy(uint32(u), uint32(v)))
		idx := lineBase + x
		if idx >= 0 && idx < len(plzVramMem) {
			plzVramMem[idx] = c
		}
		u += uStep
		v += vStep
	}
}

func plzDrawTrapezoid(yStart, yEnd int, xL, xR, uL, vL, uR, vR int32, dxL, dxR, duL, dvL, duR, dvR int32) {
	if yStart < 0 {
		delta := -yStart
		xL += dxL * int32(delta)
		uL += duL * int32(delta)
		vL += dvL * int32(delta)
		xR += dxR * int32(delta)
		uR += duR * int32(delta)
		vR += dvR * int32(delta)
		yStart = 0
	}
	if yEnd > plzVmemHeight {
		yEnd = plzVmemHeight
	}
	for y := yStart; y < yEnd; y++ {
		xStart := int(xL >> 16)
		xEnd := int(xR >> 16)
		if xStart > xEnd {
			xStart, xEnd = xEnd, xStart
			uL, uR = uR, uL
			vL, vR = vR, vL
		}
		span := xEnd - xStart
		uStep := int32(0)
		vStep := int32(0)
		if span > 0 {
			uStep = int32(int64(uR-uL) / int64(span))
			vStep = int32(int64(vR-vL) / int64(span))
		}
		plzDrawSpan(y, xStart, xEnd, uL, vL, uStep, vStep)
		xL += dxL
		uL += duL
		vL += dvL
		xR += dxR
		uR += duR
		vR += dvR
	}
}

func plzDrawFlatBottom(v0, v1, v2 plzVert) {
	if v1.x > v2.x {
		v1, v2 = v2, v1
	}
	dy1 := v1.y - v0.y
	dy2 := v2.y - v0.y
	if dy1 == 0 || dy2 == 0 {
		return
	}
	dxL := int32(int64(v1.xf-v0.xf) / int64(dy1))
	dxR := int32(int64(v2.xf-v0.xf) / int64(dy2))
	duL := int32(int64(v1.u-v0.u) / int64(dy1))
	dvL := int32(int64(v1.v-v0.v) / int64(dy1))
	duR := int32(int64(v2.u-v0.u) / int64(dy2))
	dvR := int32(int64(v2.v-v0.v) / int64(dy2))

	plzDrawTrapezoid(v0.y, v1.y, v0.xf, v0.xf, v0.u, v0.v, v0.u, v0.v, dxL, dxR, duL, dvL, duR, dvR)
}

func plzDrawFlatTop(v0, v1, v2 plzVert) {
	if v0.x > v1.x {
		v0, v1 = v1, v0
	}
	dy1 := v2.y - v0.y
	dy2 := v2.y - v1.y
	if dy1 == 0 || dy2 == 0 {
		return
	}
	dxL := int32(int64(v2.xf-v0.xf) / int64(dy1))
	dxR := int32(int64(v2.xf-v1.xf) / int64(dy2))
	duL := int32(int64(v2.u-v0.u) / int64(dy1))
	dvL := int32(int64(v2.v-v0.v) / int64(dy1))
	duR := int32(int64(v2.u-v1.u) / int64(dy2))
	dvR := int32(int64(v2.v-v1.v) / int64(dy2))

	plzDrawTrapezoid(v0.y, v2.y, v0.xf, v1.xf, v0.u, v0.v, v1.u, v1.v, dxL, dxR, duL, dvL, duR, dvR)
}

func plzDrawTri(v0, v1, v2 plzVert) {
	if v0.y > v1.y {
		v0, v1 = v1, v0
	}
	if v1.y > v2.y {
		v1, v2 = v2, v1
	}
	if v0.y > v1.y {
		v0, v1 = v1, v0
	}

	if v1.y == v0.y {
		plzDrawFlatTop(v0, v1, v2)
		return
	}
	if v1.y == v2.y {
		plzDrawFlatBottom(v0, v1, v2)
		return
	}

	dy := v2.y - v0.y
	if dy == 0 {
		return
	}
	alpha := int64(v1.y-v0.y) << 16 / int64(dy)
	split := plzVert{
		x:  0,
		y:  v1.y,
		xf: v0.xf + int32((int64(v2.xf-v0.xf)*alpha)>>16),
		u:  v0.u + int32((int64(v2.u-v0.u)*alpha)>>16),
		v:  v0.v + int32((int64(v2.v-v0.v)*alpha)>>16),
	}
	split.x = int(split.xf >> 16)

	plzDrawFlatBottom(v0, v1, split)
	plzDrawFlatTop(v1, split, v2)
}

func plzReadFromAbs(idx int) byte {
	if idx < 0 {
		return 0
	}
	switch {
	case idx < plzFromOffsetKuva2:
		return plzKuva1[idx]
	case idx < plzFromOffsetKuva3:
		return plzKuva2[idx-plzFromOffsetKuva2]
	case idx < plzFromOffsetDist1:
		return plzKuva3[idx-plzFromOffsetKuva3]
	case idx < plzFromOffsetSinx:
		return plzDist1[idx-plzFromOffsetDist1]
	case idx < plzFromOffsetSiny:
		return 0
	case idx < plzFromOffsetPal:
		return 0
	case idx < plzFromOffsetTau:
		return plzPal[idx-plzFromOffsetPal]
	case idx < plzFromOffsetClr:
		off := idx - plzFromOffsetTau
		elem := off >> 1
		byteSel := off & 1
		if elem < 0 {
			return 0
		}
		plane := elem / (256 * 2)
		rem := elem % (256 * 2)
		row := rem / 2
		slot := rem % 2
		if plane < 0 || plane >= len(plzClrTau) || row < 0 || row >= len(plzClrTau[0]) {
			return 0
		}
		val := uint16(plzClrTau[plane][row][slot])
		if byteSel == 0 {
			return byte(val)
		}
		return byte(val >> 8)
	case idx < plzFromOffsetVid:
		off := idx - plzFromOffsetClr
		clr := uint32(plzClrPtr)
		return byte(clr >> (8 * uint(off)))
	case idx < plzFromOffsetPals:
		off := idx - plzFromOffsetVid
		if off < plzSize*plzMaxY {
			return plzVidMem[0][off]
		}
		off -= plzSize * plzMaxY
		if off < plzSize*plzMaxY {
			return plzVidMem[1][off]
		}
		return 0
	case idx < plzFromOffsetFade:
		off := idx - plzFromOffsetPals
		elem := off >> 1
		byteSel := off & 1
		pal := elem / constants.PaletteByteCount
		i := elem % constants.PaletteByteCount
		if pal < 0 || pal >= len(plzPals) || i < 0 || i >= len(plzPals[pal]) {
			return 0
		}
		val := uint16(plzPals[pal][i])
		if byteSel == 0 {
			return byte(val)
		}
		return byte(val >> 8)
	case idx < plzFromMemSize:
		return plzFadePal[idx-plzFromOffsetFade]
	default:
		return 0
	}
}

func plzSampleBX(bx uint16) byte {
	idx := int(bx)
	rtIdx := plzFromOffsetDist1 + plzDsegBase + idx
	dist := plzReadFromAbs(rtIdx)
	srcIdx := plzFromBase + idx + int(dist)
	return plzReadFromAbs(srcIdx)
}

func plzDoBlock(ycount int16) {
	if ycount <= 0 {
		return
	}
	for ycount > 0 {
		if plzYY >= plzVmemHeight {
			break
		}
		if plzYY >= 0 {
			xL := int(int16(plzXX2 >> 16))
			xR := int(int16(plzXX1 >> 16))

			txL := plzTxx2
			tyL := plzTxy2
			txR := plzTxx1
			tyR := plzTxy1
			if xL > xR {
				xL, xR = xR, xL
				txL, txR = txR, txL
				tyL, tyR = tyR, tyL
			}

			ctau := &plzClrTau[plzClrPtr][plzYY]
			if int16(xL) <= ctau[0] {
				ctau[0] = int16(xL)
			}
			if int16(xR) >= ctau[1] {
				ctau[1] = int16(xR)
			}

			xLAln := int(xL) &^ 3
			leftCell := xLAln >> 2
			rightCell := (int(xR) - 1) >> 2
			cells := rightCell - leftCell

			if cells >= 0 {
				lineBase := plzYY * plzVmemWidth
				if cells == 0 {
					c := plzSampleBX(plzBxFromTxy(uint32(txL), uint32(tyL)))
					mL := plzStartMask4(uint(xL) & 3)
					mR := plzEndMask4(uint(int(xR)-1) & 3)
					plzPoke4MaskClip(lineBase, xLAln, c, mL&mR)
				} else if cells == 1 {
					cL := plzSampleBX(plzBxFromTxy(uint32(txL), uint32(tyL)))
					cR := plzSampleBX(plzBxFromTxy(uint32(txR), uint32(tyR)))
					mL := plzStartMask4(uint(xL) & 3)
					mR := plzEndMask4(uint(int(xR)-1) & 3)
					plzPoke4MaskClip(lineBase, xLAln, cL, mL)
					plzPoke4MaskClip(lineBase, xLAln+4, cR, mR)
				} else {
					cL := plzSampleBX(plzBxFromTxy(uint32(txL), uint32(tyL)))
					mL := plzStartMask4(uint(xL) & 3)
					plzPoke4MaskClip(lineBase, xLAln, cL, mL)

					interior := cells - 1
					if interior > 0 {
						dX := int32(txR - txL)
						dY := int32(tyR - tyL)
						stepX := dX / int32(cells)
						stepY := dY / int32(cells)
						tx := txL
						ty := tyL
						for i := 1; i <= interior; i++ {
							tx += stepX
							ty += stepY
							plzPoke4FullClip(lineBase, xLAln+4*i, plzSampleBX(plzBxFromTxy(uint32(tx), uint32(ty))))
						}
					}

					cR := plzSampleBX(plzBxFromTxy(uint32(txR), uint32(tyR)))
					mR := plzEndMask4(uint(int(xR)-1) & 3)
					plzPoke4MaskClip(lineBase, xLAln+4*cells, cR, mR)
				}
			}
		}

		plzYY++
		plzXX1 += plzAX1
		plzXX2 += plzAX2
		plzTxy1 += plzTay1
		plzTxx1 += plzTax1
		plzTxy2 += plzTay2
		plzTxx2 += plzTax2
		ycount--
	}
}

func plzDoPoly(x1, y1, x2, y2, x3, y3, x4, y4, color, dd int) {
	type pt struct{ x, y int }
	pnts := [4]pt{{x1, y1}, {x2, y2}, {x3, y3}, {x4, y4}}
	txt := [4]pt{{64, 4}, {190, 4}, {190, 60}, {64, 60}}

	dd = (dd + 1) & 63
	plzTextureForColor(color)
	plzDsegBase = dd * 16 * 16

	verts := [4]plzVert{}
	for i := 0; i < 4; i++ {
		verts[i] = plzVert{
			x:  pnts[i].x,
			y:  pnts[i].y,
			xf: int32(pnts[i].x<<16) + 0x8000,
			u:  int32(txt[i].x<<16) + 0x8000,
			v:  int32(txt[i].y<<16) + 0x8000,
		}
	}

	plzDrawTri(verts[0], verts[1], verts[2])
	plzDrawTri(verts[0], verts[2], verts[3])
}

func plzDoClear(mem []byte, otau, ntau *[256][2]int16) {
	base := 0
	for ycnt := 134; ; ycnt-- {
		row := 134 - ycnt
		old0 := uint16(otau[row][0])
		old1 := uint16(otau[row][1])
		new0 := uint16(ntau[row][0])
		new1 := uint16(ntau[row][1])

		if old0 != 640 {
			if new0 >= old0 {
				length := uint16(new0 - old0)
				if length != 0 {
					start := base + int(old0)
					if start >= 0 && start < len(mem) {
						end := start + int(length)
						if end > len(mem) {
							end = len(mem)
						}
						clear(mem[start:end])
					}
				}
			}
			if old1 >= new1 {
				a := uint16(new1 + 1)
				b := uint16(old1 + 1)
				length := uint16(b - a)
				if length != 0 {
					start := base + int(a)
					if start >= 0 && start < len(mem) {
						end := start + int(length)
						if end > len(mem) {
							end = len(mem)
						}
						clear(mem[start:end])
					}
				}
			}
		}

		otau[row][0] = 640
		otau[row][1] = 0
		base += plzVmemWidth
		if ycnt == 0 {
			break
		}
	}
}

func plzClear() {
	oldPtr := (plzClrPtr - 1) & 7
	plzDoClear(plzVramHalf[:], &plzClrTau[oldPtr], &plzClrTau[plzClrPtr])
	plzClrPtr = (plzClrPtr + 1) & 7
}

func plzDraw() {
	for a := 0; a < plzPolys; a++ {
		c := plzCube.pg[plzPtodraw[a].p].color
		plzDoPoly(
			plzCube.point[plzCube.pg[plzPtodraw[a].p].p1].xxx,
			plzCube.point[plzCube.pg[plzPtodraw[a].p].p1].yyy,
			plzCube.point[plzCube.pg[plzPtodraw[a].p].p2].xxx,
			plzCube.point[plzCube.pg[plzPtodraw[a].p].p2].yyy,
			plzCube.point[plzCube.pg[plzPtodraw[a].p].p3].xxx,
			plzCube.point[plzCube.pg[plzPtodraw[a].p].p3].yyy,
			plzCube.point[plzCube.pg[plzPtodraw[a].p].p4].xxx,
			plzCube.point[plzCube.pg[plzPtodraw[a].p].p4].yyy,
			c,
			plzFrames&63,
		)
	}
}

func plzCalculate() {
	plzGetSpl(4*256 + plzFrames*4)

	plzKx &= 1023
	plzKy &= 1023
	plzKz &= 1023
	plzLsKx &= 1023
	plzLsKy &= 1023

	plzLsY = int(plzKosinitAt(plzLsKx) >> 8)
	plzLsX = int((plzSinitAt(plzLsKx) >> 8) * (plzSinitAt(plzLsKy) >> 8) >> 7)
	plzLsZ = int((plzSinitAt(plzLsKx) >> 8) * (plzKosinitAt(plzLsKy) >> 8) >> 7)

	plzCountConst()
	plzRotate()
	plzSortFaces()
}

func plzVect() {
	driver.ChangeMode(constants.ScreenWidth, constants.DoubleScreenHeight, 0)

	if !shim.IsDemoFirstPart() {
		wait := 0
		for !driver.WantsToQuit() && music.GetPlusFlags() < 13 && wait < 70 {
			driver.Blit()
			wait += driver.Vsync(true)
		}
		common.FrameCount = 0
	}

	for !driver.WantsToQuit() {
		a := music.GetPlusFlags()
		if a >= -4 && a < 0 {
			break
		}
		plzSwapPage()
		plzFrames += driver.Vsync(false)
		common.FrameCount = 0
		common.CopPal = plzFpal[:]
		common.DoPal = 1

		common.Copper2()
		common.Copper3()

		plzCalculate()
		plzDraw()

		src := 160
		dst := 0
		for i := 0; i < plzVmemHeight; i++ {
			if dst+constants.ScreenWidth > len(shim.VRAM) || src+constants.ScreenWidth > len(plzVramHalf) {
				break
			}
			copy(shim.VRAM[dst:dst+constants.ScreenWidth], plzVramHalf[src:src+constants.ScreenWidth])
			dst += constants.ScreenWidth
			if dst+constants.ScreenWidth > len(shim.VRAM) {
				break
			}
			copy(shim.VRAM[dst:dst+constants.ScreenWidth], plzVramHalf[src:src+constants.ScreenWidth])
			dst += constants.ScreenWidth
			if dst+constants.ScreenWidth > len(shim.VRAM) {
				break
			}
			copy(shim.VRAM[dst:dst+constants.ScreenWidth], plzVramHalf[src:src+constants.ScreenWidth])
			dst += constants.ScreenWidth
			src += plzVmemWidth
		}

		driver.Blit()
		plzClear()
	}
}

func plzInitVect() {
	for a := 0; a < 1524; a++ {
		plzSini[a] = int(math.Sin(float64(a)/1024.0*math.Pi*4.0) * 127)
	}

	for a := 1; a < 32; a++ {
		plzPal[0*192+a*3+0] = 0
		plzPal[0*192+a*3+1] = 0
		plzPal[0*192+a*3+2] = byte(a * 2)
	}
	for a := 0; a < 32; a++ {
		plzPal[0*192+a*3+32*3+0] = byte(a * 2)
		plzPal[0*192+a*3+32*3+1] = byte(a * 2)
		plzPal[0*192+a*3+32*3+2] = 63
	}

	for a := 0; a < 32; a++ {
		plzPal[1*192+a*3+0] = byte(a * 2)
		plzPal[1*192+a*3+1] = 0
		plzPal[1*192+a*3+2] = 0
	}
	for a := 0; a < 32; a++ {
		plzPal[1*192+a*3+32*3+0] = 63
		plzPal[1*192+a*3+32*3+1] = byte(a * 2)
		plzPal[1*192+a*3+32*3+2] = 0
	}

	for a := 0; a < 32; a++ {
		plzPal[2*192+a*3+0] = byte(a)
		plzPal[2*192+a*3+1] = 0
		plzPal[2*192+a*3+2] = byte(a * 2 / 3)
	}
	for a := 0; a < 32; a++ {
		plzPal[2*192+a*3+32*3+0] = byte(31 - a)
		plzPal[2*192+a*3+32*3+1] = byte(a * 2)
		plzPal[2*192+a*3+32*3+2] = 21
	}

	for y := 0; y < 64; y++ {
		for x := 0; x < 256; x++ {
			idx := (y*4 + plzSini[x*2]) & 511
			plzKuva1[y*256+x] = byte(int8(plzSini[idx]/4 + 32))
			plzKuva2[y*256+x] = byte(int8(plzSini[idx]/4 + 32 + 64))
			plzKuva3[y*256+x] = byte(int8(plzSini[idx]/4 + 32 + 128))
		}
	}

	for y := 0; y < 128; y++ {
		for x := 0; x < 256; x++ {
			val := plzSini[y*8] / 3
			plzDist1[y*256+x] = byte(int8(val))
		}
	}

	for buf := 0; buf < 8; buf++ {
		for row := 0; row < 256; row++ {
			plzClrTau[buf][row][0] = 640
			plzClrTau[buf][row][1] = 0
		}
	}
}

func plzReset() {
	plzFrom = nil
	plzDsegBase = 0
	plzYY = 0
	plzAX1 = 0
	plzAX2 = 0
	plzXX1 = 0
	plzXX2 = 0
	plzTxx1 = 0
	plzTxy1 = 0
	plzTax1 = 0
	plzTay1 = 0
	plzTxx2 = 0
	plzTxy2 = 0
	plzTax2 = 0
	plzTay2 = 0
	if plzVramMem == nil || len(plzVramMem) != plzVmemWidth*plzVmemHeight+2*plzVramPad {
		plzVramMem = make([]byte, plzVmemWidth*plzVmemHeight+2*plzVramPad)
		plzVramBase = plzVramPad
		plzVramHalf = plzVramMem[plzVramBase : plzVramBase+plzVmemWidth*plzVmemHeight]
	}
	clear(plzSini[:])
	clear(plzVramHalf)
	clear(plzKuva1[:])
	clear(plzKuva2[:])
	clear(plzKuva3[:])
	clear(plzDist1[:])
	clear(plzPal[:])
	clear(plzClrTau[:])
	plzClrPtr = 0
	clear(plzVidMem[:])
	clear(plzPals[:])
	plzCurPal = 0
	plzTtptr = 0
	plzL1, plzL2, plzL3, plzL4 = 1000, 2000, 3000, 4000
	plzK1, plzK2, plzK3, plzK4 = 3500, 2300, 3900, 3670
	plzIL1, plzIL2, plzIL3, plzIL4 = 1000, 2000, 3000, 4000
	plzIK1, plzIK2, plzIK3, plzIK4 = 3500, 2300, 3900, 3670
	clear(plzFadePal[:])
	plzDropY = 0
	clear(plzLC1[:])
	clear(plzLC2[:])
	clear(plzLC3[:])
	clear(plzLC4[:])
	clear(plzFpal[:])
	plzPolys = 0
	clear(plzLightSrc[:])
	clear(plzLLS[:])
	plzCxx, plzCxy, plzCxz = 0, 0, 0
	plzCyx, plzCyy, plzCyz = 0, 0, 0
	plzCzx, plzCzy, plzCzz = 0, 0, 0
	plzKx, plzKy, plzKz, plzDis, plzLtx, plzTy = 0, 0, 0, 320, 0, -50
	plzLsKx, plzLsKy, plzLsKz, plzLsX, plzLsY, plzLsZ = 0, 0, 0, 0, 0, 128
	plzPage = 0
	plzFrames = 0
	clear(plzPtodraw[:])
}

func plzTextureForColor(color int) []byte {
	switch color {
	case 0:
		plzFromBase = plzFromOffsetKuva1
		return plzKuva1[:]
	case 1:
		plzFromBase = plzFromOffsetKuva2
		return plzKuva2[:]
	case 2:
		plzFromBase = plzFromOffsetKuva3
		return plzKuva3[:]
	default:
		plzFromBase = plzFromOffsetKuva1
		return plzKuva1[:]
	}
}

func plzPTauVal(idx int) int {
	if idx < 0 || idx >= len(plzPTau) {
		return 0
	}
	return int(plzPTau[idx])
}

func plzSinitAt(idx int) int32 {
	if len(plzSinit) == 0 {
		return 0
	}
	if idx < 0 {
		idx = -idx
	}
	if idx >= len(plzSinit) {
		idx = idx % len(plzSinit)
	}
	return int32(plzSinit[idx])
}

func plzKosinitAt(idx int) int32 {
	if len(plzKosinit) == 0 {
		return 0
	}
	if idx < 0 {
		idx = -idx
	}
	if idx >= len(plzKosinit) {
		idx = idx % len(plzKosinit)
	}
	return int32(plzKosinit[idx])
}
