package parts

import (
	"bytes"
	"fmt"
	"sync"

	srdata "go-secondreality"
)

const (
	creditsFontRows   = 32
	creditsFontStride = 1500
)

var (
	creditsOnce sync.Once
	creditsErr  error

	creditsPic1   []byte
	creditsPic2   []byte
	creditsPic3   []byte
	creditsPic4   []byte
	creditsPic5   []byte
	creditsPic5b  []byte
	creditsPic6   []byte
	creditsPic7   []byte
	creditsPic8   []byte
	creditsPic9   []byte
	creditsPic10  []byte
	creditsPic10b []byte
	creditsPic11  []byte
	creditsPic12  []byte
	creditsPic13  []byte
	creditsPic14  []byte
	creditsPic14b []byte
	creditsPic15  []byte
	creditsPic16  []byte
	creditsPic17  []byte
	creditsPic18  []byte

	creditsFont []byte
)

func creditsEnsureData() error {
	creditsOnce.Do(func() {
		var err error
		creditsPic1, err = creditsExtractArray(srdata.CreditsData, "credits_pic1")
		if err != nil {
			creditsErr = err
			return
		}
		creditsPic2, err = creditsExtractArray(srdata.CreditsData, "credits_pic2")
		if err != nil {
			creditsErr = err
			return
		}
		creditsPic3, err = creditsExtractArray(srdata.CreditsData, "credits_pic3")
		if err != nil {
			creditsErr = err
			return
		}
		creditsPic4, err = creditsExtractArray(srdata.CreditsData, "credits_pic4")
		if err != nil {
			creditsErr = err
			return
		}
		creditsPic5, err = creditsExtractArray(srdata.CreditsData, "credits_pic5")
		if err != nil {
			creditsErr = err
			return
		}
		creditsPic5b, err = creditsExtractArray(srdata.CreditsData, "credits_pic5b")
		if err != nil {
			creditsErr = err
			return
		}
		creditsPic6, err = creditsExtractArray(srdata.CreditsData, "credits_pic6")
		if err != nil {
			creditsErr = err
			return
		}
		creditsPic7, err = creditsExtractArray(srdata.CreditsData, "credits_pic7")
		if err != nil {
			creditsErr = err
			return
		}
		creditsPic8, err = creditsExtractArray(srdata.CreditsData, "credits_pic8")
		if err != nil {
			creditsErr = err
			return
		}
		creditsPic9, err = creditsExtractArray(srdata.CreditsData, "credits_pic9")
		if err != nil {
			creditsErr = err
			return
		}
		creditsPic10, err = creditsExtractArray(srdata.CreditsData, "credits_pic10")
		if err != nil {
			creditsErr = err
			return
		}
		creditsPic10b, err = creditsExtractArray(srdata.CreditsData, "credits_pic10b")
		if err != nil {
			creditsErr = err
			return
		}
		creditsPic11, err = creditsExtractArray(srdata.CreditsData, "credits_pic11")
		if err != nil {
			creditsErr = err
			return
		}
		creditsPic12, err = creditsExtractArray(srdata.CreditsData, "credits_pic12")
		if err != nil {
			creditsErr = err
			return
		}
		creditsPic13, err = creditsExtractArray(srdata.CreditsData, "credits_pic13")
		if err != nil {
			creditsErr = err
			return
		}
		creditsPic14, err = creditsExtractArray(srdata.CreditsData, "credits_pic14")
		if err != nil {
			creditsErr = err
			return
		}
		creditsPic14b, err = creditsExtractArray(srdata.CreditsData, "credits_pic14b")
		if err != nil {
			creditsErr = err
			return
		}
		creditsPic15, err = creditsExtractArray(srdata.CreditsData, "credits_pic15")
		if err != nil {
			creditsErr = err
			return
		}
		creditsPic16, err = creditsExtractArray(srdata.CreditsData, "credits_pic16")
		if err != nil {
			creditsErr = err
			return
		}
		creditsPic17, err = creditsExtractArray(srdata.CreditsData, "credits_pic17")
		if err != nil {
			creditsErr = err
			return
		}
		creditsPic18, err = creditsExtractArray(srdata.CreditsData, "credits_pic18")
		if err != nil {
			creditsErr = err
			return
		}

		creditsFont, err = creditsExtractArray(srdata.CreditsData, "credits_font")
		if err != nil {
			creditsErr = err
			return
		}
		need := creditsFontRows * creditsFontStride
		if len(creditsFont) < need {
			pad := make([]byte, need)
			copy(pad, creditsFont)
			creditsFont = pad
		}
	})
	return creditsErr
}

func creditsExtractArray(data []byte, name string) ([]byte, error) {
	idx := bytes.Index(data, []byte(name))
	if idx == -1 {
		return nil, fmt.Errorf("credits: array %q not found", name)
	}
	brace := bytes.IndexByte(data[idx:], '{')
	if brace == -1 {
		return nil, fmt.Errorf("credits: array %q missing '{'", name)
	}
	i := idx + brace + 1
	out := make([]byte, 0, 1024)
	for i < len(data) {
		c := data[i]
		if c == '}' {
			return out, nil
		}
		if c == '-' || (c >= '0' && c <= '9') {
			val, n := glenzParseNumber(data[i:])
			if n > 0 {
				out = append(out, byte(uint8(val)))
				i += n
				continue
			}
		}
		i++
	}
	return nil, fmt.Errorf("credits: array %q unterminated", name)
}
