package parts

import (
	"bytes"
	"fmt"
	"sync"

	srdata "go-secondreality/dck"
)

type comanLoopBlock struct {
	mode   uint8
	bxAdd  int16
	dxAdd  int16
	eaxInc uint32
}

var (
	comanOnce   sync.Once
	comanErr    error
	comanW1dta  []byte
	comanW2dta  []byte
	comanBlocks []comanLoopBlock
)

func comanEnsureData() error {
	comanOnce.Do(func() {
		data := srdata.ComanData()
		var err error
		comanW1dta, err = comanExtractArray(data, "w1dta")
		if err != nil {
			comanErr = err
			return
		}
		comanW2dta, err = comanExtractArray(data, "w2dta")
		if err != nil {
			comanErr = err
			return
		}
		comanBlocks, err = comanExtractBlocks(data)
		if err != nil {
			comanErr = err
			return
		}

		comanW1dta = comanPadWave(comanW1dta)
		comanW2dta = comanPadWave(comanW2dta)
	})
	return comanErr
}

func comanPadWave(data []byte) []byte {
	const minSize = 65536 + 2
	if len(data) >= minSize {
		return data
	}
	out := make([]byte, minSize)
	copy(out, data)
	return out
}

func comanExtractArray(data []byte, name string) ([]byte, error) {
	idx := bytes.Index(data, []byte(name))
	if idx == -1 {
		return nil, fmt.Errorf("coman: array %q not found", name)
	}
	brace := bytes.IndexByte(data[idx:], '{')
	if brace == -1 {
		return nil, fmt.Errorf("coman: array %q missing '{'", name)
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
	return nil, fmt.Errorf("coman: array %q unterminated", name)
}

func comanExtractBlocks(data []byte) ([]comanLoopBlock, error) {
	idx := bytes.Index(data, []byte("kBlocks"))
	if idx == -1 {
		return nil, fmt.Errorf("coman: kBlocks not found")
	}
	brace := bytes.IndexByte(data[idx:], '{')
	if brace == -1 {
		return nil, fmt.Errorf("coman: kBlocks missing '{'")
	}
	i := idx + brace + 1
	depth := 1
	nums := make([]int, 0, 1024)
	for i < len(data) && depth > 0 {
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
		switch c {
		case '{':
			depth++
			i++
			continue
		case '}':
			depth--
			i++
			continue
		}
		if c == '-' || (c >= '0' && c <= '9') {
			val, n := glenzParseNumber(data[i:])
			if n > 0 {
				nums = append(nums, val)
				i += n
				continue
			}
		}
		i++
	}
	if depth != 0 {
		return nil, fmt.Errorf("coman: kBlocks unterminated")
	}
	if len(nums)%4 != 0 {
		return nil, fmt.Errorf("coman: kBlocks count %d not multiple of 4", len(nums))
	}
	blocks := make([]comanLoopBlock, len(nums)/4)
	for i := 0; i < len(blocks); i++ {
		mode := nums[i*4]
		bx := nums[i*4+1]
		dx := nums[i*4+2]
		eax := nums[i*4+3]
		blocks[i] = comanLoopBlock{
			mode:   uint8(mode),
			bxAdd:  int16(bx),
			dxAdd:  int16(dx),
			eaxInc: uint32(eax),
		}
	}
	return blocks, nil
}
