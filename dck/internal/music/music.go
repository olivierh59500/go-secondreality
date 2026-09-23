package music

import (
	"log"
	"sync"
	"time"

	"github.com/olivierh59500/democonstructionkit/sound"

	audio "github.com/olivierh59500/democonstructionkit/sound/output"
)

type Song int

const (
	SongSkaven Song = iota
	SongPurpleMotion
	SongCount
)

const sampleRate = 44100

var (
	mu            sync.Mutex
	ctx           *audio.Context
	audioPlayer   *audio.Player
	musicStream   *sound.Stream
	disFrameStart uint32
	lastFrame     uint32
	lastFrameTime time.Time
	fallbackSync  bool
	fallbackBase  int
	fallbackStart time.Time
)

type syncEntry struct {
	orderAndRow uint16
	syncNumber  uint16
}

var syncData = []syncEntry{
	{0x0000, 0}, {0x0200, 1}, {0x0300, 2}, {0x032f, 3},
	{0x042f, 4}, {0x052f, 5}, {0x062f, 6}, {0x072f, 7},
	{0x082f, 8}, {0x0900, 9}, {0x0d00, 10}, {0x3d00, 1},
	{0x3f00, 2}, {0x4100, 3}, {0x4200, 4},
}

func ensureContext() {
	if ctx == nil {
		ctx = audio.NewContext(sampleRate)
	}
}

func trackerPositionLocked() sound.TrackerPosition {
	if musicStream == nil {
		return sound.TrackerPosition{}
	}
	position := musicStream.Position()
	if audioPlayer != nil {
		position = audioPlayer.Position()
	}
	snapshot, _ := musicStream.TrackerPositionAt(position)
	return snapshot
}

func orderRowFrameLocked() (uint16, uint16, uint32) {
	snapshot := trackerPositionLocked()
	return snapshot.Order, snapshot.Row, snapshot.Frame
}

func plusFlagsLocked() int16 { return trackerPositionLocked().PlusFlags }

func Start(song Song, startOrder byte) {
	mu.Lock()
	defer mu.Unlock()

	if song < 0 || song >= SongCount {
		return
	}

	ensureContext()

	if audioPlayer != nil {
		if err := audioPlayer.Close(); err != nil {
			log.Printf("music: failed to close audio player: %v", err)
		}
		audioPlayer = nil
	}
	if musicStream != nil {
		musicStream.Close()
		musicStream = nil
	}
	lastFrame = 0
	lastFrameTime = audio.Now()
	fallbackSync = false
	fallbackBase = 0
	fallbackStart = audio.Now()

	player, err := sound.Open("soundtrack.fc", RealityFC, sound.Options{Track: int(song), SampleRate: sampleRate, PCMFormat: sound.PCM16, Interpolation: true, StartOrder: int(startOrder)})
	if err != nil {
		log.Printf("music: failed to start song: %v", err)
		fallbackSync = true
		fallbackBase = 0
		fallbackStart = audio.Now()
		return
	}

	musicStream = player
	audioPlayer, err = ctx.NewPlayer(player)
	if err != nil {
		log.Printf("music: failed to create audio player: %v", err)
		musicStream.Close()
		musicStream = nil
		return
	}

	audioPlayer.Play()
}

func End() {
	mu.Lock()
	defer mu.Unlock()

	if audioPlayer != nil {
		if err := audioPlayer.Close(); err != nil {
			log.Printf("music: failed to close audio player: %v", err)
		}
		audioPlayer = nil
	}
	if musicStream != nil {
		musicStream.Close()
		musicStream = nil
	}
	fallbackSync = false
}

func Sync() int {
	mu.Lock()
	defer mu.Unlock()

	if musicStream == nil || fallbackSync {
		return syncFallbackLocked()
	}

	order, row, frame := orderRowFrameLocked()
	orderAndRow := (order << 8) | row
	now := audio.Now()
	if frame != lastFrame {
		lastFrame = frame
		lastFrameTime = now
	} else if now.Sub(lastFrameTime) > 2*time.Second {
		fallbackSync = true
		fallbackBase = lastSyncLocked(orderAndRow)
		fallbackStart = now
		log.Printf("music: sync fallback enabled (audio stalled)")
		return syncFallbackLocked()
	}

	return lastSyncLocked(orderAndRow)
}

func GetPlusFlags() int {
	mu.Lock()
	defer mu.Unlock()

	if musicStream == nil || fallbackSync {
		return 0
	}

	return int(plusFlagsLocked())
}

func GetRow() int {
	mu.Lock()
	defer mu.Unlock()

	if musicStream == nil || fallbackSync {
		return 0
	}

	_, row, _ := orderRowFrameLocked()
	return int(row)
}

func GetOrder() int {
	mu.Lock()
	defer mu.Unlock()

	if musicStream == nil || fallbackSync {
		return 0
	}

	order, _, _ := orderRowFrameLocked()
	return int(order)
}

func GetOrderRow() (int, int) {
	mu.Lock()
	defer mu.Unlock()

	if musicStream == nil || fallbackSync {
		return 0, 0
	}

	order, row, _ := orderRowFrameLocked()
	return int(order), int(row)
}

func SetFrame(frame int) {
	mu.Lock()
	defer mu.Unlock()

	if musicStream == nil {
		disFrameStart = 0
		return
	}

	_, _, current := orderRowFrameLocked()
	if frame < 0 {
		frame = 0
	}
	if current >= uint32(frame) {
		disFrameStart = current - uint32(frame)
	} else {
		disFrameStart = 0
	}
}

func GetFrame() int {
	mu.Lock()
	defer mu.Unlock()

	if musicStream == nil || fallbackSync {
		return 0
	}

	_, _, frame := orderRowFrameLocked()
	if frame < disFrameStart {
		return 0
	}
	return int(frame - disFrameStart)
}

func lastSyncLocked(orderAndRow uint16) int {
	for i := 1; i < len(syncData); i++ {
		if syncData[i].orderAndRow >= orderAndRow {
			return int(syncData[i-1].syncNumber)
		}
	}
	return 0
}

func syncFallbackLocked() int {
	if !fallbackSync {
		return 0
	}
	const step = 3 * time.Second
	elapsed := audio.Now().Sub(fallbackStart)
	advance := int(elapsed / step)
	return fallbackBase + advance
}
