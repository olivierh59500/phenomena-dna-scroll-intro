// Package phenomena implements the Phenomena DNA scroll intro remake.
package phenomena

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/scrolling"
	"image"
	"image/color"
	_ "image/png"
	"io"
	"log"
	"math"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/vector"
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
//
//go:embed assets/rasterbar.png
var rasterbarData []byte

//go:embed assets/font.png
var fontData []byte

//go:embed assets/logo.png
var logoData []byte

//go:embed assets/photon.png
var photonData []byte

//go:embed assets/music.ym
var musicData []byte

// YMPlayer wraps the YM player for Ebiten audio
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

// ScrollChar represents a character in the 3D scroller
type ScrollChar struct {
	glyph uint8
	frame uint8
	slice uint8
}

// Game represents the main game state
type Game struct {
	scrollRenderer *scrolling.Scrolling
	scrollBatch    *composite.QuadBatch
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
	msgIndex         int
	sliceCount       int
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
	scrollChars    [240]ScrollChar
	scrollHead     int
	scrollVertices []ebiten.Vertex
	scrollIndices  []uint16
	sineOffsets    [240]float64

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
		scrollVertices:   make([]ebiten.Vertex, 0, 240*4),
		scrollIndices:    make([]uint16, 0, 240*6),
		audioVolume:      1,
	}

	for i := range g.sineOffsets {
		g.sineOffsets[i] = math.Sin(float64(i)*0.05) * 15
	}

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
	// Create gradient bars - IMPORTANT: Red bar must be 480px wide to cover all frames
	cnvRedBar := createGradient(480, 9, gdcRedBar)
	cnvSilverBar := createGradient(480, 33, gdcSilverBar)
	cnvPurpleBar := createGradient(480, 33, gdcPurpleBar)
	defer cnvRedBar.Deallocate()
	defer cnvSilverBar.Deallocate()
	defer cnvPurpleBar.Deallocate()

	// Create font canvases for front and back
	cnvFont := ebiten.NewImage(len(charset)*16, 33)
	cnvFont2 := ebiten.NewImage(len(charset)*16, 33)
	defer cnvFont.Deallocate()
	defer cnvFont2.Deallocate()

	// Clear canvases
	cnvFont.Fill(color.RGBA{0, 0, 0, 0})
	cnvFont2.Fill(color.RGBA{0, 0, 0, 0})

	// Render front-side chars (inclined) - IMPORTANT: draw by 2-pixel slices like original
	for chr := 0; chr < len(charset); chr++ {
		// Draw character slice by slice (8 slices of 2 pixels each)
		for x := 0; x < 16; x += 2 {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(chr*16+x), 7) // Center vertically in 33px

			// Draw 2-pixel wide slice
			sx := chr*16 + x
			subImg := g.imgFont.SubImage(image.Rect(sx, 0, sx+2, 26)).(*ebiten.Image)
			cnvFont.DrawImage(subImg, op)
		}
	}

	// Render back-side chars (rot 180, flip horizontal)
	for chr := 0; chr < len(charset); chr++ {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(-1, -1)
		op.GeoM.Translate(float64((chr+1)*16), 26) // Adjusted for font height

		sx := chr * 16
		subImg := cnvFont.SubImage(image.Rect(sx, 0, sx+16, 33)).(*ebiten.Image)
		cnvFont2.DrawImage(subImg, op)
	}

	// Create animation frames canvas
	g.cnvFrames = ebiten.NewImage(480, len(charset)*33)

	// Generate frames for each character
	for charIndex := 0; charIndex < len(charset); charIndex++ {
		// Each character has 30 frames of animation (480 pixels / 16 pixels per frame)

		// Silver chars (front side) - character appears from bottom, rotates to top
		cnvSilverChars := ebiten.NewImage(480, 33)
		posX := 0
		for posY := 33.0; posY > -33; posY -= 2.25 { // 30 steps
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(posX), posY)

			sx := charIndex * 16
			subImg := cnvFont.SubImage(image.Rect(sx, 0, sx+16, 33)).(*ebiten.Image)
			cnvSilverChars.DrawImage(subImg, op)
			posX += 16
		}

		// Purple chars (back side) - character rotated 180 degrees
		cnvPurpleChars := ebiten.NewImage(480, 33)
		posX = 0
		// First half: character rises from 0 to 33
		for posY := 0.0; posY < 33; posY += 2.25 { // 15 steps
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(posX), posY)

			sx := charIndex * 16
			subImg := cnvFont2.SubImage(image.Rect(sx, 0, sx+16, 33)).(*ebiten.Image)
			cnvPurpleChars.DrawImage(subImg, op)
			posX += 16
		}
		// Second half: character continues from -33 to 0
		for posY := -33.0; posY < 0; posY += 2.25 { // 15 steps
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(posX), posY)

			sx := charIndex * 16
			subImg := cnvFont2.SubImage(image.Rect(sx, 0, sx+16, 33)).(*ebiten.Image)
			cnvPurpleChars.DrawImage(subImg, op)
			posX += 16
		}

		// Apply gradients with masking
		// Silver gradient
		tmpSilver := ebiten.NewImage(480, 33)
		tmpSilver.DrawImage(cnvSilverBar, nil)
		opSilver := &ebiten.DrawImageOptions{}
		opSilver.Blend = ebiten.BlendDestinationIn
		tmpSilver.DrawImage(cnvSilverChars, opSilver)

		// Purple gradient
		tmpPurple := ebiten.NewImage(480, 33)
		tmpPurple.DrawImage(cnvPurpleBar, nil)
		opPurple := &ebiten.DrawImageOptions{}
		opPurple.Blend = ebiten.BlendDestinationIn
		tmpPurple.DrawImage(cnvPurpleChars, opPurple)

		// Merge all layers for this character
		frameY := charIndex * 33

		// 1. Draw purple (back) characters
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(0, float64(frameY))
		g.cnvFrames.DrawImage(tmpPurple, op)

		// 2. Draw red bar in the middle
		op.GeoM.Reset()
		op.GeoM.Translate(0, float64(frameY+12))
		g.cnvFrames.DrawImage(cnvRedBar, op)

		// 3. Draw silver (front) characters on top
		op.GeoM.Reset()
		op.GeoM.Translate(0, float64(frameY))
		g.cnvFrames.DrawImage(tmpSilver, op)

		cnvSilverChars.Deallocate()
		cnvPurpleChars.Deallocate()
		tmpSilver.Deallocate()
		tmpPurple.Deallocate()
	}
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
func (g *Game) scrollMessage(speed int) {
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
func (g *Game) shiftLeft() {
	g.scrollHead++
	if g.scrollHead == len(g.scrollChars) {
		g.scrollHead = 0
	}
}

// addSliceOfChar adds a single slice of a character to the end of the scroll
func (g *Game) addSliceOfChar(ch byte, slice int) {
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

	g.scrollChars[tail] = ScrollChar{
		glyph: uint8(glyph),
		frame: g.scrollChars[previous].frame,
		slice: uint8(slice),
	}
}

// renderNextFrames advances the animation frames with a DNA/twist effect
func (g *Game) renderNextFrames(speed float64) {
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
		g.scrollChars[index].frame = uint8(newFrame)
	}
}

// drawScroller draws the 3D rotating text scroller
func (g *Game) drawScroller(screen *ebiten.Image) {
	if g.scrollRenderer == nil {
		var err error
		g.scrollRenderer, err = scrolling.FromImages(make([]*ebiten.Image, 240), 2)
		if err != nil {
			panic(err)
		}
		g.scrollBatch = composite.NewQuadBatch(240)
		g.scrollBatch.AlternateDiagonal = true
	}
	g.scrollBatch.Begin(screen, g.cnvFrames)
	t2 := g.t
	waveSin, waveCos := math.Sincos(5*10.50 + g.t/6)
	state := scrolling.IdentityState()
	state.Paint = func(dst *ebiten.Image, s scrolling.Sample, op ebiten.DrawImageOptions) {
		i := s.Index
		charIndex := g.scrollHead + i
		if charIndex >= len(g.scrollChars) {
			charIndex -= len(g.scrollChars)
		}
		char := g.scrollChars[charIndex]
		ypos := 80.0
		if t2 > 5*50-float64(i)*.0033 {
			ypos = 80 * waveCos
		}
		index := int(char.glyph)
		sx := int(char.frame)*16 + int(char.slice)*2
		sy := index * 33
		if index < len(charset) && sx >= 0 && sx <= 480-2 && sy >= 0 && sy <= len(charset)*33-33 {
			g.scrollBatch.Rect(image.Rect(sx, sy, sx+2, sy+33), float32(i*2)*2, 156+float32(67+ypos)*1.5, 4, 33*1.5)
		}
		t2 += 1.0 / 6.0
		waveSin, waveCos = waveSin*waveCosStep+waveCos*waveSinStep, waveCos*waveCosStep-waveSin*waveSinStep
	}
	g.scrollRenderer.DrawAt(screen, state)
	g.scrollBatch.Flush()
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
	if g.audioPlayer != nil {
		_ = g.audioPlayer.Close()
		g.audioPlayer = nil
	}
	if g.ymPlayer != nil {
		_ = g.ymPlayer.Close()
		g.ymPlayer = nil
	}
}
