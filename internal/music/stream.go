package music

import (
	"io"
	"unsafe"

	"go-secondreality/internal/st3"
)

type st3Stream struct {
	player *st3.Player
}

func newStream(player *st3.Player) io.Reader {
	return &st3Stream{player: player}
}

func (s *st3Stream) Read(p []byte) (int, error) {
	if s.player == nil {
		for i := range p {
			p[i] = 0
		}
		return len(p), nil
	}

	n := len(p) - (len(p) % 4)
	if n <= 0 {
		return 0, nil
	}

	buf := p[:n]
	out := unsafe.Slice((*int16)(unsafe.Pointer(&buf[0])), len(buf)/2)
	s.player.Fill(out)
	return n, nil
}
