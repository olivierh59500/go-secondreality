package parts

import "testing"

func TestPackedPartDataLoads(t *testing.T) {
	tests := []struct {
		name string
		load func() error
	}{
		{name: "alku", load: alkuEnsureData},
		{name: "beg", load: begEnsureData},
		{name: "coman", load: comanEnsureData},
		{name: "credits", load: creditsEnsureData},
		{name: "end", load: endEnsureData},
		{name: "endscrl", load: endscrlEnsureData},
		{name: "forest", load: forestEnsureData},
		{name: "glenz", load: glenzEnsureData},
		{name: "jplogo", load: jpEnsureData},
		{name: "koe", load: koeEnsureData},
		{name: "lens", load: lensEnsureData},
		{name: "outta", load: outtaEnsureData},
		{name: "plz", load: plzEnsureData},
		{name: "tunneli", load: tunneliEnsureData},
		{name: "u2a", load: u2aEnsureData},
		{name: "visu", load: visuEnsureData},
		{name: "water", load: waterEnsureData},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.load(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
