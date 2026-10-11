package u4

// Ports of U4_PEER.C (gem views) and U4_GET.C (chests, cannons, torches).

// ---- U4_PEER.C ----

var D_2F72 = [128]uint8{
	0x0C, 0x0B, 0x0A, 0x01, 0x01, 0x09, 0x02, 0x08, 0x07, 0x05, 0x05, 0x05, 0x05, 0x04, 0x04, 0x04,
	0x05, 0x05, 0x05, 0x05, 0x05, 0x05, 0x05, 0x04, 0x05, 0x04, 0x04, 0x05, 0x05, 0x06, 0x06, 0x05,
	0x06, 0x06, 0x06, 0x06, 0x06, 0x06, 0x06, 0x06, 0x06, 0x06, 0x06, 0x06, 0x06, 0x06, 0x06, 0x06,
	0x07, 0x07, 0x07, 0x07, 0x07, 0x07, 0x07, 0x05, 0x05, 0x07, 0x06, 0x06, 0x05, 0x00, 0x03, 0x04,
	0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x07, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03,
	0x05, 0x05, 0x05, 0x05, 0x05, 0x05, 0x05, 0x05, 0x05, 0x05, 0x05, 0x05, 0x05, 0x05, 0x05, 0x05,
	0x04, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04,
	0x04, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04, 0x06, 0x00, 0x07,
}

var D_2FF2 = [16]uint8{0x20, 0x06, 0x05, 0x04, 0x24, 0x20, 0x20, 0x0F, 0x54, 0x46, 0x5E, 0x00, 0x0E, 0x0E, 0x02, 0x03}

func MK_COORD(a, b int) int { return (a+8)*4 + b }

var (
	D_8CFC [600]uint8
	D_8F54 [22][22]uint8
)

// display the town map (bp04: with the player's position)
func C_B9EF(bp04 int) {
	loc_B := CurMode
	CurMode = MOD_VISION
	Gra_10()
	for loc_A := 0; loc_A < 32; loc_A++ {
		for loc_C := 0; loc_C < 32; loc_C++ {
			PT := func(x, y, c int) { Gra_dot(MK_COORD(loc_A, x), MK_COORD(loc_C, y), c) }
			pts := func(c int, xy ...int) {
				for i := 0; i < len(xy); i += 2 {
					PT(xy[i], xy[i+1], c)
				}
			}
			t := *D_8742.x32x32(loc_A, loc_C)
			k := uint8(5) // C_BCA0
			if t < 0x80 {
				k = D_2F72[t]
			}
			switch k {
			case 1: // grass
				pts(1, 0, 1, 1, 3, 2, 1, 3, 3)
			case 2: // forest
				pts(1, 0, 3, 1, 1, 2, 3, 3, 1, 0, 1, 1, 3, 2, 1, 3, 3)
			case 3: // lava
				pts(2, 0, 0, 1, 0, 2, 0, 3, 0, 0, 2, 1, 2, 2, 2, 3, 2)
			case 4: // castle/bridge
				pts(3, 0, 0, 0, 1, 0, 2, 0, 3, 3, 0, 3, 1, 3, 2, 3, 3)
			case 5: // dungeon/town
				pts(3, 1, 1, 2, 1, 1, 2, 2, 2)
			case 6: // shrine
				pts(3, 1, 0, 2, 0, 1, 3, 2, 3, 0, 0, 0, 1, 0, 2, 0, 3, 3, 0, 3, 1, 3, 2, 3, 3)
			case 7: // mountain
				pts(3, 1, 1, 2, 1, 1, 2, 2, 2, 1, 0, 2, 0, 1, 3, 2, 3, 0, 0, 0, 1, 0, 2, 0, 3, 3, 0, 3, 1, 3, 2, 3, 3)
			case 8: // hill
				pts(3, 0, 0, 1, 0, 0, 1, 1, 1, 2, 2, 3, 2, 2, 3, 3, 3)
			case 9: // bush
				pts(1, 0, 1, 1, 1, 2, 1, 0, 3, 2, 3, 3, 3)
			case 10: // shallow water
				pts(2, 0, 0, 2, 0, 1, 2, 3, 2)
			case 11: // water
				pts(2, 0, 0, 2, 2)
			case 12: // deep water
				PT(2*(loc_C&1), 2, 3)
			}
		}
	}
	u_kbflush()
	y, x := int(D_959C.y), int(D_959C.x)
	for loc_C := 0; u_kbhit() == 0; loc_C++ {
		if bp04 != 0 {
			for r := 0; r < 4; r++ {
				Gra_dot_XOR(MK_COORD(y, r), MK_COORD(x, loc_C&3), 3)
			}
		}
	}
	u_kbread()
	CurMode = loc_B
}

// display the dungeon level map
func C_C23B() {
	loc_E := CurMode
	CurMode = MOD_VISION
	Gra_10()
	D_8F54 = [22][22]uint8{}
	txt_Y, txt_X = -1, -1
	loc_C := 0
	txt_Y = 11
	D_8CFC[loc_C] = 11
	loc_C++
	txt_X = 11
	D_8CFC[loc_C] = 11
	loc_C++
	D_8CFC[loc_C] = Party._y
	loc_C++
	D_8CFC[loc_C] = Party._x
	loc_C++
	D_8F54[10][10] = 1
	Gra_putchar(0x01)
	for {
		loc_C--
		loc_F := int(D_8CFC[loc_C])
		loc_C--
		loc_A := int(D_8CFC[loc_C])
		loc_C--
		txt_X = int16(D_8CFC[loc_C]) - 1
		loc_C--
		txt_Y = int16(D_8CFC[loc_C]) - 1
		for loc_B := -1; loc_B < 2; txt_X, txt_Y, loc_B = txt_X+1, txt_Y-3, loc_B+1 {
			for loc_D := -1; loc_D < 2; txt_Y, loc_D = txt_Y+1, loc_D+1 {
				if loc_B|loc_D == 0 {
					continue
				}
				bp_0e := (loc_B + loc_F) & 7
				bp_10 := (loc_D + loc_A) & 7
				if txt_Y < 1 || txt_Y > 22 || txt_X < 1 || txt_X > 22 || D_8F54[txt_Y-1][txt_X-1] != 0 {
					continue
				}
				t := dngTile(bp_10, bp_0e)
				Gra_putchar(D_2FF2[t>>4])
				if t&0xf0 != 0xf0 {
					D_8CFC[loc_C] = uint8(txt_Y)
					D_8CFC[loc_C+1] = uint8(txt_X)
					D_8CFC[loc_C+2] = uint8(bp_10)
					D_8CFC[loc_C+3] = uint8(bp_0e)
					loc_C += 4
					if loc_C >= 600 {
						loc_C = 0
					}
					D_8F54[txt_Y-1][txt_X-1] = 1
				}
			}
		}
		if loc_C == 0 {
			break
		}
	}
	txt_Y = 23
	txt_X = 24
	u_kbflush()
	u_kbread()
	CurMode = loc_E
	D_9446 = 1
}

func C_C403() {
	if Party._loc >= 0x11 {
		C_C23B()
	} else {
		C_B9EF(1)
	}
}

func CMD_Peer() { /*C_C41D*/
	u4_puts("Peer at ")
	if Party._gems == 0 {
		w_What()
		return
	}
	u4_puts("a Gem!\n")
	Party._gems--
	C_C403()
}

// ---- U4_GET.C ----

func C_70CE(bp04 int) {
	C_09D9(bp04)
	hitChara(bp04, U4_RND3(30))
}

func C_70F1() {
	u4_puts("The Chest Holds:\n")
	si := U4_RND3(80) + U4_RND1(7) + 10
	u4_putl(int32(si), 1, ' ')
	u4_puts(" Gold\n")
	if int(Party._gold)+si <= 9999 {
		Party._gold += uint16(si)
	} else {
		Party._gold = 9999
	}
}

// trapped chest ?
func C_7150(bp04 int) {
	si := U4_RND1(3)
	if si&1 == 0 {
		// note: si's bit 0 is always reset here, so the Sleep and Poison traps never occur.
		si &= int(u_rand_a())
		switch si {
		case 0:
			u4_puts("Acid")
		case 1:
			u4_puts("Sleep")
		case 2:
			u4_puts("Poison")
		default:
			u4_puts("Bomb")
		}
		u4_puts(" Trap!\n")
		if bp04 == -1 || U4_RND3(100) <= int(Party.chara[bp04]._dex)+25 {
			u4_puts("Evaded!\n")
			sound(8)
		} else if si == 0 { // Acid
			C_70CE(bp04)
		} else if si == 1 { // Sleep
			Party.chara[bp04]._stat = 'S'
			if CurMode >= MOD_COMBAT {
				Fighters._chtile[bp04] = TIL_38
			}
			C_09D9(bp04)
		} else if si == 2 { // Poison
			Party.chara[bp04]._stat = 'P'
			dspl_Stats()
			C_09D9(bp04)
		} else { // Bomb
			C_1584()
		}
	}
	C_70F1()
}

func C_722F(bp04 int) {
	if CurMode == MOD_DUNGEON {
		if tile_cur != 0x40 {
			w_NotHere()
			return
		}
		*D_8742.x8x8x8(int(Party._z), int(Party._y), int(Party._x)) = 0
	} else {
		if tile_cur != TIL_3C {
			w_NotHere()
			return
		}
		if Party._loc != 0 && *D_8742.x32x32(int(Party._y), int(Party._x)) == TIL_3C {
			*D_8742.x32x32(int(Party._y), int(Party._x)) = TIL_3E
			karma_dec(&Party._hones, 1)
			karma_dec(&Party._justi, 1)
			karma_dec(&Party._honor, 1)
		} else {
			si := 0x1f
			for ; si >= 0; si-- {
				if D_8742._npc._tile[si] == TIL_3C && Party._x == D_8742._npc._x[si] && Party._y == D_8742._npc._y[si] {
					break
				}
			}
			if si >= 0 {
				D_8742._npc._gtile[si] = 0
				D_8742._npc._tile[si] = 0
			}
		}
	}
	C_7150(bp04)
}

func CMD_Get() { /*C_72EC*/
	if Party.f_1dc != 0 && Party._tile == TIL_18 {
		u4_puts("Get chest\n")
		w_DriftOnly()
		return
	}
	u4_puts("Get Chest!\n")
	si := AskChara("Who opens?\x12\x12\b")
	if si < 0 {
		return
	}
	if !isCharaConscious(si) {
		w_Disabled()
		return
	}
	C_722F(si)
}

func C_7337() {
	a := int(activeChara)
	p := Combat_MAP(int(Combat._charaY[a]), int(Combat._charaX[a]))
	if *p != TIL_3C {
		w_NotHere()
		return
	}
	*p = TIL_16
	if t := D_8742.x8x8x8(int(Party._z), int(Party._y), int(Party._x)); *t == 0x40 {
		*t = 0
	}
	C_7150(a)
}

// (G)et chest [combat]
func C_73B8() {
	u4_puts("Get Chest!\n")
	C_7337()
}

// ---- FIRE ----

const D_22C8 = "Broadsides Only!\n"

func CMD_Fire() { /*C_73C9*/
	var loc_C, loc_D int
	u4_puts("Fire ")
	if Party._tile < TIL_10 || Party._tile > TIL_13 {
		w_What()
		return
	}
	u4_puts("Cannon!\n")
	AskDir("Dir: ", &loc_C, &loc_D)
	if loc_C|loc_D == 0 {
		return
	}
	if loc_C != 0 {
		if Party._tile != TIL_11 && Party._tile != TIL_13 {
			u4_puts(D_22C8)
			sound(2)
			return
		}
	} else if Party._tile != TIL_10 && Party._tile != TIL_12 {
		u4_puts(D_22C8)
		sound(2)
		return
	}
	// -- FIRING --
	sound(3)
	loc_F := int(Party._x) + loc_C
	loc_B := int(Party._y) + loc_D
	hit_x = uint8(loc_C + 5)
	hit_y = uint8(loc_D + 5)
	hit_tile = TIL_4D
	loc_A := -1
	for loc_E := 1; ; loc_E++ {
		hitIt := loc_E > 3
		if !hitIt {
			loc_A = C_0A58(0x1f, uint8(loc_F), uint8(loc_B))
			hitIt = loc_A != -1 && D_8742._npc._tile[loc_A] != TIL_8C && D_8742._npc._tile[loc_A] != TIL_18
		}
		if hitIt {
			hit_tile = 0
			if loc_A == -1 {
				return
			}
			// -- HIT SOMETHING --
			hit_tile = TIL_4F
			C_36C7()
			hit_tile = 0
			sound(6)
			if loc_A < 8 && U4_RND1(3) != 0 {
				return
			}
			if D_8742._npc._tile[loc_A] == TIL_5E {
				return
			}
			break
		}
		C_36C7()
		hit_x += uint8(loc_C)
		loc_F += loc_C
		hit_y += uint8(loc_D)
		loc_B += loc_D
		loc_A = -1
	}
	D_8742._npc._gtile[loc_A] = 0
	D_8742._npc._tile[loc_A] = 0
}

// ---- IGNITE ----

func CMD_Ignite() { /*C_7525*/
	u4_puts("Ignite Torch!\n")
	if CurMode != MOD_DUNGEON {
		w_NotHere()
		return
	}
	if Party._torches == 0 {
		w_NoneLeft()
		return
	}
	Party._torches--
	Party.f_1dc += 100
	if CurMode == MOD_DUNGEON {
		C_AE41()
	}
}
