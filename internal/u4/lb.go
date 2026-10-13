package u4

// Talking to Lord British.

// lordBritishWords are the words Lord British answers to (four letters are enough): bye,
// help and health, then the topics of lordBritishAnswers.
var lordBritishWords = []string{
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

// lordBritishAnswers are Lord British's answers on the topics from "name" on.
var lordBritishAnswers = []string{
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

// pauseForKey waits ten seconds, or until a key.
func pauseForKey() {
	flushKeys()
	waitSeconds(10, true)
	if keyWaiting() {
		readKey()
	}
}

// putsWait prints each text, pausing between them.
func putsWait(texts ...string) {
	for i, text := range texts {
		printText(text)
		if i < len(texts)-1 {
			pauseForKey()
		}
	}
}

// lordBritishHelp is Lord British's advice for how far the Avatar has come.
func lordBritishHelp() {
	printText("He says: ")
	// (The original ANDs the karmas' bits: true when they share a set bit.)
	karmaBits := party.karmaHonesty & party.karmaCompassion & party.karmaValor & party.karmaJustice &
		party.karmaSacrifice & party.karmaHonor & party.karmaSpirituality & party.karmaHumility
	hasAll := func(items ...int) bool {
		for _, item := range items {
			if !hasBit(party.items, item) {
				return false
			}
		}
		return true
	}
	switch {
	case party.moves < 1000:
		putsWait("To survive in this hostile land thou must first know thyself! Seek ye to master thy weapons and thy magical ability!\n",
			"\nTake great care in these thy first travels in Britannia.\n",
			"\nUntil thou dost well know thyself, travel not far from the safety of the townes!\n")
	case party.size == 1:
		putsWait("Travel not the open lands alone. There are many worthy people in the diverse townes whom it would be wise to ask to Join thee!\n",
			"\nBuild thy party unto eight travellers, for only a true leader can win the Quest!\n")
	case party.runes == 0:
		putsWait("Learn ye the paths of virtue. Seek to gain entry unto the eight shrines!\n",
			"\nFind ye the Runes, needed for entry into each shrine, and learn each chant or \"Mantra\" used to focus thy meditations.\n",
			"\nWithin the Shrines thou shalt learn of the deeds which show thy inner virtue or vice!\n",
			"\nChoose thy path wisely for all thy deeds of good and evil are remembered and can return to hinder thee!\n")
	case karmaBits != 0:
		putsWait("Visit the Seer Hawkwind often and use his wisdom to help thee prove thy virtue.\n",
			"\nWhen thou art ready, Hawkwind will advise thee to seek the Elevation unto partial Avatarhood in a virtue.\n",
			"\nSeek ye to become a partial Avatar in all eight virtues, for only then shalt thou be ready to seek the codex!\n")
	case party.stones == 0:
		putsWait("Go ye now into the depths of the dungeons. Therein recover the 8 colored stones from the altar pedestals in the halls of the dungeons.\n",
			"\nFind the uses of these stones for they can help thee in the Abyss!\n")
	case anyKarma() != 0:
		printText("Thou art doing very well indeed on the path to Avatarhood! Strive ye to achieve the Elevation in all eight virtues!\n")
	case !hasAll(itemBell, itemBook, itemCandle):
		printText("Find ye the Bell, Book and Candle!  With these three things, one may enter the Great Stygian Abyss!\n")
	case !hasAll(itemKeyCourage, itemKeyLove, itemKeyTruth):
		putsWait("Before thou dost enter the Abyss thou shalt need the Key of Three Parts, and the Word of Passage.\n",
			"\nThen might thou enter the Chamber of the Codex of Ultimate Wisdom!\n")
	default:
		putsWait("Thou dost now seem ready to make the final journey into the dark Abyss! Go only with a party of eight!\n",
			"\nGood Luck, and may the powers of good watch over thee on this thy most perilous endeavor!\n",
			"\nThe hearts and souls of all Britannia go with thee now. Take care, my friend.\n")
	}
}

// lordBritishWord is the index of the word asked in lordBritishWords, or -1.
func lordBritishWord(asked string) int {
	for i, word := range lordBritishWords {
		if sameWord(asked, word, 4) {
			return i
		}
	}
	return -1
}

// printPaged prints text, pausing after every page of lines.
func printPaged(text string) {
	lines := 0
	for i := 0; i < len(text); i++ {
		if text[i] == '\n' {
			if lines == textPageLines {
				pauseForKey()
				lines = 0
			} else {
				lines++
			}
		}
		printChar(int(text[i]))
	}
}

// healParty heals the living party members fully, and cures them.
func healParty() {
	for i := int(party.size) - 1; i >= 0; i-- {
		member := &party.members[i]
		if member.status != statusDead {
			member.status = statusGood
			member.hp = member.maxHP
		}
	}
}

// askAfterHealth: Lord British asks whether the Avatar is well, and heals the party if not.
func askAfterHealth() {
	printText("\n\n\n\n\n\nHe says: I am well, thank ye.\n\nHe asks: Art thou well?" + answerRoom(2))
	switch askYesNo() {
	case 'Y':
		printText("\nHe says: That is good.\n")
	case 'N':
		printText("\nHe says: Let me heal thy wounds!\n")
		sound(soundMagic, 10)
		invertMap()
		sound(soundSweep, 192)
		invertMap()
		healParty()
		showStats()
	}
}

// maxStat is as far as a level gained raises an attribute.
const maxStat = 50

// raiseStat raises an attribute by 1 to 8, up to maxStat.
func raiseStat(stat *uint16) {
	*stat = uint16(min(int(int16(*stat))+randomBits(7)+1, maxStat))
}

// levelUp raises the members whose experience earned it a level: one for every doubling of
// their experience from 100, each worth 100 hit points, and a raise in every attribute.
func levelUp() {
	for i := 0; i < int(party.size); i++ {
		member := &party.members[i]
		maxHP := uint16(100)
		for xp := uint16(100); xp <= member.xp; xp *= 2 {
			maxHP += 100
		}
		if member.maxHP >= maxHP {
			continue
		}
		member.maxHP, member.hp = maxHP, maxHP
		member.status = statusGood
		raiseStat(&member.strength)
		raiseStat(&member.dexterity)
		raiseStat(&member.intelligence)
		newLine()
		printText(cstr(member.name[:]))
		printText("\nThou art now Level ")
		printNumber(int32(maxHP/100), 1, '0')
		newLine()
		invertMap()
		sound(soundSweep, 192)
		invertMap()
		showStats()
	}
	printText("What would thou ask of me?\n")
}

// talkToLordBritish is an audience with Lord British: his welcome (and the Avatar raised
// from the dead, or levels gained), then his answers until "bye".
func talkToLordBritish() {
	avatar := func() string { return cstr(party.members[0].name[:]) }
	if party.metLordBritish == 0 {
		party.metLordBritish = 1
		printText("\n\n\nLord British rises and says: At long last!\n")
		printText(avatar())
		printChar(',')
		printText(" thou hast come!  We have waited such a long, long time...\n\n")
		pauseForKey()
		printText("\n\nLord British sits and says: A new age is upon Britannia. The great evil Lords are gone but our people lack direction and purpose in their lives...\n\n\nA champion of virtue is called for. Thou may be this champion, but only time shall tell.  I will aid thee any way that I can!\n")
		printText("How may I help thee?\n")
	} else {
		if party.members[0].status == statusDead {
			party.members[0].status = statusGood
			printText(avatar())
			printText(", Thou shalt live again!\n")
			sound(soundMagic, 20)
			invertMap()
			sound(soundSweep, 192)
			invertMap()
			healParty()
			showStats()
		}
		printText("\n\n\nLord British says:  Welcome ")
		printText(avatar())
		switch {
		case party.size >= 3:
			printText(" and thy worthy Adventurers!\n")
		case party.size == 2:
			printText(" and thee also ")
			printText(cstr(party.members[1].name[:]))
			printText("!\n")
		default:
			newLine()
		}
		levelUp()
	}
	const (
		wordBye    = 0
		wordHelp   = 1
		wordHealth = 2
		firstTopic = 3
	)
	for {
		asked := readWord(14)
		newLine()
		word := wordBye // nothing typed
		if asked != "" {
			word = lordBritishWord(asked)
		}
		if word == wordBye {
			break
		}
		switch word {
		case wordHealth:
			askAfterHealth()
		case wordHelp:
			lordBritishHelp()
		case -1:
			printText("\nHe says: I cannot help thee with that.\n")
		default:
			printPaged(lordBritishAnswers[word-firstTopic])
		}
		printText("\nWhat else?\n")
	}
	printText("Lord British says: Fare thee well my friend")
	if party.size > 1 {
		printChar('s')
	}
	printText("!\n")
}
