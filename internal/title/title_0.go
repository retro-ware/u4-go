package title

// Port of TITLE_0.C and TITLE_X.C: the intro (Origin logo, the "Ultima IV" dissolve, the view
// opening like a scroll), the view of Britannia with its scripted actors, the two animated
// monsters, the menu, and the title's main loop. Plus Main, the entry point.

import (
	"time"

	"github.com/memmaker/u4-go/internal/ega"
	"github.com/memmaker/u4-go/internal/host"
	"github.com/memmaker/u4-go/internal/music"
	"github.com/memmaker/u4-go/internal/speaker"
)

var (
	monsterAStep  int           // where each monster is in its sequence of frames
	monsterBStep  int           //
	monsterDue    time.Duration // when animateMonsters moves the monsters next
	flowCountdown int           // view frames until the water flows again
	scriptPos     int           // the view script's last command read
	frameHold     int           // frames the view holds before reading on
	viewMap       [viewTiles]uint8
	inView        bool // the view runs: the monsters move whenever asked
)

// The menu.
const (
	menuFirstRow  = 17 // Return to the view, Journey Onward, Initiate New Game
	menuItems     = 3
	menuCol       = 11
	menuEndCol    = 29
	menuAnswerX   = 24 // where the typed choice shows
	menuAnswerY   = 16
	copyrightSign = "\t" // the font's glyph 9 is ©
)

// The cursor turns more slowly while the monsters are drawn.
const (
	cursorRateWithMonsters = 200
	cursorRateAlone        = 233
)

// Main runs the title (TITLE.EXE) and returns nil when the player journeys onward (the game
// starts next), or the error that stopped it.
//
// It must run on the host's game goroutine; it shows the screen through host.ShowFrame.
func Main(gameDir, saveDir string) (err error) {
	GameDir, SaveDir = gameDir, saveDir
	music.Play(music.Towns) // as ULTIMA.COM asked MIDPAK before running TITLE.EXE
	resetGlobals()
	defer func() {
		if r := recover(); r != nil {
			f, ok := r.(fail)
			if !ok {
				panic(r)
			}
			err = f.err
		}
	}()
	run()
	return nil
}

var frame = host.Frame{Pix: make([]byte, host.ScreenW*host.ScreenH), Palette: new([768]byte)}

// resetGlobals puts the initialised data back as a fresh load of TITLE.EXE has it (ULTIMA.COM
// runs it again after each game), and shows the framebuffer as the host's frame.
func resetGlobals() {
	monsterAStep, monsterBStep = 0, 0
	flowCountdown, scriptPos, frameHold = 1, -1, 0
	inView, monstersShown = false, false
	cursorShown, cursorRate, cursorPhase = true, 0, 0
	monsterDue, cursorDue = 0, 0
	dissolveLow, dissolveHigh = 53, 155
	party = partyRecord{}
	people = peopleTable{}

	screen = ega.New(frame.Pix)
	screen.Palette = frame.Palette
	screen.SetMode(ega.VGA)
	c64Reset()
	host.ShowFrame = func() *host.Frame { return &frame }
	setHooks()
}

// drawMonsters draws the two monsters' next frames, the lowest rows of them (all
// monsterHeight once they have risen in).
func drawMonsters(rows int) {
	f := monsterASteps[monsterAStep]
	monsterAStep = (monsterAStep + 1) % len(monsterASteps)
	copyPic(animPic, monsterAFrameCol(f), monsterFrameY(f)+monsterHeight-rows, monsterCols, rows, 0, 0)
	f = monsterBSteps[monsterBStep]
	monsterBStep = (monsterBStep + 1) % len(monsterBSteps)
	copyPic(animPic, monsterBFrameCol(f), monsterFrameY(f)+monsterHeight-rows, monsterCols, rows, monsterBCol, 0)
}

// animateMonsters moves the two monsters when it is time (called while waiting for keys).
func animateMonsters() { /*C_019A*/
	if !inView && now() < monsterDue {
		return
	}
	monsterDue = now() + monsterFrame
	if monstersShown {
		cursorRate = cursorRateWithMonsters
		drawMonsters(monsterHeight)
	} else {
		cursorRate = cursorRateAlone
	}
}

// putChar prints a character at the text cursor and moves it on; a backspace moves back and
// erases.
func putChar(c uint8) { /*C_027A*/
	if c == '\b' {
		textCol--
		putGlyph(' ')
	} else {
		putGlyph(c)
		textCol++
	}
}

// pause waits passes calibrated loop passes, unless a key is waiting.
func pause(passes int) { /*C_02A3*/
	if keyWaiting() {
		return
	}
	pace(time.Duration(passes) * delayPass)
}

// drawSignature writes "Lord British" by hand, slowly dot by dot when slowly is set.
func drawSignature(slowly bool) { /*C_02D1*/
	const left, bottom = 20, 191
	for _, dot := range signature {
		x, y := left+int(dot[0]), bottom-int(dot[1])
		plotDot(y, x, colorLightCyan)
		plotDot(y, x+1, colorLightCyan)
		if slowly {
			pause(440)
		}
	}
	if slowly {
		delay(1, true)
	}
}

// animateViewTiles animates the actors' tiles (people and sea monsters over two frames, land
// monsters over four, three times in four), then, every few frames, the water and the fields
// and the flags.
func animateViewTiles() { /*C_034D*/
	const odds = 192 // of 256
	for i := range people.tile {
		t := people.tile[i]
		switch {
		case (t >= tileNixie && t <= tileTwister) || (t >= tileMage && t <= tileShepherd) || (t >= tileGuard && t <= tileLordBritish):
			if quickRandom() < odds {
				people.baseTile[i] = (people.baseTile[i]%2 | t) ^ 1
			}
		case t >= tileRat:
			if quickRandom() < odds {
				people.baseTile[i] = (people.baseTile[i]+1)%4 | t
			}
		case t != tilePirateWest:
			people.baseTile[i] = t
		}
	}
	flowCountdown--
	if flowCountdown != 0 {
		return
	}
	if speedInfo == 1 {
		flowCountdown = 1
	} else {
		flowCountdown = int(speedInfo) - 1
	}
	for _, t := range []int{tileDeepWater, tileMediumWater, tileShallowWater, tileLava, tilePoisonField, tileFireField, tileSleepField, tileEnergyField} {
		flowTile(t)
	}
	waveFlags()
}

// stepViewScript runs the view's script up to its next frame, and builds that frame's map in
// viewMap.
func stepViewScript() { /*C_041A*/
	if frameHold != 0 {
		frameHold--
		return
	}
	for {
		scriptPos++
		c := viewScript[scriptPos]
		switch {
		case c < scriptVanish: // place(row, actor), at(frame, col)
			actor := c % 16
			pos := viewScript[scriptPos+1]
			people.y[actor] = c / 16
			people.x[actor] = pos % 32
			people.baseTile[actor] = actorTiles[actor] + pos/32
			if actor == actorPirateShip {
				people.tile[actor] = actorTiles[actor] // turned by its frame, not animated
			} else {
				people.tile[actor] = people.baseTile[actor]
			}
			scriptPos++
		case c < scriptEndFrame:
			actor := c % 16
			people.baseTile[actor] = 0
			people.tile[actor] = 0
		case c < scriptEndFrame+16:
			viewMap = viewBackground
			for i, t := range people.baseTile {
				if t != 0 {
					viewMap[viewCols*int(people.y[i])+int(people.x[i])] = t
				}
			}
			frameHold = int(c % 16)
			return
		default: // scriptRestart
			scriptPos = -1
		}
	}
}

// viewFrame moves the view on by one frame and draws it (more often on faster machines).
func viewFrame() {
	animateViewTiles()
	stepViewScript()
	drawView(screen.Pix, viewMap[:])
	for n := speedInfo / 2; n != 0 && !keyWaiting(); n-- {
		drawView(screen.Pix, viewMap[:])
	}
	pace(viewStep)
}

// showView runs the view of Britannia until a key is pressed (the key is taken).
func showView() { /*C_05A4*/
	inView = true
	cursorShown = false
	for !keyWaiting() {
		monstersShown = false
		viewFrame()
		viewFrame()
		monstersShown = true
	}
	inView = false
	readKey()
}

// playIntro shows the Origin logo and the title, piece by piece; a key skips ahead.
func playIntro() { /*C_068C*/
	delay(1, true)
	drawSignature(true)
	copyPic(titlePic, 19, 17, 4, 4, 19, 17) // "and"
	delay(1, true)
	for x := 86; x < 238; x++ { // the line under them
		plotDot(31, x, colorRed)
		pause(320)
	}
	for rows := 1; rows < 10; rows++ { // "Origin Systems Inc." slides down
		copyPic(titlePic, 9, 21, 21, rows, 9, 30-rows)
		pause(5000)
	}
	delay(1, true)
	if !keyWaiting() {
		for rows := 1; rows <= 5; rows++ { // "present" slides down
			copyPic(titlePic, 14, 5-rows, 15, rows, 14, 33)
			pause(5000)
		}
		delay(1, true)
		for density := 0; density < 57; density++ { // "Ultima IV" dissolves in
			if keyWaiting() {
				density = 56
			}
			dissolvePic(titlePic, 5, 34, 30, 45, 5, 34, density)
			if speedInfo > 4 {
				dissolvePic(titlePic, 5, 34, 30, 45, 5, 34, density)
			}
			speaker.Play(clicks(dissolveStep))
			pace(dissolveStep)
		}
		delay(1, true)
	} else {
		copyPic(titlePic, 5, 34, 30, 45, 5, 34)
		copyPic(titlePic, 14, 0, 15, 5, 14, 33)
	}
	for rows := 1; rows <= 6; rows++ { // "Quest of the Avatar" opens from its middle line
		copyPic(titlePic, 3, 81, 33, rows, 3, 86-rows)
		copyPic(titlePic, 3, 93-rows, 33, rows, 3, 86)
		pause(4000)
	}
	for width := 1; width <= 20; width++ { // the view opens like a scroll
		if keyWaiting() {
			width = 20
		}
		animateViewTiles()
		drawView(titlePic, viewBackground[:])
		if speedInfo > 4 {
			drawView(titlePic, viewBackground[:])
		}
		edge := 20 - width
		copyPic(titlePic, edge, 96, width*2, 96, edge, 96)
		pace(scrollStep)
		pause(680)
		animateViewTiles()
		drawView(titlePic, viewBackground[:])
	}
	titlePic = nil

	animPic = make([]byte, picSize)
	loadPicture("ANIMATE.EGA", animPic)
	for rows := 1; rows <= monsterHeight; rows++ { // the monsters rise in, the view moving
		if keyWaiting() {
			rows = monsterHeight
		}
		drawMonsters(rows)
		viewFrame()
		viewFrame()
	}
}

// printAt prints s at a text position.
func printAt(row, col int, s string) { /*C_0B1E*/
	textRow, textCol = row, col
	for i := 0; i < len(s); i++ {
		putChar(s[i])
	}
}

// showMenu shows the menu under the title.
func showMenu() { /*C_0B45*/
	clearMenuArea()
	printAt(14, 2, "In another world, in a time to come.")
	printAt(16, 15, "Options:")
	printAt(menuFirstRow, menuCol, "Return to the view")
	printAt(menuFirstRow+1, menuCol, "Journey Onward")
	printAt(menuFirstRow+2, menuCol, "Initiate New Game")
	printAt(21, 3, "Conversion by James Van Artsdalen")
	printAt(22, 5, copyrightSign+" Copyright 1987 Lord British")
	textCol, textRow = menuAnswerX, menuAnswerY
	cursorShown = true
}

// redrawTitle draws the title screen without the intro (back from an abandoned character
// creation).
func redrawTitle() { /*C_0BCA*/
	clearScreen()
	shapes = make([]byte, shapesSize)
	loadPacked("SHAPES.EGZ", shapes)
	screen.SetTiles(shapes)

	titlePic = make([]byte, picSize)
	loadPicture("TITLE.EGA", titlePic)
	drawSignature(false)
	copyPic(titlePic, 19, 17, 4, 4, 19, 17) // "and"
	for x := 86; x < 238; x++ {
		plotDot(31, x, colorRed)
	}
	copyPic(titlePic, 9, 21, 21, 9, 9, 21)  // "Origin Systems Inc."
	copyPic(titlePic, 5, 34, 30, 45, 5, 34) // "Ultima IV"
	copyPic(titlePic, 14, 0, 15, 5, 14, 33) // "present"
	copyPic(titlePic, 3, 81, 33, 6, 3, 80)  // "Quest of the Avatar"
	copyPic(titlePic, 3, 87, 33, 6, 3, 86)
	copyPic(titlePic, 0, 96, 40, 96, 0, 96) // the view's frame
	titlePic = nil

	animPic = make([]byte, picSize)
	loadPicture("ANIMATE.EGA", animPic)
	drawMonsters(monsterHeight)

	monstersShown = true
	showMenu()
}

// shapesSize is SHAPES.EGA's size: 256 tiles of 16x16 pixels, two a byte.
const shapesSize = 256 * tileSize * tileSize / 2

func run() { /*C_0EAA*/
	cursorShown = false
	seedRandom()

	monstersShown = false
	charset = make([]byte, charsetSize)
	Load("CHARSET.EGA", charset)
	shapes = make([]byte, shapesSize)
	loadPacked("SHAPES.EGZ", shapes)
	initDriver()

	titlePic = make([]byte, picSize)
	loadPicture("TITLE.EGA", titlePic)
	inView = false
	playIntro()

	monstersShown = true
	showView()
	showMenu()
	// The port's addition: the cursor keys pick an option, Enter takes it (the letters
	// still work). The picked line is shown inverted.
	sel := 1
	mark := func() {
		for x := menuCol; x < menuEndCol; x++ {
			screen.InvertCell(x, menuFirstRow+sel)
		}
	}
	for {
		mark()
		askThen(askMain)
		key := readKey()
		askThen(askNone)
		mark()
		switch key {
		case keyUp:
			sel = (sel + menuItems - 1) % menuItems
			continue
		case keyDown:
			sel = (sel + 1) % menuItems
			continue
		case keyEnter:
			key = menuKeys[sel]
		}
		key = lowerKey(key)
		if c := uint8(key); c > ' ' && c < 128 {
			putGlyph(c)
		}
		switch key {
		case keyR: // Return to the view
			showView()
			showMenu()
		case keyI: // Initiate New Game
			if initiateNewGame() {
				return
			}
			redrawTitle()
		case keyJ: // Journey Onward
			animPic, shapes, charset = nil, nil, nil
			return
		default:
			beep()
		}
	}
}

// menuKeys are the menu's lines' keys, top to bottom.
var menuKeys = [menuItems]int{keyR, keyJ, keyI}

// charsetSize is CHARSET.EGA's size: 128 glyphs of 8x8 pixels, two a byte.
const charsetSize = 128 * cellSize * cellSize / 2

func isUpper(c uint8) bool { return c >= 'A' && c <= 'Z' }

// lowerKey makes a key word's capital letter small.
func lowerKey(key int) int {
	if isUpper(uint8(key)) {
		return key + 'a' - 'A'
	}
	return key
}

// upperChar makes a small letter capital.
func upperChar(c int) int {
	if c >= 'a' && c <= 'z' {
		return c - ('a' - 'A')
	}
	return c
}
