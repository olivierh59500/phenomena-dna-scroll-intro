# DCK version

This directory contains the construction-kit version of phenomena-dna-scroll-intro. The original Go sources are preserved at their original paths (revision `87590875c1db88c82ce57594c01a5bc492f7311f`), with small asset accessors so both versions use the same embedded resources.

Run the original with `go run ./cmd/phenomena` and this version with `go run ./dck/cmd/phenomena` from the repository root.

The choreography and assets remain in this repository. Reusable rendering and
effects come from the published `github.com/olivierh59500/democonstructionkit`
module pinned in `go.mod`. Go downloads the dependencies automatically, including
`github.com/olivierh59500/ym-player v1.0.0` for YM playback.

## Shared DNA scroll

`scrolling.NewDNAFrames` constructs the original silver front, purple back and central raster. `DrawSlices` renders the circular scroll buffer. Multiscreen uses the same implementation. The surrounding intro, original text controls and music remain here.

For a new scroll, register `DNA.Mode()` in `scrolling.New(Config{Modes: ...})`. Each font supplies its own metrics; filmstrips adapt to glyph dimensions, including odd widths and proportional or mixed fonts. Rotation speed, phase, twist wavelength/amplitude, vertical waves and face colors are configurable. Use `{shape:dna}` or a timed mode sequence.

See the [DCK effect configuration guide](../../../lib/democonstructionkit/docs/EFFECT_OPTIONS.md) for the shared API and examples.
