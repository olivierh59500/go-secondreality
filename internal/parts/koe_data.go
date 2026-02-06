package parts

import (
	"fmt"
	"sync"

	srdata "go-secondreality"
)

var (
	koeDataOnce sync.Once
	koeDataErr  error

	koeSin1024 []int16
	koeCircle  []byte
	koeCircle2 []byte
	koePal0    []byte
	koePal1KOE []byte
	koePal2KOE []byte
	koeFlip8   []byte
)

func koeEnsureData() error {
	koeDataOnce.Do(func() {
		sinVals, err := glenzExtractArrayInts(srdata.KOEData, "koe_sin1024")
		if err != nil {
			koeDataErr = err
			return
		}
		koeSin1024 = make([]int16, len(sinVals))
		for i, v := range sinVals {
			koeSin1024[i] = int16(v)
		}

		if koeCircle, err = glenzExtractArray(srdata.KOEData, "circle"); err != nil {
			koeDataErr = err
			return
		}
		if koeCircle2, err = glenzExtractArray(srdata.KOEData, "circle2"); err != nil {
			koeDataErr = err
			return
		}
		pal2Vals, err := glenzExtractArrayExprs(srdata.KOEData, "pal2_KOEA", nil)
		if err != nil {
			koeDataErr = err
			return
		}
		koePal2KOE = make([]byte, len(pal2Vals))
		for i, v := range pal2Vals {
			koePal2KOE[i] = byte(v)
		}

		pal1Vals, err := glenzExtractArrayExprs(srdata.KOEData, "pal1_KOEA", nil)
		if err != nil {
			koeDataErr = err
			return
		}
		koePal1KOE = make([]byte, len(pal1Vals))
		for i, v := range pal1Vals {
			koePal1KOE[i] = byte(v)
		}
		if koeFlip8, err = glenzExtractArray(srdata.KOEData, "flip8"); err != nil {
			koeDataErr = err
			return
		}
		if koePal0, err = glenzExtractArray(srdata.KOEData, "pal0"); err != nil {
			koeDataErr = err
			return
		}

		if len(koeFlip8) != 256 {
			koeDataErr = fmt.Errorf("koe: flip8 len=%d", len(koeFlip8))
			return
		}
	})
	return koeDataErr
}
