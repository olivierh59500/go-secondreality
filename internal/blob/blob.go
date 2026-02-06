package blob

import (
	"bytes"
	"fmt"
	"sync"

	srdata "go-secondreality"
)

type Handle struct {
	data   []byte
	offset int
}

const (
	seekSet = 0
	seekCur = 1
	seekEnd = 2
)

var (
	blobOnce sync.Once
	blobErr  error
	blobs    map[uint32][]byte
)

func ensureBlobs() error {
	blobOnce.Do(func() {
		blobs, blobErr = parseBlobData(srdata.BlobData)
	})
	return blobErr
}

func Open(name string) *Handle {
	if err := ensureBlobs(); err != nil {
		return nil
	}
	hash := getHash(name)
	data, ok := blobs[hash]
	if !ok {
		return nil
	}
	return &Handle{data: data, offset: 0}
}

func Read(dst []byte, elementSize, elementCount int, h *Handle) int {
	if h == nil || elementSize <= 0 || elementCount <= 0 {
		return 0
	}
	count := elementSize * elementCount
	remain := len(h.data) - h.offset
	if count > remain {
		count = remain
	}
	if count <= 0 {
		return 0
	}
	if count > len(dst) {
		count = len(dst)
	}
	copy(dst, h.data[h.offset:h.offset+count])
	h.offset += count
	return count
}

func Seek(h *Handle, offset int, origin int) {
	if h == nil {
		return
	}
	switch origin {
	case seekSet:
		h.offset = offset
	case seekCur:
		h.offset += offset
	case seekEnd:
		h.offset = len(h.data) - offset
	}
	if h.offset < 0 {
		h.offset = 0
	}
	if h.offset > len(h.data) {
		h.offset = len(h.data)
	}
}

func Tell(h *Handle) int {
	if h == nil {
		return 0
	}
	return h.offset
}

func ReadAll(name string) ([]byte, error) {
	h := Open(name)
	if h == nil {
		return nil, fmt.Errorf("blob: file not found: %s", name)
	}
	if len(h.data) == 0 {
		return nil, fmt.Errorf("blob: empty data for %s", name)
	}
	out := make([]byte, len(h.data))
	copy(out, h.data)
	return out, nil
}

func getHash(name string) uint32 {
	var u1 uint16 = 0x1111
	var u2 uint16 = 0x1111
	for i := 0; i < len(name); i++ {
		ax := uint16(name[i]) &^ 0x20
		u1 ^= ax
		u1 = rotl16(u1, 1)
		u2 += ax
	}
	return (uint32(u2) << 16) | uint32(u1)
}

func rotl16(x uint16, r uint) uint16 {
	return uint16((x << (r & 15)) | (x >> ((16 - r) & 15)))
}

func parseBlobData(data []byte) (map[uint32][]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("blob: empty data source")
	}
	result := make(map[uint32][]byte)
	search := []byte("_Blob_")
	i := 0
	for {
		idx := bytes.Index(data[i:], search)
		if idx < 0 {
			break
		}
		i += idx + len(search)
		start := i
		for i < len(data) && data[i] >= '0' && data[i] <= '9' {
			i++
		}
		if start == i {
			continue
		}
		idVal := 0
		for j := start; j < i; j++ {
			idVal = idVal*10 + int(data[j]-'0')
		}
		brace := bytes.IndexByte(data[i:], '{')
		if brace < 0 {
			return nil, fmt.Errorf("blob: array %d missing '{'", idVal)
		}
		i = i + brace + 1
		out := make([]byte, 0, 1024)
		for i < len(data) {
			c := data[i]
			if c == '}' {
				i++
				break
			}
			if c == '-' || (c >= '0' && c <= '9') {
				val, n := parseNumber(data[i:])
				if n > 0 {
					out = append(out, byte(uint8(val)))
					i += n
					continue
				}
			}
			i++
		}
		result[uint32(idVal)] = out
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("blob: no arrays parsed")
	}
	return result, nil
}

func parseNumber(data []byte) (int, int) {
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
