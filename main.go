package main

import (
	"bytes"
	_ "embed"
	"image"
	_ "image/png"

	"github.com/DemmyDemon/starflight/porthole"
	"github.com/hajimehoshi/ebiten/v2"
)

var (
	//go:embed resources/viewport.png
	viewportResource []byte
)

func main() {
	ebiten.SetWindowSize(960, 540)
	ebiten.SetWindowTitle("WARP FIVE")

	viewportImg, _, err := image.Decode(bytes.NewReader(viewportResource))
	if err != nil {
		panic(err)
	}

	// ebiten.SetCursorMode(ebiten.CursorModeHidden)

	f := porthole.New(ebiten.NewImageFromImage(viewportImg))

	if err := ebiten.RunGame(f); err != nil {
		panic(err)
	}

}
