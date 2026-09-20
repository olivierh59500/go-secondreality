package parts

import (
	"log"

	"go-secondreality/dck/internal/common"
	"go-secondreality/dck/internal/constants"
	"go-secondreality/dck/internal/driver"
	"go-secondreality/dck/internal/music"
	"go-secondreality/dck/internal/shim"
)

func runEnd() {
	if err := endEnsureData(); err != nil {
		log.Printf("end: %v", err)
		return
	}

	pal2 := make([]byte, constants.PaletteByteCount)
	palette := make([]byte, constants.PaletteByteCount)

	shim.ClearScreen()

	driver.Vsync(false)

	shim.Outp(0x3c8, 0)
	for a := 0; a < constants.PaletteByteCount-3; a++ {
		shim.Outp(0x3c9, 63)
	}

	driver.Vsync(false)

	shim.Outp(0x3c8, 0)
	for a := 0; a < constants.PaletteByteCount-3; a++ {
		shim.Outp(0x3c9, 63)
	}

	for a := 0; a < 32; a++ {
		driver.Vsync(true)
	}

	common.Readp(palette, -1, endPic)

	for y := 0; y < constants.DoubleScreenHeight; y++ {
		common.Readp(shim.VRAM[y*constants.ScreenWidth:], y, endPic)
	}

	for c := 0; c <= 128; c++ {
		for a := 0; a < constants.PaletteByteCount-3; a++ {
			pal2[a] = byte(((128-c)*63 + int(palette[a])*c) / 128)
		}

		driver.Vsync(false)
		common.SetPalArea(pal2, 0, constants.PaletteColorCount-1)
		driver.Blit()
	}

	for a := 0; a < 5000 && !driver.WantsToQuit(); a++ {
		driver.Vsync(false)
		driver.Blit()
		if music.GetPlusFlags() > -16 {
			break
		}
	}

	for c := 63; c >= 0; c-- {
		for a := 0; a < constants.PaletteByteCount-3; a++ {
			pal2[a] = byte((int(palette[a]) * c) / 64)
		}
		driver.Vsync(false)
		common.SetPalArea(pal2, 0, constants.PaletteColorCount-1)
		driver.Blit()
	}
}
