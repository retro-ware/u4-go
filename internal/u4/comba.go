package u4

// Port of U4_COMBA.C (unsigned char): the combat loop, attacks and damage.

/*C_59D5*/
func COM_GetFighterId(_x, _y uint8) int {
	si := 15
	for ; si >= 0; si-- {
		if Fighters._tile[si] != 0 && Combat._npcX[si] == _x && Combat._npcY[si] == _y {
			break
		}
	}
	return si
}

// party all out ?
func C_5A04() int {
	for si := int(Party.f_1d8) - 1; si >= 0; si-- {
		if isCharaAlive(si) {
			return 0
		}
	}
	return 1
}

/*C_5A28*/
func IsCombatEnded() int {
	si := 7
	for ; si >= 0; si-- {
		if Fighters._chtile[si] != 0 {
			break
		}
	}
	if si == -1 {
		return 1
	}
	if CurMode == MOD_COM_ROOM {
		return 0
	}
	for si = 15; si >= 0; si-- {
		if Fighters._tile[si] != 0 {
			return 0
		}
	}
	return 1
}

/*C_5A6B*/
func COM_main() {
	var si uint16
	var bp_04 uint16

	u_kbflush()
	bp_04 = 0
	D_96EE = 0
	D_96F4 = 0
	for {
		// activeChara (int) < f_1d8 (unsigned): unsigned compare
		for activeChara = 0; uint16(activeChara) < Party.f_1d8 && IsCombatEnded() == 0; activeChara++ {
			/*C_5A9E*/
			a := int(activeChara)
			if Fighters._chtile[a] != 0 && isCharaConscious(a) {
				D_95C8 = 4
				Gra_11(a)
				activeCharaX = int16(Combat._charaX[a])
				activeCharaY = uint16(Combat._charaY[a])
				u4_puts(cstr(Party.chara[a]._name[:]))
				u4_puts(" with ")
				u4_puts(D_1E98[37+int(Party.chara[a]._weapon)])
				Gra_CR()
				u4_putc(0x10)
				D_95C8 = 2
				C_3C54()
				ask(pCmd)
				u_delay(25, 1)
				if u_kbhit() != 0 {
					si = uint16(u_kbread())
				} else {
					si = KBD_SPACE
				}
				if u4_isupper(int(si & 0xff)) {
					si = (si & 0xff00) | uint16(u4_lower(int(si&0xff)))
				}
				switch si {
				case KBD_SPACE:
					w_Pass()
				case KBD_UP_, KBD_UP:
					C_7B89()
				case KBD_LEFT_, KBD_LEFT:
					C_7BFD()
				case KBD_DOWN_, KBD_DOWN:
					C_7BAF()
				case KBD_RIGHT_, KBD_RIGHT:
					C_7BD5()
				case KBD_A:
					C_61D1() // (A)ttack(combat)
				case KBD_C:
					CMD_Cast()
				case KBD_G:
					C_73B8() // (G)et chest(combat)
				case KBD_R:
					CMD_Ready()
				case KBD_U:
					CMD_Use()
				case KBD_V:
					CMD_Volume()
				case KBD_Z:
					CMD_Ztats()
				case KBD_B, KBD_D, KBD_E, KBD_F, KBD_H, KBD_I, KBD_J, KBD_K, KBD_L, KBD_M, KBD_N,
					KBD_O, KBD_P, KBD_Q, KBD_S, KBD_T, KBD_W, KBD_X, KBD_Y:
					w_NotHere()
				case KBD_CTRL_S:
					if bp_04 == KBD_ALT_Z {
						C_1C21()
						break
					}
					fallthrough
				default:
					u4_puts("Bad command\n")
					sound(2)
					Gra_11(a)
					activeChara--
					continue
				}
				Gra_11(a)
				activeCharaX = -1
				dspl_Stats()
			}
			if D_07F8 == 0 {
				D_07F8 = 1
				activeChara--
			} else {
				C_5D14()
			}
			/*C_5C09*/
		}
		/*C_5CEF:*/
		C_9F7B()
		bp_04 = si
		if IsCombatEnded() != 0 {
			break
		}
	}
	if C_5A04() == 0 {
		C_837A()
	}
	D_07F8 = 1
}

func C_5D14() {
	a := int(activeChara)
	if Fighters._chtile[a] == 0 {
		return
	}
	if isCharaConscious(a) {
		C_9209()
	}
	if !isCharaAlive(a) {
		Fighters._chtile[a] = 0
	} else if isCharaConscious(a) && spell_sta == 'Q' && U4_RND1(1) != 0 {
		// Quickness!
		activeChara--
		return
	} else if Party.chara[a]._stat == 'S' && U4_RND1(7) == 0 {
		Party.chara[a]._stat = 'G'
		Fighters._chtile[a] = uint8(C_0ACF(a))
		D_944A[a] = Fighters._chtile[a]
		dspl_Stats()
	}
	food_dec(1)
}

/*C_5DAB*/
// damage is generic: callers pass int (RND) or uint8; C truncates to unsigned char.
func COM_DoDamage[T ~int | ~uint8](_npcId, _charaId int, _damage_ T) {
	_damage := uint8(_damage_)
	u4_puts(C_1513(Fighters._tile[_npcId]))
	u4_putc(' ')
	killed := false
	if Fighters._tile[_npcId] != TIL_5E {
		// C: (HP -= damage) < 0 on an unsigned char; read as the intended borrow test.
		killed = Fighters._HP[_npcId] < _damage
		Fighters._HP[_npcId] -= _damage
	}
	if killed {
		u4_puts("Killed!\n")
		if _charaId != -1 {
			di := int(D_23D2[C_7C25(Fighters._tile[_npcId])])/16 + 1
			XP_inc(_charaId, di)
			u4_puts("Exp. ")
			u4_putl(int32(di), 1, ' ')
			Gra_CR()
		}
		Fighters._tile[_npcId] = 0
	} else {
		loc_B := int(D_23D2[C_7C25(Fighters._tile[_npcId])]) / 2
		loc_A := loc_B         // 50% HP
		loc_B >>= 1            // 25% HP
		loc_C := loc_B + loc_A // 75% HP
		loc_D := int(Fighters._HP[_npcId])
		if loc_D < 24 {
			u4_puts("Fleeing!\n")
		} else if loc_D < loc_B {
			u4_puts("Critical!\n")
		} else {
			Gra_CR()
			if loc_D < loc_A {
				u4_puts("Heavily ")
			} else if loc_D < loc_C {
				u4_puts("Lightly ")
			} else {
				u4_puts("Barely ")
			}
			u4_puts("Wounded!\n")
		}
	}
}

/*C_5EA2*/
func COM_CheckHitable(_dir_x, _dir_y uint8) int {
	loc_C := _dir_x + hit_x
	loc_A := _dir_y + hit_y
	if loc_C > 10 || loc_A > 10 {
		return 0
	}
	loc_B := int(*Combat_MAP(int(loc_A), int(loc_C)))
	if loc_B == TIL_49 {
		return 0
	}
	if loc_B <= TIL_80 && loc_B > TIL_02 && loc_B != TIL_48 {
		if loc_B < TIL_31 || loc_B > TIL_34 {
			if !C_2999(uint8(loc_B)) {
				return 0
			}
		}
	}
	hit_x = loc_C
	hit_y = loc_A
	return 1
}

/*C_5F26*/
func DrawBlow(_range int) {
	if Party.chara[activeChara]._weapon != 11 || _range < 2 {
		return
	}
	bp_02 := u4_sign(int16(Combat._charaX[activeChara]) - int16(hit_x))
	bp_04 := u4_sign(int16(Combat._charaY[activeChara]) - int16(hit_y))
	hit_tile = TIL_4D
	for _range != 0 {
		COM_CheckHitable(uint8(bp_02), uint8(bp_04))
		C_3C54()
		_range--
	}
	hit_tile = 0
}

/*C_5F9D*/
func w_missed(_range int) {
	if _range == 0 {
		sound(4)
	}
	hit_tile = 0
	u4_puts("Missed!\n")
	if Party.chara[activeChara]._weapon == 9 && hit_x < 11 && hit_y < 11 && *Combat_MAP(int(hit_y), int(hit_x)) >= TIL_03 {
		*Combat_MAP(int(hit_y), int(hit_x)) = TIL_46
	}
	C_3C54()
	DrawBlow(_range)
}

/*C_6012*/
func TryDamage(_npcId, _range int) {
	loc_C := &Party.chara[activeChara]
	if Party._loc == 0x18 && loc_C._weapon < 11 {
		// abyss without good weapons
		w_missed(_range)
		return
	}
	if loc_C._weapon == 14 {
		hit_tile = TIL_4E
	} else {
		hit_tile = TIL_4F
	}
	// dexterity test
	if loc_C._dex < 40 && uint16(U4_RND1(0xff)) > loc_C._dex+0x80 {
		w_missed(_range)
		return
	}
	// success
	loc_B := Combat_MAP(int(Combat._npcY[_npcId]), int(Combat._npcX[_npcId]))
	if loc_C._weapon == 9 && *loc_B >= TIL_03 {
		*loc_B = TIL_46
	}
	// damage points
	loc_A := int(int16(loc_C._str))
	loc_A = int(int16(loc_A + int(D_2450[loc_C._weapon])))
	loc_A = SafeModulo(U4_RND1(0xff), u4_min(loc_A, 0xff))
	C_3C54()
	sound(6)
	hit_tile = 0
	COM_DoDamage(_npcId, int(activeChara), loc_A)
	DrawBlow(_range)
}

// projectile weapon?
func C_60F1(_dir_x, _dir_y int) {
	loc_D := 11
	loc_C := 0
	loc_B := 0 // note: stays uninitialised in C when the oil range is '0' (the loop is never entered)
	loc_A := &Party.chara[activeChara]
	// -- For OIL, ask range --
	if loc_A._weapon == 9 {
		loc_D = AskLetter("Range:\x12\x12\b", '0', '9')
		loc_D -= '0'
		if loc_D < 0 {
			return
		}
	}
	// -- if denumbrable weapon, decrement --
	if loc_A._weapon == 2 || loc_A._weapon == 9 {
		if Party._weapons[loc_A._weapon] == 0 {
			loc_A._weapon = 0
			u4_puts("Last one!\n")
		} else {
			Party._weapons[loc_A._weapon]--
		}
	}
	if loc_A._weapon != 14 {
		hit_tile = TIL_4D
	} else {
		hit_tile = TIL_4E
	}
	sound(4)
	for ; loc_D != 0; loc_D-- {
		if COM_CheckHitable(uint8(_dir_x), uint8(_dir_y)) == 0 {
			w_missed(loc_C)
			return
		}
		loc_C++
		C_3C54()
		if loc_B = COM_GetFighterId(hit_x, hit_y); loc_B != -1 {
			break
		}
	}
	hit_tile = 0
	if loc_B == -1 {
		w_missed(loc_C)
		return
	}
	TryDamage(loc_B, loc_C)
}

// (A)ttack [combat]
func C_61D1() {
	loc_C := &Party.chara[activeChara]
	var loc_A, loc_B int
	AskDir("Dir: ", &loc_A, &loc_B)
	if loc_A|loc_B == 0 {
		return
	}
	hit_x = Combat._charaX[activeChara]
	hit_y = Combat._charaY[activeChara]

	if D_2468[loc_C._weapon] == 0 { // is not range weapon?
		if loc_C._weapon == 10 { // Halberd
			hit_x += uint8(loc_A)
			hit_y += uint8(loc_B)
		}
		if COM_CheckHitable(uint8(loc_A), uint8(loc_B)) != 0 {
			hit_tile = TIL_4D
			C_3C54()
			hit_tile = 0
			sound(4)

			if loc_D := COM_GetFighterId(hit_x, hit_y); loc_D != -1 {
				TryDamage(loc_D, 1)
				return
			}
			if loc_C._weapon == 2 { // dagger?
				C_60F1(loc_A, loc_B)
				return
			}
		}
		w_missed(0)
	} else {
		C_60F1(loc_A, loc_B)
	}
}

/*C_628F*/
func CMD_Attack() {
	var loc_A, loc_B int

	if CurMode <= MOD_BUILDING && Party.f_1dc != 0 {
		u4_puts("Attack\n")
		w_DriftOnly()
		return
	}

	AskDir("Attack: ", &loc_A, &loc_B)
	if loc_A|loc_B == 0 {
		return
	}
	loc_C := C_0A8B(uint8(int(Party._x)+loc_A), uint8(int(Party._y)+loc_B))
	if loc_C == -1 ||
		D_8742._npc._tile[loc_C] == TIL_8C ||
		D_8742._npc._tile[loc_C] == TIL_02 ||
		D_8742._npc._tile[loc_C] == TIL_8E {
		u4_puts("Nothing to Attack!\n")
		return
	}
	if CurMode == MOD_BUILDING {
		for di := 31; di >= 0; di-- {
			if D_8742._npc._tile[di] == TIL_50 || D_8742._npc._tile[di] == TIL_5E {
				D_8742._npc._var[di] = 0xff
				D_8742._npc._tlkidx[di] = 0
			}
		}
	}
	if CurMode == MOD_BUILDING || C_0AFE(D_8742._npc._tile[loc_C]) != 0 {
		karma_dec(&Party._compa, 5)
		karma_dec(&Party._justi, 5)
		karma_dec(&Party._honor, 5)
	}
	C_7DFE(loc_C)
}
