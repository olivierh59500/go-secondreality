package parts

import "github.com/olivierh59500/democonstructionkit/indexed"

// compileScatterMaps shares the sparse surface projection used by both the
// mountain text and water text. Each keeps its own map data and blend policy.
func compileScatterMaps(streams [3][]byte, sourceSize, destinationSize int) ([3]*indexed.ScatterMap, error) {
	var maps [3]*indexed.ScatterMap
	for i, stream := range streams {
		compiled, err := indexed.DecodeScatterMap16(stream, sourceSize, destinationSize)
		if err != nil {
			return maps, err
		}
		maps[i] = compiled
	}
	return maps, nil
}
