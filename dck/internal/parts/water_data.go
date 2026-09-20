package parts

import (
	"bytes"
	"fmt"
	"sync"

	srdata "go-secondreality/dck"
)

var (
	waterOnce   sync.Once
	waterErr    error
	waterWat1   []byte
	waterWat2   []byte
	waterWat3   []byte
	waterMiekka []byte
	waterTausta []byte
)

func waterEnsureData() error {
	waterOnce.Do(func() {
		data := srdata.WaterData()
		var err error
		waterWat1, err = waterExtractArray(data, "water_wat1")
		if err != nil {
			waterErr = err
			return
		}
		waterWat2, err = waterExtractArray(data, "water_wat2")
		if err != nil {
			waterErr = err
			return
		}
		waterWat3, err = waterExtractArray(data, "water_wat3")
		if err != nil {
			waterErr = err
			return
		}
		waterMiekka, err = waterExtractArray(data, "water_miekka")
		if err != nil {
			waterErr = err
			return
		}
		waterTausta, err = waterExtractArray(data, "water_tausta")
		if err != nil {
			waterErr = err
			return
		}
	})
	return waterErr
}

func waterExtractArray(data []byte, name string) ([]byte, error) {
	idx := bytes.Index(data, []byte(name))
	if idx == -1 {
		return nil, fmt.Errorf("water: array %q not found", name)
	}
	brace := bytes.IndexByte(data[idx:], '{')
	if brace == -1 {
		return nil, fmt.Errorf("water: array %q missing '{'", name)
	}
	i := idx + brace + 1
	out := make([]byte, 0, 1024)
	for i < len(data) {
		c := data[i]
		if c == '}' {
			return out, nil
		}
		if c == '-' || (c >= '0' && c <= '9') {
			val, n := glenzParseNumber(data[i:])
			if n > 0 {
				out = append(out, byte(uint8(val)))
				i += n
				continue
			}
		}
		i++
	}
	return nil, fmt.Errorf("water: array %q unterminated", name)
}
