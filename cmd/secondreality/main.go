package main

import (
	"log"
	"os"

	"go-secondreality/internal/app"
)

func main() {
	cfg, err := app.ParseArgs(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	if err := app.Run(cfg); err != nil {
		log.Fatal(err)
	}
}
