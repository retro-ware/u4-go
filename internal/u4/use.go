package u4

import "strings"

// Use: the stones, keys, bell, book, candle, horn, wheel and skull.

func sayUseNoEffect() { printText("Hmm...No effect!\n") }
func sayNoneOwned()   { printText("None owned!\n") }

// The altar rooms of the dungeons, by the principle they stand for (altarRoom).
const (
	noAltar      = -1
	altarTruth   = 0
	altarLove    = 1
	altarCourage = 2
)

// altarStones are the four stones each altar takes: the virtues of its principle.
var altarStones = [3]uint8{
	altarTruth:   1<<virtueSpirituality | 1<<virtueHonor | 1<<virtueJustice | 1<<virtueHonesty,
	altarLove:    1<<virtueSpirituality | 1<<virtueSacrifice | 1<<virtueJustice | 1<<virtueCompassion,
	altarCourage: 1<<virtueSpirituality | 1<<virtueHonor | 1<<virtueSacrifice | 1<<virtueValor,
}

// altarKeys is the part of the three part key each altar gives.
var altarKeys = [3]int{altarTruth: itemKeyTruth, altarLove: itemKeyLove, altarCourage: itemKeyCourage}

// The altar stands in the middle of its room.
const altarX, altarY = 5, 5

// The longest word the Use prompts take.
const useWordLen = 10

// askStoneColor reads a colour and returns the stone's virtue, -1 for none, and what was
// typed.
func askStoneColor() (int, string) {
	getsWords = stoneColors()
	word := readWord(useWordLen)
	newLine()
	stone := len(colorNames) - 1
	for stone >= 0 && !strings.EqualFold(word, colorNames[stone]) {
		stone--
	}
	return stone, word
}

// useStonesOnAltar: the four right stones on an altar give its part of the key.
func useStonesOnAltar() {
	m := int(activeMember)
	if battlefield.memberX[m] != altarX || battlefield.memberY[m] != altarY || altarRoom == noAltar {
		sayUseNoEffect()
		return
	}
	printText("\nThere are holes for 4 stones. What colors:\n")
	used := uint8(0)
	for hole := 'A'; hole <= 'D'; hole++ {
		printChar(int(hole))
		printChar(':')
		stone, word := askStoneColor()
		if stone == -1 || !hasBit(party.stones, stone) {
			if word != "" {
				sayNoneOwned()
			}
			return
		}
		if hasBit(used, stone) {
			printText("Already used!\n")
			return
		}
		setBit(&used, stone)
	}
	if altarStones[altarRoom] != used {
		sayUseNoEffect()
		return
	}
	key := altarKeys[altarRoom]
	if hasBit(party.items, key) {
		sayUseNoEffect()
		return
	}
	setBit(&party.items, key)
	printText("Thou doth find one third of the Three Part Key!\n")
}

// abyssQuestions are what the altar on each level of the Abyss asks after.
var abyssQuestions = [8]string{
	"Truth",
	"Love",
	"Courage",
	"Truth and Love",
	"Love and Courage",
	"Courage and Truth",
	"Truth, Love and Courage",
	"\nA voice rings out:  What Virtue exists independently of Truth, Love and Courage",
}

// The Abyss's altars take answers this long.
const abyssAnswerLen = 12

func useStones() {
	if party.stones == 0 {
		sayNoneOwned()
		return
	}
	if gameMode != modeCombatRoom && gameMode != modeDungeon {
		printText("No place to Use them!\n")
	}
	if gameMode != modeDungeon {
		useStonesOnAltar()
		return
	}
	if hereTile != dngAltar || party.loc != locAbyss {
		sayUseNoEffect()
		return
	}
	level := int(party.z)
	lastLevel := len(abyssQuestions) - 1
	if level != lastLevel {
		printText("\nAs thou doth approach, a voice rings out: What virtue dost stem from ")
	}
	printText(abyssQuestions[level])
	printText("?\n\n")
	answer := readWord(abyssAnswerLen)
	newLine()
	if !strings.EqualFold(answer, virtueNames[level]) {
		if answer != "" {
			sayUseNoEffect()
		}
		return
	}
	printText("\nThe Voice says: Use thy Stone.\n\nColor:\n")
	stone, word := askStoneColor()
	if stone == -1 {
		if word != "" {
			sayUseNoEffect()
		}
		return
	}
	if !hasBit(party.stones, stone) {
		printText("\nYou have none!\n\n")
		return
	}
	if level != stone {
		sayUseNoEffect()
		return
	}
	if level == lastLevel {
		enterCodexChamber() // the Codex chamber
		return
	}
	*area.dungeonRoom(level, int(party.y), int(party.x)) = dngLadderDown
	printText("\nThe altar changes before thyne eyes!\n")
}

func useKeys() {
	if hasBit(party.items, itemKeyTruth) || hasBit(party.items, itemKeyLove) || hasBit(party.items, itemKeyCourage) {
		printText("No place to Use them!\n")
	} else {
		sayNoneOwned()
	}
}

// atAbyss says whether the party stands at the entrance of the Abyss.
func atAbyss() bool { return party.loc == 0 && party.x == abyssX && party.y == abyssY }

// The bell, book and candle open the Abyss: used there, in that order.
func useBell() {
	if !hasBit(party.items, itemBell) {
		sayNoneOwned()
		return
	}
	if !atAbyss() {
		sayUseNoEffect()
		return
	}
	setBit(&party.items, itemBellUsed)
	printText("The Bell rings on and on!\n")
}

func useBook() {
	if !hasBit(party.items, itemBook) {
		sayNoneOwned()
		return
	}
	if !atAbyss() || !hasBit(party.items, itemBellUsed) {
		sayUseNoEffect()
		return
	}
	setBit(&party.items, itemBookUsed)
	printText("The words resonate with the ringing!\n")
}

func useCandle() {
	if !hasBit(party.items, itemCandle) {
		sayNoneOwned()
		return
	}
	if !atAbyss() || !hasBit(party.items, itemBookUsed) {
		sayUseNoEffect()
		return
	}
	setBit(&party.items, itemCandleUsed)
	printText("As you light the Candle the Earth Trembles!\n")
	sound(soundMonsterHit)
	shakeScreen()
}

// hornTurns is how long the horn keeps the monsters away.
const hornTurns = 10

func useHorn() {
	if !hasBit(party.items, itemHorn) {
		sayNoneOwned()
		return
	}
	if party.loc != 0 {
		sayUseNoEffect()
		return
	}
	printText("The Horn sounds an eerie tone!\n")
	activeSpell = effectHorn
	spellTurns = hornTurns
	showStats()
}

// The wheel mends a ship of hull exactly wheelHull to a perfect one.
const (
	wheelHull   = 50
	perfectHull = 99
)

func useWheel() {
	if !hasBit(party.items, itemWheel) {
		sayNoneOwned()
		return
	}
	if party.loc != 0 || party.tile > tileShipSouth || party.hull != wheelHull {
		sayUseNoEffect()
		return
	}
	printText("Once mounted, the Wheel glows with a blue light!\n")
	party.hull = perfectHull
	showStats()
}

// The skull: raised, it slays everyone in sight but Lord British, at a cost to every
// virtue; cast into the Abyss, it raises them all.
const (
	skullKarmaCost  = 5
	skullKarmaBonus = 10
	outdoorPeople   = 8 // the people slots in use on the world map
)

// slayAllBut clears every person of the list but Lord British.
func slayAllBut(tile, baseTile []uint8) {
	for i := len(tile) - 1; i >= 0; i-- {
		if tile[i] != tileLordBritish {
			baseTile[i] = 0
			tile[i] = 0
		}
	}
}

func useSkull() {
	if !hasBit(party.items, itemSkull) {
		sayNoneOwned()
		return
	}
	quake := func() {
		sound(soundMonsterHit)
		shakeScreen()
		invertMap()
		sound(soundMonsterHit)
		shakeScreen()
		invertMap()
		sound(soundMonsterHit)
		shakeScreen()
	}
	if atAbyss() {
		printText("\nYou cast the Skull of Mondain into the Abyss!\n")
		setBit(&party.items, itemSkullDestroyed)
		for _, k := range karmas {
			raiseKarma(k, skullKarmaBonus)
		}
		quake()
		clearBit(&party.items, itemSkull)
		return
	}
	printText("\nYou hold the evil Skull of Mondain the Wizard aloft....\n")
	quake()
	people := &area.people
	switch {
	case gameMode >= modeCombat:
		slayAllBut(people.tile[:], people.baseTile[:])
		slayAllBut(fighters.tile[:], fighters.baseTile[:])
	case party.loc == 0:
		slayAllBut(people.tile[:outdoorPeople], people.baseTile[:outdoorPeople])
	default:
		slayAllBut(people.tile[:], people.baseTile[:])
	}
	redraw()
	for _, k := range karmas {
		lowerKarma(k, skullKarmaCost)
	}
	showStats()
	clearBit(&party.items, itemSkull)
}

// usable are the words the Use command knows.
var usable = []struct {
	word string
	use  func()
}{
	{"stone", useStones},
	{"stones", useStones},
	{"bell", useBell},
	{"book", useBook},
	{"candle", useCandle},
	{"key", useKeys},
	{"keys", useKeys},
	{"horn", useHorn},
	{"wheel", useWheel},
	{"skull", useSkull},
}

func cmdUse() {
	printText("Use which item:\n")
	getsWords = usableItems()
	word := readWord(useWordLen)
	newLine()
	for _, u := range usable {
		if strings.EqualFold(word, u.word) {
			u.use()
			return
		}
	}
	printText("Not a Usable item!\n")
}
