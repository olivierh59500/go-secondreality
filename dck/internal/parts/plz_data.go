package parts

import (
	"bytes"
	"fmt"
	"strings"
	"sync"

	srdata "go-secondreality/dck"
)

const plzSinBufferSize = 16384

var (
	plzDataOnce   sync.Once
	plzDataErr    error
	plzPTau       []byte
	plzPSini      []byte
	plzLsini4     []byte
	plzLsini16    []byte
	plzSinit      []int16
	plzKosinit    []int16
	plzInitTable  [10][8]int
	plzSplineCoef []int16
	plzBuu        [][]int16
	plzDtau       []uint16
)

var plzR0 = []int{0, 0, 500, 0, 0, 0, 256, 512}

func plzEnsureData() error {
	plzDataOnce.Do(func() {
		data := srdata.PLZData()
		var err error
		ptauVals, err := plzExtractArrayExprs(data, "ptau", nil)
		if err != nil {
			plzDataErr = err
			return
		}
		plzPTau = make([]byte, 256)
		for i := 0; i < len(ptauVals) && i < len(plzPTau); i++ {
			plzPTau[i] = byte(uint8(ptauVals[i]))
		}

		psiniVals, err := plzExtractArrayExprs(data, "psini", nil)
		if err != nil {
			plzDataErr = err
			return
		}
		plzPSini = make([]byte, plzSinBufferSize)
		for i := 0; i < len(psiniVals) && i < len(plzPSini); i++ {
			plzPSini[i] = byte(uint8(psiniVals[i]))
		}

		lsini4Vals, err := plzExtractArrayExprs(data, "lsini4", nil)
		if err != nil {
			plzDataErr = err
			return
		}
		plzLsini4 = plzIntsToLEBytes(lsini4Vals, plzSinBufferSize)

		lsini16Vals, err := plzExtractArrayExprs(data, "lsini16", nil)
		if err != nil {
			plzDataErr = err
			return
		}
		plzLsini16 = plzIntsToLEBytes(lsini16Vals, plzSinBufferSize)

		sinitVals, err := plzExtractArrayExprs(data, "_sinit", nil)
		if err != nil {
			plzDataErr = err
			return
		}
		plzSinit = plzBytesToInt16s(plzIntsToBytes(sinitVals))

		kosinitVals, err := plzExtractArrayExprs(data, "_kosinit", nil)
		if err != nil {
			plzDataErr = err
			return
		}
		plzKosinit = plzBytesToInt16s(plzIntsToBytes(kosinitVals))

		initVals, err := plzExtractArrayExprs(data, "inittable", nil)
		if err != nil {
			plzDataErr = err
			return
		}
		for i := 0; i < 10; i++ {
			for j := 0; j < 8; j++ {
				idx := i*8 + j
				if idx < len(initVals) {
					plzInitTable[i][j] = initVals[idx]
				} else {
					plzInitTable[i][j] = 0
				}
			}
		}

		splineVals, err := plzExtractArrayExprs(data, "splinecoef", nil)
		if err != nil {
			plzDataErr = err
			return
		}
		plzSplineCoef = make([]int16, len(splineVals))
		for i, v := range splineVals {
			plzSplineCoef[i] = int16(v)
		}

		buuVals, err := plzExtractBuu(data)
		if err != nil {
			plzDataErr = err
			return
		}
		rows := len(buuVals) / 8
		plzBuu = make([][]int16, rows)
		for i := 0; i < rows; i++ {
			row := make([]int16, 8)
			for j := 0; j < 8; j++ {
				row[j] = int16(buuVals[i*8+j])
			}
			plzBuu[i] = row
		}

		plzDtau = make([]uint16, 65)
		for i := 0; i < len(plzDtau); i++ {
			t := i * i
			t = t / 4
			t = t * 43
			t = t / 128
			t = t + 60
			if t < 0 {
				t = 0
			}
			if t > 0xFFFF {
				t = 0xFFFF
			}
			plzDtau[i] = uint16(t)
		}
	})
	return plzDataErr
}

func plzExtractArrayExprs(data []byte, name string, vars map[string]int) ([]int, error) {
	idx := bytes.Index(data, []byte(name))
	if idx == -1 {
		return nil, fmt.Errorf("plz: array %q not found", name)
	}
	brace := bytes.IndexByte(data[idx:], '{')
	if brace == -1 {
		return nil, fmt.Errorf("plz: array %q missing '{'", name)
	}
	i := idx + brace + 1
	depth := 1
	out := make([]int, 0, 1024)
	var token strings.Builder
	for i < len(data) {
		c := data[i]
		if c == '/' && i+1 < len(data) {
			if data[i+1] == '/' {
				i += 2
				for i < len(data) && data[i] != '\n' {
					i++
				}
				continue
			}
			if data[i+1] == '*' {
				i += 2
				for i+1 < len(data) && !(data[i] == '*' && data[i+1] == '/') {
					i++
				}
				if i+1 < len(data) {
					i += 2
				}
				continue
			}
		}
		if c == '{' {
			depth++
			if token.Len() > 0 {
				expr := strings.TrimSpace(token.String())
				if expr != "" {
					val, err := glenzEvalExpr(expr, vars)
					if err != nil {
						return nil, err
					}
					out = append(out, val)
				}
				token.Reset()
			}
			i++
			continue
		}
		if c == '}' {
			if token.Len() > 0 {
				expr := strings.TrimSpace(token.String())
				if expr != "" {
					val, err := glenzEvalExpr(expr, vars)
					if err != nil {
						return nil, err
					}
					out = append(out, val)
				}
				token.Reset()
			}
			depth--
			if depth == 0 {
				return out, nil
			}
			i++
			continue
		}
		if c == ',' {
			if token.Len() > 0 {
				expr := strings.TrimSpace(token.String())
				if expr != "" {
					val, err := glenzEvalExpr(expr, vars)
					if err != nil {
						return nil, err
					}
					out = append(out, val)
				}
				token.Reset()
			}
			i++
			continue
		}
		token.WriteByte(c)
		i++
	}
	return nil, fmt.Errorf("plz: array %q unterminated", name)
}

func plzExtractBuu(data []byte) ([]int, error) {
	idx := bytes.Index(data, []byte("buu"))
	if idx == -1 {
		return nil, fmt.Errorf("plz: array %q not found", "buu")
	}
	brace := bytes.IndexByte(data[idx:], '{')
	if brace == -1 {
		return nil, fmt.Errorf("plz: array %q missing '{'", "buu")
	}
	i := idx + brace + 1
	depth := 1
	out := make([]int, 0, 1024)
	var token strings.Builder
	lineStart := true

	flush := func() error {
		expr := strings.TrimSpace(token.String())
		token.Reset()
		if expr == "" {
			return nil
		}
		if expr == "R0" {
			out = append(out, plzR0...)
			return nil
		}
		val, err := glenzEvalExpr(expr, nil)
		if err != nil {
			return err
		}
		out = append(out, val)
		return nil
	}

	for i < len(data) {
		c := data[i]
		if lineStart {
			if c == ' ' || c == '\t' || c == '\r' {
				i++
				continue
			}
			if c == '\n' {
				i++
				continue
			}
			if c == '#' {
				for i < len(data) && data[i] != '\n' {
					i++
				}
				lineStart = true
				continue
			}
			lineStart = false
		}
		if c == '/' && i+1 < len(data) {
			if data[i+1] == '/' {
				i += 2
				for i < len(data) && data[i] != '\n' {
					i++
				}
				lineStart = true
				continue
			}
			if data[i+1] == '*' {
				i += 2
				for i+1 < len(data) && !(data[i] == '*' && data[i+1] == '/') {
					if data[i] == '\n' {
						lineStart = true
					}
					i++
				}
				if i+1 < len(data) {
					i += 2
				}
				continue
			}
		}
		if c == '\n' || c == '\r' {
			lineStart = true
			i++
			continue
		}

		if c == '{' {
			if err := flush(); err != nil {
				return nil, err
			}
			depth++
			i++
			continue
		}
		if c == '}' {
			if err := flush(); err != nil {
				return nil, err
			}
			depth--
			if depth == 0 {
				return out, nil
			}
			i++
			continue
		}
		if c == ',' {
			if err := flush(); err != nil {
				return nil, err
			}
			i++
			continue
		}
		token.WriteByte(c)
		i++
	}
	return nil, fmt.Errorf("plz: array %q unterminated", "buu")
}

func plzIntsToBytes(vals []int) []byte {
	out := make([]byte, len(vals))
	for i, v := range vals {
		out[i] = byte(uint8(v))
	}
	return out
}

func plzIntsToLEBytes(vals []int, size int) []byte {
	out := make([]byte, size)
	max := len(vals)
	if max*2 > size {
		max = size / 2
	}
	for i := 0; i < max; i++ {
		v := uint16(vals[i])
		out[i*2] = byte(v)
		out[i*2+1] = byte(v >> 8)
	}
	return out
}

func plzBytesToInt16s(b []byte) []int16 {
	n := len(b) / 2
	out := make([]int16, n)
	for i := 0; i < n; i++ {
		out[i] = int16(uint16(b[i*2]) | uint16(b[i*2+1])<<8)
	}
	return out
}
