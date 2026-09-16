package srdata

import (
	"bytes"
	"os"
	"testing"
)

func TestPackedDataMatchesSources(t *testing.T) {
	tests := []struct {
		name   string
		source string
		load   func() []byte
	}{
		{name: "alku", source: "data/Parts/ALKU/ALKU_MAIN_Data.dat", load: AlkuData},
		{name: "u2a", source: "data/Parts/INTRO3D/U2A_DATA.dat", load: U2AData},
		{name: "visu", source: "data/Parts/VISU/VISU_Data.dat", load: VisuData},
		{name: "glenz", source: "data/Parts/GLENZ/GLENZ_DATA.dat", load: GlenzData},
		{name: "tunneli", source: "data/Parts/TUNNELI/Tunneli_Data.dat", load: TunneliData},
		{name: "koe", source: "data/Parts/TECHNO/KOE_Data.dat", load: KOEData},
		{name: "forest", source: "data/Parts/FOREST/Forest_Data.dat", load: ForestData},
		{name: "lens", source: "data/Parts/LENS/LENS_MAIN_Data.dat", load: LensData},
		{name: "plz", source: "data/Parts/PLZPART/PLZ_Data.dat", load: PLZData},
		{name: "beg", source: "data/Parts/BEG/BeginTitleScreen_Data.dat", load: BegData},
		{name: "outta-data", source: "data/Parts/PAM/OUTTAA_Data.dat", load: OuttaData},
		{name: "outta-source", source: "data/Parts/PAM/OUTTAA.dat", load: OuttaSource},
		{name: "blob", source: "data/Blob/Blob_Data.dat", load: BlobData},
		{name: "water", source: "data/Parts/WATER/Water_Data.dat", load: WaterData},
		{name: "coman", source: "data/Parts/COMAN/COMAN_Data.dat", load: ComanData},
		{name: "end", source: "data/Parts/END/END_Data.dat", load: EndData},
		{name: "endscrl", source: "data/Parts/ENDSCRL/ENDSCRL_MAIN_Data.dat", load: EndScrlData},
		{name: "jplogo", source: "data/Parts/JPLOGO/JP_Data.dat", load: JPLogoData},
		{name: "credits", source: "data/Parts/CREDITS/CREDITS_MAIN_Data.dat", load: CreditsData},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want, err := os.ReadFile(tt.source)
			if err != nil {
				t.Fatal(err)
			}
			if got := tt.load(); !bytes.Equal(got, want) {
				t.Fatalf("unpacked data differs: got %d bytes, want %d", len(got), len(want))
			}
		})
	}
}
