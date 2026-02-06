package app

import (
	"errors"

	"go-secondreality/internal/demo"
)

type Config struct {
	StartPart demo.StartPart
	Loop      bool
	Windowed  bool
}

const parameterList = "Parameter List:\n\n" +
	"2: Logo\n" +
	"3: Vuori Scrolli\n" +
	"4: Vector Part II\n" +
	"5: End Scrolling\n" +
	"6: 3D Sinusfield\n\n" +
	"U: Hidden Part\n\n" +
	"L: Looping\n\n" +
	"W: Window Mode\n" +
	"F: Fullscreen Mode\n"

func ParseArgs(args []string) (Config, error) {
	cfg := Config{
		StartPart: demo.PartAlkutekstitI,
		Loop:      false,
		Windowed:  true,
	}

	if len(args) == 0 {
		return cfg, nil
	}

	invalid := false
	for _, arg := range args {
		if arg == "" {
			continue
		}
		switch arg[0] {
		case '2':
			cfg.StartPart = demo.PartLogo
		case '3':
			cfg.StartPart = demo.PartVuoriScrolli
		case '4':
			cfg.StartPart = demo.PartVectorPartII
		case '5':
			cfg.StartPart = demo.PartEndScrolling
		case '6':
			cfg.StartPart = demo.Part3DSinusfield
		case 'u', 'U':
			cfg.StartPart = demo.PartHidden
		case 'l', 'L':
			cfg.Loop = true
		case 'w', 'W':
			cfg.Windowed = true
		case 'f', 'F':
			cfg.Windowed = false
		default:
			invalid = true
		}
	}

	if invalid {
		return cfg, errors.New(parameterList)
	}

	return cfg, nil
}
