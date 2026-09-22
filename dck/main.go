// Package phenomena implements the Phenomena DNA scroll intro remake.
package phenomena

import originalassets "phenomena-dna-scroll-intro"

import (
	"bytes"

	"fmt"
	"github.com/olivierh59500/democonstructionkit/scrolling"
	"image"
	"image/color"
	_ "image/png"
	"io"
	"log"
	"math"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	audio "github.com/olivierh59500/democonstructionkit/sound/output"
	"github.com/olivierh59500/ym-player/pkg/stsound"
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

var musicData = originalassets.

	// YMPlayer wraps the YM player for Ebiten audio
	DCKAssetMusicData()

type YMPlayer struct {
	player *stsound.StSound
	buffer []int16
	mutex  sync.Mutex
	loop   bool
}

// NewYMPlayer creates a new YM player instance
func NewYMPlayer(data []byte, sampleRate int, loop bool) (*YMPlayer, error) {
	player := stsound.CreateWithRate(sampleRate)

	if err := player.LoadMemory(data); err != nil {
		player.Destroy()
		return nil, fmt.Errorf("failed to load YM data: %w", err)
	}

	player.SetLoopMode(loop)

	return &YMPlayer{
		player: player,
		buffer: make([]int16, 4096),
		loop:   loop,
	}, nil
}

// Read implements io.Reader for audio streaming
func (y *YMPlayer) Read(p []byte) (n int, err error) {
	y.mutex.Lock()
	defer y.mutex.Unlock()

	samplesNeeded := len(p) / 4
	processed := 0
	for processed < samplesNeeded {
		chunkSize := samplesNeeded - processed
		if chunkSize > len(y.buffer) {
			chunkSize = len(y.buffer)
		}

		if !y.player.Compute(y.buffer[:chunkSize], chunkSize) {
			if !y.loop {
				clear(p[processed*4 : samplesNeeded*4])
				err = io.EOF
				break
			}
		}

		for i := 0; i < chunkSize; i++ {
			sample := y.buffer[i] / 2
			offset := (processed + i) * 4
			p[offset] = byte(sample)
			p[offset+1] = byte(sample >> 8)
			p[offset+2] = byte(sample)
			p[offset+3] = byte(sample >> 8)
		}

		processed += chunkSize
	}

	return samplesNeeded * 4, err
}

// Close releases resources
func (y *YMPlayer) Close() error {
	y.mutex.Lock()
	defer y.mutex.Unlock()

	if y.player != nil {
		y.player.Destroy()
		y.player = nil
	}
	return nil
}

// Color definitions for gradients
var (
	gdcRasterBar = []GradientStop{
		{color.RGBA{0x44, 0x00, 0x44, 0xFF}, 0.0},
		{color.RGBA{0xFF, 0xDD, 0xFF, 0xFF}, 0.5},
		{color.RGBA{0x11, 0x11, 0x44, 0xFF}, 1.0},
	}
	gdcRedBar = []GradientStop{
		{color.RGBA{0x00, 0x00, 0x00, 0xFF}, 0.0},
		{color.RGBA{0xFF, 0x33, 0x00, 0xFF}, 0.5},
		{color.RGBA{0x00, 0x00, 0x00, 0xFF}, 1.0},
	}
	gdcSilverBar = []GradientStop{
		{color.RGBA{0x55, 0x55, 0x55, 0xFF}, 0.0},
		{color.RGBA{0xFF, 0xFF, 0xFF, 0xFF}, 0.5},
		{color.RGBA{0x55, 0x55, 0x55, 0xFF}, 1.0},
	}
	gdcPurpleBar = []GradientStop{
		{color.RGBA{0x34, 0x22, 0x55, 0xFF}, 0.0},
		{color.RGBA{0x60, 0x4E, 0x98, 0xFF}, 0.5},
		{color.RGBA{0x34, 0x22, 0x55, 0xFF}, 1.0},
	}
)

// GradientStop represents a color stop in a gradient
type GradientStop struct {
	Color  color.RGBA
	Offset float64
}

// Character set for the scroller - must match the font.png layout
// Font has 45 characters: " ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!'?/,.-@"
const charset = " ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!'?/,.-@"

var waveSinStep, waveCosStep = math.Sincos(1.0 / 36.0)

// Main scrolling message
const scrollMessage = `           THIS IS IMPOSSIBLE!            WHAT IS?               THIS IS!!!                    ...SO, ANOTHER DEMO FROM PHENOMENA HAS REACHED YOU...    THIS TIME WITH CODING BY                PHOTON!                ^  RASTA MUSIC BY                    FIREFOX!                &    AND SUPER GFX BY                       TERMINATOR               #  ...SO, SLAYER! HOW DO YOU LIKE @MY@ SCROLLER?  IT'S MUCH IMPOSSIBLER THAN YOURS!    ...   SO DE SO!          DOES ANYONE HAVE A PROGRAM CALLED 'PAGE RENDER 3D'? THEN CONTACT OUR NEW GFX ARTIST AT          0492-41027               % AND ASK FOR MIKAEL. NEWS NEWS NEWS NEWS   !!! LOOK OUT FOR PHENOMENA'S NEW DISK MAG CALLED ' TRANSMISSION ' ! ! ! ! IT'S A MAG ESPECIALLY MADE FOR ALL YOU CODERS OUT THERE, COMPLETE WITH CODER / DEMO / CRACK TOP-TEN,ARTICLES ABOUT CODING / CRACKING, AND SOURCES, ETC,ETC...         HERE'S MY TOP-FIVE DEMO GROUPS 1. SCOOPEX  -SLAYER IS WORKING HARD AND HIS M.H. DEMO IS STILL UNBEATEN-  ...  2. CRYPTOBURNERS  -NICE MD 2 BUT SLOOOW VECTORS-  ... 3. RSI/PARADOX  -NICE DEMOS LATELY, EXCEPT FOR THE 'FOLLOW ME' CRAP-  ...  4. KEFRENS  -ALL YOUR LATEST DEMOS HAVE BEEN PROFESSIONAL!-  ...  5. THE LINK  -ALWAYS COOL IDEAS,GIVE US SOME MORE-  ...  OF COURSE, PHENOMENA IS EXCLUDED FROM THIS LIST...        NOW OVER TO SOME INTERNAL GREETS...  @     BIG 2A-FINISH YOUR DEMO AND BUY AN A500!   @   CORE-GET YOUR HANDS ON A WORKING AMIGA!   @   DANKO-GET BUSY!   @   KLUTTAS O SPIRIT-WAKE UP FROM YOUR COMA!!!!   @   RAVE-SAME TO YOU!       ...     AND NOW, TIME FOR SOME OTHER GREETS... THEY GO TO --- CONAN/TPL-MAKE A GOOD DEMO AND JOIN ANOTHER GROUP!   @   KALLE BALLE/TSL - EVER THOUGHT ABOUT CHANGING YOUR NAME????   @   HAVOK/ECSTASY-JOIN US! I'M JUST A PHONECALL AWAY - 0381-11344 @   MAHONEY/NS-TRY TAKING SOME IDEAS FROM NT 1.2!  @   UNCLE TOM/RAZOR-STOP DRAWING AND DO SOME MUSIC @   SLAYER/SCX-AND ALL OTHER GOOD CODERS-CALL ME FOR SOME COOL TECH-TALK    0381-11344   ZEUS/ADEPT-GOOD LUCK AND CODE HARD!       ---     NOW I DON'T HAVE VERY MUCH ELSE TO SAY, EXCEPT....                    BYE!             @@@@@@@@@@@@@                `

// DemoState represents the current state of the demo
type DemoState int

const (
	StateTextPage1 DemoState = iota
	StateTextPage2
	StateShowLogo
	StateShowUpperRasterbar
	StateShowLowerRasterbar
	StateDropPhoton
	StatePhotonFadeToRed
	StateMainDemo
	StateHideLogo
	StateHideLowerRasterbar
	StateHideUpperRasterbar
	StateEnd
)

// Game represents the main game state
type Game struct {
	dnaFrames *scrolling.DNAFrames
	// Demo state
	state       DemoState
	initialized bool
	finished    bool

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
	t                float64
	pause            bool
	pauseTime        int
	scrollSpeed      int
	rotSpeed         float64
	color            float64
	percent          float64
	blackRectWidth   float64
	blackRectShow    bool
	photonY          float64
	photonGravity    float64
	photonBounce     float64
	rasterbarY       float64
	direction        float64
	scrollerRotation float64

	// Scroller data
	sliceStream *scrolling.SliceStream
	sineOffsets [240]float64

	// Audio
	audioContext *audio.Context
	audioPlayer  *audio.Player
	ymPlayer     *YMPlayer
	audioReady   bool
	audioVolume  float64

	touchIDs []ebiten.TouchID
}

// NewGame creates a new game instance
func NewGame() *Game {
	g := &Game{
		state:            StateTextPage1,
		pauseTime:        250,
		rotSpeed:         0.35,
		scrollSpeed:      1,
		blackRectWidth:   640,
		blackRectShow:    true,
		photonY:          184,
		photonBounce:     -9.50,
		rasterbarY:       -40,
		direction:        1,
		scrollerRotation: 0,
		audioVolume:      1,
	}

	for i := range g.sineOffsets {
		g.sineOffsets[i] = math.Sin(float64(i)*0.05) * 15
	}

	g.initSliceStream()
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
	g.imgFontInverted = newInvertedImage(img)
	for glyph := range g.fontGlyphs {
		sx := glyph * 16
		g.fontGlyphs[glyph] = g.imgFont.SubImage(image.Rect(sx, 0, sx+16, 26)).(*ebiten.Image)
		g.fontGlyphsInverted[glyph] = g.imgFontInverted.SubImage(image.Rect(sx, 0, sx+16, 26)).(*ebiten.Image)
	}

	// Load logo
	img, _, err = image.Decode(bytes.NewReader(logoData))
	if err != nil {
		return fmt.Errorf("failed to load logo: %w", err)
	}
	g.imgLogo = ebiten.NewImageFromImage(img)
	g.imgLogoMask = newWhiteAlphaMask(img)

	// Load photon
	img, _, err = image.Decode(bytes.NewReader(photonData))
	if err != nil {
		return fmt.Errorf("failed to load photon: %w", err)
	}
	g.imgPhoton = ebiten.NewImageFromImage(img)
	g.imgPhotonMask = newWhiteAlphaMask(img)

	g.imgMiddle = ebiten.NewImage(screenWidth, 300)
	g.imgMiddle.Fill(color.RGBA{0x00, 0x01, 0x11, 0xFF})
	g.rasterGradient = createGradient(screenWidth, 12, gdcRasterBar)

	return nil
}

// createGradient creates a horizontal gradient image
func createGradient(width, height int, stops []GradientStop) *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		t := float64(y) / float64(height-1)

		// Find the two stops to interpolate between
		var c color.RGBA
		for i := 0; i < len(stops)-1; i++ {
			if t >= stops[i].Offset && t <= stops[i+1].Offset {
				// Interpolate between stops[i] and stops[i+1]
				localT := (t - stops[i].Offset) / (stops[i+1].Offset - stops[i].Offset)
				c = lerpColor(stops[i].Color, stops[i+1].Color, localT)
				break
			}
		}

		row := img.Pix[y*img.Stride : y*img.Stride+width*4]
		for x := 0; x < len(row); x += 4 {
			row[x] = c.R
			row[x+1] = c.G
			row[x+2] = c.B
			row[x+3] = c.A
		}
	}

	return ebiten.NewImageFromImage(img)
}

// lerpColor interpolates between two colors
func lerpColor(c1, c2 color.RGBA, t float64) color.RGBA {
	r := uint8(float64(c1.R)*(1-t) + float64(c2.R)*t)
	g := uint8(float64(c1.G)*(1-t) + float64(c2.G)*t)
	b := uint8(float64(c1.B)*(1-t) + float64(c2.B)*t)
	a := uint8(float64(c1.A)*(1-t) + float64(c2.A)*t)

	return color.RGBA{r, g, b, a}
}

func newWhiteAlphaMask(source image.Image) *ebiten.Image {
	bounds := source.Bounds()
	mask := image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		row := mask.Pix[(y-bounds.Min.Y)*mask.Stride:]
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			_, _, _, alpha := source.At(x, y).RGBA()
			offset := (x - bounds.Min.X) * 4
			row[offset] = 0xff
			row[offset+1] = 0xff
			row[offset+2] = 0xff
			row[offset+3] = uint8(alpha >> 8)
		}
	}
	return ebiten.NewImageFromImage(mask)
}

func newInvertedImage(source image.Image) *ebiten.Image {
	bounds := source.Bounds()
	inverted := image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		row := inverted.Pix[(y-bounds.Min.Y)*inverted.Stride:]
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pixel := color.NRGBAModel.Convert(source.At(x, y)).(color.NRGBA)
			offset := (x - bounds.Min.X) * 4
			row[offset] = 0xff - pixel.R
			row[offset+1] = 0xff - pixel.G
			row[offset+2] = 0xff - pixel.B
			row[offset+3] = pixel.A
		}
	}
	return ebiten.NewImageFromImage(inverted)
}

// charToFontIndex converts a character to its position in the font bitmap
func charToFontIndex(ch rune) (int, bool) {
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
func (g *Game) initCharacterFrames() {
	core, front, back := createGradient(480, 9, gdcRedBar), createGradient(480, 33, gdcSilverBar), createGradient(480, 33, gdcPurpleBar)
	defer core.Deallocate()
	defer front.Deallocate()
	defer back.Deallocate()
	var err error
	g.dnaFrames, err = scrolling.NewDNAFrames(g.fontGlyphs[:], scrolling.DNAFrameConfig{Frames: 30, Height: 33, Step: 2.25, Front: front, Back: back, Core: core, CoreY: 12})
	if err != nil {
		panic(err)
	}
	g.cnvFrames = g.dnaFrames.Image
}

// initAudio opens the audio device only after Ebitengine's game loop is live.
func (g *Game) initAudio() error {
	g.audioContext = audio.NewContext(sampleRate)

	var err error

	// Create YM player
	g.ymPlayer, err = NewYMPlayer(musicData, sampleRate, true)
	if err != nil {
		return fmt.Errorf("failed to create YM player: %w", err)
	}

	// Create audio player
	g.audioPlayer, err = g.audioContext.NewPlayer(g.ymPlayer)
	if err != nil {
		if closeErr := g.ymPlayer.Close(); closeErr != nil {
			log.Printf("Failed to close YM player: %v", closeErr)
		}
		g.ymPlayer = nil
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
	g.initCharacterFrames()

	// Bring message to the start
	for i := 0; i < 320; i++ {
		g.scrollMessage(1)
		g.renderNextFrames(g.rotSpeed)
	}

	g.initialized = true
	return nil
}

// scrollMessage advances the scroll text
func (g *Game) initSliceStream() {
	tokens := make([]scrolling.SliceToken, 0, len(scrollMessage))
	for _, ch := range scrollMessage {
		if ch == '^' || ch == '#' || ch == '&' || ch == '%' {
			tokens = append(tokens, scrolling.SliceToken{Control: string(ch)})
			continue
		}
		glyph, ok := charToFontIndex(ch)
		if !ok {
			glyph = 0
		}
		tokens = append(tokens, scrolling.SliceToken{Glyph: glyph, Width: 16})
	}
	var err error
	g.sliceStream, err = scrolling.NewSliceStream(scrolling.SliceStreamConfig{Tokens: tokens, Capacity: len(g.sineOffsets), SliceWidth: 2, Repeat: true, LoopStart: 90})
	if err != nil {
		panic(err)
	}
}

func (g *Game) scrollMessage(speed int) {
	g.sliceStream.Step(speed, func(event scrolling.SliceControl) bool {
		g.pause = true
		switch event.Name {
		case "^":
			g.pauseTime = 275
			g.rotSpeed = -1
		case "&":
			g.pauseTime = 275
			g.rotSpeed = 1
		case "#":
			g.pauseTime = 250
			g.rotSpeed = -1
		case "%":
			g.pauseTime = 225
			g.rotSpeed = -1
		}
		return false
	})
}

func (g *Game) renderNextFrames(speed float64) {
	g.scrollerRotation += speed
	if g.scrollerRotation >= 30 {
		g.scrollerRotation -= 30
	}
	if g.scrollerRotation < 0 {
		g.scrollerRotation += 30
	}
	if err := g.sliceStream.SetFrames(g.scrollerRotation, g.sineOffsets[:], 30); err != nil {
		panic(err)
	}
}

func (g *Game) drawScroller(screen *ebiten.Image) {
	t2 := g.t
	waveSin, waveCos := math.Sincos(5*10.50 + g.t/6)
	g.dnaFrames.DrawSlices(screen, g.sliceStream.Slices(), g.sliceStream.Head(), scrolling.DNADrawConfig{
		SliceWidth: 2, ScaleX: 2, ScaleY: 1.5, OriginY: 156, Y: func(i int) float64 {
			y := 80.0
			if t2 > 5*50-float64(i)*.0033 {
				y = 80 * waveCos
			}
			t2 += 1.0 / 6.0
			waveSin, waveCos = waveSin*waveCosStep+waveCos*waveSinStep, waveCos*waveCosStep-waveSin*waveSinStep
			return 67 + y
		}})
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

	// Handle different demo states
	switch g.state {
	case StateTextPage1:
		g.rasterbarY += 1.5
		if g.rasterbarY >= 340 {
			g.rasterbarY = 0
			g.percent = 0
			g.state = StateTextPage2
		}

	case StateTextPage2:
		g.percent += g.direction * 1
		if g.percent > 200 {
			g.direction = -1
			g.percent = 100
		}
		if g.percent <= 0 && g.direction == -1 {
			g.percent = 0
			g.state = StateShowLogo
		}

	case StateShowLogo:
		g.percent += 4
		if g.percent >= 200 {
			g.percent = 0
			g.state = StateShowUpperRasterbar
		}

	case StateShowUpperRasterbar:
		g.percent += 4
		if g.percent >= 100 {
			g.percent = 0
			g.state = StateShowLowerRasterbar
		}

	case StateShowLowerRasterbar:
		g.percent += 4
		if g.percent >= 100 {
			g.percent = 0
			g.state = StateDropPhoton
		}

	case StateDropPhoton:
		g.photonGravity += 0.30
		g.photonY += g.photonGravity
		if g.photonY > 445 {
			g.photonGravity = g.photonBounce
			g.photonBounce *= 0.70
		}
		if g.photonBounce >= -0.70 {
			g.percent = 100
			g.state = StatePhotonFadeToRed
		}

	case StatePhotonFadeToRed:
		g.percent -= 4
		if g.percent < 50 {
			g.percent = 0
			g.state = StateMainDemo
		}

	case StateMainDemo:
		// Update animations
		g.color += 1.0 / 3.0
		if g.color > 360 {
			g.color = 0
		}

		// Update scroll
		if !g.pause {
			g.scrollMessage(g.scrollSpeed)
		} else {
			g.pauseTime--
			if g.pauseTime == 0 {
				g.pause = false
				g.rotSpeed = 0.35
				g.scrollSpeed = 1
			}
		}

		// Update timer
		g.t += 0.30

		// Update character frames
		g.renderNextFrames(g.rotSpeed)

		// Handle black rect reveal
		if g.blackRectShow {
			g.blackRectWidth -= 8
			if g.blackRectWidth < 0 {
				g.blackRectWidth = 0
				g.blackRectShow = false
			}
		}

		// Check for finish
		if g.finished {
			g.percent = 50
			g.direction = 1
			g.state = StateHideLogo
		}

	case StateHideLogo:
		if g.direction > 0 {
			g.percent += g.direction * 4
			if g.percent > 100 {
				g.direction = -1
			}
		} else {
			g.percent += g.direction * 4
			if g.percent < 0 {
				g.percent = 100
				g.direction = -1
				g.state = StateHideLowerRasterbar
			}
		}

	case StateHideLowerRasterbar:
		g.percent += g.direction * 4
		if g.percent < 0 {
			g.direction = -1
			g.percent = 100
			g.state = StateHideUpperRasterbar
		}

	case StateHideUpperRasterbar:
		g.percent += g.direction * 4
		if g.percent < 0 {
			g.state = StateEnd
		}

	case StateEnd:
		// Demo finished
		return nil
	}

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
				op.GeoM.Translate(285, g.photonY)
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
		hue := g.color / 360.0
		r, g2, b := hslToRGB(hue, 1.0, 0.5)
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

// hslToRGB converts HSL color values to RGB
func hslToRGB(h, s, l float64) (float64, float64, float64) {
	var r, g, b float64

	if s == 0 {
		r, g, b = l, l, l
	} else {
		var hue2rgb = func(p, q, t float64) float64 {
			if t < 0 {
				t += 1
			}
			if t > 1 {
				t -= 1
			}
			if t < 1.0/6.0 {
				return p + (q-p)*6*t
			}
			if t < 1.0/2.0 {
				return q
			}
			if t < 2.0/3.0 {
				return p + (q-p)*(2.0/3.0-t)*6
			}
			return p
		}

		var q float64
		if l < 0.5 {
			q = l * (1 + s)
		} else {
			q = l + s - l*s
		}
		p := 2*l - q
		r = hue2rgb(p, q, h+1.0/3.0)
		g = hue2rgb(p, q, h)
		b = hue2rgb(p, q, h-1.0/3.0)
	}

	return r, g, b
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
	if g.ymPlayer != nil {
		_ = g.ymPlayer.Close()
		g.ymPlayer = nil
	}
}
