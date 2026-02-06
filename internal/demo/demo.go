package demo

import (
	"go-secondreality/internal/driver"
	"go-secondreality/internal/music"
	"go-secondreality/internal/parts"
	"go-secondreality/internal/shim"
)

type StartPart int

const (
	PartAlkutekstitI StartPart = iota
	PartAlkutekstitII
	PartAlkutekstitIII
	PartLogo
	PartGlenz
	PartDottitunneli
	PartTechno
	PartPanicfake
	PartVuoriScrolli
	PartRotazoomer
	PartPlasmacube
	PartMiniVectorBalls
	PartPeilipalloscroll
	Part3DSinusfield
	PartJellypic
	PartVectorPartII
	PartEndpictureflash
	PartCreditsGreetings
	PartEndScrolling
	PartEmpty0
	PartHidden
	PartEmpty1
)

type Config struct {
	StartPart StartPart
	Loop      bool
}

type PartInfo struct {
	Music           music.Song
	MusicStartOrder byte
	Width           int
	Height          int
	Part            func()
}

var partList = []PartInfo{
	/* 00 Alkutekstit I     */ {Music: music.SongSkaven, MusicStartOrder: 0x00, Width: 320, Height: 400, Part: parts.Alku},
	/* 01 Alkutekstit II    */ {Music: music.SongSkaven, MusicStartOrder: 0x0C, Width: 320, Height: 200, Part: parts.U2A},
	/* 02 Alkutekstit III   */ {Music: music.SongSkaven, MusicStartOrder: 0x0D, Width: 320, Height: 200, Part: parts.OUTTA},
	/* 03 Logo              */ {Music: music.SongSkaven, MusicStartOrder: 0x0E, Width: 320, Height: 400, Part: parts.Beg},
	/* 04 Glenz             */ {Music: music.SongPurpleMotion, MusicStartOrder: 0x00, Width: 320, Height: 400, Part: parts.Glenz},
	/* 05 Dottitunneli      */ {Music: music.SongPurpleMotion, MusicStartOrder: 0x0F, Width: 320, Height: 200, Part: parts.Tunneli},
	/* 06 Techno            */ {Music: music.SongPurpleMotion, MusicStartOrder: 0x14, Width: 320, Height: 200, Part: parts.KOE},
	/* 07 Panicfake         */ {Music: music.SongPurpleMotion, MusicStartOrder: 0x27, Width: 320, Height: 400, Part: parts.Shutdown},
	/* 08 Vuori-Scrolli     */ {Music: music.SongPurpleMotion, MusicStartOrder: 0x2A, Width: 320, Height: 200, Part: parts.Forest},
	/* 09 Rotazoomer        */ {Music: music.SongPurpleMotion, MusicStartOrder: 0x2F, Width: 320, Height: 200, Part: parts.Lens},
	/* 10 Plasmacube        */ {Music: music.SongPurpleMotion, MusicStartOrder: 0x3E, Width: 320, Height: 200, Part: parts.PLZ},
	/* 11 MiniVectorBalls   */ {Music: music.SongPurpleMotion, MusicStartOrder: 0x4D, Width: 320, Height: 200, Part: parts.Dots},
	/* 12 Peilipalloscroll  */ {Music: music.SongPurpleMotion, MusicStartOrder: 0x58, Width: 320, Height: 200, Part: parts.Water},
	/* 13 3D-Sinusfield     */ {Music: music.SongPurpleMotion, MusicStartOrder: 0x5E, Width: 320, Height: 200, Part: parts.Coman},
	/* 14 Jellypic          */ {Music: music.SongPurpleMotion, MusicStartOrder: 0x62, Width: 320, Height: 400, Part: parts.JPLogo},
	/* 15 Vector Part II    */ {Music: music.SongSkaven, MusicStartOrder: 0x12, Width: 320, Height: 400, Part: parts.U2E},
	/* 16 Endpictureflash   */ {Music: music.SongSkaven, MusicStartOrder: 0x19, Width: 320, Height: 400, Part: parts.End},
	/* 17 Credits/Greetings */ {Music: music.SongSkaven, MusicStartOrder: 0x1C, Width: 320, Height: 400, Part: parts.Credits},
	/* 18 EndScrolling      */ {Music: music.SongSkaven, MusicStartOrder: 0x2B, Width: 640, Height: 350, Part: parts.EndScrl},
	{Music: music.SongCount, MusicStartOrder: 0x00, Width: 0, Height: 0, Part: nil},
	/* 19 HiddenPart        */ {Music: music.SongSkaven, MusicStartOrder: 0x46, Width: 320, Height: 400, Part: parts.DDStars},
	{Music: music.SongCount, MusicStartOrder: 0x00, Width: 0, Height: 0, Part: nil},
}

func Run(cfg Config) {
	lastMusic := music.SongCount

	startIdx := int(cfg.StartPart)
	if startIdx < 0 || startIdx >= len(partList) {
		startIdx = 0
	}

	for {
		for i := startIdx; i < len(partList) && partList[i].Width != 0; i++ {
			info := partList[i]
			if info.Music != lastMusic {
				lastMusic = info.Music
				music.Start(info.Music, info.MusicStartOrder)
			}

			driver.ChangeMode(info.Width, info.Height, shim.DefaultJSSS)
			if info.Part != nil {
				info.Part()
			}

			if driver.WantsToQuit() {
				music.End()
				return
			}

			shim.FinishedDemoFirstPart()

			if cfg.Loop && i == int(PartCreditsGreetings) {
				lastMusic = music.SongCount
				i = -1
			}
		}

		if driver.WantsToQuit() || !cfg.Loop {
			break
		}
	}

	music.End()
}
