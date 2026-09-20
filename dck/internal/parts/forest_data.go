package parts

import (
	"bytes"
	"fmt"
	"sync"

	srdata "go-secondreality/dck"
)

var (
	forestOnce  sync.Once
	forestErr   error
	forestPosi1 []byte
	forestPosi2 []byte
	forestPosi3 []byte
	forestHBack []byte
	forestO2    []byte
)

func forestEnsureData() error {
	forestOnce.Do(func() {
		data := srdata.ForestData()
		var err error
		forestPosi1, err = forestExtractArray(data, "forest_posi1")
		if err != nil {
			forestErr = err
			return
		}
		forestPosi2, err = forestExtractArray(data, "forest_posi2")
		if err != nil {
			forestErr = err
			return
		}
		forestPosi3, err = forestExtractArray(data, "forest_posi3")
		if err != nil {
			forestErr = err
			return
		}
		forestHBack, err = forestExtractArray(data, "forest_hback")
		if err != nil {
			forestErr = err
			return
		}
		forestO2, err = forestExtractArray(data, "forest_o2")
		if err != nil {
			forestErr = err
			return
		}
	})
	return forestErr
}

func forestExtractArray(data []byte, name string) ([]byte, error) {
	idx := bytes.Index(data, []byte(name))
	if idx == -1 {
		return nil, fmt.Errorf("forest: array %q not found", name)
	}
	brace := bytes.IndexByte(data[idx:], '{')
	if brace == -1 {
		return nil, fmt.Errorf("forest: array %q missing '{'", name)
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
	return nil, fmt.Errorf("forest: array %q unterminated", name)
}
