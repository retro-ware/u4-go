package u4

// Port of U4_COMBC.C (signed char): combat tables, loading CON files, starting and ending a combat.

// D_23D2, D_2406 and D_242A are contiguous in the original data segment: C_7C25 deliberately
// (bug) indexes past D_23D2 into D_2406 for non-monster livings, so keep one backing array.
var d_combat_tab = [52 + 36 + 38]uint8{
	// D_23D2: base for enemy HP/XP/attack
	// 80~8f
	0xFF, 0xFF, 0x40, 0x60, 0x80, 0x60, 0xFF, 0xFF,
	// 90~AF
	0x30, 0x30, 0x40, 0x50, 0x30, 0x60, 0x30, 0xC0, 0xFF, 0x30, 0xF0, 0x80, 0x50, 0x30, 0x50, 0x30,
	0x70, 0x40, 0x80, 0x40, 0xB0, 0xC0, 0x60, 0xF0, 0x70, 0xD0, 0xE0, 0xFF,
	// 20~2f [buggy]
	0x70, 0x30, 0x60, 0x40, 0x60, 0x80, 0x90, 0x30,
	// 50~5f [buggy]
	0x80, 0x30, 0x30, 0x30, 0x20, 0x20, 0x80, 0xFF,
	// D_2406: enemy tile "partners"
	TIL_C8, TIL_C8, TIL_8A, TIL_88, TIL_86, TIL_84, TIL_8C, TIL_8E,
	TIL_C4, TIL_E9, TIL_90, TIL_E4, TIL_A0, TIL_D0, TIL_A8, TIL_AC, TIL_B0, TIL_90, TIL_BC, TIL_9C,
	TIL_A4, TIL_E0, TIL_C8, TIL_90, TIL_F0, TIL_B8, TIL_EC, TIL_BC, TIL_F0, TIL_F0, TIL_F4, TIL_B8,
	TIL_FC, TIL_F8, TIL_FC, TIL_FC,
	// D_242A: enemy number
	1, 1, 12, 4, 4, 8, 1, 1,
	12, 12, 6, 4, 15, 6, 15, 1, 1, 15, 4, 8, 10, 12, 10, 12,
	6, 8, 6, 12, 6, 4, 8, 4, 6, 4, 4, 1,
	0, 0, // padding
}

var (
	D_23D2 = d_combat_tab[0:52]
	D_2406 = d_combat_tab[52:88]
	D_242A = d_combat_tab[88:]
)

// weapons damage
var D_2450 = [16]uint8{0x08, 0x10, 0x18, 0x20, 0x28, 0x30, 0x40, 0x28, 0x38, 0x40, 0x60, 0x60, 0x80, 0x50, 0xA0, 0xFF}

// armors protection
var D_2460 = [8]uint8{0x60, 0x80, 0x90, 0xA0, 0xB0, 0xC0, 0xD0, 0xF8}

// range weapons flag
var D_2468 = [16]uint8{0x00, 0x00, 0x00, 0xFF, 0x00, 0x00, 0x00, 0xFF, 0xFF, 0xFF, 0x00, 0xFF, 0x00, 0xFF, 0xFF, 0x00}

// Enemy index. Original bug kept: non-monsters index (tile&0x1f)+0x24, past D_23D2 into D_2406.
func C_7C25(bp04 uint8) int {
	if bp04 >= TIL_80 {
		bp04 &= 0x7f
		if (bp04 >> 1) < 8 {
			return int(bp04 >> 1)
		}
		return int(bp04>>2) + 4
	}
	// not monster
	return int(bp04&0x1f) + 0x24
}

var D_2512 = [14]string{
	"GRASS.CON", "BRIDGE.CON", "BRICK.CON", "DUNGEON.CON", "HILL.CON", "FOREST.CON", "BRUSH.CON",
	"MARSH.CON", "SHIPSEA.CON", "SHIPSHOR.CON", "SHORE.CON", "SHIPSHIP.CON", "SHORSHIP.CON", "CAMP.CON",
}

// load CON file
func C_7C65() {
	var si int
	if Party._tile <= TIL_13 || int(tile_cur)&^3 == TIL_10 {
		if D_96F8 == TIL_80 {
			D_9452 = TIL_C8
			si = 11
		} else if D_946C < TIL_03 {
			si = 8
		} else {
			si = 9
		}
	} else {
		if D_96F8 == TIL_80 {
			D_9452 = TIL_C8
			si = 12
		} else if D_946C < TIL_03 {
			si = 10
		} else {
			switch tile_cur {
			case TIL_03:
				si = 7
			case TIL_05:
				si = 6
			case TIL_06:
				si = 5
			case TIL_07:
				si = 4
			case TIL_09:
				si = 3
			case TIL_3E:
				si = 2
			case TIL_17, TIL_19, TIL_1A, TIL_3F:
				si = 1
			default:
				si = 0
			}
		}
	}
	LoadRec(D_2512[si], &Combat)
}

var D_256E = [16]int{0, 1, 2, 3, 4, 0, 0, 0, 0, 0, 0, 0, 5, 0, 6, 0}

var D_258E = [7]string{"DNG0.CON", "DNG1.CON", "DNG2.CON", "DNG3.CON", "DNG4.CON", "DNG5.CON", "DNG6.CON"}

// load dungeon CON file
func C_7D50() {
	LoadRec(D_258E[D_256E[tile_cur>>4]], &Combat)
	C_7FD7()
	D_9446 = 1 // force 3d render
}

// load inn CON file
func C_7D92() {
	LoadRec("INN.CON", &Combat)
	C_7FD7()
}

// start a combat bis ?
func C_7DBC() {
	u4_puts("\n\n**** Combat ****\n\n")
	D_9452 = D_96F8
	for si := 15; si >= 0; si-- {
		Fighters._tile[si] = 0
		Fighters._x[si] = 0
		Fighters._y[si] = 0
		Fighters._sleeping[si] = 0
	}
	for si := 7; si >= 0; si-- {
		Fighters._chtile[si] = 0
	}
	C_7C65()
	C_7FD7()
}

// start a combat ?
func C_7DFE(si int) {
	D_96F8 = D_8742._npc._tile[si]
	D_9772 = D_8742._npc._x[si]
	D_9140 = D_8742._npc._y[si]
	if Party._loc == 0x00 {
		D_946C = *D_8742.x32x32(int(D_9140)-(int(D_95A5.y)<<4), int(D_9772)-(int(D_95A5.x)<<4))
	} else {
		D_946C = *D_8742.x32x32(int(D_9140), int(D_9772))
	}
	D_8742._npc._var[si] = 0
	D_8742._npc._gtile[si] = 0
	D_8742._npc._tile[si] = 0
	C_7DBC()
}

// init fighters info?
func C_7E7E() {
	var loc_A, loc_D, loc_B, loc_C int

	for loc_D = 15; loc_D >= 0; loc_D-- {
		Fighters._tile[loc_D] = 0
	}
	partners := func() int { return int(D_242A[C_7C25(uint8(loc_A))]) }
	if D_95CC == MOD_BUILDING && CurMode != MOD_COM_CAMP {
		if D_96F8 == TIL_50 {
			loc_A = int(Party.f_1d8)*2 - 1
		} else {
			loc_A = 0
		}
	} else {
		loc_A = int(D_9452)
		if D_9452 < TIL_80 {
			loc_A = (U4_RND4(partners()) + partners()) / 2
		} else if loc_A = U4_RND1(7); loc_A == 0 {
			loc_A = (U4_RND4(partners()) + partners()) / 2
		}
	}
	for {
		loc_D = loc_A
		if uint16(loc_A>>1) < Party.f_1d8 { // int vs unsigned compare
			break
		}
		loc_A = U4_RND4(int(Party.f_1d8) * 2)
	}
	for loc_B = loc_D; loc_B >= 0; loc_B-- {
		for {
			loc_D = U4_RND1(15)
			if Fighters._tile[loc_D] == 0 {
				break
			}
		}
		loc_E := D_9452
		// add some "partners"
		if D_9452 > TIL_80 && loc_B != 0 {
			if U4_RND1(31) == 0 {
				loc_E = D_2406[C_7C25(D_2406[C_7C25(D_9452)])]
			} else if U4_RND1(7) == 0 {
				loc_E = D_2406[C_7C25(D_9452)]
			}
		}
		Fighters._tile[loc_D] = loc_E
		Fighters._gtile[loc_D] = loc_E
		loc_C = int(D_23D2[C_7C25(loc_E)])
		Fighters._HP[loc_D] = uint8((loc_C >> 1) | U4_RND4(loc_C))
	}
	for loc_D = int(Party.f_1d8) - 1; loc_D >= 0; loc_D-- {
		if !isCharaAlive(loc_D) {
			Fighters._chtile[loc_D] = 0
		} else {
			Fighters._chtile[loc_D] = uint8(C_0ACF(loc_D))
			D_944A[loc_D] = Fighters._chtile[loc_D]
		}
	}
}

func C_7FD7() {
	if CurMode != MOD_COM_CAMP {
		D_95CC = int16(CurMode)
		CurMode = MOD_COMBAT
	}
	C_7E7E()
	COM_main()
	CurMode = uint16(D_95CC)
}

// enter dungeon room?
func C_7FFD() {
	// dungeon room info (0x100 bytes):
	//  000 secret tiles info[16]; 010 enemy tiles[16]; 020 x[16]; 030 y[16];
	//  charas positions x,y[8] each: 040/048 dir 0, 050/058 dir 1, 060/068 dir 2, 070/078 dir 3;
	//  080 room map[0x80]
	var loc_A [0x100]uint8
	var loc_B, loc_C int

	ofs := uint16(8 * 8 * 8) // past the level maps
	if Party._loc == 0x18 {
		ofs += uint16(int(Party._z>>1)*0x1000) + uint16(tile_cur&0xf)*0x100
	} else {
		ofs += uint16(tile_cur&0xf) * 0x100
	}
	copy(loc_A[:], at(File_DNG, int(ofs)))
	D_95CC = MOD_DUNGEON
	CurMode = MOD_COM_ROOM
	copy(D_95B2[:], loc_A[:16])
	D_943E = -1
	if tile_cur&0xf == 0xf && Party._loc < 0x18 {
		u4_puts("\nThe Altar Room of ")
		if Party._x == 3 {
			u4_puts("Love\n")
			D_943E = 1
		} else if Party._x <= 2 {
			u4_puts("Truth\n")
			D_943E = 0
		} else {
			u4_puts("Courage\n")
			D_943E = 2
		}
	}
	var px, py int
	switch ((int(Party._dir) - 1) ^ 2) & 3 {
	case 2:
		px, py = 0x60, 0x68
	case 0:
		px, py = 0x40, 0x48
	case 3:
		px, py = 0x70, 0x78
	case 1:
		px, py = 0x50, 0x58
	}
	for loc_B = 7; loc_B >= 0; loc_B-- {
		Combat._charaX[loc_B] = loc_A[px+loc_B]
		Combat._charaY[loc_B] = loc_A[py+loc_B]
	}
	for loc_B = int(Party.f_1d8) - 1; loc_B >= 0; loc_B-- {
		if isCharaAlive(loc_B) {
			Fighters._chtile[loc_B] = uint8(C_0ACF(loc_B))
			D_944A[loc_B] = Fighters._chtile[loc_B]
		} else {
			Fighters._chtile[loc_B] = 0
		}
	}
	for loc_B = 15; loc_B >= 0; loc_B-- {
		/*C_817D*/
		// bug in the original: "if(loc_A._010)" tests the array address, always true,
		// so empty slots (tile 0) are set up as well.
		Combat._npcX[loc_B] = loc_A[0x20+loc_B]
		Combat._npcY[loc_B] = loc_A[0x30+loc_B]
		Fighters._tile[loc_B] = loc_A[0x10+loc_B]
		Fighters._gtile[loc_B] = loc_A[0x10+loc_B]
		loc_C = int(D_23D2[C_7C25(loc_A[0x10+loc_B])])
		Fighters._HP[loc_B] = uint8((loc_C >> 1) | U4_RND4(loc_C))
		if Fighters._tile[loc_B] == TIL_AC {
			Fighters._gtile[loc_B] = TIL_3C
		}
	}
	for loc_B = 10; loc_B >= 0; loc_B-- {
		for loc_C = 10; loc_C >= 0; loc_C-- {
			Combat._map[loc_B*11+loc_C] = loc_A[0x80+loc_B*11+loc_C]
		}
	}
	COM_main()
	CurMode = uint16(D_95CC)
}

func C_8214() {
	if D_95CC == MOD_DUNGEON {
		C_AE41()
	}
	u_kbflush()
}

func C_8228() int {
	// bug in the original: "si >= (cond)?8:0" parses as "(si >= cond) ? 8 : 0", so the loop
	// runs while si >= (outdoors ? 1 : 0)
	lim := 0
	if D_95CC == MOD_OUTDOORS {
		lim = 1
	}
	si := 31
	for ; si >= lim; si-- {
		if D_8742._npc._tile[si] == 0 {
			break
		}
	}
	if si == -1 || (si == 7 && D_95CC == MOD_OUTDOORS) {
		si = U4_RND3(24) + 8
	}
	return si
}

func C_8283() {
	si := int(D_96F8)
	if si == TIL_80 {
		// pirate ship becomes player ship
		si = C_8228()
		D_8742._npc._x[si] = D_9772
		D_8742._npc._y[si] = D_9140
		D_8742._npc._tile[si] = TIL_10
		D_8742._npc._gtile[si] = TIL_10
	} else if D_95CC == MOD_DUNGEON {
		// monster leaves chest?
		if si != TIL_94 && si != TIL_B4 && si != TIL_A0 {
			if *D_8742.x8x8x8(int(Party._z), int(Party._y), int(Party._x)) == 0 {
				*D_8742.x8x8x8(int(Party._z), int(Party._y), int(Party._x)) = 0x40
			}
		}
	} else if D_95CC != MOD_COM_CAMP || tile_cur < TIL_10 || tile_cur > TIL_13 {
		// monster leaves chest?
		if si >= TIL_90 || (si >= TIL_20 && si <= TIL_2E) || (si >= TIL_50 && si <= TIL_5E) {
			if si != TIL_94 && si != TIL_A0 && si != TIL_B8 && si != TIL_B4 && si != TIL_DC {
				if C_2999(D_946C) {
					si = C_8228()
					D_8742._npc._x[si] = D_9772
					D_8742._npc._y[si] = D_9140
					D_8742._npc._tile[si] = TIL_3C
					D_8742._npc._gtile[si] = TIL_3C
				}
			}
		}
	}
	C_8214()
}

// altar rooms/dungeons connection
//
//	0x11 - deceit   - truth
//	0x12 - despise  - love
//	0x13 - destard  - courage
//	0x14 - wrong    - truth love
//	0x15 - covetous - love courage
//	0x16 - shame    - truth courage
//	0x17 - hythloth - truth love courage
var D_261A = [3][4]uint8{
	{0x11, 0x16, 0x17, 0x14}, // truth
	{0x12, 0x14, 0x17, 0x15}, // love
	{0x13, 0x15, 0x17, 0x16}, // courage
}

func C_837A() {
	si := 7
	for ; ; si-- {
		if si < 0 || Fighters._chtile[si] != 0 {
			break
		}
	}

	if si == -1 || CurMode == MOD_COM_ROOM {
		if CurMode != MOD_COM_ROOM {
			if C_0AFE(D_96F8) != 0 {
				karma_inc(&Party._compa, 2)
				karma_inc(&Party._justi, 2)
			} else {
				/*C_83CE:*/
				u4_puts("Battle is lost!\n")
				karma_dec(&Party._valor, 2)
			}
			C_8214()
			return
		}
		/*C_83E6:*/
		u4_puts("Leave Room!\n")
		if D_96EE == 0 {
			if sc(D_96F4) < 0 {
				Party._dir = DIR_N
			} else {
				Party._dir = DIR_S
			}
		} else {
			if sc(D_96EE) < 0 {
				Party._dir = DIR_W
			} else {
				Party._dir = DIR_E
			}
		}
		if D_943E != -1 {
			Party._loc = uint16(D_261A[D_943E][(Party._dir-1)&3])
			Party.out_x = D_0844[Party._loc-1]
			Party.out_y = D_0864[Party._loc-1]
			u4_puts("into Dungeon\n")
			u4_puts(D_1E98[127+int(Party._loc)-1])
			File_DNG = nil
			C_3E94()
		}
		Gra_CR()
		u_kbflush()
		Party._x = uint8(DNG_X_p(int(Party._x), int(Party._dir)))
		Party._y = uint8(DNG_Y_p(int(Party._y), int(Party._dir)))
		return
	}
	/*C_849E:*/
	u4_puts("\nVictory!\n")
	if C_0AFE(D_96F8) == 0 {
		karma_inc(&Party._valor, U4_RND1(1))
	}
	if D_95CC != MOD_COM_CAMP {
		C_8283()
	}
}
