package u4

// Port of U4_MAP.C (unsigned char): the outdoor map window (four 16x16 quarters of the
// world's 32x32 chunks in D_8742._map) and moving on foot, by ship and balloon.

var (
	D_08CE = -1
	D_08D0 = -1
)

// load the portion of the map with coordinates in parameter
func C_2399(bp06, bp04 int) {
	if D_08CE != -1 {
		D_933A[D_08D0][D_08CE] = nil
	}
	if D_933A[bp06][bp04] == nil {
		m := new(tMap32x32)
		chunk := File_MAP[(bp06*8+bp04)*0x400:]
		for y := range m {
			copy(m[y][:], chunk[y*32:])
		}
		D_933A[bp06][bp04] = m
	}
	D_08CE = bp04
	D_08D0 = bp06
}

// copyQuarter loads the chunk with world quarter (qy, qx) into the window at (dy, dx).
func copyQuarter(qy, qx, dy, dx int) {
	C_2399(qy/2, qx/2)
	m := D_933A[qy/2][qx/2]
	for di := 0; di < 16; di++ {
		for si := 0; si < 16; si++ {
			*D_8742.x32x32(di+dy, si+dx) = m[(qy&1)*0x10+di][(qx&1)*0x10+si]
		}
	}
}

func tmp0() int { return (int(D_95A5.x) + 1) & 15 }
func tmp1() int { return (int(D_95A5.y) + 1) & 15 }

func C_249C() { copyQuarter(int(D_95A5.y), int(D_95A5.x), 0, 0) }
func C_2517() { copyQuarter(int(D_95A5.y), tmp0(), 0, 16) }
func C_259F() { copyQuarter(tmp1(), int(D_95A5.x), 16, 0) }
func C_2624() { copyQuarter(tmp1(), tmp0(), 16, 16) }

func C_26B6() {
	x, y := Party._x, Party._y
	if Party._loc >= 0x11 {
		x, y = Party.out_x, Party.out_y
	}
	D_95A5.x = x >> 4
	D_959C.x = x & 15
	if D_959C.x < 8 {
		D_959C.x += 0x10
		D_95A5.x = (D_95A5.x - 1) & 15
	}
	D_95A5.y = y >> 4
	D_959C.y = y & 15
	if D_959C.y < 8 {
		D_959C.y += 0x10
		D_95A5.y = (D_95A5.y - 1) & 15
	}
	C_249C()
	C_2517()
	C_259F()
	C_2624()
}

// Leaving...
func C_2747() {
	u4_puts("Leaving...\n")
	CurMode = MOD_VISION
	LoadRec("OUTMONST.SAV", &D_8742._npc)
	File_TLK = nil
	Party._x = Party.out_x
	Party._y = Party.out_y
	C_26B6()
	CurMode = MOD_OUTDOORS
	Party._loc = 0
	Party.f_1dc = 0
	D_9440 = 1
	if Party._x == 0xef && Party._y == 0xf0 {
		D_8742._npc._gtile[31] = TIL_18
		D_8742._npc._tile[31] = TIL_18
		D_8742._npc._x[31] = 0xe9
		D_8742._npc._y[31] = 0xf2
	}
	u_kbflush()
}

// spawn daemons (if horn not used)
func C_27D9() {
	if spell_sta != 1 {
		for si := 7; si >= 0; si-- {
			D_8742._npc._gtile[si] = TIL_F0
			D_8742._npc._tile[si] = TIL_F0
			D_8742._npc._x[si] = 0xe7
			D_8742._npc._y[si] = Party._y + 1
		}
	}
}

var D_08EC = [8]uint8{0xE0, 0xE0, 0xE2, 0xE3, 0xE4, 0xE5, 0xE5, 0xE4}
var D_08F4 = [8]uint8{0xDC, 0xE4, 0xDC, 0xE4, 0xE3, 0xE1, 0xDF, 0xDE}
var D_08FC = [8]uint8{TIL_82, TIL_82, TIL_82, TIL_82, TIL_83, TIL_83, TIL_81, TIL_81}

// spawn pirate ships
func C_280A() {
	for si := 7; si >= 0; si-- {
		D_8742._npc._x[si] = D_08EC[si]
		D_8742._npc._y[si] = D_08F4[si]
		D_8742._npc._tile[si] = TIL_80
		D_8742._npc._gtile[si] = D_08FC[si]
	}
}

// GOING NORTH
func C_2839() {
	Party._y--
	D_959C.y--
	if D_959C.y < 5 {
		D_959C.y += 16
		D_95A5.y = (D_95A5.y - 1) & 15
		copy(D_8742._map[16*32:32*32], D_8742._map[0:16*32])
		C_249C()
		C_2517()
	}
}

// GOING SOUTH
func C_2891() {
	Party._y++
	D_959C.y++
	if D_959C.y >= 27 {
		D_959C.y -= 16
		D_95A5.y = (D_95A5.y + 1) & 15
		copy(D_8742._map[0:16*32], D_8742._map[16*32:32*32])
		C_259F()
		C_2624()
	}
}

// GOING WEST
func C_28E9() {
	Party._x--
	D_959C.x--
	if D_959C.x < 5 {
		D_959C.x += 16
		D_95A5.x = (D_95A5.x - 1) & 15
		for di := 0; di < 32; di++ {
			copy(D_8742._map[di*32+16:di*32+32], D_8742._map[di*32:di*32+16])
		}
		C_249C()
		C_259F()
	}
}

// GOING EAST
func C_2941() {
	Party._x++
	D_959C.x++
	if D_959C.x >= 27 {
		D_959C.x -= 16
		D_95A5.x = (D_95A5.x + 1) & 15
		for di := 0; di < 32; di++ {
			copy(D_8742._map[di*32:di*32+16], D_8742._map[di*32+16:di*32+32])
		}
		C_2517()
		C_2624()
	}
}

var D_0904 = []uint8{
	TIL_03, TIL_04, TIL_05, TIL_06, TIL_07, TIL_09, TIL_0A, TIL_0B, TIL_0C,
	TIL_10, TIL_11, TIL_12, TIL_13, TIL_14, TIL_15, TIL_16, TIL_17, TIL_18,
	TIL_19, TIL_1A, TIL_1B, TIL_1C, TIL_1D, TIL_1E, TIL_3C, TIL_3E, TIL_3F,
	TIL_43, TIL_44, TIL_46, TIL_47, TIL_49, TIL_4A, TIL_4C, TIL_8E, TIL_8F,
	0, // End of list
}

// isBrickSolid (in fact: walkable)
func C_2999(bp04 uint8) bool {
	for si := 0; D_0904[si] != 0; si++ {
		if bp04 == D_0904[si] {
			return true
		}
	}
	return false
}

func w_Blocked() { /*C_29C3*/
	u4_puts("Blocked!\n")
	sound(1)
	u_kbflush()
}

func w_DriftOnly() { /*C_29DE*/
	u4_puts("Drift Only!\n")
	u_kbflush()
}

// check slow progress
func C_29EF(bp04 uint8) bool {
	switch bp04 {
	case TIL_03:
		if U4_RND1(7) != 0 {
			return false
		}
	case TIL_05, TIL_06:
		if U4_RND1(3) != 0 {
			return false
		}
	case TIL_07, TIL_46:
		if U4_RND1(1) == 0 {
			return false
		}
	default:
		return false
	}
	return true
}

func C_2A38(bp04 uint8) bool {
	return bp04 < TIL_02 || (bp04 >= TIL_8C && bp04 < TIL_90)
}

// check slow progress [boat]
func C_2A5A(bp04 uint16) bool {
	if WindDir == bp04 {
		return Party._moves&3 == 0
	}
	if WindDir == (bp04+2)&3 {
		return Party._moves&3 != 0
	}
	return true
}

// enter moongate ?
func C_2A91(bp04 uint8) bool {
	if bp04 != TIL_43 {
		return false
	}
	if Party._trammel == 4 && Party._felucca == 4 {
		t_callback()
		Gra_09()
		Party._loc = 0x1f
		sound(9, 0xa0)
		C_E72C()
		return true
	}
	t_callback()
	Party._x = D_0814[Party._felucca]
	Party._y = D_081C[Party._felucca]
	Gra_09()
	sound(9, 0xa0)
	Gra_09()
	C_26B6()
	t_callback()
	Gra_09()
	sound(9, 0xa0)
	Gra_09()
	return true
}

// move North
func C_2B19() bool {
	// LB's castle middle wing
	if tile_cur == TIL_0E {
		w_Blocked()
		return false
	}
	if tile_north != TIL_0E && !C_2999(tile_north) {
		w_Blocked()
		return false
	}
	if C_29EF(tile_north) {
		w_SlowProgress()
		return true
	}
	sound(0)
	if CurMode == MOD_OUTDOORS {
		C_2839()
		return !C_2A91(tile_north)
	}
	Party._y--
	D_959C.y = Party._y
	if D_959C.y <= 0x1f {
		return true
	}
	C_2747()
	return false
}

func CMDDIR_Up() { /*C_2B8C*/
	switch {
	case Party._tile == TIL_10 || Party._tile == TIL_12 || Party._tile == TIL_13:
		Party._tile = TIL_11
		u4_puts("Turn North!\n")
	case Party._tile == TIL_11:
		u4_puts("Sail North!\n")
		if !C_2A38(tile_north) {
			w_Blocked()
		} else if !C_2A5A(DIR_N) {
			w_SlowProgress()
		} else {
			C_2839()
		}
	case Party._tile == TIL_18:
		w_DriftOnly()
	default:
		sound(0)
		u4_puts("North\n")
		if C_2B19() && D_95C6 != 0 {
			t_callback()
			sound(0)
			C_2B19()
		}
	}
}

// move South
func C_2C25() bool {
	if !C_2999(tile_south) {
		w_Blocked()
		return false
	}
	if C_29EF(tile_south) {
		w_SlowProgress()
		return true
	}
	sound(0)
	if CurMode == MOD_OUTDOORS {
		C_2891()
		return !C_2A91(tile_south)
	}
	Party._y++
	D_959C.y = Party._y
	if D_959C.y <= 0x1f {
		return true
	}
	C_2747()
	return false
}

func CMDDIR_Down() { /*C_2C8A*/
	switch {
	case Party._tile == TIL_18:
		w_DriftOnly()
	case Party._tile == TIL_10 || Party._tile == TIL_12 || Party._tile == TIL_11:
		Party._tile = TIL_13
		u4_puts("Turn South!\n")
	case Party._tile == TIL_13:
		u4_puts("Sail South!\n")
		if !C_2A38(tile_south) {
			w_Blocked()
		} else if !C_2A5A(DIR_S) {
			w_SlowProgress()
		} else {
			C_2891()
		}
	default:
		sound(0)
		u4_puts("South\n")
		if C_2C25() && D_95C6 != 0 {
			t_callback()
			sound(0)
			C_2C25()
		}
		// spawn demons on verity island?
		if Party._x >= 0xe5 && Party._x < 0xea && Party._y >= 0xd4 && Party._y < 0xd9 {
			C_27D9()
		}
	}
}

// move West
func C_2D44() bool {
	if !C_2999(tile_west) {
		w_Blocked()
		return false
	}
	if C_29EF(tile_west) {
		w_SlowProgress()
		return true
	}
	sound(0)
	if CurMode == MOD_OUTDOORS {
		C_28E9()
		return !C_2A91(tile_west)
	}
	Party._x--
	D_959C.x = Party._x
	if D_959C.x <= 0x1f {
		return true
	}
	C_2747()
	return false
}

func CMDDIR_Left() { /*C_2DA9*/
	switch {
	case Party._tile == TIL_18:
		w_DriftOnly()
	case Party._tile == TIL_13 || Party._tile == TIL_12 || Party._tile == TIL_11:
		Party._tile = TIL_10
		u4_puts("Turn West!\n")
	case Party._tile == TIL_10:
		u4_puts("Sail West!\n")
		if !C_2A38(tile_west) {
			w_Blocked()
		} else if !C_2A5A(DIR_W) {
			w_SlowProgress()
		} else {
			C_28E9()
		}
	default:
		if Party._tile == TIL_15 {
			Party._tile = TIL_14
		}
		sound(0)
		u4_puts("West\n")
		if C_2D44() && D_95C6 != 0 {
			t_callback()
			sound(0)
			C_2D44()
		}
	}
}

// move East
func C_2E4F() bool {
	if !C_2999(tile_east) {
		w_Blocked()
		return false
	}
	if C_29EF(tile_east) {
		w_SlowProgress()
		return true
	}
	sound(0)
	if CurMode == MOD_OUTDOORS {
		C_2941()
		if Party._y == 0xe0 && Party._x == 0xdd {
			C_280A()
		}
		return !C_2A91(tile_east)
	}
	Party._x++
	D_959C.x = Party._x
	if D_959C.x <= 0x1f {
		return true
	}
	C_2747()
	return false
}

func CMDDIR_Right() { /*C_2EC5*/
	switch {
	case Party._tile == TIL_18:
		w_DriftOnly()
	case Party._tile == TIL_13 || Party._tile == TIL_10 || Party._tile == TIL_11:
		Party._tile = TIL_12
		u4_puts("Turn East!\n")
	case Party._tile == TIL_12:
		u4_puts("Sail East!\n")
		if !C_2A38(tile_east) {
			w_Blocked()
		} else if !C_2A5A(DIR_E) {
			w_SlowProgress()
		} else {
			C_2941()
			if Party._y == 0xe0 && Party._x == 0xdd {
				C_280A()
			}
		}
	default:
		if Party._tile == TIL_14 {
			Party._tile = TIL_15
		}
		sound(0)
		u4_puts("East\n")
		if C_2E4F() && D_95C6 != 0 {
			t_callback()
			sound(0)
			C_2E4F()
		}
	}
}
