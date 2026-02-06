package music

import (
	"encoding/binary"
	"log"
	"sync"
	"time"

	"go-secondreality/internal/st3"

	"github.com/hajimehoshi/ebiten/v2/audio"
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
	st3Player     *st3.Player
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

func playbackSamplePositionLocked() (uint32, bool) {
	if audioPlayer == nil {
		return 0, false
	}
	pos := audioPlayer.Position()
	if pos <= 0 {
		return 0, true
	}
	samplePos := (uint64(pos) * uint64(sampleRate)) / uint64(time.Second)
	if samplePos > uint64(^uint32(0)) {
		samplePos = uint64(^uint32(0))
	}
	return uint32(samplePos), true
}

func orderRowFrameLocked() (uint16, uint16, uint32) {
	if st3Player == nil {
		return 0, 0, 0
	}
	if samplePos, ok := playbackSamplePositionLocked(); ok {
		return st3Player.OrderRowFrameAt(samplePos)
	}
	return st3Player.OrderRowFrame()
}

func plusFlagsLocked() int16 {
	if st3Player == nil {
		return 0
	}
	if samplePos, ok := playbackSamplePositionLocked(); ok {
		return st3Player.PlusFlagsAt(samplePos)
	}
	return st3Player.PlusFlags()
}

func Start(song Song, startOrder byte) {
	mu.Lock()
	defer mu.Unlock()

	if song < 0 || song >= SongCount {
		return
	}

	ensureContext()

	if audioPlayer != nil {
		audioPlayer.Close()
		audioPlayer = nil
	}
	if st3Player != nil {
		st3Player.Close()
		st3Player = nil
	}
	lastFrame = 0
	lastFrameTime = time.Now()
	fallbackSync = false
	fallbackBase = 0
	fallbackStart = time.Now()

	offset := songOffset(song)
	player, err := st3.New(RealityFC[offset:], st3.Config{SampleRate: sampleRate, Interpolation: true, StartOrder: int(startOrder)})
	if err != nil {
		log.Printf("music: failed to start song: %v", err)
		fallbackSync = true
		fallbackBase = 0
		fallbackStart = time.Now()
		return
	}

	st3Player = player
	stream := newStream(player)

	audioPlayer, err = ctx.NewPlayer(stream)
	if err != nil {
		log.Printf("music: failed to create audio player: %v", err)
		st3Player.Close()
		st3Player = nil
		return
	}

	audioPlayer.Play()
}

func End() {
	mu.Lock()
	defer mu.Unlock()

	if audioPlayer != nil {
		audioPlayer.Close()
		audioPlayer = nil
	}
	if st3Player != nil {
		st3Player.Close()
		st3Player = nil
	}
	fallbackSync = false
}

func Sync() int {
	mu.Lock()
	defer mu.Unlock()

	if st3Player == nil || fallbackSync {
		return syncFallbackLocked()
	}

	order, row, frame := orderRowFrameLocked()
	orderAndRow := (order << 8) | row
	now := time.Now()
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

	if st3Player == nil || fallbackSync {
		return 0
	}

	return int(plusFlagsLocked())
}

func GetRow() int {
	mu.Lock()
	defer mu.Unlock()

	if st3Player == nil || fallbackSync {
		return 0
	}

	_, row, _ := orderRowFrameLocked()
	return int(row)
}

func GetOrder() int {
	mu.Lock()
	defer mu.Unlock()

	if st3Player == nil || fallbackSync {
		return 0
	}

	order, _, _ := orderRowFrameLocked()
	return int(order)
}

func SetFrame(frame int) {
	mu.Lock()
	defer mu.Unlock()

	if st3Player == nil {
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

	if st3Player == nil || fallbackSync {
		return 0
	}

	_, _, frame := orderRowFrameLocked()
	if frame < disFrameStart {
		return 0
	}
	return int(frame - disFrameStart)
}

func songOffset(song Song) int {
	idx := int(song) * 4
	if idx+4 > len(RealityFC) {
		return 0
	}
	return int(binary.LittleEndian.Uint32(RealityFC[idx : idx+4]))
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
	elapsed := time.Since(fallbackStart)
	advance := int(elapsed / step)
	return fallbackBase + advance
}
