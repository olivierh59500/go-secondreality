package parts

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"sync"

	srdata "go-secondreality"
	"go-secondreality/internal/constants"
	"go-secondreality/internal/driver"
	"go-secondreality/internal/music"
	"go-secondreality/internal/shim"
)

var (
	u2aDataOnce sync.Once
	u2aDataErr  error
	u2aBg       []byte
)

const u2aSceneName = "U2A"

func runU2A() {
	if err := u2aEnsureData(); err != nil {
		log.Printf("u2a: %v", err)
		return
	}

	if len(u2aBg) < 16+constants.PaletteByteCount+constants.ScreenSize {
		log.Printf("u2a: data too small (%d bytes)", len(u2aBg))
		return
	}

	visuReset()

	bg2 := make([]byte, constants.ScreenSize)
	copy(bg2, u2aBg[16+constants.PaletteByteCount:16+constants.PaletteByteCount+constants.ScreenSize])

	shim.ClearScreen()

	scene0 := visuReadFile(fmt.Sprintf("%s.00M", u2aSceneName))
	if len(scene0) == 0 {
		log.Printf("u2a: missing scene data")
		return
	}
	visuScene0 = scene0
	visuScenem = scene0

	if len(visuScene0) >= 16+192*3+64*3 && len(u2aBg) >= 16+64*3 {
		copy(visuScene0[16+192*3:16+192*3+64*3], u2aBg[16:16+64*3])
	}

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
					name := fmt.Sprintf("%s.%03d", u2aSceneName, e)
					visuCo[c].o = visuLoadObject(name)
					if visuCo[c].o == nil || visuCo[c].o.r == nil || visuCo[c].o.r0 == nil {
						log.Printf("u2a: failed to load object %s", name)
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

	listData := visuReadFile(fmt.Sprintf("%s.0AA", u2aSceneName))
	visuScl = 0
	if len(listData) >= 2 {
		for pos := 0; pos+2 <= len(listData); pos += 4 {
			a := int16(binary.LittleEndian.Uint16(listData[pos : pos+2]))
			if a == 0 || a == -1 {
				break
			}
			name := fmt.Sprintf("%s.0%c%c", u2aSceneName, byte(a/10)+'A', byte(a%10)+'A')
			data := visuReadFile(name)
			if len(data) == 0 {
				log.Printf("u2a: missing scene chunk %s", name)
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
		log.Printf("u2a: no scene list data")
		return
	}

	visuResetScene()

	if !shim.IsDemoFirstPart() {
		for !driver.WantsToQuit() {
			if music.GetOrder() > 10 && music.GetRow() > 46 {
				break
			}
			driver.Vsync(false)
		}
		if driver.WantsToQuit() {
			return
		}
	}

	visuInit()

	if len(visuScenem) >= 16+constants.PaletteByteCount {
		cp := visuScenem[16 : 16+constants.PaletteByteCount]
		shim.Outp(0x3c8, 0)
		for i := 0; i < len(cp); i++ {
			shim.Outp(0x3c9, uint32(cp[i]))
		}
	}

	visuWindow(0, 319, 25, 174, 512, 9999999)

	shim.ClearScreen()

	visuXit = 0
	visuCurrFrame = 0
	visuCopperCnt = 0
	visuSyncFrame = 0
	visuAvgRepeat = 1
	visuCl[0].ready = 0
	visuCl[1].ready = 0
	visuCl[2].ready = 0
	visuCl[3].ready = 1
	fov := 0

	for !driver.WantsToQuit() && visuXit == 0 {
		visuDeadlock = 0
		visuClearBG(bg2)
		visuCameraAngle(visuAngle(fov))
		visuOrderNum = 0

		for a := 1; a < visuCoNum; a++ {
			if visuCo[a].on != 0 {
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
						visuCo[a].dist = visuCalcSingleZ(b, o.v0, o.r)
					}
				}
			}
		}

		if visuCity == 1 {
			if 2 < visuCoNum {
				visuCo[2].dist = 1000000000
			}
			if 7 < visuCoNum {
				visuCo[7].dist = 1000000000
			}
			if 13 < visuCoNum {
				visuCo[13].dist = 1000000000
			}
		}
		if visuCity == 2 {
			if 14 < visuCoNum {
				visuCo[14].dist = 1000000000
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

		visuAvgRepeat = (visuAvgRepeat + (visuSyncFrame - visuCurrFrame) + 1) / 2
		repeat := visuAvgRepeat
		if repeat < 1 {
			repeat = 1
		}
		visuCl[visuClw].frames = repeat
		visuCl[visuClw].ready = 1
		visuClw = (visuClw + 1) & 3
		visuCurrFrame += repeat

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
		u2aCopper2()
	}

	visuClearBG(bg2)
}

func u2aEnsureData() error {
	u2aDataOnce.Do(func() {
		u2aBg, u2aDataErr = u2aExtractArray(srdata.U2AData, "u2a_bg")
	})
	return u2aDataErr
}

func u2aCopper2() {
	visuSyncFrame++

	if visuCl[0].ready == 2 {
		visuCl[0].ready = 0
	}
	if visuCl[1].ready == 2 {
		visuCl[1].ready = 0
	}
	if visuCl[2].ready == 2 {
		visuCl[2].ready = 0
	}
	if visuCl[3].ready == 2 {
		visuCl[3].ready = 0
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
		visuClr++
		visuClr &= 3
	} else {
		visuAvgRepeat++
	}
}

func u2aExtractArray(data []byte, name string) ([]byte, error) {
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
			val, n := u2aParseNumber(data[i:])
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

func u2aParseNumber(data []byte) (int, int) {
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
