package u4

// The people and monsters moving: in towns they stand, wander or follow the party; outdoors
// monsters close in and attack, pirate ships sail and fire broadsides, and new monsters
// appear.
//
// Where the original passed two random numbers in one call, its compiler drew the last
// argument's first; the Go draws them in that order into locals.

const attackedBy = "\nAttacked by "

// A person's behaviour (peopleTable.behaviour): below behaviourFollow they wander.
const (
	behaviourStill  = 0
	behaviourFollow = 128 // follows the party
	behaviourAttack = 255 // follows the party and attacks
)

var offMapTile uint8

// mapTileAt is the map window's tile (y, x), read as the original did, without bounds
// checks within the 32x32 map: a row's start or end reads the row next to it. Wholly off
// the map (where the original read other data, which canMoveTo then ignored) it gives a
// scratch tile.
func mapTileAt(y, x int) *uint8 {
	i := y*windowTiles + x
	if i < 0 || i >= windowTiles*windowTiles {
		offMapTile = 0
		return &offMapTile
	}
	return area.townTile(y, x)
}

// randomSign is -1, 0 or 1 at random.
func randomSign() int { return signInt(signedByte(randomByte())) }

// canMoveTo says whether person i can move to (x, y), onto tile.
func canMoveTo(i int, x, y, tile uint8) bool {
	if gameMode == modeTown && (x >= townTiles || y >= townTiles) {
		return false
	}
	kind := int(area.people.tile[i])
	if slowProgress(tile) {
		return false
	}
	if kind != tileWhirlpool && kind != tileTwister {
		people := &area.people
		for j := len(people.tile) - 1; j >= 0; j-- {
			if people.tile[j] != 0 && people.x[j] == x && people.y[j] == y && people.tile[j] != tileChest {
				return false
			}
		}
		if party.x == x && party.y == y {
			return false
		}
	}
	// Now and then a person won't step back where they came from.
	if kind != tilePirateWest && x == area.people.oldX[i] && y == area.people.oldY[i] && randomBits(3) == 0 {
		return false
	}
	switch {
	case kind < tilePirateWest: // people
		return isWalkable(tile)
	case kind == tilePirateWest:
		return tile <= tileMediumWater
	case kind < tileTwister: // sea monsters
		return tile <= tileShallowWater
	case kind == tileTwister, kind == tileBat, kind == tileInsects, kind == tileDaemon, kind == tileDragon, kind == tileBalron:
		return tile <= tileShallowWater || isWalkable(tile) // they fly
	case kind == tileZorn, kind == tileGhost:
		return tile >= tileSwamp && tile != tileEnergyField // through walls
	}
	return isWalkable(tile)
}

// moveTownPerson moves person i to (x, y) in a town. The original's test for the town's
// bounds used "or": a move is taken when either coordinate is in range.
func moveTownPerson(i int, x, y uint8) {
	if x < townTiles || y < townTiles {
		movePerson(i, x, y)
	}
}

// movePerson moves person i to (x, y), noting where they came from.
func movePerson(i int, x, y uint8) {
	people := &area.people
	people.oldX[i], people.x[i] = people.x[i], x
	people.oldY[i], people.y[i] = people.y[i], y
}

// wanderOutdoors moves monster i one tile toward (dx, dy), across or along; if it cannot it
// tries a random direction, up to tries more times.
func wanderOutdoors(i, dx, dy, tries int) {
	x, y := area.people.x[i], area.people.y[i]
	// Where the move ends in the window (the world wraps round, and so does a byte).
	wx := x - windowQuarter.x*quarterTiles + uint8(dx)
	wy := y - windowQuarter.y*quarterTiles + uint8(dy)
	if wx >= windowTiles || wy >= windowTiles {
		if tries >= 0 {
			// The original's second random direction is the sign of an unsigned byte: never
			// west.
			ry := signInt(int(randomByte()))
			rx := randomSign()
			wanderOutdoors(i, rx, ry, tries-1)
		}
		return
	}
	across := func() bool { // the move along x
		if canMoveTo(i, x+uint8(dx), y, *mapTileAt(int(wy)-dy, int(wx))) {
			movePerson(i, x+uint8(dx), y)
			return true
		}
		return false
	}
	if randomBits(1) != 0 && dx != 0 && across() {
		return
	}
	if dy != 0 && canMoveTo(i, x, y+uint8(dy), *mapTileAt(int(wy), int(wx)-dx)) {
		movePerson(i, x, y+uint8(dy))
		return
	}
	if across() {
		return
	}
	if tries >= 0 {
		ry := randomSign()
		rx := randomSign()
		wanderOutdoors(i, rx, ry, tries-1)
	}
}

// wanderInTown moves person i one tile toward (dx, dy) in a town, across or along; if the
// first choice is blocked they stay, else they try random directions, up to tries more
// times (a follower only once).
func wanderInTown(i, dx, dy, tries int) {
	x, y := int(area.people.x[i]), int(area.people.y[i])
	try := func(nx, ny int) bool {
		if canMoveTo(i, uint8(nx), uint8(ny), *mapTileAt(ny, nx)) {
			moveTownPerson(i, uint8(nx), uint8(ny))
			return true
		}
		return false
	}
	if randomBits(1) != 0 && dx != 0 {
		try(x+dx, y)
		return
	}
	if dy != 0 {
		try(x, y+dy)
		return
	}
	if try(x+dx, y) {
		return
	}
	if area.people.behaviour[i] != behaviourFollow && tries >= 0 {
		ry := randomSign()
		rx := randomSign()
		wanderInTown(i, rx, ry, tries-1)
	}
}

// moveTownPeople lets a town's people wander or follow the party; those who attack and
// stand next to it start a fight.
func moveTownPeople() {
	people := &area.people
	for i := len(people.tile) - 1; i >= 0; i-- {
		if people.tile[i] == 0 || people.behaviour[i] == behaviourStill {
			continue
		}
		if people.behaviour[i] < behaviourFollow {
			if randomBits(1) != 0 {
				ry := randomSign()
				rx := randomSign()
				wanderInTown(i, rx, ry, 2)
			}
			continue
		}
		distX := signedByte(party.x - people.x[i])
		distY := signedByte(party.y - people.y[i])
		if abs(distX)+abs(distY) >= 2 {
			wanderInTown(i, signInt(distX), signInt(distY), 2)
		} else if people.behaviour[i] == behaviourAttack {
			redraw()
			printText(attackedBy)
			printText(creatureName(people.tile[i]))
			newLine()
			startCombat(i)
		}
	}
}

// towardParty is the way from person i to the party: its direction (each -1, 0 or 1), its
// distances along x and y and the sum of their sizes.
func towardParty(i int) (dirX, dirY, distX, distY, dist int) {
	distY = signedByte(party.y - area.people.y[i])
	distX = signedByte(party.x - area.people.x[i])
	return signInt(distX), signInt(distY), distX, distY, abs(distX) + abs(distY)
}

// sailPirate sails pirate ship i ahead to (x, y) when the wind allows, or when blocked
// turns it at random.
func sailPirate(i int, x, y uint8) {
	people := &area.people
	// Unlike wanderOutdoors this doesn't wrap round the world.
	tile := *mapTileAt(int(y)-int(windowQuarter.y)*quarterTiles, int(x)-int(windowQuarter.x)*quarterTiles)
	if !canMoveTo(i, x, y, tile) {
		facing := (randomSign() + int(people.baseTile[i])) % 4 // the tile's low bits are its direction
		people.baseTile[i] = uint8(facing&3) | tilePirateWest
		return
	}
	if windAllows(uint16(people.baseTile[i] % 4)) {
		movePerson(i, x, y)
	}
}

// steerPirate turns pirate ship i toward the party when far (or one time in four), else
// sails it ahead.
func steerPirate(i int) {
	people := &area.people
	_, _, distX, distY, dist := towardParty(i)
	if dist > 5 || randomBits(3) == 0 {
		switch {
		case abs(distX) < abs(distY) && distY >= 0:
			people.baseTile[i] = tilePirateSouth
		case abs(distX) < abs(distY):
			people.baseTile[i] = tilePirateNorth
		case distX >= 0:
			people.baseTile[i] = tilePirateEast
		default:
			people.baseTile[i] = tilePirateWest
		}
		return
	}
	x, y := people.x[i], people.y[i]
	switch people.baseTile[i] {
	case tilePirateWest:
		sailPirate(i, x-1, y)
	case tilePirateNorth:
		sailPirate(i, x, y-1)
	case tilePirateEast:
		sailPirate(i, x+1, y)
	default:
		sailPirate(i, x, y+1)
	}
}

// chaseParty sails pirate ship i on when it faces the party's way (dirX, dirY), else steers.
func chaseParty(i, dirX, dirY int) {
	people := &area.people
	x, y := people.x[i], people.y[i]
	switch people.baseTile[i] {
	case tilePirateWest:
		if dirX < 0 {
			sailPirate(i, x+uint8(dirX), y)
			return
		}
	case tilePirateNorth:
		if dirY < 0 {
			sailPirate(i, x, y+uint8(dirY))
			return
		}
	case tilePirateEast:
		if dirX > 0 {
			sailPirate(i, x+uint8(dirX), y)
			return
		}
	default:
		if dirY > 0 {
			sailPirate(i, x, y+uint8(dirY))
			return
		}
	}
	steerPirate(i)
}

// fireMissile flies missile from person i toward (dx, dy), up to three tiles; it reports
// whether it reached the party. A person in the way is hit instead, and gone (a monster
// one time in four).
func fireMissile(i int, dx, dy, missile uint8) bool {
	x := area.people.x[i] + dx
	y := area.people.y[i] + dy
	sound(soundMissile)
	for range 3 {
		if party.x == x && party.y == y {
			return true
		}
		if hit := personAt(31, x, y); hit != -1 {
			sound(soundMonsterHit)
			if hit >= 8 || randomBits(3) == 0 {
				area.people.tile[hit] = 0
				area.people.baseTile[hit] = 0
			}
			return false
		}
		tile := mapTileAt(int(y)-int(windowQuarter.y)*quarterTiles, int(x)-int(windowQuarter.x)*quarterTiles)
		under := *tile
		*tile = missile
		redraw()
		*tile = under
		x += dx
		y += dy
	}
	return false
}

// shootAtParty fires missile from person i toward (dx, dy), hurting the party if it hits.
func shootAtParty(i, dx, dy int, missile uint8) {
	if !fireMissile(i, uint8(dx), uint8(dy), missile) {
		hitTile = 0
		return
	}
	redraw()
	hitX, hitY = viewCenter, viewCenter
	hitTile = tileHitFlash
	drawView()
	damageParty()
	hitTile = 0
	redraw()
}

// pirateAttack fires pirate ship i's broadside when the party is abeam, else chases it.
func pirateAttack(i, dirX, dirY int) {
	facing := area.people.baseTile[i]
	abeam := dirY == 0
	if facing == tilePirateWest || facing == tilePirateEast {
		abeam = dirX == 0
	}
	if abeam {
		shootAtParty(i, dirX, dirY, tileMissile)
	} else {
		chaseParty(i, dirX, dirY)
	}
}

// outOfWindow says whether monster i is past the window's east or south edge (above or
// left of it counts as in).
func outOfWindow(i int) bool {
	return int(area.people.x[i])-int(windowQuarter.x)*quarterTiles >= windowTiles ||
		int(area.people.y[i])-int(windowQuarter.y)*quarterTiles >= windowTiles
}

// moveMonsters moves the world's monsters (the first eight people): those next to the
// party attack, pirates hunt it, some breathe fire from afar, the rest close in, and the
// good ones, whirlpools and twisters wander.
func moveMonsters() {
	people := &area.people
	for i := 7; i >= 0; i-- {
		kind := int(people.tile[i])
		if kind == 0 {
			continue
		}
		if outOfWindow(i) {
			people.tile[i] = 0
			people.baseTile[i] = 0
			continue
		}
		var dirX, dirY, distX, distY, dist int
		if kind != tileWhirlpool && kind != tileTwister {
			dirX, dirY, distX, distY, dist = towardParty(i)
			if dist < 2 {
				redraw()
				printText(attackedBy)
				printText(creatureName(uint8(kind)))
				newLine()
				startCombat(i)
				continue
			}
		}
		if kind == tilePirateWest {
			if dist < 4 {
				pirateAttack(i, dirX, dirY)
			} else {
				chaseParty(i, dirX, dirY)
			}
			continue
		}
		breathesFire := kind == tileSeaSerpent || kind == tileLavaLizard || kind >= tileHydra
		if breathesFire && abs(distX) < 5 && abs(distY) < 5 && randomBits(1) != 0 {
			shootAtParty(i, dirX, dirY, tileHitFlash)
		}
		if kind == tileWhirlpool || kind == tileTwister || (isGoodCreature(uint8(kind)) && randomBits(3) != 0) {
			ry := randomSign()
			rx := randomSign()
			wanderOutdoors(i, rx, ry, 1)
		} else {
			wanderOutdoors(i, dirX, dirY, 2)
		}
	}
}

// movePeople moves the people of a town or the world.
func movePeople() {
	switch gameMode {
	case modeTown:
		moveTownPeople()
	case modeOutdoors:
		moveMonsters()
	}
}

// spawnMonsters now and then sets a new monster in one of the four first free places, out
// of the party's sight: at sea a sea monster or pirate, on land a monster the stronger the
// longer the game has gone on.
func spawnMonsters() {
	if gameMode != modeOutdoors || randomBits(15) != 0 {
		return
	}
	people := &area.people
	for i := 3; i >= 0; i-- {
		if people.tile[i] != 0 {
			continue
		}
		wx, wy := uint8(randomBits(31)), uint8(randomBits(31))
		tile := *area.townTile(int(wy), int(wx))
		if abs(int(wx)-int(windowPos.x)) < 6 || abs(int(wy)-int(windowPos.y)) < 6 {
			continue
		}
		worldX := uint8(int(wx) + int(windowQuarter.x)*quarterTiles)
		worldY := uint8(int(wy) + int(windowQuarter.y)*quarterTiles)
		switch {
		case tile < tileShallowWater: // at sea
			if randomBits(7) != 0 {
				continue
			}
			people.x[i], people.y[i] = worldX, worldY
			kind := uint8(tilePirateWest + randomBits(7)*2)
			if kind == tilePirateEast {
				kind = tilePirateWest
			}
			people.baseTile[i], people.tile[i] = kind, kind
		case tile >= tileGrass && tile < tileMountains: // on land
			people.x[i], people.y[i] = worldX, worldY
			mask := 15 // the kinds of land monsters that may appear, of 16
			switch {
			case party.moves < 10000:
				mask = 3
			case party.moves < 30000:
				mask = 7
			}
			kind := uint8(randomBits(randomBits(mask))*4 + tileOrc)
			people.baseTile[i], people.tile[i] = kind, kind
		}
	}
}
