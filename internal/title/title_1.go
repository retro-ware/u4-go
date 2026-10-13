package title

// Port of TITLE_1.C: character creation. The name and sex, the introduction story with its
// pictures, the gypsy's seven questions on the abacus, the new party (PARTY.NEW with the
// player swapped in as the class of the last chosen virtue), and writing PARTY.SAV and
// MONSTERS.SAV.

var (
	lastVirtue   int       // the virtue chosen last: the player's class
	sex          int       // 'M' or 'F'
	strength     uint8     // the player's attributes, from the virtues chosen
	dexterity    uint8     //
	intelligence uint8     //
	storyPic     []byte    // the introduction's current picture
	question     int       // the gypsy's question, 0..6
	playerName   [20]uint8 // zero-terminated
	karma        [8]uint8
)

// The sex glyphs of the font.
const (
	glyphMale   = 11
	glyphFemale = 12
)

const asciiEscape = 27

// putString prints s at the text cursor; a newline goes to the next row's start.
func putString(s string) { /*C_261D*/
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			textRow++
			textCol = 0
		} else {
			putChar(s[i])
		}
	}
}

// readName reads up to size-1 characters into buf (zero-terminated), then removes leading and
// trailing blanks.
func readName(buf []uint8, size int) { /*C_2656*/
	n := 0
	for {
		key := readKey()
		switch key {
		case keyBackspace, keyCtrlBackspace, keyLeft:
			if n == 0 {
				beep()
			} else {
				putChar('\b')
				n--
				buf[n] = ' '
			}
		case keyEnter:
			buf[n] = 0
		default:
			c := key % 256
			if n == size-1 || c < ' ' || c >= 128 {
				beep()
			} else {
				buf[n] = uint8(c)
				putChar(uint8(c))
				n++
			}
		}
		if key == keyEnter {
			break
		}
	}
	start := 0
	for buf[start] == ' ' {
		start++
	}
	if buf[start] != 0 {
		n = 0
		for buf[start] != 0 {
			buf[n] = buf[start]
			n++
			start++
		}
	} else {
		n = start
	}
	// The original reads on before the buffer when the name is all blanks: the byte there is
	// always 0, which stops it.
	for {
		buf[n] = 0
		n--
		if n < 0 || buf[n] != ' ' {
			break
		}
	}
}

// showPage shows a page of text and waits for a key.
func showPage(text string) { /*C_271D*/
	clearTextArea()
	textCol, textRow = 0, textTop
	putString(text)
	readKey()
}

// The moongate in the tree picture: the gate's frames and the bare grass are kept below the
// picture (columns 0 and 3 from y 152); it stands at column 9, y 68, 24 rows high.
const (
	gateSpriteCol  = 0
	grassSpriteCol = 3
	spriteY        = 152
	gateCol        = 9
	gateY          = 68
	gateRows       = 24
	gateCols       = 3
)

// moongateRises shows the moongate rising from the ground.
func moongateRises() { /*C_273E*/
	for rows := 1; rows < gateRows; rows++ {
		copyPic(storyPic, gateSpriteCol, spriteY, gateCols, rows, gateCol, gateY+gateRows-rows)
		copyPic(storyPic, gateCol, gateY, gateCols, gateRows-rows, gateCol, gateY)
		c64Gate(27-rows*27/23, 27, rows%2) // the C64's rises in 27 rows
		pace(moongateStep)
	}
	copyPic(storyPic, gateSpriteCol, spriteY, gateCols, gateRows, gateCol, gateY)
	c64Gate(0, 27, 0)
}

// moongateSinks shows the moongate sinking into the ground.
func moongateSinks() { /*C_27E0*/
	for rows := 1; rows < gateRows; rows++ {
		copyPic(storyPic, grassSpriteCol, spriteY, gateCols, rows, gateCol, gateY)
		copyPic(storyPic, gateSpriteCol, spriteY, gateCols, gateRows-rows, gateCol, gateY+rows)
		c64Gate(rows*27/23, 27, rows%2) // the C64's sinks from the top
		pace(moongateStep)
	}
	copyPic(storyPic, grassSpriteCol, spriteY, gateCols, gateRows, gateCol, gateY)
	c64Gate(0, 0, 0)
}

// showStoryPic shows the story picture above the text area.
func showStoryPic() { copyPic(storyPic, 0, 0, textCols, textAreaY, 0, 0) }

// tellStory tells the introduction, the pictures changing along the way.
func tellStory() { /*C_2883*/
	storyPic = make([]byte, picSize)
	loadPicture("tree.ega", storyPic)
	clearScreen()
	showStoryPic()
	for page, text := range storyPages {
		clearTextArea()
		textCol, textRow = 0, textTop
		putString(text)
		switch page {
		case 3: // the portal rises
			moongateRises()
		case 5: // it sinks
			moongateSinks()
			loadPicture("portal.ega", storyPic)
		case 6: // the circle of stones
			showStoryPic()
			loadPicture("tree.ega", storyPic)
		case 11: // back under the willow
			showStoryPic()
			loadPicture("outside.ega", storyPic)
		case 15: // the fair
			showStoryPic()
			loadPicture("inside.ega", storyPic)
		case 17: // inside it
			showStoryPic()
			loadPicture("wagon.ega", storyPic)
		case 20: // the wagon
			showStoryPic()
			loadPicture("gypsy.ega", storyPic)
		case 21: // the gypsy
			showStoryPic()
			loadPicture("abacus.ega", storyPic)
		case 23: // the abacus
			showStoryPic()
		}
		flushKeys()
		readKey()
	}
}

// The abacus in its picture: beads are 1 column by 12 rows below the picture (white at column
// 1, black at 3, from y 187); virtue v's beads go in column 16+v, each question's on its own
// line.
const (
	whiteBeadCol = 1
	blackBeadCol = 3
	beadY        = 187
	beadRows     = 12
	abacusCol    = 16
)

var beadLineY = [...]int{24, 40, 54, 68, 82, 98, 112}

// drawBead puts a bead on the abacus: white for the chosen virtue, black for the discarded.
func drawBead(chosen bool, virtue, line int) { /*C_2B2A*/
	src := blackBeadCol
	if chosen {
		src = whiteBeadCol
	}
	copyPic(storyPic, src, beadY, 1, beadRows, abacusCol+virtue, beadLineY[line])
	c64Bead(chosen, virtue, line)
}

// cardPictures hold the virtue cards, two to a picture (the even virtue left, the odd right).
var cardPictures = [...]string{"HONCOM.EGA", "VALJUS.EGA", "SACHONOR.EGA", "SPIRHUM.EGA"}

// The cards: 12 columns by 124 rows, from y 12; at column 1 or 27 in the pictures and on the
// table.
const (
	cardCols      = 12
	cardRows      = 124
	cardY         = 12
	leftCardCol   = 1
	rightCardCol  = 27
	leftSide      = 0
	rightSide     = 1
	startingScore = 15
	startingKarma = 50
	karmaPerVote  = 5
)

// drawCard lays virtue's card on the table, on side leftSide or rightSide.
func drawCard(virtue, side int) { /*C_2B6D*/
	pic := make([]byte, picSize)
	loadPicture(cardPictures[virtue/2], pic)
	src, dst := rightCardCol, rightCardCol
	if virtue%2 == 0 {
		src = leftCardCol
	}
	if side == leftSide {
		dst = leftCardCol
	}
	copyPic(pic, src, cardY, cardCols, cardRows, dst, cardY)
	showCard(virtue, side, dst)
}

// What each virtue adds to the attributes when chosen.
var (
	strengthBonus     = [...]uint8{0, 0, 3, 0, 1, 1, 1, 0}
	dexterityBonus    = [...]uint8{0, 3, 0, 1, 1, 0, 1, 0}
	intelligenceBonus = [...]uint8{3, 0, 0, 1, 0, 1, 1, 0}
)

// firstDilemma is where each virtue's questions against the later virtues start in dilemmas.
var firstDilemma = [...]int{0, 7, 13, 18, 22, 25, 27}

// dilemma is the question setting virtue a against a later virtue b.
func dilemma(a, b int) string { return dilemmas[firstDilemma[a]+b-a-1] }

// The cards' state during the casting.
const (
	cardFree      = 0
	cardChosen    = 1
	cardDiscarded = 255
)

// consultGypsy runs the casting: seven pairs of virtues, the winners meeting again.
func consultGypsy() { /*C_2C12*/
	var cards [8]uint8
	intelligence, dexterity, strength = startingScore, startingScore, startingScore
	question = 0
	for v := range karma {
		karma[v] = startingKarma
	}
	for {
		if question == 4 || question == 6 { // the winners come back
			for v := range cards {
				if cards[v] != cardDiscarded {
					cards[v] = cardFree
				}
			}
		}
		clearTextArea()
		textCol, textRow = 0, textTop
		var a, b uint8
		for {
			a = randomByte() % 8
			if cards[a] == cardFree {
				break
			}
		}
		for {
			b = randomByte() % 8
			if cards[b] == cardFree && b != a {
				break
			}
		}
		if b < a {
			a, b = b, a
		}
		switch question {
		case 0:
			putString(firstCards)
		case 6:
			putString(lastCards)
		default:
			putString(moreCards)
		}
		putString(cardsAre)
		drawCard(int(a), leftSide)
		putString(virtueNames[a])
		putString(" and ")
		drawCard(int(b), rightSide)
		putString(virtueNames[b])
		putString(".  She says\n\"Consider this:\"")
		flushKeys()
		readKey()
		clearTextArea()
		textCol, textRow = 0, textTop
		putString(dilemma(int(a), int(b)))
		findAnswers() // the port's addition: the answers can be picked in the text
		drawAnswers()
		var key int
		for {
			askThen(askGypsy)
			key = readKey()
			askThen(askNone)
			switch key {
			case keyLeft, keyRight, keyUp, keyDown:
				answerSel ^= 1
				drawAnswers()
				continue
			case keyEnter:
				if answerFound {
					key = []int{keyA, keyB}[answerSel]
				}
			}
			key = lowerKey(key)
			if key == keyA || key == keyB {
				break
			}
		}
		if key == keyB {
			a, b = b, a
		}
		lastVirtue = int(a)
		cards[a] = cardChosen
		karma[a] += karmaPerVote
		strength += strengthBonus[a]
		dexterity += dexterityBonus[a]
		intelligence += intelligenceBonus[a]
		drawBead(true, int(a), question)
		cards[b] = cardDiscarded
		drawBead(false, int(b), question)
		question++
		if question == 7 {
			break
		}
	}
	showPage(pathChosen)
	showPage(awakening)
}

// Where each class starts: the town of its virtue (Moonglow, Britain, Jhelom, Yew, Minoc,
// Trinsic, Skara Brae, Magincia).
var (
	startX = [...]uint8{231, 83, 35, 59, 158, 105, 23, 186}
	startY = [...]uint8{136, 105, 221, 44, 21, 183, 129, 171}
)

// createParty makes the new party from PARTY.NEW: the player, of the last chosen virtue's
// class, swapped to the front.
func createParty() { /*C_2E04*/
	LoadRec("PARTY.NEW", &party)
	party.x = startX[lastVirtue]
	party.y = startY[lastVirtue]
	party.size = 1
	for v := range karma {
		party.karma[v] = uint16(karma[v])
	}
	party.members[lastVirtue], party.members[0] = party.members[0], party.members[lastVirtue]
	player := &party.members[0]
	for i := 0; ; i++ {
		player.name[i] = playerName[i]
		if playerName[i] == 0 {
			break
		}
	}
	player.strength = uint16(strength)
	player.intelligence = uint16(intelligence)
	player.dexterity = uint16(dexterity)
	if sex == 'M' {
		player.sex = glyphMale
	} else {
		player.sex = glyphFemale
	}
	people = peopleTable{}
}

// saveNewGame writes PARTY.SAV and MONSTERS.SAV. The port has no disks: the original asked
// for the PLAYER disk first and the PROGRAM disk after.
func saveNewGame() { /*C_2FB8*/
	Save("PARTY.SAV", &party)
	Save("MONSTERS.SAV", &people)
}

// initiateNewGame creates a character: true when the new party is saved and the game starts.
func initiateNewGame() bool { /*C_3030*/
	clearMenuArea()
	printAt(16, 4, "By what name shalt thou be known")
	printAt(17, 4, "in this world and time?")
	textCol, textRow = 12, 19
	askThen(askName)
	readName(playerName[:], 12)
	askThen(askNone)
	if playerName[0] == 0 {
		animPic, shapes = nil, nil
		return false
	}
	clearMenuArea()
	printAt(17, 4, "Art thou Male or Female? ")
	askThen(askSex)
	sex = upperChar(readKey() % 256)
	for sex != 'M' && sex != 'F' {
		beep()
		sex = readKey() % 256
		// quirk: compared with the full key words, Enter and Space never match; only Escape
		// cancels
		if sex == asciiEscape || sex == keyEnter || sex == keySpace {
			askThen(askNone)
			animPic, shapes = nil, nil
			return false
		}
		sex = upperChar(sex)
	}
	askThen(askNone)
	putChar(uint8(sex))
	animPic, shapes = nil, nil
	monstersShown = false

	tellStory()
	consultGypsy()
	createParty()
	saveNewGame()
	storyPic = nil
	charset = nil
	return true
}

// dilemmas are the gypsy's questions, one for each pair of virtues in order: Honesty against
// each later virtue, then Compassion against each later one, and so on.
var dilemmas = [...]string{
	"Entrusted to deliver an uncounted purse\nof gold, thou dost meet a poor beggar.\nDost thou A) deliver the gold knowing\nthe Trust in thee was well placed, or\nB) show Compassion, giving the beggar a\ncoin, knowing it won't be missed?",
	"Thou has been prohibited by thy absent\nLord from joining thy friends in a close\npitched battle. Dost thou A) refrain, so\nthou may Honesty claim obedience, or\nB) show Valor, and aid thy comrades,\nknowing thou may deny it later?",
	"A merchant owes thy friend money, now\nlong past due.  Thou dost see the same\nmerchant drop a purse of gold. Dost thou\nA) Honestly return the purse intact, or\nB) Justly give thy friend a portion of\nthe gold first?",
	"Thee and thy friend are valiant but\npenniless warriors.  Thou both go out to\nslay a mighty dragon.  Thy friend thinks\nhe slew it; thee did.  When asked, dost\nthou A) Truthfully claim the gold, or\nB) Allow thy friend the large reward?",
	"Thou art sworn to protect thy Lord at\nany cost, yet thou know he hath\ncommitted a crime.  Authorities ask thee\nof the affair.  Dost thou A) break thine\noath by Honestly speaking, or B) uphold\nHonor by silently keeping thine oath?",
	"Thy friend seeks admittance to thy\nSpiritual order. Thou art asked to vouch\nfor his purity of Spirit, of which thou\nart unsure.  Dost thou A) Honestly\nexpress thy doubt, or B) Vouch for him\nhoping for his Spiritual improvement?",
	"Thy Lord mistakenly believes he slew a\ndragon.  Thou hast proof that thy lance\nfelled the beast.  When asked, dost thou\nA) Honestly claim the kill and the\nprize, or B) Humbly permit thy Lord his\nbelief?",
	"Thou dost manage to disarm thy mortal\nenemy in a duel.  He is at thy mercy.\nDost thou A) show Compassion by\npermitting him to yield, or B) slay him\nas expected of a Valiant duelist?",
	"After 20 years thou hast found the\nslayer of thy best friends.  The villain\nproves to be a man who provides the sole\nsupport for a young girl.  Dost thou\nA) spare him in Compassion for the girl,\nor B) slay him in the name of Justice?",
	"Thee and thy friends have been routed\nand ordered to retreat.  In defiance of\nthy orders, dost thou A) stop in\nCompassion to aid a wounded companion,\nor B) Sacrifice thyself to slow the\npursuing enemy, so others may escape?",
	"Thou art sworn to uphold a Lord who\nparticipates in the forbidden torture of\nprisoners.  Each night their cries of\npain reach thee.  Dost thou A) show\nCompassion by reporting the deeds, or\nB) Honor thy oath and ignore the deeds?",
	"Thou hast been taught to preserve all\nlife as sacred. A man lies fatally stung\nby a venomous serpent.  He pleads for a\nmerciful death.  Dost thou A) show\nCompassion and end his pain, or B) heed\nthy Spiritual beliefs and refuse?",
	"As one of the King's Guard, thy Captain\nhas asked that one amongst you visit\na hospital to cheer the children with\ntales of thy valiant deeds.  Dost thou\nA) Show thy Compassion and play the\nbraggart, or B) Humbly let another go?",
	"Thou hast been sent to secure a needed\ntreaty with a distant Lord.  Thy host is\nagreeable to the proposal but insults\nthy country at dinner.  Dost thou\nA) Valiantly bear the slurs, or\nB) Justly rise and demand an apology?",
	"A mighty knight accosts thee and demands\nthy food.  Dost thou A) Valiantly refuse\nand engage the knight, or B) Sacrifice\nthy food unto the hungry knight?",
	"During battle thou art ordered to guard\nthy commander's empty tent.  The battle\ngoes poorly and thou dost yearn to aid\nthy fellows.  Dost thou A) Valiantly\nenter the battle to aid thy companions,\nor B) Honor thy post as guard?",
	"A local bully pushes for a fight.  Dost\nthou A) Valiantly trounce the rogue, or\nB) Decline, knowing in thy Spirit that\nno lasting good will come of it?",
	"Although a teacher of music, thou art\na skillful wrestler.  Thou hast been\nasked to fight in a local championship.\nDost thou A) accept the invitation and\nValiantly fight to win, or B) Humbly\ndecline knowing thou art sure to win?",
	"During a pitched battle, thou dost see\na fellow desert his post, endangering\nmany.  As he flees, he is set upon by\nseveral enemies. Dost thou A) Justly let\nhim fight alone, or B) Risk Sacrificing\nthine own life to aid him?",
	"Thou hast sworn to do thy Lord's\nbidding in all.  He covets a piece of\nland and orders the owner removed.  Dost\nthou A) serve Justice refusing to act,\nthus being disgraced, or B) Honor thine\noath and unfairly evict the landowner?",
	"Thou dost believe that virtue resides\nin all people.  Thou dost see a rogue\nsteal from thy Lord.  Dost thou A) call\nhim to Justice, or B) personally try to\nsway him back to the Spiritual path of\ngood?",
	"Unwitnessed, thou hast slain a great\ndragon in self defense.  A poor warrior\nclaims the offered reward.  Dost thou\nA) Justly step forward to claim the\nreward, or B) Humbly go about life,\nsecure in thy self esteem?",
	"Thou art a bounty hunter sworn to return\nan alleged murder.  After his capture\nthou believest him to be innocent.  Dost\nthou A) Sacrifice thy sizable bounty\nfor thy belief, or B) Honor thy oath to\nreturn him as thou hast promised?",
	"Thou hast spent thy life in charitable\nand righteous work.  Thine uncle the\ninnkeeper lies ill and asks you to take\nover his tavern.  Dost thou A) Sacrifice\nthy life of purity to aid thy kin, or\nB) decline & follow thy Spirit's call?",
	"Thou art an elderly, wealthy eccentric.\nThy end is near. Dost thou A) donate all\nthy wealth to feed hundreds of starving\nchildren, and receive public adulation,\nor B) Humbly live out thy life, willing\nthy fortune to thy heirs?",
	"In thy youth thou pledged to marry thy\nsweetheart.  Now thou art on a sacred\nquest in distant lands.  Thy sweetheart\nasks thee to keep thy vow.  Dost thou\nA) Honor thy pledge to wed, or B) follow\nthy Spiritual crusade?",
	"Thou art at a crossroads in thy life.\nDost thou A) Choose the Honorable life\nof a Paladin, striving for Truth and\nCourage, or B) Choose the Humble life\nof a Shepherd, and a world of simplicity\nand peace?",
	"Thy parents wish thee to become an\napprentice. Two positions are available.\nDost thou A) Become an acolyte in the\nSpiritual order, or B) Become an\nassistant to a humble village cobbler?",
}

// storyPages is the introduction, a page a key press.
var storyPages = [...]string{
	"  The day is warm, yet there is a\ncooling breeze.  The latest in a series\nof personal crises seems insurmountable.\nYou are being pulled apart in all\ndirections.",
	"Yet this afternoon walk in the country-\nside slowly brings relaxation to your\nharried mind.  The soil and strain of\nmodern high-tech living begins to wash\noff in layers. That willow tree near the\nstream looks comfortable and inviting.",
	"The buzz of dragonflies and the whisper\nof the willow's swaying branches bring\na deep peace.  Searching inward for\ntranquility and happiness, you close\nyour eyes.",
	"A high-pitched cascading sound like\ncrystal wind-chimes impinges on your\nfloating awareness.  As you open your\neyes, you see a shimmering blueness rise\nfrom the ground.  The sound seems to be\nemanating from this glowing portal.",
	"It is difficult to look at the blueness.\nLight seems to bend and distort around\nit, while the sound waves become so\nintense, they appear to become visible.",
	"The portal hangs there for a moment;\nthen, with the rush of an imploding\nvacuum, it sinks into the ground.\nSomething remains suspended in mid-air\nfor a moment before falling to earth\nwith a heavy thud.",
	"Somewhat shaken by this vision, you\nrise to your feet to investigate.  A\ncrude circle of stones surrounds the\nspot where the portal appeared. There is\nsomething glinting in the grass.",
	"You pick up an amulet shaped like a\ncross with a loop at the top.  It is an\nAnkh, the sacred symbol of life and\nrebirth.  But this could not have made\nthe thud, so you look again and find\na large book wrapped in thick cloth!",
	"With trembling hands you unwrap the\nbook.  Behold, the cloth is a map, and\nwithin lies not one book, but two.  The\nmap is of a land strange to you, and the\nstyle speaks of ancient cartography.",
	"The script on the cover of the first\nbook is arcane but readable.  The title\nis:\n        The History of Britannia\n               as told by\n            Kyle the Younger",
	"The other book is disturbing to look at.\nIts small cover appears to be fashioned\nout of some sort of leathery hide, but\nfrom what creature is uncertain.  The\nreddish-black skin radiates an intense\naura suggestive of ancient power.",
	"The tongue of the title is beyond your\nken.  You dare not open the book and\ndisturb whatever sleeps within.  You\ndecide to peruse the History.  Settling\nback under the willow tree, you open the\nbook.",
	"\n\n    (You read the Book of History)",
	"\n\n(No, really! Read the Book of History!)",
	"Closing the book, you again pick up\nthe Ankh.  As you hold it, you begin to\nhear a hauntingly familiar, lute-like\nsound wafting over a nearby hill.  Still\nclutching the strange artifacts, you\nrise unbidden and climb the slope.",
	"In the valley below you see what appears\nto be a fair.  It seems strange that you\ncame that way earlier and noticed\nnothing.  As you mull this over, your\nfeet carry you down towards the site.",
	"This is no ordinary travelling\ncarnival, but a Renaissance Fair.  The\npennants on the tent tops blow briskly\nin the late afternoon breeze.",
	"The ticket taker at the RenFair's gate\nstarts to ask you for money, but upon\nspotting your Ankh says, \"Welcome,\nfriend.  Enter in peace and find your\npath.\"",
	"The music continues to pull you\nforward amongst the merchants and\nvendors.  Glimpses of fabulous treasures\ncan be seen in some of the shadowy\nbooths.",
	"These people are very happy.  They seem\nto glow with an inner light.  Some look\nup as you pass and smile, but you cannot\nstop - the music compels you to move\nonward through the crowd.",
	"Through the gathering dusk you see a\nsecluded gypsy wagon sitting off in the\nwoods.  The music seems to emanate from\nthe wagon.  As you draw near, a woman's\nvoice weaves into the music, saying:\n\"You may approach, O seeker.\"",
	"You enter to find an old gypsy sitting\nin a small curtained room.  She wears an\nAnkh around her neck. In front of her is\na round table covered in deep green\nvelvet.  The room smells so heavily of\nincense that you feel dizzy.",
	"Seeing the Ankh, the ancient gypsy\nsmiles and warns you never to part with\nit.  \"We have been waiting such a long\ntime, but at last you have come.  Sit\nhere and I shall read the path of your\nfuture.\"",
	"Upon the table she places a curious\nwooden object like an abacus but without\nbeads.  In her hands she holds eight\nunusual cards.  \"Let us begin the\ncasting.\"",
}

// The gypsy's words around the cards.
const (
	firstCards = "The gypsy places the first two cards\n"
	moreCards  = "The gypsy places two more of the cards\n"
	lastCards  = "The gypsy places the last two cards\n"
	cardsAre   = "upon the table.  They are the cards of\n"
)

var virtueNames = [...]string{
	"Honesty",
	"Compassion",
	"Valor",
	"Justice",
	"Sacrifice",
	"Honor",
	"Spirituality",
	"Humility",
}

// The end of the casting.
const (
	pathChosen = "With the final choice, the incense\nswells up around you.  The gypsy speaks\nas if from a great distance, her voice\ngrowing fainter with each word:\n\"So be it!  Thy path is chosen!\""
	awakening  = "There is a moment of intense, wrenching\nvertigo.  As you open your eyes, a\nvoice whispers within your mind, \"Seek\nthe counsel of thy sovereign.\"  After a\nmoment, the spinning subsides, and you\nopen your eyes to..."
)
