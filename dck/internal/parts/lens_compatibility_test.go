package parts

import (
	"crypto/sha256"
	"fmt"
	"testing"

	"go-secondreality/dck/internal/constants"
	"go-secondreality/dck/internal/shim"
)

// This fingerprint was recorded from the original indexed renderer. It covers
// every supplied path point as well as clipped and odd-address placements.
func TestIndexedLensPreservesOriginalPixels(t *testing.T) {
	if err := lensEnsureData(); err != nil {
		t.Fatal(err)
	}
	lensW, lensH = int(lensReadU16(lensEx0)), int(lensReadU16(lensEx0[2:]))
	var err error
	lensMapper, err = lensCreateMapper()
	if err != nil {
		t.Fatal(err)
	}
	lensBack = make([]byte, constants.ScreenSize+4096)
	for i := range lensBack {
		lensBack[i] = byte((i*37 + i/19) & 63)
	}
	prior := shim.VRAM
	shim.VRAM = make([]byte, constants.ScreenSize)
	defer func() { shim.VRAM = prior }()
	positions := [][2]int{{160, 100}, {161, 100}, {0, 0}, {-1, 0}, {320, 200}, {319, 199}, {160, -20}, {160, 220}, {-80, 100}, {400, 100}}
	n := int(lensReadU16(lensExp[2:]))
	for i := 0; i+1 < n; i += 2 {
		positions = append(positions, [2]int{int(int16(lensReadU16(lensExp[4+i*2:]))), int(int16(lensReadU16(lensExp[6+i*2:])))})
	}
	hash := sha256.New()
	for _, p := range positions {
		for i := range shim.VRAM {
			shim.VRAM[i] = byte(i*11 + i/7)
		}
		lensDrawLens(p[0], p[1])
		hash.Write(shim.VRAM)
	}
	if got, want := fmt.Sprintf("%x", hash.Sum(nil)), "373eeb2ac31ff8ddff39fdc2a8244c83c2dbae0a41cdc964e5e12f63375f615b"; got != want {
		t.Fatalf("lens pixels: %s, want %s (%d positions)", got, want, len(positions))
	}
}

func BenchmarkIndexedLens(b *testing.B) {
	if err := lensEnsureData(); err != nil {
		b.Fatal(err)
	}
	lensW, lensH = int(lensReadU16(lensEx0)), int(lensReadU16(lensEx0[2:]))
	mapper, err := lensCreateMapper()
	if err != nil {
		b.Fatal(err)
	}
	source := make([]byte, constants.ScreenSize+4096)
	destination := make([]byte, constants.ScreenSize)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mapper.Draw(destination, source, 160+i%2, 100)
	}
}
