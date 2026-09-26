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
	state        DemoState
	initialized  bool
	finished     bool
	director     *timeline.ScalarStages
	presentation *composite.ScalarStagePainter

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
	imgTextPage1       *scrolling.BitmapPage
	imgTextPage2       *scrolling.BitmapPage
	fontGlyphs         [len(charset)]*ebiten.Image
	fontGlyphsInverted [len(charset)]*ebiten.Image

	// Animation canvases
	cnvFrames *ebiten.Image // All character animation frames

	// Animation variables
	t              float64
	blackRectWidth float64
	blackRectShow  bool

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

// initTextPages initializes the intro text pages
func (g *Game) initTextPages() error {
	// Text page 1
	texts1 := []scrolling.BitmapPageLine{
		presets.PhenomenaIntroLine(18, "   FOR HOT VHS"),
		presets.PhenomenaIntroLine(75, "  AND SOFTWARE"),
		presets.PhenomenaIntroLine(133, "SWAPPING, CONTACT"),
		presets.PhenomenaIntroLine(219, " THE PUNISHER "),
		presets.PhenomenaIntroLine(291, "      AT..."),
	}
	var err error
	g.imgTextPage1, err = scrolling.NewBitmapPage(presets.PhenomenaIntroPage(g.fontGlyphsInverted[:], texts1, color.Black))
	if err != nil {
		return err
	}

	// Text page 2
	texts2 := []scrolling.BitmapPageLine{
		presets.PhenomenaIntroLine(78, "    PHENOMENA"),
		presets.PhenomenaIntroLine(158, "   SKALDEV. 69"),
		presets.PhenomenaIntroLine(238, "  16142 BROMMA"),
		presets.PhenomenaIntroLine(334, "     SWEDEN!"),
	}
	g.imgTextPage2, err = scrolling.NewBitmapPage(presets.PhenomenaIntroPage(g.fontGlyphs[:], texts2, nil))
	return err
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
	if err := g.initTextPages(); err != nil {
		return err
	}
	presentation, err := composite.NewScalarStagePainter(presets.PhenomenaStageMaterials(g.director, presets.PhenomenaStageImages{
		RasterBar: g.imgRasterbar, Page1: g.imgTextPage1.Image(), Page2: g.imgTextPage2.Image(),
		Logo: g.imgLogo, LogoMask: g.imgLogoMask, Middle: g.imgMiddle,
		Raster: g.rasterGradient, Photon: g.imgPhoton, PhotonMask: g.imgPhotonMask,
	}))
	if err != nil {
		return err
	}
	g.presentation = presentation

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
	g.director.Step()
	pose := g.director.State()
	g.state = DemoState(pose.Stage)

	// A desktop click or Android touch starts the outro.
	if (ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) || len(g.touchIDs) > 0) && g.state == StateMainDemo {
		g.finished = true
	}

	return nil
}

// Draw composes shared stage materials before the live DNA scroller and mask.
func (g *Game) Draw(screen *ebiten.Image) {
	if !g.initialized {
		return
	}
	secondary := g.photonMotion.Position()
	if g.state == StateMainDemo {
		secondary = g.hueMotion.At(0) / 360.0
	}
	g.presentation.Draw(screen, secondary)
	if g.state != StateMainDemo {
		return
	}

	g.drawScroller(screen)
	if g.blackRectShow {
		vector.FillRect(screen, 0, 375, float32(g.blackRectWidth), 55, color.RGBA{0x00, 0x01, 0x11, 0xFF}, false)
	}
}

// Layout returns the screen dimensions
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

// Cleanup cleans up resources
func (g *Game) Cleanup() {
	if g.imgTextPage1 != nil {
		g.imgTextPage1.Close()
	}
	if g.imgTextPage2 != nil {
		g.imgTextPage2.Close()
	}
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
