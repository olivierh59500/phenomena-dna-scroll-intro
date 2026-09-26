// Package phenomena implements the Phenomena DNA scroll intro remake.
package phenomena

import (
	"bytes"
	"fmt"
	"github.com/olivierh59500/democonstructionkit/presets"
	"image"
	"image/color"
	originalassets "phenomena-dna-scroll-intro"

	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/palette"
	"github.com/olivierh59500/democonstructionkit/scrolling"
	"github.com/olivierh59500/democonstructionkit/sound"
	"github.com/olivierh59500/democonstructionkit/timeline"

	_ "image/png"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	audio "github.com/olivierh59500/democonstructionkit/sound/output"
)

const (
	screenWidth  = 640
	screenHeight = 480
	sampleRate   = 48000

	// ScreenWidth and ScreenHeight are the demo's logical dimensions.
	ScreenWidth  = screenWidth
	ScreenHeight = screenHeight
)

// Embed all assets
var rasterbarData = originalassets.
	DCKAssetRasterbarData()

var fontData = originalassets.
	DCKAssetFontData()

var logoData = originalassets.
	DCKAssetLogoData()

var photonData = originalassets.
	DCKAssetPhotonData()

var musicData = originalassets.DCKAssetMusicData()

// Character set for the scroller - must match the font.png layout
// Font has 45 characters: " ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!'?/,.-@"
const charset = presets.PhenomenaAlphabet

// Main scrolling message
const scrollMessage = `           THIS IS IMPOSSIBLE!            WHAT IS?               THIS IS!!!                    ...SO, ANOTHER DEMO FROM PHENOMENA HAS REACHED YOU...    THIS TIME WITH CODING BY                PHOTON!                ^  RASTA MUSIC BY                    FIREFOX!                &    AND SUPER GFX BY                       TERMINATOR               #  ...SO, SLAYER! HOW DO YOU LIKE @MY@ SCROLLER?  IT'S MUCH IMPOSSIBLER THAN YOURS!    ...   SO DE SO!          DOES ANYONE HAVE A PROGRAM CALLED 'PAGE RENDER 3D'? THEN CONTACT OUR NEW GFX ARTIST AT          0492-41027               % AND ASK FOR MIKAEL. NEWS NEWS NEWS NEWS   !!! LOOK OUT FOR PHENOMENA'S NEW DISK MAG CALLED ' TRANSMISSION ' ! ! ! ! IT'S A MAG ESPECIALLY MADE FOR ALL YOU CODERS OUT THERE, COMPLETE WITH CODER / DEMO / CRACK TOP-TEN,ARTICLES ABOUT CODING / CRACKING, AND SOURCES, ETC,ETC...         HERE'S MY TOP-FIVE DEMO GROUPS 1. SCOOPEX  -SLAYER IS WORKING HARD AND HIS M.H. DEMO IS STILL UNBEATEN-  ...  2. CRYPTOBURNERS  -NICE MD 2 BUT SLOOOW VECTORS-  ... 3. RSI/PARADOX  -NICE DEMOS LATELY, EXCEPT FOR THE 'FOLLOW ME' CRAP-  ...  4. KEFRENS  -ALL YOUR LATEST DEMOS HAVE BEEN PROFESSIONAL!-  ...  5. THE LINK  -ALWAYS COOL IDEAS,GIVE US SOME MORE-  ...  OF COURSE, PHENOMENA IS EXCLUDED FROM THIS LIST...        NOW OVER TO SOME INTERNAL GREETS...  @     BIG 2A-FINISH YOUR DEMO AND BUY AN A500!   @   CORE-GET YOUR HANDS ON A WORKING AMIGA!   @   DANKO-GET BUSY!   @   KLUTTAS O SPIRIT-WAKE UP FROM YOUR COMA!!!!   @   RAVE-SAME TO YOU!       ...     AND NOW, TIME FOR SOME OTHER GREETS... THEY GO TO --- CONAN/TPL-MAKE A GOOD DEMO AND JOIN ANOTHER GROUP!   @   KALLE BALLE/TSL - EVER THOUGHT ABOUT CHANGING YOUR NAME????   @   HAVOK/ECSTASY-JOIN US! I'M JUST A PHONECALL AWAY - 0381-11344 @   MAHONEY/NS-TRY TAKING SOME IDEAS FROM NT 1.2!  @   UNCLE TOM/RAZOR-STOP DRAWING AND DO SOME MUSIC @   SLAYER/SCX-AND ALL OTHER GOOD CODERS-CALL ME FOR SOME COOL TECH-TALK    0381-11344   ZEUS/ADEPT-GOOD LUCK AND CODE HARD!       ---     NOW I DON'T HAVE VERY MUCH ELSE TO SAY, EXCEPT....                    BYE!             @@@@@@@@@@@@@                `

// DemoState represents the current state of the demo
type DemoState int

const (
	StateTextPage1          DemoState = presets.PhenomenaTextPage1
	StateTextPage2          DemoState = presets.PhenomenaTextPage2
	StateShowLogo           DemoState = presets.PhenomenaShowLogo
	StateShowUpperRasterbar DemoState = presets.PhenomenaShowUpperRaster
	StateShowLowerRasterbar DemoState = presets.PhenomenaShowLowerRaster
	StateDropPhoton         DemoState = presets.PhenomenaDropPhoton
	StatePhotonFadeToRed    DemoState = presets.PhenomenaPhotonFade
	StateMainDemo           DemoState = presets.PhenomenaMain
	StateHideLogo           DemoState = presets.PhenomenaHideLogo
	StateHideLowerRasterbar DemoState = presets.PhenomenaHideLowerRaster
	StateHideUpperRasterbar DemoState = presets.PhenomenaHideUpperRaster
	StateEnd                DemoState = presets.PhenomenaEnd
)

// Game represents the main game state
type Game struct {
	dnaFrames *scrolling.DNAFrames
	// Demo state
	state       DemoState
	initialized bool
	finished    bool
	director    *timeline.ScalarStages

	// Images
	imgRasterbar       *ebiten.Image
	imgFont            *ebiten.Image
	imgFontInverted    *ebiten.Image
	imgLogo            *ebiten.Image
	imgLogoMask        *ebiten.Image
	imgPhoton          *ebiten.Image
	imgPhotonMask      *ebiten.Image
	imgMiddle          *ebiten.Image
	rasterGradient     *ebiten.Image
	imgTextPage1       *ebiten.Image
	imgTextPage2       *ebiten.Image
	fontGlyphs         [len(charset)]*ebiten.Image
	fontGlyphsInverted [len(charset)]*ebiten.Image

	// Animation canvases
	cnvFrames *ebiten.Image // All character animation frames

	// Animation variables
	t              float64
	percent        float64
	blackRectWidth float64
	blackRectShow  bool
	rasterbarY     float64
	direction      float64

	// Scroller data
	sliceProgram *scrolling.SliceProgram
	rowWave      *motion.RecurrentRowWave
	dnaDraw      scrolling.DNADrawConfig
	photonMotion *motion.GravityBounce
	hueMotion    *motion.WrapBank

	// Audio
	audioContext *audio.Context
	audioPlayer  *audio.Player
	musicStream  *sound.Stream
	audioReady   bool
	audioVolume  float64

	touchIDs []ebiten.TouchID
}

// NewGame creates a new game instance
func NewGame() *Game {
	g := &Game{
		state:          StateTextPage1,
		blackRectWidth: 640,
		blackRectShow:  true,
		rasterbarY:     -40,
		direction:      1,
		audioVolume:    1,
	}
	programConfig, err := presets.PhenomenaDNAProgram(scrollMessage, charToFontIndex)
	if err != nil {
		panic(err)
	}
	g.sliceProgram, err = scrolling.NewSliceProgram(programConfig)
	if err != nil {
		panic(err)
	}
	wave, err := motion.NewRecurrentRowWave(presets.PhenomenaDNARows())
	if err != nil {
		panic(err)
	}
	g.rowWave = wave
	g.photonMotion, err = motion.NewGravityBounce(presets.PhenomenaPhotonBounce())
	if err != nil {
		panic(err)
	}
	g.hueMotion, err = motion.NewWrapBank(presets.PhenomenaPhotonHueCycle())
	if err != nil {
		panic(err)
	}
	g.director, err = timeline.NewScalarStages(presets.PhenomenaPresentation(int(StateTextPage1)))
	if err != nil {
		panic(err)
	}
	g.dnaDraw = scrolling.DNADrawConfig{SliceWidth: 2, ScaleX: 2, ScaleY: 1.5, OriginY: 156, Y: wave.At}
	return g
}

// loadImages loads all image assets
func (g *Game) loadImages() error {
	var err error

	// Load rasterbar
	img, _, err := image.Decode(bytes.NewReader(rasterbarData))
	if err != nil {
		return fmt.Errorf("failed to load rasterbar: %w", err)
	}
	g.imgRasterbar = ebiten.NewImageFromImage(img)

	// Load font
	img, _, err = image.Decode(bytes.NewReader(fontData))
	if err != nil {
		return fmt.Errorf("failed to load font: %w", err)
	}
	g.imgFont = ebiten.NewImageFromImage(img)
	g.imgFontInverted, err = composite.NewInvertedImage(img)
	if err != nil {
		return fmt.Errorf("failed to invert font: %w", err)
	}
	glyphs, err := scrolling.GridImages(g.imgFont, image.Pt(16, 26), len(g.fontGlyphs), len(g.fontGlyphs))
	if err != nil {
		return err
	}
	inverted, err := scrolling.GridImages(g.imgFontInverted, image.Pt(16, 26), len(g.fontGlyphsInverted), len(g.fontGlyphsInverted))
	if err != nil {
		return err
	}
	copy(g.fontGlyphs[:], glyphs)
	copy(g.fontGlyphsInverted[:], inverted)

	// Load logo
	img, _, err = image.Decode(bytes.NewReader(logoData))
	if err != nil {
		return fmt.Errorf("failed to load logo: %w", err)
	}
	g.imgLogo = ebiten.NewImageFromImage(img)
	g.imgLogoMask, err = composite.NewWhiteSilhouette(img)
	if err != nil {
		return fmt.Errorf("failed to build logo silhouette: %w", err)
	}

	// Load photon
	img, _, err = image.Decode(bytes.NewReader(photonData))
	if err != nil {
		return fmt.Errorf("failed to load photon: %w", err)
	}
	g.imgPhoton = ebiten.NewImageFromImage(img)
	g.imgPhotonMask, err = composite.NewWhiteSilhouette(img)
	if err != nil {
		return fmt.Errorf("failed to build photon silhouette: %w", err)
	}

	g.imgMiddle = ebiten.NewImage(screenWidth, 300)
	g.imgMiddle.Fill(color.RGBA{0x00, 0x01, 0x11, 0xFF})
	g.rasterGradient, err = composite.NewGradientImage(palette.GradientConfig{
		Width: screenWidth, Height: 12, Stops: presets.PhenomenaRasterStops(),
	})
	if err != nil {
		return fmt.Errorf("failed to build raster gradient: %w", err)
	}

	return nil
}

// charToFontIndex converts a character to its position in the font bitmap
var charToFontIndex = func() func(rune) (int, bool) {
	lookup, err := presets.TileLookup("phenomena-dna-scroll-intro", false)
	if err != nil {
		panic(err)
	}
	return lookup
}()

// makeIntroText creates intro text pages
func (g *Game) makeIntroText(mode string, backColor color.Color, texts []struct {
	Y    int
	Text string
}) *ebiten.Image {
	img := ebiten.NewImage(640, 480)

	if backColor != nil {
		img.Fill(backColor)
	}

	for _, t := range texts {
		x := 48
		for _, ch := range t.Text {
			idx, found := charToFontIndex(ch)

			if found {
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Scale(2, 2)
				op.GeoM.Translate(float64(x), float64(t.Y))

				glyph := g.fontGlyphs[idx]
				if mode == "xor" {
					glyph = g.fontGlyphsInverted[idx]
				}

				img.DrawImage(glyph, op)
			}
			x += 32
		}
	}

	return img
}

// initTextPages initializes the intro text pages
func (g *Game) initTextPages() {
	// Text page 1
	texts1 := []struct {
		Y    int
		Text string
	}{
		{18, "   FOR HOT VHS"},
		{75, "  AND SOFTWARE"},
		{133, "SWAPPING, CONTACT"},
		{219, " THE PUNISHER "},
		{291, "      AT..."},
	}
	g.imgTextPage1 = g.makeIntroText("xor", color.Black, texts1)

	// Text page 2
	texts2 := []struct {
		Y    int
		Text string
	}{
		{78, "    PHENOMENA"},
		{158, "   SKALDEV. 69"},
		{238, "  16142 BROMMA"},
		{334, "     SWEDEN!"},
	}
	g.imgTextPage2 = g.makeIntroText("source-over", nil, texts2)
}

// initCharacterFrames creates all animation frames for the 3D rotating characters
func (g *Game) initCharacterFrames() error {
	gradient := func(height int, stops []palette.GradientStop) (*ebiten.Image, error) {
		return composite.NewGradientImage(palette.GradientConfig{Width: 480, Height: height, Stops: stops})
	}
	core, err := gradient(9, presets.PhenomenaCoreStops())
	if err != nil {
		return err
	}
	defer core.Deallocate()
	front, err := gradient(33, presets.PhenomenaFrontStops())
	if err != nil {
		return err
	}
	defer front.Deallocate()
	back, err := gradient(33, presets.PhenomenaBackStops())
	if err != nil {
		return err
	}
	defer back.Deallocate()
	g.dnaFrames, err = scrolling.NewDNAFrames(g.fontGlyphs[:], scrolling.DNAFrameConfig{Frames: 30, Height: 33, Step: 2.25, Front: front, Back: back, Core: core, CoreY: 12})
	if err != nil {
		return err
	}
	g.cnvFrames = g.dnaFrames.Image
	return nil
}

// initAudio opens the audio device only after Ebitengine's game loop is live.
func (g *Game) initAudio() error {
	g.audioContext = audio.NewContext(sampleRate)

	var err error

	// Let DCK choose and configure the music decoder.
	g.musicStream, err = sound.Open("music.ym", musicData, sound.Options{SampleRate: sampleRate, Loop: true, PCMFormat: sound.PCM16, Gain: 0.5})
	if err != nil {
		return fmt.Errorf("failed to open music: %w", err)
	}

	// Create audio player
	g.audioPlayer, err = g.audioContext.NewPlayer(g.musicStream)
	if err != nil {
		if closeErr := g.musicStream.Close(); closeErr != nil {
			log.Printf("Failed to close music stream: %v", closeErr)
		}
		g.musicStream = nil
		return fmt.Errorf("failed to create audio player: %w", err)
	}

	g.audioPlayer.SetVolume(g.audioVolume)
	g.audioPlayer.Play()
	return nil
}

// Init initializes the game
func (g *Game) Init() error {
	if g.initialized {
		return nil
	}

	// Load all images
	if err := g.loadImages(); err != nil {
		return err
	}

	// Initialize text pages
	g.initTextPages()

	// Initialize character animation frames
	if err := g.initCharacterFrames(); err != nil {
		return err
	}

	// Bring message to the start
	if err := g.sliceProgram.Warmup(320, 1); err != nil {
		return err
	}

	g.initialized = true
	return nil
}

func (g *Game) drawScroller(screen *ebiten.Image) {
	if err := g.rowWave.Begin(g.t); err != nil {
		panic(err)
	}
	g.sliceProgram.Draw(screen, g.dnaFrames, g.dnaDraw)
}

func appendTexturedQuad(vertices []ebiten.Vertex, indices []uint16, dstX, dstY, dstWidth, dstHeight, srcX, srcY, srcWidth, srcHeight float32) ([]ebiten.Vertex, []uint16) {
	base := uint16(len(vertices))
	vertices = append(vertices,
		ebiten.Vertex{DstX: dstX, DstY: dstY, SrcX: srcX, SrcY: srcY, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
		ebiten.Vertex{DstX: dstX + dstWidth, DstY: dstY, SrcX: srcX + srcWidth, SrcY: srcY, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
		ebiten.Vertex{DstX: dstX, DstY: dstY + dstHeight, SrcX: srcX, SrcY: srcY + srcHeight, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
		ebiten.Vertex{DstX: dstX + dstWidth, DstY: dstY + dstHeight, SrcX: srcX + srcWidth, SrcY: srcY + srcHeight, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
	)
	indices = append(indices, base, base+1, base+2, base+1, base+2, base+3)
	return vertices, indices
}

// Update updates the game state
func (g *Game) Update() error {
	if !g.initialized {
		if err := g.Init(); err != nil {
			return err
		}
	}
	if !g.audioReady {
		g.audioReady = true
		if err := g.initAudio(); err != nil {
			log.Printf("Failed to load music: %v", err)
		}
	}

	// Handle volume control
	if g.audioPlayer != nil {
		previousVolume := g.audioVolume
		if ebiten.IsKeyPressed(ebiten.KeyUp) {
			g.audioVolume += 0.01
			if g.audioVolume > 1 {
				g.audioVolume = 1
			}
		}
		if ebiten.IsKeyPressed(ebiten.KeyDown) {
			g.audioVolume -= 0.01
			if g.audioVolume < 0 {
				g.audioVolume = 0
			}
		}
		if g.audioVolume != previousVolume {
			g.audioPlayer.SetVolume(g.audioVolume)
		}
	}
	g.touchIDs = ebiten.AppendTouchIDs(g.touchIDs[:0])

	// Stage-specific simulations run before the reusable threshold director.
	switch g.state {
	case StateDropPhoton:
		if g.photonMotion.Step() {
			g.director.Signal("photon-landed")
		}
	case StateMainDemo:
		g.hueMotion.Step()
		if err := g.sliceProgram.Step(); err != nil {
			return err
		}
		g.t += 0.30
		if g.blackRectShow {
			g.blackRectWidth -= 8
			if g.blackRectWidth < 0 {
				g.blackRectWidth = 0
				g.blackRectShow = false
			}
		}
		if g.finished {
			g.director.Signal("finish")
		}
	case StateEnd:
		return nil
	}
	previousStage := g.state
	g.director.Step()
	pose := g.director.State()
	g.state = DemoState(pose.Stage)
	if previousStage == StateTextPage1 {
		g.rasterbarY = pose.Value
	} else {
		g.percent = pose.Value
	}
	g.direction = pose.Direction

	// A desktop click or Android touch starts the outro.
	if (ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) || len(g.touchIDs) > 0) && g.state == StateMainDemo {
		g.finished = true
	}

	return nil
}

// Draw draws the game
func (g *Game) Draw(screen *ebiten.Image) {
	if !g.initialized {
		return
	}

	switch g.state {
	case StateTextPage1:
		screen.Fill(color.Black)
		var op ebiten.DrawImageOptions
		op.GeoM.Translate(0, g.rasterbarY)
		screen.DrawImage(g.imgRasterbar, &op)
		screen.DrawImage(g.imgTextPage1, nil)

	case StateTextPage2:
		screen.Fill(color.Black)
		var op ebiten.DrawImageOptions
		brightness := g.percent / 100.0
		if brightness > 1 {
			brightness = 2 - brightness
		}
		op.ColorScale.Scale(float32(brightness), float32(brightness), float32(brightness), 1)
		screen.DrawImage(g.imgTextPage2, &op)

	case StateShowLogo:
		screen.Fill(color.RGBA{0x00, 0x01, 0x11, 0xFF})

		if g.percent <= 100 {
			var op ebiten.DrawImageOptions
			brightness := g.percent / 100.0
			op.ColorScale.Scale(float32(brightness), float32(brightness), float32(brightness), 1)
			screen.DrawImage(g.imgLogo, &op)
		} else {
			screen.DrawImage(g.imgLogo, nil)
			var op ebiten.DrawImageOptions
			alpha := (200 - g.percent) / 100.0
			op.ColorScale.ScaleAlpha(float32(alpha))
			screen.DrawImage(g.imgLogoMask, &op)
		}

	case StateShowUpperRasterbar, StateShowLowerRasterbar, StateDropPhoton, StatePhotonFadeToRed:
		screen.Fill(color.Black)
		drawImageAt(screen, g.imgMiddle, 0, 130)

		screen.DrawImage(g.imgLogo, nil)

		if g.state >= StateShowUpperRasterbar {
			alpha := 1.0
			if g.state == StateShowUpperRasterbar {
				alpha = g.percent / 100.0
			}
			var op ebiten.DrawImageOptions
			op.ColorScale.ScaleAlpha(float32(alpha))
			op.GeoM.Translate(0, 129)
			screen.DrawImage(g.rasterGradient, &op)
		}

		if g.state >= StateShowLowerRasterbar {
			alpha := 1.0
			if g.state == StateShowLowerRasterbar {
				alpha = g.percent / 100.0
			}
			var op ebiten.DrawImageOptions
			op.ColorScale.ScaleAlpha(float32(alpha))
			op.GeoM.Translate(0, 430)
			screen.DrawImage(g.rasterGradient, &op)
		}

		if g.state >= StateDropPhoton {
			if g.state == StatePhotonFadeToRed {
				var op ebiten.DrawImageOptions
				op.GeoM.Translate(285, 445)
				lightness := g.percent / 100.0
				op.ColorScale.Scale(float32(lightness), float32(lightness*0.5), float32(lightness*0.5), 1)
				screen.DrawImage(g.imgPhoton, &op)
			} else {
				var op ebiten.DrawImageOptions
				op.GeoM.Translate(285, g.photonMotion.Position())
				screen.DrawImage(g.imgPhoton, &op)
			}
		}

	case StateMainDemo:
		screen.Fill(color.Black)
		drawImageAt(screen, g.imgMiddle, 0, 130)

		screen.DrawImage(g.imgLogo, nil)

		var op ebiten.DrawImageOptions
		op.GeoM.Translate(0, 129)
		screen.DrawImage(g.rasterGradient, &op)

		op.GeoM.Reset()
		op.GeoM.Translate(0, 430)
		screen.DrawImage(g.rasterGradient, &op)

		op = ebiten.DrawImageOptions{}
		hue := g.hueMotion.At(0) / 360.0
		r, g2, b := palette.HSLToRGB(hue, 1.0, 0.5)
		op.ColorScale.Scale(float32(r), float32(g2), float32(b), 1)
		op.GeoM.Translate(285, 445)
		screen.DrawImage(g.imgPhotonMask, &op)

		g.drawScroller(screen)

		if g.blackRectShow {
			vector.FillRect(screen, 0, 375, float32(g.blackRectWidth), 55, color.RGBA{0x00, 0x01, 0x11, 0xFF}, false)
		}

	case StateHideLogo, StateHideLowerRasterbar, StateHideUpperRasterbar:
		screen.Fill(color.Black)

		if g.state == StateHideLogo {
			if g.direction > 0 {
				screen.DrawImage(g.imgLogo, nil)
				var op ebiten.DrawImageOptions
				alpha := g.percent / 100.0
				op.ColorScale.ScaleAlpha(float32(alpha))
				screen.DrawImage(g.imgLogoMask, &op)
			} else {
				var op ebiten.DrawImageOptions
				alpha := g.percent / 100.0
				op.ColorScale.ScaleAlpha(float32(alpha))
				screen.DrawImage(g.imgLogoMask, &op)
			}
		}

		if g.state == StateHideLowerRasterbar {
			var op ebiten.DrawImageOptions
			op.ColorScale.ScaleAlpha(float32(g.percent / 100.0))
			op.GeoM.Translate(0, 430)
			screen.DrawImage(g.rasterGradient, &op)
		}

		if g.state == StateHideUpperRasterbar {
			var op ebiten.DrawImageOptions
			op.ColorScale.ScaleAlpha(float32(g.percent / 100.0))
			op.GeoM.Translate(0, 129)
			screen.DrawImage(g.rasterGradient, &op)
		}

	case StateEnd:
		screen.Fill(color.Black)
	}
}

func drawImageAt(dst, src *ebiten.Image, x, y int) {
	var op ebiten.DrawImageOptions
	op.GeoM.Translate(float64(x), float64(y))
	dst.DrawImage(src, &op)
}

// Layout returns the screen dimensions
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

// Cleanup cleans up resources
func (g *Game) Cleanup() {
	if g.dnaFrames != nil {
		g.dnaFrames.Close()
	}
	if g.audioPlayer != nil {
		_ = g.audioPlayer.Close()
		g.audioPlayer = nil
	}
	if g.musicStream != nil {
		_ = g.musicStream.Close()
		g.musicStream = nil
	}
}
