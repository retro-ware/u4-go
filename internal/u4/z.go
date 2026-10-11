package u4

// Port of U4_Z.C: the Ztats pages in the status area.

func C_45B5() {
	for ; txt_Y <= 8; txt_Y++ {
		txt_X = 24
		u4_puts("               ")
	}
}

// display centered string
func C_45D6(si string, bp04 int) {
	di := len(si)
	if di > 12 {
		di = 12
	}
	di -= di & 1
	txt_X = int16(b2i(di >= 8) + (12-di)/2 + bp04 + 25)
	if txt_Y == 0 {
		txt_X--
		u4_putc(0x10)
	}
	u4_puts(si)
	if txt_Y == 0 {
		u4_putc(0x11)
	}
}

func C_4649() {
	bp_02, bp_04 := txt_Y, txt_X
	txt_Y = 0
	txt_X = 24
	for {
		Gra_putchar('\r')
		txt_X++
		if txt_X >= 39 {
			break
		}
	}
	txt_Y, txt_X = bp_02, bp_04
}

func C_4689(bp04 int) {
	si := &Party.chara[bp04]
	txt_Y = 0
	C_45D6(cstr(si._name[:]), 0)
	txt_X = 24
	txt_Y = 1
	u4_putc(int(si.p_24))
	txt_Y = 1
	C_45D6(D_1E98[77+int(si._class)], 0)
	txt_X = 38
	u4_putc(int(si._stat))
	txt_X = 25
	txt_Y = 3
	u4_puts("MP:")
	u4_putl(int32(si._MP), 2, '0')
	txt_X = 32
	u4_puts("LV:")
	u4_putc(int(si._HP[1]/100) + '0')
	txt_X = 24
	txt_Y = 4
	u4_puts("STR:")
	u4_putl(int32(si._str), 2, '0')
	txt_X = 32
	u4_puts("HP:")
	u4_putl(int32(si._HP[0]), 4, '0')
	txt_X = 24
	txt_Y = 5
	u4_puts("DEX:")
	u4_putl(int32(si._dex), 2, '0')
	txt_X = 32
	u4_puts("HM:")
	u4_putl(int32(si._HP[1]), 4, '0')
	txt_X = 24
	txt_Y = 6
	u4_puts("INT:")
	u4_putl(int32(si._int), 2, '0')
	txt_X = 32
	u4_puts("EX:")
	u4_putl(int32(si._XP), 4, '0')
	txt_Y = 7
	txt_X = 24
	u4_puts("W:")
	u4_puts(D_1E98[37+int(si._weapon)])
	txt_Y = 8
	txt_X = 24
	u4_puts("A:")
	u4_puts(D_1E98[53+int(si._armor)])
}

func nextCol() {
	txt_X = (txt_X - 1) &^ 7
	txt_Y++
	if txt_Y == 9 {
		txt_Y = 1
		txt_X += 8
	}
}

func C_4832(int) {
	txt_Y = 0
	C_45D6("Weapons", 0)
	txt_Y = 1
	txt_X = 24
	u4_puts("A-Hands")
	txt_Y = 2
	txt_X = 24
	for si := 1; si < 0x10; si++ {
		if Party._weapons[si] != 0 {
			u4_putc(si + 'A')
			u4_putl(int32(Party._weapons[si]), 2, '-')
			u4_putc('-')
			u4_puts(D_1E98[61+si])
			nextCol()
		}
	}
	for txt_X < 40 {
		u4_puts("       ")
		nextCol()
	}
}

func C_48F8(int) {
	txt_Y = 0
	C_45D6("Armour", 0)
	txt_Y = 1
	txt_X = 24
	u4_puts("A  -No Armour")
	txt_Y = 2
	for si := 1; si < 8; si++ {
		if Party._armors[si] != 0 {
			txt_X = 24
			u4_putc(si + 'A')
			u4_putl(int32(Party._armors[si]), 2, '-')
			u4_putc('-')
			u4_puts(D_1E98[53+si])
			for txt_X != 39 {
				u4_putc(' ')
			}
			txt_Y++
		}
	}
	C_45B5()
}

func C_4987(int) {
	txt_Y = 0
	C_45D6("Equipment", -1)
	txt_X = 24
	txt_Y = 1
	u4_putl(int32(Party._torches), 2, ' ')
	u4_puts(" Torches")
	txt_X = 24
	txt_Y++
	u4_putl(int32(Party._gems), 2, ' ')
	u4_puts(" Gems")
	txt_X = 24
	txt_Y++
	u4_putl(int32(Party._keys), 2, ' ')
	u4_puts(" Keys")
	if Party._sextants != 0 {
		txt_X = 24
		txt_Y = 4
		u4_putl(int32(Party._sextants), 2, ' ')
		u4_puts(" Sextants")
	}
	txt_Y++
	C_45B5()
}

const D_199A = "BYRGOPWB"
const D_19A4 = "HCVJSHSH"

func C_4A3D(int) {
	it := Party.mItems
	txt_Y = 0
	C_45D6("Items", 0)
	txt_Y = 1
	if Party.mStones != 0 {
		txt_X = 24
		u4_puts("Stones:")
		for si := 0; si < 8; si++ {
			if TST_MSK8(Party.mStones, si) {
				u4_putc(int(D_199A[si]))
			}
		}
		txt_Y++
	}
	if Party.mRunes != 0 {
		txt_X = 24
		u4_puts("Runes:")
		for si := 0; si < 8; si++ {
			if TST_MSK8(Party.mRunes, si) {
				u4_putc(int(D_19A4[si]))
			}
		}
		txt_Y++
	}
	if TST_MSK(it, ST_BELL) || TST_MSK(it, ST_BOOK) || TST_MSK(it, ST_CANDLE) {
		txt_X = 24
		if TST_MSK(it, ST_BELL) {
			u4_puts("Bell ")
		}
		if TST_MSK(it, ST_BOOK) {
			u4_puts("Book ")
		}
		if TST_MSK(it, ST_CANDLE) {
			u4_puts("Candl")
			if !TST_MSK(it, ST_BOOK) || !TST_MSK(it, ST_BELL) {
				u4_putc('e')
			}
		}
		txt_Y++
	}
	if TST_MSK(it, ST_KEY_C) || TST_MSK(it, ST_KEY_L) || TST_MSK(it, ST_KEY_T) {
		txt_X = 24
		u4_puts("3 Part Key:")
		if TST_MSK(it, ST_KEY_T) {
			u4_putc('T')
		}
		if TST_MSK(it, ST_KEY_L) {
			u4_putc('L')
		}
		if TST_MSK(it, ST_KEY_C) {
			u4_putc('C')
		}
		txt_Y++
	}
	if TST_MSK(it, ST_HORN) {
		txt_X = 24
		u4_puts("Horn")
		txt_Y++
	}
	if TST_MSK(it, ST_WHEEL) {
		txt_X = 24
		u4_puts("Wheel")
		txt_Y++
	}
	if TST_MSK(it, ST_SKULL) {
		txt_X = 24
		u4_puts("Skull")
	}
}

func C_4BC7(int) {
	txt_Y = 0
	C_45D6("Reagents", 0)
	txt_Y = 1
	for si := 0; si < 8; si++ {
		if Party._reagents[si] != 0 {
			txt_X = 24
			u4_putc(si + 'A')
			u4_putl(int32(Party._reagents[si]), 2, '-')
			u4_putc('-')
			u4_puts(D_1E98[93+si])
			for txt_X != 39 {
				u4_putc(' ')
			}
			txt_Y++
		}
	}
	C_45B5()
}

func C_4C42(int) {
	txt_Y = 0
	C_45D6("Mixtures", 0)
	txt_Y = 1
	txt_X = 24
	for si := 0; si < 26; si++ {
		if Party._mixtures[si] != 0 {
			u4_putc(si + 'A')
			u4_putc('-')
			u4_putl(int32(Party._mixtures[si]), 2, '0')
			txt_X -= 4
			txt_Y++
			if txt_Y == 9 {
				txt_Y = 1
				txt_X += 5
				if txt_X >= 39 {
					return
				}
			}
		}
	}
}

var D_19C0 = [14]func(int){
	C_4689, C_4689, C_4689, C_4689, C_4689, C_4689, C_4689, C_4689,
	C_4832, C_48F8, C_4987, C_4A3D, C_4BC7, C_4C42,
}

func C_4CC1(bp04 int) {
	bp_02, bp_04 := txt_Y, txt_X
	n := int(Party.f_1d8)
	for bp04 != -1 {
		Gra_13()
		C_4649()
		D_19C0[bp04](bp04)
		for u_kbhit() == 0 {
		}
		switch si := u_kbread(); si {
		case KBD_1, KBD_2, KBD_3, KBD_4, KBD_5, KBD_6, KBD_7, KBD_8:
			if si&0xf <= n {
				bp04 = si&0xf - 1
			} else {
				sound(1)
			}
		case KBD_RIGHT, KBD_DOWN:
			bp04 = (bp04 + 1) % 14
			if bp04 == n {
				bp04 = 8
			}
		case KBD_UP, KBD_LEFT:
			if bp04 == 8 {
				bp04 = n - 1
			} else {
				bp04 = (bp04 + 13) % 14
			}
		case KBD_PGUP:
			if bp04 >= 1 && bp04 <= 8 {
				bp04 = 0
			} else if bp04 > 8 && bp04 <= 13 {
				bp04 = 8
			} else {
				bp04 = 13
			}
		case KBD_HOME:
			bp04 = 0
		case KBD_0:
			bp04 = 8
		case KBD_PGDN:
			if bp04 >= 0 && bp04 <= 7 {
				bp04 = 8
			} else if bp04 >= 8 && bp04 <= 12 {
				bp04 = 13
			} else {
				bp04 = 0
			}
		case KBD_END:
			bp04 = 13
		default:
			bp04 = -1
		}
	}
	C_4649()
	Gra_13()
	dspl_Stats()
	txt_Y, txt_X = bp_02, bp_04
}

func CMD_Ztats() { /*C_4E45*/
	if CurMode >= MOD_COMBAT {
		u4_puts("Ztats\n")
		Gra_11(int(activeChara))
		C_4CC1(int(activeChara))
		Gra_11(int(activeChara))
	} else {
		bp_02 := AskChara("Ztats for:\x12\x12\b")
		if bp_02 != -1 {
			if bp_02 == -2 {
				bp_02 = 8
			}
			C_4CC1(bp_02)
		}
	}
}
