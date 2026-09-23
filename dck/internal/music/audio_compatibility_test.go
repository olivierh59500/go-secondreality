package music

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/olivierh59500/democonstructionkit/sound"
)

// These fingerprints include five seconds of original integer PCM and 600
// audible-position marker snapshots per case, recorded before extraction.
// Soundtracks remain production assets, outside the reusable libraries.
func TestSharedReplayPreservesPCMAndMusicMarkers(t *testing.T) {
	cases := []struct {
		song, order int
		hash        string
	}{
		{0, 0, "b72d6bdf9de15fd94b916063c83ccfb55eafbae720d124a8f83d6a12bff7566a"},
		{0, 2, "931c7b6df8fd7fe336f6e1431d62e428323856a4c26e48a81bef8065d9d0e22d"},
		{0, 10, "b60b594ca36e0ae8d5cd471ffb545ba6adc5e327cc7552e4d4d9aa84be9638f6"},
		{0, 61, "68a542631e8a3788d9bcecd4b7763cc9c87b38dbec7271037cce80e7ec0c080a"},
		{1, 0, "ae1e5e1ee03bf695c3d42b2e26bf5536883b9cd7e784b59996755a7236062991"},
		{1, 2, "24a7d43497400bc331403051dde26f622ca29bb50ebe1a61f591f2bcb930ac2d"},
		{1, 10, "4599fa64e81619143b4fd2f69c6b08730558a1ef82973249eb23aba2662a67f4"},
		{1, 61, "da07eff044679795251eb0a84fcd5bfaebd44212236c7eda234452b9861920f6"},
	}
	for _, test := range cases {
		t.Run(fmt.Sprintf("song%d/order%d", test.song, test.order), func(t *testing.T) {
			stream, err := sound.Open("soundtrack.fc", RealityFC, sound.Options{Track: test.song,
				SampleRate: sampleRate, PCMFormat: sound.PCM16, Interpolation: true, StartOrder: test.order, BlockFrames: 735,
			})
			if err != nil {
				t.Fatal(err)
			}
			defer stream.Close()
			hash := sha256.New()
			pcm := make([]byte, 735*4)
			for tick := 0; tick < 300; tick++ {
				if _, err := io.ReadFull(stream, pcm); err != nil {
					t.Fatal(err)
				}
				hash.Write(pcm)
				frames := uint32((tick + 1) * 735)
				for _, frame := range []uint32{frames, frames - 735} {
					position := (time.Duration(frame)*time.Second + sampleRate - 1) / sampleRate
					snapshot, ok := stream.TrackerPositionAt(position)
					if !ok {
						t.Fatal("missing tracker markers")
					}
					binary.Write(hash, binary.LittleEndian, []uint32{uint32(snapshot.Order), uint32(snapshot.Row), snapshot.Frame, uint32(int32(snapshot.PlusFlags))})
				}
			}
			if got := fmt.Sprintf("%x", hash.Sum(nil)); got != test.hash {
				t.Fatalf("shared replay changed PCM or music markers: %s", got)
			}
		})
	}
}
