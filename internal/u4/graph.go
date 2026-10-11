package u4

// Port of U4_GRAPH.ASM: loading the graphics driver, the screen frame, Gra_line, the
// thunks into the driver (here: the ega package) and the sizzle routines that turn the
// packed EGA tiles and font into the driver's planar layout.

import "github.com/memmaker/u4-go/internal/ega"

// drv is the loaded EGA.DRV.
var drv *ega.Driver

var sizzleBuff [0x80]byte

// load graphic lib: the port has the EGA driver built in; the file must still be there,
// as the original exits (code 49h) without it.
func C_20C1(name string) {
	if _, ok := readGameFile(name); !ok {
		exit(0x49)
	}
}

// draw game screen frame
func C_213B() {
	Gra_clrscr()
	if D_943A == 1 {
		Gra_16("start.pic", 0, nil)
	} else {
		Gra_16("start.ega", 0, nil)
	}
}

func Gra_line(x0, y0, x1, y1, col int) { /*C_215A*/
	if x0 != x1 && y0 != y1 {
		C_A842(x0, y0, x1, y1) // drawLine
		return
	}
	if x0 == x1 {
		si := 0
		if d := int16(y1 - y0); d > 0 {
			si = 1
		} else if d < 0 {
			si = -1
		}
		n := int16(y1 - y0)
		if n < 0 {
			n = -n
		}
		y := y0
		for c := uint16(n + 1); ; {
			Gra_dot_OR(y, x0, col)
			y += si
			c--
			if c == 0 {
				break
			}
		}
		return
	}
	si := 0
	if d := int16(x1 - x0); d > 0 {
		si = 1
	} else if d < 0 {
		si = -1
	}
	n := int16(x1 - x0)
	if n < 0 {
		n = -n
	}
	x := x0
	for c := uint16(n + 1); ; {
		Gra_dot_OR(y0, x, col)
		x += si
		c--
		if c == 0 {
			break
		}
	}
}

// ---- the driver thunks ----

func Gra_00(y, x int)                         { drv.Gra00(y, x) }                  // active character square
func Gra_01(tile, trow, srow, scol, exp int)  { drv.Gra01(tile, trow, srow, scol, exp) }
func Gra_02(h, w, tile, row, col int)         { drv.Gra02(h, w, tile, row, col) }
func Gra_03()                                 { drv.Gra03(SoundFlag != 0); present() }
func Gra_04()                                 { drv.Gra04(SoundFlag != 0); present() }
func Gra_05(h, w int, src []byte, dst int)    { drv.Gra05(h, w, src) } // display tile map
func Gra_dot(y, x, col int)                   { drv.Dot(y, x, col) }
func Gra_CR()                                 { txt_Y = 23; txt_X = 24; drv.CR() }
func Gra_09()                                 { drv.Gra09() }
func Gra_10()                                 { drv.Gra10() }
func Gra_11(a int)                            { drv.Gra11(a) }
func Gra_12(x, y int)                         { drv.Gra12(x, y) } // (uncalled)
func Gra_13()                                 { drv.Gra13() }
func Gra_dot_OR(y, x, col int)                { drv.DotOR(y, x, col) }
func Gra_dot_XOR(y, x, col int)               { drv.DotXOR(y, x, col) }
func Gra_16(name string, xoff int, mem []byte) { drv.Gra16(name, xoff, mem) } // display image
func Gra_18(x0, x1, y0, y1, y0p, y1p, col, erase int) {
	drv.Gra18(x0, x1, y0, y1, y0p, y1p, col, erase)
}
func Gra_animSpit()          { drv.AnimSpit() }
func Gra_clrscr()            { drv.ClrScr() }
func Gra_animFlow(tile int)  { drv.AnimFlow(tile) }
func Gra_animFlag()          { drv.AnimFlag() }
func Gra_putchar(c uint8)    { drv.PutChar(int(txt_Y), int(txt_X), c) }

func Gra_17() {
	if D_943A == 1 {
		drv.Gra17("rune_5.pic")
	} else {
		drv.Gra17("rune_5.ega")
	}
}

// draw the ankh
func Gra_19() {
	var k [8]uint16
	for i, p := range pKarmas {
		k[i] = *p
	}
	drv.Gra19(k)
}

func Gra_init() {
	drv.Init(pShapes, pCharset, exit)
}

// sizzleShapes turns the 256 packed tiles (16x16, two pixels a byte) into the driver's
// planar layout, in place (C_22CF).
func sizzleShapes() {
	b := pShapes
	for t := 0; t < 0x100; t++ {
		sizzle(b[t*0x80:t*0x80+0x80], 0x10, 8, 2, 0x20)
	}
}

// sizzleCharset does the same for the 128 characters (8x8) (C_2341).
func sizzleCharset() {
	b := pCharset
	for c := 0; c < 0x80; c++ {
		sizzle(b[c*0x20:c*0x20+0x20], 8, 4, 1, 8)
	}
}

// sizzle converts one packed glyph (rows of bpr bytes, high nibble first) to four planes
// of rows*w bytes each (plane 3 first): bit 3 of each pixel goes to plane region 0.
func sizzle(g []byte, rows, bpr, w, plane int) {
	clear(sizzleBuff[:])
	si := 0
	for r := 0; r < rows; r++ {
		bx := r * w
		for i := 0; i < bpr; i++ {
			al := g[si]
			si++
			for pix := 0; pix < 2; pix++ {
				for p := 0; p < 4; p++ {
					bit := al >> 7 & 1
					al <<= 1
					if w == 2 { // rcl [bx+1] then rcl [bx]: a 16-bit shift
						v := uint16(sizzleBuff[bx+p*plane])<<8 | uint16(sizzleBuff[bx+p*plane+1])
						v = v<<1 | uint16(bit)
						sizzleBuff[bx+p*plane] = byte(v >> 8)
						sizzleBuff[bx+p*plane+1] = byte(v)
					} else {
						sizzleBuff[bx+p*plane] = sizzleBuff[bx+p*plane]<<1 | bit
					}
				}
			}
		}
	}
	copy(g, sizzleBuff[:len(g)])
}
