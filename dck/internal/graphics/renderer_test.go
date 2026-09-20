package graphics

import (
	"encoding/binary"
	"testing"

	"go-secondreality/dck/internal/constants"
)

func TestFillFramePixels(t *testing.T) {
	var palette [constants.PaletteColorCount]uint32
	palette[1] = 0xFF332211
	palette[2] = 0xFF665544

	dst := make([]byte, 8)
	fillFramePixels(dst, []byte{1, 2}, &palette)

	if got := binary.LittleEndian.Uint32(dst[:4]); got != palette[1] {
		t.Fatalf("first pixel = %#08x, want %#08x", got, palette[1])
	}
	if got := binary.LittleEndian.Uint32(dst[4:]); got != palette[2] {
		t.Fatalf("second pixel = %#08x, want %#08x", got, palette[2])
	}
}

func TestModePixelCount(t *testing.T) {
	tests := []struct {
		mode  Mode
		count int
		valid bool
	}{
		{mode: Mode{Width: 320, Height: 200}, count: 64000, valid: true},
		{mode: Mode{Width: 320, Height: 400}, count: 128000, valid: true},
		{mode: Mode{Width: 640, Height: 350}, count: 224000, valid: true},
		{mode: Mode{Width: 0, Height: 200}, valid: false},
		{mode: Mode{Width: 641, Height: 200}, valid: false},
	}

	for _, tt := range tests {
		count, valid := modePixelCount(tt.mode)
		if count != tt.count || valid != tt.valid {
			t.Errorf("mode %+v = (%d, %t), want (%d, %t)", tt.mode, count, valid, tt.count, tt.valid)
		}
	}
}

func TestFrameSnapshotIsBlack(t *testing.T) {
	var snapshot frameSnapshot
	snapshot.indices = []byte{0, 1, 2}
	snapshot.palette[0] = 0xFF000000
	snapshot.palette[1] = 0xFF000000
	snapshot.palette[2] = 0xFF000000
	if !frameSnapshotIsEffectivelyBlack(&snapshot, len(snapshot.indices)) {
		t.Fatal("all-black snapshot was reported as visible")
	}

	snapshot.palette[1] = 0xFFFFFFFF
	if frameSnapshotIsEffectivelyBlack(&snapshot, len(snapshot.indices)) {
		t.Fatal("snapshot with a visible color was reported as black")
	}
}

func BenchmarkFillFramePixels320x200(b *testing.B) {
	const pixelCount = 320 * 200
	indices := make([]byte, pixelCount)
	dst := make([]byte, pixelCount*4)
	var palette [constants.PaletteColorCount]uint32
	for i := range palette {
		palette[i] = uint32(i) | uint32(i)<<8 | uint32(i)<<16 | 0xFF<<24
	}

	b.SetBytes(pixelCount * 5)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fillFramePixels(dst, indices, &palette)
	}
}
