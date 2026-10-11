package u4

// Port of U4_USE.C (unsigned char): using the quest items.

const D_00EE = "Hmm...No effect!\n"
const D_0100 = "None owned!\n"

// colors masks
var D_010E = [3]uint8{
	1<<6 | 1<<5 | 1<<3 | 1<<0, // truth
	1<<6 | 1<<4 | 1<<3 | 1<<1, // love
	1<<6 | 1<<5 | 1<<4 | 1<<2, // courage
}

func C_01E1() {
	var bp_12 [12]byte
	a := int(activeChara)
	if Combat._charaX[a] != 5 || Combat._charaY[a] != 5 || D_943E == -1 {
		u4_puts(D_00EE)
		return
	}
	u4_puts("\nThere are holes for 4 stones. What colors:\n")
	bp_04 := uint8(0)
	for bp_02 := 1; bp_02 <= 4; bp_02++ {
		u4_putc(bp_02 + 'A' - 1)
		u4_putc(':')
		getsWords = stoneColors()
		u4_gets(bp_12[:], 11)
		Gra_CR()
		bp_10 := 7
		for ; bp_10 >= 0; bp_10-- {
			if strnicmp(bp_12[:], D_0884[bp_10], 12) == 0 {
				break
			}
		}
		if bp_10 == -1 || !TST_MSK8(Party.mStones, bp_10) {
			if bp_12[0] != 0 {
				u4_puts(D_0100)
			}
			return
		}
		if (1<<bp_10)&bp_04 != 0 {
			u4_puts("Already used!\n")
			return
		}
		bp_04 |= 1 << bp_10
	}
	if D_010E[D_943E] != bp_04 {
		u4_puts(D_00EE)
		return
	}
	key := [3]int{ST_KEY_T, ST_KEY_L, ST_KEY_C}[D_943E]
	if TST_MSK(Party.mItems, key) {
		u4_puts(D_00EE)
		return
	}
	SET_MSK(&Party.mItems, key)
	u4_puts("Thou doth find one third of the Three Part Key!\n")
}

var D_0284 = [8]string{
	"Truth",
	"Love",
	"Courage",
	"Truth and Love",
	"Love and Courage",
	"Courage and Truth",
	"Truth, Love and Courage",
	"\nA voice rings out:  What Virtue exists independently of Truth, Love and Courage",
}

// use stone(s)
func C_0311() {
	if Party.mStones == 0 {
		u4_puts(D_0100)
		return
	}
	var bp_10 [14]byte
	if CurMode != MOD_COM_ROOM && CurMode != MOD_DUNGEON {
		u4_puts("No place to Use them!\n")
	}
	if CurMode != MOD_DUNGEON {
		C_01E1()
		return
	}
	if tile_cur != 0xb0 || Party._loc != 0x18 {
		u4_puts(D_00EE)
		return
	}
	if Party._z != 7 {
		u4_puts("\nAs thou doth approach, a voice rings out: What virtue dost stem from ")
	}
	u4_puts(D_0284[Party._z])
	u4_puts("?\n\n")
	u4_gets(bp_10[:], 13)
	Gra_CR()
	if strnicmp(bp_10[:], D_1E98[151+int(Party._z)], 14) != 0 {
		if bp_10[0] != 0 {
			u4_puts(D_00EE)
		}
		return
	}
	u4_puts("\nThe Voice says: Use thy Stone.\n\nColor:\n")
	getsWords = stoneColors()
	u4_gets(bp_10[:], 11)
	Gra_CR()
	bp_02 := 7
	for strnicmp(bp_10[:], D_0884[bp_02], 12) != 0 {
		bp_02--
		if bp_02 < 0 {
			break
		}
	}
	if bp_02 == -1 {
		if bp_10[0] != 0 {
			u4_puts(D_00EE)
		}
		return
	}
	if !TST_MSK8(Party.mStones, bp_02) {
		u4_puts("\nYou have none!\n\n")
		return
	}
	if int(Party._z) != bp_02 {
		u4_puts(D_00EE)
		return
	}
	if Party._z == 7 {
		C_31F4() // last phase of game
		return
	}
	*D_8742.x8x8x8(int(Party._z), int(Party._y), int(Party._x)) = 0x20
	u4_puts("\nThe altar changes before thyne eyes!\n")
}

// use key(s)
func C_044C() {
	if (Party.mItems>>ST_KEY_T)&1|(Party.mItems>>ST_KEY_L)&1|(Party.mItems>>ST_KEY_C)&1 != 0 {
		u4_puts("No place to Use them!\n")
	} else {
		u4_puts(D_0100)
	}
}

func atAbyss() bool { return Party._loc == 0 && Party._x == 0xe9 && Party._y == 0xe9 }

// use bell
func C_0487() {
	if !TST_MSK(Party.mItems, ST_BELL) {
		u4_puts(D_0100)
		return
	}
	if !atAbyss() {
		u4_puts(D_00EE)
		return
	}
	SET_MSK(&Party.mItems, ST_USE_BELL)
	u4_puts("The Bell rings on and on!\n")
}

// use book
func C_04C0() {
	if !TST_MSK(Party.mItems, ST_BOOK) {
		u4_puts(D_0100)
		return
	}
	if !atAbyss() || !TST_MSK(Party.mItems, ST_USE_BELL) {
		u4_puts(D_00EE)
		return
	}
	SET_MSK(&Party.mItems, ST_USE_BOOK)
	u4_puts("The words resonate with the ringing!\n")
}

// use candle
func C_0501() {
	if !TST_MSK(Party.mItems, ST_CANDLE) {
		u4_puts(D_0100)
		return
	}
	if !atAbyss() || !TST_MSK(Party.mItems, ST_USE_BOOK) {
		u4_puts(D_00EE)
		return
	}
	SET_MSK(&Party.mItems, ST_USE_CANDLE)
	u4_puts("As you light the Candle the Earth Trembles!\n")
	sound(6)
	shakefx()
}

// use horn
func C_0553() {
	if !TST_MSK(Party.mItems, ST_HORN) {
		u4_puts(D_0100)
		return
	}
	if Party._loc != 0 {
		u4_puts(D_00EE)
		return
	}
	u4_puts("The Horn sounds an eerie tone!\n")
	spell_sta = 1
	spell_cnt = 10
	dspl_Stats()
}

// use wheel
func C_058C() {
	if !TST_MSK(Party.mItems, ST_WHEEL) {
		u4_puts(D_0100)
		return
	}
	if Party._loc != 0 || Party._tile > TIL_13 || Party._ship != 50 {
		u4_puts(D_00EE)
		return
	}
	u4_puts("Once mounted, the Wheel glows with a blue light!\n")
	Party._ship = 99
	dspl_Stats()
}

// use skull
func C_05CE() {
	if !TST_MSK(Party.mItems, ST_SKULL) {
		u4_puts(D_0100)
		return
	}
	shake3 := func() {
		sound(6)
		shakefx()
		Gra_09()
		sound(6)
		shakefx()
		Gra_09()
		sound(6)
		shakefx()
	}
	if !atAbyss() {
		u4_puts("\nYou hold the evil Skull of Mondain the Wizard aloft....\n")
		shake3()
		n := &D_8742._npc
		if CurMode >= MOD_COMBAT {
			for b := 31; b >= 0; b-- {
				if n._tile[b] != TIL_5E {
					n._gtile[b] = 0
					n._tile[b] = 0
				}
			}
			for b := 15; b >= 0; b-- {
				if Fighters._tile[b] != TIL_5E {
					Fighters._gtile[b] = 0
					Fighters._tile[b] = 0
				}
			}
		} else {
			b := 31
			if Party._loc == 0 {
				b = 7
			}
			for ; b >= 0; b-- {
				if n._tile[b] != TIL_5E {
					n._gtile[b] = 0
					n._tile[b] = 0
				}
			}
		}
		t_callback()
		for _, k := range pKarmas {
			karma_dec(k, 5)
		}
		dspl_Stats()
	} else {
		u4_puts("\nYou cast the Skull of Mondain into the Abyss!\n")
		SET_MSK(&Party.mItems, ST_CAST_SKULL)
		for _, k := range pKarmas {
			karma_inc(k, 10)
		}
		shake3()
	}
	RST_MSK(&Party.mItems, ST_SKULL)
}

var D_0434 = []struct {
	_00 string
	_02 func()
}{
	{"stone", C_0311},
	{"stones", C_0311},
	{"bell", C_0487},
	{"book", C_04C0},
	{"candle", C_0501},
	{"key", C_044C},
	{"keys", C_044C},
	{"horn", C_0553},
	{"wheel", C_058C},
	{"skull", C_05CE},
}

func CMD_Use() { /*C_07AE*/
	var bp_0e [12]byte
	u4_puts("Use which item:\n")
	getsWords = usableItems()
	u4_gets(bp_0e[:], 11)
	Gra_CR()
	for _, u := range D_0434 {
		if strnicmp(bp_0e[:], u._00, 12) == 0 {
			u._02()
			return
		}
	}
	u4_puts("Not a Usable item!\n")
}
