package u4

// Whirlpools and twisters, moving on the battlefield, and the dungeon's command loop.

// ---- Whirlpools and twisters ----

// clearOthersAt removes the other people where person i is (a whirlpool or twister swallows
// them), with noise.
func clearOthersAt(i, effectID int) {
	people := &area.people
	for j := len(people.tile) - 1; j >= 0; j-- {
		if j != i && people.tile[j] != 0 && people.x[j] == people.x[i] && people.y[j] == people.y[i] {
			people.baseTile[j] = 0
			people.tile[j] = 0
			redraw()
			sound(effectID)
		}
	}
}

// Where a whirlpool casts the party's ship: Lock Lake.
const whirlpoolDestX, whirlpoolDestY = 127, 78

// whirlpoolStrikes lets whirlpool i swallow the party (hurt, and cast into Lock Lake) and
// whatever else is there.
func whirlpoolStrikes(i int) {
	if party.x == area.people.x[i] && party.y == area.people.y[i] {
		redraw()
		hitX, hitY = viewCenter, viewCenter
		hitTile = tileWhirlpool
		drawView()
		party.x, party.y = whirlpoolDestX, whirlpoolDestY
		sound(soundRising)
		damageParty()
		party.tile = tileShipWest
		hitTile = 0
		centerMap()
	}
	clearOthersAt(i, soundRising)
}

// twisterStrikes lets twister i hurt the party four times, and swallow whatever else is
// there.
func twisterStrikes(i int) {
	if party.x == area.people.x[i] && party.y == area.people.y[i] {
		redraw()
		sound(soundFalling)
		for range 4 {
			damageParty()
		}
		hitTile = 0
	}
	clearOthersAt(i, soundFalling)
}

// whirlpoolsAndTwisters lets the world's whirlpools and twisters (among the first four
// monsters) strike.
func whirlpoolsAndTwisters() {
	if gameMode != modeOutdoors {
		return
	}
	for i := 3; i >= 0; i-- {
		switch area.people.tile[i] {
		case tileWhirlpool:
			whirlpoolStrikes(i)
		case tileTwister:
			twisterStrikes(i)
		}
	}
}

// ---- Moving on the battlefield ----

func saySameExit() {
	printText("All must use same exit!\n")
	sound(soundBeep)
}

// offField is the position of a member who left the battlefield.
const offField = 128

// memberLeavesField takes the active member off the battlefield. Fleeing a fight against
// evil unhurt is cowardly: valor and sacrifice suffer.
func memberLeavesField() {
	a := int(activeMember)
	battlefield.memberX[a], battlefield.memberY[a] = offField, offField
	fighters.memberTile[a] = 0
	sound(soundFailure)
	member := &party.members[a]
	if gameMode == modeCombat && !isGoodCreature(encounterTile) && member.hp == member.maxHP {
		lowerKarma(&party.karmaValor, 2)
		lowerKarma(&party.karmaSacrifice, 2)
	}
}

// memberExits lets the active member leave the battlefield at (x, y), past its edge. In a
// dungeon room all must leave by the same side, the one the first took.
func memberExits(x, y uint8) {
	if gameMode == modeCombatRoom {
		switch {
		case fleeDirX != 0: // someone left west or east
			if x != fleeDirX {
				saySameExit()
				return
			}
		case fleeDirY != 0: // someone left north or south
			if y != fleeDirY {
				saySameExit()
				return
			}
		case x >= viewSize:
			fleeDirX = x
		default:
			fleeDirY = y
		}
	}
	memberLeavesField()
}

// springTriggers springs a dungeon room's triggers at (x, y): each of the four is the tile
// it sets, the position stepping on which springs it, and the two positions it sets the
// tile at (0,0: none), every coordinate four bits.
func springTriggers(x, y uint8) {
	for t := 3; t >= 0; t-- {
		trigger := roomTriggers[t*4 : t*4+4]
		tile, at := trigger[0], trigger[1]
		if tile == 0 || at/16 != x || at%16 != y {
			continue
		}
		for _, target := range trigger[2:] {
			ty, tx := int(target%16), int(target/16)
			if ty != 0 || tx != 0 {
				*combatTile(ty, tx) = tile
			}
		}
	}
}

// moveMember moves party member i to (x, y) on the battlefield, or off it.
func moveMember(i int, x, y uint8) {
	sound(soundStep)
	if x >= viewSize || y >= viewSize {
		memberExits(x, y)
		return
	}
	tile := viewTiles[int(y)*viewSize+int(x)]
	if !isWalkable(tile) {
		sayBlocked()
		return
	}
	if slowProgress(tile) {
		saySlowProgress()
		return
	}
	battlefield.memberX[i], battlefield.memberY[i] = x, y
	sound(soundStep)
	if gameMode == modeCombatRoom {
		springTriggers(x, y)
	}
}

func cmdNorthInCombat() { // North [combat]
	a := int(activeMember)
	moveMember(a, battlefield.memberX[a], battlefield.memberY[a]-1)
}

func cmdSouthInCombat() { // South [combat]
	a := int(activeMember)
	moveMember(a, battlefield.memberX[a], battlefield.memberY[a]+1)
}

func cmdEastInCombat() { // East [combat]
	a := int(activeMember)
	moveMember(a, battlefield.memberX[a]+1, battlefield.memberY[a])
}

func cmdWestInCombat() { // West [combat]
	a := int(activeMember)
	moveMember(a, battlefield.memberX[a]-1, battlefield.memberY[a])
}

// ---- The dungeon ----

// dungeonLoop runs the party's commands in a dungeon until it climbs out above the first
// level.
func dungeonLoop() {
	var key, lastKey int
	forceRender = 1
	flushKeys()
	for {
		groundEvent()
		printChar(promptChar)
		redraw()
		if checkPartyAlive() {
			ask(pCmd)
			const commandSeconds = 25 // waiting this long passes
			waitSeconds(commandSeconds, true)
			key = keySpace
			if keyWaiting() {
				key = readKey()
			}
			key = lowerKey(key)
			switch key {
			case keySpace:
				sayPass()
			case keyUp, keyUpAlt:
				cmdAdvance()
			case keyLeft, keyLeftAlt:
				cmdTurnLeft()
			case keyDown, keyDownAlt:
				cmdRetreat()
			case keyRight, keyRightAlt:
				cmdTurnRight()
			case keyC:
				cmdCast()
			case keyD:
				cmdDescendDungeon()
			case keyG:
				cmdGet()
			case keyH:
				cmdHoleUp()
			case keyI:
				cmdIgnite()
			case keyK:
				cmdKlimbDungeon()
			case keyM:
				cmdMix()
			case keyN:
				cmdNewOrder()
			case keyP:
				cmdPeer()
			case keyQ:
				cmdQuit()
			case keyR:
				cmdReady()
			case keyS:
				cmdSearchDungeon() // search
			case keyU:
				cmdUse()
			case keyV:
				cmdVolume()
			case keyW:
				cmdWear()
			case keyZ:
				cmdZtats()
			case keyA, keyB, keyE, keyF, keyJ, keyL, keyO, keyT, keyX, keyY:
				sayNotHere()
				sound(soundBadCommand)
			case keyCtrlS:
				if lastKey == keyAltZ {
					showKarmas()
					break
				}
				fallthrough
			default:
				printText("Bad command\n")
				sound(soundBadCommand)
				turnPasses = false
			}
			lookAround()
		} else {
			printText("Zzzzz\n")
		}
		if turnPasses && party.z >= 0 {
			endDungeonTurn()
		} else {
			turnPasses = true
		}
		lastKey = key
		if party.z < 0 {
			break
		}
	}
	leaveTown()
	dungeonFile = nil
}

// dungeonEncounter starts a fight when a monster is in the party's room, taking it off the
// level; it reports whether it did.
func dungeonEncounter() bool {
	room := hereTile
	switch dngKind(room) {
	case dngTrap, dngFountain, dngField, dngRoom, dngWall:
		return false
	}
	if dngDetail(room) == 0 {
		return false
	}
	foeTile = tileRat + (dngDetail(room)-1)*4
	encounterTile = foeTile
	here := area.dungeonRoom(int(party.z), int(party.y), int(party.x))
	*here = dngKind(*here)
	people := &area.people
	for i := len(people.tile) - 1; i >= 0; i-- {
		if int16(people.behaviour[i]) == party.z && people.x[i] == party.x && people.y[i] == party.y {
			people.behaviour[i] = 0
			people.tile[i] = 0
			people.x[i] = 0
			people.y[i] = 0
			break
		}
	}
	startDungeonFight()
	return true
}

// endDungeonTurn ends a turn in the dungeon: sleepers may wake, poison and hunger hurt, the
// monsters move, the light burns down, and a monster or room ahead starts a fight.
func endDungeonTurn() {
	checkPartyAlive()
	for i := 0; i < int(party.size); i++ {
		member := &party.members[i]
		if member.status == statusAsleep && randomBits(7) == 0 {
			member.status = statusGood
		} else if member.status == statusPoisoned {
			damageMember(i, 2)
			invertStatusLine(i)
			sound(soundMonsterHit)
			invertStatusLine(i)
		}
	}
	if eatFood(int(party.size)) {
		printText("\nStarving!!!\n")
		for i := 0; i < int(party.size); i++ {
			if isAlive(i) {
				damageMember(i, 2)
			}
		}
		for i := int(party.size) - 1; i >= 0; i-- {
			invertStatusLine(i)
		}
		sound(soundMonsterHit)
		for i := int(party.size) - 1; i >= 0; i-- {
			invertStatusLine(i)
		}
	}
	regainMP()
	moveDungeonMonsters()
	spawnDungeonMonsters()
	if party.flyingOrLight != 0 {
		party.flyingOrLight-- // the light burns down
	}
	party.moves++
	if spellTurns != 0 {
		spellTurns--
		if spellTurns == 0 {
			activeSpell = effectNone
		}
	}
	if party.flyingOrLight == 0 {
		printText("It's Dark!\n")
	}
	showStats()
	if !dungeonEncounter() {
		for dngKind(hereTile) == dngRoom {
			checkPartyAlive()
			fightRoom()     // the room's fight
			renderDungeon() // render the view
		}
	}
}

// canPass says whether the party can move into a room holding contents: anything below a
// door but an energy field, and going forward also doors and battle rooms.
func canPass(contents uint8, forward bool) bool {
	switch dngKind(contents) {
	case dngDoor, dngSecretDoor, dngRoom:
		if forward {
			return true
		}
	}
	const energyField = dngField + 1
	return contents != energyField && contents < dngDoor
}

func cmdAdvance() {
	printText("Advance\n")
	if !canPass(aheadTile, true) {
		sayBlocked()
		return
	}
	party.x = uint8(aheadX(int(party.x), int(party.dir)))
	party.y = uint8(aheadY(int(party.y), int(party.dir)))
}

func cmdRetreat() {
	printText("Retreat\n")
	if !canPass(behindTile, false) {
		sayBlocked()
		return
	}
	party.x = uint8(behindX(int(party.x), int(party.dir)))
	party.y = uint8(behindY(int(party.y), int(party.dir)))
}

func cmdTurnLeft() {
	printText("Turn Left\n")
	party.dir = (party.dir + 3) % 4
	turnPasses = false
}

func cmdTurnRight() {
	printText("Turn Right\n")
	party.dir = (party.dir + 1) % 4
	turnPasses = false
}

func cmdKlimbDungeon() {
	printText("Klimb ")
	if kind := dngKind(hereTile); kind != dngLadderUp && kind != dngLadderUpDown {
		sayWhat()
		return
	}
	printText("up!\n")
	party.z--
	if party.z >= 0 {
		printText("To level ")
		printChar(int(party.z) + '1')
		newLine()
	}
}

func cmdDescendDungeon() {
	printText("Descend ")
	if kind := dngKind(hereTile); kind != dngLadderDown && kind != dngLadderUpDown {
		sayWhat()
		return
	}
	party.z++
	printText("down to level ")
	printChar(int(party.z) + '1')
	newLine()
}
