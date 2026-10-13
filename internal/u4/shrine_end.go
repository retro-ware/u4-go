package u4

// The shrines (meditation, visions and partial avatarhood) and the Chamber of the Codex at
// the bottom of the Abyss, with the ending.

import (
	"strings"

	"github.com/memmaker/u4-go/internal/host"
	"github.com/memmaker/u4-go/internal/music"
)

// ---- The shrines ----

// mantras are the words chanted at each virtue's shrine.
var mantras = [8]string{"ahm", "mu", "ra", "beh", "cah", "summ", "om", "lum"}

// visions are what each shrine shows after one, two or three cycles of meditation.
var visions = [8][3]string{
	{
		"Take not the gold of others found in towns and castles for yours it is not!\n",
		"Cheat not the merchants and peddlers for tis an evil thing to do!\n",
		"Second, read the Book of Truth at the entrance to the Great Stygian Abyss!\n",
	},
	{
		"Kill not the non-evil beasts of the land, and do not attack the fair people!\n",
		"Give of thy purse to those who beg and thy deed shall not be forgotten!\n",
		"Third, light the Candle of Love at the entrance to the Great Stygian Abyss!\n",
	},
	{
		"Victories scored over evil creatures help to build a valorous soul!\n",
		"To flee from battle with less than grievous wounds often shows a coward!\n",
		"First, ring the Bell of Courage at the entrance to the Great Stygian Abyss!\n",
	},
	{
		"To take the gold of others is injustice not soon forgotten. Take only thy due!\n",
		"Attack not a peaceful citizen for that action deserves strict punishment!\n",
		"Kill not a non-evil beast for they deserve not death, even if in hunger they attack thee!",
	},
	{
		"To give thy last gold piece unto the needy shows good measure of self-sacrifice!\n",
		"For thee to flee and leave thy companions is a self-serving action to be avoided!\n",
		"To give of thy life's blood so that others may live is a virtue of great praise!\n",
	},
	{
		"Take not the gold of others for this shall bring dishonor upon thee!\n",
		"To strike first a non-evil being is by no means an honorable deed!\n",
		"Seek ye to solve the many Quests before thee, and honor shall be a reward!\n",
	},
	{
		"Seek ye to know thyself.  Visit the seer often for he can see into thy inner being!\n",
		"Meditation leads to enlightenment Seek ye all Wisdom and Knowledge!\n",
		"If thou dost seek the White Stone, search ye not under the ground, but in Serpent's Spine!\n",
	},
	{
		"Claim not to be that which thou art not.  Humble actions speak well of thee!\n",
		"Strive not to wield the Great Force of Evil for its power will overcome thee!\n",
		"If thou dost seek the Black Stone, search ye at the Time and Place of the Gate on the darkest of all nights!\n",
	},
}

// avatarhoodRunes are the EGA pictures of the letters I N F I N I T Y, one a virtue.
var avatarhoodRunes = [8]string{"rune_1", "rune_2", "rune_0", "rune_1", "rune_2", "rune_1", "rune_3", "rune_4"}

// showAvatarhoodRune shows the virtue's letter of INFINITY on the map.
func showAvatarhoodRune(virtue int) {
	gameMode = modeVision
	clearMap()
	if screen.IsVGA() { // the VGA patch has a picture a letter: rune_1..8
		loadPicture("rune_"+string(rune('1'+virtue))+".ega", 0, nil)
		return
	}
	loadPicture(avatarhoodRunes[virtue]+".ega", 0, nil)
}

// Meditation: up to maxCycles cycles of cycleSeconds, once every meditationMoves moves.
const (
	maxCycles       = 3
	cycleSeconds    = 16
	meditationMoves = 100
	mantraKarmaCost = 3
	cycleKarma      = 3 // spirituality for every cycle
	avatarhoodKarma = 99
	shrineWordLen   = 14
)

// leaveShrine puts the party back outdoors.
func leaveShrine() {
	gameMode = modeOutdoors
	party.loc = 0
}

func enterShrine() {
	virtue := int(party.loc) - locFirstShrine
	if !hasBit(party.runes, virtue) {
		printText("\nThou dost not bear the rune of entry!  A strange force keeps you out!\n")
		leaveShrine()
		return
	}
	Load("SHRINE.CON", battlefield.tiles[:])
	gameMode = modeShrine
	music.Play(music.Shrine)
	printText("\nYou enter the ancient shrine and sit before the altar...\nUpon what virtue dost thou meditate?\n")
	subject := readWord(shrineWordLen)
	newLine()
	printText("\nFor how many\n")
	cycles := askLetter("Cycles (0-3)?"+answerRoom(2), '0', '0'+maxCycles)
	if cycles < 0 {
		leaveShrine()
		return
	}
	cycles -= '0'
	if cycles == 0 || !strings.EqualFold(subject, virtueNames[virtue]) {
		printText("\nThou art unable to focus thy thoughts on this subject!\n")
		leaveShrine()
		return
	}
	if party.moves/meditationMoves == uint32(party.lastMeditation) {
		printText("\nThy mind is still weary from thy last Meditation!\n")
		leaveShrine()
		return
	}
	party.lastMeditation = uint16(party.moves / meditationMoves)
	printText("Begin Meditation\n")
	for range cycles {
		for range cycleSeconds {
			waitSeconds(1, false)
			printChar('.')
		}
		flushKeys()
		printText("\nMantra: ")
		mantra := readWord(shrineWordLen)
		newLine()
		if !strings.EqualFold(mantra, mantras[virtue]) {
			printText("\nThou art not able to focus thy thoughts with that Mantra!\n")
			lowerKarma(&party.karmaSpirituality, mantraKarmaCost)
			leaveShrine()
			return
		}
	}
	if cycles != maxCycles || *karmas[virtue] != avatarhoodKarma {
		printText("\nThy thoughts are pure. Thou art granted a vision!\n")
		raiseKarma(&party.karmaSpirituality, cycles*cycleKarma)
		flushKeys()
		readKey()
		newLine()
		printText(visions[virtue][cycles-1])
	} else {
		printText("\nThou hast achieved partial Avatarhood in the Virtue of\n")
		printText(virtueNames[virtue])
		invertMap()
		sound(soundSweep, 255)
		invertMap()
		newLine()
		*karmas[virtue] = 0 // a virtue raised to an avatar's
		showStats()
		flushKeys()
		readKey()
		printText("\n\nThou art granted a vision!\n")
		showAvatarhoodRune(virtue)
		flushKeys()
	}
	readKey()
	newLine()
	leaveShrine()
}

// ---- The Chamber of the Codex ----

// codexQuestions are the voice's questions: the eight virtues, then the three principles.
var codexQuestions = [11]string{
	"What dost thou possess if all may rely upon your every word?",
	"What quality compels one to share in the journeys of others?",
	"What answers when great deeds are called for?",
	"What should be the same for Lord and Serf alike?",
	"What is loath to place the self above aught else?",
	"What shirks no duty?",
	"What, in knowing the true self, knows all?",
	"What is that which Serfs are born with but Nobles must strive to obtain?",
	"If all else is imaginary, this is real...",
	"What plunges to the depths, while soaring on the heights?",
	"What turns not away from any peril?",
}

// principles are the answers to the last three questions.
var principles = [3]string{"truth", "love", "courage"}

// virtuePictures are the pictures shown for each virtue answered.
var virtuePictures = [8]string{"honesty", "compassn", "valor", "justice", "sacrific", "honor", "spirit", "humility"}

// pauseForReading waits 15 seconds or a key.
func pauseForReading() {
	waitSeconds(15, true)
	if keyWaiting() {
		readKey()
	}
}

// returnPlaces are where a party turned back from the Codex is put on the surface: one for
// each of the eleven questions failed, one for the last question, and the Abyss's entrance
// for passage denied.
var returnPlaces = [13]struct{ x, y uint8 }{
	{231, 136}, {83, 105}, {35, 221}, {59, 44}, {158, 21}, {105, 183}, {23, 129}, {186, 172},
	{216, 106}, {29, 48}, {145, 243},
	{89, 106},
	{233, 233},
}

const (
	returnAfterLast = 11
	returnDenied    = 12
	codexPartySize  = 8
	codexAnswerLen  = 14
	codexTries      = 3
)

// partyBeforeCodex is the party's size before it entered the chamber alone.
var partyBeforeCodex uint16

// returnToSurface puts the party back on the surface, and back to the commands.
func returnToSurface(place int) {
	waitSeconds(5, false)
	newLine()
	party.x = returnPlaces[place].x
	party.y = returnPlaces[place].y
	party.size = partyBeforeCodex
	party.loc = 0
	gameMode = modeOutdoors
	spellTurns = 0
	hitTile = 0
	activeSpell = effectNone
	lineOfSight = 1
	gallopStep = 0
	party.flyingOrLight = 0
	centerMap()
	showStats()
	flushKeys()
	backToCommands()
}

func denyPassage() {
	printText("\nPassage is not granted.\n")
	returnToSurface(returnDenied)
}

// victory: the Codex is revealed, and the game ends.
func victory() {
	waitSeconds(2, false)
	for range 4 {
		shakeScreen()
	}
	waitSeconds(3, false)
	music.Play(music.Fanfare)
	showInfinity()
	waitSeconds(3, false)
	for _, s := range []string{
		"\n\nThe boundless knowledge of the Codex of Ultimate Wisdom is revealed unto thee.",
		"\n\nThe voice says: Thou hast proven thyself to be truly good in nature.",
		"\n\nThou must know that thy quest to become an Avatar is the endless quest of a lifetime.",
		"\n\nAvatarhood is a living gift.  It must always and forever be nurtured to flourish.",
		"\n\nFor if thou dost stray from the paths of virtue, thy way may be lost forever.",
		"\n\nReturn now unto thine own world. Live there as an example to thy people, as our memory of thy gallant deeds serves us.",
	} {
		printText(s)
		pauseForReading()
	}
	clearMap()
	printText("\n\nAs the sound of the voice trails off, darkness seems to rise around you. There is a moment of intense, wrenching vertigo.")
	pauseForReading()
	loadPicture("stoncrcl.ega", 0, nil)
	for _, s := range []string{
		"\n\nYou open your eyes to a familiar circle of stones.  You wonder of your recent adventures.",
		"\n\nIt seems a time and place very distant.  You wonder if it really happened. Then you realize that in your hand you hold The Ankh.",
		"\n\nYou walk away from the circle, knowing that you can always return from whence you came, since you now know the secret of the gates.",
	} {
		printText(s)
		pauseForReading()
	}
	printText("\n\nCONGRATULATIONS!\n   Thou hast\n   completed\n   ULTIMA IV\n  Quest of the\n    AVATAR\n  in ")
	printNumber(int32(party.moves), 0, '0')
	printText("\n turns! Report\n thy feat unto\nLord British at\nOrigin Systems!")
	for { // the game is over
		host.Yield()
	}
}

// askCodex asks the question up to three times; true for the right answer.
func askCodex(question, answer string) bool {
	for try := range codexTries {
		if try > 0 {
			waitSeconds(1, false)
			printText("\nThy thoughts are not pure.\nI ask again.\n")
			waitSeconds(2, false)
		}
		printText(question)
		newLine()
		newLine()
		said := readWord(codexAnswerLen)
		newLine()
		if strings.EqualFold(said, answer) {
			return true
		}
	}
	return false
}

// enterCodexChamber: the last altar of the Abyss leads the party's leader alone before the
// Codex, to answer the voice's questions.
func enterCodexChamber() {
	partyBeforeCodex = party.size
	party.size = 1
	gameMode = modeVision
	clearStatus()
	showStats()
	clearMap()
	music.Play(music.Stop)
	printText("\n\n\nThere is a sudden darkness, and you find yourself alone in an empty chamber.\n")
	waitSeconds(4, false)
	loadPicture("key7.ega", 0, nil)
	if !hasBit(party.items, itemKeyCourage) || !hasBit(party.items, itemKeyTruth) || !hasBit(party.items, itemKeyLove) {
		printText("\nThou dost not have the Key of Three Parts.\n")
		returnToSurface(returnDenied)
	}
	printText("\nYou use your Key of Three Parts.\n")
	waitSeconds(3, false)
	printText("\nA voice rings out:\n")
	if !askCodex("\"What is the Word of Passage?\"", "veramocor") {
		denyPassage()
	}
	if partyBeforeCodex != codexPartySize {
		printText("\nThou hast not proved thy leadership in all eight virtues.\n")
		waitSeconds(8, false)
		denyPassage()
	}
	for v := len(karmas) - 1; v >= 0; v-- {
		if *karmas[v] != 0 {
			printText("\nThou art not ready.\n")
			denyPassage()
		}
	}
	printText("\nPassage is granted.\n")
	waitSeconds(5, false)
	clearMap()
	for q, question := range codexQuestions {
		waitSeconds(2, false)
		printText("\n\nThe voice asks:\n\n")
		waitSeconds(2, false)
		answer, picture := "", ""
		if q < len(virtueNames) {
			answer, picture = virtueNames[q], virtuePictures[q]
		} else {
			answer = principles[q-len(virtueNames)]
			picture = answer
		}
		if !askCodex(question, answer) {
			waitSeconds(1, false)
			printText("\nThy quest is not yet complete.\n")
			returnToSurface(q)
		}
		loadPicture(picture+".ega", 0, nil)
		if q == len(virtueNames)-1 {
			waitSeconds(3, false)
			printText("\nThou art well versed in the virtues of the Avatar.\n")
			waitSeconds(5, false)
		}
	}
	printText("\n\nThe floor rumbles beneath your feet.\n")
	shakeScreen()
	shakeScreen()
	waitSeconds(5, false)
	printText("\nAbove the din, the voice asks:\n\nIf all eight virtues of the Avatar combine into and are derived from the Three Principles of Truth, Love and Courage...")
	pauseForReading()
	printText("\n\nThen what is the one thing which encompasses and is the whole of all undeniable Truth, unending Love, and unyielding Courage?\n\n")
	said := readWord(codexAnswerLen)
	newLine()
	for try := 1; !strings.EqualFold(said, "infinity"); try++ {
		if try == codexTries {
			waitSeconds(1, false)
			printText("\nThou dost not know the true nature of the Universe.\n")
			returnToSurface(returnAfterLast)
		}
		printText("\nThy thoughts are not pure.\nI ask again.\n")
		waitSeconds(3, false)
		printText("\n\nWhat is the one thing which encompasses and is the whole of all undeniable Truth, unending Love, and unyielding Courage?\n\n")
		said = readWord(codexAnswerLen)
		newLine()
	}
	victory()
}
