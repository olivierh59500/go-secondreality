package parts

import (
	"bytes"
	"fmt"
	"log"
	"sync"

	srdata "go-secondreality"
	"go-secondreality/internal/common"
	"go-secondreality/internal/constants"
	"go-secondreality/internal/driver"
	"go-secondreality/internal/shim"
)

var (
	begOnce sync.Once
	begErr  error
	begData []byte
)

func runBeg() {
	if err := begEnsureData(); err != nil {
		log.Printf("beg: %v", err)
		return
	}

	shim.ClearScreen()

	for i := 0; i < 32; i++ {
		driver.Vsync(true)
	}

	shim.Outp(0x3c8, 0)
	for i := 0; i < 255; i++ {
		shim.Outp(0x3c9, 63)
		shim.Outp(0x3c9, 63)
		shim.Outp(0x3c9, 63)
	}
	shim.Outp(0x3c9, 0)
	shim.Outp(0x3c9, 0)
	shim.Outp(0x3c9, 0)

	palette := make([]byte, constants.PaletteByteCount)
	common.Readp(palette, -1, begData)

	for y := 0; y < constants.DoubleScreenHeight; y++ {
		row := y * constants.ScreenWidth
		common.Readp(shim.VRAM[row:row+constants.ScreenWidth], y, begData)
	}

	pal2 := make([]byte, constants.PaletteByteCount)
	for c := 0; c <= 128; c++ {
		for i := 0; i < constants.PaletteByteCount-3; i++ {
			pal2[i] = byte(((128-c)*63 + int(palette[i])*c) / 128)
		}
		driver.Vsync(false)
		common.SetPalArea(pal2, 0, 254)
		driver.Blit()
	}

	common.SetPalArea(palette, 0, 254)
}

func begEnsureData() error {
	begOnce.Do(func() {
		var err error
		begData, err = begExtractArray(srdata.BegData, "TitleScreenData")
		if err != nil {
			begErr = err
			return
		}
	})
	return begErr
}

func begExtractArray(data []byte, name string) ([]byte, error) {
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
			val, n := begParseNumber(data[i:])
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

func begParseNumber(data []byte) (int, int) {
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
