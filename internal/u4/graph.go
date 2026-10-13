package u4

// Port of U4_GRAPH.ASM: loading the graphics driver, the screen frame, drawLine and the
// thunks into the driver (here: the ega package, plain pixel code).

import (
	"github.com/memmaker/u4-go/internal/ega"
	"github.com/memmaker/u4-go/internal/host"
)

// screen is the graphics driver (EGA.DRV, with the VGA upgrade as a mode of it).
var screen *ega.Driver

// checkDriverFile insists on the driver's file: the port has the driver built in, but the
// original stopped without it.
func checkDriverFile(name string) {
	if _, ok := readGameFile(name); !ok {
		fatalf("can't open %s", name)
	}
}

// drawFrame draws the screen's frame (START.EGA).
func drawFrame() {
	clearScreen()
	loadPicture("start.ega", 0, nil)
}

// drawLine ORs a line of colour onto the screen: straight lines here, slanted ones by
// drawSlantedLine.
func drawLine(x0, y0, x1, y1, color int) {
	switch {
	case x0 != x1 && y0 != y1:
		drawSlantedLine(x0, y0, x1, y1)
	case x0 == x1:
		step := signInt(y1 - y0)
		for i, y := 0, y0; i <= abs(y1-y0); i, y = i+1, y+step {
			plotOR(y, x0, color)
		}
	default:
		step := signInt(x1 - x0)
		for i, x := 0, x0; i <= abs(x1-x0); i, x = i+1, x+step {
			plotOR(y0, x, color)
		}
	}
}

// ---- the driver's entries ----

func drawCursorBox(y, x int)                      { screen.DrawCursorBox(y, x) } // active character square
func drawTileRow(tile, trow, srow, scol, exp int) { screen.DrawTileRow(tile, trow, srow, scol, exp) }
func fillTiles(h, w, tile, row, col int)          { screen.FillTiles(h, w, tile, row, col) }
func shakeDown()                                  { screen.ShakeDown(soundOn != 0); present() }
func shakeUp()                                    { screen.ShakeUp(soundOn != 0); present() }
func drawMap(h, w int, src []byte, dst int)       { screen.DrawMap(h, w, src) } // display tile map
func plot(y, x, col int)                          { screen.Dot(y, x, col) }
func newLine()                                    { textRow = textBottom; textCol = textLeft; screen.CR(); histCR() }
func invertMap()                                  { screen.InvertMap() }
func clearMap()                                   { screen.ClearMap() }
func invertStatusLine(a int)                      { screen.InvertStatusLine(a) }
func Gra_12(x, y int)                             { screen.InvertCell(x, y) } // (uncalled)
func clearStatus()                                { screen.ClearStatus() }
func plotOR(y, x, col int)                        { screen.DotOR(y, x, col) }
func plotXOR(y, x, col int)                       { screen.DotXOR(y, x, col) }
func loadPicture(name string, xoff int, mem []byte) { // display image
	if mem == nil {
		host.InWorld = false // a full-screen picture: the original layout until the next command
	}
	screen.LoadPicture(name, xoff, mem)
}
func fillQuad(x0, x1, y0, y1, y0p, y1p, col, erase int) {
	screen.FillQuad(x0, x1, y0, y1, y0p, y1p, col, erase)
}
func flickerFlames()    { screen.FlickerFlames() }
func clearScreen()      { screen.Clear() }
func flowTile(tile int) { screen.FlowTile(tile) }
func waveFlags()        { screen.WaveFlags() }
func putGlyph(c uint8)  { screen.PutChar(int(textRow), int(textCol), c) }

// showInfinity opens the map window onto the symbol of infinity (RUNE_5).
func showInfinity() {
	if screen.IsVGA() { // the VGA patch's infinity is rune_0
		screen.OpenDoors("rune_0.ega")
	} else {
		screen.OpenDoors("rune_5.ega")
	}
}

// updateAnkh draws the ankh on the status line from the mastered virtues.
func updateAnkh() {
	var k [8]uint16
	for i, p := range karmas {
		k[i] = *p
	}
	screen.BuildAnkh(k)
}

func initDriver() {
	screen.Init(driverExit)
}

// drawSlantedLine ORs a slanted line in colour 3 from (x0, y0) to (x1, y1), stepping along
// the longer axis (Bresenham's way).
func drawSlantedLine(x0, y0, x1, y1 int) {
	const color = 3
	x, y := x0, y0
	stepX, stepY := signInt(x1-x0), signInt(y1-y0)
	width, height := abs(x1-x0), abs(y1-y0)
	if width < height {
		err := height / 2
		for {
			plotOR(y, x, color)
			y += stepY
			if err -= width; err <= 0 {
				x += stepX
				err += height
			}
			if y == y1 {
				break
			}
		}
	} else {
		err := width / 2
		for {
			plotOR(y, x, color)
			x += stepX
			if err -= height; err <= 0 {
				y += stepY
				err += width
			}
			if x == x1 {
				break
			}
		}
	}
	plotOR(y, x, color)
}
