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
	glenzNRows      = 256
	glenzRowsActive = 200
	glenzMaxLines   = 4096
)

const (
	glenzNE_X      = 0
	glenzNE_Y1     = 4
	glenzNE_Y2     = 6
	glenzNE_COLOR  = 8
	glenzNE_NEXT   = 10
	glenzNE_DX     = 14
	glenzNE_STRIDE = 18
)

var (
	glenzBackPal [16 * 3]byte = [16 * 3]byte{
		16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16,
		16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16,
	}
	glenzLightShift int8

	glenzDemoMode [3]func(ax, dx uint16, idx int)

	glenzProjXMul  int32
	glenzProjYMul  int32
	glenzProjXAdd  uint16
	glenzProjYAdd  uint16
	glenzProjMinZ  int32
	glenzProjMinZS uint16

	glenzWMinX int16
	glenzWMinY int16
	glenzWMaxX int16
	glenzWMaxY int16

	glenzCount uint16

	glenzBgPic  [65535]byte
	glenzFCRow  [100][]byte
	glenzFCRow2 [16][]byte

	glenzPal1   [constants.PaletteByteCount]byte
	glenzPal2   [constants.PaletteByteCount]byte
	glenzPal    [constants.PaletteByteCount]byte
	glenzTmpPal [constants.PaletteByteCount]byte

	glenzPoints2  [256]int32
	glenzPoints2B [256]int32
	glenzPoints3  [1024]uint16

	glenzEdges2   [64]int
	glenzPolylist [256]uint16
	glenzMatrix   [9]int16

	glenzRepeat int
	glenzFrame  int

	glenzRows [512]uint16

	glenzNewData1 [0x10000]byte
	glenzNDP0     uint32
	glenzNDP      uint32

	glenzNEP [glenzNRows]uint32
	glenzNL  [8192]uint32
	glenzNEC uint32 = glenzNE_STRIDE
	glenzNE  [glenzMaxLines * glenzNE_STRIDE]byte

	glenzYRow   uint32
	glenzYRowAd uint32

	glenzRolCol  [256]uint8
	glenzRolUsed [256]uint8

	glenzCF uint8

	glenzXAdd int32
	glenzYAdd int32
	glenzZAdd int32
	glenzM    [9]int32
)

func runGlenz() {
	if err := glenzEnsureData(); err != nil {
		log.Printf("glenz: %v", err)
		return
	}

	glenzReset()

	if !shim.IsDemoFirstPart() {
		for !driver.WantsToQuit() && music.GetPlusFlags() < -19 {
			driver.Vsync(false)
			driver.Blit()
		}
	}

	music.SetFrame(0)

	glenzZoomer2()

	for i := 0; i < 65535 && i < len(shim.VRAM); i++ {
		shim.VRAM[i] = 0
	}
	driver.ChangeMode(constants.ScreenWidth, constants.ScreenHeight, shim.DefaultJSSS)
	glenzInit320x200C()

	for a := 0; a < 100; a++ {
		base := constants.PaletteByteCount + 16 + a*constants.ScreenWidth
		if base+constants.ScreenWidth <= len(glenzFC) {
			glenzFCRow[a] = glenzFC[base : base+constants.ScreenWidth]
		}
	}
	for a := 0; a < 16; a++ {
		base := constants.PaletteByteCount + 16 + a*constants.ScreenWidth + 100*constants.ScreenWidth
		if base+constants.ScreenWidth <= len(glenzFC) {
			glenzFCRow2[a] = glenzFC[base : base+constants.ScreenWidth]
		}
	}

	shim.Outp(0x3c8, 0)
	for a := 0; a < constants.PaletteByteCount; a++ {
		shim.Outp(0x3c9, 0)
	}

	shim.Outp(0x3c8, 0)
	for a := 0; a < 16*3 && 16+a < len(glenzFC); a++ {
		shim.Outp(0x3c9, uint32(glenzFC[a+16]))
	}

	yy := 0
	ya := 0
	for !driver.WantsToQuit() {
		ya++
		yy += ya
		if yy > 48*16 {
			yy -= ya
			ya = -ya * 2 / 3
			if ya > -4 && ya < 4 {
				break
			}
		}
		y := yy / 16
		y1 := 130 + y/2
		y2 := 130 + y*3/2
		b := 0
		if y2 != y1 {
			b = 25600 / (y2 - y1)
		}
		pd := (y1 - 4) * constants.ScreenWidth
		for ry := y1 - 4; ry < y1; ry++ {
			if ry > 199 {
				pd += constants.ScreenWidth
				continue
			}
			for i := 0; i < constants.ScreenWidth; i++ {
				shim.VRAM[pd+i] = 0
			}
			pd += constants.ScreenWidth
		}
		for c, ry := 0, y1; ry < y2; ry, pd, c = ry+1, pd+constants.ScreenWidth, c+b {
			if ry > 199 {
				continue
			}
			row := c / 256
			if row >= 0 && row < len(glenzFCRow) && glenzFCRow[row] != nil {
				copy(shim.VRAM[pd:pd+constants.ScreenWidth], glenzFCRow[row])
			}
		}
		for c := 0; c < 16; c, pd = c+1, pd+constants.ScreenWidth {
			ry := y2 + c
			if ry > 199 {
				continue
			}
			if c > 7 {
				for i := 0; i < constants.ScreenWidth; i++ {
					shim.VRAM[pd+i] = 0
				}
			} else if glenzFCRow2[c] != nil {
				copy(shim.VRAM[pd:pd+constants.ScreenWidth], glenzFCRow2[c])
			}
		}
		driver.Blit()
		driver.Vsync(false)
	}

	for !driver.WantsToQuit() && music.GetFrame() < 300 {
		driver.Vsync(false)
		driver.Blit()
	}

	for a := 0; a < 16; a++ {
		ps := 0x10 + a*3
		if ps+2 < len(glenzFC) {
			glenzBackPal[a*3+0] = glenzFC[ps+0]
			glenzBackPal[a*3+1] = glenzFC[ps+1]
			glenzBackPal[a*3+2] = glenzFC[ps+2]
		}
	}

	pp := 0
	for a := 0; a < 256; a++ {
		b := a
		if a >= 16 {
			b = a & 7
		}
		r := int(glenzBackPal[b*3+0])
		g := int(glenzBackPal[b*3+1])
		bl := int(glenzBackPal[b*3+2])
		if (a&8) != 0 && a > 15 {
			r += 16
			g += 16
			bl += 16
		}
		if r > 63 {
			r = 63
		}
		if g > 63 {
			g = 63
		}
		if bl > 63 {
			bl = 63
		}
		glenzTmpPal[pp] = byte(r)
		pp++
		glenzTmpPal[pp] = byte(g)
		pp++
		glenzTmpPal[pp] = byte(bl)
		pp++
	}

	shim.Outp(0x3c8, 0)
	for a := 0; a < constants.PaletteByteCount; a++ {
		shim.Outp(0x3c9, uint32(glenzTmpPal[a]))
	}
	glenzLightShift = 9
	rx, ry, rz := 0, 0, 0
	ypos := -9000
	yposa := 0
	driver.Vsync(false)
	copy(glenzBgPic[:], shim.VRAM[:constants.ScreenSize])

	for !driver.WantsToQuit() && music.GetFrame() < 333 {
		driver.Vsync(false)
		driver.Blit()
	}

	copy(glenzPal[:], glenzBackPal[:])

	shim.Outp(0x3c7, 0)
	for a := 0; a < 16*3; a++ {
		glenzPal[a] = shim.Inp(0x3c9)
	}

	zpos := 7500
	boingm := 6
	boingd := 7
	jello := 0
	jelloa := 0
	xscale := 120
	yscale := 120
	zscale := 120
	bscale := 0
	oxp := 0
	oyp := 0
	ozp := 0
	oxb := 0
	oyb := 0
	ozb := 0

	for glenzFrame < 7000 && !driver.WantsToQuit() {
		if !shim.IsDemoFirstPart() {
			a := music.GetPlusFlags()
			if a < 0 && a > -16 {
				break
			}
		}

		glenzRepeat = driver.Vsync(false)
		shim.Outp(0x3c8, 0)
		for a := 0; a < 16*3; a++ {
			shim.Outp(0x3c9, uint32(glenzPal[a]))
		}

		for glenzRepeat > 0 {
			glenzRepeat--
			glenzFrame++
			rx += 32
			ry += 7
			rx %= 3 * 3600
			ry %= 3 * 3600
			rz %= 3 * 3600

			if glenzFrame > 900 {
				a := glenzFrame - 900
				b := glenzFrame - 900
				if b > 50 {
					b = 50
				}
				oxp = int(common.Sin1024[(a*3)&1023]) * b / 10
				oyp = int(common.Sin1024[(a*5)&1023]) * b / 10
				ozp = (int(common.Sin1024[(a*4)&1023])/2 + 128) * b / 16
				if glenzFrame > 1800 {
					a = glenzFrame - 1800 + 64
					if a > 1024 {
						a = 1024
					}
					oxb = -int(common.Sin1024[(a*6)&1023]) * a / 40
					oyb = -int(common.Sin1024[(a*7)&1023]) * a / 40
					ozb = (int(common.Sin1024[(a*8)&1023]) + 128) * a / 40
				} else {
					oxb = -int(common.Sin1024[(a*6)&1023])
					oyb = -int(common.Sin1024[(a*7)&1023])
					ozb = int(common.Sin1024[(a*8)&1023]) + 128
				}
				b = 1800 - glenzFrame
				if b < 0 {
					if b < -99 {
						b = -99
					}
					oyp -= b * b / 2
				}
			}

			if glenzFrame > 800 {
				if glenzFrame > 1220+789 {
					if xscale > 0 {
						xscale--
					}
					if yscale > 0 {
						yscale--
					}
					if zscale > 0 {
						zscale--
					}
					if bscale > 0 {
						bscale--
					}
				} else if glenzFrame > 1400+789 {
					if bscale > 0 {
						bscale -= 8
						if bscale < 0 {
							bscale = 0
						}
					}
				} else {
					if bscale < 180 {
						bscale += 2
					} else {
						bscale = 180
					}
				}
				if bscale > xscale {
					glenzLightShift = 10
				}
			} else {
				if glenzFrame < 640+70 {
					yposa += 31
					ypos += yposa / 40
					if ypos > -300 {
						ypos -= yposa / 40
						yposa = -yposa * boingm / boingd
						boingm += 2
						boingd++
					}
					if ypos > -900 && yposa > 0 {
						jello = (ypos + 900) * 5 / 3
						jelloa = 0
					}
				} else {
					if ypos > -2800 {
						ypos -= 16
					} else if ypos < -2800 {
						ypos += 16
					}
				}
				yscale = 120 + jello/30
				xscale = yscale
				zscale = 120 - jello/30
				a := jello
				jello += jelloa
				if (a < 0 && jello > 0) || (a > 0 && jello < 0) {
					jelloa = jelloa * 5 / 6
				}
				a = jello / 20
				jelloa -= a
			}

			if glenzFrame > 1280+789 {
				b := 1280 + 789 + 64 - glenzFrame
				if b < 0 {
					b = 0
				}
				for a := 0; a < 16*3; a++ {
					glenzPal[a] = byte(int(glenzBackPal[a]) * b / 64)
				}
			} else if glenzFrame > 700 {
				if glenzFrame < 765 {
					b := 764 - glenzFrame
					if b < 0 {
						b = 0
					}
					for a := 0; a < 16*3; a++ {
						glenzPal[a] = byte(int(glenzBackPal[a]) * b / 64)
					}
				} else if glenzFrame < 790 {
					y := 150 + (glenzFrame-765)*2
					if y >= 0 && y < constants.ScreenHeight {
						start := y * constants.ScreenWidth
						if start+640 <= len(glenzBgPic) {
							for i := 0; i < 640; i++ {
								glenzBgPic[start+i] = 0
							}
						}
						if start+640 <= len(shim.VRAM) {
							for i := 0; i < 640; i++ {
								shim.VRAM[start+i] = 0
							}
						}
					}
					if glenzFrame > 785 {
						for a := 0; a < 16; a++ {
							r := 0
							g := 0
							bl := 0
							if (a & 1) != 0 {
								r += 10
							}
							if (a & 2) != 0 {
								r += 30
							}
							if (a & 4) != 0 {
								r += 20
							}
							if (a & 8) != 0 {
								r += 16
								g += 16
								bl += 16
							}
							if r > 63 {
								r = 63
							}
							if g > 63 {
								g = 63
							}
							if bl > 63 {
								bl = 63
							}
							glenzBackPal[a*3+0] = byte(r)
							glenzBackPal[a*3+1] = byte(g)
							glenzBackPal[a*3+2] = byte(bl)
						}
					}
				} else if glenzFrame < 795 {
					copy(glenzPal[:], glenzBackPal[:])
				}
			}
		}

		glenzInitGroup()

		if xscale > 4 {
			glenzDemoMode[0] = glenzDemoMode[1]
			glenzMatrixYXZ(int16(rx), int16(ry), int16(rz), glenzMatrix[:])
			glenzSetMatrix(glenzMatrix[:], 0, 0, 0)
			glenzPoints2B[0] = 0
			glenzRotList(glenzPoints2B[:], glenzPoints)
			glenzMatrix[0] = int16(xscale * 64)
			glenzMatrix[1] = 0
			glenzMatrix[2] = 0
			glenzMatrix[3] = 0
			glenzMatrix[4] = int16(yscale * 64)
			glenzMatrix[5] = 0
			glenzMatrix[6] = 0
			glenzMatrix[7] = 0
			glenzMatrix[8] = int16(zscale * 64)
			glenzSetMatrix(glenzMatrix[:], int32(oxp), int32(ypos+1500+oyp), int32(zpos+ozp))
			glenzPoints2[0] = 0
			glenzRotList(glenzPoints2[:], glenzPoints2B[:])
			if glenzFrame < 800 {
				glenzClipList(glenzPoints2[:])
			}
			glenzPoints3[0] = 0
			glenzProjList(glenzPoints3[:], glenzPoints2[:])
			glenzCeasyPolylist(glenzPolylist[:], glenzEPolys, glenzPoints3[:])
			glenzPolyList(glenzPolylist[:])
		}

		if glenzFrame > 800 && bscale > 4 {
			glenzDemoMode[0] = glenzDemoMode[2]
			glenzMatrixYXZ(int16(3600-rx/3), int16(3600-ry/3), int16(3600-rz/3), glenzMatrix[:])
			glenzSetMatrix(glenzMatrix[:], 0, 0, 0)
			glenzPoints2B[0] = 0
			glenzRotList(glenzPoints2B[:], glenzPointsB)
			glenzMatrix[0] = int16(bscale * 64)
			glenzMatrix[1] = 0
			glenzMatrix[2] = 0
			glenzMatrix[3] = 0
			glenzMatrix[4] = int16(bscale * 64)
			glenzMatrix[5] = 0
			glenzMatrix[6] = 0
			glenzMatrix[7] = 0
			glenzMatrix[8] = int16(bscale * 64)
			glenzSetMatrix(glenzMatrix[:], int32(oxb), int32(ypos+1500+oyb), int32(zpos+ozb))
			glenzPoints2[0] = 0
			glenzRotList(glenzPoints2[:], glenzPoints2B[:])
			glenzPoints3[0] = 0
			glenzProjList(glenzPoints3[:], glenzPoints2[:])
			glenzCeasyPolylist(glenzPolylist[:], glenzEPolysB, glenzPoints3[:])
			glenzPolyList(glenzPolylist[:])
		}

		glenzDoneGroup()
		driver.Blit()
	}
}

func glenzClamp63(v uint32) uint8 {
	if v > 63 {
		return 63
	}
	return uint8(v)
}

func glenzSetPalXXX(r, g, b uint8) {
	if r > 63 {
		r = 63
	}
	if g > 63 {
		g = 63
	}
	if b > 63 {
		b = 63
	}
	shim.Outp(0x03C9, uint32(r))
	shim.Outp(0x03C9, uint32(g))
	shim.Outp(0x03C9, uint32(b))
}

func glenzCheckHiddenDet(polylist []uint16, idx int) int32 {
	x0 := int32(int16(polylist[idx+0]))
	y0 := int32(int16(polylist[idx+1]))
	x1 := int32(int16(polylist[idx+2]))
	y1 := int32(int16(polylist[idx+3]))
	x2 := int32(int16(polylist[idx+4]))
	y2 := int32(int16(polylist[idx+5]))
	a := int64(x0-x1) * int64(y0-y2)
	b := int64(y0-y1) * int64(x0-x2)
	det := a - b
	if det < 0 {
		glenzCF = 1
	} else {
		glenzCF = 0
	}
	return int32(det)
}

func glenzAddDotEmit(polylist []uint16, di *int, idx uint16, lastIdx *uint16, points3 []uint16) {
	if idx == *lastIdx {
		return
	}
	*lastIdx = idx
	base := 2 + int(idx)*6
	if base+1 >= len(points3) || *di+1 >= len(polylist) {
		return
	}
	polylist[*di] = points3[base]
	polylist[*di+1] = points3[base+1]
	*di += 2
}

func glenzDemoGlz2(ax, dx uint16, idx int) {
	if glenzCF != 0 {
		return
	}
	w := glenzPolylist[idx]
	al := uint8(w & 0xFF)
	newVal := (al >> 1) & 1
	glenzPolylist[idx] = (w & 0xFF00) | uint16(newVal)
}

func glenzDemoGlz(ax, dx uint16, idx int) {
	if glenzCF != 0 {
		w := glenzPolylist[idx]
		id := uint8(w & 0xFF)
		slot := glenzRolCol[id]
		glenzRolCol[id] = 0
		if slot != 0 {
			glenzRolUsed[slot] = 0
		}
		low := uint8(((id >> 1) & 1) << 2)
		glenzPolylist[idx] = (w & 0xFF00) | uint16(low)
		return
	}

	cat := (uint32(dx) << 16) | uint32(ax)
	var inten uint16
	if glenzLightShift != 9 {
		t1 := uint16((cat >> 8) & 0xFFFF)
		cat2 := (uint32(dx) << 16) | uint32(t1)
		t2 := uint16((cat2 >> 1) & 0xFFFF)
		inten = uint16(t1 + t2)
	} else {
		inten = uint16((cat >> 7) & 0xFFFF)
	}
	if inten > 63 {
		inten = 63
	}
	ah := uint8(inten)

	w := glenzPolylist[idx]
	bp := w
	id := uint8(w & 0xFF)
	slot := glenzRolCol[id]
	if slot == 0 {
		cand := uint8(2)
		for i := 0; i < 15; i++ {
			if glenzRolUsed[cand] == 0 {
				slot = cand
				break
			}
			cand += 2
		}
		if slot == 0 {
			slot = 2
		}
		glenzRolCol[id] = slot
		glenzRolUsed[slot] = 1
	}

	glenzPolylist[idx] = (w & 0xFF00) | uint16(slot<<3)
	shim.Outp(0x03C8, uint32(slot<<3))

	var r0, g0, b0 uint8
	if (bp & 2) != 0 {
		b0 = ah
		g0 = ah >> 1
		r0 = 0
	} else {
		r0 = ah
		g0 = ah
		b0 = ah
	}

	for i := 0; i < 16; i++ {
		ra := glenzBackPal[i*3+0] >> 2
		ga := glenzBackPal[i*3+1] >> 2
		ba := glenzBackPal[i*3+2] >> 2
		glenzSetPalXXX(glenzClamp63(uint32(r0)+uint32(ra)), glenzClamp63(uint32(g0)+uint32(ga)), glenzClamp63(uint32(b0)+uint32(ba)))
	}
}

func glenzCeasyPolylist(polylist []uint16, polys []uint16, points3 []uint16) {
	di := 0
	p := 0
	for {
		if p >= len(polys) {
			if di < len(polylist) {
				polylist[di] = 0
			}
			return
		}
		sides := polys[p]
		p++
		if sides == 0 {
			if di < len(polylist) {
				polylist[di] = 0
			}
			return
		}

		if di+1 >= len(polylist) {
			return
		}
		di++
		if p >= len(polys) {
			polylist[di] = 0
			return
		}
		polylist[di] = polys[p]
		p++
		di++
		cntPtr := di
		lastIdx := uint16(0xFFFF)
		for c := uint16(0); c < sides && p < len(polys); c++ {
			idx := polys[p]
			p++
			glenzAddDotEmit(polylist, &di, idx, &lastIdx, points3)
		}
		if di-cntPtr >= 2 {
			if polylist[cntPtr] == polylist[di-2] && polylist[cntPtr+1] == polylist[di-1] {
				di -= 2
			}
		}
		dwords := uint16((di - cntPtr) / 2)
		if cntPtr-2 >= 0 && cntPtr-2 < len(polylist) {
			polylist[cntPtr-2] = dwords
		}

		det := glenzCheckHiddenDet(polylist, cntPtr)
		ax := uint16(det & 0xFFFF)
		dx := uint16(uint32(det) >> 16)
		if cntPtr-1 >= 0 {
			glenzDemoMode[0](ax, dx, cntPtr-1)
		}
	}
}

func glenzNEWrite32(p uint32, v uint32) {
	if int(p)+4 > len(glenzNE) {
		return
	}
	glenzNE[p+0] = byte(v)
	glenzNE[p+1] = byte(v >> 8)
	glenzNE[p+2] = byte(v >> 16)
	glenzNE[p+3] = byte(v >> 24)
}

func glenzNEWrite16(p uint32, v uint16) {
	if int(p)+2 > len(glenzNE) {
		return
	}
	glenzNE[p+0] = byte(v)
	glenzNE[p+1] = byte(v >> 8)
}

func glenzNERead32(p uint32) uint32 {
	if int(p)+4 > len(glenzNE) {
		return 0
	}
	return uint32(glenzNE[p]) | (uint32(glenzNE[p+1]) << 8) | (uint32(glenzNE[p+2]) << 16) | (uint32(glenzNE[p+3]) << 24)
}

func glenzNERead16(p uint32) uint16 {
	if int(p)+2 > len(glenzNE) {
		return 0
	}
	return uint16(glenzNE[p]) | (uint16(glenzNE[p+1]) << 8)
}

func glenzEnsureSeed(base uint32) {
	if int(base)+4 > len(glenzNewData1) {
		return
	}
	dw := uint32(glenzNewData1[base]) | (uint32(glenzNewData1[base+1]) << 8) | (uint32(glenzNewData1[base+2]) << 16) | (uint32(glenzNewData1[base+3]) << 24)
	if dw == 0 {
		glenzNewData1[base+0] = 0xFF
		glenzNewData1[base+1] = 0xFF
		glenzNewData1[base+2] = 0
		glenzNewData1[base+3] = 0
	}
}

func glenzInitGroup() {
	glenzNDP0 ^= 0x8000
	glenzNDP0 &= 0x8000
	glenzNDP = glenzNDP0

	glenzNEC = glenzNE_STRIDE
	for i := range glenzNEP {
		glenzNEP[i] = 0
	}

	glenzEnsureSeed(0x0000)
	glenzEnsureSeed(0x8000)

	glenzYRow = 0
	glenzYRowAd = 0
}

func glenzAddEdgeIfVisible(x1, y1, x2, y2 int16, color uint16) uint32 {
	if y1 > y2 {
		x1, x2 = x2, x1
		y1, y2 = y2, y1
	}
	if y1 == y2 {
		return 0
	}
	num := (int32(x2) - int32(x1)) << 16
	den := int32(y2) - int32(y1)
	if den == 0 {
		return 0
	}
	dx := num / den
	xFixed := int32(x1) << 16

	if y1 < 0 {
		if y2 <= 0 {
			return 0
		}
		xFixed += dx * int32(-y1)
		y1 = 0
	}
	if y1 > 199 {
		return 0
	}
	if glenzNEC+glenzNE_STRIDE > uint32(len(glenzNE)) {
		return 0
	}
	e := glenzNEC
	glenzNEC += glenzNE_STRIDE

	glenzNEWrite32(e+glenzNE_X, uint32(xFixed))
	glenzNEWrite16(e+glenzNE_Y1, uint16(y1))
	glenzNEWrite16(e+glenzNE_Y2, uint16(y2))
	glenzNEWrite16(e+glenzNE_COLOR, color)
	glenzNEWrite32(e+glenzNE_DX, uint32(dx))

	head := glenzNEP[y1]
	if head == 0 {
		glenzNEP[y1] = e
		glenzNEWrite32(e+glenzNE_NEXT, 0)
		return e
	}
	it := head
	var prev uint32
	for it != 0 {
		if glenzNERead16(e+glenzNE_Y2) == glenzNERead16(it+glenzNE_Y2) &&
			glenzNERead32(e+glenzNE_X) == glenzNERead32(it+glenzNE_X) &&
			glenzNERead32(e+glenzNE_DX) == glenzNERead32(it+glenzNE_DX) {
			old := glenzNERead16(it + glenzNE_COLOR)
			merged := (old & 0xFF00) | ((old ^ color) & 0x00FF)
			glenzNEWrite16(it+glenzNE_COLOR, merged)
			glenzNEC -= glenzNE_STRIDE
			return 0
		}
		prev = it
		it = glenzNERead32(it + glenzNE_NEXT)
	}
	glenzNEWrite32(prev+glenzNE_NEXT, e)
	glenzNEWrite32(e+glenzNE_NEXT, 0)
	return e
}

func glenzNGPass2() {
	listEnd := 0
	glenzYRow = 0
	glenzYRowAd = 0

	for row := 0; row < glenzRowsActive; row++ {
		for e := glenzNEP[row]; e != 0; e = glenzNERead32(e + glenzNE_NEXT) {
			if listEnd < len(glenzNL) {
				glenzNL[listEnd] = e
				listEnd++
			}
		}
		if listEnd > 1 {
			for i := 1; i < listEnd; i++ {
				cur := glenzNL[i]
				x := int32(glenzNERead32(cur + glenzNE_X))
				j := i - 1
				for j >= 0 {
					prev := glenzNL[j]
					if x >= int32(glenzNERead32(prev+glenzNE_X)) {
						break
					}
					glenzNL[j+1] = prev
					j--
				}
				glenzNL[j+1] = cur
			}
		}

		ndpScanStart := glenzNDP
		ebpNdp := glenzNDP
		var cxLast uint16
		writeIdx := 0

		for i := 0; i < listEnd; i++ {
			e := glenzNL[i]
			y2 := int16(glenzNERead16(e + glenzNE_Y2))
			if int16(glenzYRow) >= y2 {
				continue
			}
			glenzNL[writeIdx] = e
			writeIdx++

			xFixed := int32(glenzNERead32(e + glenzNE_X))
			dx := int32(glenzNERead32(e + glenzNE_DX))
			xi := int16(xFixed >> 16)
			if xi > constants.ScreenWidth-1 {
				xi = constants.ScreenWidth - 1
			}
			if xi < 1 {
				xi = 1
			}
			color := glenzNERead16(e + glenzNE_COLOR)
			if ebpNdp > ndpScanStart && cxLast == uint16(xi) {
				prevOff := ebpNdp - 2
				old := uint16(glenzNewData1[prevOff]) | (uint16(glenzNewData1[prevOff+1]) << 8)
				old ^= color
				glenzNewData1[prevOff] = byte(old)
				glenzNewData1[prevOff+1] = byte(old >> 8)
			} else {
				pos := uint16(xi) + uint16(glenzYRowAd)
				glenzNewData1[ebpNdp+0] = byte(pos)
				glenzNewData1[ebpNdp+1] = byte(pos >> 8)
				glenzNewData1[ebpNdp+2] = byte(color)
				glenzNewData1[ebpNdp+3] = byte(color >> 8)
				ebpNdp += 4
				cxLast = uint16(xi)
			}
			glenzNEWrite32(e+glenzNE_X, uint32(int32(xFixed)+int32(dx)))
		}

		glenzNDP = ebpNdp
		listEnd = writeIdx

		glenzYRow++
		glenzYRowAd += constants.ScreenWidth
	}

	pos := uint16(63999)
	glenzNewData1[glenzNDP+0] = byte(pos)
	glenzNewData1[glenzNDP+1] = byte(pos >> 8)
	glenzNewData1[glenzNDP+2] = 0
	glenzNewData1[glenzNDP+3] = 0
	glenzNDP += 4

	glenzNewData1[glenzNDP+0] = 0xFF
	glenzNewData1[glenzNDP+1] = 0xFF
	glenzNewData1[glenzNDP+2] = 0
	glenzNewData1[glenzNDP+3] = 0
	glenzNDP += 4
}

func glenzNGPass3() {
	base := uint16(glenzNDP0)
	newp := uint32(base & 0x8000)
	lastp := newp ^ 0x8000

	edx := uint32(glenzNewData1[lastp]) | (uint32(glenzNewData1[lastp+1]) << 8) | (uint32(glenzNewData1[lastp+2]) << 16) | (uint32(glenzNewData1[lastp+3]) << 24)
	lastp += 4
	ecx := uint32(glenzNewData1[newp]) | (uint32(glenzNewData1[newp+1]) << 8) | (uint32(glenzNewData1[newp+2]) << 16) | (uint32(glenzNewData1[newp+3]) << 24)
	newp += 4

	var di uint16
	var al uint8
	var ah uint8

	for {
		cx := uint16(ecx)
		dx := uint16(edx)

		if dx < cx {
			if al != ah && dx > di {
				cnt := int(dx - di)
				baseDi := int(di)
				for i := 0; i < cnt; i++ {
					shim.VRAM[baseDi+i] = al | glenzBgPic[baseDi+i]
				}
			}
			di = dx
			ah ^= uint8(edx >> 16)
			edx = uint32(glenzNewData1[lastp]) | (uint32(glenzNewData1[lastp+1]) << 8) | (uint32(glenzNewData1[lastp+2]) << 16) | (uint32(glenzNewData1[lastp+3]) << 24)
			lastp += 4
		} else if dx == cx {
			if cx == 0xFFFF {
				break
			}
			if al != ah && dx > di {
				cnt := int(dx - di)
				baseDi := int(di)
				for i := 0; i < cnt; i++ {
					shim.VRAM[baseDi+i] = al | glenzBgPic[baseDi+i]
				}
			}
			di = dx
			ah ^= uint8(edx >> 16)
			edx = uint32(glenzNewData1[lastp]) | (uint32(glenzNewData1[lastp+1]) << 8) | (uint32(glenzNewData1[lastp+2]) << 16) | (uint32(glenzNewData1[lastp+3]) << 24)
			lastp += 4
		} else {
			if al != ah && cx > di {
				cnt := int(cx - di)
				baseDi := int(di)
				for i := 0; i < cnt; i++ {
					shim.VRAM[baseDi+i] = al | glenzBgPic[baseDi+i]
				}
			}
			di = cx
			al ^= uint8(ecx >> 16)
			ecx = uint32(glenzNewData1[newp]) | (uint32(glenzNewData1[newp+1]) << 8) | (uint32(glenzNewData1[newp+2]) << 16) | (uint32(glenzNewData1[newp+3]) << 24)
			newp += 4
		}
	}
}

func glenzNewGroup(mode uint16, pg []uint16) {
	if mode == 0 {
		glenzInitGroup()
		return
	}
	if mode == 1 {
		idx := 0
		for {
			if idx >= len(pg) {
				break
			}
			sides := pg[idx]
			if sides == 0 {
				break
			}
			color := pg[idx+1]
			verts := idx + 2
			first := verts
			prev := verts
			cur := verts + 2
			for k := uint16(0); k < sides; k++ {
				if k == sides-1 {
					cur = first
				}
				x1 := int16(pg[prev+0])
				y1 := int16(pg[prev+1])
				x2 := int16(pg[cur+0])
				y2 := int16(pg[cur+1])
				glenzAddEdgeIfVisible(x1, y1, x2, y2, color)
				prev = cur
				cur += 2
			}
			idx = verts + int(sides)*2
		}
		return
	}
	glenzNGPass2()
	glenzNGPass3()
}

func glenzPolyList(polylist []uint16) {
	glenzNewGroup(1, polylist)
}

func glenzDoneGroup() {
	glenzNewGroup(2, nil)
}

func glenzSetMatrix(m9 []int16, X, Y, Z int32) {
	glenzXAdd = X
	glenzYAdd = Y
	glenzZAdd = Z
	for i := 0; i < 9; i++ {
		glenzM[i] = int32(m9[i])
	}
}

func glenzDotQ15(a0, a1, a2, b0, b1, b2 int32) int32 {
	s := int64(a0)*int64(b0) + int64(a1)*int64(b1) + int64(a2)*int64(b2)
	return int32(s >> 15)
}

func glenzRotList(dst []int32, src []int32) int {
	if len(src) == 0 {
		return 0
	}
	srcCount := uint16(src[0] & 0xFFFF)
	dstHdr := dst[0]
	dstCount := uint16(dstHdr & 0xFFFF)
	newDstCount := uint16(dstCount + srcCount)
	dst[0] = (dstHdr & ^int32(0xFFFF)) | int32(newDstCount)

	out := 1 + int(dstCount)*3
	in := 1
	for i := uint16(0); i < srcCount; i++ {
		x := src[in+0]
		y := src[in+1]
		z := src[in+2]

		X := glenzDotQ15(glenzM[0], glenzM[1], glenzM[2], x, y, z) + glenzXAdd
		row1 := glenzDotQ15(glenzM[3], glenzM[4], glenzM[5], x, y, z)
		Y := glenzDotQ15(glenzM[6], glenzM[7], glenzM[8], x, y, z) + glenzYAdd
		Z := row1 + glenzZAdd

		dst[out+0] = X
		dst[out+1] = Y
		dst[out+2] = Z

		in += 3
		out += 3
	}
	return int(srcCount)
}

func glenzClipList(src []int32) {
	if len(src) == 0 {
		return
	}
	cx := uint16(src[0] & 0xFFFF)
	idx := 1
	for i := uint16(0); i < cx; i++ {
		if src[idx+1] >= 1500 {
			src[idx+1] = 1500
		}
		idx += 3
	}
}

func glenzProjList(dst []uint16, src []int32) {
	if len(src) == 0 || len(dst) < 2 {
		return
	}
	cx := uint16(src[0] & 0xFFFF)
	glenzCount = cx
	have := dst[0]
	dst[0] = have + cx

	out := 2 + int(have)*6
	in := 1
	for i := uint16(0); i < cx; i++ {
		X := src[in+0]
		Y := src[in+1]
		Z := src[in+2]

		dst[out+4] = uint16(Z & 0xFFFF)
		dst[out+5] = uint16(uint32(Z) >> 16)

		Zdiv := Z
		bp := uint16(0)
		if Zdiv < glenzProjMinZ {
			Zdiv = glenzProjMinZ
			bp |= 16
		}

		qY := int32(int64(Y) * int64(glenzProjYMul) / int64(Zdiv))
		y := int16(qY + int32(glenzProjYAdd))
		if y > glenzWMaxY {
			bp |= 8
		}
		if y < glenzWMinY {
			bp |= 4
		}
		dst[out+1] = uint16(y)

		qX := int32(int64(X) * int64(glenzProjXMul) / int64(Zdiv))
		x := int16(qX + int32(glenzProjXAdd))
		if x > glenzWMaxX {
			bp |= 2
		}
		if x < glenzWMinX {
			bp |= 1
		}
		dst[out+0] = uint16(x)
		dst[out+2] = bp
		dst[out+3] = 0

		in += 3
		out += 6
	}
}

func glenzSetRows(dx, lcount uint16) {
	ax := uint16(0)
	for i := uint16(0); i < lcount && int(i) < len(glenzRows); i++ {
		glenzRows[i] = ax
		ax = ax + dx
	}
}

func glenzInit320x200C() {
	glenzProjXMul = 256
	glenzProjYMul = 213
	glenzProjXAdd = 160
	glenzProjYAdd = 130
	glenzProjMinZ = 128
	glenzProjMinZS = 7
	glenzWMinX = 0
	glenzWMinY = 0
	glenzWMaxX = constants.ScreenWidth - 1
	glenzWMaxY = constants.ScreenHeight - 1
	glenzSetRows(constants.PlanarWidth, 200)
}

func glenzWrapDeg10(a int16) int16 {
	x := int32(a)
	for x >= 3600 {
		x -= 3600
	}
	for x < 0 {
		x += 3600
	}
	return int16(x)
}

func glenzQ15MulExact(a, b int16) int16 {
	p := int32(a) * int32(b)
	up := uint32(p)
	shr := up >> 15
	return int16(shr & 0xFFFF)
}

func glenzAdd16(acc, v int16) int16 {
	return int16(int32(acc) + int32(v))
}

func glenzSub16(acc, v int16) int16 {
	return int16(int32(acc) - int32(v))
}

func glenzTrigDeg10(a int16) (int16, int16) {
	w := glenzWrapDeg10(a)
	if int(w) >= len(glenzSinTable) {
		return 0, 0
	}
	return glenzSinTable[w], glenzCosTable[w]
}

func glenzCalcMatrix(rx, ry, rz int16, d []int16) {
	sx, cx := glenzTrigDeg10(rx)
	sy, cy := glenzTrigDeg10(ry)
	sz, cz := glenzTrigDeg10(rz)

	xsYs := glenzQ15MulExact(sx, sy)
	xsYc := glenzQ15MulExact(sx, cy)
	xcYs := glenzQ15MulExact(cx, sy)
	xcYc := glenzQ15MulExact(cx, cy)

	v0 := glenzQ15MulExact(cy, cz)
	v0 = glenzSub16(v0, glenzQ15MulExact(xsYs, sz))
	d[0] = v0

	v2 := glenzQ15MulExact(xsYs, cz)
	v2 = glenzAdd16(v2, glenzQ15MulExact(cy, sz))
	d[1] = v2

	d[2] = int16(-xcYs)
	d[3] = int16(-glenzQ15MulExact(cx, sz))
	d[4] = glenzQ15MulExact(cx, cz)
	d[5] = sx

	v12 := glenzQ15MulExact(xsYc, sz)
	v12 = glenzAdd16(v12, glenzQ15MulExact(sy, cz))
	d[6] = v12

	v14 := glenzQ15MulExact(sy, sz)
	v14 = glenzSub16(v14, glenzQ15MulExact(xsYc, cz))
	d[7] = v14

	d[8] = xcYc
}

func glenzMatrixYXZ(rx, ry, rz int16, dst []int16) {
	glenzCalcMatrix(ry, rx, rz, dst)
}

func glenzCalcMatrix0(rx, ry, rz int16, d []int16) {
	sx, cx := glenzTrigDeg10(rx)
	sy, cy := glenzTrigDeg10(ry)
	sz, cz := glenzTrigDeg10(rz)

	xsYs := glenzQ15MulExact(sx, sy)
	xcYs := glenzQ15MulExact(cx, sy)

	d[0] = glenzQ15MulExact(cy, cz)
	d[1] = glenzQ15MulExact(cy, sz)
	d[2] = int16(-sy)

	r10 := glenzQ15MulExact(xsYs, cz)
	r10 = glenzSub16(r10, glenzQ15MulExact(cx, sz))
	d[3] = r10

	r11 := glenzQ15MulExact(xsYs, sz)
	r11 = glenzAdd16(r11, glenzQ15MulExact(cx, cz))
	d[4] = r11

	d[5] = glenzQ15MulExact(cy, sx)

	r20 := glenzQ15MulExact(xcYs, cz)
	r20 = glenzAdd16(r20, glenzQ15MulExact(sx, sz))
	d[6] = r20

	r21 := glenzQ15MulExact(xcYs, sz)
	r21 = glenzSub16(r21, glenzQ15MulExact(sx, cz))
	d[7] = r21

	d[8] = glenzQ15MulExact(cy, cx)
}

func glenzCalcMatrixSep(rx, ry, rz int16, dst []int16) {
	sx, cx := glenzTrigDeg10(rx)
	sy, cy := glenzTrigDeg10(ry)
	sz, cz := glenzTrigDeg10(rz)

	m := dst
	m[0] = 32767
	m[1] = 0
	m[2] = 0
	m[3] = 0
	m[4] = cx
	m[5] = sx
	m[6] = 0
	m[7] = int16(-sx)
	m[8] = cx

	m = m[9:]
	m[0] = cy
	m[1] = 0
	m[2] = int16(-sy)
	m[3] = 0
	m[4] = 32767
	m[5] = 0
	m[6] = sy
	m[7] = 0
	m[8] = cy

	m = m[9:]
	m[0] = cz
	m[1] = sz
	m[2] = 0
	m[3] = int16(-sz)
	m[4] = cz
	m[5] = 0
	m[6] = 0
	m[7] = 0
	m[8] = 32767
}

func glenzMulMatrices(m1 []int16, m2 []int16) {
	a00, a01, a02 := m1[0], m1[1], m1[2]
	a10, a11, a12 := m1[3], m1[4], m1[5]
	a20, a21, a22 := m1[6], m1[7], m1[8]

	b00, b01, b02 := m2[0], m2[1], m2[2]
	b10, b11, b12 := m2[3], m2[4], m2[5]
	b20, b21, b22 := m2[6], m2[7], m2[8]

	r00 := glenzQ15MulExact(a00, b00)
	r00 = glenzAdd16(r00, glenzQ15MulExact(a01, b10))
	r00 = glenzAdd16(r00, glenzQ15MulExact(a02, b20))

	r01 := glenzQ15MulExact(a00, b01)
	r01 = glenzAdd16(r01, glenzQ15MulExact(a01, b11))
	r01 = glenzAdd16(r01, glenzQ15MulExact(a02, b21))

	r02 := glenzQ15MulExact(a00, b02)
	r02 = glenzAdd16(r02, glenzQ15MulExact(a01, b12))
	r02 = glenzAdd16(r02, glenzQ15MulExact(a02, b22))

	r10 := glenzQ15MulExact(a10, b00)
	r10 = glenzAdd16(r10, glenzQ15MulExact(a11, b10))
	r10 = glenzAdd16(r10, glenzQ15MulExact(a12, b20))

	r11 := glenzQ15MulExact(a10, b01)
	r11 = glenzAdd16(r11, glenzQ15MulExact(a11, b11))
	r11 = glenzAdd16(r11, glenzQ15MulExact(a12, b21))

	r12 := glenzQ15MulExact(a10, b02)
	r12 = glenzAdd16(r12, glenzQ15MulExact(a11, b12))
	r12 = glenzAdd16(r12, glenzQ15MulExact(a12, b22))

	r20 := glenzQ15MulExact(a20, b00)
	r20 = glenzAdd16(r20, glenzQ15MulExact(a21, b10))
	r20 = glenzAdd16(r20, glenzQ15MulExact(a22, b20))

	r21 := glenzQ15MulExact(a20, b01)
	r21 = glenzAdd16(r21, glenzQ15MulExact(a21, b11))
	r21 = glenzAdd16(r21, glenzQ15MulExact(a22, b21))

	r22 := glenzQ15MulExact(a20, b02)
	r22 = glenzAdd16(r22, glenzQ15MulExact(a21, b12))
	r22 = glenzAdd16(r22, glenzQ15MulExact(a22, b22))

	m1[0] = r00
	m1[1] = r01
	m1[2] = r02
	m1[3] = r10
	m1[4] = r11
	m1[5] = r12
	m1[6] = r20
	m1[7] = r21
	m1[8] = r22
}

func glenzZoomer2() {
	shim.Outp(0x3c7, 0)
	for a := 0; a < constants.PaletteByteCount; a++ {
		glenzPal1[a] = shim.Inp(0x3c9)
	}

	zy := 0
	zya := 0
	zly := 0
	zy2 := 0
	zly2 := 0
	framez2 := 0

	for !driver.WantsToQuit() {
		if zy == 260 {
			break
		}
		zly = zy
		zya++
		zy += zya / 4
		if zy > 260 {
			zy = 260
		}
		v := zly * constants.PlanarWidth * 4
		for y := zly; y <= zy; y++ {
			for i := 0; i < constants.PlanarWidth*4; i++ {
				shim.VRAM[v+i] = 255
			}
			v += constants.PlanarWidth * 4
		}
		zly2 = zy2
		zy2 = 125 * zy / 260
		v = (399 - zy2) * constants.PlanarWidth * 4
		for y := zly2; y <= zy2; y++ {
			for i := 0; i < constants.PlanarWidth*4; i++ {
				shim.VRAM[v+i] = 255
			}
			v += constants.PlanarWidth * 4
		}
		c := framez2
		if c > 32 {
			c = 32
		}
		b := 32 - c
		for a := 0; a < 128*3; a++ {
			glenzPal2[a] = byte((int(glenzPal1[a])*b + 30*c) >> 5)
		}
		framez2++
		driver.Vsync(false)
		common.SetPalArea(glenzPal2[:], 0, 128)
		driver.Blit()
	}

	shim.Outp(0x3c8, 0)
	shim.Outp(0x3c9, 0)
	shim.Outp(0x3c9, 0)
	shim.Outp(0x3c9, 0)
	v := 0
	for y := 0; y <= 399; y++ {
		for i := 0; i < constants.PlanarWidth*4; i++ {
			shim.VRAM[v+i] = 0
		}
		v += constants.PlanarWidth * 4
	}
}

func glenzReset() {
	for i := range glenzBackPal {
		glenzBackPal[i] = 16
	}
	glenzLightShift = 0
	glenzDemoMode[0] = glenzDemoGlz
	glenzDemoMode[1] = glenzDemoGlz
	glenzDemoMode[2] = glenzDemoGlz2

	glenzProjXMul = 0
	glenzProjYMul = 0
	glenzProjXAdd = 0
	glenzProjYAdd = 0
	glenzProjMinZ = 0
	glenzProjMinZS = 0

	glenzWMinX = 0
	glenzWMinY = 0
	glenzWMaxX = 100
	glenzWMaxY = 100

	glenzCount = 0

	for i := range glenzBgPic {
		glenzBgPic[i] = 0
	}
	for i := range glenzFCRow {
		glenzFCRow[i] = nil
	}
	for i := range glenzFCRow2 {
		glenzFCRow2[i] = nil
	}
	for i := range glenzPal1 {
		glenzPal1[i] = 0
		glenzPal2[i] = 0
		glenzPal[i] = 0
		glenzTmpPal[i] = 0
	}
	for i := range glenzPoints2 {
		glenzPoints2[i] = 0
		glenzPoints2B[i] = 0
	}
	for i := range glenzPoints3 {
		glenzPoints3[i] = 0
	}
	for i := range glenzEdges2 {
		glenzEdges2[i] = 0
	}
	for i := range glenzPolylist {
		glenzPolylist[i] = 0
	}
	for i := range glenzMatrix {
		glenzMatrix[i] = 0
	}
	glenzRepeat = 0
	glenzFrame = 0
	for i := range glenzRows {
		glenzRows[i] = 0
	}
	for i := range glenzNewData1 {
		glenzNewData1[i] = 0
	}
	glenzNDP0 = 0
	glenzNDP = 0
	for i := range glenzNEP {
		glenzNEP[i] = 0
	}
	for i := range glenzNL {
		glenzNL[i] = 0
	}
	glenzNEC = glenzNE_STRIDE
	for i := range glenzNE {
		glenzNE[i] = 0
	}
	glenzYRow = 0
	glenzYRowAd = 0
	for i := range glenzRolCol {
		glenzRolCol[i] = 0
		glenzRolUsed[i] = 0
	}
	glenzCF = 0
	glenzXAdd = 0
	glenzYAdd = 0
	glenzZAdd = 0
	for i := range glenzM {
		glenzM[i] = 0
	}
}
