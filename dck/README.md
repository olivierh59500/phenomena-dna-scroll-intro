# DCK version

This directory contains the construction-kit version of phenomena-dna-scroll-intro. The original Go sources are preserved at their original paths (revision `87590875c1db88c82ce57594c01a5bc492f7311f`), with small asset accessors so both versions use the same embedded resources.

Run the original with `go run ./cmd/phenomena` and this version with `go run ./dck/cmd/phenomena` from the repository root.

The choreography and assets remain in this repository. Reusable rendering and
effects come from the published `github.com/olivierh59500/democonstructionkit`
module pinned in `go.mod`. Music is opened with `sound.Open`; DCK selects the decoder from the asset and
provides the configured stereo PCM format. The demo keeps its playback level and loop settings.

## Shared DNA scroll

`scrolling.NewDNAFrames` constructs the silver front, purple back and central raster from the shared gradient materials. `SliceProgram` owns the two-pixel circular transport, control-character pauses, rotation and loop point; `RecurrentRowWave` owns the strip baseline. `ScalarStages` drives the introduction and exit thresholds. Multiscreen uses the same DCK recipes. The artwork, message, music and layer order remain here.

For a new scroll, register `DNA.Mode()` in `scrolling.New(Config{Modes: ...})`. Each font supplies its own metrics; filmstrips adapt to glyph dimensions, including odd widths and proportional or mixed fonts. Rotation speed, phase, twist wavelength/amplitude, vertical waves and face colors are configurable. Use `{shape:dna}` or a timed mode sequence.

See the [DCK effect configuration guide](../../../lib/democonstructionkit/docs/EFFECT_OPTIONS.md) for the shared API and examples.
