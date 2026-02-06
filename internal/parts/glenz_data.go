package parts

import (
	"bytes"
	"fmt"
	"strings"
	"sync"

	srdata "go-secondreality"
)

var (
	glenzDataOnce  sync.Once
	glenzDataErr   error
	glenzFC        []byte
	glenzSinTable  []int16
	glenzCosTable  []int16
	glenzPoints    []int32
	glenzPointsB   []int32
	glenzEPolys    []uint16
	glenzEPolysB   []uint16
)

func glenzEnsureData() error {
	glenzDataOnce.Do(func() {
		var err error
		glenzFC, err = glenzExtractArray(srdata.GlenzData, "fc")
		if err != nil {
			glenzDataErr = err
			return
		}
		sinVals, err := glenzExtractArrayInts(srdata.GlenzData, "sintable16")
		if err != nil {
			glenzDataErr = err
			return
		}
		cosVals, err := glenzExtractArrayInts(srdata.GlenzData, "costable16")
		if err != nil {
			glenzDataErr = err
			return
		}
		pointVals, err := glenzExtractArrayExprs(srdata.GlenzData, "points", map[string]int{
			"ZZZ": 50,
			"QQQ": 99,
		})
		if err != nil {
			glenzDataErr = err
			return
		}
		pointBVals, err := glenzExtractArrayExprs(srdata.GlenzData, "pointsb", map[string]int{
			"ZZZ": 50,
			"QQQ": 99,
		})
		if err != nil {
			glenzDataErr = err
			return
		}
		epolysVals, err := glenzExtractArrayInts(srdata.GlenzData, "epolys")
		if err != nil {
			glenzDataErr = err
			return
		}
		epolysBVals, err := glenzExtractArrayInts(srdata.GlenzData, "epolysb")
		if err != nil {
			glenzDataErr = err
			return
		}

		glenzSinTable = make([]int16, len(sinVals))
		for i, v := range sinVals {
			glenzSinTable[i] = int16(v)
		}
		glenzCosTable = make([]int16, len(cosVals))
		for i, v := range cosVals {
			glenzCosTable[i] = int16(v)
		}
		glenzPoints = make([]int32, len(pointVals))
		for i, v := range pointVals {
			glenzPoints[i] = int32(v)
		}
		glenzPointsB = make([]int32, len(pointBVals))
		for i, v := range pointBVals {
			glenzPointsB[i] = int32(v)
		}
		glenzEPolys = make([]uint16, len(epolysVals))
		for i, v := range epolysVals {
			glenzEPolys[i] = uint16(v)
		}
		glenzEPolysB = make([]uint16, len(epolysBVals))
		for i, v := range epolysBVals {
			glenzEPolysB[i] = uint16(v)
		}
	})
	return glenzDataErr
}

func glenzExtractArray(data []byte, name string) ([]byte, error) {
	idx := bytes.Index(data, []byte(name))
	if idx == -1 {
		return nil, fmt.Errorf("glenz: array %q not found", name)
	}
	brace := bytes.IndexByte(data[idx:], '{')
	if brace == -1 {
		return nil, fmt.Errorf("glenz: array %q missing '{'", name)
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
	return nil, fmt.Errorf("glenz: array %q unterminated", name)
}

func glenzExtractArrayInts(data []byte, name string) ([]int, error) {
	idx := bytes.Index(data, []byte(name))
	if idx == -1 {
		return nil, fmt.Errorf("glenz: array %q not found", name)
	}
	brace := bytes.IndexByte(data[idx:], '{')
	if brace == -1 {
		return nil, fmt.Errorf("glenz: array %q missing '{'", name)
	}
	i := idx + brace + 1
	out := make([]int, 0, 1024)
	for i < len(data) {
		c := data[i]
		if c == '}' {
			return out, nil
		}
		if c == '-' || (c >= '0' && c <= '9') {
			val, n := glenzParseNumber(data[i:])
			if n > 0 {
				out = append(out, val)
				i += n
				continue
			}
		}
		i++
	}
	return nil, fmt.Errorf("glenz: array %q unterminated", name)
}

func glenzExtractArrayExprs(data []byte, name string, vars map[string]int) ([]int, error) {
	idx := bytes.Index(data, []byte(name))
	if idx == -1 {
		return nil, fmt.Errorf("glenz: array %q not found", name)
	}
	brace := bytes.IndexByte(data[idx:], '{')
	if brace == -1 {
		return nil, fmt.Errorf("glenz: array %q missing '{'", name)
	}
	i := idx + brace + 1
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
			}
			return out, nil
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
	return nil, fmt.Errorf("glenz: array %q unterminated", name)
}

func glenzEvalExpr(expr string, vars map[string]int) (int, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return 0, fmt.Errorf("glenz: empty expression")
	}
	val, rest, ok := glenzParseTermExpr(expr, vars)
	if !ok {
		return 0, fmt.Errorf("glenz: invalid expression %q", expr)
	}
	rest = strings.TrimSpace(rest)
	for rest != "" {
		op := rest[0]
		if op != '*' && op != '/' {
			return 0, fmt.Errorf("glenz: unsupported expression %q", expr)
		}
		term, rest2, ok := glenzParseTermExpr(rest[1:], vars)
		if !ok {
			return 0, fmt.Errorf("glenz: invalid expression %q", expr)
		}
		if op == '*' {
			val *= term
		} else {
			if term == 0 {
				return 0, fmt.Errorf("glenz: division by zero in %q", expr)
			}
			val /= term
		}
		rest = strings.TrimSpace(rest2)
	}
	return val, nil
}

func glenzParseTermExpr(expr string, vars map[string]int) (int, string, bool) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return 0, "", false
	}
	if expr[0] == '+' {
		expr = strings.TrimSpace(expr[1:])
		if expr == "" {
			return 0, "", false
		}
	}
	if isIdentStart(expr[0]) {
		i := 1
		for i < len(expr) && isIdentChar(expr[i]) {
			i++
		}
		name := expr[:i]
		val, ok := vars[name]
		if !ok {
			return 0, "", false
		}
		return val, expr[i:], true
	}
	val, n := glenzParseNumber([]byte(expr))
	if n == 0 {
		return 0, "", false
	}
	return val, expr[n:], true
}

func isIdentStart(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_'
}

func isIdentChar(c byte) bool {
	return isIdentStart(c) || (c >= '0' && c <= '9')
}

func glenzParseNumber(data []byte) (int, int) {
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
