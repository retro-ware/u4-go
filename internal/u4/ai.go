package u4

// The monsters' minds: wandering the dungeon levels, and fighting.

// combatTileAt is the battlefield's tile (y, x). Off the 11x11 field (a monster that fled,
// a look past the edge) the original read whatever followed; this reads 0.
func combatTileAt(y, x int) uint8 {
	i := y*viewSize + x
	if i < 0 || i >= len(battlefield.tiles) {
		return 0
	}
	return battlefield.tiles[i]
}

// dungeonMonsterCanEnter says whether dungeon monster i may move to room (x, y) of the
// party's level: an empty room that is no trap, fountain, field, battle room or wall, and
// (but one time in eight) not the room it came from.
func dungeonMonsterCanEnter(x, y uint8, i int) bool {
	room := *area.dungeonRoom(int(party.z), int(y), int(x))
	if dngDetail(room) != 0 {
		return false // a monster is there
	}
	switch dngKind(room) {
	case dngTrap, dngFountain, dngField, dngRoom, dngWall:
		return false
	}
	return randomBits(7) == 0 || area.people.oldX[i] != x || area.people.oldY[i] != y
}

// wrapLevel wraps a coordinate round a dungeon level.
func wrapLevel(c int) uint8 { return uint8(c) % levelSize }

// moveDungeonMonsters lets the monsters of the party's level (but reapers and mimics) wander
// a room, trying up to eight random directions, and marks them on the level's map.
func moveDungeonMonsters() {
	people := &area.people
	for i := len(people.tile) - 1; i >= 0; i-- {
		kind := people.tile[i]
		if kind == 0 || int16(people.behaviour[i]) != party.z || kind == tileReaper || kind == tileMimic {
			continue // behaviour holds a dungeon monster's level
		}
		x, y := people.x[i], people.y[i]
		if party.x == x && party.y == y {
			continue
		}
		var toX, toY uint8
		canEnter := func(nx, ny uint8) bool {
			toX, toY = nx, ny
			return dungeonMonsterCanEnter(nx, ny, i)
		}
		found := false
		for range 8 {
			dx := randomSign()
			dy := randomSign()
			if randomBits(1) != 0 && dx != 0 && canEnter(wrapLevel(int(x)+dx), y) ||
				dy != 0 && canEnter(x, wrapLevel(int(y)+dy)) ||
				canEnter(wrapLevel(int(x)+dx), y) {
				found = true
				break
			}
		}
		if !found {
			continue
		}
		*area.dungeonRoom(int(party.z), int(toY), int(toX)) |= uint8((int(kind)-tileRat)/4 + 1)
		*area.dungeonRoom(int(party.z), int(y), int(x)) = dngKind(*area.dungeonRoom(int(party.z), int(y), int(x)))
		people.oldX[i], people.oldY[i] = x, y
		people.x[i], people.y[i] = toX, toY
	}
	redraw()
}

// spawnDungeonMonsters fills the party's level's two monster places, when free, with
// monsters the stronger the deeper the level, in empty rooms.
func spawnDungeonMonsters() {
	people := &area.people
	z := int(party.z)
	for slot := z*4 + 1; slot >= z*4 && slot >= 0; slot-- {
		if people.tile[slot] != 0 {
			continue
		}
		x := uint8(randomBits(7))
		y := uint8(randomBits(7))
		// The original asks for both coordinates to differ from the party's.
		if party.x == x || party.y == y || *area.dungeonRoom(z, int(y), int(x)) != dngCorridor {
			continue
		}
		kind := uint8((randomBits(3)+z)*4 + tileRat)
		if kind == tileMimic {
			continue
		}
		people.tile[slot] = kind
		people.x[slot], people.oldX[slot] = x, x
		people.y[slot], people.oldY[slot] = y, y
		people.behaviour[slot] = uint8(party.z)
		*area.dungeonRoom(z, int(y), int(x)) = (kind-tileRat)/4 + 1
	}
}

// memberUnderHit is the living party member where the missile is (hitX, hitY), or -1.
func memberUnderHit() int {
	for i := int(party.size) - 1; i >= 0; i-- {
		if isAlive(i) && hitX == battlefield.memberX[i] && hitY == battlefield.memberY[i] {
			return i
		}
	}
	return -1
}

// monsterHurts lets monster hurt member, by up to a quarter of its strength; it reports
// whether the member died.
func monsterHurts(member, monster int) bool {
	invertStatusLine(member)
	drawCombat()
	sound(soundMemberHit)
	hitTile = 0
	drawCombat()
	damage := randomBelowSafe(int(monsterStrength[monsterIndex(fighters.tile[monster])]) / 4)
	// The original took the damage for binary-coded decimal, reading its high four bits as
	// the tens: 27 (one 16 and 11) counts 1*10 + 27%10 = 17.
	damage = damage/16*10 + damage%10
	invertStatusLine(member)
	if !damageMember(member, damage) {
		showStats()
		return false
	}
	printText(cstr(party.members[member].name[:]))
	printText(" is Killed!\n")
	raiseKarma(&party.karmaSacrifice, 1)
	showStats()
	checkPartyAlive()
	return true
}

// flashMemberHit flashes member's status line on the battlefield.
func flashMemberHit(member int) {
	invertStatusLine(member)
	drawCombat()
	sound(soundMemberHit)
	invertStatusLine(member)
	hitTile = 0
}

// monsterShoots flies missile from monster toward (dx, dy) across the battlefield, and
// lets it hit the first party member in its way.
func monsterShoots(missile uint8, monster, dx, dy int) {
	hitTile = uint16(missile)
	hitX, hitY = battlefield.monsterX[monster], battlefield.monsterY[monster]
	sound(soundMissile)
	member := -1
	for advanceHit(uint8(dx), uint8(dy)) {
		drawCombat()
		if member = memberUnderHit(); member != -1 {
			break
		}
	}
	if member == -1 {
		if hitTile == tileLava {
			*combatTile(int(hitY), int(hitX)) = tileLava // lava stays where it lands
		}
		hitTile = 0
		return
	}
	newLine()
	printText(cstr(party.members[member].name[:]))
	printChar(' ')
	status := &party.members[member].status
	switch hitTile {
	case tilePoisonField:
		printText("Poisoned!\n")
		flashMemberHit(member)
		if randomBits(1) != 0 || *status != statusGood {
			printText("Failed.\n")
		} else {
			*status = statusPoisoned
		}
	case tileEnergyField:
		printText("Electrified!\n")
		monsterHurts(member, monster)
	case tileFireField:
		printText("Fiery Hit!\n")
		monsterHurts(member, monster)
	case tileSleepField:
		printText("Slept!\n")
		flashMemberHit(member)
		if randomBits(1) != 0 || *status != statusGood {
			printText("Failed.\n")
		} else {
			*status = statusAsleep
			fighters.memberTile[member] = tileAsleep
		}
	case tileLava:
		printText("Lava Hit!\n")
		monsterHurts(member, monster)
	case tileMagicSphere:
		printText("Magical Hit!\n")
		monsterHurts(member, monster)
	default:
		printText("Hit!\n")
		monsterHurts(member, monster)
	}
	hitTile = 0
	showStats()
}

// monsterCanMoveTo says whether monster i can move to (x, y) on the battlefield. Off the
// field it can when weak enough to flee.
func monsterCanMoveTo(x, y uint8, i int) bool {
	tile := combatTileAt(int(y), int(x)) // read before the range check, as the original
	if slowProgress(tile) {
		return false
	}
	if x >= viewSize || y >= viewSize {
		return fighters.hp[i] < fleeingHP
	}
	if fighters.x[i] == x && fighters.y[i] == y && randomBits(3) != 0 {
		return false // rarely back where it came from
	}
	for j := len(fighters.tile) - 1; j >= 0; j-- {
		if fighters.tile[j] != 0 && battlefield.monsterX[j] == x && battlefield.monsterY[j] == y {
			return false
		}
	}
	for j := 7; j >= 0; j-- {
		if fighters.memberTile[j] != 0 && battlefield.memberX[j] == x && battlefield.memberY[j] == y {
			return false
		}
	}
	switch {
	case fighters.tile[i] < tilePirateWest: // people
		return isWalkable(tile)
	case fighters.tile[i] < tileTwister: // sea monsters
		return tile <= tileShallowWater
	}
	switch fighters.baseTile[i] {
	case tileInsects, tileTwister, tileBat, tileDaemon, tileDragon, tileBalron: // they fly
		return tile <= tileShallowWater || isWalkable(tile)
	case tileGhost, tileZorn: // through walls
		return tile >= tileSwamp && tile != tileEnergyField
	}
	return isWalkable(tile)
}

// fleeingHP: a monster weaker than this flees.
const fleeingHP = 24

// monsterMissile is what monster kind shoots or casts, 0 if nothing.
func monsterMissile(kind uint8) uint8 {
	switch kind {
	case tileMage, tileLordBritish, tileSeahorse, tileEvilMage, tileLiche, tileDaemon:
		return tileMagicSphere
	case tileNixie, tileTroll:
		return tileMissile
	case tileSquid:
		return tileEnergyField
	case tileSeaSerpent, tileHydra, tileDragon:
		return tileHitFlash // fire breath
	case tileSpider, tileMimic, tilePython:
		return tilePoisonField
	case tileGazer:
		return tileSleepField
	case tileEttin, tileCyclops:
		return tileRocks
	case tileLavaLizard:
		return tileLava
	case tileReaper, tileBalron:
		return uint8(tilePoisonField + randomBits(3)) // any field
	}
	return 0
}

// jinxStrikes: under Jinx a monster moving onto another hurts it instead; it reports
// whether that happened.
func jinxStrikes(x, y uint8) bool {
	if activeSpell != effectJinx {
		return false
	}
	other := monsterAt(x, y)
	if other == -1 {
		return false
	}
	hitTile = tileHitFlash
	hitX, hitY = x, y
	drawCombat()
	hitTile = 0
	sound(soundMonsterHit)
	damageMonster(other, -1, randomBits(63))
	drawCombat()
	return true
}

func stealFood() {
	party.food = max(party.food-2500, 0)
	showStats()
	sound(soundFailure)
}

func stealGold() {
	if randomBits(3) == 0 {
		party.gold = uint16(max(int(party.gold)-randomBits(63), 0))
		showStats()
		sound(soundFailure)
	}
}

// monsterStrikes lets monster strike member at close range, unless the member's armour (or
// Protection, half the time) wards it off.
func monsterStrikes(monster, member int) {
	hitX, hitY = battlefield.memberX[member], battlefield.memberY[member]
	hitTile = tileMissile
	drawCombat()
	sound(soundMonsterBlow)
	hitTile = 0
	if (activeSpell == effectProtection && randomBits(1) != 0) || armorDefense[party.members[member].armor] >= randomByte() {
		drawCombat()
		return
	}
	monsterHurts(member, monster)
}

// monsterFlees takes monster off the battlefield; fleeing good creatures were spared, which
// is compassionate and just.
func monsterFlees(monster int) {
	newLine()
	printText(creatureName(fighters.tile[monster]))
	printText(" Flees!\n")
	if isGoodCreature(fighters.tile[monster]) {
		raiseKarma(&party.karmaCompassion, 1)
		raiseKarma(&party.karmaJustice, 1)
	}
	sound(soundFailure)
	fighters.tile[monster] = 0
	fighters.hp[monster] = 0
}

// monsterTurn is monster's turn against the nearest party member target, dist away: it may
// shoot, put the party to sleep (balrons and reapers), steal, strike, flee or close in.
func monsterTurn(monster, target, dist int) {
	kind := fighters.tile[monster]
	dirX := signInt(int(battlefield.memberX[target]) - int(battlefield.monsterX[monster]))
	dirY := signInt(int(battlefield.memberY[target]) - int(battlefield.monsterY[monster]))
	mayShoot := true
	if kind == tileMimic { // a chest until the party comes near
		fighters.baseTile[monster] = tileChest
		if dist >= 5 {
			mayShoot = false
		} else {
			fighters.baseTile[monster] = tileMimic
		}
	}
	if mayShoot && randomBits(3) == 0 {
		missile := monsterMissile(kind)
		if missile != 0 && (missile != tileMagicSphere || activeSpell != effectNegate) {
			monsterShoots(missile, monster, dirX, dirY)
			redraw()
			return
		}
	}
	if randomBits(3) == 0 && (kind == tileBalron || kind == tileReaper) && activeSpell != effectNegate {
		redraw()
		invertMap()
		sound(soundSweep, 128)
		invertMap()
		for i := int(party.size) - 1; i >= 0; i-- {
			if fighters.memberTile[i] != 0 && party.members[i].status == statusGood && randomBits(1) != 0 {
				party.members[i].status = statusAsleep
				fighters.memberTile[i] = tileAsleep
			}
		}
		showStats()
		return
	}
	if fighters.hp[monster] < fleeingHP {
		dirX, dirY = -dirX, -dirY
	} else if dist < 2 {
		switch kind {
		case tileGremlin:
			stealFood()
		case tileRogue:
			stealGold()
		}
		monsterStrikes(monster, target)
		return
	}
	if kind == tileMimic || kind == tileReaper {
		return // they don't move
	}

	// Close in (or flee): across or along, then twice at random.
	var toX, toY uint8
	moved, jinxed := false, false
	tryMove := func(x, y uint8) bool {
		toX, toY = x, y
		if jinxStrikes(x, y) {
			jinxed = true
			return true
		}
		moved = monsterCanMoveTo(x, y, monster)
		return moved
	}
	for range 2 {
		x, y := battlefield.monsterX[monster], battlefield.monsterY[monster]
		if randomBits(1) != 0 && dirX != 0 && tryMove(x+uint8(dirX), y) ||
			dirY != 0 && tryMove(x, y+uint8(dirY)) ||
			dirX != 0 && tryMove(x+uint8(dirX), y) {
			break
		}
		dirX = randomSign()
		dirY = randomSign()
	}
	if jinxed {
		return
	}
	if moved {
		fighters.x[monster], fighters.y[monster] = battlefield.monsterX[monster], battlefield.monsterY[monster]
		battlefield.monsterX[monster], battlefield.monsterY[monster] = toX, toY
		if toX >= viewSize || toY >= viewSize {
			monsterFlees(monster)
			return
		}
		drawCombat()
		return
	}
	if dist < 2 {
		monsterStrikes(monster, target)
	}
}

// monstersTurn gives each monster on the battlefield its turn (sleepers wake one time in
// eight) against the nearest party member, then lets the fields hurt them.
func monstersTurn() {
	redraw()
	for m := len(fighters.tile) - 1; m >= 0; m-- {
		kind := fighters.tile[m]
		if kind == 0 {
			continue
		}
		if fighters.sleeping[m] == 0 || randomBits(7) == 0 {
			fighters.sleeping[m] = 0
			if kind == tileZorn { // a zorn negates magic
				activeSpell = effectNegate
				spellTurns = 2
				showStats()
			} else if kind == tileWisp && randomBits(7) == 0 { // a wisp teleports
				var x, y uint8
				for {
					x = uint8(randomBelow(viewSize))
					y = uint8(randomBelow(viewSize))
					if monsterCanMoveTo(x, y, m) {
						break
					}
				}
				battlefield.monsterX[m], battlefield.monsterY[m] = x, y
			}
			x, y := battlefield.monsterX[m], battlefield.monsterY[m]
			nearest, target := 99, -1
			for i := int(party.size) - 1; i >= 0; i-- {
				if isAlive(i) && fighters.memberTile[i] != 0 {
					d := abs(int(battlefield.memberX[i])-int(x)) + abs(int(battlefield.memberY[i])-int(y))
					if d < nearest {
						nearest, target = d, i
					}
				}
			}
			if target != -1 {
				monsterTurn(m, target, nearest)
			}
		}
		switch combatTileAt(int(battlefield.monsterY[m]), int(battlefield.monsterX[m])) {
		case tileFireField, tileLava:
			if fighters.tile[m] == tileLavaLizard || fighters.tile[m] >= tileDaemon {
				break // immune to fire
			}
			fallthrough
		case tilePoisonField:
			sound(soundMonsterHit)
			damageMonster(m, -1, randomBits(127))
		case tileSleepField:
			if !isUndead(fighters.tile[m]) && fighters.hp[m] >= randomByte() {
				fighters.sleeping[m] = 1
				sound(soundMonsterHit)
			}
		}
	}
	regainMP()
	party.moves++
	if spellTurns != 0 {
		spellTurns--
		if spellTurns == 0 {
			activeSpell = effectNone
		}
	}
}
