package u4

// Port of U4_3D.C: the dungeon's first person view, drawn as wire frames with the driver's
// Gra_18 quads and lines, and the monsters' sprites scaled by distance.

import (
	"time"

	"github.com/memmaker/u4-go/internal/host"
)

var Wall_x0 = [6]uint8{0x08, 0x28, 0x48, 0x58, 0x5F, 0x00} /*D_2D74*/
var Wall_x1 = [6]uint8{0xB7, 0x97, 0x77, 0x67, 0x60, 0x00} /*D_2D7A*/
var Wall_y0 = [6]uint8{0x08, 0x28, 0x48, 0x58, 0x5F, 0x00} /*D_2D80*/
var Wall_y1 = [6]uint8{0xB7, 0x97, 0x77, 0x67, 0x60, 0x00} /*D_2D86*/

var LDoor_x0 = [4]uint8{0x08, 0x30, 0x4C, 0x5A} /*D_2D8C*/
var LDoor_x1 = [4]uint8{0x18, 0x40, 0x54, 0x5C} /*D_2D90*/
var RDoor_x0 = [4]uint8{0xB7, 0x8F, 0x72, 0x64} /*D_2D94*/
var RDoor_x1 = [4]uint8{0xA7, 0x7F, 0x6A, 0x62} /*D_2D98*/
var Door_y0 = [4]uint8{0x34, 0x48, 0x56, 0x5A}  /*D_2D9C*/
var Door_y0p = [4]uint8{0x3D, 0x51, 0x5B, 0x5F} /*D_2DA0*/
var Door_y1 = [4]uint8{0xB7, 0x8F, 0x73, 0x65}  /*D_2DA4*/
var Door_y1p = [4]uint8{0xA7, 0x7F, 0x6B, 0x62} /*D_2DA8*/
var D_2DAC = [4]uint8{0x08, 0x30, 0x4C, 0x5A}
var D_2DB0 = [4]uint8{0x18, 0x40, 0x54, 0x5D}
var D_2DB4 = [4]uint8{0xB7, 0x8F, 0x73, 0x65}
var D_2DB8 = [4]uint8{0xA7, 0x7F, 0x6B, 0x62}
var D_2DBC = [4]uint8{0x35, 0x48, 0x55, 0x5C}
var D_2DC0 = [4]uint8{0x3C, 0x50, 0x5A, 0x5D}
var D_2DC4 = [4]uint8{0x8B, 0x78, 0x6A, 0x62}
var D_2DC8 = [4]uint8{0x83, 0x70, 0x65, 0x61}
var Ladder_x0 = [4]uint8{0x48, 0x54, 0x5C, 0x5F} /*D_2DCC*/
var Ladder_x1 = [4]uint8{0x77, 0x6B, 0x63, 0x60} /*D_2DD0*/
var D_2DD4 = [4]uint8{0x10, 0x38, 0x50, 0x5D}    // up ladder rung 1 y
var D_2DD8 = [4]uint8{0x30, 0x48, 0x56, 0x5E}    // up ladder rung 2 y
var D_2DDC = [4]uint8{0x50, 0x58, 0x5C, 0x5F}    // up ladder rung 3 y
var D_2DE0 = [4]uint8{0xAF, 0x87, 0x6F, 0x62}    // down ladder rung 1 y
var D_2DE4 = [4]uint8{0x8F, 0x77, 0x69, 0x61}    // down ladder rung 2 y
var D_2DE8 = [4]uint8{0x6F, 0x67, 0x63, 0x60}    // down ladder rung 3 y

var D_2DEC, D_2DEE uint16

var (
	STA_A  uint8  // tile_forward_1
	STA_B  uint8  // tile_forward_2
	STA_C  bool   // <backup (Party.f_1dc == 0)>
	STA_D  int16  // <backup Party._z>
	STA_E  int    // current rendering depth
	STA_F  uint8  // <backup Party._x>
	STA_G  uint16 // <backup Party._dir>
	STA_H  uint8  // <backup Party._y>
	STA_I  uint8  // <backup tile_cur>
	STA_J  uint8  // <backup tile_forward>
	STA_K  uint8  // <backup tile_forward_1>
	STA_L  uint8  // <backup tile_forward_2>
	D_8CFA int
)

func u4_abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

func isign(a int) int { return int(u4_sign(int16(a))) }

func w8(a [6]uint8, i int) int { return int(a[i]) }
func q8(a [4]uint8, i int) int { return int(a[i]) }

// render force fields?
func C_A7F2(bp04 uint8) {
	e := STA_E
	Gra_02((w8(Wall_y1, e)-w8(Wall_y0, e)+1)/16, (w8(Wall_x1, e)-w8(Wall_x0, e)+1)/16,
		int(bp04&3)|0x44, w8(Wall_y0, e), w8(Wall_x0, e)>>2)
}

// drawLine
func C_A842(bp0a, bp08, bp06, bp04 int) {
	loc_E := bp0a
	loc_A := bp08
	loc_B := isign(bp06 - bp0a)
	loc_C := isign(bp04 - bp08)
	if u4_abs(bp0a-bp06) < u4_abs(bp08-bp04) {
		loc_D := u4_abs(bp04-bp08) / 2
		for {
			Gra_dot_OR(loc_A, loc_E, 3)
			loc_A += loc_C
			loc_D -= u4_abs(bp06 - bp0a)
			if loc_D <= 0 {
				loc_E += loc_B
				loc_D += u4_abs(bp04 - bp08)
			}
			if loc_A == bp04 {
				break
			}
		}
	} else {
		loc_D := u4_abs(bp06-bp0a) / 2
		for {
			Gra_dot_OR(loc_A, loc_E, 3)
			loc_E += loc_B
			loc_D -= u4_abs(bp04 - bp08)
			if loc_D <= 0 {
				loc_A += loc_C
				loc_D += u4_abs(bp06 - bp0a)
			}
			if loc_E == bp06 {
				break
			}
		}
	}
	Gra_dot_OR(loc_A, loc_E, 3)
}

// pit: up (top row tables) or low (bottom)
func pit(a, b [4]uint8) {
	si := STA_E
	Gra_line(q8(D_2DBC, si), q8(a, si), q8(D_2DC0, si), q8(b, si), 3)
	Gra_line(q8(D_2DBC, si)+1, q8(a, si), q8(D_2DC0, si)+1, q8(b, si), 3)
	Gra_line(q8(D_2DC0, si)+1, q8(b, si), q8(D_2DC8, si), q8(b, si), 3)
	Gra_line(q8(D_2DC8, si), q8(b, si), q8(D_2DC4, si), q8(a, si), 3)
	Gra_line(q8(D_2DC8, si)-1, q8(b, si), q8(D_2DC4, si)-1, q8(a, si), 3)
	Gra_line(q8(D_2DC4, si)-1, q8(a, si), q8(D_2DBC, si)+1, q8(a, si), 3)
}

func C_A94A() { pit(D_2DAC, D_2DB0) } // up pit
func C_AA29() { pit(D_2DB4, D_2DB8) } // low pit

func C_AB08(bp04 uint8) int {
	D_2DEC--
	if D_2DEC+1 == 0 {
		D_2DEC = speed_info - 1
	}
	switch bp04 & 0xf0 {
	default:
		if bp04&0xf0 < 0xc0 {
			return 0
		}
		return -1
	case 0x40:
		return TIL_3C
	case 0x70:
		return TIL_4E
	case 0x80:
		if D_2DEC == 1 && U4_RND1(0x3f) == 0 && bp04&0xf != 0 {
			if bp04&0x0f >= 8 {
				C_AA29()
			} else {
				C_A94A()
			}
		}
		return 0
	case 0x90:
		return TIL_02
	case 0xa0:
		C_A7F2(bp04&3 | TIL_44)
		switch STA_E {
		case 0:
			cursor_rate = 0
		case 1:
			cursor_rate = 1
		case 2:
			cursor_rate = 6
		case 3:
			cursor_rate = 10
		}
		return TIL_48
	case 0xb0:
		return TIL_4A
	}
}

// isSpriteRenderable?
func C_ABE5(bp04 uint8) bool {
	si := bp04 & 0xf0
	if si == 0x80 || si == 0x90 || si == 0xa0 || si >= 0xd0 {
		return false
	}
	return bp04&0x0f != 0
}

// The sprite pause of C_AC23 per STA_E: a calibrated while(--long) of k*speed_info passes
// (each about 3 delay-loop passes of 55ms/8000), k = 1000, 1200, 2400, 3000.
var spritePause = [4]time.Duration{20625 * time.Microsecond, 24750 * time.Microsecond, 49500 * time.Microsecond, 61875 * time.Microsecond}

// render sprites?
func C_AC23(bp04 int) {
	cursor_rate = 0
	switch STA_E {
	case 0:
		for b := 0; b < 48; b++ {
			Gra_01(bp04, b/3, b+0x70, 0x14, 2)
		}

	case 1:
		for b := 0; b < 32; b++ {
			Gra_01(bp04, b>>1, b+0x68, 0x14, 2)
		}

	case 2:
		Gra_02(1, 1, bp04, 0x60, 0x16)

	case 3:
		for b := 0; b < 8; b++ {
			Gra_01(bp04, b<<1, b+0x60, 0x16, 1)
		}
	}
	host.Wait(spritePause[STA_E&3])
}

func C_AD6A(bp06 uint8, bp04 int) bool {
	STA_E = bp04
	D_2DEE--
	if D_2DEE+1 == 0 {
		D_2DEE = speed_info - 1
		D_8CFA = U4_RND1(3)
	}
	if C_ABE5(bp06) {
		C_AC23(TIL_90 + int(bp06&0xf)<<2 - 4 + D_8CFA)
		return false
	}
	bp_02 := C_AB08(bp06)
	if bp_02 > 0 {
		if bp_02 != TIL_48 {
			C_AC23(bp_02)
		}
		return false
	}
	if bp_02 != 0 {
		if STA_E != 0 || bp06 >= 0xd0 {
			return false
		}
	}
	return true
}

// animate 3D sprites?
func C_ADEF() {
	if Party.f_1dc == 0 {
		Gra_10()
		return
	}
	if !C_AD6A(tile_cur, 0) {
		return
	}
	if !C_AD6A(tile_forward, 1) {
		return
	}
	if !C_AD6A(STA_A, 2) {
		return
	}
	C_AD6A(STA_B, 3)
}

func dngTile(y, x int) uint8 { return *D_8742.x8x8x8(int(Party._z), y, x) }

// compute nearest tiles
func C_AE41() {
	d := int(Party._dir)
	// back
	x := DNG_X_m(int(Party._x), d)
	y := DNG_Y_m(int(Party._y), d)
	tile_back = dngTile(y, x)
	// current tile
	tile_cur = dngTile(int(Party._y), int(Party._x))
	// forward
	x = DNG_X_p(int(Party._x), d)
	y = DNG_Y_p(int(Party._y), d)
	tile_forward = dngTile(y, x)
	// another 2 steps forward
	x = DNG_X_p(x, d)
	y = DNG_Y_p(y, d)
	STA_A = dngTile(y, x)
	x = DNG_X_p(x, d)
	y = DNG_Y_p(y, d)
	STA_B = dngTile(y, x)
}

// compute the actual x,y of the square at depth STA_E and column bp04 (-1 left, 0
// straight, +1 right) from the party's position and orientation
func C_AF36(getX, getY *int, bp04 int) {
	*getX = int(Party._x)
	*getY = int(Party._y)
	switch Party._dir {
	case DIR_N:
		*getY = (*getY - STA_E) & 7
		*getX = (*getX + bp04) & 7
	case DIR_S:
		*getY = (*getY + STA_E) & 7
		*getX = (*getX - bp04) & 7
	case DIR_E:
		*getX = (*getX + STA_E) & 7
		*getY = (*getY + bp04) & 7
	case DIR_W:
		*getX = (*getX - STA_E) & 7
		*getY = (*getY - bp04) & 7
	}
}

func VERTICAL(x, y0, y1, c int)   { Gra_line(x&0xff, y0, x&0xff, y1, c) }
func HORIZONTAL(x0, y, x1, c int) { Gra_line(x0&0xff, y, x1&0xff, y, c) }

// up ladder
func C_AFB5() {
	si := STA_E
	C_A94A()
	if Party._dir == DIR_N || Party._dir == DIR_S {
		VERTICAL(0x5f, q8(D_2DAC, si), 0x5f, 3)
		VERTICAL(0x5f+1, q8(D_2DAC, si), 0x5f, 3)
	} else {
		VERTICAL(q8(Ladder_x0, si), q8(D_2DAC, si), 0x5e, 3)
		VERTICAL(q8(Ladder_x0, si)+1, q8(D_2DAC, si), 0x5e, 3)
		VERTICAL(q8(Ladder_x1, si), q8(D_2DAC, si), 0x5e, 3)
		VERTICAL(q8(Ladder_x1, si)-1, q8(D_2DAC, si), 0x5e, 3)
		HORIZONTAL(q8(Ladder_x0, si), q8(D_2DD4, si), q8(Ladder_x1, si), 3)
		HORIZONTAL(q8(Ladder_x0, si), q8(D_2DD8, si), q8(Ladder_x1, si), 3)
		HORIZONTAL(q8(Ladder_x0, si), q8(D_2DDC, si), q8(Ladder_x1, si), 3)
	}
}

// down ladder
func C_B0E9() {
	si := STA_E
	C_AA29()
	if Party._dir == DIR_N || Party._dir == DIR_S {
		VERTICAL(0x5f, q8(D_2DB4, si), 0x5f, 3)
		VERTICAL(0x5f+1, q8(D_2DB4, si), 0x5f, 3)
	} else {
		VERTICAL(q8(Ladder_x0, si), 0x5f, q8(D_2DB4, si), 3)
		VERTICAL(q8(Ladder_x0, si)+1, 0x5f, q8(D_2DB4, si), 3)
		VERTICAL(q8(Ladder_x1, si), 0x5f, q8(D_2DB4, si), 3)
		VERTICAL(q8(Ladder_x1, si)-1, 0x5f, q8(D_2DB4, si), 3)
		HORIZONTAL(q8(Ladder_x0, si), q8(D_2DE0, si), q8(Ladder_x1, si), 3)
		HORIZONTAL(q8(Ladder_x0, si), q8(D_2DE4, si), q8(Ladder_x1, si), 3)
		HORIZONTAL(q8(Ladder_x0, si), q8(D_2DE8, si), q8(Ladder_x1, si), 3)
	}
}

func dirCol() int    { return int(Party._dir&1) + 1 }
func dirColNot() int { return b2i(Party._dir&1 == 0) + 1 }

// front wall?
func C_B21D() {
	e := STA_E
	Gra_18(w8(Wall_x0, e), w8(Wall_x1, e)+1, w8(Wall_y0, e), w8(Wall_y1, e), w8(Wall_y0, e), w8(Wall_y1, e), dirCol(), 0)
}

// left front wall?
func C_B25E() {
	e := STA_E
	Gra_18(w8(Wall_x0, e), w8(Wall_x0, e+1), w8(Wall_y0, e+1), w8(Wall_y1, e+1), w8(Wall_y0, e+1), w8(Wall_y1, e+1), dirCol(), 0)
}

// right front wall?
func C_B29C() {
	e := STA_E
	Gra_18(w8(Wall_x1, e), w8(Wall_x1, e+1), w8(Wall_y0, e+1), w8(Wall_y1, e+1), w8(Wall_y0, e+1), w8(Wall_y1, e+1), dirCol(), 0)
}

// left lateral wall?
func C_B2DA() {
	e := STA_E
	Gra_18(w8(Wall_x0, e), w8(Wall_x0, e+1), w8(Wall_y0, e), w8(Wall_y1, e), w8(Wall_y0, e+1), w8(Wall_y1, e+1), dirColNot(), 0)
}

// right lateral wall?
func C_B31B() {
	e := STA_E
	Gra_18(w8(Wall_x1, e), w8(Wall_x1, e+1), w8(Wall_y0, e), w8(Wall_y1, e), w8(Wall_y0, e+1), w8(Wall_y1, e+1), dirColNot(), 0)
}

// front door?
func C_B35C() {
	e := STA_E
	c0, c1 := 3, 0
	if e == 0 {
		c0, c1 = 0, 3
	}
	Gra_18(w8(Wall_x0, e+1), w8(Wall_x1, e+1), w8(Wall_y0, e+1), w8(Wall_y1, e), w8(Wall_y0, e+1), w8(Wall_y1, e), c0, c1)
}

// left door?
func C_B3AE() {
	e := STA_E
	Gra_18(q8(LDoor_x0, e), q8(LDoor_x1, e), q8(Door_y0, e), q8(Door_y1, e), q8(Door_y0p, e), q8(Door_y1p, e), 3, 0)
}

// right door?
func C_B3E3() {
	e := STA_E
	Gra_18(q8(RDoor_x0, e), q8(RDoor_x1, e), q8(Door_y0, e), q8(Door_y1, e), q8(Door_y0p, e), q8(Door_y1p, e), 3, 0)
}

// right walls
func C_B418() {
	var x, y int
	C_AF36(&x, &y, 1)
	switch dngTile(y, x) >> 4 {
	case 0xc, 0xd:
		C_B31B()
		C_B3E3()
	case 0xe, 0xf:
		C_B31B()
	default:
		C_B29C()
	}
}

// left & right walls
func C_B47A() {
	var x, y int
	C_AF36(&x, &y, -1)
	switch dngTile(y, x) >> 4 {
	case 0xc, 0xd:
		C_B2DA()
		C_B3AE()
	case 0xe, 0xf:
		C_B2DA()
	default:
		C_B25E()
	}
	C_B418()
}

// render 3d
func C_B4DF() {
	x, y := int(Party._x), int(Party._y)
	for STA_E = 0; STA_E <= 3; STA_E++ {
		C_AF36(&x, &y, 0)
		t := dngTile(y, x)
		switch t & 0xf0 {
		case 0x30:
			C_B0E9()
			C_AFB5()
			C_B47A()
		case 0x10:
			C_AFB5()
			C_B47A()
		case 0x00, 0x40, 0x70, 0x80, 0x90, 0xb0:
			C_B47A()
		case 0x20:
			C_B0E9()
			C_B47A()
		case 0x50:
			C_A94A()
			C_B47A()
		case 0x60:
			C_AA29()
			C_B47A()
		case 0xc0:
			C_B21D()
			C_B35C()
			if STA_E != 0 {
				return
			}
		case 0xd0:
			C_B21D()
			C_B35C()
			return
		case 0xa0:
			C_A7F2(t)
			return
		case 0xe0, 0xf0:
			C_B21D()
			return
		}
	}
	STA_E = 4
	C_AF36(&x, &y, 0)
	if t := dngTile(y, x) & 0xf0; t == 0xf0 || t == 0xe0 {
		C_B21D()
	}
}

// backup dungeon infos & render 3D
func C_B61C() {
	STA_D = Party._z
	STA_F = Party._x
	STA_H = Party._y
	STA_G = Party._dir
	STA_C = Party.f_1dc == 0
	STA_I = tile_cur
	STA_J = tile_forward
	STA_K = STA_A
	STA_L = STA_B
	D_9446 = 0
	Gra_10()
	if Party.f_1dc != 0 {
		C_B4DF()
	}
}

// render 3d if necessary
func C_B677() {
	if Party.f_1dc != 0 {
		cursor_rate = 13
	} else {
		cursor_rate = 1
	}
	C_AE41()
	if D_9446 == 0 && STA_D == Party._z && STA_F == Party._x && STA_H == Party._y &&
		STA_G == Party._dir && STA_I == tile_cur && STA_C == (Party.f_1dc == 0) {
		bp_02 := tile_cur & 0xf0
		if bp_02 == 0xa0 || bp_02 >= 0xd0 {
			return
		}
		if tile_forward == STA_J {
			if bp_02 = tile_forward & 0xf0; bp_02 == 0xa0 || bp_02 >= 0xc0 {
				return
			}
			if STA_A == STA_K {
				if bp_02 = STA_A & 0xf0; bp_02 == 0xa0 || bp_02 >= 0xc0 {
					return
				}
				return // "if(STA_L == STA_L/*bug?*/) return;"
			}
		}
	}
	C_B61C()
}
