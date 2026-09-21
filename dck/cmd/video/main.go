// Command video records the entire production, including the final scroll.
package main

import (
	"flag"
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/video"
	"go-secondreality/dck/internal/app"
)

func main() {
	config := video.Config{Output: "second-reality.mp4", PosterAt: 105 * time.Second, Title: "Second Reality", Width: 960, Height: 600, FPS: 60, TPS: 70, SampleRate: 44100}
	config.Flags(flag.CommandLine)
	flag.Parse()
	if err := video.Run(config, func() (ebiten.Game, error) { return app.NewRecordingGame(app.DefaultConfig()), nil }); err != nil {
		log.Fatal(err)
	}
}
