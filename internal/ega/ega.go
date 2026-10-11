// Package ega is a port of Ultima IV's EGA graphics driver (EGA.DRV), the 4.7 KB module
// AVATAR.EXE loads at run time and calls through the U4_GRAPH.ASM thunks (Gra_00..Gra_putchar).
//
// The driver drew into the EGA's four bit planes (mode 0Dh, 320x200, 40 bytes a row) with the
// sequencer map mask, the graphics controller's function select (OR/XOR/AND) and bit mask, and
// write mode 1 latch copies. The port keeps the four planes as byte arrays and does exactly
// those operations on them, so every routine stays byte for byte what the driver wrote; Pixels
// turns the planes into a 320x200 picture of colour indexes 0..15.
//
// The tile set and the font live in the game's memory (pShapes, pCharset) in the planar layout
// sizzleShapes / sizzleCharset left them in: a tile is 128 bytes, four 32-byte planes (plane 3
// first) of 16 rows x 2 bytes; a character is 32 bytes, four 8-byte planes (plane 3 first).
// The animations (Gra_animFlow, Gra_animFlag, Gra_animSpit, Gra_19) edit those bytes in place.
//
// Offsets in the comments are EGA.DRV's (the jump table at 0000h, the routines after 0277h).
package ega

const (
	W        = 320
	H        = 200
	rowBytes = 40
	planeLen = 0x10000 // offsets are 16-bit; keep the whole segment so stray writes wrap as on DOS
)

// Driver is the loaded EGA.DRV with its video memory.
type Driver struct {
	Planes [4][]byte // video memory, plane 0..3

	Shapes  []byte // the game's pShapes buffer (0x8000 bytes, planar after sizzleShapes)
	Charset []byte // the game's pCharset buffer (0x1000 bytes, planar after sizzleCharset)

	// ReadFile returns a file's contents for Gra_16/Gra_17 (the driver's own dopen/dread);
	// ok false is the driver's error path.
	ReadFile func(name string) ([]byte, bool)
	// Exit is the pExit callback Gra_init got (the game's exit), called with the error code.
	Exit func(code int)
	// Speaker toggles the PC speaker (the shake's random clicks, port 61h bit 1).
	Speaker func(on bool)
	// Present lets the host show the screen inside a routine that animates on its own (shake,
	// Gra_17's opening doors); DOS drew those straight to the screen.
	Present func()

	seed [4]byte // 0036h: random seed 12h,34h,56h,78h

	speakerOn bool
}

// New returns the driver in the state EGA.DRV is loaded in.
func New() *Driver {
	d := &Driver{seed: [4]byte{0x12, 0x34, 0x56, 0x78}}
	for p := range d.Planes {
		d.Planes[p] = make([]byte, planeLen)
	}
	return d
}

// line is the driver's line table (0052h, filled by Gra_init): row * 40.
func line(row int) int { return row * rowBytes }

// colorMask (01F3h) maps the game's dot and wall colours (CGA's 0..3) to EGA plane masks;
// indexes past the table read the zero bytes after it.
func colorMask(c int) byte {
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

// Palette is the EGA palette Gra_init sets (01E2h: the default 0..7, 38h..3Fh) as 6-bit
// VGA DAC values, for colour indexes 0..15.
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

// Pixels writes the screen as colour indexes (0..15) into pix (320*200 bytes, top row first).
func (d *Driver) Pixels(pix []byte) {
	p0, p1, p2, p3 := d.Planes[0], d.Planes[1], d.Planes[2], d.Planes[3]
	for y := 0; y < H; y++ {
		for xb := 0; xb < rowBytes; xb++ {
			o := y*rowBytes + xb
			b0, b1, b2, b3 := p0[o], p1[o], p2[o], p3[o]
			out := pix[y*W+xb*8:]
			for bit := 0; bit < 8; bit++ {
				m := byte(0x80) >> bit
				var c byte
				if b0&m != 0 {
					c |= 1
				}
				if b1&m != 0 {
					c |= 2
				}
				if b2&m != 0 {
					c |= 4
				}
				if b3&m != 0 {
					c |= 8
				}
				out[bit] = c
			}
		}
	}
}

// --- planar write helpers ---------------------------------------------------------------

// put writes v to the planes in mask (map mask, write mode 0, replace, bit mask FFh).
func (d *Driver) put(mask byte, off int, v byte) {
	off &= 0xffff
	for p := 0; p < 4; p++ {
		if mask&(1<<p) != 0 {
			d.Planes[p][off] = v
		}
	}
}

// or sets the bits of bm on the planes in mask: write mode 0 writing FFh through the bit
// mask (replace, or function OR) - the driver's "or byte [es:di],0ffh".
func (d *Driver) or(mask byte, off int, bm byte) {
	off &= 0xffff
	for p := 0; p < 4; p++ {
		if mask&(1<<p) != 0 {
			d.Planes[p][off] |= bm
		}
	}
}

// xor flips the bits of bm on the planes in mask (function select XOR).
func (d *Driver) xor(mask byte, off int, bm byte) {
	off &= 0xffff
	for p := 0; p < 4; p++ {
		if mask&(1<<p) != 0 {
			d.Planes[p][off] ^= bm
		}
	}
}

// latch copies a byte on all four planes (write mode 1: read loads the latches, write
// stores them).
func (d *Driver) latch(dst, src int) {
	dst &= 0xffff
	src &= 0xffff
	for p := 0; p < 4; p++ {
		d.Planes[p][dst] = d.Planes[p][src]
	}
}

func rd(b []byte, off int) byte {
	if off < 0 || off >= len(b) {
		return 0
	}
	return b[off]
}

func wr(b []byte, off int, v byte) {
	if off >= 0 && off < len(b) {
		b[off] = v
	}
}

// random (0EC6h): the driver's own generator. Every caller enters with the carry clear (the
// thunk's last add, a cmp with 0, an and), so the first adc is an add, and so is the second
// (after an xor).
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

// --- the jump table ----------------------------------------------------------------------

// Init is Gra_init (0E2Eh): mode 0Dh, the palette, and the pointers to the tiles, the font
// and the exit callback.
func (d *Driver) Init(shapes, charset []byte, exit func(int)) {
	d.Shapes, d.Charset, d.Exit = shapes, charset, exit
	d.ClrScr()
}

// ClrScr is Gra_clrscr (0EA5h): all four planes zero.
func (d *Driver) ClrScr() {
	for p := range d.Planes {
		clear(d.Planes[p][:0x1f40])
	}
}

// Gra00 (11C6h) ORs a white 16x16 frame (2 pixels thick) around map cell (y, x): the cursor.
func (d *Driver) Gra00(y, x int) {
	di := line((y*16+8)&0xffff) + x*2 + 1
	for i := 0; i < 2; i++ {
		d.or(0xf, di, 0xff)
		d.or(0xf, di+1, 0xff)
		di += rowBytes
	}
	for i := 0; i < 12; i++ {
		d.or(0xf, di+i*rowBytes, 0xc0)
		d.or(0xf, di+1+i*rowBytes, 0x03)
	}
	di += 12 * rowBytes
	for i := 0; i < 2; i++ {
		d.or(0xf, di, 0xff)
		d.or(0xf, di+1, 0xff)
		di += rowBytes
	}
}

// Gra01 (0277h) draws row trow of a tile at screen row srow, x = scol*4, 1x or (exp != 1)
// 2x wide.
func (d *Driver) Gra01(tile, trow, srow, scol, exp int) {
	si := tile*128 + trow*2
	for p := 3; p >= 0; p-- {
		di := scol>>1 + line(srow)
		for i := 0; i < 2; i++ {
			al := rd(d.Shapes, si)
			si++
			if exp == 1 {
				d.put(1<<p, di, al)
				di++
				continue
			}
			var bx uint16
			for b := 7; b >= 0; b-- {
				bit := uint16(al>>b) & 1
				bx = bx<<2 | bit<<1 | bit
			}
			d.put(1<<p, di, byte(bx>>8))
			d.put(1<<p, di+1, byte(bx))
			di += 2
		}
		si += 0x20 - 2
	}
}

// drawTile copies a 16x16 tile (128 planar bytes at si) to di.
func (d *Driver) drawTile(si, di int) {
	for p := 3; p >= 0; p-- {
		for r := 0; r < 16; r++ {
			d.put(1<<p, di+r*rowBytes, rd(d.Shapes, si))
			d.put(1<<p, di+r*rowBytes+1, rd(d.Shapes, si+1))
			si += 2
		}
	}
}

// Gra02 (0350h) fills h rows of w tiles with one tile, from pixel row row, x = col*4.
func (d *Driver) Gra02(h, w, tile, row, col int) {
	si := tile*128 + 0 // the tile is reread for every cell
	r := row
	for h = h & 0xffff; ; {
		di := col>>1 + line(r)
		for c := 0; c < w&0xff; c++ {
			d.drawTile(si, di)
			di += 2
		}
		r += 16
		h--
		if h&0xffff == 0 {
			break
		}
	}
}

func (d *Driver) click(sound bool, dh *bool) {
	if sound && d.random()&1 != 0 {
		*dh = !*dh
		if d.Speaker != nil {
			d.Speaker(*dh)
		}
	}
}

// Gra03 (041Dh, pascal) moves the map window 2 rows down (the earthquake), clicking the
// speaker at random when sound is on.
func (d *Driver) Gra03(sound bool) {
	dh := false
	for bx := 0x15a; bx != 0; bx -= 2 {
		di := line(bx/2+10) + 1
		si := line(bx/2+8) + 1
		for i := 0; i < 0x16; i++ {
			d.latch(di+i, si+i)
		}
		d.click(sound, &dh)
	}
	if d.Speaker != nil {
		d.Speaker(false)
	}
	d.present()
}

// Gra04 (048Fh, pascal) moves the map window 2 rows up.
func (d *Driver) Gra04(sound bool) {
	dh := false
	for bx := 0; bx != 0x15a; bx += 2 {
		di := line(bx/2+8) + 1
		si := line(bx/2+10) + 1
		for i := 0; i < 0x16; i++ {
			d.latch(di+i, si+i)
		}
		d.click(sound, &dh)
	}
	if d.Speaker != nil {
		d.Speaker(false)
	}
	d.present()
}

// Gra05 (0505h) draws the map window: h rows of 11 tiles from the tile numbers in src,
// starting at pixel (8, 8). The width argument is not used.
func (d *Driver) Gra05(h, w int, src []byte) {
	k := 0
	for r := 0; h&0xffff != 0; r++ {
		di := line(8+r*16) + 1
		for c := 0; c < 11; c++ {
			t := 0
			if k < len(src) {
				t = int(src[k])
			}
			k++
			d.drawTile(t*128, di)
			di += 2
		}
		h--
	}
}

// Dot is Gra_dot (088Eh): pixel (x, y) set to a colour (AND 0 on all planes, then OR the
// colour's planes).
func (d *Driver) Dot(y, x, col int) {
	bm := byte(0x80) >> (x & 7)
	di := line(y&0xffff) + (x&0xffff)>>3
	off := di & 0xffff
	for p := 0; p < 4; p++ {
		d.Planes[p][off] &^= bm
	}
	d.or(colorMask(col), di, bm)
}

// DotOR is Gra_dot_OR (07C4h).
func (d *Driver) DotOR(y, x, col int) {
	d.or(colorMask(col), line(y&0xffff)+(x&0xffff)>>3, byte(0x80)>>(x&7))
}

// DotXOR is Gra_dot_XOR (081Dh).
func (d *Driver) DotXOR(y, x, col int) {
	d.xor(colorMask(col), line(y&0xffff)+(x&0xffff)>>3, byte(0x80)>>(x&7))
}

// CR is Gra_CR (05D3h): the text window (x 192..319, rows 96..191) scrolls up one line and
// the bottom line (rows 184..191) is cleared.
func (d *Driver) CR() {
	for r := 0; r < 0x58; r++ {
		di := line(96+r) + 0x18
		si := line(104+r) + 0x18
		for i := 0; i < 0x10; i++ {
			d.latch(di+i, si+i)
		}
	}
	for r := 0; r < 8; r++ {
		di := line(184+r) + 0x18
		for i := 0; i < 0x10; i++ {
			d.put(0xf, di+i, 0)
		}
	}
}

// Gra09 (0646h) inverts the map window (XOR colour 7, x 8..183, rows 8..183).
func (d *Driver) Gra09() {
	for r := 0; r < 0xb0; r++ {
		di := line(8+r) + 1
		for i := 0; i < 0x16; i++ {
			d.xor(7, di+i, 0xff)
		}
	}
}

// Gra10 (0694h) clears the map window.
func (d *Driver) Gra10() {
	for r := 8; r < 0xb8; r++ {
		di := line(r) + 1
		for i := 0; i < 0x16; i++ {
			d.put(0xf, di+i, 0)
		}
	}
}

// Gra11 (06C9h, pascal) inverts party member a's line in the status box (XOR 7).
func (d *Driver) Gra11(a int) {
	r := ((a + 1) * 8) & 0xffff
	for i := 0; i < 8; i++ {
		di := line(r+i) + 0x18
		for c := 0; c < 0xf; c++ {
			d.xor(7, di+c, 0xff)
		}
	}
}

// Gra12 (0724h) inverts one character cell (x, y) (XOR 7). The game never calls it.
func (d *Driver) Gra12(x, y int) {
	bx := line((y*8)&0xffff) + x
	for i := 0; i < 8; i++ {
		d.xor(7, bx+i*rowBytes, 0xff)
	}
}

// Gra13 (0792h) clears the status box (x 192..311, rows 8..71).
func (d *Driver) Gra13() {
	for r := 8; r < 72; r++ {
		di := line(r) + 0x18
		for i := 0; i < 0xf; i++ {
			d.put(0xf, di+i, 0)
		}
	}
}

// Gra16 is the picture loader (091Dh, pic_inflate): an RLE file (byte 02h, count, value; every
// byte two pixels, high nibble first). With mem == nil the picture is ORed onto the screen,
// starting at byte xoff of row 0; otherwise it is unpacked into mem (32000 bytes: plane 3,
// 2, 1, 0 of 8000 bytes each, cleared first).
func (d *Driver) Gra16(name string, xoff int, mem []byte) {
	data, ok := d.ReadFile(name)
	if !ok {
		d.Exit(3)
		return
	}
	si := 0
	lodsb := func() byte {
		b := rd(data, si)
		si++
		return b
	}
	di := line(0) + xoff
	if mem != nil {
		clear(mem[:min(len(mem), 0x7d00)])
		ah := byte(0x80)
		bx := 0x1f40
		for {
			cx := 1
			al := lodsb()
			ch := byte(0)
			if al == 2 {
				cx = int(lodsb())
				al = lodsb()
				ch = al
			}
			for {
				for half := 0; half < 2; half++ {
					for pl := 0; pl < 4; pl++ {
						if al&0x80 != 0 {
							o := di + pl*0x1f40
							wr(mem, o, rd(mem, o)|ah)
						}
						al <<= 1
					}
					cf := ah&1 != 0
					ah >>= 1
					if half == 1 {
						al = ch
						if cf {
							ah = 0x80
							di++
							bx--
							if bx == 0 {
								return
							}
						}
					} else if cf { // a nibble boundary never ends a byte (ah starts at 80h)
						ah = 0x80
					}
				}
				cx = (cx & 0xff00) | int(byte(cx)-1)
				if byte(cx) == 0 {
					break
				}
			}
		}
	}
	mask := byte(0x80)
	bx, dx := 0, 0
	for {
		cx := 1
		al := lodsb()
		if al == 2 {
			cx = int(lodsb())
			al = lodsb()
		}
		for {
			d.or(al>>4, di+bx, mask)
			mask = mask>>1 | mask<<7
			if mask == 0x80 {
				bx++
			}
			d.or(al&0xf, di+bx, mask)
			mask = mask>>1 | mask<<7
			if mask == 0x80 {
				bx++
			}
			if bx >= 0x28 {
				dx++
				if dx == 0xc8 {
					return
				}
				di = line(dx)
				bx = 0
			}
			cx = (cx - 1) & 0xffff
			if cx == 0 {
				break
			}
		}
	}
}

// Gra17 (0AE3h, pascal) opens the map window like a pair of doors onto a picture (the
// shrine/mystic views): the old picture slides out to both sides while the new one grows
// from the middle, 11 steps. Without memory for the picture it clears the window and draws
// the picture straight to the screen.
func (d *Driver) Gra17(name string) {
	buf := make([]byte, 0x7d00)
	d.Gra16(name, 0, buf)
	for cl := 1; cl <= 0xb; cl++ {
		if cl < 0xb {
			for di := 8; di < 0xb8; di++ {
				si := line(di)
				for bx := 1; bx < 0xc-cl; bx++ {
					d.latch(si+bx, si+bx+1)
				}
				for bx := 0x16; bx >= 0xc+cl; bx-- {
					d.latch(si+bx, si+bx-1)
				}
			}
		}
		for p := 0; p < 4; p++ {
			plane := 3 - p
			for r := 8; r < 0xb8; r++ {
				base := line(r)
				bx := 0xc - cl
				for n := 0; n < 0xb+cl-bx+1; n++ {
					d.Planes[plane][base+bx+n] = rd(buf, p*0x1f40+base+bx+n)
				}
			}
		}
		d.present() // the driver's delay loop (dec ax 65536 times)
	}
}

// Gra18 (0C14h) fills a quadrilateral column by column: from x0 towards x1 (x1 itself not
// drawn), the column runs from the edge y0->y0p to the edge y1->y1p (Bresenham steps). It ORs
// colour col, or clears when erase is set. The 3D dungeon walls.
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
		bm := byte(0x80) >> (uint16(x) & 7)
		si := int(uint16(x) >> 3)
		n := int(uint16(yb - ya + 1))
		for i := 0; i < n; i++ {
			off := line(int(uint16(ya))+i) + si
			if erase&0xff != 0 {
				for p := 0; p < 4; p++ {
					d.Planes[p][off&0xffff] &^= bm
				}
			} else {
				d.or(colorMask(col), off, bm)
			}
		}
		x += sx
		ea -= abs(int16(y0p) - int16(y0))
		eb -= abs(int16(y1p) - int16(y1))
		// The driver keeps |dx| in SI only when it set it for the top edge; otherwise SI
		// still holds the column's byte offset, and that is what the bottom edge adds.
		step := int16(si)
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

// Gra19 (0D39h, pascal) builds character 7 (the ankh in the status line) row by row from
// character 0: row i is shown while virtue i is non-zero... the driver copies the row when
// *karma[i] == 0 (a virtue reached its highest point) and blanks it otherwise.
func (d *Driver) Gra19(karmas [8]uint16) {
	for si := 7; si >= 0; si-- {
		for p := 0; p < 4; p++ {
			if karmas[si] == 0 {
				wr(d.Charset, si+0xe0+p*8, rd(d.Charset, si+p*8))
			} else {
				wr(d.Charset, si+0xe0+p*8, 0)
			}
		}
	}
}

// AnimSpit (0D9Eh) makes the flames of tile 4Bh flicker: random bits flipped in plane 3,
// rows 7..14.
func (d *Driver) AnimSpit() {
	masks := [8][2]byte{{0x00, 0x0c}, {0x80, 0x0d}, {0xc0, 0x1f}, {0xc0, 0x1f}, {0x80, 0x0f}, {0xc0, 0x1f}, {0x40, 0x12}, {0xc0, 0x18}}
	for i, m := range masks {
		ah := d.random()
		al := d.random()
		o := 0x25ae + i*2
		wr(d.Shapes, o, rd(d.Shapes, o)^(al&m[1]))
		wr(d.Shapes, o+1, rd(d.Shapes, o+1)^(ah&m[0]))
	}
}

// AnimFlow (0FACh, pascal) scrolls a tile down one row (water, lava): each plane's rows
// rotate, the last row coming back on top.
func (d *Driver) AnimFlow(tile int) {
	bx := (tile * 128) & 0xffff
	if bx+128 > len(d.Shapes) {
		return
	}
	t := d.Shapes[bx : bx+128]
	var last [4][2]byte
	for p := 0; p < 4; p++ {
		last[p] = [2]byte{t[p*0x20+0x1e], t[p*0x20+0x1f]}
	}
	copy(t[2:], t[:126])
	for p := 0; p < 4; p++ {
		t[p*0x20], t[p*0x20+1] = last[p][0], last[p][1]
	}
}

// AnimFlag (0FFFh) waves the flags: five pairs of words swapped at random in four planes.
func (d *Driver) AnimFlag() {
	for _, o := range []int{0x506, 0x582, 0x804, 0x904, 0x702} {
		if int8(d.random()) < 0 {
			continue
		}
		for p := 0; p < 4; p++ {
			a := o + p*0x20
			b := a + 2
			if b+1 >= len(d.Shapes) {
				continue
			}
			d.Shapes[a], d.Shapes[b] = d.Shapes[b], d.Shapes[a]
			d.Shapes[a+1], d.Shapes[b+1] = d.Shapes[b+1], d.Shapes[a+1]
		}
	}
}

// PutChar is Gra_putchar (114Ch, pascal): character c at text cell (x, y).
func (d *Driver) PutChar(y, x int, c byte) {
	si := int(c) * 32
	bx := line((y*8)&0xffff) + x
	for p := 3; p >= 0; p-- {
		for r := 0; r < 8; r++ {
			d.put(1<<p, bx+r*rowBytes, rd(d.Charset, si))
			si++
		}
	}
}
