package parts

import (
	"encoding/binary"

	"go-secondreality/dck/internal/common"
	"go-secondreality/dck/internal/constants"
	"go-secondreality/dck/internal/driver"
	"go-secondreality/dck/internal/music"
	"go-secondreality/dck/internal/shim"
)

type dotsDot struct {
	x    int16
	y    int16
	z    int16
	old1 int16
	old2 int16
	old3 int16
	old4 int16
	yadd int16
}

var (
	dotsDotNum int

	dotsRotSin int16
	dotsRotCos int16

	dotsBgPic [constants.ScreenSize]byte
	dotsRows  [constants.ScreenHeight]int16

	dotsDepth1 [128]int32
	dotsDepth2 [128]int32
	dotsDepth3 [128]int32

	dotsPal  common.Palette
	dotsPal2 common.Palette

	dotsDots [1024]dotsDot

	dotsGravityBottom int16
	dotsGravity       int16
	dotsGravityD      int16

	dotsTau [1024]int

	dotsRandSeed int32 = 10
)

var dotsCols = [...]int{0, 0, 0, 4, 25, 30, 8, 40, 45, 16, 55, 60}

func dotsRand() int {
	dotsRandSeed = dotsRandSeed*1103515245 + 12345
	val := int32(dotsRandSeed / 65536)
	return int(uint32(val) & 0x7FFF)
}

func dotsSetPalette(p []byte) {
	common.SetPalArea(p, 0, constants.PaletteColorCount)
}

func dotsSin(deg int) int {
	return int(common.Sin1024[deg&1023])
}

func dotsCos(deg int) int {
	return int(common.Sin1024[(deg+256)&1023])
}

func dotsRestoreBall(pos uint16) {
	p := int(pos)
	dotsCopyBG(p, 4)
	dotsCopyBG(p+constants.ScreenWidth, 4)
	dotsCopyBG(p+2*constants.ScreenWidth, 4)
}

func dotsRestoreShadow(pos uint16) {
	p := int(pos)
	dotsCopyBG(p, 2)
}

func dotsCopyBG(pos, count int) {
	if pos < 0 || count <= 0 {
		return
	}
	if pos+count > len(shim.VRAM) || pos+count > len(dotsBgPic) {
		return
	}
	copy(shim.VRAM[pos:pos+count], dotsBgPic[pos:pos+count])
}

func dotsDrawShadow(pos int) {
	if pos < 0 || pos+2 > len(shim.VRAM) {
		return
	}
	shim.VRAM[pos+0] = 87
	shim.VRAM[pos+1] = 87
}

func dotsPutUint16(pos int, val uint16) {
	if pos < 0 || pos+2 > len(shim.VRAM) {
		return
	}
	binary.LittleEndian.PutUint16(shim.VRAM[pos:pos+2], val)
}

func dotsPutUint32(pos int, val uint32) {
	if pos < 0 || pos+4 > len(shim.VRAM) {
		return
	}
	binary.LittleEndian.PutUint32(shim.VRAM[pos:pos+4], val)
}

func dotsDrawBallFromDepth(pos int, idx int) {
	if idx < 0 || idx >= len(dotsDepth1) {
		return
	}
	dotsPutUint16(pos+1, uint16(dotsDepth1[idx]))
	dotsPutUint32(pos+constants.ScreenWidth, uint32(dotsDepth2[idx]))
	dotsPutUint16(pos+2*constants.ScreenWidth+1, uint16(dotsDepth3[idx]))
}

func dotsDraw() {
	for k := 0; k < dotsDotNum; k++ {
		currentDot := &dotsDots[k]

		x := int32(currentDot.x)
		z := int32(currentDot.z)

		t1 := z * int32(dotsRotCos)
		t2 := x * int32(dotsRotSin)
		depth32 := t1 - t2
		bp := depth32 >> 16
		bp += 9000
		if bp == 0 {
			bp = 1
		}

		sum32 := x*int32(dotsRotCos) + z*int32(dotsRotSin)
		xnum := (sum32 >> 8) + (sum32 >> 11)
		sx := int(xnum/bp) + 160

		if sx < 0 || sx > 319 {
			dotsRestoreBall(uint16(currentDot.old2))
			dotsRestoreShadow(uint16(currentDot.old1))
			continue
		}

		syShadow := int((int32(8<<16))/bp) + 100
		if syShadow >= 0 && syShadow <= 199 {
			rowOff := int(uint16(dotsRows[syShadow]))
			shPos := rowOff + sx

			dotsRestoreShadow(uint16(currentDot.old1))
			dotsDrawShadow(shPos)
			currentDot.old1 = int16(shPos)
		} else {
			dotsRestoreShadow(uint16(currentDot.old1))
		}

		currentDot.yadd = int16(int32(currentDot.yadd) + int32(dotsGravity))

		yNew := int32(currentDot.y) + int32(currentDot.yadd)
		if yNew >= int32(dotsGravityBottom) {
			v := int32(currentDot.yadd)
			v = -v
			v = (v * int32(dotsGravityD)) >> 4
			currentDot.yadd = int16(v)
			yNew += v
		}
		currentDot.y = int16(yNew)

		ynum := (int32(currentDot.y) << 6) / bp
		sy := int(ynum) + 100

		if sy >= 0 && sy <= 199 {
			rowOff := int(uint16(dotsRows[sy]))
			pos := rowOff + sx

			dotsRestoreBall(uint16(currentDot.old2))

			idx := int((bp >> 6) &^ 3)
			if idx < 0 {
				idx = 0
			} else if idx > 127 {
				idx = 127
			}

			dotsDrawBallFromDepth(pos, idx)
			currentDot.old2 = int16(pos)
		} else {
			dotsRestoreBall(uint16(currentDot.old2))
		}
	}
}

func runDots() {
	clear(dotsBgPic[:])
	clear(dotsRows[:])
	clear(dotsDepth1[:])
	clear(dotsDepth2[:])
	clear(dotsDepth3[:])
	clear(dotsPal[:])
	clear(dotsPal2[:])
	clear(dotsDots[:])
	clear(dotsTau[:])

	dotsDotNum = 0
	dotsRotSin = 0
	dotsRotCos = 0
	dotsGravityBottom = 8000
	dotsGravity = 0
	dotsGravityD = 16
	dotsRandSeed = 10

	dropper := 0
	repeat := 0
	frame := 0
	rota := -1 * 64
	rot := 0
	rots := 0
	a := 0
	b := 0
	c := 0
	d := 0
	i := 0
	j := 0
	grav := 0
	gravd := 0
	f := 0

	shim.ClearScreen()

	dotsDotNum = 512

	for a = 0; a < dotsDotNum; a++ {
		dotsTau[a] = a
	}

	for a = 0; a < 500; a++ {
		b = dotsRand() % dotsDotNum
		c = dotsRand() % dotsDotNum
		d = dotsTau[b]
		dotsTau[b] = dotsTau[c]
		dotsTau[c] = d
	}

	dropper = 22000
	for a = 0; a < dotsDotNum; a++ {
		dotsDots[a].x = 0
		dotsDots[a].y = int16(2560 - dropper)
		dotsDots[a].z = 0
		dotsDots[a].yadd = 0
		dotsDots[a].old1 = 0
		dotsDots[a].old2 = 0
	}

	grav = 3
	gravd = 13
	dotsGravityBottom = 8105

	for a = 0; a < 500; a++ {
		b = dotsRand() % dotsDotNum
		c = dotsRand() % dotsDotNum
		d = int(dotsDots[b].x)
		dotsDots[b].x = dotsDots[c].x
		dotsDots[c].x = int16(d)
		d = int(dotsDots[b].y)
		dotsDots[b].y = dotsDots[c].y
		dotsDots[c].y = int16(d)
		d = int(dotsDots[b].z)
		dotsDots[b].z = dotsDots[c].z
		dotsDots[c].z = int16(d)
	}

	for a = 0; a < constants.ScreenHeight; a++ {
		dotsRows[a] = int16(a * constants.ScreenWidth)
	}

	shim.Outp(0x3c8, 0)
	for a = 0; a < 16; a++ {
		for b = 0; b < 4; b++ {
			c = 100 + a*9
			shim.Outp(0x3c9, uint32(dotsCols[b*3+0]))
			shim.Outp(0x3c9, uint32(dotsCols[b*3+1]*c/256))
			shim.Outp(0x3c9, uint32(dotsCols[b*3+2]*c/256))
		}
	}

	shim.Outp(0x3c8, 255)
	shim.Outp(0x3c9, 31)
	shim.Outp(0x3c9, 0)
	shim.Outp(0x3c9, 15)
	shim.Outp(0x3c8, 64)

	for a = 0; a < 100; a++ {
		c = 64 - 256/(a+4)
		c = c * c / 64
		shim.Outp(0x3c9, uint32(c/4))
		shim.Outp(0x3c9, uint32(c/4))
		shim.Outp(0x3c9, uint32(c/4))
	}

	shim.Outp(0x3c7, 0)
	for a = 0; a < constants.PaletteByteCount; a++ {
		dotsPal[a] = shim.Inp(0x3c9)
	}

	shim.Outp(0x3c8, 0)
	for a = 0; a < constants.PaletteByteCount; a++ {
		shim.Outp(0x3c9, 0)
	}

	for a = 0; a < 100; a++ {
		row := (100 + a) * constants.ScreenWidth
		if row < 0 || row+constants.ScreenWidth > len(shim.VRAM) {
			break
		}
		for x := 0; x < constants.ScreenWidth; x++ {
			shim.VRAM[row+x] = byte(a + 64)
		}
	}

	for a = 0; a < 128; a++ {
		c = a - (43+20)/2
		c = c * 3 / 4
		c += 8
		if c < 0 {
			c = 0
		} else if c > 15 {
			c = 15
		}
		c = 15 - c
		dotsDepth1[a] = int32(0x202 + 0x04040404*c)
		dotsDepth2[a] = int32(0x02030302 + 0x04040404*c)
		dotsDepth3[a] = int32(0x202 + 0x04040404*c)
	}

	copy(dotsBgPic[:], shim.VRAM[:constants.ScreenSize])

	for b = 64; b >= 0; b-- {
		for c = 0; c < constants.PaletteByteCount; c++ {
			a = int(dotsPal[c]) - b
			if a < 0 {
				a = 0
			}
			dotsPal2[c] = byte(a)
		}

		driver.Vsync(false)
		driver.Vsync(false)
		dotsSetPalette(dotsPal2[:])
		driver.Blit()
	}

	for !driver.WantsToQuit() && frame < 2450 {
		repeat = driver.Vsync(false)

		if frame > 2300 {
			dotsSetPalette(dotsPal2[:])
		}

		a = music.GetPlusFlags()
		if a > -4 && a < 0 {
			break
		}

		for repeat > 0 {
			repeat--
			frame++

			if frame == 500 {
				f = 0
			}

			i = dotsTau[j]
			j++
			if j >= dotsDotNum {
				j = 0
			}

			if frame < 500 {
				dotsDots[i].x = int16(dotsSin(f*11) * 40)
				dotsDots[i].y = int16(dotsCos(f*13)*10 - dropper)
				dotsDots[i].z = int16(dotsSin(f*17) * 40)
				dotsDots[i].yadd = 0
			} else if frame < 900 {
				dotsDots[i].x = int16(dotsCos(f*15) * 55)
				dotsDots[i].y = int16(dropper)
				dotsDots[i].z = int16(dotsSin(f*15) * 55)
				dotsDots[i].yadd = -260
			} else if frame < 1700 {
				a = dotsSin(frame) / 8
				dotsDots[i].x = int16(dotsCos(f*66) * a)
				dotsDots[i].y = 8000
				dotsDots[i].z = int16(dotsSin(f*66) * a)
				dotsDots[i].yadd = -300
			} else if frame < 2360 {
				dotsDots[i].x = int16(dotsRand() - 16384)
				dotsDots[i].y = int16(8000 - dotsRand()/2)
				dotsDots[i].z = int16(dotsRand() - 16384)
				dotsDots[i].yadd = 0
				if frame > 1900 && (frame&31) == 0 && grav > 0 {
					grav--
				}
			} else if frame < 2400 {
				a = frame - 2360
				for b = 0; b < constants.PaletteByteCount; b += 3 {
					c = int(dotsPal[b+0]) + a*3
					if c > 63 {
						c = 63
					}
					dotsPal2[b+0] = byte(c)
					c = int(dotsPal[b+1]) + a*3
					if c > 63 {
						c = 63
					}
					dotsPal2[b+1] = byte(c)
					c = int(dotsPal[b+2]) + a*4
					if c > 63 {
						c = 63
					}
					dotsPal2[b+2] = byte(c)
				}
			} else if frame < 2440 {
				a = frame - 2400
				for b = 0; b < constants.PaletteByteCount; b += 3 {
					c = 63 - a*2
					if c < 0 {
						c = 0
					}
					dotsPal2[b+0] = byte(c)
					dotsPal2[b+1] = byte(c)
					dotsPal2[b+2] = byte(c)
				}
			}

			if dropper > 4000 {
				dropper -= 100
			}

			dotsRotCos = int16(dotsCos(rot) * 64)
			dotsRotSin = int16(dotsSin(rot) * 64)
			rots += 2

			if frame > 1900 {
				rot += rota / 64
				rota--
			} else {
				rot = dotsSin(rots)
			}

			f++
			dotsGravity = int16(grav)
			dotsGravityD = int16(gravd)
		}

		dotsDraw()
		driver.Blit()
	}
}
