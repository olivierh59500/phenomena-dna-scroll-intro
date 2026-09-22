package phenomena

import (
	"github.com/olivierh59500/democonstructionkit/scrolling"
	"math"
	"slices"
	"testing"
)

// This oracle retains the original stream arithmetic independently of DCK.
type sliceReference struct {
	msgIndex, sliceCount, scrollHead, pauseTime int
	pause                                       bool
	rotSpeed, scrollerRotation                  float64
	scrollChars                                 [240]scrolling.DNASlice
	sineOffsets                                 [240]float64
}

func (g *sliceReference) scrollMessage(speed int) {
	for i := 0; i < speed; i++ {
		ch := scrollMessage[g.msgIndex]
		isCtrl := ch == '^' || ch == '#' || ch == '&' || ch == '%'

		if isCtrl && g.sliceCount == 0 {
			// Handle control characters only when starting a new character
			switch ch {
			case '^':
				g.pause = true
				g.pauseTime = 275
				g.rotSpeed = -1
			case '&':
				g.pause = true
				g.pauseTime = 275
				g.rotSpeed = 1
			case '#':
				g.pause = true
				g.pauseTime = 250
				g.rotSpeed = -1
			case '%':
				g.pause = true
				g.pauseTime = 225
				g.rotSpeed = -1
			}
			g.msgIndex++
			if g.msgIndex >= len(scrollMessage) {
				g.msgIndex = 90
			}
			// Don't scroll this frame, the pause will take effect
		} else {
			// Regular scroll: shift left, add new slice
			g.shiftLeft()
			g.addSliceOfChar(ch, g.sliceCount)

			g.sliceCount++
			if g.sliceCount > 7 {
				g.sliceCount = 0
				g.msgIndex++
				if g.msgIndex >= len(scrollMessage) {
					g.msgIndex = 90
				}
			}
		}
	}
}

// shiftLeft advances the logical start of the circular scroller buffer.
func (g *sliceReference) shiftLeft() {
	g.scrollHead++
	if g.scrollHead == len(g.scrollChars) {
		g.scrollHead = 0
	}
}

// addSliceOfChar adds a single slice of a character to the end of the scroll
func (g *sliceReference) addSliceOfChar(ch byte, slice int) {
	previous := g.scrollHead + len(g.scrollChars) - 2
	if previous >= len(g.scrollChars) {
		previous -= len(g.scrollChars)
	}
	tail := g.scrollHead + len(g.scrollChars) - 1
	if tail >= len(g.scrollChars) {
		tail -= len(g.scrollChars)
	}
	glyph, ok := charToFontIndex(rune(ch))
	if !ok {
		glyph = 0
	}

	g.scrollChars[tail] = scrolling.DNASlice{
		Glyph: glyph,
		Frame: g.scrollChars[previous].Frame,
		Slice: slice,
	}
}

// renderNextFrames advances the animation frames with a DNA/twist effect
func (g *sliceReference) renderNextFrames(speed float64) {
	// Update the base rotation
	g.scrollerRotation += speed
	if g.scrollerRotation >= 30 {
		g.scrollerRotation -= 30
	}
	if g.scrollerRotation < 0 {
		g.scrollerRotation += 30
	}

	for i := range g.scrollChars {
		index := g.scrollHead + i
		if index >= len(g.scrollChars) {
			index -= len(g.scrollChars)
		}
		newFrame := g.scrollerRotation + g.sineOffsets[i]
		if newFrame >= 30 {
			newFrame -= 30
		} else if newFrame < 0 {
			newFrame += 30
		}
		g.scrollChars[index].Frame = int(newFrame)
	}
}

// drawScroller draws the 3D rotating text scroller

func TestSharedSliceTransportMatchesOriginalThroughControlsAndLoops(t *testing.T) {
	scene := NewGame()
	reference := &sliceReference{rotSpeed: .35, pauseTime: 250}
	for i := range reference.sineOffsets {
		reference.sineOffsets[i] = math.Sin(float64(i)*.05) * 15
	}
	count := len(scrollMessage)*8*3 + 2000
	for tick := 0; tick < count; tick++ {
		if !scene.pause {
			scene.scrollMessage(1)
		} else {
			scene.pauseTime--
			if scene.pauseTime == 0 {
				scene.pause = false
				scene.rotSpeed = .35
			}
		}
		if !reference.pause {
			reference.scrollMessage(1)
		} else {
			reference.pauseTime--
			if reference.pauseTime == 0 {
				reference.pause = false
				reference.rotSpeed = .35
			}
		}
		scene.renderNextFrames(scene.rotSpeed)
		reference.renderNextFrames(reference.rotSpeed)
		token, strip := scene.sliceStream.Cursor()
		if token != reference.msgIndex || strip != reference.sliceCount || scene.sliceStream.Head() != reference.scrollHead || scene.pause != reference.pause || scene.pauseTime != reference.pauseTime || scene.rotSpeed != reference.rotSpeed || scene.scrollerRotation != reference.scrollerRotation {
			t.Fatalf("transport or cue changed at tick %d", tick)
		}
		if tick%97 == 0 || scene.pause {
			if !slices.Equal(scene.sliceStream.Slices(), reference.scrollChars[:]) {
				t.Fatalf("strip history changed at tick %d", tick)
			}
		}
	}
}
