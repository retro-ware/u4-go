package u4

// Port of U4_SPELL.C (unsigned char): casting spells.

/*spells m.p. cost*/
var D_208C = [26]uint8{
	5,  /*Awaken*/
	15, /*Blink*/
	5,  /*Cure*/
	20, /*Dispel*/
	10, /*Energy*/
	15, /*Fireball*/
	40, /*Gate*/
	10, /*Heal*/
	20, /*Iceball*/
	30, /*Jinx*/
	25, /*Kill*/
	5,  /*Light*/
	5,  /*Magic Misl*/
	20, /*Negate*/
	5,  /*Open*/
	15, /*Protection*/
	20, /*Quickness*/
	45, /*Resurrect*/
	15, /*Sleep*/
	30, /*Tremor*/
	15, /*Undead*/
	15, /*View*/
	10, /*Winds*/
	15, /*X-it*/
	10, /*Y-up*/
	5,  /*Z-down*/
}

var D_8CCC int /*current spell*/

// spellScratch stands in for the bytes the C reads/writes when a map index runs off the
// map (negative combat index, or past the town map): those are never valid tiles for the
// spell, so a fresh 0 byte gives the same "Failed!".
var spellScratch uint8

// spellMap returns the map byte at a flat index of D_8742._map, or the scratch byte.
func spellMap(idx int) *uint8 {
	if idx < 0 || idx >= len(D_8742._map) {
		spellScratch = 0
		return &spellScratch
	}
	return &D_8742._map[idx]
}

// spellCombatMap is Combat_MAP that tolerates a negative index (C reads the bytes before the map).
func spellCombatMap(y, x int) *uint8 {
	if y*11+x < 0 {
		spellScratch = 0
		return &spellScratch
	}
	return Combat_MAP(y, x)
}

/*isUndead?*/
func C_636D(bp04 int) int {
	if bp04 == TIL_9C || bp04 == TIL_BC || bp04 == TIL_C4 || bp04 == TIL_E4 {
		return 1
	}
	return 0
}

func w_Failed() { /*C_6399*/
	u4_puts( /*D_20A6*/ "Failed!\n")
	hit_tile = 0
	sound(8)
}

/*try "cast"*/
/*
	.consume m.p
	.check 'N'egate
*/
func C_63B4() int {
	Party.chara[activeChara]._MP -= uint16(D_208C[D_8CCC])
	sound(10, int(D_208C[D_8CCC]))
	Gra_09()
	sound(9, D_8CCC+0x60)
	Gra_09()
	if spell_sta == 'N' {
		w_Failed()
		return 0
	}
	return 1
}

/*assert outdoors only*/
func C_6409() int {
	if CurMode != MOD_OUTDOORS {
		u4_puts( /*D_20AF*/ "Outdoors Only!\n")
		w_Failed()
		return 0
	}
	return 1
}

/*assert combat only*/
func C_6428() int {
	if CurMode < MOD_COMBAT {
		u4_puts( /*D_20BF*/ "Combat Only!\n")
		w_Failed()
		return 0
	}
	return 1
}

/*assert dungeon only*/
func C_6447() int {
	if CurMode != MOD_DUNGEON {
		u4_puts( /*D_20CD*/ "Dungeon Only!\n")
		w_Failed()
		return 0
	}
	return 1
}

/*magic projectile ?*/
func C_6466(bp04 int) {
	var loc_C, loc_D int
	var loc_A, loc_B int

	if C_6428() == 0 {
		return
	}
	AskDir( /*D_20DC*/ "Dir: ", &loc_A, &loc_B)
	if loc_A|loc_B == 0 {
		return
	}
	if C_63B4() == 0 {
		return
	}
	hit_tile = uint16(bp04)
	hit_x = Combat._charaX[activeChara]
	hit_y = Combat._charaY[activeChara]
	loc_C = -1
	for {
		if COM_CheckHitable(uint8(loc_A), uint8(loc_B)) == 0 {
			break
		}
		C_3C54()
		loc_C = COM_GetFighterId(hit_x, hit_y)
		if loc_C != -1 {
			break
		}
	}
	if loc_C == -1 {
		w_Failed()
		return
	}
	sound(6)
	switch hit_tile {
	case TIL_4D:
		loc_D = U4_RND3(64) | 16
	case TIL_4E:
		loc_D = U4_RND3(224) | 32
	case TIL_4F:
		loc_D = U4_RND3(128) | 24
	default:
		loc_D = 224 | 8
	}
	hit_tile = 0
	COM_DoDamage(loc_C, int(activeChara), uint8(loc_D))
}

func SPL_Awaken() { /*C_6558*/
	si := AskChara( /*D_20E2*/ "Who:\x12\x12\b")
	if si < 0 {
		return
	}
	if C_63B4() == 0 {
		return
	}
	if Party.chara[si]._stat != 'S' {
		w_Failed()
		return
	}
	Party.chara[si]._stat = 'G'
	if CurMode >= MOD_COMBAT {
		Fighters._chtile[si] = uint8(C_0ACF(si))
		D_944A[si] = Fighters._chtile[si]
	}
}

func SPL_Blink() { /*C_65AA*/
	var loc_A, loc_B int
	var loc_C, loc_D uint8

	if Party._tile >= TIL_14 && Party._tile != TIL_18 {
		if C_6409() == 0 {
			return
		}
		AskDir( /*D_20EA*/ "Dir: ", &loc_A, &loc_B)
		if loc_A|loc_B == 0 {
			return
		}
		if C_63B4() == 0 {
			return
		}
		if Party._x&Party._y < 0xc0 {
			loc_C = D_959C.x
			loc_D = D_959C.y
			for loc_C < 32 && loc_D < 32 {
				loc_C = uint8(int(loc_C) + loc_A)
				loc_D = uint8(int(loc_D) + loc_B)
			}
			for {
				loc_C = uint8(int(loc_C) - loc_A)
				loc_D = uint8(int(loc_D) - loc_B)
				if !((loc_C != D_959C.x || loc_D != D_959C.y) && !C_2999(*D_8742.x32x32(int(loc_D), int(loc_C)))) {
					break
				}
			}
			if D_959C.x != loc_C || D_959C.y != loc_D {
				Party._x += loc_C - D_959C.x
				Party._y += loc_D - D_959C.y
				C_26B6()
				return
			}
		}
	}
	w_Failed()
}

func SPL_Cure() { /*C_669B*/
	si := AskChara( /*D_20F0*/ "Who:\x12\x12\b")
	if si < 0 {
		return
	}
	if C_63B4() == 0 {
		return
	}
	if Party.chara[si]._stat != 'P' {
		w_Failed()
		return
	}
	Party.chara[si]._stat = 'G'
}

func SPL_Dispel() { /*C_66DA*/
	var loc_A, loc_B int
	var loc_C *uint8

	if CurMode == MOD_DUNGEON {
		if C_63B4() == 0 {
			return
		}
		loc_C = D_8742.x8x8x8(int(Party._z), DNG_Y_p(int(Party._y), int(Party._dir)), DNG_X_p(int(Party._x), int(Party._dir)))
		if *loc_C&0xf0 != 0xa0 {
			w_Failed()
			return
		}
		*loc_C = 0
		return
	}
	if CurMode <= MOD_BUILDING && Party.f_1dc != 0 {
		w_Failed()
		return
	}
	AskDir( /*D_20F8*/ "Dir: ", &loc_A, &loc_B)
	if loc_A|loc_B == 0 {
		return
	}
	if C_63B4() == 0 {
		return
	}
	if CurMode >= MOD_COMBAT {
		cy, cx := int(Combat._charaY[activeChara]), int(Combat._charaX[activeChara])
		if cy+loc_B >= 11 || cx+loc_A >= 11 {
			w_Failed()
			return
		}
		loc_C = spellCombatMap(cy+loc_B, cx+loc_A)
		if *loc_C < TIL_44 || *loc_C > TIL_47 {
			w_Failed()
			return
		}
		*loc_C = *Combat_MAP(cy, cx)
		return
	}
	if CurMode == MOD_OUTDOORS {
		loc_C = spellMap((int(D_959C.y)+loc_B)*32 + int(D_959C.x) + loc_A)
	} else {
		if int(Party._y)+loc_B >= 32 || int(Party._x)+loc_A >= 32 {
			w_Failed()
			return
		}
		loc_C = spellMap((int(Party._y)+loc_B)*32 + int(Party._x) + loc_A)
	}
	if *loc_C < TIL_44 || *loc_C > TIL_47 {
		w_Failed()
		return
	}
	*loc_C = tile_cur
}

func SPL_Energy() { /*C_6882*/
	var loc_D *uint8
	var loc_B int
	var loc_A, loc_C, loc_E int

	u4_puts( /*D_20FE*/ "Energy type? ")
	askKeys("Energy type", item("Poison", key('p')), item("Lightning", key('l')), item("Fire", key('f')), item("Sleep", key('s')))
	loc_B = u_kbread() & 0xff
	u4_toupper(&loc_B)
	switch loc_B {
	case 'P' /*oison*/ :
		loc_E = TIL_44
	case 'L' /*ightning*/ :
		loc_E = TIL_45
	case 'F' /*ire*/ :
		loc_E = TIL_46
	case 'S' /*leep*/ :
		loc_E = TIL_47
	default:
		loc_E = -1
	}
	if loc_E == -1 {
		w_Failed()
		return
	}
	u4_putc(loc_B)
	Gra_CR()
	if CurMode == MOD_DUNGEON {
		if C_63B4() == 0 {
			return
		}
		loc_D = D_8742.x8x8x8(int(Party._z), DNG_Y_p(int(Party._y), int(Party._dir)), DNG_X_p(int(Party._x), int(Party._dir)))
		if *loc_D != 0 {
			w_Failed() // quirk: no return, the field is overwritten anyway
		}
		*loc_D = uint8(loc_E&3 | 0xa0)
		return
	}
	if CurMode >= MOD_COMBAT {
		AskDir( /*D_210C*/ "Dir: ", &loc_A, &loc_C)
		if loc_A|loc_C == 0 {
			return
		}
		if C_63B4() == 0 {
			return
		}
		if int(Combat._charaY[activeChara])+loc_C < 11 && int(Combat._charaX[activeChara])+loc_A < 11 {
			loc_D = spellCombatMap(loc_C+int(Combat._charaY[activeChara]), loc_A+int(Combat._charaX[activeChara]))
			if C_2999(*loc_D) {
				*loc_D = uint8(loc_E)
				return
			}
		}
	}
	w_Failed()
}

func SPL_Fireball() { /*C_69D7*/
	C_6466(TIL_4F)
}

func SPL_Gate() { /*C_69E5*/
	if Party._tile < TIL_14 || Party._tile == TIL_18 {
		w_Failed()
		return
	}
	if C_6409() == 0 {
		return
	}
	si := AskLetter( /*D_2112*/ "To Phase:\x12\x12\b", '0', '8')
	if si < 0 || si == '0' {
		return
	}
	if C_63B4() == 0 {
		return
	}
	si -= '1'
	Party._x = D_0814[si]
	Party._y = D_081C[si]
	C_26B6()
}

func SPL_Heal() { /*C_6A40*/
	si := AskChara( /*D_211F*/ "Who?\x12\x12\b")
	if si < 0 {
		return
	}
	if C_63B4() == 0 {
		return
	}
	if !isCharaAlive(si) {
		w_Failed()
		return
	}
	HP_inc(si, U4_RND3(25)+75)
}

func SPL_Iceball() { /*C_6A82*/
	C_6466(TIL_4E)
}

func SPL_Jinx() { /*C_6A90*/
	if C_63B4() == 0 {
		return
	}
	spell_sta = 'J'
	spell_cnt = 10
}

func SPL_Kill() { /*C_6AA9*/
	C_6466(TIL_8C)
}

func SPL_Light() { /*C_6AB7*/
	if C_6447() == 0 {
		return
	}
	if C_63B4() == 0 {
		return
	}
	Party.f_1dc += 100
	if CurMode == MOD_DUNGEON {
		C_AE41()
	}
}

func SPL_MagicMisl() { /*C_6ADB*/
	C_6466(TIL_4D)
}

func SPL_Negate() { /*C_6AE9*/
	if C_63B4() == 0 {
		return
	}
	spell_sta = 'N'
	spell_cnt = 10
}

func SPL_Open() { /*C_6B02*/
	if C_63B4() == 0 {
		return
	}
	if CurMode == MOD_OUTDOORS && Party._tile == TIL_18 {
		w_Failed()
		return
	}
	if CurMode >= MOD_COMBAT {
		C_7337()
	} else {
		C_722F(-1) /*100% chance evasion*/
	}
}

func SPL_Protection() { /*C_6B36*/
	if C_63B4() == 0 {
		return
	}
	spell_sta = 'P'
	spell_cnt = 10
}

func SPL_Quickness() { /*C_6B4F*/
	if C_63B4() == 0 {
		return
	}
	spell_sta = 'Q'
	spell_cnt = 10
}

func SPL_Resurrect() { /*C_6B68*/
	if CurMode >= MOD_COMBAT {
		w_Failed()
		return
	}
	si := AskChara( /*D_2127*/ "Who:\x12\x12\b")
	if si < 0 {
		return
	}
	if C_63B4() == 0 {
		return
	}
	if Party.chara[si]._stat != 'D' {
		w_Failed()
		return
	}
	Party.chara[si]._stat = 'G'
}

func SPL_Sleep() { /*C_6BAE*/
	if C_6428() == 0 {
		return
	}
	if C_63B4() == 0 {
		return
	}
	for si := 15; si >= 0; si-- {
		if Fighters._tile[si] != 0 && C_636D(int(Fighters._tile[si])) == 0 && Fighters._tile[si] != TIL_FC && Fighters._HP[si] < u_rand_a() {
			Fighters._sleeping[si] = 1
		}
	}
}

func SPL_Tremor() { /*C_6BF8*/
	if C_6428() == 0 {
		return
	}
	if C_63B4() == 0 {
		return
	}
	shakefx()
	hit_tile = TIL_4F
	for si := 15; si >= 0; si-- {
		if Fighters._tile[si] != 0 && Fighters._HP[si] < 192 {
			if U4_RND1(1) != 0 {
				if U4_RND1(1) != 0 {
					Fighters._HP[si] = 23
				}
				continue
			}
			hit_x = Combat._npcX[si]
			hit_y = Combat._npcY[si]
			C_3C54()
			sound(6)
			COM_DoDamage(si, int(activeChara), 0xff)
		}
	}
	hit_tile = 0
}

func SPL_Undead() { /*C_6C71*/
	if C_6428() == 0 {
		return
	}
	if C_63B4() == 0 {
		return
	}
	for si := 15; si >= 0; si-- {
		if C_636D(int(Fighters._tile[si])) != 0 && U4_RND1(1) != 0 && Fighters._HP[si] > 23 {
			Fighters._HP[si] = 23
		}
	}
}

func SPL_View() { /*C_6CB2*/
	if C_63B4() == 0 {
		return
	}
	C_C403()
}

func SPL_Winds() { /*C_6CC3*/
	var bp_02, bp_04 int

	if C_6409() == 0 {
		return
	}
	AskDir( /*D_212F*/ "From Dir: ", &bp_02, &bp_04)
	if bp_02|bp_04 == 0 {
		return
	}
	if C_63B4() == 0 {
		return
	}
	if bp_02 != 0 {
		if bp_02 == 1 {
			WindDir = DIR_E
		} else {
			WindDir = DIR_W
		}
	} else {
		if bp_04 == 1 {
			WindDir = DIR_S
		} else {
			WindDir = DIR_N
		}
	}
}

func SPL_X_it() { /*C_6D22*/
	if C_6447() == 0 {
		return
	}
	if C_63B4() == 0 {
		return
	}
	Party._z = -1
}

func SPL_Y_up() { /*C_6D3D*/
	var loc_A int
	var loc_B, loc_C uint8

	if C_6447() == 0 {
		return
	}
	if C_63B4() == 0 {
		return
	}
	if Party._loc == 0x18 {
		w_Failed()
		return
	}
	Party._z--
	if Party._z < 0 {
		return
	}
	for loc_A = 32; loc_A > 0; loc_A-- {
		loc_B = uint8(U4_RND1(7))
		loc_C = uint8(U4_RND1(7))
		if *D_8742.x8x8x8(int(Party._z), int(loc_C), int(loc_B)) == 0 {
			break
		}
	}
	if loc_A == 0 {
		Party._z++
		w_Failed()
		return
	}
	Party._x = loc_B
	Party._y = loc_C
}

func SPL_Z_down() { /*C_6DC1*/
	var loc_A int
	var loc_B, loc_C uint8

	if C_6447() == 0 {
		return
	}
	if C_63B4() == 0 {
		return
	}
	if Party._loc == 0x18 || Party._z == 7 {
		w_Failed()
		return
	}
	Party._z++
	for loc_A = 32; loc_A > 0; loc_A-- {
		loc_B = uint8(U4_RND1(7))
		loc_C = uint8(U4_RND1(7))
		if *D_8742.x8x8x8(int(Party._z), int(loc_C), int(loc_B)) == 0 {
			break
		}
	}
	if loc_A == 0 {
		Party._z--
		w_Failed()
		return
	}
	Party._x = loc_B
	Party._y = loc_C
}

var D_216E = [26]func(){
	SPL_Awaken,
	SPL_Blink,
	SPL_Cure,
	SPL_Dispel,
	SPL_Energy,
	SPL_Fireball,
	SPL_Gate,
	SPL_Heal,
	SPL_Iceball,
	SPL_Jinx,
	SPL_Kill,
	SPL_Light,
	SPL_MagicMisl,
	SPL_Negate,
	SPL_Open,
	SPL_Protection,
	SPL_Quickness,
	SPL_Resurrect,
	SPL_Sleep,
	SPL_Tremor,
	SPL_Undead,
	SPL_View,
	SPL_Winds,
	SPL_X_it,
	SPL_Y_up,
	SPL_Z_down,
}

func CMD_Cast() { /*C_6E4A*/
	u4_puts( /*D_213A*/ "Cast Spell!\n")
	if CurMode < MOD_COMBAT {
		activeChara = int16(AskChara( /*D_2147*/ "Player:\x12\x12\b"))
		if activeChara < 0 {
			return
		}
		if !isCharaConscious(int(activeChara)) {
			w_Disabled()
			return
		}
	}
	Gra_13()
	C_4C42(0)
	txt_Y = 23
	txt_X = 24
	D_8CCC = AskLetter( /*D_2152*/ "Spell:\x12\x12\b", 'A', 'Z')
	Gra_13()
	C_4649()
	dspl_Stats()
	if CurMode >= MOD_COMBAT {
		Gra_11(int(activeChara))
	}
	if D_8CCC < 0 {
		return
	}
	D_8CCC -= 'A'
	u4_puts(D_1E98[101+D_8CCC])
	u4_puts( /*D_215C*/ "!\n")
	/*-- check mixture --*/
	if Party._mixtures[D_8CCC] == 0 {
		w_NoneLeft()
		return
	}
	Party._mixtures[D_8CCC]--
	/*-- check m.p --*/
	if Party.chara[activeChara]._MP < uint16(D_208C[D_8CCC]) {
		u4_puts( /*D_215F*/ "M.P. too low!\n")
		w_Failed()
		return
	}
	/*-- cast spell --*/
	D_216E[D_8CCC]()
	dspl_Stats()
}
