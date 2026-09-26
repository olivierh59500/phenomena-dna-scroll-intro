package phenomena

import (
	"encoding/binary"
	"testing"

	"github.com/olivierh59500/democonstructionkit/sound"
)

func TestCharsetIndex(t *testing.T) {
	for want, ch := range charset {
		got, ok := charToFontIndex(ch)
		if !ok || got != want {
			t.Fatalf("charToFontIndex(%q) = (%d, %v), want (%d, true)", ch, got, ok, want)
		}
	}

	if got, ok := charToFontIndex('a'); !ok || got != 1 {
		t.Fatalf("lowercase mapping = (%d, %v), want (1, true)", got, ok)
	}
	if _, ok := charToFontIndex('^'); ok {
		t.Fatal("control character unexpectedly mapped to a glyph")
	}
}

func TestCircularScrollerShift(t *testing.T) {
	game := NewGame()
	game.sliceProgram.Insert(3)
	if game.sliceProgram.Stream().Head() != 3 {
		t.Fatalf("head = %d", game.sliceProgram.Stream().Head())
	}
	token, strip := game.sliceProgram.Stream().Cursor()
	if token != 0 || strip != 3 {
		t.Fatalf("cursor = %d,%d", token, strip)
	}
	tail := (game.sliceProgram.Stream().Head() + len(game.sliceProgram.Stream().Slices()) - 1) % len(game.sliceProgram.Stream().Slices())
	if got := game.sliceProgram.Stream().Slices()[tail]; got.Glyph != 0 || got.Slice != 2 {
		t.Fatalf("tail = %v", got)
	}
}

func TestScrollerUpdateDoesNotAllocate(t *testing.T) {
	game := NewGame()
	allocations := testing.AllocsPerRun(1000, func() {
		_ = game.sliceProgram.Step()
	})
	if allocations != 0 {
		t.Fatalf("scroller update allocations = %v, want 0", allocations)
	}
}

func TestMusicStreamReadProducesStereoWithoutAllocating(t *testing.T) {
	player, err := sound.Open("music.ym", musicData, sound.Options{SampleRate: sampleRate, Loop: true, PCMFormat: sound.PCM16, Gain: 0.5})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := player.Close(); err != nil {
			t.Errorf("Close: %v", err)
		}
	}()

	buffer := make([]byte, 4096*4)
	read := func() {
		n, err := player.Read(buffer)
		if err != nil {
			t.Fatalf("Read: %v", err)
		}
		if n != len(buffer) {
			t.Fatalf("Read bytes = %d, want %d", n, len(buffer))
		}
	}

	read()
	for i := 0; i < len(buffer); i += 4 {
		left := binary.LittleEndian.Uint16(buffer[i : i+2])
		right := binary.LittleEndian.Uint16(buffer[i+2 : i+4])
		if left != right {
			t.Fatalf("frame %d is not mono duplicated to stereo: %d != %d", i/4, left, right)
		}
	}

	if allocations := testing.AllocsPerRun(20, read); allocations != 0 {
		t.Fatalf("Read allocations = %v, want 0", allocations)
	}
}

func BenchmarkScrollerUpdate(b *testing.B) {
	game := NewGame()
	b.ReportAllocs()
	for b.Loop() {
		_ = game.sliceProgram.Step()
	}
}

func BenchmarkMusicStreamRead4096(b *testing.B) {
	player, err := sound.Open("music.ym", musicData, sound.Options{SampleRate: sampleRate, Loop: true, PCMFormat: sound.PCM16, Gain: 0.5})
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() {
		if err := player.Close(); err != nil {
			b.Errorf("Close: %v", err)
		}
	})

	buffer := make([]byte, 4096*4)
	b.ReportAllocs()
	b.SetBytes(int64(len(buffer)))
	for b.Loop() {
		if _, err := player.Read(buffer); err != nil {
			b.Fatal(err)
		}
	}
}
