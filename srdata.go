package srdata

import (
	"bytes"
	"compress/gzip"
	"embed"
	"fmt"
	"io"
)

//go:generate go run ./internal/cmd/packdata

// packedFiles contains deterministic gzip archives generated from the original
// C-style data files. Keeping the textual sources out of the native library
// substantially reduces the installed APK size.
//
//go:embed packed/*.gz
var packedFiles embed.FS

func AlkuData() []byte    { return readPacked("alku.dat.gz") }
func U2AData() []byte     { return readPacked("u2a.dat.gz") }
func VisuData() []byte    { return readPacked("visu.dat.gz") }
func GlenzData() []byte   { return readPacked("glenz.dat.gz") }
func TunneliData() []byte { return readPacked("tunneli.dat.gz") }
func KOEData() []byte     { return readPacked("koe.dat.gz") }
func ForestData() []byte  { return readPacked("forest.dat.gz") }
func LensData() []byte    { return readPacked("lens.dat.gz") }
func PLZData() []byte     { return readPacked("plz.dat.gz") }
func BegData() []byte     { return readPacked("beg.dat.gz") }
func OuttaData() []byte   { return readPacked("outta-data.dat.gz") }
func OuttaSource() []byte { return readPacked("outta-source.dat.gz") }
func BlobData() []byte    { return readPacked("blob.dat.gz") }
func WaterData() []byte   { return readPacked("water.dat.gz") }
func ComanData() []byte   { return readPacked("coman.dat.gz") }
func EndData() []byte     { return readPacked("end.dat.gz") }
func EndScrlData() []byte { return readPacked("endscrl.dat.gz") }
func JPLogoData() []byte  { return readPacked("jplogo.dat.gz") }
func CreditsData() []byte { return readPacked("credits.dat.gz") }

func readPacked(name string) []byte {
	compressed, err := packedFiles.ReadFile("packed/" + name)
	if err != nil {
		panic(fmt.Errorf("srdata: read %s: %w", name, err))
	}

	reader, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		panic(fmt.Errorf("srdata: open %s: %w", name, err))
	}
	data, readErr := io.ReadAll(reader)
	closeErr := reader.Close()
	if readErr != nil {
		panic(fmt.Errorf("srdata: unpack %s: %w", name, readErr))
	}
	if closeErr != nil {
		panic(fmt.Errorf("srdata: close %s: %w", name, closeErr))
	}
	return data
}
