package title

// Port of GRAPH.ASM, EGA entries only (the CGA and Tandy routines of the dispatch table D_03CE
// are dropped): tile maps (Gra_0 to the screen, Gra_C into a picture), dots, the window
// clears, picture copies (Gra_3) and the dissolve (Gra_B), the picture loader (Gra_inflate),
// the tile/flag animations and characters.
//
// The screen is an ega.Driver's 320x200 framebuffer of colour indexes; pictures are the same
// (picSize bytes, one a pixel). x positions and widths the game passes are in 8-pixel
// columns, as on the EGA.

import "github.com/memmaker/u4-go/internal/ega"

var drv *ega.Driver

var (
	D_03CC uint8       = 0x35 // Gra_B's random state
	D_03CD uint8       = 0x9b
	D_3A86 [0x10]uint8 // C_208A's random seed (only 4 bytes used; _BSS, so zero)
	D_3B16 uint16      // inflated file buffer offset (Gra_inflate: the pixel)
)

// picSize is a picture buffer's size: one byte a pixel.
const picSize = ega.W * ega.H

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

// D_337A: Gra_dot's colour to colour bits; 4 and 5 read on into GR_SEGMENT (A000h after init).
var D_337A = [...]uint8{0x00, 0x02, 0x04, 0x0B, 0x00, 0xA0}

func rd(b []byte, i int) byte {
	if i < 0 || i >= len(b) {
		return 0
	}
	return b[i]
}

func wr(b []byte, i int, v byte) {
	if i >= 0 && i < len(b) {
		b[i] = v
	}
}

// Gra_0 (GR_0_EGA) draws an h x w map of tiles (16x16) at pixel row `row`, from x 8. The
// other arguments are the CGA/Tandy routines'.
func Gra_0(h, w int, m []uint8, row, _, _, _ int) { Gra_C(h, w, m, row, drv.Pix, 0) }

// Gra_C (GR_C_EGA) is Gra_0 into the picture pic.
func Gra_C(h, w int, m []uint8, row int, pic []byte, _ int) {
	for r, k := 0, 0; r < h; r++ {
		for c := 0; c < w; c, k = c+1, k+1 {
			g := &drv.Tiles[m[k]]
			x, y := 8+c*16, row+r*16
			for ty := 0; ty < 16; ty++ {
				for tx := 0; tx < 16; tx++ {
					wr(pic, (y+ty)*ega.W+x+tx, g[ty*16+tx])
				}
			}
		}
	}
}

// Gra_dot (GR_DOT_EGA) ORs a pixel of colour col (0..3) at (x, y).
func Gra_dot(y, x, col int) {
	i := (y&0xffff)*ega.W + x&0xffff
	wr(drv.Pix, i, rd(drv.Pix, i)|drv.Color(D_337A[col]&0xf))
}

func clearRows(y0, y1, x, w int) {
	for r := y0; r < y1; r++ {
		clear(drv.Pix[r*ega.W+x : r*ega.W+x+w])
	}
}

// Gra_2 (GR_2_EGA) clears the view/menu area: rows 104..183, x 8..311.
func Gra_2() { clearRows(104, 184, 8, 304) }

// Gra_5 (GR_5_EGA) clears the text area: rows 152..199.
func Gra_5() { clearRows(152, 200, 0, ega.W) }

// Gra_3 (GR_3_EGA) copies a w (columns) x h block of picture pic from (src_x, src_y) to the
// screen at (dst_x, dst_y). rand_pts is the CGA routine's.
func Gra_3(w, h, src_x, src_y int, pic []byte, dst_y, rand_pts, dst_x int) {
	for i := 0; i < h; i++ {
		di := (dst_y+i)*ega.W + dst_x*8
		si := (src_y+i)*ega.W + src_x*8
		for c := 0; c < w*8; c++ {
			wr(drv.Pix, di+c, rd(pic, si+c))
		}
	}
}

// Gra_B (GR_B_EGA) is Gra_3 through a random mask of density rand_pts..rand_pts+7 (of 63),
// bottom row first, skipping 8-pixel columns blank in the picture; it clicks the speaker
// when its generator overflows past rand_pts (more often as the picture fills in).
func Gra_B(w, h, src_x, src_y int, pic []byte, dst_y, rand_pts, dst_x int) {
	for ; h != 0; h-- {
		di := (dst_y+h-1)*ega.W + dst_x*8
		si := (src_y+h-1)*ega.W + src_x*8
		for c := 0; c < w; c, si, di = c+1, si+8, di+8 {
			var al byte
			for k := 0; k < 8; k++ {
				al |= rd(pic, si+k)
			}
			if al == 0 {
				continue
			}
			s := uint16(D_03CC) + 0x1d
			dl := uint8(s)
			dh := dl
			dl = uint8(uint16(dl) + uint16(D_03CD) + s>>8)
			D_03CC, D_03CD = dl, dh
			dh = dl & 7
			if uint16(dl)+uint16(uint8(rand_pts)) > 0xff {
				dissolveClicks++
			}
			m := uint8(D_32D0[uint8(rand_pts)+dh])
			for k := 0; k < 8; k++ {
				var v byte
				if m&(0x80>>k) != 0 {
					v = rd(pic, si+k)
				}
				wr(drv.Pix, di+k, v)
			}
		}
	}
}

// Gra_inflate (GR_INFLATE_EGA) decodes picture file name into pic (cleared first): each byte
// is two pixels, high nibble first. (In VGA mode the EGA colours map to the VGA palette's.)
func Gra_inflate(name string, pic []byte) {
	D_3B16 = 0
	clear(pic)
	CB_ega := func(al uint8) { /*C_199E*/
		wr(pic, int(D_3B16), drv.Color(al>>4))
		wr(pic, int(D_3B16)+1, drv.Color(al&0xf))
		D_3B16 += 2
	}
	if C_23C3(name, CB_ega) == -1 {
		fatalf("can't read %s", name)
	}
}

// Gra_init (GR_INIT_EGA): mode 0Dh (screen cleared), the palette (the EGA default:
// ega.Palette), then the font and tiles decoded to pixels.
func Gra_init() {
	drv.ClrScr()
	drv.SetGlyphs(pCharset)
	drv.SetTiles(pShapes)
}

func Gra_animFlow(tile int) { drv.AnimFlow(tile) } /*GR_ANIMFLOW_EGA: same as EGA.DRV's*/

// Gra_animFlag (GR_ANIMFLAG_EGA) waves the flags, each when C_208A comes out non-negative.
// TITLE.EXE's seed is zero (_BSS) and the carry going in is clear, so the generator stays
// at 0 and every flag waves every time (EGA.DRV seeds 12345678h instead).
func Gra_animFlag() {
	for i := 0; i < 5; i++ {
		if int8(C_208A()) >= 0 {
			drv.WaveFlag(i)
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

func Gra_putchar(c uint8) { drv.PutChar(int(txt_Y), int(txt_X), c) } /*GR_PUTCHAR_EGA*/

func Gra_clrscr() { drv.ClrScr() } /*GR_CLRSCR_EGA*/

// SetVGA switches the title to VGA (or EGA) graphics, its pictures included.
func SetVGA(on bool) {
	if drv != nil {
		drv.SetMode(on, pTitle, pAnim, D_6940)
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
		fatalf("can't read %s", name)
	}
}
