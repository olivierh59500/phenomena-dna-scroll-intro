package phenomena

import "testing"

func legacyAtlasIndex(ch rune) (int, bool) {
	if ch >= 'a' && ch <= 'z' {
		ch -= 'a' - 'A'
	}
	switch {
	case ch == ' ':
		return 0, true
	case ch >= 'A' && ch <= 'Z':
		return int(ch-'A') + 1, true
	case ch >= '0' && ch <= '9':
		return int(ch-'0') + 27, true
	case ch == '!':
		return 37, true
	case ch == '\'':
		return 38, true
	case ch == '?':
		return 39, true
	case ch == '/':
		return 40, true
	case ch == ',':
		return 41, true
	case ch == '.':
		return 42, true
	case ch == '-':
		return 43, true
	case ch == '@':
		return 44, true
	default:
		return 0, false
	}
}
func TestSharedAtlasIndicesMatchOriginalAlphabet(t *testing.T) {
	for r := rune(0); r < 256; r++ {
		got, ok := charToFontIndex(r)
		want, found := legacyAtlasIndex(r)
		if got != want || ok != found {
			t.Fatalf("rune %U: got (%d,%t), want (%d,%t)", r, got, ok, want, found)
		}
	}
}
