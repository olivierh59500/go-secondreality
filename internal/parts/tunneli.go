package parts

import (
	"log"
	"math"

	"go-secondreality/internal/constants"
	"go-secondreality/internal/driver"
	"go-secondreality/internal/shim"
)

const tunneliFrameCountToExit = 1060

type tunneliBC struct {
	x int16
	y int16
}

type tunneliRing struct {
	x int32
	y int32
	c uint8
}

var (
	tunneliPutki  [103]tunneliRing
	tunneliPcalc  [138][64]tunneliBC
	tunneliRows   [501]uint16
	tunneliSinit  [4097]int16
	tunneliCosit  [2049]int16
	tunneliSade   [103]uint16
	tunneliOldPos [7501]uint16

	tunneliX int32
	tunneliY int32
	tunneliZ int32

	tunneliSX uint16
	tunneliSY uint16
	tunneliFrame int
	tunneliQuit  bool
)

func runTunneli() {
	if err := tunneliEnsureData(); err != nil {
		log.Printf("tunneli: %v", err)
		return
	}
	tunneliInit()
	tunneliLoop()
}

func tunneliInit() {
	tunneliX = 0
	tunneliY = 0
	tunneliZ = 0
	tunneliSX = 0
	tunneliSY = 0
	tunneliFrame = 0
	tunneliQuit = false

	clear(tunneliPutki[:])
	clear(tunneliPcalc[:])
	clear(tunneliSinit[:])
	clear(tunneliCosit[:])
	clear(tunneliSade[:])

	screenSize := uint16(constants.ScreenSize)
	for i := range tunneliRows {
		if i < 100 || i >= 300 {
			tunneliRows[i] = screenSize
		} else {
			tunneliRows[i] = uint16((i - 100) * constants.ScreenWidth)
		}
	}

	if len(tunneliSiniData) >= len(tunneliSinit) {
		copy(tunneliSinit[:], tunneliSiniData[:len(tunneliSinit)])
	} else {
		copy(tunneliSinit[:], tunneliSiniData)
	}

	cosStart := len(tunneliSinit)
	if len(tunneliSiniData) > cosStart {
		cosSlice := tunneliSiniData[cosStart:]
		if len(cosSlice) > len(tunneliCosit) {
			cosSlice = cosSlice[:len(tunneliCosit)]
		}
		copy(tunneliCosit[:], cosSlice)
	}

	idx := 0
	for i := range tunneliPcalc {
		for j := range tunneliPcalc[i] {
			if idx+1 >= len(tunneliTunData) {
				break
			}
			tunneliPcalc[i][j].x = tunneliTunData[idx]
			idx++
			tunneliPcalc[i][j].y = tunneliTunData[idx]
			idx++
		}
	}

	for i := 0; i < constants.PaletteColorCount; i++ {
		shim.SetPal(i, 0, 0, 0)
	}
	for i := 0; i <= 64; i++ {
		v := byte(64 - i)
		shim.SetPal(64+i, v, v, v)
	}
	for i := 0; i <= 64; i++ {
		v := byte((64 - i) * 3 / 4)
		shim.SetPal(128+i, v, v, v)
	}
	shim.SetPal(68, 0, 0, 0)
	shim.SetPal(132, 0, 0, 0)
	shim.SetPal(255, 0, 63, 0)

	for i := 0; i <= 100 && i < len(tunneliPutki); i++ {
		tunneliPutki[i].x = 0
		tunneliPutki[i].y = 0
		tunneliPutki[i].c = 0
	}

	for z := 0; z <= 100; z++ {
		tunneliSade[z] = uint16(16384 / ((z * 7) + 95))
	}

	for i := range tunneliOldPos {
		tunneliOldPos[i] = screenSize
	}
}

func tunneliLoop() {
	screenSize := uint16(constants.ScreenSize)

	for !tunneliQuit && !driver.WantsToQuit() {
		frames := driver.Vsync(false)

		ry := 0
		for xi := 80; xi >= 4; xi-- {
			bx := int16(tunneliPutki[xi].x - tunneliPutki[5].x)
			by := int16(tunneliPutki[xi].y - tunneliPutki[5].y)
			br := int(tunneliSade[xi])
			if br < 0 || br >= len(tunneliPcalc) {
				continue
			}
			bbc := uint8(int(tunneliPutki[xi].c) + int(math.Round(float64(xi)/1.3)))
			if bbc < 64 {
				continue
			}
			pcp := tunneliPcalc[br]
			ax := ry
			for i := 0; i < 64; i++ {
				idx := ax
				if idx >= len(tunneliOldPos)-1 {
					idx = len(tunneliOldPos) - 1
				}

				if prev := tunneliOldPos[idx]; prev < screenSize {
					shim.VRAM[prev] = 0
				}

				newpos := constants.ScreenSize
				di := int(pcp[i].x) + int(bx)
				if uint32(di) <= uint32(constants.ScreenWidth-1) {
					yv := int(pcp[i].y) + int(by)
					row := 100 + yv
					if uint32(row) <= 500 {
						base := tunneliRows[row]
						if base < screenSize {
							newpos = int(base) + di
							if newpos >= 0 && newpos < len(shim.VRAM) {
								shim.VRAM[newpos] = bbc
							}
						}
					}
				}

				tunneliOldPos[idx] = uint16(newpos)
				ax++
			}
			ry = ax
		}

		for sync := 0; sync < frames; sync++ {
			sx := tunneliSX
			sy := tunneliSY
			tunneliPutki[100].x = int32(tunneliCosit[int(sy&2047)]) -
				int32(tunneliSinit[int((uint32(sy)*3)&4095)]) -
				int32(tunneliCosit[int(sx&2047)])
			tunneliPutki[100].y = int32(tunneliSinit[int((uint32(sx)*2)&4095)]) -
				int32(tunneliCosit[int(sx&2047)]) +
				int32(tunneliSinit[int(uint32(tunneliY)&4095)])

			copy(tunneliPutki[0:100], tunneliPutki[1:101])

			tunneliSY++
			tunneliSX++

			if (tunneliSY & 15) > 7 {
				tunneliPutki[99].c = 128
			} else {
				tunneliPutki[99].c = 64
			}

			if tunneliFrame >= (tunneliFrameCountToExit - 102) {
				tunneliPutki[99].c = 0
			}

			tunneliFrame++
			if tunneliFrame == tunneliFrameCountToExit || driver.WantsToQuit() {
				tunneliQuit = true
				break
			}
		}

		driver.Blit()
	}
}
