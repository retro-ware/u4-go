package u4

// Setting up a battle: the monsters' tables, the battlefields (CON files), the dungeon's
// rooms, and the battle's end.

// monsterTables holds three tables that ran on into each other in the original's data, and
// still do: monsterIndex gives a tile that is no monster (a townsperson attacked) 36 to 67,
// past the 52 strengths into the partners and group sizes.
var monsterTables = [52 + 36 + 38]uint8{
	// The monsters' strengths (their hit points, experience and blows), by monsterIndex:
	// the sea monsters,
	255, 255, 64, 96, 128, 96, 255, 255,
	// the land monsters,
	48, 48, 64, 80, 48, 96, 48, 192, 255, 48, 240, 128, 80, 48, 80, 48,
	112, 64, 128, 64, 176, 192, 96, 240, 112, 208, 224, 255,
	// and what the townspeople (tiles 32 to 47 and 80 to 95) read.
	112, 48, 96, 64, 96, 128, 144, 48,
	128, 48, 48, 48, 32, 32, 128, 255,

	// The partners that may join a group of each monster.
	tileRogue, tileRogue, tileSeahorse, tileSeaSerpent, tileSquid, tileNixie, tileWhirlpool, tileTwister,
	tileSkeleton, tileLavaLizard2, tileRat, tileLiche, tileSlime, tileEttin, tileGremlin, tileMimic, tileReaper, tileRat, tilePhantom, tileGhost,
	tileTroll, tileEvilMage, tileRogue, tileRat, tileDaemon, tileGazer, tileZorn, tilePhantom, tileDaemon, tileDaemon, tileHydra, tileGazer,
	tileBalron, tileDragon, tileBalron, tileBalron,

	// How many of each monster come at most.
	1, 1, 12, 4, 4, 8, 1, 1,
	12, 12, 6, 4, 15, 6, 15, 1, 1, 15, 4, 8, 10, 12, 10, 12,
	6, 8, 6, 12, 6, 4, 8, 4, 6, 4, 4, 1,
	0, 0,
}

var (
	monsterStrength  = monsterTables[0:]
	monsterPartner   = monsterTables[52:]
	monsterGroupSize = monsterTables[88:]
)

// weaponDamage is the most damage each weapon adds to the wielder's strength.
var weaponDamage = [16]uint8{8, 16, 24, 32, 40, 48, 64, 40, 56, 64, 96, 96, 128, 80, 160, 255}

// armorDefense: a blow fails when a random byte is at most this.
var armorDefense = [8]uint8{96, 128, 144, 160, 176, 192, 208, 248}

// weaponRanged says which weapons shoot (or are thrown).
var weaponRanged = [16]bool{
	weaponSling: true, weaponBow: true, weaponCrossbow: true, weaponOil: true,
	weaponMagicAxe: true, weaponMagicBow: true, weaponMagicWand: true,
}

// monsterIndex is a monster's place in the monster tables: the sea monsters (two tiles each)
// 0 to 7, the land monsters (four tiles each) 8 to 35; anything else 36 to 67.
func monsterIndex(tile uint8) int {
	if tile < tilePirateWest {
		return int(tile%32) + 36
	}
	t := int(tile - tilePirateWest)
	if t/2 < 8 {
		return t / 2
	}
	return t/4 + 4
}

// loadBattlefield loads the battlefield for a fight outdoors or in a town: by ship, against
// pirates, on the shore, or the ground the party stands on.
func loadBattlefield() {
	var file string
	aboard := party.tile <= tileShipSouth || int(hereTile)&^3 == tileShipWest
	switch {
	case aboard && encounterTile == tilePirateWest:
		foeTile = tileRogue
		file = "SHIPSHIP.CON"
	case aboard && foeGround < tileSwamp:
		file = "SHIPSEA.CON"
	case aboard:
		file = "SHIPSHOR.CON"
	case encounterTile == tilePirateWest:
		foeTile = tileRogue
		file = "SHORSHIP.CON"
	case foeGround < tileSwamp:
		file = "SHORE.CON"
	default:
		switch hereTile {
		case tileSwamp:
			file = "MARSH.CON"
		case tileScrub:
			file = "BRUSH.CON"
		case tileForest:
			file = "FOREST.CON"
		case tileHills:
			file = "HILL.CON"
		case tileDungeon:
			file = "DUNGEON.CON"
		case tileBrickFloor:
			file = "BRICK.CON"
		case tileBridge, tileBridgeNorth, tileBridgeSouth, tileWoodFloor:
			file = "BRIDGE.CON"
		default:
			file = "GRASS.CON"
		}
	}
	LoadRec(file, &battlefield)
}

// startDungeonFight fights the monster in the party's dungeon room, on a battlefield like
// the room.
func startDungeonFight() {
	file := "DNG0.CON"
	switch dngKind(hereTile) {
	case dngLadderUp:
		file = "DNG1.CON"
	case dngLadderDown:
		file = "DNG2.CON"
	case dngLadderUpDown:
		file = "DNG3.CON"
	case dngChest:
		file = "DNG4.CON"
	case dngDoor:
		file = "DNG5.CON"
	case dngSecretDoor:
		file = "DNG6.CON"
	}
	LoadRec(file, &battlefield)
	runCombat()
	forceRender = 1
}

// startInnFight fights the thief in the inn at night.
func startInnFight() {
	LoadRec("INN.CON", &battlefield)
	runCombat()
}

// beginCombat clears the battlefield's fighters and fights encounterTile.
func beginCombat() {
	printText("\n\n**** Combat ****\n\n")
	foeTile = encounterTile
	for i := len(fighters.tile) - 1; i >= 0; i-- {
		fighters.tile[i] = 0
		fighters.x[i] = 0
		fighters.y[i] = 0
		fighters.sleeping[i] = 0
	}
	for i := 7; i >= 0; i-- {
		fighters.memberTile[i] = 0
	}
	loadBattlefield()
	runCombat()
}

// startCombat fights person i, noting the ground they stand on.
func startCombat(i int) {
	encounterTile = area.people.tile[i]
	foeX, foeY = area.people.x[i], area.people.y[i]
	if party.loc == 0 {
		foeGround = *area.townTile(int(foeY)-int(windowQuarter.y)*quarterTiles, int(foeX)-int(windowQuarter.x)*quarterTiles)
	} else {
		foeGround = *area.townTile(int(foeY), int(foeX))
	}
	area.people.behaviour[i] = 0
	area.people.baseTile[i] = 0
	area.people.tile[i] = 0
	beginCombat()
}

// setUpFighters sets out the monsters (how many: from the group size, the guards by the
// party's size, never more than twice the party) with partners now and then, and the party.
func setUpFighters() {
	for i := len(fighters.tile) - 1; i >= 0; i-- {
		fighters.tile[i] = 0
	}
	var last int // the monsters' count, less one
	groupOf := func(tile int) int { return int(monsterGroupSize[monsterIndex(uint8(tile))]) }
	someOf := func(n int) int { return (randomBelowSafe(n) + n) / 2 }
	switch {
	case modeBeforeCombat == modeTown && gameMode != modeCombatCamp:
		if encounterTile == tileGuard {
			last = int(party.size)*2 - 1
		}
	case foeTile < tilePirateWest:
		last = someOf(groupOf(int(foeTile)))
	default:
		// Pirates and monsters: up to eight; none drawn asks the group size of tile 0,
		// which is 0.
		if last = randomBits(7); last == 0 {
			last = someOf(groupOf(0))
		}
	}
	for last/2 >= int(party.size) {
		last = randomBelowSafe(int(party.size) * 2)
	}
	for n := last; n >= 0; n-- {
		var slot int
		for {
			slot = randomBits(15)
			if fighters.tile[slot] == 0 {
				break
			}
		}
		kind := foeTile
		if foeTile > tilePirateWest && n != 0 {
			if randomBits(31) == 0 {
				kind = monsterPartner[monsterIndex(monsterPartner[monsterIndex(foeTile)])]
			} else if randomBits(7) == 0 {
				kind = monsterPartner[monsterIndex(foeTile)]
			}
		}
		fighters.tile[slot] = kind
		fighters.baseTile[slot] = kind
		strength := int(monsterStrength[monsterIndex(kind)])
		fighters.hp[slot] = uint8(strength/2 | randomBelowSafe(strength))
	}
	setUpMembers()
}

// setUpMembers puts the living party members on the battlefield.
func setUpMembers() {
	for i := int(party.size) - 1; i >= 0; i-- {
		if isAlive(i) {
			fighters.memberTile[i] = uint8(memberShape(i))
			memberFrame[i] = fighters.memberTile[i]
		} else {
			fighters.memberTile[i] = 0
		}
	}
}

// runCombat sets the fighters out and fights, then returns to the mode before.
func runCombat() {
	if gameMode != modeCombatCamp {
		modeBeforeCombat = int16(gameMode)
		gameMode = modeCombat
	}
	setUpFighters()
	fight()
	gameMode = uint16(modeBeforeCombat)
}

// A dungeon's file holds after its levels' maps (dungeonMapSize bytes) its rooms, each
// roomSize bytes: the triggers (see springTriggers), the monsters' tiles and positions, the
// party members' positions for each side the party may come in by, and the room's map.
const (
	roomSize         = 256
	roomMonsterTiles = 16
	roomMonsterX     = 32
	roomMonsterY     = 48
	roomMembers      = 64 // x then y of the 8 members for each side, 16 bytes a side
	roomMap          = 128
	abyssRoomsPerTwo = 16 // the Abyss has 16 rooms for every two levels
	altarRoomNumber  = 15 // a room number below the Abyss: an altar room
	altarLoveX       = 3  // the altar room's entrance column for love
	altarTruthMaxX   = 2  // and for truth (courage's is east of love's)
)

// fightRoom fights the battle of the dungeon room the party entered.
func fightRoom() {
	number := int(dngDetail(hereTile))
	offset := dungeonMapSize + number*roomSize
	if party.loc == locAbyss {
		offset += int(party.z/2) * abyssRoomsPerTwo * roomSize
	}
	var room [roomSize]uint8
	copy(room[:], at(dungeonFile, offset))
	modeBeforeCombat = modeDungeon
	gameMode = modeCombatRoom
	copy(roomTriggers[:], room[:16])
	altarRoom = noAltar
	if number == altarRoomNumber && party.loc < locAbyss {
		printText("\nThe Altar Room of ")
		switch {
		case party.x == altarLoveX:
			printText("Love\n")
			altarRoom = altarLove
		case party.x <= altarTruthMaxX:
			printText("Truth\n")
			altarRoom = altarTruth
		default:
			printText("Courage\n")
			altarRoom = altarCourage
		}
	}
	side := (int(party.dir) + 1) % 4 // the side the party comes in by
	xs := room[roomMembers+side*16:]
	ys := xs[8:]
	for i := 7; i >= 0; i-- {
		battlefield.memberX[i], battlefield.memberY[i] = xs[i], ys[i]
	}
	setUpMembers()
	for i := len(fighters.tile) - 1; i >= 0; i-- {
		// Empty places (tile 0) are set up too: the original tested an array's address.
		kind := room[roomMonsterTiles+i]
		battlefield.monsterX[i] = room[roomMonsterX+i]
		battlefield.monsterY[i] = room[roomMonsterY+i]
		fighters.tile[i] = kind
		fighters.baseTile[i] = kind
		strength := int(monsterStrength[monsterIndex(kind)])
		fighters.hp[i] = uint8(strength/2 | randomBelowSafe(strength))
		if kind == tileMimic {
			fighters.baseTile[i] = tileChest
		}
	}
	copy(battlefield.tiles[:], room[roomMap:roomMap+viewSize*viewSize])
	fight()
	gameMode = uint16(modeBeforeCombat)
}

// afterFight redraws the dungeon (after a fight there) and drops the keys pressed meanwhile.
func afterFight() {
	if modeBeforeCombat == modeDungeon {
		lookAround()
	}
	flushKeys()
}

// freePersonSlot is a free place among the people for something left behind, else a random
// one of the last 24. The original's loop ran down to 1 outdoors (it meant to stop at 8):
// there finding none gives 0.
func freePersonSlot() int {
	lowest := 0
	if modeBeforeCombat == modeOutdoors {
		lowest = 1
	}
	i := 31
	for ; i >= lowest; i-- {
		if area.people.tile[i] == 0 {
			break
		}
	}
	if i == -1 || (i == 7 && modeBeforeCombat == modeOutdoors) {
		i = randomBelow(24) + 8
	}
	return i
}

// leaveSpoils leaves what the beaten foe had: a pirate its ship, others a chest (but bats,
// insects, slimes and the like), in the dungeon in an empty room.
func leaveSpoils() {
	kind := int(encounterTile)
	switch {
	case kind == tilePirateWest:
		i := freePersonSlot()
		area.people.x[i], area.people.y[i] = foeX, foeY
		area.people.tile[i] = tileShipWest
		area.people.baseTile[i] = tileShipWest
	case modeBeforeCombat == modeDungeon:
		if kind != tileBat && kind != tileInsects && kind != tileSlime {
			if here := area.dungeonRoom(int(party.z), int(party.y), int(party.x)); *here == dngCorridor {
				*here = dngChest
			}
		}
	case modeBeforeCombat != modeCombatCamp || hereTile < tileShipWest || hereTile > tileShipSouth:
		carries := kind >= tileRat || (kind >= tileMage && kind <= tileShepherd) || (kind >= tileGuard && kind <= tileLordBritish)
		switch kind {
		case tileBat, tileSlime, tileGazer, tileInsects, tileWisp:
			carries = false
		}
		if carries && isWalkable(foeGround) {
			i := freePersonSlot()
			area.people.x[i], area.people.y[i] = foeX, foeY
			area.people.tile[i] = tileChest
			area.people.baseTile[i] = tileChest
		}
	}
	afterFight()
}

// altarRoomExits: the altar room of each virtue (truth, love, courage) leads by its four
// sides (north, east, south, west) into the dungeons of the virtues joined with it.
var altarRoomExits = [3][4]uint8{
	{locDeceit, locShame, locHythloth, locWrong},     // truth
	{locDespise, locWrong, locHythloth, locCovetous}, // love
	{locDestard, locCovetous, locHythloth, locShame}, // courage
}

// endCombat ends the battle: won, lost (all fled or fell), or a dungeon room left (an altar
// room into another dungeon).
func endCombat() {
	membersLeft := false
	for i := 7; i >= 0; i-- {
		if fighters.memberTile[i] != 0 {
			membersLeft = true
			break
		}
	}
	if membersLeft && gameMode != modeCombatRoom {
		printText("\nVictory!\n")
		if !isGoodCreature(encounterTile) {
			raiseKarma(&party.karmaValor, randomBits(1))
		}
		if modeBeforeCombat != modeCombatCamp {
			leaveSpoils()
		}
		return
	}
	if gameMode != modeCombatRoom {
		if isGoodCreature(encounterTile) {
			raiseKarma(&party.karmaCompassion, 2) // fled from a good creature
			raiseKarma(&party.karmaJustice, 2)
		} else {
			printText("Battle is lost!\n")
			lowerKarma(&party.karmaValor, 2)
		}
		afterFight()
		return
	}
	printText("Leave Room!\n")
	switch {
	case fleeDirX == 0 && signedByte(fleeDirY) < 0:
		party.dir = dirNorth
	case fleeDirX == 0:
		party.dir = dirSouth
	case signedByte(fleeDirX) < 0:
		party.dir = dirWest
	default:
		party.dir = dirEast
	}
	if altarRoom != noAltar {
		party.loc = uint16(altarRoomExits[altarRoom][(party.dir+3)%4])
		party.outX = placeX[party.loc-1]
		party.outY = placeY[party.loc-1]
		printText("into Dungeon\n")
		printText(placeName(party.loc))
		dungeonFile = nil
		loadDungeon()
	}
	newLine()
	flushKeys()
	party.x = uint8(aheadX(int(party.x), int(party.dir)))
	party.y = uint8(aheadY(int(party.y), int(party.dir)))
}
