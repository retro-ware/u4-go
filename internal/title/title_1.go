package title

// Port of TITLE_1.C: character creation. The name and sex, the introduction story with its
// pictures, the gypsy's seven questions on the abacus, the new party (PARTY.NEW with the
// player swapped in as the class of the last chosen virtue), and writing PARTY.SAV and
// MONSTERS.SAV.

/*====---- _BSS ----====*/
var (
	lastVirtue       int16     /*static_B D_693A*/
	M_or_F           int16     /*static_C D_693C*/
	tmp_str          uint8     /*static_D D_693E*/
	D_6940           []byte    /*static_E*/
	tmp_dex          uint8     /*static_F D_6944*/
	curQuestionIndex uint16    /*static_G D_6946*/
	player_name      [20]uint8 /*static_H D_6948*/
	tmp_int          uint8     /*static_I D_695C*/
	tmp_karma        [8]uint8  /*static_J D_695E*/
)

func STR(i int) string { return D_2EE6[i-1] }

func u4_puts(bp04 string) { /*C_261D*/
	for i := 0; i < len(bp04); i++ {
		if bp04[i] == '\n' {
			txt_Y++
			txt_X = 0
		} else {
			u4_putc(bp04[i])
		}
	}
}

// text input: up to bp04-1 characters into si, then leading and trailing blanks removed
func C_2656(si []uint8, bp04 uint16) {
	var di, bp_04 uint16
	for {
		bp_04 = uint16(u_kbread())
		switch bp_04 {
		case KBD_BS, KBD_0e7f, KBD_LEFT:
			if di == 0 {
				sound_1()
			} else {
				u4_putc(8)
				di--
				si[di] = ' '
			}
		case KBD_ENTER:
			si[di] = 0
		default:
			bp_04 &= 0x00ff
			if bp04-1 == di || bp_04 < 0x20 || bp_04 >= 0x80 {
				sound_1()
			} else {
				si[di] = uint8(bp_04)
				u4_putc(uint8(bp_04))
				di++
			}
		}
		if bp_04 == KBD_ENTER {
			break
		}
	}
	bp_04 = 0
	for si[bp_04] == ' ' {
		bp_04++
	}
	if si[bp_04] != 0 {
		di = 0
		for si[bp_04] != 0 {
			si[di] = si[bp_04]
			di++
			bp_04++
		}
	} else {
		di = bp_04
	}
	// si[-1] is the high byte of curQuestionIndex (static_G sits just before player_name),
	// always 0 here: an all-blank name stops there
	at := func(i uint16) uint8 {
		if i == 0xffff {
			return uint8(curQuestionIndex >> 8)
		}
		return si[i]
	}
	for {
		si[di] = 0
		di--
		if at(di) != ' ' {
			break
		}
	}
}

func C_271D(bp04 string) {
	Gra_5()
	txt_X = 0
	txt_Y = 19
	u4_puts(bp04)
	u_kbread()
}

// moongateLoopCycles: one pass of the moongate animations' empty for(int) loop, not scaled
// by speed_info (so the animation runs faster on faster machines, as it did).
const moongateLoopCycles = 25

/*moongate animation #1*/
func C_273E() {
	for bp_02 := 1; bp_02 <= 23; bp_02++ {
		Gra_3(3, bp_02, 0, 152, D_6940, 92-bp_02, -1, 9)
		Gra_3(3, 24-bp_02, 9, 68, D_6940, 68, -1, 9)
		cpu(10000 * moongateLoopCycles)
	}
	Gra_3(3, 24, 0, 152, D_6940, 68, -1, 9)
}

/*moongate animation #2*/
func C_27E0() {
	for bp_02 := 1; bp_02 <= 23; bp_02++ {
		Gra_3(3, bp_02, 3, 152, D_6940, 68, -1, 9)
		Gra_3(3, 24-bp_02, 0, 152, D_6940, bp_02+68, -1, 9)
		cpu(10000 * moongateLoopCycles)
	}
	Gra_3(3, 24, 3, 152, D_6940, 68, -1, 9)
}

// the introduction: texts 1Dh..34h, the pictures changing along the way
func C_2883() {
	D_6940 = make([]byte, 0x8000)
	bp_02 := 0x1d
	Gra_inflate("tree.ega", D_6940)
	Gra_clrscr()
	Gra_3(40, 152, 0, 0, D_6940, 0, -1, 0)
	for bp_02 != 0x35 {
		Gra_5()
		txt_X = 0
		txt_Y = 19
		u4_puts(STR(bp_02))
		bp_02++
		switch bp_02 - 1 {
		case 0x20:
			C_273E() /*moongate animation #1*/
		case 0x22:
			C_27E0() /*moongate animation #2*/
			Gra_inflate("portal.ega", D_6940)
		case 0x23:
			Gra_3(40, 152, 0, 0, D_6940, 0, -1, 0)
			Gra_inflate("tree.ega", D_6940)
		case 0x28:
			Gra_3(40, 152, 0, 0, D_6940, 0, -1, 0)
			Gra_inflate("outside.ega", D_6940)
		case 0x2c:
			Gra_3(40, 152, 0, 0, D_6940, 0, -1, 0)
			Gra_inflate("inside.ega", D_6940)
		case 0x2e:
			Gra_3(40, 152, 0, 0, D_6940, 0, -1, 0)
			Gra_inflate("wagon.ega", D_6940)
		case 0x31:
			Gra_3(40, 152, 0, 0, D_6940, 0, -1, 0)
			Gra_inflate("gypsy.ega", D_6940)
		case 0x32:
			Gra_3(40, 152, 0, 0, D_6940, 0, -1, 0)
			Gra_inflate("abacus.ega", D_6940)
		case 0x34:
			Gra_3(40, 152, 0, 0, D_6940, 0, -1, 0)
		}
		u_kbflush()
		u_kbread()
	}
}

var D_3012 = [...]int{0x18, 0x28, 0x36, 0x44, 0x52, 0x62, 0x70}

// draw an abacus bead: bp08 0 chosen (white), 1 discarded (black); virtue bp06, line bp04
func C_2B2A(bp08, bp06, bp04 int) {
	src := 3
	if bp08 == 0 {
		src = 1
	}
	Gra_3(1, 12, src, 187, D_6940, D_3012[bp04], -1, bp06+16)
}

var D_307E = [...]string{
	"HONCOM.EGA", "VALJUS.EGA", "SACHONOR.EGA", "SPIRHUM.EGA",
	"HONCOM.PIC", "VALJUS.PIC", "SACHONOR.PIC", "SPIRHUM.PIC", // CGA, unused
}

// draw virtue card bp06 on the left (bp04 0) or the right
func C_2B6D(bp06, bp04 int) {
	bp_04 := make([]byte, 0x8000)
	Gra_inflate(D_307E[bp06/2], bp_04)
	sx, dx := 27, 27
	if bp06&1 == 0 {
		sx = 1
	}
	if bp04 == 0 {
		dx = 1
	}
	Gra_3(12, 124, sx, 12, bp_04, 12, -1, dx)
}

/*characterisics increments by virtue*/
var D_30B2 = [...]uint8{0, 0, 3, 0, 1, 1, 1, 0} /*str*/
var D_30BA = [...]uint8{0, 3, 0, 1, 1, 0, 1, 0} /*dex*/
var D_30C2 = [...]uint8{3, 0, 0, 1, 0, 1, 1, 0} /*int*/

/*dilemmas text indexes*/
var D_30CA = [...]uint8{
	0x01 - 1, /*Honesty vs {Compassion|...|Humility}*/
	0x08 - 2, /*Compassion vs {Valor|...|Humility}*/
	0x0e - 3, /*Valor vs {Justice|...|Humility}*/
	0x13 - 4, /*Justice vs {Sacrifice|...|Humility}*/
	0x17 - 5, /*Sacrifice vs {Honor|...|Humility}*/
	0x1a - 6, /*Honor vs {Spirituality|Humility}*/
	0x1c - 7, /*Spirituality vs Humility*/
}

// the gypsy: seven pairs of virtues, the winners meet again
func C_2C12() {
	var loc_A int
	var loc_B, loc_C uint8
	var loc_D [8]uint8

	tmp_int = 15
	tmp_dex = 15
	tmp_str = 15
	curQuestionIndex = 0
	for loc_A = 7; loc_A >= 0; loc_A-- {
		loc_D[loc_A] = 0
		tmp_karma[loc_A] = 50
	}
	for {
		if curQuestionIndex == 4 || curQuestionIndex == 6 {
			for loc_A = 7; loc_A >= 0; loc_A-- {
				if loc_D[loc_A] < 0x80 {
					loc_D[loc_A] = 0
				}
			}
		}
		Gra_5()
		txt_X = 0
		txt_Y = 19
		for {
			loc_B = u_rand_a() & 7
			if loc_D[loc_B] == 0 {
				break
			}
		}
		for {
			loc_C = u_rand_a() & 7
			if loc_D[loc_C] == 0 && loc_C != loc_B {
				break
			}
		}
		if loc_C < loc_B {
			loc_B, loc_C = loc_C, loc_B
		}
		if curQuestionIndex == 0 {
			u4_puts(STR(0x35))
		} else if curQuestionIndex == 6 {
			u4_puts(STR(0x37))
		} else {
			u4_puts(STR(0x36))
		}
		u4_puts(STR(0x38))
		C_2B6D(int(loc_B), 0)
		u4_puts(STR(0x39 + int(loc_B)))
		u4_puts(" and ")
		C_2B6D(int(loc_C), 1)
		u4_puts(STR(0x39 + int(loc_C)))
		u4_puts(".  She says\n\"Consider this:\"")
		u_kbflush()
		u_kbread()
		Gra_5()
		txt_X = 0
		txt_Y = 19
		u4_puts(STR(int(D_30CA[loc_B]) + int(loc_C)))
		for {
			loc_A = u_kbread()
			if u4_isupper(uint8(loc_A)) {
				loc_A += 'a' - 'A'
			}
			if loc_A == KBD_A || loc_A == KBD_B {
				break
			}
		}
		if loc_A == KBD_B {
			loc_B, loc_C = loc_C, loc_B
		}
		/*chosen virtue*/
		lastVirtue = int16(loc_B)
		loc_D[lastVirtue] = 1
		tmp_karma[lastVirtue] += 5
		tmp_str += D_30B2[lastVirtue]
		tmp_dex += D_30BA[lastVirtue]
		tmp_int += D_30C2[lastVirtue]
		C_2B2A(0, int(lastVirtue), int(curQuestionIndex))
		/*discarded virtue*/
		loc_D[loc_C] = 0xff
		C_2B2A(1, int(loc_C), int(curQuestionIndex))
		curQuestionIndex++
		if curQuestionIndex == 7 {
			break
		}
	}
	C_271D(STR(0x42))
	C_271D(STR(0x43))
}

// starting positions by class (the virtue's town)
var D_30DC = [...]uint8{0xE7, 0x53, 0x23, 0x3B, 0x9E, 0x69, 0x17, 0xBA}
var D_30E4 = [...]uint8{0x88, 0x69, 0xDD, 0x2C, 0x15, 0xB7, 0x81, 0xAB}

// karmas: offsets of Party._hones.._humil (D_30EC)
var pKarmas = [8]int{party_hones, party_hones + 2, party_hones + 4, party_hones + 6, party_hones + 8, party_hones + 10, party_hones + 12, party_hones + 14}

// prepare the party structure
func C_2E04() {
	var loc_B [sizeofChara]byte

	if Load("PARTY.NEW", sizeofParty, Party[:]) == -1 {
		exit(3)
	}
	Party[party_x] = D_30DC[lastVirtue]
	Party[party_y] = D_30E4[lastVirtue]
	put16(Party[:], party_f_1d8, 1)
	for loc_A := 7; loc_A >= 0; loc_A-- {
		put16(Party[:], pKarmas[loc_A], uint16(tmp_karma[loc_A]))
	}
	/*-- swap(&(Party.chara[lastVirtue]), &(Party.chara[0])) --*/
	cv := Party[party_chara+int(lastVirtue)*sizeofChara:][:sizeofChara]
	c0 := Party[party_chara:][:sizeofChara]
	copy(loc_B[:], cv)
	copy(cv, c0)
	copy(c0, loc_B[:])
	/*strcpy(Party.chara[0]._name, player_name)*/
	for i := 0; ; i++ {
		c0[chara_name+i] = player_name[i]
		if player_name[i] == 0 {
			break
		}
	}
	put16(c0, chara_str, uint16(tmp_str))
	put16(c0, chara_int, uint16(tmp_int))
	put16(c0, chara_dex, uint16(tmp_dex))
	if M_or_F == 'M' {
		c0[chara_p_24] = 0x0b
	} else {
		c0[chara_p_24] = 0x0c
	}
	D_6976._npc = tNPC{}
}

// write PARTY.SAV and MONSTERS.SAV. The port has no disks: the original asked for the PLAYER
// disk first and the PROGRAM disk after.
func C_2FB8() {
	if Save("PARTY.SAV", sizeofParty, Party[:]) == -1 {
		exit(3)
	}
	if Save("MONSTERS.SAV", 0x100, D_6976._npc.bytes()) == -1 {
		exit(3)
	}
}

// (I)nitiate New Game
func C_3030() {
	Gra_2()
	C_0B1E(16, 4, "By what name shalt thou be known")
	C_0B1E(17, 4, "in this world and time?")
	txt_X = 12
	txt_Y = 19
	C_2656(player_name[:], 12)
	if player_name[0] == 0 {
		pAnim, pShapes = nil, nil
		return
	}
	Gra_2()
	C_0B1E(17, 4, "Art thou Male or Female? ")
	M_or_F = int16(uint8(u_kbread()))
	u4_toupper16(&M_or_F)
	for M_or_F != 'M' && M_or_F != 'F' {
		sound_1()
		M_or_F = int16(uint8(u_kbread()))
		// quirk: compared with the full key words, Enter and Space never match; only Esc cancels
		if M_or_F == 0x1b || M_or_F == KBD_ENTER || M_or_F == KBD_SPACE {
			pAnim, pShapes = nil, nil
			return
		}
		u4_toupper16(&M_or_F)
	}
	u4_putc(uint8(M_or_F))
	pAnim, pShapes = nil, nil
	D_6E80 = 0

	C_2883() /*introduction*/
	C_2C12() /*gipsy*/
	C_2E04() /*prepare party structure*/
	C_2FB8() /*save the new party*/
	D_6940 = nil
	/*-- clean & return to dos --*/
	pCharset = nil
	low_clean()
	exit('0') // the original: 'A'+drive of the player disk; AVATAR only tells '0'/'1' apart
}

func u4_toupper16(c *int16) {
	if *c >= 'a' && *c <= 'z' {
		*c -= 'a' - 'A'
	}
}

// The texts of TITLE_1.C (D_2EE6): STR(i) is D_2EE6[i-1].

var D_2EE6 = [0x43]string{
	/*0x01*/ /*D_0424*/ "Entrusted to deliver an uncounted purse\nof gold, thou dost meet a poor beggar.\nDost thou A) deliver the gold knowing\nthe Trust in thee was well placed, or\nB) show Compassion, giving the beggar a\ncoin, knowing it won't be missed?",
	/*0x02*/ /*D_0509*/ "Thou has been prohibited by thy absent\nLord from joining thy friends in a close\npitched battle. Dost thou A) refrain, so\nthou may Honesty claim obedience, or\nB) show Valor, and aid thy comrades,\nknowing thou may deny it later?",
	/*0x03*/ /*D_05EC*/ "A merchant owes thy friend money, now\nlong past due.  Thou dost see the same\nmerchant drop a purse of gold. Dost thou\nA) Honestly return the purse intact, or\nB) Justly give thy friend a portion of\nthe gold first?",
	/*0x04*/ /*D_06C1*/ "Thee and thy friend are valiant but\npenniless warriors.  Thou both go out to\nslay a mighty dragon.  Thy friend thinks\nhe slew it; thee did.  When asked, dost\nthou A) Truthfully claim the gold, or\nB) Allow thy friend the large reward?",
	/*0x05*/ /*D_07AB*/ "Thou art sworn to protect thy Lord at\nany cost, yet thou know he hath\ncommitted a crime.  Authorities ask thee\nof the affair.  Dost thou A) break thine\noath by Honestly speaking, or B) uphold\nHonor by silently keeping thine oath?",
	/*0x06*/ /*D_0891*/ "Thy friend seeks admittance to thy\nSpiritual order. Thou art asked to vouch\nfor his purity of Spirit, of which thou\nart unsure.  Dost thou A) Honestly\nexpress thy doubt, or B) Vouch for him\nhoping for his Spiritual improvement?",
	/*0x07*/ /*D_0975*/ "Thy Lord mistakenly believes he slew a\ndragon.  Thou hast proof that thy lance\nfelled the beast.  When asked, dost thou\nA) Honestly claim the kill and the\nprize, or B) Humbly permit thy Lord his\nbelief?",
	/*0x08*/ /*D_0A40*/ "Thou dost manage to disarm thy mortal\nenemy in a duel.  He is at thy mercy.\nDost thou A) show Compassion by\npermitting him to yield, or B) slay him\nas expected of a Valiant duelist?",
	/*0x09*/ /*D_0AF6*/ "After 20 years thou hast found the\nslayer of thy best friends.  The villain\nproves to be a man who provides the sole\nsupport for a young girl.  Dost thou\nA) spare him in Compassion for the girl,\nor B) slay him in the name of Justice?",
	/*0x0a*/ /*D_0BE0*/ "Thee and thy friends have been routed\nand ordered to retreat.  In defiance of\nthy orders, dost thou A) stop in\nCompassion to aid a wounded companion,\nor B) Sacrifice thyself to slow the\npursuing enemy, so others may escape?",
	/*0x0b*/ /*D_0CC0*/ "Thou art sworn to uphold a Lord who\nparticipates in the forbidden torture of\nprisoners.  Each night their cries of\npain reach thee.  Dost thou A) show\nCompassion by reporting the deeds, or\nB) Honor thy oath and ignore the deeds?",
	/*0x0c*/ /*D_0DA5*/ "Thou hast been taught to preserve all\nlife as sacred. A man lies fatally stung\nby a venomous serpent.  He pleads for a\nmerciful death.  Dost thou A) show\nCompassion and end his pain, or B) heed\nthy Spiritual beliefs and refuse?",
	/*0x0d*/ /*D_0E89*/ "As one of the King's Guard, thy Captain\nhas asked that one amongst you visit\na hospital to cheer the children with\ntales of thy valiant deeds.  Dost thou\nA) Show thy Compassion and play the\nbraggart, or B) Humbly let another go?",
	/*0x0e*/ /*D_0F6E*/ "Thou hast been sent to secure a needed\ntreaty with a distant Lord.  Thy host is\nagreeable to the proposal but insults\nthy country at dinner.  Dost thou\nA) Valiantly bear the slurs, or\nB) Justly rise and demand an apology?",
	/*0x0f*/ /*D_104C*/ "A mighty knight accosts thee and demands\nthy food.  Dost thou A) Valiantly refuse\nand engage the knight, or B) Sacrifice\nthy food unto the hungry knight?",
	/*0x10*/ /*D_10E6*/ "During battle thou art ordered to guard\nthy commander's empty tent.  The battle\ngoes poorly and thou dost yearn to aid\nthy fellows.  Dost thou A) Valiantly\nenter the battle to aid thy companions,\nor B) Honor thy post as guard?",
	/*0x11*/ /*D_11C9*/ "A local bully pushes for a fight.  Dost\nthou A) Valiantly trounce the rogue, or\nB) Decline, knowing in thy Spirit that\nno lasting good will come of it?",
	/*0x12*/ /*D_1261*/ "Although a teacher of music, thou art\na skillful wrestler.  Thou hast been\nasked to fight in a local championship.\nDost thou A) accept the invitation and\nValiantly fight to win, or B) Humbly\ndecline knowing thou art sure to win?",
	/*0x13*/ /*D_1346*/ "During a pitched battle, thou dost see\na fellow desert his post, endangering\nmany.  As he flees, he is set upon by\nseveral enemies. Dost thou A) Justly let\nhim fight alone, or B) Risk Sacrificing\nthine own life to aid him?",
	/*0x14*/ /*D_1425*/ "Thou hast sworn to do thy Lord's\nbidding in all.  He covets a piece of\nland and orders the owner removed.  Dost\nthou A) serve Justice refusing to act,\nthus being disgraced, or B) Honor thine\noath and unfairly evict the landowner?",
	/*0x15*/ /*D_150B*/ "Thou dost believe that virtue resides\nin all people.  Thou dost see a rogue\nsteal from thy Lord.  Dost thou A) call\nhim to Justice, or B) personally try to\nsway him back to the Spiritual path of\ngood?",
	/*0x16*/ /*D_15D4*/ "Unwitnessed, thou hast slain a great\ndragon in self defense.  A poor warrior\nclaims the offered reward.  Dost thou\nA) Justly step forward to claim the\nreward, or B) Humbly go about life,\nsecure in thy self esteem?",
	/*0x17*/ /*D_16AA*/ "Thou art a bounty hunter sworn to return\nan alleged murder.  After his capture\nthou believest him to be innocent.  Dost\nthou A) Sacrifice thy sizable bounty\nfor thy belief, or B) Honor thy oath to\nreturn him as thou hast promised?",
	/*0x18*/ /*D_1791*/ "Thou hast spent thy life in charitable\nand righteous work.  Thine uncle the\ninnkeeper lies ill and asks you to take\nover his tavern.  Dost thou A) Sacrifice\nthy life of purity to aid thy kin, or\nB) decline & follow thy Spirit's call?",
	/*0x19*/ /*D_187B*/ "Thou art an elderly, wealthy eccentric.\nThy end is near. Dost thou A) donate all\nthy wealth to feed hundreds of starving\nchildren, and receive public adulation,\nor B) Humbly live out thy life, willing\nthy fortune to thy heirs?",
	/*0x1a*/ /*D_195E*/ "In thy youth thou pledged to marry thy\nsweetheart.  Now thou art on a sacred\nquest in distant lands.  Thy sweetheart\nasks thee to keep thy vow.  Dost thou\nA) Honor thy pledge to wed, or B) follow\nthy Spiritual crusade?",
	/*0x1b*/ /*D_1A39*/ "Thou art at a crossroads in thy life.\nDost thou A) Choose the Honorable life\nof a Paladin, striving for Truth and\nCourage, or B) Choose the Humble life\nof a Shepherd, and a world of simplicity\nand peace?",
	/*0x1c*/ /*D_1B05*/ "Thy parents wish thee to become an\napprentice. Two positions are available.\nDost thou A) Become an acolyte in the\nSpiritual order, or B) Become an\nassistant to a humble village cobbler?",
	/*0x1d*/ /*D_1BBF*/ "  The day is warm, yet there is a\ncooling breeze.  The latest in a series\nof personal crises seems insurmountable.\nYou are being pulled apart in all\ndirections.",
	/*0x1e*/ /*D_1C60*/ "Yet this afternoon walk in the country-\nside slowly brings relaxation to your\nharried mind.  The soil and strain of\nmodern high-tech living begins to wash\noff in layers. That willow tree near the\nstream looks comfortable and inviting.",
	/*0x1f*/ /*D_1D4B*/ "The buzz of dragonflies and the whisper\nof the willow's swaying branches bring\na deep peace.  Searching inward for\ntranquility and happiness, you close\nyour eyes.",
	/*0x20*/ /*D_1DEE*/ "A high-pitched cascading sound like\ncrystal wind-chimes impinges on your\nfloating awareness.  As you open your\neyes, you see a shimmering blueness rise\nfrom the ground.  The sound seems to be\nemanating from this glowing portal.",
	/*0x21*/ /*D_1ED2*/ "It is difficult to look at the blueness.\nLight seems to bend and distort around\nit, while the sound waves become so\nintense, they appear to become visible.",
	/*0x22*/ /*D_1F6E*/ "The portal hangs there for a moment;\nthen, with the rush of an imploding\nvacuum, it sinks into the ground.\nSomething remains suspended in mid-air\nfor a moment before falling to earth\nwith a heavy thud.",
	/*0x23*/ /*D_2038*/ "Somewhat shaken by this vision, you\nrise to your feet to investigate.  A\ncrude circle of stones surrounds the\nspot where the portal appeared. There is\nsomething glinting in the grass.",
	/*0x24*/ /*D_20F0*/ "You pick up an amulet shaped like a\ncross with a loop at the top.  It is an\nAnkh, the sacred symbol of life and\nrebirth.  But this could not have made\nthe thud, so you look again and find\na large book wrapped in thick cloth!",
	/*0x25*/ /*D_21D1*/ "With trembling hands you unwrap the\nbook.  Behold, the cloth is a map, and\nwithin lies not one book, but two.  The\nmap is of a land strange to you, and the\nstyle speaks of ancient cartography.",
	/*0x26*/ /*D_2292*/ "The script on the cover of the first\nbook is arcane but readable.  The title\nis:\n        The History of Britannia\n               as told by\n            Kyle the Younger",
	/*0x27*/ /*D_233B*/ "The other book is disturbing to look at.\nIts small cover appears to be fashioned\nout of some sort of leathery hide, but\nfrom what creature is uncertain.  The\nreddish-black skin radiates an intense\naura suggestive of ancient power.",
	/*0x28*/ /*D_2422*/ "The tongue of the title is beyond your\nken.  You dare not open the book and\ndisturb whatever sleeps within.  You\ndecide to peruse the History.  Settling\nback under the willow tree, you open the\nbook.",
	/*0x29*/ /*D_24EA*/ "\n\n    (You read the Book of History)",
	/*0x2a*/ /*D_250F*/ "\n\n(No, really! Read the Book of History!)",
	/*0x2b*/ /*D_2539*/ "Closing the book, you again pick up\nthe Ankh.  As you hold it, you begin to\nhear a hauntingly familiar, lute-like\nsound wafting over a nearby hill.  Still\nclutching the strange artifacts, you\nrise unbidden and climb the slope.",
	/*0x2c*/ /*D_261C*/ "In the valley below you see what appears\nto be a fair.  It seems strange that you\ncame that way earlier and noticed\nnothing.  As you mull this over, your\nfeet carry you down towards the site.",
	/*0x2d*/ /*D_26DC*/ "This is no ordinary travelling\ncarnival, but a Renaissance Fair.  The\npennants on the tent tops blow briskly\nin the late afternoon breeze.",
	/*0x2e*/ /*D_2767*/ "The ticket taker at the RenFair's gate\nstarts to ask you for money, but upon\nspotting your Ankh says, \"Welcome,\nfriend.  Enter in peace and find your\npath.\"",
	/*0x2f*/ /*D_2804*/ "The music continues to pull you\nforward amongst the merchants and\nvendors.  Glimpses of fabulous treasures\ncan be seen in some of the shadowy\nbooths.",
	/*0x30*/ /*D_289A*/ "These people are very happy.  They seem\nto glow with an inner light.  Some look\nup as you pass and smile, but you cannot\nstop - the music compels you to move\nonward through the crowd.",
	/*0x31*/ /*D_2952*/ "Through the gathering dusk you see a\nsecluded gypsy wagon sitting off in the\nwoods.  The music seems to emanate from\nthe wagon.  As you draw near, a woman's\nvoice weaves into the music, saying:\n\"You may approach, O seeker.\"",
	/*0x32*/ /*D_2A32*/ "You enter to find an old gypsy sitting\nin a small curtained room.  She wears an\nAnkh around her neck. In front of her is\na round table covered in deep green\nvelvet.  The room smells so heavily of\nincense that you feel dizzy.",
	/*0x33*/ /*D_2B13*/ "Seeing the Ankh, the ancient gypsy\nsmiles and warns you never to part with\nit.  \"We have been waiting such a long\ntime, but at last you have come.  Sit\nhere and I shall read the path of your\nfuture.\"",
	/*0x34*/ /*D_2BDB*/ "Upon the table she places a curious\nwooden object like an abacus but without\nbeads.  In her hands she holds eight\nunusual cards.  \"Let us begin the\ncasting.\"",
	/*0x35*/ /*D_2C79*/ "The gypsy places the first two cards\n",
	/*0x36*/ /*D_2C9F*/ "The gypsy places two more of the cards\n",
	/*0x37*/ /*D_2CC7*/ "The gypsy places the last two cards\n",
	/*0x38*/ /*D_2CEC*/ "upon the table.  They are the cards of\n",
	/*0x39*/ /*D_2D14*/ "Honesty",
	/*0x3a*/ /*D_2D1C*/ "Compassion",
	/*0x3b*/ /*D_2D27*/ "Valor",
	/*0x3c*/ /*D_2D2D*/ "Justice",
	/*0x3d*/ /*D_2D35*/ "Sacrifice",
	/*0x3e*/ /*D_2D3F*/ "Honor",
	/*0x3f*/ /*D_2D45*/ "Spirituality",
	/*0x40*/ /*D_2D52*/ "Humility",
	/*0x41*/ /*D_2D5B*/ "",
	/*0x42*/ /*D_2D5C*/ "With the final choice, the incense\nswells up around you.  The gypsy speaks\nas if from a great distance, her voice\ngrowing fainter with each word:\n\"So be it!  Thy path is chosen!\"",
	/*0x43*/ /*D_2E0F*/ "There is a moment of intense, wrenching\nvertigo.  As you open your eyes, a\nvoice whispers within your mind, \"Seek\nthe counsel of thy sovereign.\"  After a\nmoment, the spinning subsides, and you\nopen your eyes to...",
}
