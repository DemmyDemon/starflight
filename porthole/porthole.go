package porthole

import (
	"image/color"
	"math/rand/v2"
	"os"
	"runtime"
	"slices"

	"github.com/DemmyDemon/starflight/shaders"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	WIDTH     = 480
	HEIGHT    = 270
	DEPTH     = 20
	MINDEPTH  = 5
	STARS     = 75
	SLOWNESS  = 45.0
	USESHADER = true
)

var (
	ColSpace = color.RGBA{0, 0, 7, 255}
)

type Porthole struct {
	starsImage *ebiten.Image
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

	shaders.MustLoad()
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

	f.SortStars()

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
	luma := uint8(255 * depth)
	cb := uint8(rand.Uint() % 256)
	cr := uint8(rand.Uint() % 256)
	r, g, b := color.YCbCrToRGB(luma, cb, cr)
	return color.RGBA{r, g, b, 255}
}

func cmpDepth(a Star, b Star) int {
	if a.Z > b.Z {
		return 1
	}
	return -1
}

func (p *Porthole) SortStars() {
	slices.SortFunc(p.Stars, cmpDepth)
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
	needSort := false
	for i := 0; i < STARS; i++ {
		p.Stars[i].X += ((p.WarpFactor * p.Stars[i].Z) / SLOWNESS)
		if p.Stars[i].X > WIDTH {
			p.Stars[i].X = 0
			p.Stars[i].Y = rand.Float64() * HEIGHT
			p.Stars[i].Z = rand.Float64()*DEPTH + MINDEPTH
			p.Stars[i].Color = PickStarColor(p.Stars[i].Z / (DEPTH + MINDEPTH))
			needSort = true
		}
	}
	if needSort {
		p.SortStars()
	}
}

func (p *Porthole) Draw(screen *ebiten.Image) {

	screen.Fill(ColSpace)

	if USESHADER {
		if p.starsImage == nil {
			p.starsImage = ebiten.NewImage(WIDTH, HEIGHT)
		}
		p.starsImage.Clear()
		p.DrawStill(p.starsImage)
		screen.DrawRectShader(
			WIDTH, HEIGHT, shaders.Get("starsmudge"),
			&ebiten.DrawRectShaderOptions{
				Uniforms: map[string]any{
					"WarpFactor": p.WarpFactor,
				},
				Images: [4]*ebiten.Image{
					p.starsImage,
				},
			},
		)
		screen.DrawImage(p.starsImage, p.fgOptions)
	} else if p.WarpFactor == 0 {
		p.DrawStill(screen)
	} else {
		p.DrawWarp(screen)
	}

	if p.foreground != nil {
		screen.DrawImage(p.foreground, p.fgOptions)
	}
}

func (p *Porthole) DrawStill(img *ebiten.Image) {
	for _, star := range p.Stars {
		img.Set(int(star.X), int(star.Y), star.Color)
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
