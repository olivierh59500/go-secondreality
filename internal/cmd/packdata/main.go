package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"
)

type dataFile struct {
	source string
	target string
}

var dataFiles = []dataFile{
	{source: "data/Parts/ALKU/ALKU_MAIN_Data.dat", target: "alku.dat.gz"},
	{source: "data/Parts/INTRO3D/U2A_DATA.dat", target: "u2a.dat.gz"},
	{source: "data/Parts/VISU/VISU_Data.dat", target: "visu.dat.gz"},
	{source: "data/Parts/GLENZ/GLENZ_DATA.dat", target: "glenz.dat.gz"},
	{source: "data/Parts/TUNNELI/Tunneli_Data.dat", target: "tunneli.dat.gz"},
	{source: "data/Parts/TECHNO/KOE_Data.dat", target: "koe.dat.gz"},
	{source: "data/Parts/FOREST/Forest_Data.dat", target: "forest.dat.gz"},
	{source: "data/Parts/LENS/LENS_MAIN_Data.dat", target: "lens.dat.gz"},
	{source: "data/Parts/PLZPART/PLZ_Data.dat", target: "plz.dat.gz"},
	{source: "data/Parts/BEG/BeginTitleScreen_Data.dat", target: "beg.dat.gz"},
	{source: "data/Parts/PAM/OUTTAA_Data.dat", target: "outta-data.dat.gz"},
	{source: "data/Parts/PAM/OUTTAA.dat", target: "outta-source.dat.gz"},
	{source: "data/Blob/Blob_Data.dat", target: "blob.dat.gz"},
	{source: "data/Parts/WATER/Water_Data.dat", target: "water.dat.gz"},
	{source: "data/Parts/COMAN/COMAN_Data.dat", target: "coman.dat.gz"},
	{source: "data/Parts/END/END_Data.dat", target: "end.dat.gz"},
	{source: "data/Parts/ENDSCRL/ENDSCRL_MAIN_Data.dat", target: "endscrl.dat.gz"},
	{source: "data/Parts/JPLOGO/JP_Data.dat", target: "jplogo.dat.gz"},
	{source: "data/Parts/CREDITS/CREDITS_MAIN_Data.dat", target: "credits.dat.gz"},
}

func main() {
	root, err := os.Getwd()
	if err != nil {
		fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		fatal(fmt.Errorf("run from the repository root: %w", err))
	}

	targetDir := filepath.Join(root, "packed")
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		fatal(err)
	}

	for _, file := range dataFiles {
		if err := pack(filepath.Join(root, file.source), filepath.Join(targetDir, file.target)); err != nil {
			fatal(fmt.Errorf("pack %s: %w", file.source, err))
		}
	}
}

func pack(source, target string) error {
	data, err := os.ReadFile(source)
	if err != nil {
		return err
	}

	var compressed bytes.Buffer
	writer, err := gzip.NewWriterLevel(&compressed, gzip.BestCompression)
	if err != nil {
		return err
	}
	writer.Header.OS = 255
	if _, err := writer.Write(data); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}

	return os.WriteFile(target, compressed.Bytes(), 0o644)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
