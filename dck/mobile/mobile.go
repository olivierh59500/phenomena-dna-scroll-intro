// Package mobile exposes the demo to ebitenmobile.
package mobile

import (
	enginemobile "github.com/hajimehoshi/ebiten/v2/mobile"
	phenomena "phenomena-dna-scroll-intro/dck"
)

func init() {
	enginemobile.SetGame(phenomena.NewGame())
}

// Dummy forces gomobile to include this package in the Android binding.
func Dummy() {}
