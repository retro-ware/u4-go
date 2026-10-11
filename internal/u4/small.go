package u4

// Ports of the small command modules: U4_HOLE.C, U4_Q_N_V.C (unsigned char), U4_EVT.C,
// U4_MIX.C and U4_LOCAT.C.

// ---- U4_HOLE.C ----

var D_26B8 = [8]uint8{TIL_C0, TIL_C4, TIL_C8, TIL_CC, TIL_B4, TIL_A0, TIL_A4, TIL_DC}

func C_8A5A() {
	u4_puts("Ambushed!\n")
	D_9772 = Party._x
	D_9140 = Party._y
	D_9452 = D_26B8[U4_RND1(7)]
	D_96F8 = D_9452
	for si := int(Party.f_1d8) - 1; si >= 0; si-- {
		if Party.chara[si]._stat == 'G' {
			Party.chara[si]._stat = 'S'
		}
	}
	dspl_Stats()
	C_7FD7()
}

func CMD_HoleUp() { /*C_8AB0*/
	u4_puts("Hole up & Camp\n")
	if CurMode != MOD_DUNGEON && Party._loc != 0 {
		w_NotHere()
		return
	}
	if Party._tile != TIL_1F {
		w_OnlyOnFoot()
		return
	}
	name := "CAMP.CON"
	if CurMode == MOD_DUNGEON {
		name = "CAMP.DNG"
	}
	if Load(name, sizeof(&Combat), &Combat) == -1 {
		exit(3)
	}
	// the loop clears 32 bytes from Fighters._tile: _tile and _gtile
	Fighters._tile = [16]uint8{}
	Fighters._gtile = [16]uint8{}
	for si := int(Party.f_1d8) - 1; si >= 0; si-- {
		t := uint8(0)
		if isCharaAlive(si) {
			t = TIL_38
		}
		Fighters._chtile[si] = t
		D_944A[si] = t
	}
	dspl_Stats()
	u4_puts("Resting...\n")
	D_95CC = int16(CurMode)
	CurMode = MOD_COM_CAMP
	u_delay(10, 0)
	if U4_RND1(7) == 0 {
		C_8A5A()
	} else if Party._moves/100 == uint32(Party.f_1e6) {
		u4_puts("No effect.\n")
	} else {
		Party.f_1e6 = uint16(Party._moves / 100)
		for si := int(Party.f_1d8) - 1; si >= 0; si-- {
			if isCharaAlive(si) {
				if Party.chara[si]._stat == 'S' {
					Party.chara[si]._stat = 'G'
				}
				HP_inc(si, U4_RND1(119))
				HP_inc(si, 99)
				Party.chara[si]._MP = 99
			}
		}
		MP_recover()
		u4_puts("Players Healed!\n")
	}
	dspl_Stats()
	CurMode = uint16(D_95CC)
	if CurMode == MOD_DUNGEON {
		D_9446 = 1
	}
}

// ---- U4_Q_N_V.C ----

func CMD_Quit() { /*C_6F29*/
	u4_puts("Quit & Save...\n")
	u4_putl(int32(Party._moves), 1, '0')
	u4_puts(" moves\n")
	if Party._loc != 0 {
		if Party._loc < 0x11 || Party._loc > 0x18 {
			u4_puts("Not Here!\n")
			return
		}
	}
	if Save("PARTY.SAV", sizeof(&Party), &Party) == -1 {
		exit(3)
	}
	if Party._loc == 0 {
		if Save("MONSTERS.SAV", sizeof(&D_8742._npc), &D_8742._npc) == -1 {
			exit(3)
		}
	}
	if Party._loc < 0x11 || Party._loc > 0x18 {
		return
	}
	if Save("MONSTERS.SAV", sizeof(&D_8742._npc), &D_8742._npc) == -1 {
		exit(3)
	}
	if Save("DNGMAP.SAV", 0x200, &D_8742._map) == -1 {
		exit(3)
	}
}

const D_21F4 = ", You must lead!\n"

// exchange two party members
func C_6FF9(a, b *tChara) { *a, *b = *b, *a }

func CMD_NewOrder() { /*C_7034*/
	u4_puts("New Order!\n")
	loc_C := AskChara("Exchange #\x12\x12\b")
	if loc_C < 0 {
		return
	}
	if loc_C == 0 {
		u4_puts(cstr(Party.chara[0]._name[:]))
		u4_puts(D_21F4)
		return
	}
	loc_D := AskChara("    with #\x12\x12\b")
	if loc_D < 0 {
		return
	}
	if loc_D == 0 {
		u4_puts(cstr(Party.chara[0]._name[:]))
		u4_puts(D_21F4)
		return
	}
	if loc_C == loc_D {
		w_What()
		return
	}
	C_6FF9(&Party.chara[loc_C], &Party.chara[loc_D])
	Gra_13()
	dspl_Stats()
}

func CMD_Volume() { /*C_70AD*/
	u4_puts("Volume O")
	SoundFlag ^= 1
	if SoundFlag != 0 {
		u4_puts("n!\n")
	} else {
		u4_puts("ff!\n")
	}
}

// ---- U4_EVT.C ----

func C_919A() {
	for si := int(Party.f_1d8) - 1; si >= 0; si-- {
		if isCharaConscious(si) && U4_RND1(3) == 0 {
			Party.chara[si]._stat = 'S'
			C_09D9(si)
			dspl_Stats()
		}
	}
}

func C_91D1() {
	for si := int(Party.f_1d8) - 1; si >= 0; si-- {
		if Party.chara[si]._stat == 'G' && U4_RND1(7) == 0 {
			Party.chara[si]._stat = 'P'
			C_09D9(si)
			dspl_Stats()
		}
	}
}

// randomEvent
func C_9209() {
	if CurMode == MOD_OUTDOORS && tile_cur == TIL_17 && U4_RND1(7) == 0 {
		u4_puts("\nBridge Trolls!\n")
		D_96F8 = TIL_A4
		D_9772 = Party._x
		D_9140 = Party._y
		D_946C = TIL_17
		C_7DBC()
		return
	}
	if CurMode <= MOD_BUILDING && Party.f_1dc != 0 {
		return
	}
	if CurMode >= MOD_COMBAT {
		a := int(activeChara)
		if Fighters._chtile[a] == 0 {
			return
		}
		switch *Combat_MAP(int(Combat._charaY[a]), int(Combat._charaX[a])) {
		case TIL_03, TIL_44:
			if Party.chara[a]._stat == 'G' {
				Party.chara[a]._stat = 'P'
				C_09D9(a)
				dspl_Stats()
			}
		case TIL_47:
			if isCharaConscious(a) {
				Party.chara[a]._stat = 'S'
				Fighters._chtile[a] = TIL_38
				C_09D9(a)
				dspl_Stats()
			}
		case TIL_46, TIL_4C:
			C_09D9(a)
			hitChara(a, U4_RND3(32)+16)
		}
		return
	}
	if CurMode == MOD_DUNGEON {
		if tile_cur&0xf0 == 0xa0 {
			switch tile_cur & 3 {
			case 0:
				C_91D1()
			case 2:
				C_1584()
			case 3:
				C_919A()
			}
		} else if tile_cur&0xf0 == 0x80 {
			if tile_cur == 0x80 {
				u4_puts("\nWinds!\n")
				Party.f_1dc = 0
			} else {
				if tile_cur < 0x88 {
					u4_puts("\nFalling Rocks!\n")
				} else {
					u4_puts("\nPit!\n")
				}
				C_1584()
			}
		}
		return
	}
	switch tile_cur {
	case TIL_03, TIL_44:
		C_91D1()
	case TIL_46, TIL_4C:
		C_1584()
	case TIL_47:
		C_919A()
	}
}

// ---- U4_MIX.C ----

const (
	RG_ash        = 0x80 >> 0
	RG_ginseng    = 0x80 >> 1
	RG_garlic     = 0x80 >> 2
	RG_spiderSilk = 0x80 >> 3
	RG_bloodMoss  = 0x80 >> 4
	RG_blackPearl = 0x80 >> 5
	RG_nightshade = 0x80 >> 6
	RG_mandrake   = 0x80 >> 7
)

var D_277E = [26]uint8{
	/*A*/ RG_ginseng | RG_garlic,
	/*B*/ RG_spiderSilk | RG_bloodMoss,
	/*C*/ RG_ginseng | RG_garlic,
	/*D*/ RG_ash | RG_garlic | RG_blackPearl,
	/*E*/ RG_ash | RG_spiderSilk | RG_blackPearl,
	/*F*/ RG_ash | RG_blackPearl,
	/*G*/ RG_ash | RG_blackPearl | RG_mandrake,
	/*H*/ RG_ginseng | RG_spiderSilk,
	/*I*/ RG_blackPearl | RG_mandrake,
	/*J*/ RG_blackPearl | RG_nightshade | RG_mandrake,
	/*K*/ RG_blackPearl | RG_nightshade,
	/*L*/ RG_ash,
	/*M*/ RG_ash | RG_blackPearl,
	/*N*/ RG_ash | RG_garlic | RG_mandrake,
	/*O*/ RG_ash | RG_bloodMoss,
	/*P*/ RG_ash | RG_ginseng | RG_garlic,
	/*Q*/ RG_ash | RG_ginseng | RG_bloodMoss,
	/*R*/ RG_ash | RG_ginseng | RG_garlic | RG_spiderSilk | RG_bloodMoss | RG_mandrake,
	/*S*/ RG_ginseng | RG_spiderSilk,
	/*T*/ RG_ash | RG_bloodMoss | RG_mandrake,
	/*U*/ RG_ash | RG_garlic,
	/*V*/ RG_nightshade | RG_mandrake,
	/*W*/ RG_ash | RG_bloodMoss,
	/*X*/ RG_ash | RG_spiderSilk | RG_bloodMoss,
	/*Y*/ RG_spiderSilk | RG_bloodMoss,
	/*Z*/ RG_spiderSilk | RG_bloodMoss,
}

func CMD_Mix() { /*C_8C08*/
	var loc_A [8]uint16
	loc_D := 0
	for loc_D == 0 {
		Gra_13()
		C_4649()
		C_4C42(0)
		txt_Y = 23
		txt_X = 24
		u4_puts("Mix Reagents\n")
		loc_C := AskLetter("For Spell:\x12\x12\b", 'A', 'Z')
		if loc_C < 0 {
			break
		}
		loc_C -= 'A'
		u4_puts(D_1E98[101+loc_C])
		Gra_CR()
		Gra_13()
		C_4649()
		for si := 7; si >= 0; si-- {
			loc_A[si] = Party._reagents[si] & 0xff // kept in a byte
		}
		loc_E := uint8(0) // char; the compiler compares it as a byte
		for {
			C_4BC7(0)
			txt_Y = 23
			txt_X = 24
			loc_B := AskLetter("Reagent:\x12\x12\b", 'A', 'H')
			if loc_B == -2 {
				Party._reagents = loc_A
				loc_D = 1
				break
			}
			if loc_B == -1 {
				if loc_E == 0 {
					u4_puts("\nNothing mixed!\n")
				} else {
					u4_puts("\nYou mix the Reagents, and...\n")
					if D_277E[loc_C] == loc_E {
						u4_puts("Success!\n\n")
						Party._mixtures[loc_C]++
						if Party._mixtures[loc_C] > 99 {
							Party._mixtures[loc_C] = 99
						}
					} else {
						u4_puts("It Fizzles!\n\n")
					}
				}
				break
			}
			loc_B -= 'A'
			if Party._reagents[loc_B] == 0 {
				w_NoneLeft()
				break
			}
			loc_E |= 0x80 >> loc_B
			Party._reagents[loc_B]--
		}
	}
	Gra_13()
	C_4649()
	dspl_Stats()
}

// ---- U4_LOCAT.C ----

func CMD_Locate() { /*C_755E*/
	u4_puts("Locate position with ")
	if Party._sextants == 0 {
		w_What()
		return
	}
	u4_puts("sextant\n Latitude: ")
	u4_putc(int(Party._y>>4) + 'A')
	u4_putc('\'')
	u4_putc(int(Party._y&0xf) + 'A')
	u4_puts("\"\nLongitude: ")
	u4_putc(int(Party._x>>4) + 'A')
	u4_putc('\'')
	u4_putc(int(Party._x&0xf) + 'A')
	u4_putc('"')
	Gra_CR()
}

var D_2334 = [24]uint8{
	// weapons class masks
	0xFF, 0xFF, 0xFF, 0xFF, 0x7F, 0x6F, 0x6F, 0x7E, 0x7E, 0xFF, 0x2C, 0x0C, 0x2E, 0x5E, 0xD0, 0xFF,
	// D_2344: armor class masks
	0xFF, 0xFF, 0x7F, 0x2C, 0x2C, 0x24, 0x04, 0xFF,
}

// "may not use"
func C_75DC(bp08, bp06, bp04 int) {
	u4_puts("\nA ")
	u4_puts(D_1E98[77+int(Party.chara[bp08]._class)])
	u4_puts(" may NOT use ")
	if bp04 != 0 {
		u4_putc('a')
	}
	Gra_CR()
	u4_puts(D_1E98[37+bp06])
	Gra_CR()
	sound(1)
}

func CMD_Ready() { /*C_7631*/
	var loc_A int
	if CurMode >= MOD_COMBAT {
		u4_puts("Ready Weapon!\n")
		loc_A = int(activeChara)
	} else {
		u4_puts("Ready a weapon\n")
		if loc_A = AskChara("for:\x12\x12\b"); loc_A < 0 {
			return
		}
	}
	Gra_13()
	C_4832(0)
	txt_Y = 23
	txt_X = 24
	loc_B := AskLetter("Weapon:\x12\x12\b", 'A', 'P')
	Gra_13()
	C_4649()
	dspl_Stats()
	if CurMode >= MOD_COMBAT {
		Gra_11(int(activeChara))
	}
	if loc_B < 0 {
		return
	}
	loc_B -= 'A'
	if loc_B != 0 && Party._weapons[loc_B] == 0 {
		w_NoneLeft()
		return
	}
	if loc_B != 0 {
		if (0x80>>Party.chara[loc_A]._class)&D_2334[loc_B] == 0 {
			C_75DC(loc_A, loc_B, 1)
			return
		}
	}
	if loc_C := Party.chara[loc_A]._weapon; loc_C != 0 {
		Party._weapons[loc_C]++
		if Party._weapons[loc_C] > 99 {
			Party._weapons[loc_C] = 99
		}
	}
	if loc_B != 0 {
		Party._weapons[loc_B]--
	}
	Party.chara[loc_A]._weapon = uint16(loc_B)
	u4_puts(D_1E98[37+loc_B])
	Gra_CR()
}

func CMD_Wear() { /*C_7732*/
	u4_puts("Wear Armour\n")
	loc_A := AskChara("for:\x12\x12\b")
	if loc_A < 0 {
		return
	}
	Gra_13()
	C_48F8(0)
	txt_Y = 23
	txt_X = 24
	loc_B := AskLetter("Armour:\x12\x12\b", 'A', 'H')
	Gra_13()
	C_4649()
	dspl_Stats()
	if CurMode >= MOD_COMBAT {
		Gra_11(int(activeChara))
	}
	if loc_B < 0 {
		return
	}
	loc_B -= 'A'
	if loc_B != 0 && Party._armors[loc_B] == 0 {
		w_NoneLeft()
		return
	}
	if loc_B != 0 {
		if (0x80>>Party.chara[loc_A]._class)&D_2334[loc_B+0x10] == 0 {
			C_75DC(loc_A, loc_B+0x10, 0)
			return
		}
	}
	if loc_C := Party.chara[loc_A]._armor; loc_C != 0 {
		Party._armors[loc_C]++
		if Party._armors[loc_C] > 99 {
			Party._armors[loc_C] = 99
		}
	}
	if loc_B != 0 {
		Party._armors[loc_B]--
	}
	Party.chara[loc_A]._armor = uint16(loc_B)
	u4_puts(D_1E98[53+loc_B])
	Gra_CR()
}
