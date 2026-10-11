package u4

// Port of U4_UTIL.C (signed char).

// shake screen + noise
func shakefx() { /*C_095E*/
	Gra_04()
	Gra_03()
	Gra_04()
	Gra_03()
	Gra_04()
	Gra_03()
	Gra_04()
	Gra_03()
}

// increase exp (max 9999)
func XP_inc(bp06, bp04 int) { /*C_097D*/
	Party.chara[bp06]._XP += uint16(bp04)
	if Party.chara[bp06]._XP > 9999 {
		Party.chara[bp06]._XP = 9999
	}
}

func w_Pass() { /*C_09A3*/
	u4_puts("Pass\n")
}

// increments hit points
func HP_inc(bp06, bp04 int) { /*C_09B1*/
	si := &Party.chara[bp06]._HP
	si[0] += uint16(bp04)
	if si[0] > si[1] {
		si[0] = si[1]
	}
}

func C_09D9(bp04 int) {
	Gra_11(bp04)
	sound(7)
	Gra_11(bp04)
}

// increment karma (max 99)
func karma_inc(bp06 *uint16, bp04 int) { /*C_09F8*/
	if *bp06 != 0 {
		*bp06 += uint16(bp04)
		if *bp06 > 99 {
			*bp06 = 99
		}
	}
}

// decrement karma
func karma_dec(bp06 *uint16, bp04 int) { /*C_0A17*/
	if *bp06 == 0 {
		*bp06 = 99
		Gra_19()
		txt_X = 31
		txt_Y = 10
		u4_putc(int(spell_sta))
		u4_puts("\n Thou hast lost\n  an Eighth!\n")
	}
	*bp06 -= uint16(bp04)
	if *bp06 == 0 { // unsigned "<= 0"
		*bp06 = 1
	}
}

// findEnemyAt
func C_0A58(bp08 int, bp06, bp04 uint8) int {
	var i int
	for i = bp08; i >= 0; i-- {
		if D_8742._npc._tile[i] != 0 &&
			D_8742._npc._x[i] == bp06 &&
			D_8742._npc._y[i] == bp04 {
			break
		}
	}
	return i
}

func C_0A8B(bp06, bp04 uint8) int {
	if Party._loc == 0 {
		return C_0A58(7, bp06, bp04)
	}
	return C_0A58(31, bp06, bp04)
}

func SafeModulo(bp06, bp04 int) int { /*C_0AB1*/
	if bp04 == 0 {
		return 0
	}
	return int(int16(bp06) % int16(bp04))
}

// get character sprite id ?
func C_0ACF(bp04 int) int {
	if Party.chara[bp04]._stat == 'S' {
		return TIL_38
	}
	return (sc(Party.chara[bp04]._class) << 1) + TIL_20
}

// isNotEvil: 1 if the creature's defeat leads to karma increase
func C_0AFE(bp04 uint8) int {
	if bp04 < TIL_80 ||
		bp04 == TIL_8A || bp04 == TIL_90 ||
		bp04 == TIL_94 || bp04 == TIL_98 ||
		bp04 == TIL_B4 || bp04 == TIL_CC {
		return 1
	}
	return 0
}

func u4_puts(txt string) int { /*C_0B38*/
	at := func(i int) byte {
		if i < len(txt) {
			return txt[i]
		}
		return 0
	}
	loc_A, loc_B := 0, 0
	i := 0
	for at(i) != 0 {
		loc_B--
		if loc_B+1 == 0 {
			for loc_B = 0; ; loc_B++ {
				loc_C := at(i + loc_B)
				if loc_C == 0 || loc_C == '\n' || loc_C == ' ' {
					break
				}
			}
			if loc_B+int(txt_X) > 40 && txt_X != 24 {
				if loc_A == 12 {
					loc_A++
					u_kbflush()
					if txt_X == 39 {
						for u_kbhit() == 0 {
						}
					}
					u_kbread()
					loc_A = 0
				} else {
					loc_A++
				}
				Gra_CR()
			}
		}
		if at(i) == '\n' {
			if loc_A == 12 {
				u_kbflush()
				if txt_X == 39 {
					for u_kbhit() == 0 {
					}
				}
				u_kbread()
				loc_A = 0
			} else {
				loc_A++
			}
		}
		u4_putc(int(at(i)))
		i++
	}
	return i
}

func u4_putl(bp0a int32, bp06 int, bp04 int) { /*C_0C03*/
	if bp0a < 0 {
		u4_putc('-')
		u4_putl(-bp0a, bp06-1, bp04)
	}
	if bp0a < 10 {
		if bp06 > 1 {
			u4_putc(bp04)
			u4_putl(bp0a, bp06-1, bp04)
		} else {
			u4_putc(int(bp0a + '0'))
		}
	} else {
		u4_putl(bp0a/10, bp06-1, bp04)
		u4_putc(int(bp0a%10 + '0'))
	}
}

func u4_putc(c int) { /*C_0C9F*/
	bp04 := uint8(c)
	switch bp04 {
	case '\b':
		if txt_X > 1 {
			txt_X--
			Gra_putchar(' ')
		}
	case '\n':
		Gra_CR()
	case ' ':
		if txt_X <= 39 {
			Gra_putchar(bp04)
			txt_X++
		}
	default:
		if txt_X > 39 {
			Gra_CR()
		}
		Gra_putchar(bp04)
		txt_X++
	}
}

// display party info
func dspl_Stats() { /*C_0CF7*/
	Gra_19()
	loc_E := txt_Y
	loc_A := txt_X
	for loc_B := 0; loc_B < int(Party.f_1d8); loc_B++ {
		loc_D := &Party.chara[loc_B]
		txt_Y = int16(loc_B + 1)
		txt_X = 24
		u4_putc(loc_B + '1')
		u4_putc('-')
		for loc_C := 0; loc_C < 16 && loc_D._name[loc_C] != 0 && loc_C < 8; loc_C++ {
			u4_putc(int(loc_D._name[loc_C]))
		}
		txt_X = 35
		u4_putl(int32(loc_D._HP[0]), 3, ' ')
		u4_putc(int(loc_D._stat))
	}
	txt_X = 24
	txt_Y = 10
	u4_putc('F')
	u4_putc(':')
	u4_putl(Party._food/100, 4, '0')
	u4_putc(' ')
	u4_putc(int(spell_sta))
	if Party._tile < TIL_14 {
		u4_puts(" SHP:")
		u4_putl(int32(Party._ship), 2, '0')
	} else {
		u4_puts(" G:")
		u4_putl(int32(Party._gold), 4, '0')
	}
	txt_Y = loc_E
	txt_X = loc_A
}

// display gold
func dspl_Gold() { /*C_0E0E*/
	bp_02 := txt_X
	bp_04 := txt_Y
	txt_X = 35
	txt_Y = 10
	u4_putl(int32(Party._gold), 4, '0')
	txt_Y = bp_04
	txt_X = bp_02
}

// i.e. state 'G' or 'P' or 'S'
func isCharaAlive(bp04 int) bool { /*C_0E4E*/
	if bp04 >= int(Party.f_1d8) {
		return false
	}
	switch Party.chara[bp04]._stat {
	case 'G', 'P', 'S':
		return true
	}
	return false
}

// i.e. state 'G' or 'P'
func isCharaConscious(bp04 int) bool { /*C_0E82*/
	if bp04 >= int(Party.f_1d8) {
		return false
	}
	switch Party.chara[bp04]._stat {
	case 'G', 'P':
		return true
	}
	return false
}

// death/revive
func C_0EB1() {
	for bp_02 := int(Party.f_1d8) - 1; bp_02 >= 0; bp_02-- {
		Party.chara[bp_02]._stat = 'D'
		Party.chara[bp_02]._HP[0] = 0
	}
	Gra_13()
	dspl_Stats()
	u_delay(5, 0)
	Gra_10()
	CurMode = MOD_VISION
	u4_puts("\n\n\nAll is Dark...\n")
	u_delay(5, 0)
	u4_puts("\nBut wait...\n")
	u_delay(5, 0)
	u4_puts("Where am I?...\n")
	u_delay(5, 0)
	u4_puts("Am I dead?...\n")
	u_delay(5, 0)
	u4_puts("Afterlife?...\n")
	u_delay(5, 0)
	u4_puts("You hear:\n")
	txt_X = int16(uint16(12-uint16(strlen(Party.chara[0]._name[:])))/2 + 25)
	u4_puts(cstr(Party.chara[0]._name[:]))
	Gra_CR()
	u_delay(5, 0)
	u4_puts("I feel motion...\n")
	u_delay(5, 0)
	Party._tile = TIL_1F
	Party._x = 0x13
	Party._y = 0x08
	Party.out_x = 0x56
	Party.out_y = 0x6c

	File_TLK = nil
	if File_DNG != nil {
		File_TLK = nil /*bug??: the C closed File_DNG but left its handle set*/
	}
	if Party._loc == 0 {
		Save("OUTMONST.SAV", &D_8742._npc)
	}
	LoadRec("LCB_2.ULT", &D_8742)
	File_TLK = readFile("LCB.TLK")
	CurMode = MOD_BUILDING
	Party._loc = 0x01
	Party.f_1dc = 0
	D_95CC = MOD_OUTDOORS
	spell_cnt = 0
	hit_tile = 0
	spell_sta = 7
	D_9440 = 1
	D_95C6 = 0
	t_callback()
	C_3A80()
	u4_puts("\n\nLord British says: I have pulled thy spirit and some possessions from the void.  Be more careful in the future!\n")
	for bp_02 := int(Party.f_1d8) - 1; bp_02 >= 0; bp_02-- {
		Party.chara[bp_02]._stat = 'G'
		Party.chara[bp_02]._HP[0] = Party.chara[bp_02]._HP[1]
	}
	for bp_02 := 15; bp_02 >= 0; bp_02-- {
		Party._weapons[bp_02] = 0
	}
	for bp_02 := 7; bp_02 >= 0; bp_02-- {
		Party._armors[bp_02] = 0
	}
	Party._food = 20099
	Party._gold = 200
	dspl_Stats()
	u_kbflush()
	longjmp_D_9458()
}

// 1: at least one chara conscious; 0: all charas disabled and at least one alive
func C_10FD() bool {
	for si := int(Party.f_1d8) - 1; si >= 0; si-- {
		if isCharaAlive(si) {
			for ; si >= 0; si-- {
				if isCharaConscious(si) {
					return true
				}
			}
			return false
		}
	}
	C_0EB1() // does not return
	return false
}

// character bp06 takes bp04 hits -> 0: alive, 1: died
func hitChara(bp06 int, bp04 int) int { /*C_1135*/
	si := &Party.chara[bp06]
	if si._HP[0] >= uint16(bp04) {
		si._HP[0] -= uint16(bp04)
		dspl_Stats()
		return 0
	}
	// character dies
	si._HP[0] = 0
	si._stat = 'D'
	if CurMode >= MOD_COMBAT {
		Fighters._chtile[bp06] = TIL_38
		D_944A[bp06] = TIL_38
	}
	dspl_Stats()
	return 1
}

func w_What() { /*C_1188*/
	u4_puts("What?\n")
	u_kbflush()
}

func w_Cant_t() { /*C_1199*/
	u4_puts("Can't!\n")
	u_kbflush()
}

func w_NotHere() { /*C_11AA*/
	u4_puts("Not Here!\n")
	u_kbflush()
}

func w_Disabled() { /*C_11BB*/
	u4_puts("Disabled!\n")
	u_kbflush()
}

func w_OnlyOnFoot() { /*C_11CC*/
	u4_puts("Only on foot!\n")
	u_kbflush()
}

func w_SlowProgress() { /*C_11DD*/
	u4_puts("Slow progress!\n")
}

func w_NoneLeft() { /*C_11EB*/
	u4_puts("None left!\n")
}

func AskLetter(bp08 string, bp06, bp04 int) int { /*C_11F9*/
	si := -1
	for {
		if si != -1 {
			sound(1)
		}
		u4_puts(bp08)
		si = u_kbread()
		if si == KBD_ENTER || si == KBD_ESC || si == KBD_SPACE {
			if si == KBD_ENTER || si == KBD_SPACE {
				si = -1
			} else {
				si = -2
			}
			Gra_CR()
			break
		}
		si &= 0xff
		if u4_islower(si) {
			si += 'A' - 'a'
		}
		if si > ' ' && si < 0x7f {
			u4_putc(si)
		}
		Gra_CR()
		if bp06 <= si && si <= bp04 {
			break
		}
		if si < 0 {
			break
		}
	}
	return si
}

func AskChara(bp04 string) int { /*C_1287*/
	if Party.f_1d8 == 1 {
		u4_puts(bp04)
		u4_puts("1\n")
		return 0
	}
	si := AskLetter(bp04, '0', '0'+int(Party.f_1d8))
	if si == '0' {
		return -2
	}
	if si < 0 {
		return -1
	}
	return si - '1'
}

func AskDir(bp08 string, si, di *int) { /*C_12D6*/
	*si, *di = 0, 0
	for {
		bp_02 := 1
		u4_puts(bp08)
		bp_04 := u_kbread()
		switch bp_04 {
		case KBD_SPACE, KBD_ESC, KBD_BS, KBD_ENTER:
		case KBD_UP:
			*di = -1
			u4_puts("North")
		case KBD_DOWN:
			*di = 1
			u4_puts("South")
		case KBD_RIGHT:
			*si = 1
			u4_puts("East")
		case KBD_LEFT:
			*si = -1
			u4_puts("West")
		default:
			bp_04 &= 0xff
			if bp_04 >= ' ' && bp_04 < 0x7f {
				u4_putc(bp_04)
			}
			Gra_CR()
			sound(1)
			bp_02 = 0
		}
		if bp_02 != 0 {
			break
		}
	}
	Gra_CR()
}

func food_dec(bp04 int) bool { /*C_138B*/
	Party._food -= int32(bp04)
	if Party._food >= 0 {
		return false
	}
	Party._food = 0
	return true
}

func MP_recover() { /*C_13B6*/
	for bp_04 := int(Party.f_1d8) - 1; bp_04 >= 0; bp_04-- {
		if isCharaAlive(bp_04) {
			si := &Party.chara[bp_04]
			var di int
			in := int(si._int)
			switch sc(si._class) {
			case 0:
				di = in * 2
			case 1, 5, 6:
				di = in
			case 3:
				di = in/2 + in
			case 4:
				di = in / 2
			default:
				di = 0
			}
			if di > 99 {
				di = 99
			}
			si._MP++
			if si._MP > uint16(di) {
				si._MP = uint16(di)
			}
		}
	}
}

// u4_gets reads a line into si (a C buffer of bp04 bytes, zero-terminated).
func u4_gets(si []byte, bp04 int) { /*C_1445*/
	loc_A := 0
	var loc_B int
	for {
		loc_B = u_kbread()
		switch loc_B {
		case KBD_BS, KBD_0e7f, KBD_LEFT:
			if loc_A == 0 {
				sound(1)
			} else {
				u4_putc(8)
				loc_A--
				si[loc_A] = ' '
			}
		case KBD_ENTER:
			si[loc_A] = 0
		default:
			loc_B &= 0xff
			if bp04-1 == loc_A || loc_B < ' ' || loc_B >= 0x80 {
				sound(1)
			} else {
				si[loc_A] = byte(loc_B)
				u4_putc(loc_B)
				loc_A++
			}
		}
		if loc_B == KBD_ENTER {
			break
		}
	}
	// strip leading and trailing blanks
	loc_B = 0
	for si[loc_B] == ' ' {
		loc_B++
	}
	if si[loc_B] != 0 {
		loc_A = 0
		for si[loc_B] != 0 {
			si[loc_A] = si[loc_B]
			loc_A++
			loc_B++
		}
	} else {
		loc_A = loc_B
	}
	for {
		si[loc_A] = 0
		loc_A--
		if loc_A < 0 || si[loc_A] != ' ' { // C reads si[-1] when the line was empty
			break
		}
	}
}

// get creature name
func C_1513(bp04 uint8) string {
	if bp04 >= TIL_80 {
		if bp04 >= TIL_90 {
			return D_1E98[9+int(bp04-TIL_80)/4-(TIL_90-TIL_80)/4]
		}
		return D_1E98[1+int(bp04-TIL_80)/2]
	}
	if bp04 == TIL_38 {
		return D_1E98[159]
	}
	if bp04 < TIL_20 || bp04 >= TIL_60 || (bp04 >= TIL_30 && bp04 < TIL_50) {
		return D_1E98[20] /*"Phantom"*/
	}
	return D_1E98[77+int(bp04&0x1f)/2]
}

// all party damage
func C_1584() {
	for loc_B := int(Party.f_1d8) - 1; loc_B >= 0; loc_B-- {
		Gra_11(loc_B)
	}
	sound(6)
	shakefx()
	for loc_B := int(Party.f_1d8) - 1; loc_B >= 0; loc_B-- {
		Gra_11(loc_B)
	}
	if CurMode >= MOD_COMBAT || Party._tile > TIL_13 {
		// normal case
		for loc_B := int(Party.f_1d8) - 1; loc_B >= 0; loc_B-- {
			if U4_RND1(1) != 0 && isCharaAlive(loc_B) {
				loc_A := U4_RND3(15) + 10
				if CurMode < MOD_COMBAT || Fighters._chtile[loc_B] != 0 {
					hitChara(loc_B, loc_A)
				}
			}
		}
	} else {
		// on ship: _ship is unsigned, so the C's "< 0" test is never true
		Party._ship -= 10
	}
	dspl_Stats()
}

func AskY_N() int { /*C_162F*/
	bp_02 := -1
	for {
		if bp_02 != -1 {
			sound(1)
		}
		bp_02 = u_kbread() & 0x7f
		u4_toupper(&bp_02)
		if bp_02 == 'N' || bp_02 == 'Y' || bp_02 == ' ' || bp_02 == 0x1b || bp_02 == '\r' {
			break
		}
	}
	if bp_02 == 'Y' || bp_02 == 'N' {
		u4_putc(bp_02)
	}
	Gra_CR()
	return int(int8(bp_02))
}

func AskInt(bp04 int) int32 { /*C_169C*/
	var bp_14 [20]byte
	u4_gets(bp_14[:], bp04+1)
	Gra_CR()
	if bp_14[0] != 0 {
		return atol(bp_14[:])
	}
	return -1
}
