# DCK version

This directory contains the construction-kit version of go-secondreality. The original Go sources are preserved at their original paths (revision `ea713e594d41f9f70f1cb741ab682fdb0c03fef6`), with small asset accessors so both versions use the same embedded resources.

Run the original with `go run ./cmd/secondreality` and this version with `go run ./dck/cmd/secondreality` from the repository root.

The choreography and assets remain in this repository. Reusable rendering and
effects come from the published `github.com/olivierh59500/democonstructionkit`
module pinned in `go.mod`. Go resolves its dependencies automatically, including
`github.com/olivierh59500/ym-player v1.0.0` in the DCK module graph.
Second Reality retains its original ST3 music synchronization.
