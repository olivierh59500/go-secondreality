package parts

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"
)

// plzLegacyFields retains the pre-extraction byte/word lookup and column-order
// rules independently of DCK's compiled spatial terms and row renderer.
func plzLegacyFields(dst *[2][plzSize * plzMaxY]byte, k, l [4]int) {
	readWord := func(data []byte, offset int) uint16 {
		index := (offset & (plzSinBufferSize - 1)) &^ 1
		return binary.LittleEndian.Uint16(data[index:])
	}
	var columns [2][4][plzSize]int
	for group, c := range [2][4]int{k, l} {
		c1, c2 := uint32(uint16(c[0])), uint32(uint16(c[1]))
		c3, c4 := uint32(uint16(c[2])), uint32(uint16(c[3]))
		for x := 0; x < plzSize; x++ {
			columns[group][0][x] = int(c1 + uint32(8*x))
			columns[group][1][x] = int((c2 << 1) + uint32(80*8-8*x))
			columns[group][2][x] = int(c3 + uint32(80*4-4*x))
			columns[group][3][x] = int((c4 << 1) + uint32(32*x))
		}
	}
	for field := 0; field < 2; field++ {
		for y := 0; y < plzMaxY; y++ {
			group := (y & 1) ^ field
			lc := &columns[group]
			y2 := int(uint32(uint16(y)) << 1)
			for x := 0; x < plzSize; x++ {
				column := (x &^ 3) + (3 - (x & 3))
				b1 := readWord(plzLsini16, lc[1][column]+y2)
				s1 := plzPSini[(lc[0][column]+int(b1))&(plzSinBufferSize-1)]
				b2 := readWord(plzLsini4, lc[3][column]+y2)
				s2 := plzPSini[(lc[2][column]+int(b2)+y2)&(plzSinBufferSize-1)]
				dst[field][y*plzSize+column] = byte(uint16(s1) + uint16(s2))
			}
		}
	}
}

func TestLookupPlasmaPreservesOriginalFields(t *testing.T) {
	if err := plzEnsureData(); err != nil {
		t.Fatal(err)
	}
	kernel, err := plzCreatePlasma()
	if err != nil {
		t.Fatal(err)
	}
	var got, want [2][plzSize * plzMaxY]byte
	for preset, values := range plzInitTable {
		for _, ticks := range []int{0, 1, 127, 4095, 65535, -1} {
			k := [4]int{values[4] - 3*ticks, values[5] - 2*ticks, values[6] + ticks, values[7] + 2*ticks}
			l := [4]int{values[0] - ticks, values[1] - 2*ticks, values[2] + 2*ticks, values[3] + 3*ticks}
			plzLegacyFields(&want, k, l)
			if err := plzRenderFields(kernel, &got, k, l); err != nil {
				t.Fatal(err)
			}
			for field := range got {
				if !bytes.Equal(got[field][:], want[field][:]) {
					for pixel := range got[field] {
						if got[field][pixel] != want[field][pixel] {
							t.Fatalf("preset %d tick %d field %d pixel (%d,%d): got %d want %d", preset, ticks, field, pixel%plzSize, pixel/plzSize, got[field][pixel], want[field][pixel])
						}
					}
				}
			}
		}
	}
	k, l := [4]int{3500, 2300, 3900, 3670}, [4]int{1000, 2000, 3000, 4000}
	if allocations := testing.AllocsPerRun(20, func() {
		if err := plzRenderFields(kernel, &got, k, l); err != nil {
			panic(err)
		}
	}); allocations != 0 {
		t.Fatalf("field generation allocations = %g", allocations)
	}
}

func TestLookupPlasmaOriginalInterleaveAndDrop(t *testing.T) {
	if err := plzEnsureData(); err != nil {
		t.Fatal(err)
	}
	kernel, err := plzCreatePlasma()
	if err != nil {
		t.Fatal(err)
	}
	var fields, legacy [2][plzSize * plzMaxY]byte
	k, l := [4]int{3500, 2300, 3900, 3670}, [4]int{1000, 2000, 3000, 4000}
	if err := plzRenderFields(kernel, &fields, k, l); err != nil {
		t.Fatal(err)
	}
	plzLegacyFields(&legacy, k, l)
	for _, drop := range []int{0, 60, 120, 399, 400} {
		t.Run(fmt.Sprint(drop), func(t *testing.T) {
			var got, want [320 * 400]byte
			for y := 0; y < plzMaxY && y+drop < 400; y++ {
				for x := 0; x < 80; x++ {
					source := y*plzSize + x + 2
					target := (y+drop)*320 + x*4
					got[target+0], got[target+2] = fields[1][source], fields[1][source]
					got[target+1], got[target+3] = fields[0][source], fields[0][source]
				}
			}
			for y := drop; y < min(400, drop+plzMaxY); y++ {
				for x := 0; x < 320; x++ {
					want[y*320+x] = legacy[1-(x&1)][(y-drop)*plzSize+2+x/4]
				}
			}
			if got != want {
				t.Fatal("interleaved output changed")
			}
		})
	}
}

func BenchmarkPlasmaFields(b *testing.B) {
	if err := plzEnsureData(); err != nil {
		b.Fatal(err)
	}
	kernel, err := plzCreatePlasma()
	if err != nil {
		b.Fatal(err)
	}
	var dst [2][plzSize * plzMaxY]byte
	k, l := [4]int{3500, 2300, 3900, 3670}, [4]int{1000, 2000, 3000, 4000}
	b.ReportAllocs()
	b.SetBytes(2 * plzSize * plzMaxY)
	b.ResetTimer()
	for b.Loop() {
		if err := plzRenderFields(kernel, &dst, k, l); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkOriginalPlasmaFields(b *testing.B) {
	if err := plzEnsureData(); err != nil {
		b.Fatal(err)
	}
	var dst [2][plzSize * plzMaxY]byte
	k, l := [4]int{3500, 2300, 3900, 3670}, [4]int{1000, 2000, 3000, 4000}
	b.ReportAllocs()
	b.SetBytes(2 * plzSize * plzMaxY)
	b.ResetTimer()
	for b.Loop() {
		plzLegacyFields(&dst, k, l)
	}
}
