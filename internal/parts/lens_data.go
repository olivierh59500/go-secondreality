package parts

import (
	"bytes"
	"fmt"
	"sync"

	srdata "go-secondreality"
)

var (
	lensOnce sync.Once
	lensErr  error
	lensExb  []byte
	lensEx0  []byte
	lensEx1  []byte
	lensEx2  []byte
	lensEx3  []byte
	lensEx4  []byte
	lensExp  []byte
)

func lensEnsureData() error {
	lensOnce.Do(func() {
		data := srdata.LensData()
		var err error
		lensExb, err = lensExtractArray(data, "lensexbBase")
		if err != nil {
			lensErr = err
			return
		}
		lensEx0, err = lensExtractArray(data, "lensex0")
		if err != nil {
			lensErr = err
			return
		}
		lensEx1, err = lensExtractArray(data, "lensex1")
		if err != nil {
			lensErr = err
			return
		}
		lensEx2, err = lensExtractArray(data, "lensex2")
		if err != nil {
			lensErr = err
			return
		}
		lensEx3, err = lensExtractArray(data, "lensex3")
		if err != nil {
			lensErr = err
			return
		}
		lensEx4, err = lensExtractArray(data, "lensex4")
		if err != nil {
			lensErr = err
			return
		}
		lensExp, err = lensExtractArray(data, "lensexp")
		if err != nil {
			lensErr = err
			return
		}
	})
	return lensErr
}

func lensExtractArray(data []byte, name string) ([]byte, error) {
	idx := bytes.Index(data, []byte(name))
	if idx == -1 {
		return nil, fmt.Errorf("lens: array %q not found", name)
	}
	brace := bytes.IndexByte(data[idx:], '{')
	if brace == -1 {
		return nil, fmt.Errorf("lens: array %q missing '{'", name)
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
	return nil, fmt.Errorf("lens: array %q unterminated", name)
}
