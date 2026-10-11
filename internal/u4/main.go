package u4

// Port of U4_MAIN.C and U4_MAIN2.C (unsigned char), plus Main, the game's entry point.

// Main runs the game (AVATAR.EXE) on the host's game goroutine. It returns nil when the game
// wants the title back (no party formed), else the error that stopped it.
func Main() (err error) {
	defer func() {
		if r := recover(); r != nil {
			s, ok := r.(stop)
			if !ok {
				panic(r)
			}
			err = s.err
		}
	}()
	resetGlobals()
	main_()
	return nil
}

// catchJmp runs f, returning normally when it longjmps to D_9458.
func catchJmp(f func()) (jumped bool) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(jmpMain); ok {
				jumped = true
				return
			}
			panic(r)
		}
	}()
	f()
	return false
}

func main_() { /*C_191E*/
	var si int
	low_init()
	C_C51C()
	if Party._loc >= 0x11 && Party._loc <= 0x18 {
		CurMode = MOD_DUNGEON
		Load("DNGMAP.SAV", D_8742._map[:0x200])
		File_DNG = readFile(D_0894[Party._loc-0x11])
		catchJmp(DNG_main)
	} else {
		Party._loc = 0
	}
	u_kbflush()
	bp_04 := 0
	for {
		catchJmp(mainStep(&si, &bp_04))
	}
}

// mainStep is one pass of main's command loop (setjmp(D_9458) sits at its top).
func mainStep(psi, pbp_04 *int) func() {
	return func() {
		for {
			si := *psi
			t_callback()
			C_9209()
			if C_10FD() {
				u4_putc(0x10)
				ask(pCmd)
				u_delay(25, 1)
				if u_kbhit() != 0 {
					si = u_kbread()
				} else {
					si = KBD_SPACE
				}
				if u4_isupper(si & 0xff) {
					si = (si & 0xff00) | u4_lower(si&0xff)
				}
				*psi = si
				switch si {
				case KBD_SPACE:
					w_Pass()
				case 0x487e, KBD_UP:
					CMDDIR_Up()
				case 0x4b7c, KBD_LEFT:
					CMDDIR_Left()
				case 0x5060, KBD_DOWN:
					CMDDIR_Down()
				case 0xf400, KBD_RIGHT:
					CMDDIR_Right()
				case KBD_A:
					CMD_Attack()
				case KBD_B:
					CMD_Board()
				case KBD_C:
					CMD_Cast()
				case KBD_D:
					CMD_Descend()
				case KBD_E:
					CMD_Enter()
				case KBD_F:
					CMD_Fire()
				case KBD_G:
					CMD_Get()
				case KBD_H:
					CMD_HoleUp()
				case KBD_I:
					CMD_Ignite()
				case KBD_J:
					CMD_Jimmy()
				case KBD_K:
					CMD_Klimb()
				case KBD_L:
					CMD_Locate()
				case KBD_M:
					CMD_Mix()
				case KBD_N:
					CMD_NewOrder()
				case KBD_O:
					CMD_Open()
				case KBD_P:
					CMD_Peer()
				case KBD_Q:
					CMD_Quit()
				case KBD_R:
					CMD_Ready()
				case KBD_S:
					CMD_Search()
				case KBD_T:
					CMD_Talk()
				case KBD_U:
					CMD_Use()
				case KBD_V:
					CMD_Volume()
				case KBD_W:
					CMD_Wear()
				case KBD_X:
					CMD_X_it()
				case KBD_Y:
					CMD_Yell()
				case KBD_Z:
					CMD_Ztats()
				case KBD_CTRL_S:
					if *pbp_04 == KBD_ALT_Z {
						C_1C21()
						break
					}
					fallthrough
				default:
					u4_puts("Bad command!\n")
					sound(2)
					D_07F8 = 0
				}
			} else {
				u4_puts("Zzzzz\n")
			}
			if D_07F8 != 0 {
				C_1C53()
			} else {
				t_callback()
				D_07F8 = 1
			}
			*pbp_04 = *psi
		}
	}
}

// cheat code --> reveal the different levels of karmas
func C_1C21() {
	Gra_CR()
	for si := 0; si < 8; si++ {
		u4_putl(int32(*pKarmas[si]), 2, '0')
	}
	Gra_CR()
}

func C_1C53() {
	C_10FD()
	// restore some hull
	if Party._ship < 50 && U4_RND1(3) == 0 {
		Party._ship++
	}
	// S=>G(randomly), P=>takes 2 hit
	for si := 0; si < int(Party.f_1d8); si++ {
		di := &Party.chara[si]
		if di._stat == 'S' && U4_RND1(7) == 0 {
			di._stat = 'G'
		} else if di._stat == 'P' {
			hitChara(si, 2)
			Gra_11(si)
			sound(6)
			Gra_11(si)
		}
	}
	// FOOD management
	if food_dec(int(Party.f_1d8)) {
		u4_puts("\nStarving!!!\n")
		for si := 0; si < int(Party.f_1d8); si++ {
			if isCharaAlive(si) {
				hitChara(si, 2)
			}
		}
		for si := int(Party.f_1d8) - 1; si >= 0; si-- {
			Gra_11(si)
		}
		sound(6)
		for si := int(Party.f_1d8) - 1; si >= 0; si-- {
			Gra_11(si)
		}
	}
	MP_recover()
	if Party.f_1dc == 0 {
		C_7918()
		C_5834()
		C_5851()
		C_7918()
	}
	Party._moves++
	if spell_cnt != 0 {
		spell_cnt--
		if spell_cnt == 0 {
			spell_sta = 7
		}
	}
	dspl_Stats()
	// automatic door closing
	C_431D()
}

// ---- U4_MAIN2.C ----

var D_07F8 uint16 = 1

// karmas
var pKarmas = [8]*uint16{
	&Party._hones, &Party._compa, &Party._valor, &Party._justi,
	&Party._sacri, &Party._honor, &Party._spiri, &Party._humil,
}

var activeCharaX int16 = -1 /*D_080A*/

var D_080C = [4]uint8{0xff, 0, 1, 0} // {-1,0,1,0}
var D_0810 = [4]uint8{0, 0xff, 0, 1} // {0,-1,0,1}

// moongates positions
var D_0814 = [8]uint8{0xE0, 0x60, 0x26, 0x32, 0xA6, 0x68, 0x17, 0xBB}
var D_081C = [8]uint8{0x85, 0x66, 0xE0, 0x25, 0x13, 0xC2, 0x7E, 0xA7}

// towns/castles files
var D_0824 = [16]string{
	// Castles
	"LCB_1.ULT", "LYCAEUM.ULT", "EMPATH.ULT", "SERPENT.ULT",
	// Townes
	"MOONGLOW.ULT", "BRITAIN.ULT", "JHELOM.ULT", "YEW.ULT",
	"MINOC.ULT", "TRINSIC.ULT", "SKARA.ULT", "MAGINCIA.ULT",
	// Villages
	"PAWS.ULT", "DEN.ULT", "VESPER.ULT", "COVE.ULT",
}

// places world's coordinates
var D_0844 = [32]uint8{0x56, 0xDA, 0x1C, 0x92, 0xE8, 0x52, 0x24, 0x3A, 0x9F, 0x6A, 0x16, 0xBB, 0x62, 0x88, 0xC9, 0x88, 0xF0, 0x5B, 0x48, 0x7E, 0x9C, 0x3A, 0xEF, 0xE9, 0xE9, 0x80, 0x24, 0x49, 0xCD, 0x51, 0xE7, 0xE7}
var D_0864 = [32]uint8{0x6B, 0x6B, 0x32, 0xF1, 0x87, 0x6A, 0xDE, 0x2B, 0x14, 0xB8, 0x80, 0xA9, 0x91, 0x9E, 0x3B, 0x5A, 0x49, 0x43, 0xA8, 0x14, 0x1B, 0x66, 0xF0, 0xE9, 0x42, 0x5C, 0xE5, 0x0B, 0x2D, 0xCF, 0xD8, 0xD8}

// colors
var D_0884 = [8]string{"Blue", "Yellow", "Red", "Green", "Orange", "Purple", "White", "Black"}

// dungeons files
var D_0894 = [8]string{
	"Deceit.Dng", "Despise.Dng", "Destard.Dng", "Wrong.Dng",
	"Covetous.Dng", "Shame.Dng", "Hythloth.Dng", "Abyss.Dng",
}
