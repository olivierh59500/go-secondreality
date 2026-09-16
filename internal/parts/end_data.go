package parts

import (
	"bytes"
	"fmt"
	"sync"

	srdata "go-secondreality"
)

var (
	endOnce sync.Once
	endErr  error
	endPic  []byte
)

func endEnsureData() error {
	endOnce.Do(func() {
		var err error
		endPic, err = endExtractArray(srdata.EndData(), "end_pic")
		if err != nil {
			endErr = err
			return
		}
	})
	return endErr
}

func endExtractArray(data []byte, name string) ([]byte, error) {
	idx := bytes.Index(data, []byte(name))
	if idx == -1 {
		return nil, fmt.Errorf("end: array %q not found", name)
	}
	brace := bytes.IndexByte(data[idx:], '{')
	if brace == -1 {
		return nil, fmt.Errorf("end: array %q missing '{'", name)
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
	return nil, fmt.Errorf("end: array %q unterminated", name)
}
