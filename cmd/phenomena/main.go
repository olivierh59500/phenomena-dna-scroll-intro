package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	phenomena "phenomena-dna-scroll-intro"
)

func main() {
	ebiten.SetWindowSize(phenomena.ScreenWidth*2, phenomena.ScreenHeight*2)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowTitle("Phenomena - Enigma (Go/Ebitengine Port)")
	ebiten.SetVsyncEnabled(true)
	ebiten.SetScreenClearedEveryFrame(false)

	game := phenomena.NewGame()
	err := ebiten.RunGame(newDrawOnUpdateGame(game))
	game.Cleanup()
	if err != nil {
		log.Fatal(err)
	}
}
