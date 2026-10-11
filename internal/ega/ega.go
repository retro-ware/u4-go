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

// Driver is the screen, the tile set and the font.
type Driver struct {
	Pix    []byte // W*H colour indexes, top row first
	Tiles  [256][16 * 16]byte
	Glyphs [128][8 * 8]byte

	// ReadFile returns a file's contents for Gra16/Gra17; ok false is the driver's error path.
	ReadFile func(name string) ([]byte, bool)
	// Exit is the game's exit, called with the error code.
	Exit func(code int)
	// Speaker toggles the PC speaker (the shake's random clicks).
	Speaker func(on bool)
	// Present lets the host show the screen inside a routine that animates on its own (shake,
	// Gra17's opening doors).
	Present func()

	// Palette is filled with the mode's 256 colours (6-bit) by SetMode, when set.
	Palette *[768]byte

	seed [4]byte // random seed 12h,34h,56h,78h

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
		return vgaOfEGA[c&0xf]
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
	return &Driver{Pix: pix, seed: [4]byte{0x12, 0x34, 0x56, 0x78}}
}

// unpack decodes packed pixels (two a byte, high nibble first) into dst; missing bytes are 0.
func unpack(dst []byte, packed []byte) {
	for i := range dst {
		var b byte
		if i/2 < len(packed) {
			b = packed[i/2]
		}
		if i&1 == 0 {
			dst[i] = b >> 4
		} else {
			dst[i] = b & 0xf
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
	switch c & 0xff {
	case 1:
		return 2
	case 2:
		return 1
	case 3:
		return 7
	}
	return 0
}

// Palette is the EGA palette (the default 0..7, 38h..3Fh) as 6-bit VGA DAC values, for
// colour indexes 0..15.
var Palette = func() [16][3]byte {
	regs := [16]byte{0, 1, 2, 3, 4, 5, 6, 7, 0x38, 0x39, 0x3a, 0x3b, 0x3c, 0x3d, 0x3e, 0x3f}
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
	g := &d.Tiles[t&0xff]
	for r := 0; r < 16; r++ {
		for c := 0; c < 16; c++ {
			d.set((y+r)*W+x+c, g[r*16+c])
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

// Init is Gra_init: the exit callback and a clear screen.
func (d *Driver) Init(exit func(int)) {
	d.Exit = exit
	d.ClrScr()
}

// ClrScr is Gra_clrscr.
func (d *Driver) ClrScr() { clear(d.Pix) }

// Gra00 ORs a white 16x16 frame (2 pixels thick) around map cell (y, x): the cursor.
func (d *Driver) Gra00(y, x int) {
	x0, y0 := x*16+8, (y*16+8)&0xffff
	for r := 0; r < 16; r++ {
		for c := 0; c < 16; c++ {
			if r < 2 || r >= 14 || c < 2 || c >= 14 {
				if d.vga { // VGA: set
					d.set((y0+r)*W+x0+c, 15)
					continue
				}
				d.or((y0+r)*W+x0+c, 15)
			}
		}
	}
}

// Gra01 draws row trow of a tile at screen row srow, x = (scol/2)*8 (VGA: scol*4), 1x or
// (exp != 1) 2x wide.
func (d *Driver) Gra01(tile, trow, srow, scol, exp int) {
	i := srow*W + (scol>>1)*8
	if d.vga {
		i = srow*W + scol*4
	}
	row := d.Tiles[tile&0xff][trow*16:]
	for c := 0; c < 16; c++ {
		if exp == 1 {
			d.set(i+c, row[c])
		} else {
			d.set(i+2*c, row[c])
			d.set(i+2*c+1, row[c])
		}
	}
}

// Gra02 fills h rows of w tiles with one tile, from pixel row row, x = (col/2)*8.
func (d *Driver) Gra02(h, w, tile, row, col int) {
	for r := row; ; r += 16 {
		for c := 0; c < w&0xff; c++ {
			d.Tile(tile, ((col>>1)+c*2)*8, r)
		}
		h--
		if h&0xffff == 0 {
			break
		}
	}
}

// shake moves the map window (x 8..183, rows 8..185) 2 rows down or up, a row at a time,
// clicking the speaker at random when sound is on.
func (d *Driver) shake(sound, down bool) {
	on := false
	for k := 0; k < 173; k++ {
		if down {
			r := 173 - k
			d.copyRow(r+10, r+8, 8, 176)
		} else {
			d.copyRow(k+8, k+10, 8, 176)
		}
		if sound && d.random()&1 != 0 {
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

// Gra03 moves the map window 2 rows down (the earthquake).
func (d *Driver) Gra03(sound bool) { d.shake(sound, true) }

// Gra04 moves the map window 2 rows up.
func (d *Driver) Gra04(sound bool) { d.shake(sound, false) }

// Gra05 draws the map window: h rows of 11 tiles from the tile numbers in src, starting at
// pixel (8, 8). The width argument is not used.
func (d *Driver) Gra05(h, w int, src []byte) {
	k := 0
	for r := 0; h&0xffff != 0; r++ {
		for c := 0; c < 11; c++ {
			t := 0
			if k < len(src) {
				t = int(src[k])
			}
			k++
			d.Tile(t, 8+c*16, 8+r*16)
		}
		h--
	}
}

func dot(y, x int) int { return (y&0xffff)*W + x&0xffff }

// Dot is Gra_dot: pixel (x, y) set to a colour.
func (d *Driver) Dot(y, x, col int) { d.set(dot(y, x), colorOf(col)) }

// DotOR is Gra_dot_OR.
func (d *Driver) DotOR(y, x, col int) { d.or(dot(y, x), colorOf(col)) }

// DotXOR is Gra_dot_XOR.
func (d *Driver) DotXOR(y, x, col int) { d.xor(dot(y, x), colorOf(col)) }

// CR is Gra_CR: the text window (x 192..319, rows 96..191) scrolls up one line and the
// bottom line (rows 184..191) is cleared.
func (d *Driver) CR() {
	for r := 96; r < 184; r++ {
		d.copyRow(r, r+8, 192, 128)
	}
	d.fill(192, 184, 128, 8, 0)
}

// inv is the inverting XOR: colour 7, VGA FFh.
func (d *Driver) inv() byte {
	if d.vga {
		return 0xff
	}
	return 7
}

// Gra09 inverts the map window (XOR inv, x 8..183, rows 8..183).
func (d *Driver) Gra09() { d.invert(8, 8, 176, 176, d.inv()) }

// Gra10 clears the map window.
func (d *Driver) Gra10() { d.fill(8, 8, 176, 176, 0) }

// Gra11 inverts party member a's line in the status box (XOR 7).
func (d *Driver) Gra11(a int) { d.invert(192, ((a+1)*8)&0xffff, 120, 8, d.inv()) }

// Gra12 inverts one character cell (x, y) (XOR 7). The game never calls it.
func (d *Driver) Gra12(x, y int) { d.invert(x*8, (y*8)&0xffff, 8, 8, d.inv()) }

// Gra13 clears the status box (x 192..311, rows 8..71).
func (d *Driver) Gra13() { d.fill(192, 8, 120, 64, 0) }

// Gra16 loads an RLE picture (byte 02h, count, value; every byte two pixels, high nibble
// first). With mem == nil the picture is ORed onto the screen, its first row starting at
// x = xoff*8; otherwise it is unpacked into mem (W*H pixels). In VGA mode the patch's
// picture is used (one byte a pixel, XORed); a picture it lacks comes in EGA colours.
func (d *Driver) Gra16(name string, xoff int, mem []byte) {
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
		i := xoff * 8
		for n := W * H; ; {
			cnt, b := 1, next()
			if b == 2 {
				cnt, b = int(next()), next()
				if cnt == 0 {
					cnt = 0x100
				}
			}
			for ; cnt > 0; cnt-- {
				if i >= 0 && i+1 < len(mem) {
					mem[i], mem[i+1] = d.Color(b>>4), d.Color(b&0xf)
				}
				i += 2
				if n -= 2; n == 0 {
					return
				}
			}
		}
	}
	i, x, y := xoff*8, 0, 0
	for {
		cnt, b := 1, next()
		if b == 2 {
			cnt, b = int(next()), next()
			if cnt == 0 {
				cnt = 0x10000
			}
		}
		for ; cnt > 0; cnt-- {
			d.or(i+x, d.Color(b>>4))
			d.or(i+x+1, d.Color(b&0xf))
			if x += 2; x >= W {
				if y++; y == H {
					return
				}
				i, x = y*W, 0
			}
		}
	}
}

// gra16VGA is the VGA driver's Gra16: RLE bytes XORed onto the screen (or mem, cleared
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
		if b == 2 {
			cnt, b = int(next()), next()
			if cnt == 0 {
				cnt = 0x10000
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

// Gra17 opens the map window like a pair of doors onto a picture (the shrine/mystic views):
// the old picture slides out to both sides 8 pixels a step while the new one grows from
// the middle, 11 steps (VGA: a pixel a step, 87 steps).
func (d *Driver) Gra17(name string) {
	buf := make([]byte, W*H)
	d.Gra16(name, 0, buf)
	if d.vga {
		for n := 87; n > 0; n-- {
			for r := 8; r < 184; r++ {
				row, src := d.Pix[r*W:r*W+W], buf[r*W:r*W+W]
				copy(row[8:8+n], row[9:9+n])
				copy(row[184-n:184], row[183-n:183])
				copy(row[8+n:184-n], src[8+n:184-n])
			}
			d.present()
		}
		for r := 8; r < 184; r++ {
			d.Pix[r*W+8], d.Pix[r*W+183] = buf[r*W+8], buf[r*W+183]
		}
		d.present()
		return
	}
	for cl := 1; cl <= 11; cl++ {
		for r := 8; r < 184; r++ {
			row := d.Pix[r*W : r*W+W]
			if cl < 11 {
				copy(row[8:(12-cl)*8], row[16:(13-cl)*8])
				copy(row[(12+cl)*8:184], row[(11+cl)*8:176])
			}
			copy(row[(12-cl)*8:(12+cl)*8], buf[r*W+(12-cl)*8:])
		}
		d.present()
	}
}

// Gra18 fills a quadrilateral column by column: from x0 towards x1 (x1 itself not drawn),
// the column runs from the edge y0->y0p to the edge y1->y1p (Bresenham steps). It ORs colour
// col, or clears when erase is set. The 3D dungeon walls.
func (d *Driver) Gra18(x0, x1, y0, y1, y0p, y1p, col, erase int) {
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
			p := ((int(uint16(ya))+i)*W/8+px>>3)&0xffff*8 + px&7
			if erase&0xff != 0 {
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
		step := int16(px >> 3)
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

// Gra19 builds glyph 7 (the ankh in the status line) row by row from glyph 0: row i is
// copied when *karma[i] == 0 (that virtue reached its highest point) and blanked otherwise.
func (d *Driver) Gra19(karmas [8]uint16) {
	for r := 0; r < 8; r++ {
		row := d.Glyphs[7][r*8 : r*8+8]
		if karmas[r] == 0 {
			copy(row, d.Glyphs[0][r*8:r*8+8])
		} else {
			clear(row)
		}
	}
}

// spitVGA: the VGA driver's flame pixels of tile 4Bh.
var spitVGA = [30]byte{0x76, 0x77, 0x85, 0x86, 0x94, 0x95, 0x96, 0x97, 0xa4, 0xa5, 0xa6, 0xa7, 0xa8,
	0xb5, 0xb6, 0xb7, 0xb8, 0xb9, 0xc4, 0xc5, 0xc6, 0xc7, 0xc8, 0xc9, 0xd5, 0xd6, 0xd7, 0xd8, 0xe6, 0xe7}

// AnimSpit makes the flames of tile 4Bh flicker: colour bit 4 flipped at random in rows
// 7..14. VGA: each of 30 flame pixels changes one time in 32.
func (d *Driver) AnimSpit() {
	if d.vga {
		t := &d.Tiles[0x4b]
		for i := len(spitVGA) - 1; i >= 0; i-- {
			if d.random()&0x1f == 0 {
				// ponytail: the driver XORs with a 32-byte colour table indexed by the
				// pixel's colour, which for the VGA tile's colours reads on into its own
				// code; the port flips bit 2 (a shade within the palette's ramp) instead.
				t[spitVGA[i]] ^= 4
			}
		}
		return
	}
	masks := [8][2]byte{{0x00, 0x0c}, {0x80, 0x0d}, {0xc0, 0x1f}, {0xc0, 0x1f}, {0x80, 0x0f}, {0xc0, 0x1f}, {0x40, 0x12}, {0xc0, 0x18}}
	t := &d.Tiles[0x4b]
	for i, m := range masks {
		right := d.random() & m[0]
		left := d.random() & m[1]
		row := t[(7+i)*16:]
		for k := 0; k < 8; k++ {
			if left&(0x80>>k) != 0 {
				row[k] ^= 4
			}
			if right&(0x80>>k) != 0 {
				row[8+k] ^= 4
			}
		}
	}
}

// AnimFlow scrolls a tile down one row (water, lava), the last row coming back on top.
func (d *Driver) AnimFlow(tile int) {
	if tile < 0 || tile >= len(d.Tiles) {
		return
	}
	t := &d.Tiles[tile]
	var last [16]byte
	copy(last[:], t[15*16:])
	copy(t[16:], t[:15*16])
	copy(t[:16], last[:])
}

// flagRows are the flag tiles' waving rows: {tile, row}; row and row+1 swap.
var flagRows = [5][2]int{{0x0a, 3}, {0x0b, 1}, {0x10, 2}, {0x12, 2}, {0x0e, 1}}

// flagsVGA are the VGA driver's: {tile, offset, pixels} swapped with the row below.
var flagsVGA = [5][3]int{{0x0a, 0x34, 6}, {0x0b, 0x14, 4}, {0x0e, 0x14, 8}, {0x10, 0x2a, 4}, {0x12, 0x27, 4}}

// WaveFlag swaps the two rows of flag i (0..4).
func (d *Driver) WaveFlag(i int) {
	a, n := d.Tiles[flagRows[i][0]][flagRows[i][1]*16:], 16
	if d.vga {
		f := flagsVGA[i]
		a, n = d.Tiles[f[0]][f[1]:], f[2]
	}
	for k := 0; k < n; k++ {
		a[k], a[16+k] = a[16+k], a[k]
	}
}

// AnimFlag waves the flags, each when the generator comes out non-negative (VGA: one time
// in 16).
func (d *Driver) AnimFlag() {
	for i := range flagRows {
		if r := d.random(); d.vga && r&0xf == 0 || !d.vga && int8(r) >= 0 {
			d.WaveFlag(i)
		}
	}
}

// PutChar is Gra_putchar: character c at text cell (x, y); past the font it is blank.
func (d *Driver) PutChar(y, x int, c byte) {
	var g [64]byte
	if int(c) < len(d.Glyphs) {
		g = d.Glyphs[c]
	}
	i := ((y*8)&0xffff)*W + x*8
	for r := 0; r < 8; r++ {
		for k := 0; k < 8; k++ {
			d.set(i+r*W+k, g[r*8+k])
		}
	}
}
