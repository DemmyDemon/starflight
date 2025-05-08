package porthole

import (
	"image/color"
	"math/rand/v2"
	"os"
	"runtime"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	WIDTH    = 480
	HEIGHT   = 270
	DEPTH    = 20
	MINDEPTH = 5
	STARS    = 75
	SLOWNESS = 45.0
)

var (
	ColSpace = color.RGBA{0, 0, 7, 255}
)

type Porthole struct {
	foreground *ebiten.Image
	fgOptions  *ebiten.DrawImageOptions
	Stars      []Star
	Warp       bool
	Run        bool
	WarpFactor float64
	TargetWarp float64
	WarpRamp   float64
	Counter    int
	lastClick  int
}

type Star struct {
	X     float64
	Y     float64
	Z     float64
	Color color.Color
}

func New(foreground *ebiten.Image) ebiten.Game {

	stars := make([]Star, STARS)

	for i := 0; i < STARS; i++ {
		stars[i] = NewStar()
	}

	f := &Porthole{
		foreground: foreground,
		fgOptions:  &ebiten.DrawImageOptions{},
		Stars:      stars,
		Warp:       true,
		Run:        true,
		WarpFactor: 0,
		TargetWarp: 9.9,
		WarpRamp:   0.1,
	}

	return f
}

func NewStar() Star {
	star := Star{
		X: rand.Float64() * WIDTH,
		Y: rand.Float64() * HEIGHT,
		Z: rand.Float64()*DEPTH + MINDEPTH,
	}

	color := PickStarColor(star.Z / (DEPTH + MINDEPTH))
	star.Color = color

	return star
}

func PickStarColor(depth float64) color.Color {

	r := rand.Uint32() % 256
	g := rand.Uint32() % 256
	b := rand.Uint32() % 256

	// Calculate luminance (Cargo cult FTW)
	luminance := (0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(b))

	scale := 255 / luminance
	finalR := float64(r) * scale
	finalG := float64(g) * scale
	finalB := float64(b) * scale

	base := color.RGBA{clamp(finalR), clamp(finalG), clamp(finalB), 255}
	return AlphaPrecalc(base, depth)
}

func (p *Porthole) Update() error {
	p.Counter++

	if ebiten.IsKeyPressed(ebiten.KeyQ) {
		if runtime.GOARCH != "wasm" { // Because this will exit the program, but not exit fullscreen or clear the canvas...
			os.Exit(0)
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsMouseButtonJustPressed(ebiten.MouseButton2) {
		p.Warp = !p.Warp
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		p.Run = !p.Run
	}

	if p.Run || inpututil.IsKeyJustPressed(ebiten.KeyN) {
		if p.Warp && p.WarpFactor < p.TargetWarp {
			p.WarpFactor += p.WarpRamp
			if p.WarpFactor > p.TargetWarp {
				p.WarpFactor = p.TargetWarp
			}
		}

		if (!p.Warp || p.WarpFactor > p.TargetWarp) && p.WarpFactor > 0 {
			p.WarpFactor -= p.WarpRamp
			if p.WarpFactor < 0 {
				p.WarpFactor = 0
			}
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButton0) {
		//if ebiten.IsFocused() {
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButton0) {
		if p.Counter-p.lastClick < 15 {
			ebiten.SetFullscreen(!ebiten.IsFullscreen())
		}
		p.lastClick = p.Counter
	}

	if p.Run || inpututil.IsKeyJustPressed(ebiten.KeyN) {
		p.Move()
	}

	return nil
}

func (p *Porthole) Move() {
	for i := 0; i < STARS; i++ {
		p.Stars[i].X += ((p.WarpFactor * p.Stars[i].Z) / SLOWNESS)
		if p.Stars[i].X > WIDTH {
			p.Stars[i].X = 0
			p.Stars[i].Y = rand.Float64() * HEIGHT
			p.Stars[i].Z = rand.Float64()*DEPTH + MINDEPTH
			p.Stars[i].Color = PickStarColor(p.Stars[i].Z / (DEPTH + MINDEPTH))
		}
	}
}

func (p *Porthole) Draw(screen *ebiten.Image) {

	screen.Fill(ColSpace)

	if p.WarpFactor == 0 {
		p.DrawStill(screen)
	} else {
		p.DrawWarp(screen)
	}

	if p.foreground != nil {
		screen.DrawImage(p.foreground, p.fgOptions)
	}
}

func (p *Porthole) DrawStill(screen *ebiten.Image) {
	if p.WarpFactor == 0 {
		for _, star := range p.Stars {
			screen.Set(int(star.X), int(star.Y), star.Color)
		}
	}
}

func (p *Porthole) DrawWarp(screen *ebiten.Image) {

	stepSize := 1 / float64(p.WarpFactor)

	// ebitenutil.DebugPrint(screen, fmt.Sprintf("WarpFactor %d, stepSize %0.5f", p.WarpFactor, stepSize))

	for _, star := range p.Stars {
		screen.Set(int(star.X), int(star.Y), star.Color)
		for i := int(p.WarpFactor); i > 0; i-- {
			color := AlphaPrecalc(star.Color, 1-(stepSize*float64(i)))
			screen.Set(int(star.X)-i, int(star.Y), color)
		}
	}

}

func AlphaPrecalc(col color.Color, alpha float64) color.Color {
	r, g, b, a := col.RGBA()

	newR := massage(r) * alpha
	newG := massage(g) * alpha
	newB := massage(b) * alpha
	newA := massage(a) * alpha

	return color.RGBA{
		R: clamp(newR),
		G: clamp(newG),
		B: clamp(newB),
		A: clamp(newA),
	}
}

func massage(value uint32) float64 {
	fval := float64(value) / 65535
	return fval
}

func clamp(value float64) uint8 {
	if value > 255 {
		return 255
	}
	return uint8(value * 255)
}

func (p *Porthole) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return WIDTH, HEIGHT
}
