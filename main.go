package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"io"
	"log"
	"math"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/olivierh59500/ym-player/pkg/stsound"
)

const (
	screenWidth  = 640
	screenHeight = 480
	fps          = 60
	timePerFrame = 1000.0 / fps
	sampleRate   = 44100
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
	player       *stsound.StSound
	sampleRate   int
	buffer       []int16
	mutex        sync.Mutex
	position     int64
	totalSamples int64
	loop         bool
	volume       float64
}

// NewYMPlayer creates a new YM player instance
func NewYMPlayer(data []byte, sampleRate int, loop bool) (*YMPlayer, error) {
	player := stsound.CreateWithRate(sampleRate)

	if err := player.LoadMemory(data); err != nil {
		player.Destroy()
		return nil, fmt.Errorf("failed to load YM data: %w", err)
	}

	player.SetLoopMode(loop)

	info := player.GetInfo()
	totalSamples := int64(info.MusicTimeInMs) * int64(sampleRate) / 1000

	return &YMPlayer{
		player:       player,
		sampleRate:   sampleRate,
		buffer:       make([]int16, 4096),
		totalSamples: totalSamples,
		loop:         loop,
		volume:       0.5,
	}, nil
}

// Read implements io.Reader for audio streaming
func (y *YMPlayer) Read(p []byte) (n int, err error) {
	y.mutex.Lock()
	defer y.mutex.Unlock()

	samplesNeeded := len(p) / 4
	outBuffer := make([]int16, samplesNeeded*2)

	processed := 0
	for processed < samplesNeeded {
		chunkSize := samplesNeeded - processed
		if chunkSize > len(y.buffer) {
			chunkSize = len(y.buffer)
		}

		if !y.player.Compute(y.buffer[:chunkSize], chunkSize) {
			if !y.loop {
				for i := processed * 2; i < len(outBuffer); i++ {
					outBuffer[i] = 0
				}
				err = io.EOF
				break
			}
		}

		for i := 0; i < chunkSize; i++ {
			sample := int16(float64(y.buffer[i]) * y.volume)
			outBuffer[(processed+i)*2] = sample
			outBuffer[(processed+i)*2+1] = sample
		}

		processed += chunkSize
		y.position += int64(chunkSize)
	}

	buf := make([]byte, 0, len(outBuffer)*2)
	for _, sample := range outBuffer {
		buf = append(buf, byte(sample), byte(sample>>8))
	}

	copy(p, buf)
	n = len(buf)
	if n > len(p) {
		n = len(p)
	}

	return n, err
}

// SetVolume sets the playback volume (0.0 to 1.0)
func (y *YMPlayer) SetVolume(volume float64) {
	y.mutex.Lock()
	defer y.mutex.Unlock()
	y.volume = volume
}

// GetVolume returns the current volume
func (y *YMPlayer) GetVolume() float64 {
	y.mutex.Lock()
	defer y.mutex.Unlock()
	return y.volume
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
	Color  color.Color
	Offset float64
}

// Character set for the scroller - must match the font.png layout
// Font has 45 characters: " ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!'?/,.-@"
var charset = []string{
	" ", "A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K",
	"L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V",
	"W", "X", "Y", "Z", "0", "1", "2", "3", "4", "5", "6",
	"7", "8", "9", "!", "'", "?", "/", ",", ".", "-", "@",
}

// Control characters for special effects
var ctrlChars = []string{"^", "#", "&", "%"}

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
	char  string // The character
	frame int    // Current animation frame (0-29)
	slice int    // Current slice (0-7)
}

// Game represents the main game state
type Game struct {
	// Demo state
	state       DemoState
	initialized bool
	finished    bool

	// Images
	imgRasterbar *ebiten.Image
	imgFont      *ebiten.Image
	imgLogo      *ebiten.Image
	imgPhoton    *ebiten.Image
	imgTextPage1 *ebiten.Image
	imgTextPage2 *ebiten.Image

	// Animation canvases
	cnvFrames    *ebiten.Image // All character animation frames
	cnvScroller  *ebiten.Image // Scroller rendering buffer
	cnvPhoton    *ebiten.Image // Photon coloring buffer
	cnvLogoWhite *ebiten.Image // White logo for effects

	// Animation variables
	t                float64
	msgIndex         int
	sliceCount       int
	pause            bool
	pauseTime        int
	scrollSpeed      float64
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
	scrollChars []ScrollChar

	// Audio
	audioContext *audio.Context
	audioPlayer  *audio.Player
	ymPlayer     *YMPlayer

	// Timing
	timePrev time.Time
}

// NewGame creates a new game instance
func NewGame() *Game {
	g := &Game{
		state:            StateTextPage1,
		pauseTime:        250,
		rotSpeed:         0.35,
		scrollSpeed:      1.0, // Initialize scroll speed
		blackRectWidth:   640,
		blackRectShow:    true,
		photonY:          184,
		photonBounce:     -9.50,
		rasterbarY:       -40,
		direction:        1,
		scrollerRotation: 0,
		timePrev:         time.Now(),
		scrollChars:      make([]ScrollChar, 240),
		audioContext:     audio.NewContext(sampleRate),
	}

	// Initialize scroll chars
	for i := range g.scrollChars {
		g.scrollChars[i] = ScrollChar{
			char:  " ",
			frame: 0,
			slice: 0,
		}
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

	// Load logo
	img, _, err = image.Decode(bytes.NewReader(logoData))
	if err != nil {
		return fmt.Errorf("failed to load logo: %w", err)
	}
	g.imgLogo = ebiten.NewImageFromImage(img)

	// Load photon
	img, _, err = image.Decode(bytes.NewReader(photonData))
	if err != nil {
		return fmt.Errorf("failed to load photon: %w", err)
	}
	g.imgPhoton = ebiten.NewImageFromImage(img)

	return nil
}

// createGradient creates a horizontal gradient image
func createGradient(width, height int, stops []GradientStop) *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		t := float64(y) / float64(height-1)

		// Find the two stops to interpolate between
		var c color.Color
		for i := 0; i < len(stops)-1; i++ {
			if t >= stops[i].Offset && t <= stops[i+1].Offset {
				// Interpolate between stops[i] and stops[i+1]
				localT := (t - stops[i].Offset) / (stops[i+1].Offset - stops[i].Offset)
				c = lerpColor(stops[i].Color, stops[i+1].Color, localT)
				break
			}
		}

		for x := 0; x < width; x++ {
			img.Set(x, y, c)
		}
	}

	return ebiten.NewImageFromImage(img)
}

// lerpColor interpolates between two colors
func lerpColor(c1, c2 color.Color, t float64) color.Color {
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()

	r := uint8((float64(r1>>8)*(1-t) + float64(r2>>8)*t))
	g := uint8((float64(g1>>8)*(1-t) + float64(g2>>8)*t))
	b := uint8((float64(b1>>8)*(1-t) + float64(b2>>8)*t))
	a := uint8((float64(a1>>8)*(1-t) + float64(a2>>8)*t))

	return color.RGBA{r, g, b, a}
}

// charToFontIndex converts a character to its position in the font bitmap
func charToFontIndex(ch rune) (int, bool) {
	// The charset array defines the order of characters in the font
	// We need to find the index of the character in this array

	charStr := string(ch)

	// Handle uppercase/lowercase
	if ch >= 'a' && ch <= 'z' {
		charStr = string(ch - 32) // Convert to uppercase
	}

	// Find character in charset
	for i, c := range charset {
		if c == charStr {
			return i, true
		}
	}

	// Character not found, treat as space
	return 0, false
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

			if found && idx >= 0 {
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Scale(2, 2)
				op.GeoM.Translate(float64(x), float64(t.Y))

				// Extract character from font (assuming single row layout)
				sx := idx * 16
				sy := 0
				subImg := g.imgFont.SubImage(image.Rect(sx, sy, sx+16, sy+26)).(*ebiten.Image)

				if mode == "xor" {
					// For XOR mode, we invert the colors
					op.ColorM.Scale(-1, -1, -1, 1)
					op.ColorM.Translate(1, 1, 1, 0)
				}

				img.DrawImage(subImg, op)
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

	// Create font canvases for front and back
	cnvFont := ebiten.NewImage(len(charset)*16, 33)
	cnvFont2 := ebiten.NewImage(len(charset)*16, 33)

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
		opSilver.CompositeMode = ebiten.CompositeModeDestinationIn
		tmpSilver.DrawImage(cnvSilverChars, opSilver)

		// Purple gradient
		tmpPurple := ebiten.NewImage(480, 33)
		tmpPurple.DrawImage(cnvPurpleBar, nil)
		opPurple := &ebiten.DrawImageOptions{}
		opPurple.CompositeMode = ebiten.CompositeModeDestinationIn
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
	}

	// Create scroller canvas
	g.cnvScroller = ebiten.NewImage(480, 180)

	// Create photon coloring canvas
	g.cnvPhoton = ebiten.NewImage(70, 15)

	// Create white logo canvas
	g.cnvLogoWhite = ebiten.NewImage(640, 129)

	// log.Printf("Character frames initialized. cnvFrames size: %v", g.cnvFrames.Bounds())
}

// loadMusic loads and plays the YM music
func (g *Game) loadMusic() error {
	var err error

	// Create YM player
	g.ymPlayer, err = NewYMPlayer(musicData, sampleRate, true)
	if err != nil {
		return fmt.Errorf("failed to create YM player: %w", err)
	}

	// Create audio player
	g.audioPlayer, err = g.audioContext.NewPlayer(g.ymPlayer)
	if err != nil {
		g.ymPlayer.Close()
		g.ymPlayer = nil
		return fmt.Errorf("failed to create audio player: %w", err)
	}

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

	// Load music
	if err := g.loadMusic(); err != nil {
		log.Printf("Failed to load music: %v", err)
	}

	// Bring message to the start
	for i := 0; i < 320; i++ {
		g.scrollMessage(1)
		g.renderNextFrames(g.rotSpeed)
	}

	g.initialized = true
	return nil
}

// scrollMessage advances the scroll text
func (g *Game) scrollMessage(speed float64) {
	for i := 0; i < int(speed); i++ {
		// We need to add a new slice.
		// First, determine which character and slice index.
		chStr := string(scrollMessage[g.msgIndex])

		// Is it a control character?
		isCtrl := false
		for _, ctrl := range ctrlChars {
			if chStr == ctrl {
				isCtrl = true
				break
			}
		}

		if isCtrl && g.sliceCount == 0 {
			// Handle control characters only when starting a new character
			switch chStr {
			case "^":
				g.pause = true
				g.pauseTime = 275
				g.rotSpeed = -1
			case "&":
				g.pause = true
				g.pauseTime = 275
				g.rotSpeed = 1
			case "#":
				g.pause = true
				g.pauseTime = 250
				g.rotSpeed = -1
			case "%":
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
			g.addSliceOfChar(chStr, g.sliceCount)

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

// shiftLeft shifts all scroll characters left
func (g *Game) shiftLeft() {
	for i := 0; i < len(g.scrollChars)-1; i++ {
		g.scrollChars[i] = g.scrollChars[i+1]
	}
}

// addSliceOfChar adds a single slice of a character to the end of the scroll
func (g *Game) addSliceOfChar(ch string, slice int) {
	// The new slice inherits the frame from its left neighbor
	f := g.scrollChars[len(g.scrollChars)-2].frame

	// The last element is overwritten
	g.scrollChars[len(g.scrollChars)-1] = ScrollChar{
		char:  ch,
		frame: f,
		slice: slice,
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

	// Apply rotation to each character slice with a sine wave offset
	for i := range g.scrollChars {
		// Calculate the sine wave offset based on the slice's position
		// This creates the DNA-like twist
		sineOffset := math.Sin(float64(i)*0.05) * 15 // Adjust multiplier for twist tightness

		// Combine base rotation with the sine offset
		newFrame := g.scrollerRotation + sineOffset

		// Wrap around at 30 frames
		for newFrame >= 30 {
			newFrame -= 30
		}
		for newFrame < 0 {
			newFrame += 30
		}

		g.scrollChars[i].frame = int(newFrame)
	}
}

// drawScroller draws the 3D rotating text scroller
func (g *Game) drawScroller(screen *ebiten.Image) {
	g.cnvScroller.Fill(color.RGBA{0x00, 0x01, 0x11, 0xFF})

	t2 := g.t
	for i := 0; i < 240; i++ {
		var ypos float64
		if t2 > 5*50-float64(i)*0.0033 {
			ypos = 80 * math.Cos(5*10.50+t2/6)
		} else {
			ypos = 80
		}

		// Find character in charset
		charsetIdx := -1
		for j, c := range charset {
			if g.scrollChars[i].char == c {
				charsetIdx = j
				break
			}
		}

		if charsetIdx >= 0 && charsetIdx < len(charset) {
			// Calculate source position for this slice
			frame := g.scrollChars[i].frame
			slice := g.scrollChars[i].slice

			// Each frame is 16 pixels wide, each slice is 2 pixels
			sx := frame*16 + slice*2
			sy := charsetIdx * 33

			// Make sure we're within bounds
			if sx >= 0 && sx <= 480-2 && sy >= 0 && sy <= len(charset)*33-33 {
				// Draw 2x33 pixel slice
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(float64(i*2), 67+ypos)

				subImg := g.cnvFrames.SubImage(image.Rect(sx, sy, sx+2, sy+33)).(*ebiten.Image)
				g.cnvScroller.DrawImage(subImg, op)
			}
		}

		t2 += 1.0 / 6.0
	}

	// Draw scroller to screen with scaling
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(2, 1.5)
	op.GeoM.Translate(0, 156)
	screen.DrawImage(g.cnvScroller, op)
}

// Update updates the game state
func (g *Game) Update() error {
	if !g.initialized {
		if err := g.Init(); err != nil {
			return err
		}
	}

	// Handle volume control
	if g.ymPlayer != nil {
		if ebiten.IsKeyPressed(ebiten.KeyUp) {
			vol := g.ymPlayer.GetVolume() + 0.01
			if vol > 1.0 {
				vol = 1.0
			}
			g.ymPlayer.SetVolume(vol)
		}
		if ebiten.IsKeyPressed(ebiten.KeyDown) {
			vol := g.ymPlayer.GetVolume() - 0.01
			if vol < 0 {
				vol = 0
			}
			g.ymPlayer.SetVolume(vol)
		}
	}

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

	// Check for mouse click to finish demo
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) && g.state == StateMainDemo {
		g.finished = true
	}

	return nil
}

// Draw draws the game
func (g *Game) Draw(screen *ebiten.Image) {
	if !g.initialized {
		return
	}

	// Draw based on current state
	switch g.state {
	case StateTextPage1:
		screen.Fill(color.Black)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(0, g.rasterbarY)
		screen.DrawImage(g.imgRasterbar, op)
		screen.DrawImage(g.imgTextPage1, nil)

	case StateTextPage2:
		screen.Fill(color.Black)
		// Draw with fade effect
		op := &ebiten.DrawImageOptions{}
		brightness := g.percent / 100.0
		if brightness > 1 {
			brightness = 2 - brightness
		}
		op.ColorM.Scale(brightness, brightness, brightness, 1)
		screen.DrawImage(g.imgTextPage2, op)

	case StateShowLogo:
		screen.Fill(color.Black)
		// Draw background
		screen.Fill(color.RGBA{0x00, 0x01, 0x11, 0xFF})

		// Draw logo with white overlay effect
		if g.percent <= 100 {
			// Fade in white
			op := &ebiten.DrawImageOptions{}
			brightness := g.percent / 100.0
			op.ColorM.Scale(brightness, brightness, brightness, 1)
			screen.DrawImage(g.imgLogo, op)
		} else {
			// Show normal logo with fading white overlay
			screen.DrawImage(g.imgLogo, nil)

			// Draw white overlay
			g.cnvLogoWhite.Clear()
			g.cnvLogoWhite.DrawImage(g.imgLogo, nil)
			op := &ebiten.DrawImageOptions{}
			op.ColorM.Scale(1, 1, 1, 1)
			op.ColorM.Translate(1, 1, 1, 0)
			alpha := (200 - g.percent) / 100.0
			op.ColorM.Scale(1, 1, 1, alpha)
			screen.DrawImage(g.cnvLogoWhite, op)
		}

	case StateShowUpperRasterbar, StateShowLowerRasterbar, StateDropPhoton, StatePhotonFadeToRed:
		screen.Fill(color.Black)
		// Draw background
		for y := 130; y < 430; y++ {
			for x := 0; x < 640; x++ {
				screen.Set(x, y, color.RGBA{0x00, 0x01, 0x11, 0xFF})
			}
		}

		// Draw logo
		screen.DrawImage(g.imgLogo, nil)

		// Draw upper rasterbar
		if g.state >= StateShowUpperRasterbar {
			alpha := 1.0
			if g.state == StateShowUpperRasterbar {
				alpha = g.percent / 100.0
			}
			op := &ebiten.DrawImageOptions{}
			op.ColorM.Scale(1, 1, 1, alpha)
			op.GeoM.Translate(0, 129)
			// Use gradient instead of image
			rasterGrad := createGradient(640, 12, gdcRasterBar)
			screen.DrawImage(rasterGrad, op)
		}

		// Draw lower rasterbar
		if g.state >= StateShowLowerRasterbar {
			alpha := 1.0
			if g.state == StateShowLowerRasterbar {
				alpha = g.percent / 100.0
			}
			op := &ebiten.DrawImageOptions{}
			op.ColorM.Scale(1, 1, 1, alpha)
			op.GeoM.Translate(0, 430)
			rasterGrad := createGradient(640, 12, gdcRasterBar)
			screen.DrawImage(rasterGrad, op)
		}

		// Draw photon
		if g.state >= StateDropPhoton {
			if g.state == StatePhotonFadeToRed {
				// Draw with red tint
				g.cnvPhoton.Clear()
				g.cnvPhoton.DrawImage(g.imgPhoton, nil)

				// Apply red tint based on percent
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(285, 445)

				// Create red tint
				lightness := g.percent / 100.0
				op.ColorM.Scale(lightness, lightness*0.5, lightness*0.5, 1)

				screen.DrawImage(g.cnvPhoton, op)
			} else {
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(285, g.photonY)
				screen.DrawImage(g.imgPhoton, op)
			}
		}

	case StateMainDemo:
		screen.Fill(color.Black)
		// Draw background
		for y := 130; y < 430; y++ {
			for x := 0; x < 640; x++ {
				screen.Set(x, y, color.RGBA{0x00, 0x01, 0x11, 0xFF})
			}
		}

		// Draw logo
		screen.DrawImage(g.imgLogo, nil)

		// Draw raster bars
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(0, 129)
		rasterGrad := createGradient(640, 12, gdcRasterBar)
		screen.DrawImage(rasterGrad, op)

		op.GeoM.Reset()
		op.GeoM.Translate(0, 430)
		screen.DrawImage(rasterGrad, op)

		// Draw photon with color cycling
		g.cnvPhoton.Clear()
		g.cnvPhoton.DrawImage(g.imgPhoton, nil)

		op = &ebiten.DrawImageOptions{}
		// Apply HSL color based on g.color
		hue := g.color / 360.0
		r, g2, b := hslToRGB(hue, 1.0, 0.5)
		op.ColorM.Scale(0, 0, 0, 1)
		op.ColorM.Translate(r, g2, b, 0)
		op.GeoM.Translate(285, 445)
		screen.DrawImage(g.cnvPhoton, op)

		// Draw scroller
		g.drawScroller(screen)

		// Draw black rect for reveal effect
		if g.blackRectShow {
			for y := 375; y < 430; y++ {
				for x := 0; x < int(g.blackRectWidth); x++ {
					screen.Set(x, y, color.RGBA{0x00, 0x01, 0x11, 0xFF})
				}
			}
		}

	case StateHideLogo, StateHideLowerRasterbar, StateHideUpperRasterbar:
		screen.Fill(color.Black)

		if g.state == StateHideLogo {
			// Draw logo with white effect
			if g.direction > 0 {
				screen.DrawImage(g.imgLogo, nil)

				// White overlay
				g.cnvLogoWhite.Clear()
				g.cnvLogoWhite.DrawImage(g.imgLogo, nil)
				op := &ebiten.DrawImageOptions{}
				alpha := g.percent / 100.0
				op.ColorM.Scale(1, 1, 1, alpha)
				op.ColorM.Translate(1, 1, 1, 0)
				screen.DrawImage(g.cnvLogoWhite, op)
			} else {
				// Just white logo fading out
				g.cnvLogoWhite.Clear()
				g.cnvLogoWhite.DrawImage(g.imgLogo, nil)
				op := &ebiten.DrawImageOptions{}
				alpha := g.percent / 100.0
				op.ColorM.Scale(1, 1, 1, alpha)
				op.ColorM.Translate(1, 1, 1, 0)
				screen.DrawImage(g.cnvLogoWhite, op)
			}
		}

		// Draw raster bars with fade
		if g.state >= StateHideLowerRasterbar {
			if g.state == StateHideLowerRasterbar {
				op := &ebiten.DrawImageOptions{}
				op.ColorM.Scale(1, 1, 1, g.percent/100.0)
				op.GeoM.Translate(0, 430)
				rasterGrad := createGradient(640, 12, gdcRasterBar)
				screen.DrawImage(rasterGrad, op)
			}
		}

		if g.state >= StateHideUpperRasterbar {
			if g.state == StateHideUpperRasterbar {
				op := &ebiten.DrawImageOptions{}
				op.ColorM.Scale(1, 1, 1, g.percent/100.0)
				op.GeoM.Translate(0, 129)
				rasterGrad := createGradient(640, 12, gdcRasterBar)
				screen.DrawImage(rasterGrad, op)
			}
		}

	case StateEnd:
		screen.Fill(color.Black)
	}
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
		g.audioPlayer.Close()
	}
	if g.ymPlayer != nil {
		g.ymPlayer.Close()
	}
}

func main() {
	ebiten.SetWindowSize(screenWidth*2, screenHeight*2)
	ebiten.SetWindowTitle("Phenomena - Enigma (Go/Ebiten Port)")
	ebiten.SetVsyncEnabled(true)

	game := NewGame()

	// Ensure cleanup on exit
	defer game.Cleanup()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
