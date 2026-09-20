package parts

import (
	"bytes"
	"fmt"
	"log"
	"sync"

	srdata "go-secondreality/dck"
	"go-secondreality/dck/internal/common"
	"go-secondreality/dck/internal/constants"
	"go-secondreality/dck/internal/driver"
	"go-secondreality/dck/internal/music"
	"go-secondreality/dck/internal/shim"
)

var (
	outtaOnce sync.Once
	outtaErr  error

	outtaBasePalette []byte
	outtaMemBlock    []byte
	outtaWfade       []int
)

func runOutta() {
	if err := outtaEnsureData(); err != nil {
		log.Printf("outta: %v", err)
		return
	}
	if len(outtaBasePalette) < constants.PaletteByteCount {
		log.Printf("outta: base palette too small (%d bytes)", len(outtaBasePalette))
		return
	}
	if len(outtaMemBlock) == 0 {
		log.Printf("outta: missing pam_memblock")
		return
	}
	if len(outtaWfade) == 0 {
		log.Printf("outta: missing wfade")
		return
	}
	pamPal := make([]byte, len(outtaBasePalette)+constants.PaletteByteCount*64)
	copy(pamPal, outtaBasePalette)

	for a := 1; a < 64; a++ {
		for b := 0; b < constants.PaletteByteCount; b++ {
			base := int(pamPal[b])
			pamPal[a*constants.PaletteByteCount+b] = byte((63*a + (64-a)*base) / 64)
		}
	}

	if !shim.IsDemoFirstPart() {
		for music.Sync() < 10 && !driver.WantsToQuit() {
			driver.Vsync(false)
		}
		if driver.WantsToQuit() {
			return
		}
	}

	shim.ClearScreen()

	frameIdx := 0
	for f := 0; f < 45 && !driver.WantsToQuit(); f++ {
		if f <= 40 {
			frameIdx = outtaBlitFrame(outtaMemBlock, frameIdx)
		}

		wf := 0
		if f < len(outtaWfade) {
			wf = outtaWfade[f]
		}
		if wf < 0 {
			wf = 0
		}
		if wf > 63 {
			wf = 63
		}

		base := wf * constants.PaletteByteCount
		if base+constants.PaletteByteCount <= len(pamPal) {
			common.CopPal = pamPal[base : base+constants.PaletteByteCount]
		} else {
			common.CopPal = pamPal[:constants.PaletteByteCount]
		}

		common.DoPal = 1
		common.Copper2()
		common.Copper3()

		for i := 0; i < 4; i++ {
			driver.Vsync(true)
		}
		driver.Blit()
	}
}

func outtaBlitFrame(mem []byte, start int) int {
	if start < 0 {
		start = 0
	}
	if start >= len(mem) {
		return start
	}
	idx := start
	dest := 0
	for {
		if idx >= len(mem) {
			break
		}
		val := int8(mem[idx])
		idx++
		if val == 0 {
			break
		}
		if val < 0 {
			for {
				dest += int(-val)
				if idx >= len(mem) {
					return idx
				}
				val = int8(mem[idx])
				idx++
				if val == 0 {
					return outtaAlign16(start, idx)
				}
				if val >= 0 {
					break
				}
			}
		}
		if idx >= len(mem) {
			break
		}
		count := int(val)
		repeatValue := mem[idx]
		idx++
		if count > 0 {
			if dest < 0 {
				dest = 0
			}
			if dest >= len(shim.VRAM) {
				return outtaAlign16(start, idx)
			}
			end := dest + count
			if end > len(shim.VRAM) {
				end = len(shim.VRAM)
			}
			for i := dest; i < end; i++ {
				shim.VRAM[i] = repeatValue
			}
			dest = end
		}
	}
	return outtaAlign16(start, idx)
}

func outtaAlign16(start, idx int) int {
	delta := idx - start
	if mod := delta & 15; mod != 0 {
		delta += 16 - mod
	}
	next := start + delta
	return next
}

func outtaEnsureData() error {
	outtaOnce.Do(func() {
		source := srdata.OuttaSource()
		data := srdata.OuttaData()
		var err error
		outtaBasePalette, err = outtaExtractArray(source, "basePalette")
		if err != nil {
			outtaErr = err
			return
		}
		outtaMemBlock, err = outtaExtractArray(data, "pam_memblock")
		if err != nil {
			outtaErr = err
			return
		}
		outtaWfade, err = outtaExtractArrayInts(data, "wfade")
		if err != nil {
			outtaErr = err
			return
		}
	})
	return outtaErr
}

func outtaExtractArray(data []byte, name string) ([]byte, error) {
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
			val, n := outtaParseNumber(data[i:])
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

func outtaExtractArrayInts(data []byte, name string) ([]int, error) {
	idx := bytes.Index(data, []byte(name))
	if idx == -1 {
		return nil, fmt.Errorf("array %q not found", name)
	}
	brace := bytes.IndexByte(data[idx:], '{')
	if brace == -1 {
		return nil, fmt.Errorf("array %q missing '{'", name)
	}
	i := idx + brace + 1
	out := make([]int, 0, 128)
	for i < len(data) {
		c := data[i]
		if c == '}' {
			return out, nil
		}
		if c == '-' || (c >= '0' && c <= '9') {
			val, n := outtaParseNumber(data[i:])
			if n > 0 {
				out = append(out, val)
				i += n
				continue
			}
		}
		i++
	}
	return nil, fmt.Errorf("array %q unterminated", name)
}

func outtaParseNumber(data []byte) (int, int) {
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
