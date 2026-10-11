package u4

// Ports of U4_COMBB.C (unsigned char: whirlpools, twisters, moving in combat) and
// U4_DNG.C (the dungeon command loop).

// ---- U4_COMBB.C: Whirlpool and Twister management ----

// remove monster?
func C_7821(si, bp04 int) {
	n := &D_8742._npc
	for di := 31; di >= 0; di-- {
		if di != si && n._tile[di] != 0 && n._x[di] == n._x[si] && n._y[di] == n._y[si] {
			n._gtile[di] = 0
			n._tile[di] = 0
			t_callback()
			sound(bp04)
		}
	}
}

// whirlpool contact?
func C_786F(si int) {
	if Party._x == D_8742._npc._x[si] && Party._y == D_8742._npc._y[si] {
		t_callback()
		hit_x = 5
		hit_y = 5
		hit_tile = TIL_8C
		C_36C7()
		Party._x = 0x7f
		Party._y = 0x4e
		sound(11)
		C_1584()
		Party._tile = TIL_10
		hit_tile = 0
		C_26B6()
	}
	C_7821(si, 11)
}

// twister contact?
func C_78D1(si int) {
	if Party._x == D_8742._npc._x[si] && Party._y == D_8742._npc._y[si] {
		t_callback()
		sound(12)
		C_1584()
		C_1584()
		C_1584()
		C_1584()
		hit_tile = 0
	}
	C_7821(si, 12)
}

func C_7918() {
	if CurMode != MOD_OUTDOORS {
		return
	}
	for si := 3; si >= 0; si-- {
		if D_8742._npc._tile[si] == TIL_8C {
			C_786F(si)
		} else if D_8742._npc._tile[si] == TIL_8E {
			C_78D1(si)
		}
	}
}

func w_SameExit() { /*C_794D*/
	u4_puts("All must use same exit!\n")
	sound(1)
}

// exit combat zone
func C_7962() {
	a := int(activeChara)
	Combat._charaX[a] = 0x80
	Combat._charaY[a] = 0x80
	Fighters._chtile[a] = 0
	sound(8)
	if CurMode == MOD_COMBAT && C_0AFE(D_96F8) == 0 && Party.chara[a]._HP[0] == Party.chara[a]._HP[1] {
		karma_dec(&Party._valor, 2)
		karma_dec(&Party._sacri, 2)
	}
}

// exiting fight
func C_79C9(bp06, bp04 uint8) {
	if CurMode == MOD_COM_ROOM {
		// someone already exited W/E
		if D_96EE != 0 {
			if bp06 != D_96EE {
				w_SameExit()
				return
			}
			C_7962()
			return
		}
		// someone already exited N/S
		if D_96F4 != 0 {
			if bp04 != D_96F4 {
				w_SameExit()
				return
			}
			C_7962()
			return
		}
		if bp06 > 10 {
			D_96EE = bp06
		} else {
			D_96F4 = bp04
		}
	}
	C_7962()
}

func u16at(b []uint8, i int) uint16 { return uint16(b[i]) | uint16(b[i+1])<<8 }

// room special tiles
func C_7A15(bp06, bp04 uint8) {
	for a := 3; a >= 0; a-- {
		o := a << 2
		if D_95B2[o] == 0 {
			continue
		}
		w0 := u16at(D_95B2[:], o)
		if uint16(bp06)<<12 == w0&0xf000 && uint16(bp04)<<8 == w0&0x0f00 {
			w1 := u16at(D_95B2[:], o+2)
			b := int(w1 & 0xf)
			c := int(w1>>4) & 0xf
			if b|c != 0 {
				*Combat_MAP(b, c) = D_95B2[o]
			}
			b = int(w1>>8) & 0xf
			c = int(w1>>12) & 0xf
			if b|c != 0 {
				*Combat_MAP(b, c) = D_95B2[o]
			}
		}
	}
}

// move [fight]
func C_7AE3(si int, bp06, bp04 uint8) {
	sound(0)
	// out of bounds => exiting
	if bp06 > 10 || bp04 > 10 {
		C_79C9(bp06, bp04)
		return
	}
	di := D_96F9[int(bp04)*11+int(bp06)]
	if !C_2999(di) {
		w_Blocked()
		return
	}
	if C_29EF(di) {
		w_SlowProgress()
		return
	}
	Combat._charaX[si] = bp06
	Combat._charaY[si] = bp04
	sound(0)
	if CurMode == MOD_COM_ROOM {
		C_7A15(bp06, bp04)
	}
}

func C_7B89() { // North [combat]
	u4_puts("North\n")
	a := int(activeChara)
	C_7AE3(a, Combat._charaX[a], Combat._charaY[a]-1)
}

func C_7BAF() { // South [combat]
	u4_puts("South\n")
	a := int(activeChara)
	C_7AE3(a, Combat._charaX[a], Combat._charaY[a]+1)
}

func C_7BD5() { // East [combat]
	u4_puts("East\n")
	a := int(activeChara)
	C_7AE3(a, Combat._charaX[a]+1, Combat._charaY[a])
}

func C_7BFD() { // West [combat]
	u4_puts("West\n")
	a := int(activeChara)
	C_7AE3(a, Combat._charaX[a]-1, Combat._charaY[a])
}

// ---- U4_DNG.C ----

func DNG_main() { /*C_84D2*/
	var si int
	D_9446 = 1
	u_kbflush()
	bp_04 := 0
	for {
		C_9209()
		u4_putc(0x10)
		t_callback()
		if C_10FD() {
			u_delay(25, 1)
			if u_kbhit() != 0 {
				si = u_kbread()
			} else {
				si = KBD_SPACE
			}
			if u4_isupper(si & 0xff) {
				si = si&0xff00 | u4_lower(si&0xff)
			}
			switch si {
			case KBD_SPACE:
				w_Pass()
			case KBD_UP, 0x487e:
				C_891E()
			case KBD_LEFT, 0x4b7c:
				C_899F()
			case KBD_DOWN, 0x5060:
				C_895F()
			case KBD_RIGHT, 0xf400:
				C_89BD()
			case KBD_C:
				CMD_Cast()
			case KBD_D:
				C_8A1F() // (D)escend [dungeon]
			case KBD_G:
				CMD_Get()
			case KBD_H:
				CMD_HoleUp()
			case KBD_I:
				CMD_Ignite()
			case KBD_K:
				C_89DB() // (K)limb [dungeon]
			case KBD_M:
				CMD_Mix()
			case KBD_N:
				CMD_NewOrder()
			case KBD_P:
				CMD_Peer()
			case KBD_Q:
				CMD_Quit()
			case KBD_R:
				CMD_Ready()
			case KBD_S:
				C_B9B2() // (S)earch [dungeon]
			case KBD_U:
				CMD_Use()
			case KBD_V:
				CMD_Volume()
			case KBD_W:
				CMD_Wear()
			case KBD_Z:
				CMD_Ztats()
			case KBD_A, KBD_B, KBD_E, KBD_F, KBD_J, KBD_L, KBD_O, KBD_T, KBD_X, KBD_Y:
				w_NotHere()
				sound(2)
			case KBD_CTRL_S:
				if bp_04 == KBD_ALT_Z {
					C_1C21()
					break
				}
				fallthrough
			default:
				u4_puts("Bad command\n")
				sound(2)
				D_07F8 = 0
			}
			C_AE41()
		} else {
			u4_puts("Zzzzz\n")
		}
		if D_07F8 != 0 && Party._z >= 0 {
			C_87E2()
		} else {
			D_07F8 = 1
		}
		bp_04 = si
		if Party._z < 0 {
			break
		}
	}
	C_2747()
	dclose(File_DNG)
	File_DNG = 0
}

func C_8743() bool {
	t := tile_cur
	if t&0x0f == 0 || t&0xf0 == 0x80 || t&0xf0 == 0x90 || t&0xf0 == 0xA0 || t&0xf0 == 0xD0 || t&0xf0 == 0xF0 {
		return false
	}
	D_9452 = TIL_90 + (t&0x0f)*4 - 4
	D_96F8 = D_9452
	*D_8742.x8x8x8(int(Party._z), int(Party._y), int(Party._x)) &= 0xf0
	n := &D_8742._npc
	for si := 0x1f; si >= 0; si-- {
		if int16(n._var[si]) == Party._z && n._x[si] == Party._x && n._y[si] == Party._y {
			n._var[si] = 0
			n._tile[si] = 0
			n._x[si] = 0
			n._y[si] = 0
			break
		}
	}
	C_7D50()
	return true
}

func C_87E2() {
	C_10FD()
	for si := 0; si < int(Party.f_1d8); si++ {
		di := &Party.chara[si]
		if di._stat == 'S' && U4_RND1(7) == 0 {
			di._stat = 'G'
		} else if di._stat == 'P' {
			hitChara(si, 2)
			Gra_11(si)
			sound(6)
			Gra_11(si)
		}
	}
	if food_dec(int(Party.f_1d8)) {
		u4_puts("\nStarving!!!\n")
		for si := 0; si < int(Party.f_1d8); si++ {
			if isCharaAlive(si) {
				hitChara(si, 2)
			}
		}
		for si := int(Party.f_1d8) - 1; si >= 0; si-- {
			Gra_11(si)
		}
		sound(6)
		for si := int(Party.f_1d8) - 1; si >= 0; si-- {
			Gra_11(si)
		}
	}
	MP_recover()
	C_9414()
	C_95AA()
	if Party.f_1dc != 0 {
		Party.f_1dc--
	}
	Party._moves++
	if spell_cnt != 0 {
		spell_cnt--
		if spell_cnt == 0 {
			spell_sta = 7
		}
	}
	if Party.f_1dc == 0 {
		u4_puts("It's Dark!\n")
	}
	dspl_Stats()
	if !C_8743() {
		for tile_cur&0xf0 == 0xd0 {
			C_10FD() // check death
			C_7FFD() // enter dungeon room
			C_B677() // render 3d
		}
	}
}

func C_88E8(bp06 uint8, bp04 int) bool {
	if bp04 == 0 {
		if bp06&0xf0 == 0xc0 || bp06&0xf0 == 0xe0 || bp06&0xf0 == 0xd0 {
			return true
		}
	}
	return bp06 != 0xa1 && bp06 < 0xc0
}

// Advance
func C_891E() {
	u4_puts("Advance\n")
	if !C_88E8(tile_forward, 0) {
		w_Blocked()
		return
	}
	Party._x = uint8(DNG_X_p(int(Party._x), int(Party._dir)))
	Party._y = uint8(DNG_Y_p(int(Party._y), int(Party._dir)))
}

// Retreat
func C_895F() {
	u4_puts("Retreat\n")
	if !C_88E8(tile_back, 1) {
		w_Blocked()
		return
	}
	Party._x = uint8(DNG_X_m(int(Party._x), int(Party._dir)))
	Party._y = uint8(DNG_Y_m(int(Party._y), int(Party._dir)))
}

// Turn Left
func C_899F() {
	u4_puts("Turn Left\n")
	Party._dir = (Party._dir - 1) & 3
	D_07F8 = 0
}

// Turn Right
func C_89BD() {
	u4_puts("Turn Right\n")
	Party._dir = (Party._dir + 1) & 3
	D_07F8 = 0
}

// (K)limb [dungeon]
func C_89DB() {
	u4_puts("Klimb ")
	if tile_cur&0xf0 == 0x10 || tile_cur&0xf0 == 0x30 {
		u4_puts("up!\n")
		Party._z--
		if Party._z >= 0 {
			u4_puts("To level ")
			u4_putc(int(Party._z) + '1')
			Gra_CR()
		}
	} else {
		w_What()
	}
}

// (D)escend [dungeon]
func C_8A1F() {
	u4_puts("Descend ")
	if tile_cur&0xf0 == 0x20 || tile_cur&0xf0 == 0x30 {
		Party._z++
		u4_puts("down to level ")
		u4_putc(int(Party._z) + '1')
		Gra_CR()
	} else {
		w_What()
	}
}
