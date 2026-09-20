package parts

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sync"

	srdata "go-secondreality/dck"
)

var (
	tunneliDataOnce sync.Once
	tunneliDataErr  error

	tunneliSiniData []int16
	tunneliTunData  []int16
)

func tunneliEnsureData() error {
	tunneliDataOnce.Do(func() {
		data := srdata.TunneliData()
		var err error
		var raw []byte
		raw, err = tunneliExtractArray(data, "_tunnel_sini")
		if err != nil {
			tunneliDataErr = err
			return
		}
		tunneliSiniData = tunneliBytesToInt16(raw)

		raw, err = tunneliExtractArray(data, "_tunnel_tun")
		if err != nil {
			tunneliDataErr = err
			return
		}
		tunneliTunData = tunneliBytesToInt16(raw)
	})
	return tunneliDataErr
}

func tunneliExtractArray(data []byte, name string) ([]byte, error) {
	idx := bytes.Index(data, []byte(name))
	if idx == -1 {
		return nil, fmt.Errorf("tunneli: array %q not found", name)
	}
	brace := bytes.IndexByte(data[idx:], '{')
	if brace == -1 {
		return nil, fmt.Errorf("tunneli: array %q missing '{'", name)
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
	return nil, fmt.Errorf("tunneli: array %q unterminated", name)
}

func tunneliBytesToInt16(b []byte) []int16 {
	n := len(b) / 2
	out := make([]int16, n)
	for i := 0; i < n; i++ {
		out[i] = int16(binary.LittleEndian.Uint16(b[i*2:]))
	}
	return out
}
