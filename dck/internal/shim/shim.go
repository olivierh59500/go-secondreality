package shim

import "go-secondreality/dck/internal/constants"

const DefaultJSSS = 0.80

var VRAM = make([]byte, constants.VRAMX*constants.VRAMY)
var Palette [constants.PaletteColorCount][4]byte
var StartPixel int

var (
	paletteReadIndex     int
	paletteReadComponent int
	paletteIndex         int
	paletteComponent     int
	isFirstPart          = true
)

func ClearScreen() {
	clear(VRAM)
}

func SetPal(idx int, r, g, b byte) {
	if idx < 0 || idx >= constants.PaletteColorCount {
		return
	}
	Palette[idx][0] = b << 2
	Palette[idx][1] = g << 2
	Palette[idx][2] = r << 2
	Palette[idx][3] = 0
}

func Outp(reg int, value uint32) {
	switch reg {
	case 0x3c7:
		paletteReadIndex = int(value & 0xFF)
		paletteReadComponent = 0
	case 0x3c8:
		paletteIndex = int(value & 0xFF)
		paletteComponent = 0
	case 0x3c9:
		if paletteIndex >= constants.PaletteColorCount {
			paletteIndex = 0
		}
		component := 2 - paletteComponent
		Palette[paletteIndex][component] = byte(value&0xFF) << 2
		paletteComponent++
		if paletteComponent == 3 {
			paletteIndex++
			paletteComponent = 0
			if paletteIndex >= constants.PaletteColorCount {
				paletteIndex = 0
			}
		}
	}
}

func Inp(reg int) byte {
	switch reg {
	case 0x3c9:
		if paletteReadIndex >= constants.PaletteColorCount {
			paletteReadIndex = 0
		}
		component := 2 - paletteReadComponent
		result := Palette[paletteReadIndex][component] >> 2
		paletteReadComponent++
		if paletteReadComponent == 3 {
			paletteReadIndex++
			paletteReadComponent = 0
			if paletteReadIndex >= constants.PaletteColorCount {
				paletteReadIndex = 0
			}
		}
		return result
	default:
		return 0
	}
}

func SetStartPixel(v int) {
	StartPixel = v
}

func IsDemoFirstPart() bool {
	return isFirstPart
}

func FinishedDemoFirstPart() {
	isFirstPart = false
}
