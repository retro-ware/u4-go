package u4

import "github.com/memmaker/u4-go/internal/music"

// The small commands: hole up & camp, quit (save), new order, volume, mix, locate,
// ready and wear; and the random events of the ground the party stands on.

// ---- Hole up & camp ----

// ambushers are the monsters that may surprise a sleeping camp.
var ambushers = [8]uint8{tileOrc, tileSkeleton, tileRogue, tilePython, tileInsects, tileSlime, tileTroll, tileWisp}

// The camp is ambushed one time in campAmbushOdds, and rests at most once every
// campInterval moves.
const (
	campAmbushOdds = 8
	campInterval   = 100
	campSeconds    = 10
)

// ambush: monsters fall on the camp, and the sleeping wake to fight.
func ambush() {
	printText("Ambushed!\n")
	foeX = party.x
	foeY = party.y
	foeTile = ambushers[randomBits(len(ambushers)-1)]
	encounterTile = foeTile
	for m := int(party.size) - 1; m >= 0; m-- {
		if party.members[m].status == statusGood {
			party.members[m].status = statusAsleep
		}
	}
	showStats()
	runCombat()
}

func cmdHoleUp() {
	printText("Hole up & Camp\n")
	if gameMode != modeDungeon && party.loc != 0 {
		sayNotHere()
		return
	}
	if party.tile != tileOnFoot {
		sayOnlyOnFoot()
		return
	}
	name := "CAMP.CON"
	if gameMode == modeDungeon {
		name = "CAMP.DNG"
	}
	LoadRec(name, &battlefield)
	music.Play(music.Stop)
	fighters.tile = [16]uint8{}
	fighters.baseTile = [16]uint8{}
	for m := int(party.size) - 1; m >= 0; m-- {
		tile := uint8(0)
		if isAlive(m) {
			tile = tileAsleep
		}
		fighters.memberTile[m] = tile
		memberFrame[m] = tile
	}
	showStats()
	printText("Resting...\n")
	modeBeforeCombat = int16(gameMode)
	gameMode = modeCombatCamp
	waitSeconds(campSeconds, false)
	if randomBits(campAmbushOdds-1) == 0 {
		ambush()
	} else if party.moves/campInterval == uint32(party.lastCamp) {
		printText("No effect.\n")
	} else {
		party.lastCamp = uint16(party.moves / campInterval)
		for m := int(party.size) - 1; m >= 0; m-- {
			if !isAlive(m) {
				continue
			}
			if party.members[m].status == statusAsleep {
				party.members[m].status = statusGood
			}
			healMember(m, randomBits(119))
			healMember(m, 99)
			party.members[m].mp = maxMP
		}
		regainMP()
		printText("Players Healed!\n")
	}
	showStats()
	gameMode = uint16(modeBeforeCombat)
	if gameMode == modeDungeon {
		forceRender = 1
	}
}

// ---- Quit, new order and volume ----

// inDungeon says whether the party is in one of the dungeons (or the Abyss).
func inDungeon() bool { return party.loc >= locFirstDungeon && party.loc <= locAbyss }

func cmdQuit() {
	// The replies say plainly what happened (the original: "Quit & Save...", "Not Here!").
	printText("Save game: ")
	printNumber(int32(party.moves), 1, '0')
	printText(" moves\n")
	if party.loc != 0 && !inDungeon() {
		printText("Can't save in\ntowns. Not saved!\n")
		return
	}
	dungeon := inDungeon()
	Save("PARTY.SAV", &party)
	Save("MONSTERS.SAV", &area.people)
	if dungeon {
		SaveBytes("DNGMAP.SAV", area.tiles[:dungeonMapSize])
	}
	printText("Game saved.\n")
}

// sayMustLead tells that the first member leads and cannot be moved.
func sayMustLead() {
	printText(cstr(party.members[0].name[:]))
	printText(", You must lead!\n")
}

// swapMembers exchanges two members' places in the party.
func swapMembers(a, b int) {
	party.members[a], party.members[b] = party.members[b], party.members[a]
}

func cmdNewOrder() {
	printText("New Order!\n")
	first := askMember("Exchange #" + answerRoom(2))
	if first < 0 {
		return
	}
	if first == 0 {
		sayMustLead()
		return
	}
	second := askMember("    with #" + answerRoom(2))
	if second < 0 {
		return
	}
	if second == 0 {
		sayMustLead()
		return
	}
	if first == second {
		sayWhat()
		return
	}
	swapMembers(first, second)
	clearStatus()
	showStats()
}

func cmdVolume() {
	printText("Volume O")
	soundOn ^= 1
	if soundOn != 0 {
		printText("n!\n")
	} else {
		printText("ff!\n")
	}
}

// ---- Random events ----

// The odds (one in so many) of the ground's random harms.
const (
	trollOdds        = 8 // bridge trolls on a bridge
	fallAsleepOdds   = 4 // a member falls asleep in a sleep field
	getPoisonedOdds  = 8 // a member is poisoned in a swamp or poison field
	burnDamageBase   = 16
	burnDamageSpread = 32
)

// The fields of the dungeon (the detail of a dngField room).
const (
	fieldPoison = iota
	fieldEnergy
	fieldFire
	fieldSleep
)

// The traps of the dungeon (the detail of a dngTrap room): winds, falling rocks (details
// 1 to 7), and pits (8 on).
const (
	trapWinds = 0
	trapPit   = 8
)

// sleepField may put each conscious member to sleep.
func sleepField() {
	for m := int(party.size) - 1; m >= 0; m-- {
		if isConscious(m) && randomBits(fallAsleepOdds-1) == 0 {
			party.members[m].status = statusAsleep
			flashMember(m)
			showStats()
		}
	}
}

// poisonField may poison each healthy member.
func poisonField() {
	for m := int(party.size) - 1; m >= 0; m-- {
		if party.members[m].status == statusGood && randomBits(getPoisonedOdds-1) == 0 {
			party.members[m].status = statusPoisoned
			flashMember(m)
			showStats()
		}
	}
}

// groundEvent: the ground the party (or the active member in combat) stands on harms it.
func groundEvent() {
	if gameMode == modeOutdoors && hereTile == tileBridge && randomBits(trollOdds-1) == 0 {
		printText("\nBridge Trolls!\n")
		encounterTile = tileTroll
		foeX = party.x
		foeY = party.y
		foeGround = tileBridge
		beginCombat()
		return
	}
	if gameMode <= modeTown && party.flyingOrLight != 0 {
		return // the balloon floats above it
	}
	if gameMode >= modeCombat {
		groundStrikesMember(int(activeMember))
		return
	}
	if gameMode == modeDungeon {
		dungeonRoomEvent()
		return
	}
	switch hereTile {
	case tileSwamp, tilePoisonField:
		poisonField()
	case tileFireField, tileLava:
		damageParty()
	case tileSleepField:
		sleepField()
	}
}

// groundStrikesMember: the battlefield's ground under the member harms them.
func groundStrikesMember(m int) {
	if fighters.memberTile[m] == 0 {
		return
	}
	switch *combatTile(int(battlefield.memberY[m]), int(battlefield.memberX[m])) {
	case tileSwamp, tilePoisonField:
		if party.members[m].status == statusGood {
			party.members[m].status = statusPoisoned
			flashMember(m)
			showStats()
		}
	case tileSleepField:
		if isConscious(m) {
			party.members[m].status = statusAsleep
			fighters.memberTile[m] = tileAsleep
			flashMember(m)
			showStats()
		}
	case tileFireField, tileLava:
		flashMember(m)
		damageMember(m, randomBelow(burnDamageSpread)+burnDamageBase)
	}
}

// dungeonRoomEvent: the field or trap the party stands on in a dungeon.
func dungeonRoomEvent() {
	switch dngKind(hereTile) {
	case dngField:
		switch dngDetail(hereTile) & 3 {
		case fieldPoison:
			poisonField()
		case fieldFire:
			damageParty()
		case fieldSleep:
			sleepField()
		}
	case dngTrap:
		switch detail := dngDetail(hereTile); {
		case detail == trapWinds:
			printText("\nWinds!\n")
			party.flyingOrLight = 0 // they blow out the torch
		case detail < trapPit:
			printText("\nFalling Rocks!\n")
			damageParty()
		default:
			printText("\nPit!\n")
			damageParty()
		}
	}
}

// ---- Mix ----

// The reagents, as the bits of a recipe.
const (
	usesAsh = 128 >> iota
	usesGinseng
	usesGarlic
	usesSpiderSilk
	usesBloodMoss
	usesBlackPearl
	usesNightshade
	usesMandrake
)

// reagentBit is reagent r's bit in a recipe.
func reagentBit(r int) uint8 { return 128 >> r }

// recipes are the reagents that each spell, A to Z, is mixed from.
var recipes = [26]uint8{
	/*Awaken*/ usesGinseng | usesGarlic,
	/*Blink*/ usesSpiderSilk | usesBloodMoss,
	/*Cure*/ usesGinseng | usesGarlic,
	/*Dispel*/ usesAsh | usesGarlic | usesBlackPearl,
	/*Energy*/ usesAsh | usesSpiderSilk | usesBlackPearl,
	/*Fireball*/ usesAsh | usesBlackPearl,
	/*Gate*/ usesAsh | usesBlackPearl | usesMandrake,
	/*Heal*/ usesGinseng | usesSpiderSilk,
	/*Iceball*/ usesBlackPearl | usesMandrake,
	/*Jinx*/ usesBlackPearl | usesNightshade | usesMandrake,
	/*Kill*/ usesBlackPearl | usesNightshade,
	/*Light*/ usesAsh,
	/*Magic missile*/ usesAsh | usesBlackPearl,
	/*Negate*/ usesAsh | usesGarlic | usesMandrake,
	/*Open*/ usesAsh | usesBloodMoss,
	/*Protection*/ usesAsh | usesGinseng | usesGarlic,
	/*Quickness*/ usesAsh | usesGinseng | usesBloodMoss,
	/*Resurrect*/ usesAsh | usesGinseng | usesGarlic | usesSpiderSilk | usesBloodMoss | usesMandrake,
	/*Sleep*/ usesGinseng | usesSpiderSilk,
	/*Tremor*/ usesAsh | usesBloodMoss | usesMandrake,
	/*Undead*/ usesAsh | usesGarlic,
	/*View*/ usesNightshade | usesMandrake,
	/*Winds*/ usesAsh | usesBloodMoss,
	/*X-it*/ usesAsh | usesSpiderSilk | usesBloodMoss,
	/*Y-up*/ usesSpiderSilk | usesBloodMoss,
	/*Z-down*/ usesSpiderSilk | usesBloodMoss,
}

// maxMixtures is the most of one spell the party can carry.
const maxMixtures = 99

// The prompts' place at the bottom of the status area.
const (
	promptRow = 23
	promptCol = 24
)

func cmdMix() {
	for done := false; !done; {
		clearStatus()
		clearStatusTitle()
		showMixtures()
		textRow, textCol = promptRow, promptCol
		printText("Mix Reagents\n")
		spell := askLetter("For Spell:"+answerRoom(2), 'A', 'Z')
		if spell < 0 {
			break
		}
		spell -= 'A'
		printText(spellNames[spell])
		newLine()
		clearStatus()
		clearStatusTitle()
		var before [8]uint16
		for r := range before {
			before[r] = uint16(uint8(party.reagents[r])) // the original kept only the low byte
		}
		mixed := uint8(0)
		for {
			showReagents()
			textRow, textCol = promptRow, promptCol
			reagent := askLetter("Reagent:"+answerRoom(2), 'A', 'H')
			if reagent == -2 { // cancelled: the reagents go back
				party.reagents = before
				done = true
				break
			}
			if reagent == -1 { // mixed
				if mixed == 0 {
					printText("\nNothing mixed!\n")
				} else {
					printText("\nYou mix the Reagents, and...\n")
					if recipes[spell] == mixed {
						printText("Success!\n\n")
						party.mixtures[spell] = min(party.mixtures[spell]+1, maxMixtures)
					} else {
						printText("It Fizzles!\n\n")
					}
				}
				break
			}
			reagent -= 'A'
			if party.reagents[reagent] == 0 {
				sayNoneLeft()
				break
			}
			mixed |= reagentBit(reagent)
			party.reagents[reagent]--
		}
	}
	clearStatus()
	clearStatusTitle()
	showStats()
}

// ---- Locate ----

// cmdLocate reads the sextant: each coordinate as two letters, its sixteenths and its
// sixteenths' sixteenths.
func cmdLocate() {
	printText("Locate position with ")
	if party.sextants == 0 {
		sayWhat()
		return
	}
	printText("sextant\n Latitude: ")
	printCoordinate(party.y)
	printText("\"\nLongitude: ")
	printCoordinate(party.x)
	printChar('"')
	newLine()
}

// printCoordinate prints a world coordinate as degrees and minutes, A'A to P'P.
func printCoordinate(c uint8) {
	printChar(int(c/16) + 'A')
	printChar('\'')
	printChar(int(c%16) + 'A')
}

// ---- Ready and wear ----

// classBit is a class's bit in a mask of classes.
func classBit(class int) uint8 { return 128 >> class }

const allClasses = 255

// onlyClasses is the mask of the classes given.
func onlyClasses(classes ...int) uint8 {
	mask := uint8(0)
	for _, c := range classes {
		mask |= classBit(c)
	}
	return mask
}

// allClassesBut is the mask of every class but the ones given.
func allClassesBut(classes ...int) uint8 { return ^onlyClasses(classes...) }

// weaponClasses are the classes that may ready each weapon.
var weaponClasses = [16]uint8{
	weaponHands:       allClasses,
	weaponStaff:       allClasses,
	weaponDagger:      allClasses,
	weaponSling:       allClasses,
	weaponMace:        allClassesBut(classMage),
	weaponAxe:         allClassesBut(classMage, classDruid),
	weaponSword:       allClassesBut(classMage, classDruid),
	weaponBow:         allClassesBut(classMage, classShepherd),
	weaponCrossbow:    allClassesBut(classMage, classShepherd),
	weaponOil:         allClasses,
	weaponHalberd:     onlyClasses(classFighter, classTinker, classPaladin),
	weaponMagicAxe:    onlyClasses(classTinker, classPaladin),
	weaponMagicSword:  allClassesBut(classMage, classBard, classDruid, classShepherd),
	weaponMagicBow:    allClassesBut(classMage, classFighter, classShepherd),
	weaponMagicWand:   onlyClasses(classMage, classBard, classDruid),
	weaponMysticSword: allClasses,
}

// armorClasses are the classes that may wear each armour.
var armorClasses = [8]uint8{
	armorSkin:       allClasses,
	armorCloth:      allClasses,
	armorLeather:    allClassesBut(classMage),
	armorChainMail:  onlyClasses(classFighter, classTinker, classPaladin),
	armorPlateMail:  onlyClasses(classFighter, classTinker, classPaladin),
	armorMagicChain: onlyClasses(classFighter, classPaladin),
	armorMagicPlate: onlyClasses(classPaladin),
	armorMysticRobe: allClasses,
}

// mayUse says whether the member's class may use an item allowed to the classes given.
func mayUse(member int, classes uint8) bool {
	return classBit(int(party.members[member].class))&classes != 0
}

// sayMayNotUse tells that the member's class may not use the item (a weapon is "a ...").
func sayMayNotUse(member int, item string, isWeapon bool) {
	printText("\nA ")
	printText(peopleNames[int(party.members[member].class)])
	printText(" may NOT use ")
	if isWeapon {
		printChar('a')
	}
	newLine()
	printText(item)
	newLine()
	sound(soundBeep)
}

// maxCarried is the most of one weapon or armour the party can carry.
const maxCarried = 99

// putBack returns an item to the party's stock (hands and skin are not stocked).
func putBack(stock []uint16, item uint16) {
	if item != 0 {
		stock[item] = min(stock[item]+1, maxCarried)
	}
}

// afterChoosing restores the status area after a list was shown for a choice.
func afterChoosing() {
	clearStatus()
	clearStatusTitle()
	showStats()
	if gameMode >= modeCombat {
		invertStatusLine(int(activeMember))
	}
}

func cmdReady() {
	var member int
	if gameMode >= modeCombat {
		printText("Ready Weapon!\n")
		member = int(activeMember)
	} else {
		printText("Ready a weapon\n")
		if member = askMember("for:" + answerRoom(2)); member < 0 {
			return
		}
	}
	clearStatus()
	showWeapons()
	textRow, textCol = promptRow, promptCol
	weapon := askLetter("Weapon:"+answerRoom(2), 'A', 'P')
	afterChoosing()
	if weapon < 0 {
		return
	}
	weapon -= 'A'
	if weapon != weaponHands {
		if party.weapons[weapon] == 0 {
			sayNoneLeft()
			return
		}
		if !mayUse(member, weaponClasses[weapon]) {
			sayMayNotUse(member, weaponNames[weapon], true)
			return
		}
	}
	putBack(party.weapons[:], party.members[member].weapon)
	if weapon != weaponHands {
		party.weapons[weapon]--
	}
	party.members[member].weapon = uint16(weapon)
	printText(weaponNames[weapon])
	newLine()
}

func cmdWear() {
	printText("Wear Armour\n")
	member := askMember("for:" + answerRoom(2))
	if member < 0 {
		return
	}
	clearStatus()
	showArmour()
	textRow, textCol = promptRow, promptCol
	armor := askLetter("Armour:"+answerRoom(2), 'A', 'H')
	afterChoosing()
	if armor < 0 {
		return
	}
	armor -= 'A'
	if armor != armorSkin {
		if party.armors[armor] == 0 {
			sayNoneLeft()
			return
		}
		if !mayUse(member, armorClasses[armor]) {
			sayMayNotUse(member, armorNames[armor], false)
			return
		}
	}
	putBack(party.armors[:], party.members[member].armor)
	if armor != armorSkin {
		party.armors[armor]--
	}
	party.members[member].armor = uint16(armor)
	printText(armorNames[armor])
	newLine()
}
