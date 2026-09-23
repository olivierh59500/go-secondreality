
## Optional DCK version

The original implementation remains at its original paths. Run it with `go run ./cmd/secondreality`.

The construction-kit version is in [dck/](dck/README.md). Run `go run ./dck/cmd/secondreality` from this directory. Both versions share the original assets.

The DCK version opens the bundled soundtrack through `sound.Open` with a track
index and starting order. DCK selects the timing-compatible replay inside
`go-zikmu`; the demo reads audible order, row and frame markers through DCK.
The local DCK copy of the replay engine has been removed. Regression tests
compare PCM and synchronization markers with the original implementation.
