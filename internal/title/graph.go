package title

// Port of GRAPH.ASM, EGA entries only (the CGA and Tandy routines of the dispatch table D_03CE
// are dropped): tile maps (Gra_0 to the screen, Gra_C into a picture buffer), dots, the
// window clears, picture copies (Gra_3) and the dissolve (Gra_B), the picture loader
// (Gra_inflate), the tile/flag animations, characters, and the sizzle routines that turn the
// packed tiles and font into planar data.
//
// The screen is an ega.Driver's four planes (mode 0Dh, 40 bytes a row). Picture buffers hold
// four 8000-byte planes, plane 3 first (offsets 0, 1F40h, 3E80h, 5DC0h). After GR_INIT_EGA
// the Lines table is row*40; lines() computes it.

import (
	"github.com/memmaker/u4-go/internal/ega"
	"github.com/memmaker/u4-go/internal/speaker"
)

var drv *ega.Driver

var (
	D_03CC     uint8       = 0x35 // Gra_B's random state
	D_03CD     uint8       = 0x9b
	D_3A86     [0x10]uint8 // C_208A's random seed (only 4 bytes used; _BSS, so zero)
	sizzleBuff [0x80]uint8
	D_3B16     uint16 // inflated file buffer offset
	D_3B1C     uint8  // pixel mask (a word in the original; the low byte is used)
)

// D_32D0: Gra_B's dissolve masks (the low byte of each word is used), by density 0..63.
var D_32D0 = [64]uint16{
	0x0000, 0x0000, 0x0000, 0x0000, 0x0000, 0x0000, 0x0000, 0x0000,
	0x0000, 0x0001, 0xC004, 0x0010, 0x0080, 0x0104, 0x0400, 0x0401,
	0x0100, 0xC3C0, 0x1000, 0xC410, 0x1004, 0x0411, 0x0110, 0x03C4,
	0x1080, 0xC481, 0x0110, 0xC104, 0x1010, 0xD101, 0x1084, 0x13C1,
	0x1010, 0x1304, 0xC484, 0x3CF0, 0x0F3C, 0x0F3F, 0x3C84, 0x3C3F,
	0x10F1, 0x13FC, 0xCFC1, 0x3CF1, 0xCFCF, 0x3CFF, 0xD13F, 0x3FF1,
	0xD3F1, 0xF3FC, 0x3C8F, 0xCF8F, 0x3F3F, 0xFFCF, 0x13FC, 0xCFFF,
	0xFFFF, 0xFFFF, 0xFFFF, 0xFFFF, 0xFFFF, 0xFFFF, 0xFFFF, 0xFFFF,
}

// D_337A: Gra_dot's colour to plane mask; 4 and 5 read on into GR_SEGMENT (A000h after init).
var D_337A = [...]uint8{0x00, 0x02, 0x04, 0x0B, 0x00, 0xA0}

// Cycle model costs (calibration knobs): a tile (64 word copies), a byte of a picture copy,
// a row's set-up, Gra_B's work per byte.
const (
	tileCycles   = 1900
	byteCycles   = 9
	rowCycles    = 60
	bEmptyCycles = 40
	bByteCycles  = 220
)

const planeSize = 0x1f40

func lines(row int) int { return (row * 40) & 0xffff }

func rd(b []byte, off int) byte {
	if off < 0 || off >= len(b) {
		return 0
	}
	return b[off]
}

// wr stores into a picture buffer; past its 8000h bytes the original wrote over whatever
// followed in the far heap, the port drops the byte.
func wr(b []byte, off int, v byte) {
	if off >= 0 && off < len(b) {
		b[off] = v
	}
}

// scr is video memory, plane p, offset off (16-bit).
func scr(p, off int) *byte { return &drv.Planes[p][off&0xffff] }

// Gra_0 (GR_0_EGA) draws an h x w map of tiles (16x16) at pixel row `row`, from
// byte 1 (x 8). The other arguments are the CGA/Tandy routines'.
func Gra_0(h, w int, m []uint8, row, _, _, _ int) {
	k := 0
	n := h * w
	for bx := row; h > 0; h-- {
		di := lines(bx) + 1
		for c := 0; c < w; c++ {
			si := int(m[k]) << 7
			k++
			for p := 3; p >= 0; p-- {
				for r := 0; r < 16; r++ {
					*scr(p, di+r*40) = rd(pShapes, si)
					*scr(p, di+r*40+1) = rd(pShapes, si+1)
					si += 2
				}
			}
			di += 2
		}
		bx += 16
	}
	cpu(n * tileCycles)
}

// Gra_C (GR_C_EGA) is Gra_0 into the picture buffer pic.
func Gra_C(h, w int, m []uint8, row int, pic []byte, _ int) {
	k := 0
	n := h * w
	for bx := row; h > 0; h-- {
		di := lines(bx) + 1
		for c := 0; c < w; c++ {
			si := int(m[k]) << 7
			k++
			for p := 0; p < 4; p++ {
				for r := 0; r < 16; r++ {
					wr(pic, di+r*40, rd(pShapes, si))
					wr(pic, di+r*40+1, rd(pShapes, si+1))
					si += 2
				}
				di += planeSize
			}
			di = (di - 0x7d00 + 2) & 0xffff
		}
		bx += 16
	}
	cpu(n * tileCycles)
}

// Gra_dot (GR_DOT_EGA) ORs a pixel of colour col (0..3) at (x, y).
func Gra_dot(y, x, col int) {
	bm := byte(0x80) >> (x & 7)
	off := lines(y&0xffff) + (x&0xffff)>>3
	mask := D_337A[col]
	for p := 0; p < 4; p++ {
		if mask&(1<<p) != 0 {
			*scr(p, off) |= bm
		}
	}
}

// Gra_2 (GR_2_EGA) clears the view/menu area: rows 104..183, x 8..311.
func Gra_2() {
	for r := 104; r < 184; r++ {
		for i := 1; i <= 38; i++ {
			for p := 0; p < 4; p++ {
				*scr(p, lines(r)+i) = 0
			}
		}
	}
}

// Gra_5 (GR_5_EGA) clears the text area: rows 152..199.
func Gra_5() {
	for r := 152; r < 200; r++ {
		for i := 0; i < 40; i++ {
			for p := 0; p < 4; p++ {
				*scr(p, lines(r)+i) = 0
			}
		}
	}
}

// Gra_3 (GR_3_EGA) copies a w (bytes) x h block of picture pic from (src_x, src_y) to the
// screen at (dst_x, dst_y). rand_pts is the CGA routine's.
func Gra_3(w, h, src_x, src_y int, pic []byte, dst_y, rand_pts, dst_x int) {
	for p, po := 3, 0; p >= 0; p, po = p-1, po+planeSize {
		for i := 0; i < h; i++ {
			di := dst_x + lines(dst_y+i)
			si := lines(src_y+i) + src_x + po
			for c := 0; c < w; c++ {
				*scr(p, di+c) = rd(pic, (si+c)&0xffff)
			}
		}
	}
	cpu(4 * h * (w*byteCycles + rowCycles))
}

// Gra_B (GR_B_EGA) is Gra_3 through a random mask of density rand_pts..rand_pts+7 (of 63),
// bottom row first, skipping bytes blank in the picture; it clicks the speaker when its
// generator overflows past rand_pts (more often as the picture fills in).
func Gra_B(w, h, src_x, src_y int, pic []byte, dst_y, rand_pts, dst_x int) {
	var segs []speaker.Seg
	on := false // port 61h bit 1 (the original restores the port at the end)
	us, cycles := 0.0, 0
	cost := func(c int) {
		cycles += c
		us += float64(c) / (4.77 * float64(speed_info))
	}
	for ; h != 0; h-- {
		di := lines(dst_y+h-1) + dst_x
		si := lines(src_y+h-1) + src_x
		cost(rowCycles)
		for c := 0; c < w; c, si, di = c+1, si+1, di+1 {
			al := rd(pic, si) | rd(pic, si+planeSize) | rd(pic, si+2*planeSize) | rd(pic, si+3*planeSize)
			if al == 0 {
				cost(bEmptyCycles)
				continue
			}
			cost(bByteCycles)
			s := uint16(D_03CC) + 0x1d
			dl := uint8(s)
			dh := dl
			dl = uint8(uint16(dl) + uint16(D_03CD) + s>>8)
			D_03CC, D_03CD = dl, dh
			dh = dl & 7
			if uint16(dl)+uint16(uint8(rand_pts)) > 0xff {
				segs = append(segs, speaker.Seg{On: on, Us: us})
				on, us = !on, 0
			}
			m := uint8(D_32D0[uint8(rand_pts)+dh])
			*scr(3, di) = rd(pic, si) & m
			*scr(2, di) = rd(pic, si+planeSize) & m
			*scr(1, di) = rd(pic, si+2*planeSize) & m
			*scr(0, di) = rd(pic, si+3*planeSize) & m
		}
	}
	segs = append(segs, speaker.Seg{On: on, Us: us})
	speaker.Play(segs)
	cpu(cycles)
}

// Gra_inflate (GR_INFLATE_EGA) decodes picture file name into pic (cleared first): each byte
// is two pixels, spread over the four planes.
func Gra_inflate(name string, pic []byte) {
	D_3B1C = 0x80
	D_3B16 = 0
	clear(pic[:min(len(pic), 0x7d00)])
	CB_ega := func(al uint8) { /*C_199E*/
		di := int(D_3B16)
		ah := D_3B1C
		for half := 0; half < 2; half++ {
			for po := 0; po < 4*planeSize; po += planeSize {
				if al&0x80 != 0 {
					wr(pic, (di+po)&0xffff, rd(pic, (di+po)&0xffff)|ah)
				}
				al <<= 1
			}
			ah >>= 1
		}
		cf := D_3B1C&2 != 0
		D_3B1C = D_3B1C>>2 | D_3B1C<<6
		if cf {
			D_3B16++
		}
	}
	if C_23C3(name, CB_ega) == -1 {
		exit(0x36)
	}
}

// Gra_init (GR_INIT_EGA): mode 0Dh (video memory cleared), the palette (D_3369, the EGA
// default: ega.Palette), the Lines table, then the planar font and tiles.
func Gra_init() {
	for p := range drv.Planes {
		clear(drv.Planes[p])
	}
	sizzleCharset()
	sizzleShapes()
}

func Gra_animFlow(tile int) { /*GR_ANIMFLOW_EGA: same as EGA.DRV's*/
	drv.Shapes = pShapes
	drv.AnimFlow(tile)
}

// Gra_animFlag (GR_ANIMFLAG_EGA) swaps the flag tiles' word pairs, each when C_208A comes
// out non-negative. TITLE.EXE's seed is zero (_BSS) and the carry going in is clear, so the
// generator stays at 0 and every flag swaps every time (EGA.DRV seeds 12345678h instead).
func Gra_animFlag() {
	for _, o := range []int{0x506, 0x582, 0x804, 0x904, 0x702} {
		if int8(C_208A()) < 0 {
			continue
		}
		for p := 0; p < 4; p++ {
			a := o + p*0x20
			pShapes[a], pShapes[a+2] = pShapes[a+2], pShapes[a]
			pShapes[a+1], pShapes[a+3] = pShapes[a+3], pShapes[a+1]
		}
	}
}

// C_208A: the driver's generator (like u_rand_b on its own seed; carry in assumed clear).
func C_208A() uint8 {
	s := &D_3A86
	a := s[3] + s[2]
	s[2] = a
	s[1] ^= a
	a = s[1] + s[0]
	a = a>>1 | a<<7
	s[0] = a
	s[3] = a
	return a
}

func Gra_putchar(c uint8) { /*GR_PUTCHAR_EGA: same as EGA.DRV's*/
	drv.Charset = pCharset
	drv.PutChar(int(txt_Y), int(txt_X), c)
}

func Gra_clrscr() { drv.ClrScr() } /*GR_CLRSCR_EGA*/

// sizzleCharset turns the 128 packed characters (8x8, two pixels a byte) into four 8-byte
// planes each, plane 3 first (C_20B4).
func sizzleCharset() {
	for c := 0; c < 0x80; c++ {
		g := pCharset[c*0x20 : c*0x20+0x20]
		si := 0
		for bx := 0; bx < 8; bx++ {
			for dl := 0; dl < 4; dl++ {
				al := g[si]
				si++
				for k := 0; k < 8; k++ {
					o := bx + (k&3)*8
					sizzleBuff[o] = sizzleBuff[o]<<1 | al>>7
					al <<= 1
				}
			}
		}
		copy(g, sizzleBuff[:0x20])
	}
}

// sizzleShapes does the same for the 256 tiles (16x16): four planes of 16 rows x 2 bytes
// (C_210C).
func sizzleShapes() {
	for t := 0; t < 0x100; t++ {
		g := pShapes[t*0x80 : t*0x80+0x80]
		si := 0
		for bx := 0; bx < 0x20; bx += 2 {
			for dl := 0; dl < 8; dl++ {
				al := g[si]
				si++
				for k := 0; k < 8; k++ {
					o := bx + (k&3)*0x20 // rcl [o+1] then rcl [o]: a 16-bit shift, [o] high
					v := uint16(sizzleBuff[o])<<8 | uint16(sizzleBuff[o+1])
					v = v<<1 | uint16(al>>7)
					sizzleBuff[o], sizzleBuff[o+1] = byte(v>>8), byte(v)
					al <<= 1
				}
			}
		}
		copy(g, sizzleBuff[:])
	}
}

// C_217E decodes file name (SHAPES.EGZ) into buf as is (CB_dummy).
func C_217E(name string, buf []byte) {
	D_3B16 = 0
	CB_dummy := func(al uint8) { /*C_21AD*/
		wr(buf, int(D_3B16), al)
		D_3B16++
	}
	if C_23C3(name, CB_dummy) == -1 {
		exit(0x36)
	}
}
