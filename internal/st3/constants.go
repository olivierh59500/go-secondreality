package st3

const (
	c2Freq = 8363

	soundcardGUS   = 0
	soundcardSBPro = 1

	pattSep = 254
	pattEnd = 255

	initialDitherSeed = 0x12345000

	mixBufSamples = 4096
)

func clamp[T ~int | ~int8 | ~int16 | ~int32 | ~int64](v, low, high T) T {
	if v < low {
		return low
	}
	if v > high {
		return high
	}
	return v
}

func clamp16(v int32) int32 {
	if int16(v) != int16(v) {
		return 0x7FFF ^ (v >> 31)
	}
	return v
}
