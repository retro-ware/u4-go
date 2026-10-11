package u4

// Port of U4_INIT.C: start-up, reading the command line ULTIMA.COM passed and the graphics
// and save files. The port only has the EGA driver, so the adapter is always EGA.

func C_C51C() {
	// set drive number: dropped, the port has no drives or disks
	// set graphic type: EGA only
	D_943A = 2
	pShapes = dalloc(0x8000)
	pCharset = dalloc(0x1000)
	if Load("CHARSET.EGA", 0x1000, pCharset) == -1 {
		exit(3)
	}
	C_20C1("EGA.DRV")
	Gra_init()
	sizzleCharset()
	Gra_clrscr()
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
