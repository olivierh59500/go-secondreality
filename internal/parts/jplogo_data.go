package parts

import (
	"bytes"
	"fmt"
	"sync"

	srdata "go-secondreality"
)

var (
	jpOnce sync.Once
	jpErr  error
	jpPic  []byte
)

func jpEnsureData() error {
	jpOnce.Do(func() {
		var err error
		jpPic, err = jpExtractArray(srdata.JPLogoData(), "jl_pic")
		if err != nil {
			jpErr = err
			return
		}
	})
	return jpErr
}

func jpExtractArray(data []byte, name string) ([]byte, error) {
	idx := bytes.Index(data, []byte(name))
	if idx == -1 {
		return nil, fmt.Errorf("jplogo: array %q not found", name)
	}
	brace := bytes.IndexByte(data[idx:], '{')
	if brace == -1 {
		return nil, fmt.Errorf("jplogo: array %q missing '{'", name)
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
	return nil, fmt.Errorf("jplogo: array %q unterminated", name)
}
