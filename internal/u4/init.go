package u4

// Port of U4_INIT.C: start-up, reading the command line ULTIMA.COM passed and the graphics
// and save files. The port only has the EGA driver, so the adapter is always EGA.

var D_9138 int

// confirmFileInDrive: the game files are always there, unless the player deleted one;
// then this waits for a key forever, as the original did with no disk in the drive.
func C_C454(bp06, bp04 string) {
	if C_182F(bp04) != 0 {
		return
	}
	CurMode = MOD_VISION
	txt_Y = 10
	txt_X = int16(uint(40-len(bp06)) >> 1)
	u4_puts(bp06)
	txt_Y++
	txt_X = 18
	u4_puts("and")
	txt_Y++
	txt_X = 11
	u4_puts("press drive letter")
	for u_kbhit() == 0 {
	}
	for {
		bp_02 := u_kbread() & 0xff
		u4_toupper(&bp_02)
		if C_1814() == bp_02 && C_182F(bp04) != 0 {
			return
		}
		sound(1)
		for u_kbhit() == 0 {
		}
	}
}

func C_C51C() {
	NPARAM := psp[0x80]
	PARAM1 := int(psp[0x84])
	u4_toupper(&PARAM1)
	psp[0x84] = byte(PARAM1)
	// set drive number
	if NPARAM < 4 {
		D_9138 = b2i((equip_flags&0xc0)>>6 != 0)
		psp[0x84] = byte(D_9138 + '0')
	} else if PARAM1 == '0' {
		D_9138 = 0
	} else if PARAM1 == '1' {
		D_9138 = 1
	} else {
		D_9138 = b2i((equip_flags&0xc0)>>6 != 0)
	}
	// set graphic type: EGA only
	D_943A = 2
	pShapes = dalloc(0x8000)
	pCharset = dalloc(0x1000)
	if Load("CHARSET.EGA", 0x1000, pCharset) == -1 {
		exit(3)
	}
	bp_02 := C_1814()
	C_20C1("EGA.DRV")
	C_184F() // piracy check
	Gra_init()
	sizzleCharset()
	C_18A2()
	Gra_clrscr()
	C_C454("Insert PLAYER disk", "WORLD.MAP")
	if Load("SHAPES.EGA", 0x8000, pShapes) == -1 {
		exit(3)
	}
	sizzleShapes()
	if Load("PARTY.SAV", sizeof(&Party), &Party) == -1 {
		exit(3)
	}
	if Party.f_1d8 == 0 {
		Gra_clrscr()
		txt_Y = 8
		txt_X = 12
		u4_puts("No party formed!")
		sound(8)
		u_delay(3, 0)
		u_kbflush()
		if bp_02 == C_1814() {
			C_C454("Re-Insert PROGRAM Disk", "TITLE.EXE")
		}
		low_clean()
		exit(4)
	}
	if Load("MONSTERS.SAV", sizeof(&D_8742._npc), &D_8742._npc) == -1 {
		exit(3)
	}
	if File_MAP = dopen("WORLD.MAP", 0); File_MAP == -1 {
		Console("Can't open map file\r\n")
		exit(3)
	}
	File_TLK = 0
	D_9445 = D_0814[Party._trammel]
	D_9448 = D_081C[Party._felucca]
	C_213B() // draw game screen frame
	Gra_CR()
	spell_sta = 7
	spell_cnt = 0
	D_9440 = uint16(b2i(Party.f_1dc == 0)) // flying status
	D_95C6 = 0
	CurMode = MOD_OUTDOORS
	WindDir = DIR_N
	SoundFlag = 1
	dspl_Stats()
	C_26B6()
}
