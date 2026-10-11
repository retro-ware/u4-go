package u4

// Port of U4_LB.C: talking to Lord British.

var D_6FF0 = []string{
	"bye",
	"help",
	"health",
	"name",
	"look",
	"job",
	"truth",
	"love",
	"courage",
	"honesty",
	"compassion",
	"valor",
	"justice",
	"sacrifice",
	"honor",
	"spirituality",
	"humility",
	"pride",
	"avatar",
	"quest",
	"britannia",
	"ankh",
	"abyss",
	"mondain",
	"minax",
	"exodus",
	"virtue",
}

var D_7028 = []string{
	"\n\n\nHe says:\nMy name is\nLord British,\nSovereign of\nall Britannia!\n",
	"\n\n\n\n\nThou see the\nKing with the\nRoyal Sceptre.\n",
	"\n\n\nHe says:\nI rule all\nBritannia, and\nshall do my best\nto help thee!\n",
	"He says:\nMany truths can\nbe learned at\nthe Lycaeum.  It\nlies on the\nnorthwestern\nshore of Verity\nIsle!\n",
	"He says:\nLook for the\nmeaning of Love\nat Empath Abbey.\nThe Abbey sits\non the western\nedge of the Deep\nForest!\n",
	"\n\nHe says:\nSerpent's Castle\non the Isle of\nDeeds is where\nCourage should\nbe sought!\n",
	"\nHe says:\nThe fair towne\nof Moonglow on\nVerity Isle is\nwhere the virtue\nof Honesty\nthrives!\n",
	"\n\nHe says:\nThe bards in the\ntowne of Britain\nare well versed\nin the virtue of\nCompassion!\n",
	"\n\nHe says:\nMany valiant\nfighters come\nfrom Jhelom\nin the Valarian\nIsles!\n",
	"\n\n\nHe says:\nIn the city of\nYew, in the Deep\nForest, Justice\nis served!\n",
	"\nHe says:\nMinoc, towne of\nself-sacrifice,\nlies on the\neastern shores\nof Lost Hope\nBay!\n",
	"\nHe says:\nThe Paladins who\nstrive for Honor\nare oft seen in\nTrinsic, north\nof the Cape of\nHeroes!\n",
	"\nHe says:\nIn Skara Brae\nthe Spiritual\npath is taught.\nFind it on an\nisle near\nSpiritwood!\n",
	"\n\n\nHe says:\nHumility is the\nfoundation of\nVirtue!  The\nruins of proud\nMagincia are a\ntestimony unto\nthe Virtue of\nHumility!\n\nFind the Ruins\nof Magincia far\noff the shores\nof Britannia,\non a small isle\nin the vast\nOcean!\n",
	"\n\n\nHe says:\nOf the eight\ncombinations of\nTruth, Love and\nCourage, that\nwhich contains\nneither Truth,\nLove nor Courage\nis Pride.\n\nPride being not\na Virtue must be\nshunned in favor\nof Humility, the\nVirtue which is\nthe antithesis\nof Pride!\n",
	"\n\n\n\n\n\nLord British\nsays:\nTo be an Avatar\nis to be the\nembodiment of\nthe Eight\nVirtues.\n\n\nIt is to live a\nlife constantly\nand forever in\nthe Quest to\nbetter thyself\nand the world in\nwhich we live.\n",
	"\n\n\nLord British\nsays:\nThe Quest of\nthe Avatar is\nto know and\nbecome the\nembodiment of\nthe Eight\nVirtues of\nGoodness!\nIt is known that\nall who take on\nthis Quest must\nprove themselves\nby conquering\nthe Abyss and\nViewing the\nCodex of\nUltimate Wisdom!\n",
	"\n\n\n\n\n\n\nHe says:\nEven though the\nGreat Evil Lords\nhave been routed\nevil yet remains\nin Britannia.\n\n\n\n\n\nIf but one soul\ncould complete\nthe Quest of the\nAvatar, our\npeople would\nhave a new hope,\na new goal for\nlife.\n\nThere would be a\nshining example\nthat there is\nmore to life\nthan the endless\nstruggle for\npossessions\nand gold!\n",
	"He says:\nThe Ankh is the\nsymbol of one\nwho strives for\nVirtue.  Keep it\nwith thee at all\ntimes for by\nthis mark thou\nshalt be known!\n",
	"\n\n\n\n\n\nHe says:\nThe Great\nStygian Abyss\nis the darkest\npocket of evil\nremaining in\nBritannia!\n\n\n\n\n\n\n\nIt is said that\nin the deepest\nrecesses of the\nAbyss is the\nChamber of the\nCodex!\n\n\nIt is also said\nthat only one of\nhighest Virtue\nmay enter this\nChamber, one\nsuch as an\nAvatar!!!\n",
	"\n\n\n\n\n\nHe says:\nMondain is dead!\n",
	"\n\n\n\n\n\nHe says:\nMinax is dead!\n",
	"\n\n\n\n\n\nHe says:\nExodus is dead!\n",
	"\nHe says:\nThe Eight\nVirtues of the\nAvatar are:\nHonesty,\nCompassion,\nValor,\nJustice,\nSacrifice,\nHonor,\nSpirituality,\nand Humility!\n",
}

// 10 seconds delay
func C_E1FC() {
	u_kbflush()
	u_delay(10, 1)
	if u_kbhit() != 0 {
		u_kbread()
	}
}

// putsWait prints each text, waiting between them.
func putsWait(ss ...string) {
	for i, s := range ss {
		u4_puts(s)
		if i < len(ss)-1 {
			C_E1FC()
		}
	}
}

// when asked 'help'
func C_E21E() {
	u4_puts("He says: ")
	it := Party.mItems
	switch {
	case Party._moves < 1000:
		putsWait("To survive in this hostile land thou must first know thyself! Seek ye to master thy weapons and thy magical ability!\n",
			"\nTake great care in these thy first travels in Britannia.\n",
			"\nUntil thou dost well know thyself, travel not far from the safety of the townes!\n")
	case Party.f_1d8 == 1:
		putsWait("Travel not the open lands alone. There are many worthy people in the diverse townes whom it would be wise to ask to Join thee!\n",
			"\nBuild thy party unto eight travellers, for only a true leader can win the Quest!\n")
	case Party.mRunes == 0:
		putsWait("Learn ye the paths of virtue. Seek to gain entry unto the eight shrines!\n",
			"\nFind ye the Runes, needed for entry into each shrine, and learn each chant or \"Mantra\" used to focus thy meditations.\n",
			"\nWithin the Shrines thou shalt learn of the deeds which show thy inner virtue or vice!\n",
			"\nChoose thy path wisely for all thy deeds of good and evil are remembered and can return to hinder thee!\n")
	case Party._hones&Party._compa&Party._valor&Party._justi&Party._sacri&Party._honor&Party._spiri&Party._humil != 0:
		putsWait("Visit the Seer Hawkwind often and use his wisdom to help thee prove thy virtue.\n",
			"\nWhen thou art ready, Hawkwind will advise thee to seek the Elevation unto partial Avatarhood in a virtue.\n",
			"\nSeek ye to become a partial Avatar in all eight virtues, for only then shalt thou be ready to seek the codex!\n")
	case Party.mStones == 0:
		putsWait("Go ye now into the depths of the dungeons. Therein recover the 8 colored stones from the altar pedestals in the halls of the dungeons.\n",
			"\nFind the uses of these stones for they can help thee in the Abyss!\n")
	case anyKarma() != 0:
		u4_puts("Thou art doing very well indeed on the path to Avatarhood! Strive ye to achieve the Elevation in all eight virtues!\n")
	case (it>>ST_BELL)&(it>>ST_BOOK)&(it>>ST_CANDLE)&1 == 0:
		u4_puts("Find ye the Bell, Book and Candle!  With these three things, one may enter the Great Stygian Abyss!\n")
	case (it>>ST_KEY_C)&(it>>ST_KEY_L)&(it>>ST_KEY_T)&1 == 0:
		putsWait("Before thou dost enter the Abyss thou shalt need the Key of Three Parts, and the Word of Passage.\n",
			"\nThen might thou enter the Chamber of the Codex of Ultimate Wisdom!\n")
	default:
		putsWait("Thou dost now seem ready to make the final journey into the dark Abyss! Go only with a party of eight!\n",
			"\nGood Luck, and may the powers of good watch over thee on this thy most perilous endeavor!\n",
			"\nThe hearts and souls of all Britannia go with thee now. Take care, my friend.\n")
	}
}

// looks for the parameter in Lord British's keywords (4 chars are enough): the index of
// the keyword, or -1
func C_E37E(bp04 []byte) int {
	for i, k := range D_6FF0 {
		if strnicmp(bp04, k, 4) == 0 {
			return i
		}
	}
	return -1
}

func C_E3D2(bp04 string) {
	di := 0
	for i := 0; i < len(bp04); i++ {
		if bp04[i] == '\n' {
			if di == 12 {
				C_E1FC()
				di = 0
			} else {
				di++
			}
		}
		u4_putc(int(bp04[i]))
	}
}

func C_E408() {
	for b := int(Party.f_1d8) - 1; b >= 0; b-- {
		c := &Party.chara[b]
		if c._stat != 'D' {
			c._stat = 'G'
			c._HP[0] = c._HP[1]
		}
	}
}

// when asked 'health'
func C_E442() {
	u4_puts("\n\n\n\n\n\nHe says: I am well, thank ye.\n\nHe asks: Art thou well?\x12\x12\b")
	switch AskY_N() {
	case 'Y':
		u4_puts("\nHe says: That is good.\n")
	case 'N':
		u4_puts("\nHe says: Let me heal thy wounds!\n")
		// heal party
		sound(10, 10)
		Gra_09()
		sound(9, 0xc0)
		Gra_09()
		C_E408()
		dspl_Stats()
	}
}

func C_E498(bp04 *uint16) {
	v := int(int16(*bp04)) + U4_RND1(7) + 1
	*bp04 = uint16(u4_min(v, 50))
}

func C_E4C3() {
	for b := 0; b < int(Party.f_1d8); b++ {
		c := &Party.chara[b]
		loc_C := uint16(100)
		for loc_A := uint16(100); loc_A <= c._XP; loc_A <<= 1 {
			loc_C += 100
		}
		if c._HP[1] < loc_C {
			c._HP[1] = loc_C
			c._HP[0] = loc_C
			c._stat = 'G'
			C_E498(&c._str)
			C_E498(&c._dex)
			C_E498(&c._int)
			Gra_CR()
			u4_puts(cstr(c._name[:]))
			u4_puts("\nThou art now Level ")
			u4_putl(int32(loc_C/100), 1, '0')
			Gra_CR()
			Gra_09()
			sound(9, 0xc0)
			Gra_09()
			dspl_Stats()
		}
	}
	u4_puts("What would thou ask of me?\n")
}

// conversation with Lord british
func C_E59B() {
	var bp_12 [15]byte
	name0 := func() string { return cstr(Party.chara[0]._name[:]) }
	if Party.f_1e4 == 0 {
		// -- first time --
		Party.f_1e4 = 1
		u4_puts("\n\n\nLord British rises and says: At long last!\n")
		u4_puts(name0())
		u4_putc(',')
		u4_puts(" thou hast come!  We have waited such a long, long time...\n\n")
		C_E1FC()
		u4_puts("\n\nLord British sits and says: A new age is upon Britannia. The great evil Lords are gone but our people lack direction and purpose in their lives...\n\n\nA champion of virtue is called for. Thou may be this champion, but only time shall tell.  I will aid thee any way that I can!\n")
		u4_puts("How may I help thee?\n")
	} else {
		// -- second time and after --
		if Party.chara[0]._stat == 'D' {
			// resurection
			Party.chara[0]._stat = 'G'
			u4_puts(name0())
			u4_puts(", Thou shalt live again!\n")
			// heal party
			sound(10, 20)
			Gra_09()
			sound(9, 0xc0)
			Gra_09()
			C_E408()
			dspl_Stats()
		}
		u4_puts("\n\n\nLord British says:  Welcome ")
		u4_puts(name0())
		if Party.f_1d8 >= 3 {
			u4_puts(" and thy worthy Adventurers!\n")
		} else if Party.f_1d8 == 2 {
			u4_puts(" and thee also ")
			u4_puts(cstr(Party.chara[1]._name[:]))
			u4_puts("!\n")
		} else {
			Gra_CR()
		}
		C_E4C3()
	}
	for {
		u4_gets(bp_12[:], 15)
		Gra_CR()
		bp_02 := 0
		if bp_12[0] != 0 {
			bp_02 = C_E37E(bp_12[:])
		}
		if bp_02 == 0 { // "bye"
			break
		}
		switch bp_02 {
		case 2:
			C_E442()
		case 1:
			C_E21E()
		case -1:
			u4_puts("\nHe says: I cannot help thee with that.\n")
		default:
			C_E3D2(D_7028[bp_02-3])
		}
		u4_puts("\nWhat else?\n")
	}
	u4_puts("Lord British says: Fare thee well my friend")
	if Party.f_1d8 > 1 {
		u4_putc('s')
	}
	u4_puts("!\n")
}
