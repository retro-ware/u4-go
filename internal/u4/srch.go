package u4

// Search: the hidden things of the world and the towns, and the orbs, fountains and stones
// of the dungeons.

// The rewards of a find.
const (
	findKarma      = 5   // honor for every find
	questItemXP    = 400 // the bell, horn, wheel, skull, book, candle and mystic gear
	stoneXP        = 200
	runeXP         = 100
	mysticCount    = 8 // mystic robes or swords found: one for every member
	reagentsFound  = 2 // at least, and up to seven more
	maxReagents    = 99
	findMoveWindow = 16 // herbs grow back after this many moves
)

func sayNothingHere() { printText("Nothing Here!\n") }

// foundSomething tells of a find, honours the finder and remembers when it was.
func foundSomething() {
	printText("You find...\n")
	raiseKarma(&party.karmaHonor, findKarma)
	party.lastFound = uint16(foundWindow())
}

// foundWindow is the stretch of moves the party is in; one find of herbs per stretch.
func foundWindow() uint32 { return party.moves / findMoveWindow % 16 * findMoveWindow }

// gatherReagent adds a few of a reagent, dropping what does not fit.
func gatherReagent(r int) {
	party.reagents[r] += uint16(randomBits(7) + reagentsFound)
	if party.reagents[r] > maxReagents {
		party.reagents[r] = maxReagents
		printText("Dropped some!\n")
		sound(soundBeep)
	}
}

// moonsUp says whether either moon is up (not in its new phase).
func moonsUp() bool { return party.trammel|party.felucca != 0 }

// findHerb: the herbs grow only under the new moons, and not twice in a stretch of moves.
func findHerb(reagent int, text string) {
	if moonsUp() || foundWindow() == uint32(party.lastFound) {
		sayNothingHere()
		return
	}
	foundSomething()
	printText(text)
	gatherReagent(reagent)
}

const (
	reagentNightshade = 6
	reagentMandrake   = 7
)

func findMandrake()   { findHerb(reagentMandrake, "Mandrake Root!\n") }
func findNightshade() { findHerb(reagentNightshade, "Nightshade!\n") }

// findItem gives the quest item once.
func findItem(item int, text string) {
	if hasBit(party.items, item) {
		sayNothingHere()
		return
	}
	setBit(&party.items, item)
	foundSomething()
	printText(text)
	addXP(0, questItemXP)
}

func findBell()   { findItem(itemBell, "The Bell of Courage!\n") }
func findHorn()   { findItem(itemHorn, "A Silver Horn!\n") }
func findWheel()  { findItem(itemWheel, "The Wheel from the H.M.S. Cape!\n") }
func findBook()   { findItem(itemBook, "The Book of Truth!\n") }
func findCandle() { findItem(itemCandle, "The Candle of Love!\n") }

// findSkull: the skull lies only under the new moons, and is gone once destroyed.
func findSkull() {
	if hasBit(party.items, itemSkull) || moonsUp() || hasBit(party.items, itemSkullDestroyed) {
		sayNothingHere()
		return
	}
	setBit(&party.items, itemSkull)
	foundSomething()
	printText("The Skull of Mondain the Wizard!\n")
	addXP(0, questItemXP)
}

// The stones found outdoors: the black under the new moons, the white any time.
const (
	stoneWhite = virtueSpirituality
	stoneBlack = virtueHumility
)

func findBlackStone() {
	if hasBit(party.stones, stoneBlack) || moonsUp() {
		sayNothingHere()
		return
	}
	setBit(&party.stones, stoneBlack)
	foundSomething()
	printText("The Black Stone!\n")
	addXP(0, stoneXP)
}

func findWhiteStone() {
	if hasBit(party.stones, stoneWhite) {
		sayNothingHere()
		return
	}
	setBit(&party.stones, stoneWhite)
	foundSomething()
	printText("The White Stone!\n")
	addXP(0, stoneXP)
}

// useTelescope shows a town's map, chosen by its letter.
func useTelescope() {
	printText("You see a knob on the Telescope marked A-P\n")
	town := askLetter("You Select:"+answerRoom(2), 'A', 'P')
	if town < 0 {
		return
	}
	town -= 'A'
	Load(townFiles[town], area.tiles[:])
	showTownMap(false)
	Load(townFiles[party.loc-1], area.tiles[:])
}

// anyKarma is nonzero while any virtue is unfinished (karma 0 is a virtue raised to an
// avatar's).
func anyKarma() uint16 {
	return party.karmaHonesty | party.karmaCompassion | party.karmaValor | party.karmaJustice |
		party.karmaSacrifice | party.karmaHonor | party.karmaSpirituality | party.karmaHumility
}

// The mystic gear is found only by an avatar of all eight virtues, and only once.
func findMysticArmour() {
	if party.armors[armorMysticRobe]|anyKarma() != 0 {
		sayNothingHere()
		return
	}
	party.armors[armorMysticRobe] = mysticCount
	foundSomething()
	printText("Mystic Armour!\n")
	addXP(0, questItemXP)
}

func findMysticWeapons() {
	if party.weapons[weaponMysticSword]|anyKarma() != 0 {
		sayNothingHere()
		return
	}
	party.weapons[weaponMysticSword] = mysticCount
	foundSomething()
	printText("Mystic Weapons!\n")
	addXP(0, questItemXP)
}

// runeTowns are where each virtue's rune lies.
var runeTowns = [8]uint16{
	virtueHonesty:      locMoonglow,
	virtueCompassion:   locBritain,
	virtueValor:        locJhelom,
	virtueJustice:      locYew,
	virtueSacrifice:    locMinoc,
	virtueHonor:        locTrinsic,
	virtueSpirituality: locCastleBritannia,
	virtueHumility:     locPaws,
}

// findRune gives the rune of the town the party is in.
func findRune() {
	virtue := virtueHumility
	for virtue > virtueHonesty && runeTowns[virtue] != party.loc {
		virtue--
	}
	if hasBit(party.runes, virtue) {
		sayNothingHere()
		return
	}
	setBit(&party.runes, virtue)
	foundSomething()
	printText("The rune of ")
	printText(virtueNames[virtue])
	printText("!\n")
	addXP(0, runeXP)
}

// hiddenThing is something found by searching at a place.
type hiddenThing struct {
	loc  uint16
	x, y uint8
	find func()
}

var hiddenThings = []hiddenThing{
	{0, 182, 54, findMandrake},
	{0, 100, 165, findMandrake},
	{0, 46, 149, findNightshade},
	{0, 205, 44, findNightshade},
	{0, 176, 208, findBell},
	{0, 45, 173, findHorn},
	{0, 96, 215, findWheel},
	{0, 197, 245, findSkull},
	{0, 224, 133, findBlackStone},
	{0, 64, 80, findWhiteStone},
	{locLycaeum, 6, 6, findBook},
	{locCove, 22, 1, findCandle},
	{locLycaeum, 22, 3, useTelescope},
	{locEmpathAbbey, 22, 4, findMysticArmour},
	{locSerpentsHold, 8, 15, findMysticWeapons},
	{locMoonglow, 8, 6, findRune},
	{locBritain, 25, 1, findRune},
	{locJhelom, 30, 30, findRune},
	{locYew, 13, 6, findRune},
	{locMinoc, 28, 30, findRune},
	{locTrinsic, 2, 29, findRune},
	{locCastleBritannia, 17, 8, findRune},
	{locPaws, 29, 29, findRune},
}

func cmdSearch() {
	printText("Search...\n")
	if gameMode <= modeTown && party.flyingOrLight != 0 {
		sayDriftOnly()
		return
	}
	for _, h := range hiddenThings {
		if h.loc == party.loc && h.x == party.x && h.y == party.y {
			h.find()
			return
		}
	}
	sayNothingHere()
}

// ---- Searching in a dungeon ----

func sayNoEffect() { printText("\nHmmm--No Effect!\n") }

// shockMember flashes the member's line and hurts them by 100 points a level.
func shockMember(member, levels int) {
	invertStatusLine(member)
	sound(soundMemberHit)
	sound(soundMemberHit)
	sound(soundMemberHit)
	invertStatusLine(member)
	damageMember(member, 100*levels)
}

// raiseStatBy raises a stat, up to maxStat (the stat read as signed, as the original did).
func raiseStatBy(stat *uint16, amount int) {
	*stat += uint16(amount)
	if int16(*stat) > maxStat {
		*stat = maxStat
	}
}

// Each dungeon's orb: how hard it shocks (in levels) and which stats it raises.
type orb struct {
	shock                          int
	strength, dexterity, intellect bool
}

const orbBonus = 5

var orbs = [8]orb{
	locDeceit - locFirstDungeon:   {2, false, false, true},
	locDespise - locFirstDungeon:  {2, false, true, false},
	locDestard - locFirstDungeon:  {2, true, false, false},
	locWrong - locFirstDungeon:    {4, false, true, true},
	locCovetous - locFirstDungeon: {4, true, true, false},
	locShame - locFirstDungeon:    {4, true, false, true},
	locHythloth - locFirstDungeon: {6, true, true, true},
	locAbyss - locFirstDungeon:    {6, false, false, false},
}

// askWhoTouches asks for a conscious member; -1 if none.
func askWhoTouches(question string) int {
	member := askMember(question)
	if member < 0 {
		return -1
	}
	if !isConscious(member) {
		printText("\nDisabled!\n")
		return -1
	}
	return member
}

func touchOrb() {
	printText("\nYou find a Magical Ball...\n")
	member := askWhoTouches("Who touches?" + answerRoom(2))
	if member < 0 {
		return
	}
	*area.dungeonRoom(int(party.z), int(party.y), int(party.x)) = dngCorridor
	o := orbs[int(party.loc)-locFirstDungeon]
	shockMember(member, o.shock)
	m := &party.members[member]
	if o.strength {
		raiseStatBy(&m.strength, orbBonus)
		printText("Strength + 5\n")
	}
	if o.dexterity {
		raiseStatBy(&m.dexterity, orbBonus)
		printText("Dexterity + 5\n")
	}
	if o.intellect {
		raiseStatBy(&m.intelligence, orbBonus)
		printText("Intelligence + 5\n")
	}
}

// The fountains (the detail of a dngFountain room).
const (
	fountainPlain = iota
	fountainHealing
	fountainAcid
	fountainCure
	fountainPoison
)

func drinkFountain() {
	printText("You find a Fountain.\n")
	member := askWhoTouches("Who drinks?" + answerRoom(2))
	if member < 0 {
		return
	}
	m := &party.members[member]
	switch dngDetail(hereTile) {
	case fountainPlain:
		sayNoEffect()
	case fountainHealing:
		if m.hp == m.maxHP {
			sayNoEffect()
		} else {
			printText("\nAhh-Refreshing!\n")
			m.hp = m.maxHP
		}
	case fountainAcid:
		printText("\nBleck--Nasty!\n")
		shockMember(member, 1)
	case fountainCure:
		if m.status != statusPoisoned {
			sayNoEffect()
		} else {
			m.status = statusGood
			printText("\nHmm--Delicious!\n")
		}
	case fountainPoison:
		if m.status == statusPoisoned {
			sayNoEffect()
		} else {
			m.status = statusPoisoned
			printText("\nArgh-Choke-Gasp!\n")
			shockMember(member, 1)
		}
	}
}

// findDungeonStone: each dungeon but Hythloth and the Abyss hides its virtue's stone on an
// altar.
func findDungeonStone() {
	stone := int(party.loc) - locFirstDungeon
	if party.loc == locHythloth || party.loc == locAbyss || hasBit(party.stones, stone) {
		printText("\nYou find Nothing!\n")
		return
	}
	setBit(&party.stones, stone)
	printText("\nYou find the ")
	printText(colorNames[stone])
	printText(" stone!\n")
	raiseKarma(&party.karmaHonor, findKarma)
	addXP(0, stoneXP)
}

func cmdSearchDungeon() {
	printText("Search...\n")
	switch dngKind(hereTile) {
	case dngOrb:
		touchOrb()
	case dngFountain:
		drinkFountain()
	case dngAltar:
		findDungeonStone()
	default:
		printText("\nYou find Nothing!\n")
	}
}
