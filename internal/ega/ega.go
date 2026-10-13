// Package ega is Ultima IV's EGA graphics driver (EGA.DRV) redone as plain pixel code, and
// the VGA driver of the "Ultima 4 VGA upgrade patch" (internal/upgrade) as a mode of it.
//
// The screen is a 320x200 framebuffer of 8-bit colour indexes into a 256-colour palette:
// in EGA mode only 0..15 are used (the EGA palette, Palette), in VGA mode all (u4vga.pal).
// The tile set (SHAPES.EGA: 256 tiles of 16x16, or shapes.vga) and the font (CHARSET.EGA:
// 128 glyphs of 8x8, or charset.vga) are decoded into per-pixel arrays, again on a mode
// switch (SetMode). Every routine keeps the original's visible result, quirks included
// (the shake, the opening doors, the dungeon walls' Bresenham slip).
//
// The game's colour numbers for dots and walls (CGA's 0..3) map to EGA colours through
// colorOf; the driver's random generator is kept for the shake clicks and the animations.
package ega

import "github.com/memmaker/u4-go/internal/upgrade"

// VGA is the mode the program runs in: new drivers' callers pass it to SetMode.
var VGA = true

const (
	W = 320
	H = 200
)

// Sizes and the screen's regions, in pixels.
const (
	tileSize  = 16
	glyphSize = 8

	mapLeft   = 8   // the map window: 11x11 tiles inside the frame
	mapTop    = 8   //
	mapSize   = 176 //
	mapTiles  = 11  //
	mapRight  = mapLeft + mapSize
	mapBottom = mapTop + mapSize

	sideLeft      = 192 // the status box and the text window, right of the map
	statusWidth   = 120
	statusHeight  = 64
	textWidth     = 128
	textTop       = 96
	textBottom    = 192
	lastTextRow   = textBottom - glyphSize
	tileCampfire  = 75 // its flames flicker
	white         = 15
	vgaInvertMask = 255
	egaInvertMask = 7
)

// wrap16 wraps a video memory offset to 16 bits, as the 8086's address arithmetic did.
func wrap16(v int) int { return v & (1<<16 - 1) }

// Driver is the screen, the tile set and the font.
type Driver struct {
	Pix    []byte // W*H colour indexes, top row first
	Tiles  [256][16 * 16]byte
	Glyphs [128][8 * 8]byte

	// ReadFile returns a file's contents for LoadPicture/OpenDoors; ok false is the driver's error path.
	ReadFile func(name string) ([]byte, bool)
	// Exit is the game's exit, called with the error code.
	Exit func(code int)
	// Speaker toggles the PC speaker (the shake's random clicks).
	Speaker func(on bool)
	// Present lets the host show the screen inside a routine that animates on its own (shake,
	// OpenDoors's opening doors).
	Present func()

	// Palette is filled with the mode's 256 colours (6-bit) by SetMode, when set.
	Palette *[768]byte

	seed [4]byte // the random state

	vga                 bool
	egaTiles, egaGlyphs []byte // the EGA files, for decoding again on a mode switch
}

// IsVGA says whether the driver is in VGA mode.
func (d *Driver) IsVGA() bool { return d.vga }

// SetMode switches to VGA (or EGA) mode: tiles, font and palette, and the pixels already on
// the screen and in bufs (pictures kept for later) mapped to the nearest colours.
func (d *Driver) SetMode(vga bool, bufs ...[]byte) {
	if vga != d.vga {
		lut := egaOf
		if vga {
			lut = vgaOfEGA
		}
		for _, b := range append(bufs, d.Pix) {
			for i, c := range b {
				b[i] = lut[c]
			}
		}
	}
	d.vga = vga
	d.load()
	if d.Palette != nil {
		clear(d.Palette[:])
		if vga {
			copy(d.Palette[:], upgrade.Palette)
		} else {
			for i, c := range Palette {
				copy(d.Palette[i*3:], c[:])
			}
		}
	}
}

func (d *Driver) load() {
	for t := range d.Tiles {
		if d.vga {
			copy(d.Tiles[t][:], upgrade.Shapes[t*256:])
		} else {
			unpack(d.Tiles[t][:], d.egaTiles[min(len(d.egaTiles), t*128):])
		}
	}
	for c := range d.Glyphs {
		if d.vga {
			copy(d.Glyphs[c][:], upgrade.Charset[c*64:])
		} else {
			unpack(d.Glyphs[c][:], d.egaGlyphs[min(len(d.egaGlyphs), c*32):])
		}
	}
}

// Color is EGA colour c (0..15) as an index in the current mode's palette.
func (d *Driver) Color(c byte) byte {
	if d != nil && d.vga {
		return vgaOfEGA[c%16]
	}
	return c
}

// nearest returns the index of pal's colour nearest to c.
func nearest(pal [][3]byte, c [3]byte) byte {
	best, bd := 0, 1<<30
	for i, p := range pal {
		dd := 0
		for k := range p {
			v := int(p[k]) - int(c[k])
			dd += v * v
		}
		if dd < bd {
			best, bd = i, dd
		}
	}
	return byte(best)
}

// egaOf maps VGA palette indexes to the nearest EGA colour; vgaOfEGA EGA colours to the
// nearest VGA index (indexes past 15 map to themselves).
var egaOf, vgaOfEGA = func() (e, v [256]byte) {
	var vp [][3]byte
	for i := 0; i+3 <= len(upgrade.Palette); i += 3 {
		vp = append(vp, [3]byte(upgrade.Palette[i:i+3]))
	}
	for i := range v {
		v[i] = byte(i)
		if i < len(vp) {
			e[i] = nearest(Palette[:], vp[i])
		}
	}
	for i, c := range Palette {
		v[i] = nearest(vp, c)
	}
	return
}()

// New returns the driver drawing into pix (W*H bytes; a fresh buffer when nil).
func New(pix []byte) *Driver {
	if pix == nil {
		pix = make([]byte, W*H)
	}
	return &Driver{Pix: pix, seed: [4]byte{18, 52, 86, 120}}
}

// unpack decodes packed pixels (two a byte, high nibble first) into dst; missing bytes are 0.
func unpack(dst []byte, packed []byte) {
	for i := range dst {
		var b byte
		if i/2 < len(packed) {
			b = packed[i/2]
		}
		if i%2 == 0 {
			dst[i] = b >> 4
		} else {
			dst[i] = b % 16
		}
	}
}

// SetTiles decodes SHAPES.EGA's 256 tiles (128 bytes each).
func (d *Driver) SetTiles(packed []byte) {
	d.egaTiles = packed
	d.load()
}

// SetGlyphs decodes CHARSET.EGA's 128 glyphs (32 bytes each).
func (d *Driver) SetGlyphs(packed []byte) {
	d.egaGlyphs = packed
	d.load()
}

// colorOf maps the game's dot and wall colours (CGA's 0..3) to EGA colours; past the
// driver's table it read zeros.
func colorOf(c int) byte {
	switch byte(c) {
	case 1:
		return 2
	case 2:
		return 1
	case 3:
		return 7
	}
	return 0
}

// Palette is the EGA palette (the default registers 0..7 and 56..63) as 6-bit VGA DAC
// values, for colour indexes 0..15.
var Palette = func() [16][3]byte {
	regs := [16]byte{0, 1, 2, 3, 4, 5, 6, 7, 56, 57, 58, 59, 60, 61, 62, 63}
	var p [16][3]byte
	for i, r := range regs {
		// EGA 6-bit colour: bits 5..0 = r' g' b' R G B (primary = 2/3, secondary = 1/3)
		p[i] = [3]byte{
			(r>>2&1)*42 + (r>>5&1)*21,
			(r>>1&1)*42 + (r>>4&1)*21,
			(r&1)*42 + (r>>3&1)*21,
		}
	}
	return p
}()

// --- pixel helpers -----------------------------------------------------------------------
// Pixels are addressed by linear index y*W + x: an x past the right edge runs on into the
// next row, as it did in video memory; anything off the screen is dropped (reads 0).

func (d *Driver) at(i int) *byte {
	if i < 0 || i >= len(d.Pix) {
		return nil
	}
	return &d.Pix[i]
}

func (d *Driver) set(i int, c byte) {
	if p := d.at(i); p != nil {
		*p = c
	}
}

func (d *Driver) or(i int, c byte) {
	if p := d.at(i); p != nil {
		*p |= c
	}
}

func (d *Driver) xor(i int, c byte) {
	if p := d.at(i); p != nil {
		*p ^= c
	}
}

func (d *Driver) get(i int) byte {
	if p := d.at(i); p != nil {
		return *p
	}
	return 0
}

// fill sets w x h pixels from (x, y) to c.
func (d *Driver) fill(x, y, w, h int, c byte) {
	for r := y; r < y+h; r++ {
		for i := 0; i < w; i++ {
			d.set(r*W+x+i, c)
		}
	}
}

// invert XORs w x h pixels from (x, y) with c.
func (d *Driver) invert(x, y, w, h int, c byte) {
	for r := y; r < y+h; r++ {
		for i := 0; i < w; i++ {
			d.xor(r*W+x+i, c)
		}
	}
}

// copyRow copies n pixels of row src to row dst, from x.
func (d *Driver) copyRow(dst, src, x, n int) {
	for i := 0; i < n; i++ {
		d.set(dst*W+x+i, d.get(src*W+x+i))
	}
}

// Tile draws tile t with its top left at (x, y).
func (d *Driver) Tile(t, x, y int) {
	g := &d.Tiles[byte(t)]
	for r := 0; r < tileSize; r++ {
		for c := 0; c < tileSize; c++ {
			d.set((y+r)*W+x+c, g[r*tileSize+c])
		}
	}
}

// random: the driver's own generator (every caller enters with the carry clear).
func (d *Driver) random() byte {
	s := &d.seed
	a := s[3] + s[2]
	s[2] = a
	s[1] ^= a
	a = s[1] + s[0]
	a = a>>1 | a<<7
	s[0] = a
	s[3] = a
	return a
}

func (d *Driver) present() {
	if d.Present != nil {
		d.Present()
	}
}

// --- the driver's entries ----------------------------------------------------------------

// Init is initDriver: the exit callback and a clear screen.
func (d *Driver) Init(exit func(int)) {
	d.Exit = exit
	d.Clear()
}

// Clear is clearScreen.
func (d *Driver) Clear() { clear(d.Pix) }

// DrawCursorBox ORs a white 16x16 frame (2 pixels thick) around map cell (y, x): the cursor.
func (d *Driver) DrawCursorBox(y, x int) {
	const thickness = 2
	x0, y0 := x*tileSize+mapLeft, wrap16(y*tileSize+mapTop)
	for r := 0; r < tileSize; r++ {
		for c := 0; c < tileSize; c++ {
			if r < thickness || r >= tileSize-thickness || c < thickness || c >= tileSize-thickness {
				if d.vga { // VGA: set
					d.set((y0+r)*W+x0+c, white)
					continue
				}
				d.or((y0+r)*W+x0+c, white)
			}
		}
	}
}

// DrawTileRow draws row trow of a tile at screen row srow, x = (scol/2)*8 (VGA: scol*4), 1x or
// (exp != 1) 2x wide.
func (d *Driver) DrawTileRow(tile, trow, srow, scol, exp int) {
	i := srow*W + (scol>>1)*8
	if d.vga {
		i = srow*W + scol*4
	}
	row := d.Tiles[byte(tile)][trow*tileSize:]
	for c := 0; c < tileSize; c++ {
		if exp == 1 {
			d.set(i+c, row[c])
		} else {
			d.set(i+2*c, row[c])
			d.set(i+2*c+1, row[c])
		}
	}
}

// FillTiles fills h rows of w tiles with one tile, from pixel row row, x = (col/2)*8.
func (d *Driver) FillTiles(h, w, tile, row, col int) {
	for r := row; ; r += tileSize {
		for c := 0; c < int(byte(w)); c++ {
			d.Tile(tile, (col/2+c*2)*glyphSize, r)
		}
		h--
		if uint16(h) == 0 { // a count of 0 runs on through 65536
			break
		}
	}
}

// shake moves the map window (x 8..183, rows 8..185) 2 rows down or up, a row at a time,
// clicking the speaker at random when sound is on.
func (d *Driver) shake(sound, down bool) {
	const rows = 173 // the 2-row move stops short of the window's last row
	on := false
	for k := 0; k < rows; k++ {
		if down {
			r := rows - k
			d.copyRow(r+mapTop+2, r+mapTop, mapLeft, mapSize)
		} else {
			d.copyRow(k+mapTop, k+mapTop+2, mapLeft, mapSize)
		}
		if sound && d.random()%2 != 0 {
			on = !on
			if d.Speaker != nil {
				d.Speaker(on)
			}
		}
	}
	if d.Speaker != nil {
		d.Speaker(false)
	}
	d.present()
}

// ShakeDown moves the map window 2 rows down (the earthquake).
func (d *Driver) ShakeDown(sound bool) { d.shake(sound, true) }

// ShakeUp moves the map window 2 rows up.
func (d *Driver) ShakeUp(sound bool) { d.shake(sound, false) }

// DrawMap draws the map window: h rows of 11 tiles from the tile numbers in src, starting at
// pixel (8, 8). The width argument is not used.
func (d *Driver) DrawMap(h, w int, src []byte) {
	k := 0
	for r := 0; uint16(h) != 0; r++ {
		for c := 0; c < mapTiles; c++ {
			t := 0
			if k < len(src) {
				t = int(src[k])
			}
			k++
			d.Tile(t, mapLeft+c*tileSize, mapTop+r*tileSize)
		}
		h--
	}
}

func dot(y, x int) int { return wrap16(y)*W + wrap16(x) }

// Dot is plot: pixel (x, y) set to a colour.
func (d *Driver) Dot(y, x, col int) { d.set(dot(y, x), colorOf(col)) }

// DotOR is plotOR.
func (d *Driver) DotOR(y, x, col int) { d.or(dot(y, x), colorOf(col)) }

// DotXOR is plotXOR.
func (d *Driver) DotXOR(y, x, col int) { d.xor(dot(y, x), colorOf(col)) }

// CR is newLine: the text window (x 192..319, rows 96..191) scrolls up one line and the
// bottom line (rows 184..191) is cleared.
func (d *Driver) CR() {
	for r := textTop; r < lastTextRow; r++ {
		d.copyRow(r, r+glyphSize, sideLeft, textWidth)
	}
	d.fill(sideLeft, lastTextRow, textWidth, glyphSize, 0)
}

// inv is the inverting XOR: colour 7, VGA all bits.
func (d *Driver) inv() byte {
	if d.vga {
		return vgaInvertMask
	}
	return egaInvertMask
}

// InvertMap inverts the map window (XOR inv, x 8..183, rows 8..183).
func (d *Driver) InvertMap() { d.invert(mapLeft, mapTop, mapSize, mapSize, d.inv()) }

// ClearMap clears the map window.
func (d *Driver) ClearMap() { d.fill(mapLeft, mapTop, mapSize, mapSize, 0) }

// InvertStatusLine inverts party member a's line in the status box (XOR 7).
func (d *Driver) InvertStatusLine(a int) {
	d.invert(sideLeft, wrap16((a+1)*glyphSize), statusWidth, glyphSize, d.inv())
}

// InvertCell inverts one character cell (x, y) (XOR 7). The game never calls it.
func (d *Driver) InvertCell(x, y int) {
	d.invert(x*glyphSize, wrap16(y*glyphSize), glyphSize, glyphSize, d.inv())
}

// ClearStatus clears the status box (x 192..311, rows 8..71).
func (d *Driver) ClearStatus() { d.fill(sideLeft, mapTop, statusWidth, statusHeight, 0) }

// LoadPicture loads an RLE picture (the byte runMark, a count, a value; every byte two pixels, high nibble
// first). With mem == nil the picture is ORed onto the screen, its first row starting at
// x = xoff*8; otherwise it is unpacked into mem (W*H pixels). In VGA mode the patch's
// picture is used (one byte a pixel, XORed); a picture it lacks comes in EGA colours.
func (d *Driver) LoadPicture(name string, xoff int, mem []byte) {
	if d.vga {
		if data, ok := upgrade.Picture(name); ok {
			d.gra16VGA(data, xoff, mem)
			return
		}
	}
	data, ok := d.ReadFile(name)
	if !ok {
		d.Exit(3)
		return
	}
	si := 0
	next := func() byte {
		var b byte
		if si < len(data) {
			b = data[si]
		}
		si++
		return b
	}
	if mem != nil {
		clear(mem)
		i := xoff * glyphSize
		for n := W * H; ; {
			cnt, b := 1, next()
			if b == runMark {
				cnt, b = int(next()), next()
				if cnt == 0 {
					cnt = 256
				}
			}
			for ; cnt > 0; cnt-- {
				if i >= 0 && i+1 < len(mem) {
					mem[i], mem[i+1] = d.Color(b>>4), d.Color(b%16)
				}
				i += 2
				if n -= 2; n == 0 {
					return
				}
			}
		}
	}
	i, x, y := xoff*glyphSize, 0, 0
	for {
		cnt, b := 1, next()
		if b == runMark {
			cnt, b = int(next()), next()
			if cnt == 0 {
				cnt = 65536 // the screen loop's 16-bit count
			}
		}
		for ; cnt > 0; cnt-- {
			d.or(i+x, d.Color(b>>4))
			d.or(i+x+1, d.Color(b%16))
			if x += 2; x >= W {
				if y++; y == H {
					return
				}
				i, x = y*W, 0
			}
		}
	}
}

// gra16VGA is the VGA driver's LoadPicture: RLE bytes XORed onto the screen (or mem, cleared
// first) from offset xoff, until 320x200 pixels are done (a run may go past).
func (d *Driver) gra16VGA(data []byte, xoff int, mem []byte) {
	dst := d.Pix
	if mem != nil {
		clear(mem)
		dst = mem
	}
	si := 0
	next := func() byte {
		var b byte
		if si < len(data) {
			b = data[si]
		}
		si++
		return b
	}
	for i := 0; i < W*H; {
		cnt, b := 1, next()
		if b == runMark {
			cnt, b = int(next()), next()
			if cnt == 0 {
				cnt = 65536
			}
		}
		for ; cnt > 0; cnt-- {
			if j := xoff + i; j >= 0 && j < len(dst) {
				dst[j] ^= b
			}
			i++
		}
	}
}

// runMark starts a run in an RLE picture: the count and the repeated byte follow.
const runMark = 2

// OpenDoors opens the map window like a pair of doors onto a picture (the shrine/mystic views):
// the old picture slides out to both sides 8 pixels a step while the new one grows from
// the middle, 11 steps (VGA: a pixel a step, 87 steps).
func (d *Driver) OpenDoors(name string) {
	buf := make([]byte, W*H)
	d.LoadPicture(name, 0, buf)
	if d.vga {
		for n := mapSize/2 - 1; n > 0; n-- {
			for r := mapTop; r < mapBottom; r++ {
				row, src := d.Pix[r*W:r*W+W], buf[r*W:r*W+W]
				copy(row[mapLeft:mapLeft+n], row[mapLeft+1:mapLeft+1+n])
				copy(row[mapRight-n:mapRight], row[mapRight-1-n:mapRight-1])
				copy(row[mapLeft+n:mapRight-n], src[mapLeft+n:mapRight-n])
			}
			d.present()
		}
		for r := mapTop; r < mapBottom; r++ {
			d.Pix[r*W+mapLeft], d.Pix[r*W+mapRight-1] = buf[r*W+mapLeft], buf[r*W+mapRight-1]
		}
		d.present()
		return
	}
	// EGA: in 8-pixel columns, the doors meeting at column 12 (the window's middle)
	const middle, steps, left, right = 12, 11, mapLeft / 8, mapRight / 8
	for step := 1; step <= steps; step++ {
		for r := mapTop; r < mapBottom; r++ {
			row := d.Pix[r*W : r*W+W]
			if step < steps {
				copy(row[left*8:(middle-step)*8], row[(left+1)*8:(middle+1-step)*8])
				copy(row[(middle+step)*8:right*8], row[(middle-1+step)*8:(right-1)*8])
			}
			copy(row[(middle-step)*8:(middle+step)*8], buf[r*W+(middle-step)*8:])
		}
		d.present()
	}
}

// FillQuad fills a quadrilateral column by column: from x0 towards x1 (x1 itself not drawn),
// the column runs from the edge y0->y0p to the edge y1->y1p (Bresenham steps). It ORs colour
// col, or clears when erase is set. The 3D dungeon walls.
func (d *Driver) FillQuad(x0, x1, y0, y1, y0p, y1p, col, erase int) {
	sign := func(v int16) int16 {
		if v == 0 {
			return 0
		}
		if v < 0 {
			return -1
		}
		return 1
	}
	abs := func(v int16) int16 {
		if v < 0 {
			return -v
		}
		return v
	}
	x := int16(x0)
	ya, yb := int16(y0), int16(y1)
	dx := int16(x1) - int16(x0)
	sx := sign(dx)
	sa := sign(int16(y0p) - int16(y0))
	sb := sign(int16(y1p) - int16(y1))
	ea := int16(uint16(abs(dx)) >> 1)
	eb := ea
	for {
		px := int(uint16(x))
		n := int(uint16(yb - ya + 1))
		for i := 0; i < n; i++ {
			// A column whose edges crossed (yb < ya) runs on for some 65535 rows: its offset
			// wraps around video memory's 64K and comes back over the screen.
			p := wrap16((int(uint16(ya))+i)*W/8+px/8)*8 + px%8
			if byte(erase) != 0 {
				d.set(p, 0)
			} else {
				d.or(p, colorOf(col))
			}
		}
		x += sx
		ea -= abs(int16(y0p) - int16(y0))
		eb -= abs(int16(y1p) - int16(y1))
		// The driver keeps |dx| for the bottom edge only when it set it for the top edge;
		// otherwise its register still holds the column's byte offset (x/8), and that is
		// what the bottom edge adds.
		step := int16(px / 8)
		if ea < 0 {
			ya += sa
			step = abs(dx)
			ea += step
		}
		if eb < 0 {
			yb += sb
			eb += step
		}
		if x == int16(x1) {
			break
		}
	}
}

// The ankh glyphs: the whole one, and the status line's, built from the virtues.
const (
	glyphFullAnkh = 0
	glyphAnkh     = 7
)

// BuildAnkh builds glyph 7 (the ankh in the status line) row by row from glyph 0: row i is
// copied when *karma[i] == 0 (that virtue reached its highest point) and blanked otherwise.
func (d *Driver) BuildAnkh(karmas [8]uint16) {
	for r := 0; r < 8; r++ {
		row := d.Glyphs[glyphAnkh][r*glyphSize : (r+1)*glyphSize]
		if karmas[r] == 0 {
			copy(row, d.Glyphs[glyphFullAnkh][r*glyphSize:(r+1)*glyphSize])
		} else {
			clear(row)
		}
	}
}

// flamePixelsVGA are the VGA driver's flame pixels of the campfire tile, {row, column}.
var flamePixelsVGA = [30][2]int{
	{7, 6}, {7, 7}, {8, 5}, {8, 6}, {9, 4}, {9, 5}, {9, 6}, {9, 7}, {10, 4}, {10, 5}, {10, 6},
	{10, 7}, {10, 8}, {11, 5}, {11, 6}, {11, 7}, {11, 8}, {11, 9}, {12, 4}, {12, 5}, {12, 6},
	{12, 7}, {12, 8}, {12, 9}, {13, 5}, {13, 6}, {13, 7}, {13, 8}, {14, 6}, {14, 7},
}

// FlickerFlames makes the campfire's flames flicker: colour bit 4 flipped at random in rows
// 7..14. VGA: each of 30 flame pixels changes one time in 32.
func (d *Driver) FlickerFlames() {
	if d.vga {
		t := &d.Tiles[tileCampfire]
		for i := len(flamePixelsVGA) - 1; i >= 0; i-- {
			if d.random()%32 == 0 {
				// ponytail: the driver XORs with a 32-byte colour table indexed by the
				// pixel's colour, which for the VGA tile's colours reads on into its own
				// code; the port flips bit 2 (a shade within the palette's ramp) instead.
				p := flamePixelsVGA[i]
				t[p[0]*tileSize+p[1]] ^= 4
			}
		}
		return
	}
	// which pixels of each flame row may flicker: {right half, left half}
	masks := [8][2]byte{
		{0b00000000, 0b00001100},
		{0b10000000, 0b00001101},
		{0b11000000, 0b00011111},
		{0b11000000, 0b00011111},
		{0b10000000, 0b00001111},
		{0b11000000, 0b00011111},
		{0b01000000, 0b00010010},
		{0b11000000, 0b00011000},
	}
	const firstRow, half = 7, 8
	t := &d.Tiles[tileCampfire]
	for i, m := range masks {
		right := d.random() & m[0]
		left := d.random() & m[1]
		row := t[(firstRow+i)*tileSize:]
		for k := 0; k < half; k++ {
			if left&(128>>k) != 0 {
				row[k] ^= 4
			}
			if right&(128>>k) != 0 {
				row[half+k] ^= 4
			}
		}
	}
}

// FlowTile scrolls a tile down one row (water, lava), the last row coming back on top.
func (d *Driver) FlowTile(tile int) {
	if tile < 0 || tile >= len(d.Tiles) {
		return
	}
	t := &d.Tiles[tile]
	var last [16]byte
	copy(last[:], t[(tileSize-1)*tileSize:])
	copy(t[tileSize:], t[:(tileSize-1)*tileSize])
	copy(t[:tileSize], last[:])
}

// The tiles with flags.
const (
	tileTowne                   = 10
	tileCastle                  = 11
	tileCastleBritanniaEntrance = 14
	tileShipWest                = 16
	tileShipEast                = 18
)

// flagRows are the flag tiles' waving rows: {tile, row}; row and row+1 swap.
var flagRows = [5][2]int{{tileTowne, 3}, {tileCastle, 1}, {tileShipWest, 2}, {tileShipEast, 2}, {tileCastleBritanniaEntrance, 1}}

// flagsVGA are the VGA driver's: {tile, row, column, pixels} swapped with the row below.
var flagsVGA = [5][4]int{{tileTowne, 3, 4, 6}, {tileCastle, 1, 4, 4}, {tileCastleBritanniaEntrance, 1, 4, 8}, {tileShipWest, 2, 10, 4}, {tileShipEast, 2, 7, 4}}

// WaveFlag swaps the two rows of flag i (0..4).
func (d *Driver) WaveFlag(i int) {
	a, n := d.Tiles[flagRows[i][0]][flagRows[i][1]*tileSize:], tileSize
	if d.vga {
		f := flagsVGA[i]
		a, n = d.Tiles[f[0]][f[1]*tileSize+f[2]:], f[3]
	}
	for k := 0; k < n; k++ {
		a[k], a[tileSize+k] = a[tileSize+k], a[k]
	}
}

// WaveFlags waves the flags, each when the generator comes out non-negative (VGA: one time
// in 16).
func (d *Driver) WaveFlags() {
	for i := range flagRows {
		if r := d.random(); d.vga && r%16 == 0 || !d.vga && r < 128 {
			d.WaveFlag(i)
		}
	}
}

// PutChar is putGlyph: character c at text cell (x, y); past the font it is blank.
func (d *Driver) PutChar(y, x int, c byte) {
	var g [64]byte
	if int(c) < len(d.Glyphs) {
		g = d.Glyphs[c]
	}
	i := wrap16(y*glyphSize)*W + x*glyphSize
	for r := 0; r < glyphSize; r++ {
		for k := 0; k < glyphSize; k++ {
			d.set(i+r*W+k, g[r*glyphSize+k])
		}
	}
}
