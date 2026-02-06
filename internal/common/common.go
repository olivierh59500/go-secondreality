package common

import (
	"encoding/binary"

	"go-secondreality/internal/constants"
	"go-secondreality/internal/shim"
)

type Palette [constants.PaletteByteCount]byte

type readpHeader struct {
	Magic int16
	Wid   int16
	Hig   int16
	Cols  int16
	Add   int16
}

var (
	FrameCount int
	CopPal     []byte
	DoPal      int
	CopStart   int
	CopScrl    int
	CopDoFade  int
	CopDrop    int

	CopFadePal   []int16
	FadePal      Palette
	FadePalShort [constants.PaletteByteCount]int16
)

func Reset() {
	FrameCount = 0
	CopPal = nil
	DoPal = 0
	CopStart = 0
	CopScrl = 0
	CopDoFade = 0
	CopDrop = 0
	CopFadePal = nil

	clear(FadePalShort[:])
	clear(FadePal[:])
}

func Readp(dest []byte, row int, src []byte) {
	if len(src) < 16 {
		return
	}

	hdr := readpHeader{
		Magic: int16(binary.LittleEndian.Uint16(src[0:2])),
		Wid:   int16(binary.LittleEndian.Uint16(src[2:4])),
		Hig:   int16(binary.LittleEndian.Uint16(src[4:6])),
		Cols:  int16(binary.LittleEndian.Uint16(src[6:8])),
		Add:   int16(binary.LittleEndian.Uint16(src[8:10])),
	}

	if row == -1 {
		count := int(hdr.Cols) * 3
		if count <= 0 {
			return
		}
		start := 16
		end := start + count
		if start >= len(src) {
			return
		}
		if end > len(src) {
			end = len(src)
		}
		copy(dest, src[start:end])
		return
	}

	if row >= int(hdr.Hig) {
		return
	}

	pos := int(hdr.Add) * 16
	if pos < 0 || pos >= len(src) {
		return
	}

	for i := 0; i < row; i++ {
		if pos+2 > len(src) {
			return
		}
		skip := int(binary.LittleEndian.Uint16(src[pos : pos+2]))
		pos += 2 + skip
		if pos > len(src) {
			return
		}
	}

	if pos+2 > len(src) {
		return
	}
	bytes := int(binary.LittleEndian.Uint16(src[pos : pos+2]))
	pos += 2
	end := pos + bytes
	if end > len(src) {
		end = len(src)
	}

	di := 0
	for pos < end && di < len(dest) {
		code := src[pos]
		pos++
		if code&0x80 != 0 {
			count := int(code & 0x7F)
			if pos >= end {
				break
			}
			val := src[pos]
			pos++
			if count <= 0 {
				continue
			}
			if di+count > len(dest) {
				count = len(dest) - di
			}
			for i := 0; i < count; i++ {
				dest[di+i] = val
			}
			di += count
		} else {
			dest[di] = code
			di++
		}
	}
}

func SetPalArea(p []byte, offset, count int) {
	if count <= 0 || len(p) == 0 {
		return
	}

	shim.Outp(0x3c8, uint32(offset))

	limit := count * 3
	if limit > len(p) {
		limit = len(p)
	}

	for c := 0; c < limit; c++ {
		shim.Outp(0x3c9, uint32(p[c]))
	}
}

func GetPalArea(p []byte, offset, count int) {
	if count <= 0 || len(p) == 0 {
		return
	}

	shim.Outp(0x3c8, uint32(offset))

	limit := count * 3
	if limit > len(p) {
		limit = len(p)
	}

	for c := 0; c < limit; c++ {
		p[c] = shim.Inp(0x3c9)
	}
}

func Copper2() {
	FrameCount++
	if DoPal != 0 {
		for i := 0; i < constants.PaletteColorCount; i++ {
			SetPalArea(CopPal, 0, constants.PaletteColorCount)
		}
		DoPal = 0
	}
}

func Copper3() {
	if CopDoFade != 0 {
		DoPal = 1
	}

	if DoPal != 0 {
		CopPal = FadePal[:]
		if len(CopFadePal) == constants.PaletteByteCount {
			for i := 0; i < constants.PaletteByteCount; i++ {
				FadePalShort[i] += CopFadePal[i]
				FadePal[i] = byte(FadePalShort[i] >> 8)
			}
		}
		CopDoFade--
	}
}
