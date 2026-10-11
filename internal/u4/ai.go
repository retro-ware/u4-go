package u4

// Port of U4_AI.C (signed char): town/dungeon creature shuffling and the combat AI.

// aiCombatMap is Combat_MAP(y, x) with the C's unchecked indexing: outside the 11x11 map
// (creatures that fled, probes at x/y >= 11) the C reads whatever follows; we read 0.
func aiCombatMap(y, x int) uint8 {
	i := y*11 + x
	if i < 0 || i >= len(Combat._map) {
		return 0
	}
	return Combat._map[i]
}

func C_93A5(bp08, bp06 uint8, bp04 int) bool {
	bp_02 := *D_8742.x8x8x8(int(Party._z), int(bp06), int(bp08))
	if bp_02&0x0f == 0 {
		if bp_02&0xe0 != 0x80 &&
			bp_02&0xf0 != 0xa0 &&
			bp_02&0xf0 != 0xd0 &&
			bp_02&0xf0 != 0xf0 {
			if U4_RND1(7) == 0 || D_8742._npc._old_x[bp04] != bp08 || D_8742._npc._old_y[bp04] != bp06 {
				return true
			}
		}
	}
	return false
}

func C_9414() {
	var loc_A, loc_C, loc_D int
	var loc_E, loc_F uint8
	n := &D_8742._npc

	for loc_B := 31; loc_B >= 0; loc_B-- {
		if n._tile[loc_B] != 0 &&
			int16(n._var[loc_B]) == Party._z &&
			n._tile[loc_B] != TIL_B0 &&
			n._tile[loc_B] != TIL_AC {
			if Party._x != n._x[loc_B] || Party._y != n._y[loc_B] {
				for loc_D = 7; loc_D >= 0; loc_D-- {
					loc_A = isign(sc(u_rand_a()))
					loc_C = isign(sc(u_rand_a()))
					if U4_RND1(1) != 0 && loc_A != 0 {
						loc_E = uint8(int(n._x[loc_B])+loc_A) & 7
						loc_F = n._y[loc_B]
						if C_93A5(loc_E, loc_F, loc_B) {
							break
						}
					}
					if loc_C != 0 {
						loc_E = n._x[loc_B]
						loc_F = uint8(int(n._y[loc_B])+loc_C) & 7
						if C_93A5(loc_E, loc_F, loc_B) {
							break
						}
					}
					loc_E = uint8(int(n._x[loc_B])+loc_A) & 7
					loc_F = n._y[loc_B]
					if C_93A5(loc_E, loc_F, loc_B) {
						break
					}
				}
				if loc_D != -1 {
					*D_8742.x8x8x8(int(Party._z), int(loc_F), int(loc_E)) |= uint8((int(n._tile[loc_B])-TIL_90)/4 + 1)
					*D_8742.x8x8x8(int(Party._z), int(n._y[loc_B]), int(n._x[loc_B])) &= 0xf0
					n._old_x[loc_B] = n._x[loc_B]
					n._old_y[loc_B] = n._y[loc_B]
					n._x[loc_B] = loc_E
					n._y[loc_B] = loc_F
				}
			}
		}
	}
	t_callback()
}

func C_95AA() {
	n := &D_8742._npc
	for loc_A := int(Party._z)*4 + 1; loc_A >= 0; loc_A-- {
		if (loc_A >> 2) < int(Party._z) {
			break
		}
		if n._tile[loc_A] == 0 {
			loc_B := uint8(U4_RND1(7))
			loc_D := uint8(U4_RND1(7))
			if Party._x != loc_B && Party._y != loc_D {
				if *D_8742.x8x8x8(int(Party._z), int(loc_D), int(loc_B)) == 0 {
					loc_C := uint8((U4_RND1(3)+int(Party._z))*4 + TIL_90)
					if loc_C != TIL_AC {
						n._tile[loc_A] = loc_C
						n._x[loc_A] = loc_B
						n._old_x[loc_A] = loc_B
						n._y[loc_A] = loc_D
						n._old_y[loc_A] = loc_D
						n._var[loc_A] = uint8(Party._z)
						*D_8742.x8x8x8(int(Party._z), int(loc_D), int(loc_B)) = ((loc_C - 0x90) >> 2) + 1
					}
				}
			}
		}
	}
}

func C_9683() int {
	for si := int(Party.f_1d8); si > 0; {
		si--
		if isCharaAlive(si) && hit_x == Combat._charaX[si] && hit_y == Combat._charaY[si] {
			return si
		}
	}
	return -1
}

func C_96B9(bp06, bp04 int) bool {
	Gra_11(bp06)
	C_3C54()
	sound(7)
	hit_tile = 0
	C_3C54()
	bp_02 := uint16(U4_RND4(int(D_23D2[C_7C25(Fighters._tile[bp04])]) >> 2))
	bp_02 = (bp_02>>4)*10 + (bp_02 % 10)
	Gra_11(bp06)
	if hitChara(bp06, int(bp_02)) == 0 {
		dspl_Stats()
		return false
	}
	u4_puts(cstr(Party.chara[bp06]._name[:]))
	u4_puts(/*D_29CA*/ " is Killed!\n")
	karma_inc(&Party._sacri, 1)
	dspl_Stats()
	C_10FD()

	return true
}

func C_9764(bp04 int) {
	Gra_11(bp04)
	C_3C54()
	sound(7)
	Gra_11(bp04)
	hit_tile = 0
}

// monster projectile?: bp0a: type, bp08: monster id
func C_978C(bp0a uint8, bp08, bp06, bp04 int) {
	hit_tile = uint16(bp0a)
	hit_x = Combat._npcX[bp08]
	hit_y = Combat._npcY[bp08]
	sound(3)
	si := -1
	for {
		if COM_CheckHitable(uint8(bp06), uint8(bp04)) == 0 {
			break
		}
		C_3C54()
		if si = C_9683(); si != -1 {
			break
		}
	}
	if si == -1 {
		if hit_tile == TIL_4C {
			*Combat_MAP(int(hit_y), int(hit_x)) = TIL_4C
		}
		hit_tile = 0
		return
	}
	Gra_CR()
	u4_puts(cstr(Party.chara[si]._name[:]))
	u4_putc(' ')
	switch hit_tile {
	case TIL_44:
		u4_puts(/*D_29D7*/ "Poisoned!\n")
		C_9764(si)
		if U4_RND1(1) != 0 || Party.chara[si]._stat != 'G' {
			u4_puts(/*D_29E2*/ "Failed.\n")
		} else {
			Party.chara[si]._stat = 'P'
		}
	case TIL_45:
		u4_puts(/*D_29EB*/ "Electrified!\n")
		C_96B9(si, bp08)
	case TIL_46:
		u4_puts(/*D_29F9*/ "Fiery Hit!\n")
		C_96B9(si, bp08)
	case TIL_47:
		u4_puts(/*D_2A05*/ "Slept!\n")
		C_9764(si)
		if U4_RND1(1) != 0 || Party.chara[si]._stat != 'G' {
			u4_puts(/*D_2A0D*/ "Failed.\n")
		} else {
			Party.chara[si]._stat = 'S'
			Fighters._chtile[si] = TIL_38
		}
	case TIL_4C:
		u4_puts(/*D_2A16*/ "Lava Hit!\n")
		C_96B9(si, bp08)
	case TIL_4E:
		u4_puts(/*D_2A21*/ "Magical Hit!\n")
		C_96B9(si, bp08)
	default:
		u4_puts(/*D_2A2F*/ "Hit!\n")
		C_96B9(si, bp08)
	}
	hit_tile = 0
	dspl_Stats()
}

func C_98E4(bp08, bp06 uint8, bp04 int) bool {
	bp_02 := aiCombatMap(int(bp06), int(bp08)) // read before the range check, as in the C
	if C_29EF(bp_02) {
		return false
	}
	if bp08 >= 11 || bp06 >= 11 {
		return Fighters._HP[bp04] < 24
	}
	if Fighters._x[bp04] == bp08 && Fighters._y[bp04] == bp06 && U4_RND1(3) != 0 {
		return false
	}

	for bp_04 := 15; bp_04 >= 0; bp_04-- {
		if Fighters._tile[bp_04] != 0 && Combat._npcX[bp_04] == bp08 && Combat._npcY[bp_04] == bp06 {
			return false
		}
	}
	for bp_04 := 7; bp_04 >= 0; bp_04-- {
		if Fighters._chtile[bp_04] != 0 && Combat._charaX[bp_04] == bp08 && Combat._charaY[bp_04] == bp06 {
			return false
		}
	}

	if Fighters._tile[bp04] < 0x80 {
		return C_2999(aiCombatMap(int(bp06), int(bp08)))
	}
	if Fighters._tile[bp04] < 0x8e {
		return bp_02 < 3
	}
	switch Fighters._gtile[bp04] {
	case TIL_B4, TIL_8E, TIL_94, TIL_F0, TIL_F8, TIL_FC:
		return bp_02 < 3 || C_2999(bp_02)
	case TIL_9C, TIL_EC:
		return bp_02 >= 3 && bp_02 != 0x45
	default:
		return C_2999(bp_02)
	}
}

func C_9A41(bp04 uint8) uint8 {
	switch bp04 {
	case TIL_20, TIL_5E, TIL_8A, TIL_E0, TIL_E4, TIL_F0:
		return TIL_4E // magic sphere
	case TIL_84, TIL_A4:
		return TIL_4D // Missile
	case TIL_86:
		return TIL_45 // energy field
	case TIL_88, TIL_F4, TIL_F8:
		return TIL_4F // flash
	case TIL_98, TIL_AC, TIL_CC:
		return TIL_44 // poison field
	case TIL_B8:
		return TIL_47 // sleep field
	case TIL_D0, TIL_D8:
		return TIL_37 // rocks
	case TIL_E8:
		return TIL_4C // lava
	case TIL_B0, TIL_FC:
		return uint8(U4_RND1(3) + TIL_44) // [poison|energy|fire] field
	default:
		return 0
	}
}

/*C_9B03*/
func CheckJinx(_x, _y uint8) bool {
	if spell_sta == 'J' {
		if bp_02 := COM_GetFighterId(_x, _y); bp_02 != -1 {
			hit_tile = TIL_4F
			hit_x = _x
			hit_y = _y
			C_3C54()
			hit_tile = 0
			sound(6)
			COM_DoDamage(bp_02, -1, U4_RND1(0x3f))
			C_3C54()
			return true
		}
	}
	return false
}

// steal food
func C_9B6B() {
	Party._food -= 2500
	if Party._food < 0 {
		Party._food = 0
	}
	dspl_Stats()
	sound(8)
}

// steal gold
func C_9BA6() {
	if U4_RND1(3) == 0 {
		si := U4_RND1(0x3f)
		Party._gold = uint16(u4_max(int(Party._gold)-si, 0))
		dspl_Stats()
		sound(8)
	}
}

// contact attack?
func C_9BE5(bp06, bp04 int) int {
	hit_x = Combat._charaX[bp04]
	hit_y = Combat._charaY[bp04]
	hit_tile = TIL_4D
	C_3C54()
	sound(5)
	hit_tile = 0
	if (spell_sta == 'P' && U4_RND1(1) != 0) || D_2460[Party.chara[bp04]._armor] >= u_rand_a() {
		C_3C54()
	} else if !C_96B9(bp04, bp06) {
		return 1
	}
	return 0
}

func C_9C56(bp04 int) {
	Gra_CR()
	u4_puts(C_1513(Fighters._tile[bp04]))
	u4_puts(/*D_2A35*/ " Flees!\n")
	if C_0AFE(Fighters._tile[bp04]) != 0 {
		karma_inc(&Party._compa, 1)
		karma_inc(&Party._justi, 1)
	}
	sound(8)
	Fighters._tile[bp04] = 0
	Fighters._HP[bp04] = 0
}

// move enemies during fight: bp08: monster id, bp06: chara id
func C_9CBC(bp08, bp06, bp04 int) {
	var loc_B int
	var loc_D, loc_E uint8

	loc_A := isign(int(Combat._charaX[bp06]) - int(Combat._npcX[bp08]))
	loc_C := isign(int(Combat._charaY[bp06]) - int(Combat._npcY[bp08]))
	// Mimic
	if Fighters._tile[bp08] == TIL_AC {
		Fighters._gtile[bp08] = TIL_3C
		if bp04 >= 5 {
			goto C_9D51
		}
		Fighters._gtile[bp08] = TIL_AC
	}
	// attempt projectile
	if U4_RND1(3) == 0 {
		if bp_0c := C_9A41(Fighters._tile[bp08]); bp_0c != 0 {
			if bp_0c != TIL_4E || spell_sta != 'N' {
				C_978C(bp_0c, bp08, loc_A, loc_C)
				t_callback()
				return
			}
		}
	}
C_9D51:
	// balron or reaper
	if U4_RND1(3) == 0 {
		if Fighters._tile[bp08] == TIL_FC || Fighters._tile[bp08] == TIL_B0 {
			if spell_sta != 'N' {
				t_callback()
				Gra_09()
				sound(9, 0x80)
				Gra_09()
				for loc_B = int(Party.f_1d8) - 1; loc_B >= 0; loc_B-- {
					if Fighters._chtile[loc_B] != 0 && Party.chara[loc_B]._stat == 'G' && U4_RND1(1) != 0 {
						Party.chara[loc_B]._stat = 'S'
						Fighters._chtile[loc_B] = TIL_38
					}
				}
				dspl_Stats()
				return
			}
		}
	}
	// C_9DE5 (the C jumps into it from the other branches)
	if Fighters._HP[bp08] < 24 {
		loc_A = -loc_A
		loc_C = -loc_C
	} else if bp04 < 2 {
		// contact?
		switch Fighters._tile[bp08] {
		case TIL_A8:
			C_9B6B()
		case TIL_C8:
			C_9BA6()
		}
		C_9BE5(bp08, bp06) // contact attack?
		return
	}
	if Fighters._tile[bp08] == TIL_AC || Fighters._tile[bp08] == TIL_B0 {
		return
	}
	for loc_B = 2; loc_B > 0; loc_B-- {
		if U4_RND1(1) != 0 && loc_A != 0 {
			loc_D = uint8(int(Combat._npcX[bp08]) + loc_A)
			loc_E = Combat._npcY[bp08]
			if CheckJinx(loc_D, loc_E) {
				return
			}
			if C_98E4(loc_D, loc_E, bp08) {
				break
			}
		}
		if loc_C != 0 {
			loc_D = Combat._npcX[bp08]
			loc_E = uint8(int(Combat._npcY[bp08]) + loc_C)
			if CheckJinx(loc_D, loc_E) {
				return
			}
			if C_98E4(loc_D, loc_E, bp08) {
				break
			}
		}
		if loc_A != 0 {
			loc_D = uint8(int(Combat._npcX[bp08]) + loc_A)
			loc_E = Combat._npcY[bp08]
			if CheckJinx(loc_D, loc_E) {
				return
			}
			if C_98E4(loc_D, loc_E, bp08) {
				break
			}
		}
		loc_A = isign(sc(u_rand_a()))
		loc_C = isign(sc(u_rand_a()))
	}
	if loc_B != 0 {
		Fighters._x[bp08] = Combat._npcX[bp08]
		Fighters._y[bp08] = Combat._npcY[bp08]
		Combat._npcX[bp08] = loc_D
		Combat._npcY[bp08] = loc_E
		if loc_D >= 11 || loc_E >= 11 {
			C_9C56(bp08) // ... Flees
			return
		}
		C_3C54()
		return
	}
	if bp04 < 2 {
		C_9BE5(bp08, bp06) // contact attack?
	}
}

// combat ai
func C_9F7B() {
	t_callback()

	for loc_A := 15; loc_A >= 0; loc_A-- {
		if Fighters._tile[loc_A] != 0 {
			if Fighters._sleeping[loc_A] == 0 || U4_RND1(7) == 0 {
				Fighters._sleeping[loc_A] = 0
				if Fighters._tile[loc_A] == TIL_EC {
					spell_sta = 'N'
					spell_cnt = 2
					dspl_Stats()
				} else if Fighters._tile[loc_A] == TIL_DC && U4_RND1(7) == 0 {
					var loc_D, loc_E uint8
					for {
						loc_D = uint8(U4_RND3(11))
						loc_E = uint8(U4_RND3(11))
						if C_98E4(loc_D, loc_E, loc_A) {
							break
						}
					}
					Combat._npcX[loc_A] = loc_D
					Combat._npcY[loc_A] = loc_E
				}
				loc_D := Combat._npcX[loc_A]
				loc_E := Combat._npcY[loc_A]
				loc_F := 99
				loc_C := -1
				for loc_B := int(Party.f_1d8) - 1; loc_B >= 0; loc_B-- {
					if isCharaAlive(loc_B) && Fighters._chtile[loc_B] != 0 {
						si := u4_abs(int(Combat._charaX[loc_B])-int(loc_D)) +
							u4_abs(int(Combat._charaY[loc_B])-int(loc_E))
						if loc_F > si {
							loc_F = si
							loc_C = loc_B
						}
					}
				}
				if loc_C != -1 {
					C_9CBC(loc_A, loc_C, loc_F)
				}
			}
			switch aiCombatMap(int(Combat._npcY[loc_A]), int(Combat._npcX[loc_A])) {
			case TIL_46, TIL_4C:
				if Fighters._tile[loc_A] == TIL_E8 || Fighters._tile[loc_A] >= TIL_F0 {
					break
				}
				fallthrough
			case TIL_44:
				sound(6)
				COM_DoDamage(loc_A, -1, U4_RND1(0x7f))
			case TIL_47:
				if C_636D(int(Fighters._tile[loc_A])) == 0 && Fighters._HP[loc_A] >= u_rand_a() {
					Fighters._sleeping[loc_A] = 1
					sound(6)
				}
			}
		}
	}
	MP_recover()
	Party._moves++
	if spell_cnt != 0 {
		spell_cnt--
		if spell_cnt == 0 {
			spell_sta = 7
		}
	}
}
