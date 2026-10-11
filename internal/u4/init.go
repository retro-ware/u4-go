package u4

// Port of U4_INIT.C: start-up, reading the graphics and save files. The port only has the EGA driver, so the adapter is always EGA.

func C_C51C() {
	// set drive number: dropped, the port has no drives or disks
	// set graphic type: EGA only
	D_943A = 2
	pShapes = make([]byte, 0x8000)
	pCharset = make([]byte, 0x1000)
	Load("CHARSET.EGA", pCharset)
	C_20C1("EGA.DRV")
	Gra_init()
	drv.SetGlyphs(pCharset)
	Gra_clrscr()
	Load("SHAPES.EGA", pShapes)
	drv.SetTiles(pShapes)
	LoadRec("PARTY.SAV", &Party)
	if Party.f_1d8 == 0 {
		Gra_clrscr()
		txt_Y = 8
		txt_X = 12
		u4_puts("No party formed!")
		sound(8)
		u_delay(3, 0)
		u_kbflush()
		low_clean()
		backToTitle()
	}
	LoadRec("MONSTERS.SAV", &D_8742._npc)
	if File_MAP = readFile("WORLD.MAP"); len(File_MAP) < 64*0x400 {
		fatalf("WORLD.MAP: file wrong size")
	}
	File_TLK = nil
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
