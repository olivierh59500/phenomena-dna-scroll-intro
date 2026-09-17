package phenomena

import (
	"encoding/binary"
	"testing"
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
	for i := range game.scrollChars {
		game.scrollChars[i].glyph = uint8(i)
	}

	game.shiftLeft()
	game.addSliceOfChar('A', 3)

	if game.scrollHead != 1 {
		t.Fatalf("scroll head = %d, want 1", game.scrollHead)
	}
	if got := game.scrollChars[game.scrollHead].glyph; got != 1 {
		t.Fatalf("first logical glyph = %d, want 1", got)
	}
	tail := (game.scrollHead + len(game.scrollChars) - 1) % len(game.scrollChars)
	if got, want := game.scrollChars[tail].glyph, uint8(1); got != want {
		t.Fatalf("tail glyph = %d, want %d", got, want)
	}
	if got, want := game.scrollChars[tail].slice, uint8(3); got != want {
		t.Fatalf("tail slice = %d, want %d", got, want)
	}
}

func TestScrollerUpdateDoesNotAllocate(t *testing.T) {
	game := NewGame()
	allocations := testing.AllocsPerRun(1000, func() {
		game.scrollMessage(1)
		game.renderNextFrames(game.rotSpeed)
	})
	if allocations != 0 {
		t.Fatalf("scroller update allocations = %v, want 0", allocations)
	}
}

func TestYMPlayerReadProducesStereoWithoutAllocating(t *testing.T) {
	player, err := NewYMPlayer(musicData, sampleRate, true)
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
		game.scrollMessage(1)
		game.renderNextFrames(game.rotSpeed)
	}
}

func BenchmarkYMPlayerRead4096(b *testing.B) {
	player, err := NewYMPlayer(musicData, sampleRate, true)
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
