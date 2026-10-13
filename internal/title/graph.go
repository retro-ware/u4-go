package title

// Port of GRAPH.ASM, EGA entries only (the CGA and Tandy routines are dropped): the view's
// tile map, dots, clearing the menu and text areas, copies from pictures (plain and as a
// dissolve), the picture loader, the tile and flag animations and characters.
//
// The screen is an ega.Driver's 320x200 framebuffer of colour indexes; pictures are the same
// (picSize bytes, one a pixel). Horizontal positions and widths of picture copies are in
// 8-pixel columns, as on the EGA.

import "github.com/memmaker/u4-go/internal/ega"

var screen *ega.Driver

// picSize is a picture buffer's size: one byte a pixel.
const picSize = ega.W * ega.H

// The view of Britannia: 5 rows of 19 tiles, from x 8 and y 104.
const (
	viewRows  = 5
	viewCols  = 19
	viewX     = 8
	viewY     = 104
	tileSize  = 16
	viewTiles = viewRows * viewCols
)

// Text positions: 40 columns of 8 pixels, 25 rows of 8.
const (
	textCols  = 40
	textRows  = 25
	cellSize  = 8
	textTop   = 19  // the text area's first row (the story, the gypsy)
	textAreaY = 152 // the text area's first pixel row, cleared with the picture's lower part
)

// EGA colours of the dots.
const (
	colorRed       = 4
	colorLightCyan = 11
)

// The dissolve's random state (two bytes of a simple add-with-carry generator).
var dissolveLow, dissolveHigh uint8 = 53, 155

// dissolveMasks has, by density 0..63, which of eight pixels show: the dissolve picks among
// eight neighbouring densities at random.
var dissolveMasks = [64]uint8{
	0b00000000, 0b00000000, 0b00000000, 0b00000000, 0b00000000, 0b00000000, 0b00000000, 0b00000000,
	0b00000000, 0b00000001, 0b00000100, 0b00010000, 0b10000000, 0b00000100, 0b00000000, 0b00000001,
	0b00000000, 0b11000000, 0b00000000, 0b00010000, 0b00000100, 0b00010001, 0b00010000, 0b11000100,
	0b10000000, 0b10000001, 0b00010000, 0b00000100, 0b00010000, 0b00000001, 0b10000100, 0b11000001,
	0b00010000, 0b00000100, 0b10000100, 0b11110000, 0b00111100, 0b00111111, 0b10000100, 0b00111111,
	0b11110001, 0b11111100, 0b11000001, 0b11110001, 0b11001111, 0b11111111, 0b00111111, 0b11110001,
	0b11110001, 0b11111100, 0b10001111, 0b10001111, 0b00111111, 0b11001111, 0b11111100, 0b11111111,
	0b11111111, 0b11111111, 0b11111111, 0b11111111, 0b11111111, 0b11111111, 0b11111111, 0b11111111,
}

// pixelAt is pixel i of b, or 0 outside it.
func pixelAt(b []byte, i int) byte {
	if i < 0 || i >= len(b) {
		return 0
	}
	return b[i]
}

// setPixel sets pixel i of b unless it is outside it.
func setPixel(b []byte, i int, v byte) {
	if i >= 0 && i < len(b) {
		b[i] = v
	}
}

// drawView draws the view's tiles (viewTiles of them, row by row) into pic (GR_0_EGA to the
// screen, GR_C_EGA into a picture).
func drawView(pic []byte, tiles []uint8) {
	for row, k := 0, 0; row < viewRows; row++ {
		for col := 0; col < viewCols; col, k = col+1, k+1 {
			glyph := &screen.Tiles[tiles[k]]
			x, y := viewX+col*tileSize, viewY+row*tileSize
			for ty := 0; ty < tileSize; ty++ {
				for tx := 0; tx < tileSize; tx++ {
					setPixel(pic, (y+ty)*ega.W+x+tx, glyph[ty*tileSize+tx])
				}
			}
		}
	}
}

// plotDot (GR_DOT_EGA) adds colour to the pixel at (x, y).
func plotDot(y, x int, color uint8) {
	i := y*ega.W + x
	setPixel(screen.Pix, i, pixelAt(screen.Pix, i)|screen.Color(color))
}

func clearRows(y0, y1, x, w int) {
	for r := y0; r < y1; r++ {
		clear(screen.Pix[r*ega.W+x : r*ega.W+x+w])
	}
}

// clearMenuArea (GR_2_EGA) clears the view and menu area: rows 104..183, x 8..311.
func clearMenuArea() { clearRows(viewY, 184, viewX, 304) }

// clearTextArea (GR_5_EGA) clears the text area: rows 152..199.
func clearTextArea() {
	clearRows(textAreaY, ega.H, 0, ega.W)
	c64Clear(textAreaY, ega.H, 0, ega.W)
	clear(textCells[textTop:])
	clearAnswers()
}

// copyPic (GR_3_EGA) copies a block of cols x rows from picture pic at (srcCol, srcY) to the
// screen at (dstCol, dstY).
func copyPic(pic []byte, srcCol, srcY, cols, rows, dstCol, dstY int) {
	c64Blit(pic, srcCol, srcY, cols, rows, dstCol, dstY)
	for i := 0; i < rows; i++ {
		dst := (dstY+i)*ega.W + dstCol*cellSize
		src := (srcY+i)*ega.W + srcCol*cellSize
		for c := 0; c < cols*cellSize; c++ {
			setPixel(screen.Pix, dst+c, pixelAt(pic, src+c))
		}
	}
}

// dissolvePic (GR_B_EGA) is copyPic through a random mask of density density..density+7 (of
// 63), bottom row first, skipping 8-pixel cells blank in the picture. It clicks the speaker
// each time its generator overflows past density (more often as the picture fills in).
func dissolvePic(pic []byte, srcCol, srcY, cols, rows, dstCol, dstY, density int) {
	for ; rows != 0; rows-- {
		dst := (dstY+rows-1)*ega.W + dstCol*cellSize
		src := (srcY+rows-1)*ega.W + srcCol*cellSize
		for c := 0; c < cols; c, src, dst = c+1, src+cellSize, dst+cellSize {
			var any byte
			for k := 0; k < cellSize; k++ {
				any |= pixelAt(pic, src+k)
			}
			if any == 0 {
				continue
			}
			sum := int(dissolveLow) + 29
			next := byte(sum)
			r := byte(int(next) + int(dissolveHigh) + sum>>8)
			dissolveLow, dissolveHigh = r, next
			if int(r)+density > 255 {
				dissolveClicks++
			}
			mask := dissolveMasks[density+int(r%8)]
			for k := 0; k < cellSize; k++ {
				var v byte
				if mask&(128>>k) != 0 {
					v = pixelAt(pic, src+k)
				}
				setPixel(screen.Pix, dst+k, v)
			}
		}
	}
}

// loadPicture (GR_INFLATE_EGA) decodes picture file name into pic (cleared first): each byte
// is two pixels, the high half first. (In VGA mode the EGA colours map to the VGA palette's.)
func loadPicture(name string, pic []byte) {
	clear(pic)
	i := 0
	ok := decodeFile(name, func(b uint8) {
		setPixel(pic, i, screen.Color(b>>4))
		setPixel(pic, i+1, screen.Color(b%16))
		i += 2
	})
	if !ok {
		fatalf("can't read %s", name)
	}
	c64Inflate(name, pic)
}

// loadPacked decodes file name (SHAPES.EGZ) into buf as is.
func loadPacked(name string, buf []byte) {
	i := 0
	ok := decodeFile(name, func(b uint8) {
		setPixel(buf, i, b)
		i++
	})
	if !ok {
		fatalf("can't read %s", name)
	}
}

// initDriver (GR_INIT_EGA): the screen cleared, the default palette, then the font and tiles
// decoded to pixels.
func initDriver() {
	screen.Clear()
	screen.SetGlyphs(charset)
	screen.SetTiles(shapes)
}

func flowTile(tile int) { screen.FlowTile(tile) } /*GR_ANIMFLOW_EGA: the same as EGA.DRV's*/

// waveFlags (GR_ANIMFLAG_EGA) waves the five flags. The original waves each when its own
// random generator comes out non-negative, but the title leaves that generator's seed zero,
// where it stays, so every flag waves every time.
func waveFlags() {
	for i := 0; i < 5; i++ {
		screen.WaveFlag(i)
	}
}

func putGlyph(c uint8) { /*GR_PUTCHAR_EGA*/
	screen.PutChar(textRow, textCol, c)
	if textRow >= 0 && textRow < textRows && textCol >= 0 && textCol < textCols {
		textCells[textRow][textCol] = c // for the gypsy's answers (menus.go)
	}
}

func clearScreen() { screen.Clear(); clear(c64Scr.Pix) } /*GR_CLRSCR_EGA*/

// SetVGA switches the title to VGA (or EGA) graphics, its pictures included.
func SetVGA(on bool) {
	if screen != nil {
		screen.SetMode(on, titlePic, animPic, storyPic)
	}
}
