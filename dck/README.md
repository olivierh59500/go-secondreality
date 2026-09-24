# DCK version

This directory contains the construction-kit version of go-secondreality. The original Go sources are preserved at their original paths (revision `ea713e594d41f9f70f1cb741ab682fdb0c03fef6`), with small asset accessors so both versions use the same embedded resources.

Run the original with `go run ./cmd/secondreality` and this version with `go run ./dck/cmd/secondreality` from the repository root.

The choreography and assets remain in this repository. Reusable rendering and
effects come from the published `github.com/olivierh59500/democonstructionkit`
module pinned in `go.mod`. Go resolves its dependencies automatically, including
`github.com/olivierh59500/ym-player v1.0.0` in the DCK module graph.
Second Reality retains its original ST3 music synchronization.

The Rotozoomer uses DCK's `indexed.Rotozoom256` sampler. It preserves the
original palette-indexed 256 × 256 textures, 16-bit address wrapping and
fixed-point row stepping. Three indexed reference poses and an isolated
43-second recording matched the previous implementation exactly across all
2,580 decoded frames. The original demo package remains unchanged.
