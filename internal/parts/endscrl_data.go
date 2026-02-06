package parts

import (
	"bytes"
	"fmt"
	"sync"

	srdata "go-secondreality"
)

const (
	endscrlFontRows   = 25
	endscrlFontStride = 1550
)

var (
	endscrlOnce sync.Once
	endscrlErr  error
	endscrlFont []byte
)

func endscrlEnsureData() error {
	endscrlOnce.Do(func() {
		var err error
		endscrlFont, err = endscrlExtractArray(srdata.EndScrlData, "endscrl_font")
		if err != nil {
			endscrlErr = err
			return
		}
		need := endscrlFontRows * endscrlFontStride
		if len(endscrlFont) < need {
			pad := make([]byte, need)
			copy(pad, endscrlFont)
			endscrlFont = pad
		}
	})
	return endscrlErr
}

func endscrlExtractArray(data []byte, name string) ([]byte, error) {
	idx := bytes.Index(data, []byte(name))
	if idx == -1 {
		return nil, fmt.Errorf("endscrl: array %q not found", name)
	}
	brace := bytes.IndexByte(data[idx:], '{')
	if brace == -1 {
		return nil, fmt.Errorf("endscrl: array %q missing '{'", name)
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
	return nil, fmt.Errorf("endscrl: array %q unterminated", name)
}
