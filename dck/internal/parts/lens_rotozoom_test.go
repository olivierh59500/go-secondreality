package parts

import (
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/olivierh59500/democonstructionkit/indexed"
)

func TestLensRotozoomReference(t *testing.T) {
	previousPicture, previousRotated := lensRotPic, lensRot90
	defer func() { lensRotPic, lensRot90 = previousPicture, previousRotated }()
	previousRenderer := lensRotozoom
	defer func() { lensRotozoom = previousRenderer }()
	var err error
	lensRotozoom, err = indexed.NewRotozoom256(indexed.Rotozoom256Config{Width: lensZoomXW, Height: lensZoomYW})
	if err != nil {
		t.Fatal(err)
	}
	lensRotPic, lensRot90 = make([]byte, 256*256), make([]byte, 256*256)
	for i := range lensRotPic {
		lensRotPic[i] = byte((i*37 + i/256*19) % 251)
		lensRot90[i] = byte((i*11 + i/256*53) % 247)
	}
	for _, tc := range []struct {
		name         string
		x, y, xa, ya int
		want         string
	}{
		{"ordinary", 12, 34, 61, 40, "679bbf2157e8f14b5d93a365ef81c789b0e21969736c9941c084737be55203b8"},
		{"rotated", 250, 255, -30, 120, "3ecc148e073605a86709e425c45230e47afaca6a595d787643c3ea39d787cdb2"},
		{"wrapped", -12, 320, 0, -128, "089cfb513dc1cc9394997a4b377991c77a65f0b9722fcc376a73fa0f4ac9936a"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			clear(lensZoomerPlanar[:])
			lensRotate(tc.x, tc.y, tc.xa, tc.ya)
			if got := fmt.Sprintf("%x", sha256.Sum256(lensZoomerPlanar[:16000])); got != tc.want {
				t.Fatalf("indexed rotozoom output = %s, want %s", got, tc.want)
			}
		})
	}
}
