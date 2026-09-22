package parts

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/olivierh59500/democonstructionkit/indexed"
)

// scatterLegacy decodes the original per-source count/address stream every
// frame. It intentionally does not use DCK's compiled map or combine helpers.
func scatterLegacy(dst, stream, source, background []byte, additive bool) {
	position := 0
	for _, value := range source {
		if position+2 > len(stream) {
			return
		}
		count := int(binary.LittleEndian.Uint16(stream[position:]))
		position += 2
		for i := 0; i < count && position+2 <= len(stream); i++ {
			target := int(binary.LittleEndian.Uint16(stream[position:]))
			position += 2
			if target >= len(dst) || target >= len(background) {
				continue
			}
			if additive {
				dst[target] = background[target] + value
			} else if value == 0 {
				dst[target] = background[target]
			} else {
				dst[target] = value
			}
		}
	}
}

func TestSharedScatterPreservesForestAndWater(t *testing.T) {
	if err := forestEnsureData(); err != nil {
		t.Fatal(err)
	}
	if err := waterEnsureData(); err != nil {
		t.Fatal(err)
	}
	for _, effect := range []struct {
		name    string
		streams [3][]byte
		size    int
		bg      []byte
		mode    indexed.ScatterMode
	}{
		{"forest", [3][]byte{forestPosi1, forestPosi2, forestPosi3}, 237 * 31, forestHBack[778:], indexed.ScatterAddBackground},
		{"water", [3][]byte{waterWat1, waterWat2, waterWat3}, waterFBSize, waterTausta[778:], indexed.ScatterOverBackground},
	} {
		maps, err := compileScatterMaps(effect.streams, effect.size, 320*200)
		if err != nil {
			t.Fatalf("%s maps: %v", effect.name, err)
		}
		for phase, stream := range effect.streams {
			for frame := 0; frame < 16; frame++ {
				t.Run(fmt.Sprintf("%s/phase%d/frame%d", effect.name, phase, frame), func(t *testing.T) {
					source := make([]byte, effect.size)
					for i := range source {
						// All 256 indices appear, including transparency, with shifted
						// source columns and wrapping addition on successive frames.
						source[i] = byte(i*17 + frame*13)
					}
					got := bytes.Repeat([]byte{0x71}, 320*200)
					want := append([]byte(nil), got...)
					scatterLegacy(want, stream, source, effect.bg, effect.mode == indexed.ScatterAddBackground)
					if err := maps[phase].Render(got, source, effect.bg, effect.mode); err != nil {
						t.Fatal(err)
					}
					if !bytes.Equal(got, want) {
						for i := range got {
							if got[i] != want[i] {
								t.Fatalf("pixel %d: got %d, want %d", i, got[i], want[i])
							}
						}
					}
					if allocations := testing.AllocsPerRun(5, func() {
						if err := maps[phase].Render(got, source, effect.bg, effect.mode); err != nil {
							panic(err)
						}
					}); allocations != 0 {
						t.Fatalf("scatter allocations = %g", allocations)
					}
				})
			}
		}
	}
}

func BenchmarkSharedScatter(b *testing.B) {
	if err := forestEnsureData(); err != nil {
		b.Fatal(err)
	}
	maps, err := compileScatterMaps([3][]byte{forestPosi1, forestPosi2, forestPosi3}, 237*31, 320*200)
	if err != nil {
		b.Fatal(err)
	}
	dst, source := make([]byte, 320*200), make([]byte, 237*31)
	for i := range source {
		source[i] = byte(i)
	}
	b.Run("compiled", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			if err := maps[0].Render(dst, source, forestHBack[778:], indexed.ScatterAddBackground); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("legacy", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			scatterLegacy(dst, forestPosi1, source, forestHBack[778:], true)
		}
	})
}
