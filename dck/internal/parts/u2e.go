package parts

import (
	"encoding/binary"
	"fmt"
	"log"

	"go-secondreality/dck/internal/common"
	"go-secondreality/dck/internal/constants"
	"go-secondreality/dck/internal/driver"
	"go-secondreality/dck/internal/music"
	"go-secondreality/dck/internal/shim"
)

const u2eSceneName = "U2E"

func u2eCopper2() {
	visuSyncFrame++

	for i := 0; i < 4; i++ {
		if visuCl[i].ready == 2 {
			visuCl[i].ready = 0
		}
	}

	visuDeadlock++
	visuCopperCnt++

	if visuCopperDelay > 0 {
		visuCopperDelay--
	}
	if visuCopperDelay > 0 {
		return
	}
	visuCopperDelay = 0

	if visuCl[visuClr].ready != 0 {
		visuCl[(visuClr-1)&3].ready = 2
		visuCopperDelay = visuCl[visuClr].frames
		visuClr = (visuClr + 1) & 3
	} else {
		visuAvgRepeat++
	}
}

func u2eFadeSet(vram []byte) {
	const left = 17 * 4
	const mid = 47 * 4
	const right = 16 * 4
	if left+mid+right != constants.ScreenWidth {
		return
	}

	for y := 0; y < 200; y++ {
		var leftVal, midVal, rightVal byte
		if y < 25 || y >= 175 {
			leftVal = 0
			midVal = 252
			rightVal = 0
		} else {
			leftVal = 254
			midVal = 253
			rightVal = 254
		}
		for dy := 0; dy < 2; dy++ {
			row := (y*2 + dy) * constants.ScreenWidth
			if row+constants.ScreenWidth > len(vram) {
				continue
			}
			line := vram[row : row+constants.ScreenWidth]
			for x := 0; x < left; x++ {
				line[x] = leftVal
			}
			for x := 0; x < mid; x++ {
				line[left+x] = midVal
			}
			for x := 0; x < right; x++ {
				line[left+mid+x] = rightVal
			}
		}
	}
}

func runU2E() {
	visuReset()

	visuScene0 = nil
	visuScenem = nil

	scene0 := visuReadFile(fmt.Sprintf("%s.00M", u2eSceneName))
	if len(scene0) == 0 {
		log.Printf("u2e: missing scene data")
		return
	}
	visuScene0 = scene0
	visuScenem = scene0

	if len(visuScene0) > 15 {
		switch visuScene0[15] {
		case 'C':
			visuCity = 1
		case 'R':
			visuCity = 2
		}
	}

	if len(visuScene0) >= 8 {
		offset := int(int32(binary.LittleEndian.Uint32(visuScene0[4:8])))
		if offset < 0 {
			offset = 0
		}
		if offset+2 <= len(visuScene0) {
			pos := offset
			visuCoNum = int(int16(binary.LittleEndian.Uint16(visuScene0[pos : pos+2])))
			if visuCoNum < 0 {
				visuCoNum = 0
			}
			if visuCoNum > len(visuCo) {
				visuCoNum = len(visuCo)
			}
			pos += 2
			f := -1
			for c := 1; c < visuCoNum && pos+2 <= len(visuScene0); c++ {
				e := int(int16(binary.LittleEndian.Uint16(visuScene0[pos : pos+2])))
				pos += 2
				if e > f {
					f = e
					name := fmt.Sprintf("%s.%03d", u2eSceneName, e)
					visuCo[c].o = visuLoadObject(name)
					if visuCo[c].o == nil || visuCo[c].o.r == nil || visuCo[c].o.r0 == nil {
						log.Printf("u2e: failed to load object %s", name)
						return
					}
					*visuCo[c].o.r = visuRMatrix{}
					*visuCo[c].o.r0 = visuRMatrix{}
					visuCo[c].index = e
					visuCo[c].on = 0
				} else {
					g := 0
					for g < c {
						if visuCo[g].index == e {
							break
						}
						g++
					}
					visuCo[c] = visuCo[g]
					src := visuCo[g].o
					clone := visuGetNewObject()
					if src != nil {
						*clone = *src
					}
					clone.r = visuGetNewRMatrix()
					clone.r0 = visuGetNewRMatrix()
					*clone.r = visuRMatrix{}
					*clone.r0 = visuRMatrix{}
					visuCo[c].o = clone
					visuCo[c].on = 0
				}
			}
		}
	}

	visuCo[0].o = &visuCamObject
	visuCamObject.r = &visuCam
	visuCamObject.r0 = &visuCam

	listData := visuReadFile(fmt.Sprintf("%s.0AA", u2eSceneName))
	visuScl = 0
	if len(listData) >= 2 {
		for pos := 0; pos+2 <= len(listData); pos += 2 {
			a := int16(binary.LittleEndian.Uint16(listData[pos : pos+2]))
			if a == 0 || a == -1 {
				break
			}
			name := fmt.Sprintf("%s.0%c%c", u2eSceneName, byte(a/10)+'A', byte(a%10)+'A')
			data := visuReadFile(name)
			if len(data) == 0 {
				log.Printf("u2e: missing scene chunk %s", name)
				return
			}
			visuSceneList[visuScl].data = data
			visuScl++
			if visuScl >= len(visuSceneList) {
				break
			}
		}
	}
	if visuScl == 0 {
		log.Printf("u2e: no scene list data")
		return
	}

	visuResetScene()

	fpal := make([]byte, constants.PaletteByteCount)
	common.GetPalArea(fpal, 0, constants.PaletteColorCount)

	for b := 0; b < 33 && !driver.WantsToQuit(); b++ {
		for a := 3; a < constants.PaletteByteCount-6; a++ {
			v := int(fpal[a]) + 2
			if v > 63 {
				v = 63
			}
			fpal[a] = byte(v)
		}
		driver.Vsync(false)
		common.SetPalArea(fpal, 0, constants.PaletteColorCount)
		driver.Blit()
	}

	for b := 0; b < 16 && !driver.WantsToQuit(); b++ {
		driver.Vsync(true)
	}

	u2eFadeSet(shim.VRAM)

	for b := 0; b < 16 && !driver.WantsToQuit(); b++ {
		driver.Vsync(true)
	}

	for b := 0; b < 33 && !driver.WantsToQuit(); b++ {
		for a := 3; a < constants.PaletteByteCount-9; a++ {
			v := int(fpal[a]) - 2
			if v < 0 {
				v = 0
			}
			fpal[a] = byte(v)
		}
		for a := constants.PaletteByteCount - 9; a < constants.PaletteByteCount-3; a++ {
			v := int(fpal[a]) + 2
			if v > 63 {
				v = 63
			}
			fpal[a] = byte(v)
		}
		driver.Vsync(false)
		common.SetPalArea(fpal, 0, constants.PaletteColorCount)
		driver.Blit()
	}

	driver.ChangeMode(constants.ScreenWidth, constants.ScreenHeight, shim.DefaultJSSS)
	visuInit()

	if len(visuScenem) >= 16+constants.PaletteByteCount {
		cp := visuScenem[16 : 16+constants.PaletteByteCount]
		cp[255*3+0] = 0
		cp[255*3+1] = 0
		cp[255*3+2] = 0
		cp[252*3+0] = 0
		cp[252*3+1] = 0
		cp[252*3+2] = 0
		cp[253*3+0] = 63
		cp[253*3+1] = 63
		cp[253*3+2] = 63
		cp[254*3+0] = 63
		cp[254*3+1] = 63
		cp[254*3+2] = 63
		common.SetPalArea(cp, 0, constants.PaletteColorCount)
	}

	visuWindow(0, 319, 25, 174, 512, 9999999)

	shim.ClearScreen()

	for !driver.WantsToQuit() {
		if music.GetOrder() > 18 {
			break
		}
		driver.Vsync(false)
	}

	visuCopperCnt = 0
	visuSyncFrame = 0
	visuAvgRepeat = 1
	visuCl[0].ready = 0
	visuCl[1].ready = 0
	visuCl[2].ready = 0
	visuCl[3].ready = 1
	fov := 0

	for !driver.WantsToQuit() && visuXit == 0 {
		clear(shim.VRAM[:constants.DoubleScreenSize])
		if visuFirstFrame == 0 {
			visuDeadlock = 0
			visuClear255()
			visuCameraAngle(visuAngle(fov))
			visuOrderNum = 0

			for a := 1; a < visuCoNum; a++ {
				if visuCo[a].on == 0 {
					continue
				}
				if visuOrderNum < len(visuOrder) {
					visuOrder[visuOrderNum] = a
					visuOrderNum++
				}
				o := visuCo[a].o
				if o == nil || o.r == nil || o.r0 == nil {
					continue
				}
				*o.r = *o.r0
				visuCalcApplyRMatrix(o.r, &visuCam)
				if len(o.pl[0]) >= 2 {
					b := int(o.pl[0][1])
					if b >= 0 && b < len(o.v0) {
						if len(o.name) > 1 && o.name[1] == '_' {
							visuCo[a].dist = 1000000000
						} else {
							visuCo[a].dist = visuCalcSingleZ(b, o.v0, o.r)
						}
					}
				}
				if visuCurrFrame > 900*2 && visuCurrFrame < 1100*2 {
					if len(o.name) > 3 && o.name[1] == 's' && o.name[2] == '0' && o.name[3] == '1' {
						visuCo[a].dist = 1
					}
				}
			}

			for a := 0; a < visuOrderNum; a++ {
				c := visuOrder[a]
				dis := visuCo[c].dist
				b := a - 1
				for b >= 0 && dis > visuCo[visuOrder[b]].dist {
					visuOrder[b+1] = visuOrder[b]
					b--
				}
				visuOrder[b+1] = c
			}

			for a := 0; a < visuOrderNum; a++ {
				o := visuCo[visuOrder[a]].o
				visuDrawObject(o)
			}
		} else {
			visuSyncFrame = 0
			visuFirstFrame = 0
			visuCopperCnt = 1
		}

		a := visuSyncFrame - visuCurrFrame
		repeat := a + 1
		if repeat < 0 {
			repeat = 0
		}
		if repeat == 0 {
			visuCl[visuClw].frames = 1
		} else {
			visuCl[visuClw].frames = repeat
		}
		visuCl[visuClw].ready = 1
		visuClw = (visuClw + 1) & 3

		repeat = (repeat + 1) / 2
		visuCurrFrame += repeat * 2

		for repeat > 0 && visuXit == 0 {
			repeat--
			onum := 0
			for visuXit == 0 {
				if len(visuSp) == 0 {
					visuXit = 1
					break
				}
				a := int(visuSp[0])
				visuSp = visuSp[1:]
				if a == 0xFF {
					if len(visuSp) == 0 {
						visuXit = 1
						break
					}
					a = int(visuSp[0])
					visuSp = visuSp[1:]
					if a <= 0x7F {
						fov = a << 8
						break
					}
					if a == 0xFF {
						visuResetScene()
						visuXit = 1
						continue
					}
				}
				if (a & 0xC0) == 0xC0 {
					onum = (a & 0x3F) << 4
					if len(visuSp) == 0 {
						visuXit = 1
						break
					}
					a = int(visuSp[0])
					visuSp = visuSp[1:]
				}
				onum = (onum & 0xFF0) | (a & 0xF)
				if onum >= len(visuCo) {
					return
				}
				switch a & 0xC0 {
				case 0x80:
					visuCo[onum].on = 1
				case 0x40:
					visuCo[onum].on = 0
				}
				if onum >= visuCoNum {
					return
				}
				if visuCo[onum].o == nil || visuCo[onum].o.r0 == nil {
					continue
				}
				r := visuCo[onum].o.r0
				pflag := int32(0)
				switch a & 0x30 {
				case 0x10:
					if len(visuSp) == 0 {
						visuXit = 1
						break
					}
					pflag |= int32(visuSp[0])
					visuSp = visuSp[1:]
				case 0x20:
					if len(visuSp) < 2 {
						visuXit = 1
						break
					}
					pflag |= int32(visuSp[0]) | (int32(visuSp[1]) << 8)
					visuSp = visuSp[2:]
				case 0x30:
					if len(visuSp) < 3 {
						visuXit = 1
						break
					}
					pflag |= int32(visuSp[0]) | (int32(visuSp[1]) << 8) | (int32(visuSp[2]) << 16)
					visuSp = visuSp[3:]
				}
				l := visuLsGet(uint8(pflag))
				r.x += l
				l = visuLsGet(uint8(pflag >> 2))
				r.y += l
				l = visuLsGet(uint8(pflag >> 4))
				r.z += l

				if pflag&0x40 != 0 {
					for b := 0; b < 9; b++ {
						if pflag&(0x80<<b) != 0 {
							r.m[b] += visuLsGet(2)
						}
					}
				} else {
					for b := 0; b < 9; b++ {
						if pflag&(0x80<<b) != 0 {
							r.m[b] += visuLsGet(1)
						}
					}
				}
			}
		}

		driver.Blit()
		driver.Vsync(false)
		u2eCopper2()
	}

	common.GetPalArea(fpal, 0, constants.PaletteColorCount)
	for b := 0; b < 16 && !driver.WantsToQuit(); b++ {
		for a := 0; a < constants.PaletteByteCount; a++ {
			v := int(fpal[a]) + 4
			if v > 63 {
				v = 63
			}
			fpal[a] = byte(v)
		}
		driver.Vsync(false)
		shim.Outp(0x3c8, 255)
		for a := 0; a < constants.PaletteByteCount; a++ {
			shim.Outp(0x3c9, uint32(fpal[a]))
		}
		driver.Blit()
	}
}
