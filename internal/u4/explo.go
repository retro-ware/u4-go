package u4

// Port of U4_EXPLO.C: entering and leaving places, boarding, doors, klimb and descend.

var D_1738 = [16]string{
	// Castles
	"LCB.TLK", "LYCAEUM.TLK", "EMPATH.TLK", "SERPENT.TLK",
	// Townes
	"MOONGLOW.TLK", "BRITAIN.TLK", "JHELOM.TLK", "YEW.TLK",
	"MINOC.TLK", "TRINSIC.TLK", "SKARA.TLK", "MAGINCIA.TLK",
	// Villages
	"PAWS.TLK", "DEN.TLK", "VESPER.TLK", "COVE.TLK",
}

// load castle/towne/village files
func C_3E30(bp04 int) {
	if Save("OUTMONST.SAV", sizeof(&D_8742._npc), &D_8742._npc) == -1 {
		exit(3)
	}
	if Load(D_0824[bp04-0x01], sizeof(&D_8742), &D_8742) == -1 {
		exit(3)
	}
	File_TLK = dopen(D_1738[Party._loc-1], 0)
	CurMode = MOD_BUILDING
}

// load dungeon files
func C_3E94() {
	if Load(D_0894[Party._loc-0x11], 0x200, &D_8742._map) == -1 {
		exit(3)
	}
	File_DNG = dopen(D_0894[Party._loc-0x11], 0)
	D_8742._npc._tile = [32]uint8{}
}

func C_3EE4() {
	C_3E94()
	CurMode = MOD_DUNGEON
	Party._dir = DIR_E
	Party._z = 0
	DNG_main()
}

// enter dungeon
func C_3F03() {
	Gra_CR()
	Gra_CR()
	if Party._tile != TIL_1F {
		Party._loc = 0
		w_OnlyOnFoot()
		return
	}
	Party._x = 1
	Party._y = 1
	if Save("OUTMONST.SAV", sizeof(&D_8742._npc), &D_8742._npc) == -1 {
		exit(3)
	}
	C_3EE4()
}

// enter towne
func C_3F4A() {
	u4_puts(D_1E98[127+int(Party._loc)-1])
	Gra_CR()
	Gra_CR()
	Party._y = 15
	D_959C.y = 15
	Party._x = 1
	D_959C.x = 1
	C_3E30(int(Party._loc))
	if Party._loc-0x05 < 8 {
		for si := int(Party.f_1d8) - 1; si >= 1; si-- {
			if uint16(Party.chara[si]._class) == Party._loc-0x05 {
				D_8742._npc._var[31] = 0
				D_8742._npc._gtile[31] = 0
				D_8742._npc._tile[31] = 0
			}
		}
	}
}

// enter the abyss?
func C_3FB9() {
	it := Party.mItems
	if (it>>ST_BOOK)&(it>>ST_BELL)&(it>>ST_CANDLE)&(it>>ST_USE_BOOK)&(it>>ST_USE_BELL)&(it>>ST_USE_CANDLE)&1 != 0 {
		u4_puts(D_1E98[127+int(Party._loc)-1])
		C_3F03()
	} else {
		Party._loc = 0
		w_Cant_t()
	}
}

func CMD_Enter() { /*C_4018*/
	u4_puts("Enter ")
	if Party._loc != 0 || Party._tile == TIL_18 {
		w_What()
		return
	}
	si := 31
	for ; si >= 0; si-- {
		if Party._x == D_0844[si] && Party._y == D_0864[si] {
			break
		}
	}
	if si == -1 {
		w_What()
		return
	}
	Party._loc = uint16(si + 1)
	Party.out_x = Party._x
	Party.out_y = Party._y
	name := func() string { return D_1E98[127+int(Party._loc)-1] }
	switch *D_8742.x32x32(int(D_959C.y), int(D_959C.x)) {
	case TIL_09:
		u4_puts("dungeon!\n\n")
		u4_puts(name())
		C_3F03()
	case TIL_0A:
		u4_puts("towne!\n\n")
		C_3F4A()
	case TIL_0B, TIL_0E:
		u4_puts("castle!\n\n")
		u4_puts(name())
		Gra_CR()
		Gra_CR()
		Party._y = 0x1e
		D_959C.y = 0x1e
		Party._x = 0x0f
		D_959C.x = 0x0f
		Party._z = 0
		C_3E30(int(Party._loc))
	case TIL_0C:
		u4_puts("village!\n\n")
		C_3F4A()
	case TIL_1D:
		u4_puts("ruin!\n\n")
		C_3F4A()
	case TIL_4C:
		if Party._x != 0xe9 || Party._y != 0xe9 {
			w_What()
			return
		}
		C_3FB9()
	case TIL_46:
		C_3FB9()
	case TIL_1E:
		u4_puts("the Shrine of\n")
		u4_puts(D_1E98[151+int(Party._loc)-0x19])
		Gra_CR()
		C_E72C()
	default:
		w_What()
	}
}

func C_4164() {
	n := &D_8742._npc
	si := 31
	for ; si >= 8; si-- {
		if n._tile[si] == 0 {
			break
		}
	}
	if si == 7 {
		si = U4_RND1(15) | 0x10
	}
	n._tile[si] = uint8(Party._tile)
	n._gtile[si] = uint8(Party._tile)
	n._x[si] = Party._x
	n._y[si] = Party._y
	n._tlkidx[si] = 0
	n._var[si] = 0
	Party._tile = TIL_1F
	Gra_CR()
}

func CMD_X_it() { /*C_41C0*/
	u4_puts("X-it ")
	switch {
	case Party._tile < TIL_14:
		ship_x = Party._x
		ship_y = Party._y
		C_4164()
	case Party._tile < TIL_16:
		D_95C6 = 0
		C_4164()
	case Party._tile == TIL_18 && Party.f_1dc == 0:
		C_4164()
	default:
		w_What()
	}
}

func C_4206(bp04 int) {
	n := &D_8742._npc
	si := 31
	for ; si >= 8; si-- {
		if Party._x == n._x[si] && Party._y == n._y[si] && bp04 == int(n._tile[si]&^3) {
			break
		}
	}
	if si != 7 {
		n._tile[si] = 0
		n._gtile[si] = 0
	}
	Party._tile = uint16(tile_cur)
}

func CMD_Board() { /*C_4253*/
	if Party._tile != TIL_1F {
		u4_puts("Board: ")
		w_Cant_t()
		return
	}
	if tile_cur == TIL_14 || tile_cur == TIL_15 {
		C_4206(TIL_14)
		u4_puts("Mount Horse!\n")
		return
	}
	u4_puts("Board ")
	if tile_cur == TIL_18 {
		C_4206(TIL_18)
		u4_puts("Balloon!\n")
		Party.f_1dc = 0
		return
	}
	if tile_cur < TIL_10 || tile_cur > TIL_13 {
		w_What()
		return
	}
	C_4206(TIL_10)
	u4_puts("Frigate!\n")
	if ship_x != Party._x || ship_y != Party._y {
		Party._ship = 50
	}
}

func CMD_Yell() { /*C_42E8*/
	u4_puts("Yell ")
	if Party._tile == TIL_14 || Party._tile == TIL_15 {
		D_95C6 ^= 1
		if D_95C6 != 0 {
			u4_puts("Giddyup!\n")
		} else {
			u4_puts("Whoa!\n")
		}
	} else {
		w_What()
	}
}

var (
	D_17FA uint16 // current opened door's x
	D_17FC uint16 // current opened door's y
	D_17FE uint16 // current opened door's delay
)

// automatic door close
func C_431D() {
	if CurMode == MOD_BUILDING && D_17FE != 0 {
		D_17FE--
		if D_17FE == 0 {
			*D_8742.x32x32(int(D_17FC), int(D_17FA)) = TIL_3B
		}
	} else {
		D_17FE = 0
	}
}

func CMD_Open() { /*C_4353*/
	var bp_02, bp_04 int
	for D_17FE != 0 {
		C_431D()
	}
	if CurMode != MOD_BUILDING {
		u4_puts("Open; ")
		w_NotHere()
		return
	}
	AskDir("Open: ", &bp_02, &bp_04)
	if bp_02|bp_04 == 0 {
		return
	}
	x := uint16(bp_02) + uint16(Party._x) // unsigned
	if x > 0x1f || bp_04 != 0 {
		w_NotHere()
		return
	}
	si := D_8742.x32x32(int(Party._y), int(x))
	switch *si {
	case TIL_3A:
		w_Cant_t()
	case TIL_3B:
		D_17FA = x
		D_17FC = uint16(Party._y)
		D_17FE = 5
		*si = TIL_3E
		u4_puts("\nOpened!\n")
	default:
		w_NotHere()
	}
}

func CMD_Jimmy() { /*C_43F8*/
	var bp_02, bp_04 int
	u4_puts("Jimmy lock!\n")
	AskDir("Dir: ", &bp_02, &bp_04)
	if bp_02|bp_04 == 0 {
		return
	}
	x := bp_02 + int(Party._x)
	if CurMode != MOD_BUILDING || x > 0x1f || bp_04 != 0 {
		w_NotHere()
		return
	}
	si := D_8742.x32x32(int(Party._y), x)
	if *si != TIL_3A {
		w_NotHere()
		return
	}
	if Party._keys == 0 {
		u4_puts("No keys left!\n")
		return
	}
	Party._keys--
	*si = TIL_3B
}

func CMD_Klimb() { /*C_4477*/
	u4_puts("Klimb ")
	if Party._loc == 0 {
		if Party._tile != TIL_18 {
			w_What()
			return
		}
		u4_puts("altitude\n")
		Party.f_1dc = 1
		D_9440 = 0
		return
	}
	if Party._loc == 1 && tile_cur == TIL_1B {
		if Party._tile != TIL_1F {
			Gra_CR()
			w_OnlyOnFoot()
			return
		}
		u4_puts("to second floor!\n")
		if Load("LCB_2.ULT", sizeof(&D_8742), &D_8742) == -1 {
			exit(3)
		}
	} else {
		w_What()
	}
}

func CMD_Descend() { /*C_44EE*/
	if Party._tile == TIL_18 {
		u4_puts("Land Balloon\n")
		if tile_cur != TIL_04 {
			sound(1)
			w_NotHere()
			return
		}
		if Party.f_1dc == 0 {
			u4_puts("Already Landed!\n")
			return
		}
		Party.f_1dc = 0
		D_9440 = 1
		return
	}
	if Party._loc != 0x01 {
		u4_puts("Descend ")
		w_What()
		return
	}
	// in LB's castle
	if Party._tile != TIL_1F {
		w_OnlyOnFoot()
		return
	}
	u4_puts("Descend ")
	if tile_cur != TIL_1C {
		w_What()
		return
	}
	if Party._y == 2 {
		u4_puts("into the depths!\n")
		Party.out_x = 0xef
		Party.out_y = 0xf0
		Party._x = 5
		Party._y = 5
		Party._loc = 0x17
		C_3EE4()
		return
	}
	u4_puts("to first floor!\n")
	if Load("LCB_1.ULT", sizeof(&D_8742), &D_8742) == -1 {
		exit(3)
	}
}
