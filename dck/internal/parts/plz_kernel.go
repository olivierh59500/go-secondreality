package parts

import (
	"encoding/binary"

	"github.com/olivierh59500/democonstructionkit/plasma"
)

// plzCreatePlasma preserves the original tables and address increments while
// sharing the nested lookup renderer. Palettes, music and cube rendering remain
// independent of the indexed texture generator.
func plzCreatePlasma() (*plasma.Lookup, error) {
	words := func(data []byte) []uint16 {
		values := make([]uint16, len(data)/2)
		for i := range values {
			values[i] = binary.LittleEndian.Uint16(data[i*2:])
		}
		return values
	}
	return plasma.NewLookup(plasma.LookupConfig{
		Width: plzSize, ColorTable: plzPSini,
		Waves: []plasma.Wave{
			{
				Displacement: words(plzLsini16), ColorStepX: 8,
				DisplacementOffset: 640, DisplacementStepX: -8, DisplacementStepY: 2,
				DisplacementFractionBits: 1,
			},
			{
				Displacement: words(plzLsini4), ColorOffset: 320, ColorStepX: -4, ColorStepY: 2,
				DisplacementStepX: 32, DisplacementStepY: 2, DisplacementFractionBits: 1,
			},
		},
	})
}

func plzRenderFields(kernel *plasma.Lookup, dst *[2][plzSize * plzMaxY]byte, k, l [4]int) error {
	phases := func(values [4]int) [2]plasma.Phase {
		return [2]plasma.Phase{
			{Color: int64(uint16(values[0])), Displacement: int64(uint16(values[1])) * 2},
			{Color: int64(uint16(values[2])), Displacement: int64(uint16(values[3])) * 2},
		}
	}
	kPhase, lPhase := phases(k), phases(l)
	for field := 0; field < 2; field++ {
		if err := kernel.RenderRows(dst[field][:], plzSize,
			plasma.Rows{First: field, Count: plzMaxY / 2, Step: 2}, kPhase[:]); err != nil {
			return err
		}
		if err := kernel.RenderRows(dst[field][:], plzSize,
			plasma.Rows{First: 1 - field, Count: plzMaxY / 2, Step: 2}, lPhase[:]); err != nil {
			return err
		}
	}
	return nil
}
