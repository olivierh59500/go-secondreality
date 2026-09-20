package parts

import (
	"bytes"
	"fmt"
	"sync"

	srdata "go-secondreality/dck"
	"go-secondreality/dck/internal/constants"
)

var (
	visuDataOnce sync.Once
	visuDataErr  error
	visuSinTable []int16
	visuAvistan  []uint16
	visuAFillDiv []int16
	visuRowTable []uint16
)

func visuEnsureData() error {
	visuDataOnce.Do(func() {
		data := srdata.VisuData()
		var err error
		sinVals, err := visuExtractArrayInts(data, "sintable")
		if err != nil {
			visuDataErr = err
			return
		}
		avVals, err := visuExtractArrayInts(data, "avistan")
		if err != nil {
			visuDataErr = err
			return
		}
		afVals, err := visuExtractArrayInts(data, "afilldiv")
		if err != nil {
			visuDataErr = err
			return
		}
		visuSinTable = make([]int16, len(sinVals))
		for i, v := range sinVals {
			visuSinTable[i] = int16(v)
		}
		visuAvistan = make([]uint16, len(avVals))
		for i, v := range avVals {
			visuAvistan[i] = uint16(v)
		}
		visuAFillDiv = make([]int16, len(afVals))
		for i, v := range afVals {
			visuAFillDiv[i] = int16(v)
		}
		visuRowTable = make([]uint16, constants.ScreenHeight+1)
		for i := 0; i < constants.ScreenHeight; i++ {
			visuRowTable[i] = uint16(i * constants.ScreenWidth)
		}
	})
	return visuDataErr
}

func visuExtractArrayInts(data []byte, name string) ([]int, error) {
	idx := bytes.Index(data, []byte(name))
	if idx == -1 {
		return nil, fmt.Errorf("visu: array %q not found", name)
	}
	brace := bytes.IndexByte(data[idx:], '{')
	if brace == -1 {
		return nil, fmt.Errorf("visu: array %q missing '{'", name)
	}
	i := idx + brace + 1
	out := make([]int, 0, 1024)
	for i < len(data) {
		c := data[i]
		if c == '}' {
			return out, nil
		}
		if c == '-' || (c >= '0' && c <= '9') {
			val, n := visuParseNumber(data[i:])
			if n > 0 {
				out = append(out, val)
				i += n
				continue
			}
		}
		i++
	}
	return nil, fmt.Errorf("visu: array %q unterminated", name)
}

func visuParseNumber(data []byte) (int, int) {
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
