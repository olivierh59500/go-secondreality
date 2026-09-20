package parts

import (
	"log"

	"go-secondreality/dck/internal/blob"
	"go-secondreality/dck/internal/common"
	"go-secondreality/dck/internal/constants"
	"go-secondreality/dck/internal/driver"
	"go-secondreality/dck/internal/music"
	"go-secondreality/dck/internal/shim"
)

const (
	koeWMinY = 0
	koeWMaxY = constants.ScreenHeight - 1
	koeWMinX = 0
	koeWMaxX = constants.ScreenWidth - 1

	koePitch    = constants.PlanarWidth * constants.DoubleScreenHeight
	koeVbufSize = 8000
)

type koeXY struct {
	x int16
	y int16
}

type koeRing struct {
	x int32
	y int32
	c uint8
}

var (
	koeClip1 koeXY
	koeClip2 koeXY

	koePolyISides uint16
	koePolySides  uint16

	koeClipXY2 [32]koeXY
	koePolyIXY [16]koeXY
	koePolyXY  [16]koeXY

	koeClipLeft int16

	koeFrameCountA uint16
	koePalAnimCA   uint16
	koePalAnimCA2  uint16
	koeScrnPosA    uint16
	koeScrnPosAl   uint16
	koeScrnXA      uint16
	koeScrnYA      uint16
	koeScrnRotA    uint16
	koeSinuRotA    uint16
	koeOverRotA    uint16 = 211
	koeOverXA      uint16
	koeOverYA      uint16
	koePatDirA     int16

	koeSizeFade  uint16
	koeRotSpeed  uint16
	koePalFader  uint16
	koePalFader2 uint8 = 255
	koeZumPlane  uint8 = 0x11

	koeSinusPower uint8
	koePowerCnt   uint8

	koePalKOEB [32 * 3]byte

	koeRowsKOEA [200]uint16
	koeBlit16t  [256]uint16

	koeCircles [8][]byte

	koeFrameCountB uint16
	koePalAnimCB   int
	koePalAnimC2B  int
	koeScrnPosB    uint16
	koeScrnPosBl   uint16
	koeScrnX       uint16
	koeScrnY       uint16
	koeScrnRot     uint16
	koeSinuRot     uint16
	koeOverRot     uint16
	koeOverX       uint16
	koeOverYA2     uint16
	koePatDir      int = -3

	koeVBuf [8192]byte

	koePlanarVRAM [4][8][koeVbufSize]byte

	koeCircleMem [16384 * 16]byte
	koePic       [20000 * 4]byte

	koePalette [constants.PaletteByteCount]byte
	koePalFade [13000]byte

	koePl  int = 1
	koePlv int

	koeCurrentPal [16 * 16 * 3]byte

	koeCurPal int
	koeLastA  int

	koePower0 [16 * 256]byte
	koePower1 [16 * 256]byte

	koeTempBuf [koePitch]byte
	koePlanar  [koePitch * 4]byte

	koeFlashPal1 [16 * 3]int
)

func runKOE() {
	if err := koeEnsureData(); err != nil {
		log.Printf("koe: %v", err)
		return
	}
	koeMain()
}

func koeReset() {
	koeClip1 = koeXY{}
	koeClip2 = koeXY{}

	koePolyISides = 0
	koePolySides = 0

	clear(koeClipXY2[:])
	clear(koePolyIXY[:])
	clear(koePolyXY[:])

	koeClipLeft = 0

	koeFrameCountA = 0
	koePalAnimCA = 0
	koePalAnimCA2 = 0
	koeScrnPosA = 0
	koeScrnPosAl = 0
	koeScrnXA = 0
	koeScrnYA = 0
	koeScrnRotA = 0
	koeSinuRotA = 0
	koeOverRotA = 211
	koeOverXA = 0
	koeOverYA = 0
	koePatDirA = 0

	koeSizeFade = 0
	koeRotSpeed = 0
	koePalFader = 0
	koePalFader2 = 255
	koeZumPlane = 0x11

	koeSinusPower = 0
	koePowerCnt = 0

	clear(koePalKOEB[:])

	clear(koeRowsKOEA[:])
	clear(koeBlit16t[:])

	for i := range koeCircles {
		koeCircles[i] = nil
	}

	koeFrameCountB = 0
	koePalAnimCB = 0
	koePalAnimC2B = 0
	koeScrnPosB = 0
	koeScrnPosBl = 0
	koeScrnX = 0
	koeScrnY = 0
	koeScrnRot = 0
	koeSinuRot = 0
	koeOverRot = 0
	koeOverX = 0
	koeOverYA2 = 0
	koePatDir = -3

	clear(koeVBuf[:])
	clear(koePlanarVRAM[:])

	clear(koeCircleMem[:])
	clear(koePic[:])
	clear(koePalette[:])
	clear(koePalFade[:])

	koePl = 1
	koePlv = 0

	clear(koeCurrentPal[:])

	koeCurPal = 0
	koeLastA = 0

	clear(koePower0[:])
	clear(koePower1[:])

	clear(koeTempBuf[:])
	clear(koePlanar[:])

	clear(koeFlashPal1[:])
}

func koeMain() {
	koeReset()
	shim.ClearScreen()

	for c := 0; c < 16; c++ {
		base := 3 * 16 * c
		for a := 0; a < 16; a++ {
			x := 0
			if (a & 1) != 0 {
				x++
			}
			if (a & 2) != 0 {
				x++
			}
			if (a & 4) != 0 {
				x++
			}
			if (a & 8) != 0 {
				x++
			}

			var r, g, b int
			switch x {
			case 0:
				r, g, b = 0, 0, 0
			case 1:
				r = 38 * 64 / 111
				g = 33 * 64 / 111
				b = 44 * 64 / 111
			case 2:
				r = 52 * 64 / 111
				g = 45 * 64 / 111
				b = 58 * 64 / 111
			case 3:
				r = 67 * 64 / 111
				g = 61 * 64 / 111
				b = 73 * 64 / 111
			case 4:
				r = 83 * 64 / 111
				g = 77 * 64 / 111
				b = 89 * 64 / 111
			}

			scaleR := (10 + c*9/9)
			scaleG := (10 + c*7/9)
			scaleB := (10 + c*5/9)

			r = r * scaleR / 10
			g = g * scaleG / 10
			b = b * scaleB / 10

			if r > 63 {
				r = 63
			}
			if g > 63 {
				g = 63
			}
			if b > 63 {
				b = 63
			}

			koeCurrentPal[base] = byte(r)
			koeCurrentPal[base+1] = byte(g)
			koeCurrentPal[base+2] = byte(b)
			base += 3
		}
	}

	koeBlitInit()

	p := 0
	for b := 0; b < 16; b++ {
		for c := 0; c < 256; c++ {
			var ch int
			if b == 15 {
				ch = c
			} else {
				a := c
				if c > 127 {
					a = c - 256
				}
				ch = a * b / 15
			}
			koePower0[p] = byte(int8(ch))
			p++
		}
	}

	p = 0
	for b := 0; b < 16; b++ {
		for c := 0; c < 256; c++ {
			var ch int
			if b == 15 {
				ch = c
			} else {
				a := c
				if c > 127 {
					a = c - 256
				}
				ch = a * b / 15
			}
			koePower1[p] = byte(int8(ch))
			p++
		}
	}

	if !shim.IsDemoFirstPart() {
		for !driver.WantsToQuit() && music.GetPlusFlags() < -4 {
			driver.Vsync(false)
		}
	}

	music.SetFrame(0)

	koeInitInterference()
	koeDoInterference()

	koeInitInterferenceKOEA(koeCircleMem[:])

	if !driver.WantsToQuit() {
		koeDoInterferenceKOEA()

		koeFlash(-1)
		koeFlash(32)
		koeFlash(64)
		koeFlash(192)
		koeFlash(256)

		if constants.ScreenSize <= len(shim.VRAM) {
			clear(shim.VRAM[:constants.ScreenSize])
		}

		fill := 8000 * 8
		if fill > len(shim.VRAM) {
			fill = len(shim.VRAM)
		}
		for i := 0; i < fill; i++ {
			shim.VRAM[i] = 15
		}

		ip := koeFlash(-2)
		for a := 0; a < 16; a++ {
			ip[a*3+0] = a * 6 / 2
			ip[a*3+1] = a * 7 / 2
			ip[a*3+2] = a * 8 / 2
		}

		for !driver.WantsToQuit() && music.GetPlusFlags() < -3 {
			driver.Vsync(false)
		}

		for !driver.WantsToQuit() {
			if (music.GetRow() & 7) == 7 {
				break
			}
			driver.Vsync(false)
		}

		for b := 0; b < 4 && !driver.WantsToQuit(); b++ {
			zy := 0
			zya := 0
			zly := 0
			for a := 256; a > -400; a -= 32 {
				zly = zy
				zya++
				zy += zya
				v := zly * constants.ScreenWidth
				for y := zly; y < zy; y++ {
					idx := v + b*constants.PlanarWidth
					if idx+constants.PlanarWidth <= len(shim.VRAM) {
						for i := 0; i < constants.PlanarWidth; i++ {
							shim.VRAM[idx+i] = 0
						}
					}
					v += constants.ScreenWidth
				}
				if a >= 0 {
					koeFlash(a)
				} else {
					koeFlash(0)
				}
			}

			for !driver.WantsToQuit() {
				if (music.GetRow() & 7) == 7 {
					break
				}
				driver.Vsync(false)
			}

			koeFlash(-1)
			koeFlash(32)
			koeFlash(64)
			koeFlash(192)
			koeFlash(256)
		}
	}

	shim.ClearScreen()

	koeDoit1(70 * 6)
	koeDoit2(70 * 12)
	koeDoit3(70 * 14)
}

func koeBlitInit() {
	var off uint16
	for y := 0; y < 200; y++ {
		koeRowsKOEA[y] = off
		off += 40
	}

	for al := 0; al < 256; al++ {
		dh := uint8(255)
		dl := uint8(al)
		ah := uint8(0)

		for i := 0; i < 8; i++ {
			msb := dl >> 7
			dl = (dl << 1) | msb
			if msb != 0 {
				ah ^= dh
			}
			dh >>= 1
		}

		high := uint8(0)
		if (ah & 1) != 0 {
			high = 0x80
		}

		koeBlit16t[al] = uint16(ah) | (uint16(high) << 8)
	}
}

func koeAsmDoit(src []byte, dst []byte) {
	if len(src) < koeVbufSize || len(dst) < koeVbufSize {
		return
	}
	srcPos := 0
	dstPos := 0
	for line := 0; line < 200; line++ {
		dh := uint8(0)
		for i := 0; i < 20; i++ {
			bl := src[srcPos] ^ dh
			t0 := koeBlit16t[bl]
			al := byte(t0)
			ah := byte(t0 >> 8)

			bl = src[srcPos+1] ^ ah
			t1 := koeBlit16t[bl]
			dl := byte(t1)
			dh = byte(t1 >> 8)

			dst[dstPos] = al
			dst[dstPos+1] = dl

			srcPos += 2
			dstPos += 2
		}
	}
}

func koeBltLine(src []byte, dst []byte, planeCount int) {
	off := 0
	for plane := 0; plane < planeCount; plane++ {
		if off+40 > len(src) || plane*koePitch+40 > len(dst) {
			return
		}
		copy(dst[plane*koePitch:plane*koePitch+40], src[off:off+40])
		off += 40
	}
}

func koeBltLineRev(src []byte, dst []byte, planeCount int) {
	off := 0
	for plane := 0; plane < planeCount; plane++ {
		if off+40 > len(src) || plane*koePitch+40 > len(dst) {
			return
		}
		base := plane * koePitch
		for i := 0; i < 40; i++ {
			dst[base+i] = koeFlip8[src[off+39-i]]
		}
		off += 40
	}
}

func koeMixPal(src []byte, dst []byte, count int, fade int) {
	if fade <= 256 {
		for i := 0; i < count && i < len(src) && i < len(dst); i++ {
			dst[i] = byte((int(src[i]) * fade) >> 8)
		}
		return
	}
	add := fade - 256
	for i := 0; i < count && i < len(src) && i < len(dst); i++ {
		val := int(src[i]) + add
		if val > 63 {
			val = 63
		}
		dst[i] = byte(val)
	}
}

func koeInitInterference() {
	src := koeCircle2
	top := 0
	bottom := (constants.DoubleScreenHeight - 1) * constants.PlanarWidth
	for y := 0; y < 200; y++ {
		if len(src) >= 40 {
			koeBltLineMono(src, koeTempBuf[top:])
			koeBltLineRevMono(src, koeTempBuf[top+40:])
			koeBltLineMono(src, koeTempBuf[bottom:])
			koeBltLineRevMono(src, koeTempBuf[bottom+40:])
		}
		top += constants.PlanarWidth
		bottom -= constants.PlanarWidth
		if len(src) >= 40 {
			src = src[40:]
		}
	}

	src = koeCircle
	top = 0
	bottom = (constants.DoubleScreenHeight - 1) * constants.PlanarWidth
	for y := 0; y < 200; y++ {
		if len(src) >= 40*3 {
			koeBltLinePlanes(src, koePlanar[top:], 3)
			koeBltLineRevPlanes(src, koePlanar[top+40:], 3)
			koeBltLinePlanes(src, koePlanar[bottom:], 3)
			koeBltLineRevPlanes(src, koePlanar[bottom+40:], 3)
		}
		top += constants.PlanarWidth
		bottom -= constants.PlanarWidth
		if len(src) >= 40*3 {
			src = src[40*3:]
		}
	}

	koeFrameCountB = 0
}

func koeResolvePlanar() {
	srcPos := int(koeScrnPosB)
	dstPos := 0
	pixelsLeft := constants.ScreenWidth * 200
	rowPixels := constants.ScreenWidth
	bitMask := uint8(0x80)

	for pixelsLeft > 0 && dstPos < len(shim.VRAM) {
		var color uint8
		if koePlanar[srcPos+0*koePitch]&bitMask != 0 {
			color |= 1
		}
		if koePlanar[srcPos+1*koePitch]&bitMask != 0 {
			color |= 2
		}
		if koePlanar[srcPos+2*koePitch]&bitMask != 0 {
			color |= 4
		}
		if koePlanar[srcPos+3*koePitch]&bitMask != 0 {
			color |= 8
		}
		shim.VRAM[dstPos] = color
		dstPos++

		bitMask = (bitMask >> 1) | ((bitMask & 1) << 7)
		if bitMask == 0x80 {
			srcPos++
		}

		rowPixels--
		if rowPixels == 0 {
			srcPos += 40
			rowPixels = constants.ScreenWidth
		}

		pixelsLeft--
	}
}

func koeDoInterference() {
	for {
		driver.Vsync(false)

		common.SetPalArea(koePalKOEB[:], 0, 16)

		animPos := koePalAnimCB + koePatDir
		if animPos < 0 {
			animPos = 8*3 - 3
		} else if animPos >= 8*3 {
			animPos = 0
		}
		koePalAnimCB = animPos
		koePalAnimC2B = animPos

		koePalFader += 2
		if koePalFader > 512 {
			koePalFader = 512
		}

		if animPos < len(koePal0) {
			koeMixPal(koePal0[animPos:], koePalKOEB[:], 8*3, int(koePalFader))
			koeMixPal(koePal0[animPos:], koePalKOEB[8*3:], 8*3, int(koePalFader))
		}

		common.SetPalArea(koePalKOEB[:], 0, 16)

		row := music.GetRow() & 7
		if row == 0 || row == 4 {
			koePatDir = -3
		}

		koeScrnRot = (koeScrnRot + 5) & 1023

		if koeFrameCountB >= 64 {
			koeRotSpeed++
		}

		if len(koeSin1024) > 0 {
			sinIdx := (int(koeScrnRot) * 2) & 1023
			scrn := (int32(koeSin1024[sinIdx]) * int32(koeSizeFade)) >> 16
			koeScrnX = uint16(int32(scrn) + 160)

			sinIdx = (sinIdx + 512) & 1023
			scrn = (int32(koeSin1024[sinIdx]) * int32(koeSizeFade)) >> 16
			koeScrnY = uint16(int32(scrn) + 100)

			koeOverRot = (koeOverRot + koeRotSpeed) & 1023
			sinIdx = (int(koeOverRot) * 2) & 1023
			over := (int32(koeSin1024[sinIdx]) >> 2)
			over = (over * int32(koeSizeFade)) >> 16
			koeOverX = uint16(int32(over) + 160)

			sinIdx = (sinIdx + 512) & 1023
			over = (int32(koeSin1024[sinIdx]) >> 2)
			over = (over * int32(koeSizeFade)) >> 16
			koeOverYA2 = uint16((int32(over) + 100) * 80)
		}

		koeScrnPosBl = koeScrnX & 7
		koeScrnPosB = uint16(int(koeScrnY)*constants.PlanarWidth + int(koeScrnX>>3))

		koeFrameCountB++
		if koeFrameCountB >= 256 {
			break
		}

		koeResolvePlanar()
		driver.Blit()

		if driver.WantsToQuit() {
			break
		}
	}
}

func koeWaitBorder() int {
	if a := music.GetRow(); a != koeLastA {
		koeLastA = a
		if (a & 7) == 7 {
			koeCurPal = 15
		}
	}

	r := driver.Vsync(false)
	if r > 10 {
		r = 10
	}
	if r < 1 {
		r = 1
	}

	start := 16 * 3 * koeCurPal
	if start+16*3 <= len(koeCurrentPal) {
		shim.Outp(0x3c8, 0)
		for a := 0; a < 16*3; a++ {
			shim.Outp(0x3c9, uint32(koeCurrentPal[start+a]))
		}
	}

	if koeCurPal != 0 {
		koeCurPal--
	}

	return r
}

func koeFlash(i int) []int {
	if i == -2 {
		return koeFlashPal1[:]
	}
	if i == -1 {
		shim.Outp(0x3c7, 0)
		for a := 0; a < 16*3; a++ {
			koeFlashPal1[a] = int(shim.Inp(0x3c9))
		}
		return koeFlashPal1[:]
	}

	j := 256 - i
	var pal2 [16 * 3]int
	for a := 0; a < 16*3; a++ {
		pal2[a] = (koeFlashPal1[a]*j + 63*i) >> 8
	}

	driver.Vsync(false)
	shim.Outp(0x3c8, 0)
	for a := 0; a < 16*3; a++ {
		shim.Outp(0x3c9, uint32(pal2[a]))
	}
	driver.Blit()

	return koeFlashPal1[:]
}

func koeResolve16Color(page int, xShift int) {
	if page < 0 || page >= 8 {
		return
	}
	if xShift < 0 {
		xShift = 0
	}
	if xShift > constants.ScreenWidth {
		xShift = constants.ScreenWidth
	}

	dstPos := 0
	dstStride := constants.ScreenWidth - xShift
	srcStride := (xShift + 7) >> 3
	idx := 0

	for y := 0; y < 200; y++ {
		for i := 0; i < xShift && dstPos < len(shim.VRAM); i++ {
			shim.VRAM[dstPos] = 0
			dstPos++
		}
		for x := 0; x < dstStride && dstPos < len(shim.VRAM); x++ {
			bit := x & 7
			mask := uint8(1 << (7 - bit))
			var color uint8
			if koePlanarVRAM[0][page][idx]&mask != 0 {
				color |= 1
			}
			if koePlanarVRAM[1][page][idx]&mask != 0 {
				color |= 2
			}
			if koePlanarVRAM[2][page][idx]&mask != 0 {
				color |= 4
			}
			if koePlanarVRAM[3][page][idx]&mask != 0 {
				color |= 8
			}
			shim.VRAM[dstPos] = color
			dstPos++
			if bit == 7 {
				idx++
			}
		}
		idx += srcStride
	}
}

func koeSideScroll256(src []byte, xShift int) {
	if xShift >= 0 {
		if xShift > constants.ScreenWidth {
			xShift = constants.ScreenWidth
		}
		dstStride := constants.ScreenWidth - xShift
		dstPos := 0
		srcPos := 0
		for y := 0; y < constants.DoubleScreenHeight; y++ {
			for i := 0; i < xShift && dstPos < len(shim.VRAM); i++ {
				shim.VRAM[dstPos] = 0
				dstPos++
			}
			if srcPos+dstStride <= len(src) && dstPos+dstStride <= len(shim.VRAM) {
				copy(shim.VRAM[dstPos:dstPos+dstStride], src[srcPos:srcPos+dstStride])
			}
			dstPos += dstStride
			srcPos += constants.ScreenWidth
		}
		return
	}

	xShift = -xShift
	if xShift > constants.ScreenWidth {
		xShift = constants.ScreenWidth
	}
	stride := constants.ScreenWidth - xShift
	dstPos := 0
	srcPos := 0
	for y := 0; y < constants.DoubleScreenHeight; y++ {
		srcPos += xShift
		if srcPos+stride <= len(src) && dstPos+stride <= len(shim.VRAM) {
			copy(shim.VRAM[dstPos:dstPos+stride], src[srcPos:srcPos+stride])
		}
		srcPos += stride
		for i := 0; i < xShift && dstPos+stride+i < len(shim.VRAM); i++ {
			shim.VRAM[dstPos+stride+i] = 0
		}
		dstPos += constants.ScreenWidth
	}
}

func koeOutPal(pal []byte, start, count int) {
	common.SetPalArea(pal, start, count)
}

func koeBltLineMono(src []byte, dst []byte) {
	if len(src) < 40 || len(dst) < 40 {
		return
	}
	copy(dst[:40], src[:40])
}

func koeBltLineRevMono(src []byte, dst []byte) {
	if len(src) < 40 || len(dst) < 40 {
		return
	}
	for i := 0; i < 40; i++ {
		dst[i] = koeFlip8[src[39-i]]
	}
}

func koeBltLinePlanes(src []byte, dst []byte, planes int) {
	for p := 0; p < planes; p++ {
		off := p * 40
		if off+40 > len(src) || p*koePitch+40 > len(dst) {
			return
		}
		copy(dst[p*koePitch:p*koePitch+40], src[off:off+40])
	}
}

func koeBltLineRevPlanes(src []byte, dst []byte, planes int) {
	for p := 0; p < planes; p++ {
		off := p * 40
		if off+40 > len(src) || p*koePitch+40 > len(dst) {
			return
		}
		base := p * koePitch
		for i := 0; i < 40; i++ {
			dst[base+i] = koeFlip8[src[off+39-i]]
		}
	}
}

func koeRotate1PageRCR(src []byte, dst []byte) {
	const n = 32000
	if len(src) < n || len(dst) < n {
		return
	}
	var cf uint8
	for i := 0; i < n; i++ {
		b := src[i]
		newCF := b & 1
		dst[i] = (b >> 1) | (cf << 7)
		cf = newCF
	}
}

func koeResolvePlanarKOEA() {
	esi := int(koeScrnPosA)
	edi := 0
	bl := uint8(0x80 >> (koeScrnPosAl & 7))
	ecx := constants.ScreenWidth * 200
	edx := constants.ScreenWidth

	for ecx > 0 && edi < len(shim.VRAM) {
		var al uint8
		if koePlanar[esi+0*koePitch]&bl != 0 {
			al |= 1
		}
		if koePlanar[esi+1*koePitch]&bl != 0 {
			al |= 2
		}
		if koePlanar[esi+2*koePitch]&bl != 0 {
			al |= 4
		}
		if koePlanar[esi+3*koePitch]&bl != 0 {
			al |= 8
		}
		shim.VRAM[edi] = al
		edi++

		bl = (bl >> 1) | (bl << 7)
		if (bl & 0x80) != 0 {
			esi++
		}

		edx--
		if edx == 0 {
			esi += 40
			edx = constants.ScreenWidth
		}
		ecx--
	}
}

func koeInitInterferenceKOEA(memory []byte) {
	pageSize := 2048 * 16
	for i := 0; i < 8; i++ {
		if len(memory) < pageSize {
			koeCircles[i] = nil
			continue
		}
		koeCircles[i] = memory[:pageSize]
		memory = memory[pageSize:]
	}

	esi := koeCircle2
	ediTop := 0
	ediBot := (constants.DoubleScreenHeight - 1) * constants.PlanarWidth
	for y := 0; y < 200; y++ {
		if len(esi) >= 40 {
			koeBltLineMono(esi, koeTempBuf[ediTop:])
			koeBltLineRevMono(esi, koeTempBuf[ediTop+40:])

			koeBltLineMono(esi, koeTempBuf[ediBot:])
			koeBltLineRevMono(esi, koeTempBuf[ediBot+40:])

			esi = esi[40:]
		}
		ediTop += constants.PlanarWidth
		ediBot -= constants.PlanarWidth
	}

	if len(koeCircles[0]) >= 32000 {
		copy(koeCircles[0], koeTempBuf[:32000])
	}
	for i := 0; i < 7; i++ {
		if len(koeCircles[i]) >= 32000 && len(koeCircles[i+1]) >= 32000 {
			koeRotate1PageRCR(koeCircles[i], koeCircles[i+1])
		}
	}

	sc := koeCircle
	dt := 0
	db := (constants.DoubleScreenHeight - 1) * constants.PlanarWidth
	for y := 0; y < constants.ScreenHeight; y++ {
		if len(sc) >= 40*3 {
			koeBltLinePlanes(sc, koePlanar[dt:], 3)
			koeBltLineRevPlanes(sc, koePlanar[dt+40:], 3)

			koeBltLinePlanes(sc, koePlanar[db:], 3)
			koeBltLineRevPlanes(sc, koePlanar[db+40:], 3)

			sc = sc[40*3:]
		}
		dt += constants.PlanarWidth
		db -= constants.PlanarWidth
	}

	koeFrameCountA = 0
}

func koeSarDiv8(x int32) int32 {
	if x >= 0 {
		return x >> 3
	}
	return -(((-x) + 7) >> 3)
}

func koeSinByByteOffset(byteOff uint16) int16 {
	if len(koeSin1024) == 0 {
		return 0
	}
	idx := int(byteOff&2046) / 2
	if idx < 0 {
		idx = 0
	}
	if idx >= len(koeSin1024) {
		idx %= len(koeSin1024)
	}
	return koeSin1024[idx]
}

func koeDoInterferenceKOEA() {
	for {
		driver.Vsync(false)

		si := int(koePalAnimCA) + int(koePatDirA)
		if si < 0 {
			si = 8*3 - 3
		}
		if si >= 8*3 {
			si = 0
		}
		koePalAnimCA = uint16(si)
		koePalAnimCA2 = uint16(si)

		if si < len(koePal1KOE) {
			koeOutPal(koePal1KOE[si:], 0, 8*3)
		}
		if si < len(koePal2KOE) {
			koeOutPal(koePal2KOE[si:], 8, 8*3)
		}

		esiBase := 0
		edi := 3*koePitch + int(koeScrnPosA)

		ebp := (koeSinuRotA + 7*2) & 2047
		koeSinuRotA = ebp

		for lines := 200; lines > 0; lines-- {
			ebp = (ebp + 9*2) & 2047

			s := koeSinByByteOffset(ebp)
			ebx := int32(s)
			ebx >>= 3

			idx := int((uint16(koeSinusPower) << 8) | uint16(uint8(ebx)))
			eax := int32(int8(koePower0[idx]))
			eax += int32(koeOverXA)
			eax -= int32(koeScrnPosAl)

			page := 7 - int(uint32(eax)&7)
			if page < 0 || page >= len(koeCircles) || koeCircles[page] == nil {
				edi += constants.PlanarWidth
				esiBase += constants.PlanarWidth
				continue
			}

			byteOff := koeSarDiv8(eax)
			srcPos := int(uint16(koeOverYA)) + int(byteOff) + esiBase
			if srcPos < 0 {
				srcPos = 0
			}
			if srcPos+44 <= len(koeCircles[page]) && edi+44 <= len(koePlanar) {
				copy(koePlanar[edi:edi+44], koeCircles[page][srcPos:srcPos+44])
			}

			edi += constants.PlanarWidth
			esiBase += constants.PlanarWidth
		}

		koeResolvePlanarKOEA()
		driver.Blit()

		row := music.GetRow() & 7
		if row == 0 || row == 4 {
			koePatDirA = -3
		}

		bx := (koeScrnRotA + 5) & 1023
		koeScrnRotA = bx
		if len(koeSin1024) > 0 {
			v := int16(koeSin1024[bx])
			v >>= 2
			v += 160
			koeScrnXA = uint16(v)

			bx2 := (bx + 256) & 1023
			v = int16(koeSin1024[bx2])
			v >>= 2
			v += 100
			koeScrnYA = uint16(v)

			bx = (koeOverRotA + 7) & 1023
			koeOverRotA = bx
			v = int16(koeSin1024[bx])
			v >>= 2
			v += 160
			koeOverXA = uint16(v)

			bx2 = (bx + 256) & 1023
			v = int16(koeSin1024[bx2])
			v >>= 2
			v += 100
			koeOverYA = uint16(uint16(v) * 80)
		}

		ax := koeScrnXA
		koeScrnPosAl = ax & 7
		rowBytes := uint16(constants.PlanarWidth) * koeScrnYA
		koeScrnPosA = rowBytes + (ax >> 3)

		if koeFrameCountA >= 70*5 {
			koePowerCnt++
			if koePowerCnt >= 16 {
				koePowerCnt = 0
				if koeSinusPower < 15 {
					koeSinusPower++
				}
			}
		}

		koeFrameCountA++
		if music.GetFrame() >= 925 {
			break
		}
		if driver.WantsToQuit() {
			break
		}
	}
}

func koeClipIntersect(v1 *int16, v2 int16, w1 *int16, w2 int16, wl int16) {
	dw := int32(w2) - int32(*w1)
	if dw != 0 {
		num := (int32(wl) - int32(*w1)) * (int32(v2) - int32(*v1))
		dv := num / dw
		*v1 = int16(int32(*v1) + dv)
	}
	*w1 = wl
}

func koeOutcodeX(x int16) uint8 {
	var c uint8
	if x < koeWMinX {
		c |= 1
	}
	if x > koeWMaxX {
		c |= 2
	}
	return c
}

func koeOutcodeY(y int16) uint8 {
	var c uint8
	if y < koeWMinY {
		c |= 4
	}
	if y > koeWMaxY {
		c |= 8
	}
	return c
}

func koeDrawLine(x1, y1, x2, y2 uint16) {
	if y1 > y2 {
		x1, x2 = x2, x1
		y1, y2 = y2, y1
	}

	dy := int(y2 - y1)
	si := int(koeRowsKOEA[y1])

	if koeClipLeft != 0 {
		cnt := koeClipLeft
		tesi := si
		if cnt < 0 {
			for ; cnt != 0; cnt++ {
				tesi += 40
				if tesi >= 0 && tesi < len(koeVBuf) {
					koeVBuf[tesi] ^= 0x80
				}
			}
		} else {
			for ; cnt != 0; cnt-- {
				tesi -= 40
				if tesi >= 0 && tesi < len(koeVBuf) {
					koeVBuf[tesi] ^= 0x80
				}
			}
		}
	}

	if dy == 0 {
		return
	}

	si += int(x1 >> 3)
	mask := uint8(0x80 >> (x1 & 7))
	bx := -(dy / 2)

	if x1 >= x2 {
		dx := int(x1) - int(x2)
		cx := dy
		for cx > 0 {
			if si >= 0 && si < len(koeVBuf) {
				koeVBuf[si] ^= mask
			}
			si += 40
			bx += dx
			for bx >= 0 {
				bx -= dy
				mask <<= 1
				if mask == 0 {
					mask = 0x01
					si--
				}
			}
			cx--
		}
		return
	}

	dx := int(x2) - int(x1)
	cx := dy
	for cx > 0 {
		if si >= 0 && si < len(koeVBuf) {
			koeVBuf[si] ^= mask
		}
		si += 40
		bx += dx
		for bx >= 0 {
			bx -= dy
			mask >>= 1
			if mask == 0 {
				mask = 0x80
				si++
			}
		}
		cx--
	}
}

func koeClipLineX(a *koeXY, b *koeXY) bool {
	ca := koeOutcodeX(a.x)
	cb := koeOutcodeX(b.x)
	if (ca & cb) != 0 {
		return true
	}
	if (ca & 1) != 0 {
		koeClipIntersect(&a.y, b.y, &a.x, b.x, koeWMinX)
	}
	if (ca & 2) != 0 {
		koeClipIntersect(&a.y, b.y, &a.x, b.x, koeWMaxX)
	}
	if (cb & 1) != 0 {
		koeClipIntersect(&b.y, a.y, &b.x, a.x, koeWMinX)
	}
	if (cb & 2) != 0 {
		koeClipIntersect(&b.y, a.y, &b.x, a.x, koeWMaxX)
	}
	return false
}

func koeClipLineY(a *koeXY, b *koeXY) bool {
	ca := koeOutcodeY(a.y)
	cb := koeOutcodeY(b.y)
	if (ca & cb) != 0 {
		return true
	}
	if (ca & 4) != 0 {
		koeClipIntersect(&a.x, b.x, &a.y, b.y, koeWMinY)
	}
	if (ca & 8) != 0 {
		koeClipIntersect(&a.x, b.x, &a.y, b.y, koeWMaxY)
	}
	if (cb & 4) != 0 {
		koeClipIntersect(&b.x, a.x, &b.y, a.y, koeWMinY)
	}
	if (cb & 8) != 0 {
		koeClipIntersect(&b.x, a.x, &b.y, a.y, koeWMaxY)
	}
	return false
}

func koeClipAnyPoly() {
	n := koePolyISides
	if n == 0 {
		koePolySides = 0
		return
	}
	if n == 1 {
		p := koePolyIXY[0]
		if p.x < koeWMinX || p.x > koeWMaxX || p.y < koeWMinY || p.y > koeWMaxY {
			koePolySides = 0
			return
		}
		koePolyXY[0] = p
		koePolySides = 1
		return
	}
	if n == 2 {
		a := koePolyIXY[0]
		b := koePolyIXY[1]
		if koeClipLineY(&a, &b) {
			koePolySides = 0
			return
		}
		if koeClipLineX(&a, &b) {
			koePolySides = 0
			return
		}
		koePolyXY[0] = a
		koePolyXY[1] = b
		if a.x == b.x && a.y == b.y {
			koePolySides = 1
		} else {
			koePolySides = 2
		}
		return
	}

	outc := uint16(0)
	last := koePolyIXY[n-1]
	lastDW := (uint32(uint16(last.y)) << 16) | uint32(uint16(last.x))

	for i := uint16(0); i < n; i++ {
		cur := koePolyIXY[i]
		a := last
		b := cur
		last = cur

		if !koeClipLineY(&a, &b) {
			da := (uint32(uint16(a.y)) << 16) | uint32(uint16(a.x))
			if da != lastDW && int(outc) < len(koeClipXY2) {
				koeClipXY2[outc] = a
				outc++
				lastDW = da
			}
			db := (uint32(uint16(b.y)) << 16) | uint32(uint16(b.x))
			if db != lastDW && int(outc) < len(koeClipXY2) {
				koeClipXY2[outc] = b
				outc++
				lastDW = db
			}
		}
	}

	m := outc
	if m >= 2 {
		if koeClipXY2[0].x == koeClipXY2[m-1].x && koeClipXY2[0].y == koeClipXY2[m-1].y {
			m--
		}
	}
	if m <= 2 {
		if m == 0 {
			koePolySides = 0
			return
		}
		a := koeClipXY2[0]
		b := a
		if m == 2 {
			b = koeClipXY2[1]
		}
		if koeClipLineX(&a, &b) {
			koePolySides = 0
			return
		}
		koePolyXY[0] = a
		koePolyXY[1] = b
		if m == 2 {
			if a.x == b.x && a.y == b.y {
				koePolySides = 1
			} else {
				koePolySides = 2
			}
		} else {
			koePolySides = 1
		}
		return
	}

	outc = 0
	last = koeClipXY2[m-1]
	lastDW = (uint32(uint16(last.y)) << 16) | uint32(uint16(last.x))

	for i := uint16(0); i < m; i++ {
		cur := koeClipXY2[i]
		a := last
		b := cur
		last = cur

		if !koeClipLineX(&a, &b) {
			da := (uint32(uint16(a.y)) << 16) | uint32(uint16(a.x))
			if da != lastDW && int(outc) < len(koePolyXY) {
				koePolyXY[outc] = a
				outc++
				lastDW = da
			}
			db := (uint32(uint16(b.y)) << 16) | uint32(uint16(b.x))
			if db != lastDW && int(outc) < len(koePolyXY) {
				koePolyXY[outc] = b
				outc++
				lastDW = db
			}
		}
	}

	koePolySides = outc
	if koePolySides > 0 && koePolyXY[0].x == koePolyXY[koePolySides-1].x && koePolyXY[0].y == koePolyXY[koePolySides-1].y {
		koePolySides--
	}
}

func koeAsmBox(x0, y0, x1, y1, x2, y2, x3, y3 int) {
	koePolyIXY[0] = koeXY{int16(x0), int16(y0)}
	koePolyIXY[1] = koeXY{int16(x1), int16(y1)}
	koePolyIXY[2] = koeXY{int16(x2), int16(y2)}
	koePolyIXY[3] = koeXY{int16(x3), int16(y3)}
	koePolyISides = 4

	koeClipAnyPoly()

	n := koePolySides
	if n == 0 {
		return
	}
	for i := uint16(0); i+1 < n; i++ {
		koeDrawLine(uint16(koePolyXY[i].x), uint16(koePolyXY[i].y), uint16(koePolyXY[i+1].x), uint16(koePolyXY[i+1].y))
	}
	koeDrawLine(uint16(koePolyXY[n-1].x), uint16(koePolyXY[n-1].y), uint16(koePolyXY[0].x), uint16(koePolyXY[0].y))
}

func koeDoit1(count int) int {
	rot := 45
	vm := 50
	vma := 0
	koeWaitBorder()
	koePlv = 0
	koePl = 1
	for !driver.WantsToQuit() && count > 0 {
		count -= koeWaitBorder()
		clear(koeVBuf[:koeVbufSize])

		hx := int(koeSin1024[(rot+0)&1023]) * 16 * 6 / 5
		hy := int(koeSin1024[(rot+256)&1023]) * 16
		vx := int(koeSin1024[(rot+256)&1023]) * 6 / 5
		vy := int(koeSin1024[(rot+512)&1023])
		vx = vx * vm / 100
		vy = vy * vm / 100
		for c := -10; c < 11; c += 2 {
			cx := vx * c * 2
			cy := vy * c * 2
			x1 := (-hx-vx+cx)/16 + 160
			y1 := (-hy-vy+cy)/16 + 100
			x2 := (-hx+vx+cx)/16 + 160
			y2 := (-hy+vy+cy)/16 + 100
			x3 := (+hx+vx+cx)/16 + 160
			y3 := (+hy+vy+cy)/16 + 100
			x4 := (+hx-vx+cx)/16 + 160
			y4 := (+hy-vy+cy)/16 + 100
			koeAsmBox(x1, y1, x2, y2, x3, y3, x4, y4)
		}
		rot += 2
		vm += vma
		if vm < 25 {
			vm -= vma
			vma = -vma
		}
		vma--

		koeAsmDoit(koeVBuf[:], koePlanarVRAM[koePl][koePlv][:])
		koeResolve16Color(koePlv, 0)

		koePlv++
		koePlv &= 7
		if koePlv == 0 {
			koePl = (koePl + 1) & 3
		}
		driver.Blit()
	}
	return 0
}

func koeDoit2(count int) int {
	rot := 50
	rota := 10
	vm := 100 * 64
	vma := 0
	koeWaitBorder()
	koePlv = 0
	koePl = 1
	for !driver.WantsToQuit() && count > 0 {
		count -= koeWaitBorder()
		clear(koeVBuf[:koeVbufSize])

		hx := int(koeSin1024[(rot+0)&1023]) * 16 * 6 / 5
		hy := int(koeSin1024[(rot+256)&1023]) * 16
		vx := int(koeSin1024[(rot+256)&1023]) * 6 / 5
		vy := int(koeSin1024[(rot+512)&1023])
		vx = vx * (vm / 64) / 100
		vy = vy * (vm / 64) / 100
		for c := -10; c < 11; c += 2 {
			cx := vx * c * 2
			cy := vy * c * 2
			x1 := (-hx-vx+cx)/16 + 160
			y1 := (-hy-vy+cy)/16 + 100
			x2 := (-hx+vx+cx)/16 + 160
			y2 := (-hy+vy+cy)/16 + 100
			x3 := (+hx+vx+cx)/16 + 160
			y3 := (+hy+vy+cy)/16 + 100
			x4 := (+hx-vx+cx)/16 + 160
			y4 := (+hy-vy+cy)/16 + 100
			koeAsmBox(x1, y1, x2, y2, x3, y3, x4, y4)
		}
		rot += rota / 10
		vm += vma
		if vm < 0 {
			vm -= vma
			vma = -vma
		}
		vma--
		rota++

		koeAsmDoit(koeVBuf[:], koePlanarVRAM[koePl][koePlv][:])
		koeResolve16Color(koePlv, 0)

		koePlv++
		koePlv &= 7
		if koePlv == 0 {
			koePl = (koePl + 1) & 3
		}
		driver.Blit()
	}
	return 0
}

func koeDoit3(count int) int {
	rot := 45
	rota := 10
	rot2 := 0
	vm := 100 * 64
	vma := 0
	xpos := constants.ScreenWidth
	xposa := 0
	repeat := 1

	koeWaitBorder()
	shim.Outp(0x3c8, 0)
	for a := 0; a < 16*3; a++ {
		shim.Outp(0x3c9, 0)
	}

	if len(shim.VRAM) >= 65536 {
		clear(shim.VRAM[:65536])
	}

	koePlv = 0
	koePl = 1
	for !driver.WantsToQuit() && count > 0 {
		_ = music.GetFrame()

		if count < 333 {
			for repeat > 0 {
				repeat--
				xpos -= xposa / 4
				if xpos < 0 {
					xpos = 0
				} else {
					xposa++
				}
			}
			if xpos == 0 {
				break
			}
		}

		var wx, wy int
		if rot2 < 32 {
			wx = int(koeSin1024[(rot2+0)&1023])*rot2/8 + 160
			wy = int(koeSin1024[(rot2+256)&1023])*rot2/8 + 100
		} else {
			wx = int(koeSin1024[(rot2+0)&1023])/4 + 160
			wy = int(koeSin1024[(rot2+256)&1023])/4 + 100
		}
		rot2 += 17

		repeat = koeWaitBorder()
		count -= repeat
		_ = xpos & 7

		clear(koeVBuf[:koeVbufSize])

		hx := int(koeSin1024[(rot+0)&1023]) * 16 * 6 / 5
		hy := int(koeSin1024[(rot+256)&1023]) * 16
		vx := int(koeSin1024[(rot+256)&1023]) * 6 / 5
		vy := int(koeSin1024[(rot+512)&1023])
		vx = vx * (vm / 64) / 100
		vy = vy * (vm / 64) / 100
		for c := -10; c < 11; c += 2 {
			cx := vx * c * 2
			cy := vy * c * 2
			x1 := (-hx-vx+cx)/16 + wx
			y1 := (-hy-vy+cy)/16 + wy
			x2 := (-hx+vx+cx)/16 + wx
			y2 := (-hy+vy+cy)/16 + wy
			x3 := (+hx+vx+cx)/16 + wx
			y3 := (+hy+vy+cy)/16 + wy
			x4 := (+hx-vx+cx)/16 + wx
			y4 := (+hy-vy+cy)/16 + wy
			koeAsmBox(x1, y1, x2, y2, x3, y3, x4, y4)
		}
		rot += rota / 10
		vm += vma
		if vm < 0 {
			vm -= vma
			vma = -vma
		}
		vma--
		rota++

		koeAsmDoit(koeVBuf[:], koePlanarVRAM[koePl][koePlv][:])
		koeResolve16Color(koePlv, constants.ScreenWidth-xpos)

		koePlv += 2
		koePlv &= 7
		if koePlv == 0 {
			koePl = (koePl + 1) & 3
		}

		driver.Blit()
	}

	driver.ChangeMode(constants.ScreenWidth, constants.DoubleScreenHeight, shim.DefaultJSSS)
	if constants.ScreenWidth*constants.DoubleScreenHeight <= len(shim.VRAM) {
		clear(shim.VRAM[:constants.ScreenWidth*constants.DoubleScreenHeight])
	}
	driver.Blit()

	if driver.WantsToQuit() {
		return 0
	}

	h := blob.Open("troll.up")
	if h != nil {
		driver.Vsync(false)
		blob.Read(koePic[:40000], 40000, 1, h)
		driver.Vsync(false)
		blob.Read(koePic[40000:], 40000, 1, h)
	}

	driver.Vsync(false)

	picdata := make([]byte, constants.ScreenWidth*constants.DoubleScreenHeight)
	common.Readp(koePalette[:], -1, koePic[:])
	for y := 0; y < constants.DoubleScreenHeight; y++ {
		common.Readp(picdata[y*constants.ScreenWidth:], y, koePic[:])
	}

	pp := 0
	for y := 0; y < 16; y++ {
		x := 45 - y*3
		for a := 0; a < constants.PaletteByteCount; a++ {
			c := int(koePalette[a]) + x
			if c > 63 {
				c = 63
			}
			if pp < len(koePalFade) {
				koePalFade[pp] = byte(c)
				pp++
			}
		}
	}

	driver.Vsync(false)
	driver.Blit()
	common.SetPalArea(koePalette[:], 0, 256)

	for !driver.WantsToQuit() {
		b := music.GetRow()
		a := music.GetOrder()
		driver.Blit()
		if a > 35 || (a == 35 && b > 48) {
			break
		}
		driver.Vsync(false)
	}

	count = 300
	xposa = 0
	xpos = 0
	for !driver.WantsToQuit() && count > 0 {
		if xpos == constants.ScreenWidth {
			break
		}
		xpos += xposa / 4
		if xpos > constants.ScreenWidth {
			xpos = constants.ScreenWidth
		} else {
			xposa++
		}
		_ = xpos / 4
		count -= driver.Vsync(false)
		_ = (xpos & 3) * 2
		koeSideScroll256(picdata, constants.ScreenWidth-xpos)
		driver.Blit()
	}

	count = 50
	c := 0
	ripple := 0
	ripplep := 8
	for !driver.WantsToQuit() && count > 0 {
		if ripplep > 1023 {
			ripplep = 1024
		} else {
			ripplep = ripplep * 5 / 4
		}
		xpos = constants.ScreenWidth + int(koeSin1024[ripple&1023])/ripplep
		ripple += ripplep + 100
		_ = xpos / 4

		count -= driver.Vsync(false)
		_ = (xpos & 3) * 2

		if c < 16 {
			offset := c * constants.PaletteByteCount
			if offset < len(koePalFade) {
				common.SetPalArea(koePalFade[offset:], 0, 256)
			}
			c++
		}

		koeSideScroll256(picdata, constants.ScreenWidth-xpos)
		driver.Blit()
	}

	common.SetPalArea(koePalette[:], 0, 256)
	count = 420
	xpos = constants.ScreenWidth

	for !driver.WantsToQuit() && count > 0 {
		a := music.GetPlusFlags()
		if a > -6 && a < 16 {
			break
		}
		_ = xpos / 4
		count -= driver.Vsync(false)
		_ = (xpos & 3) * 2
		koeSideScroll256(picdata, constants.ScreenWidth-xpos)
		driver.Blit()
	}

	return 0
}
