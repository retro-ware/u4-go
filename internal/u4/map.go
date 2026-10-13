package u4

// The outdoor map window and moving on foot, on horseback, by ship and by balloon.
//
// The world is 256x256 tiles, stored in WORLD.MAP as 8x8 chunks of 32x32 tiles. Outdoors the
// map window (area.tiles, 32x32) holds four 16x16 quarters of the world around the party,
// and scrolls by a quarter when the party comes near its edge.

const (
	quarterTiles  = 16 // a quarter of a chunk is 16x16 tiles
	worldQuarters = 16 // the world is 16x16 quarters, wrapping round
	windowTiles   = 32 // the map window is 32x32 tiles, two by two quarters

	scrollLow  = 5  // the window scrolls when the party's place in it falls below this
	scrollHigh = 27 // or reaches this

	townTiles = 32 // a town is 32x32 tiles; walking off its edge leaves it
)

// Special places of the world map.
const (
	hythlothX, hythlothY = 239, 240 // leaving the dungeon Hythloth, the party finds
	balloonX, balloonY   = 233, 242 // the balloon waiting here

	daemonX                          = 231 // the daemons guarding the Isle of the Abyss appear here
	daemonsWestX, daemonsEastX       = 229, 234
	daemonsNorthY, daemonsSouthY     = 212, 217
	piratesTriggerX, piratesTriggerY = 221, 224 // stepping east onto this spot brings pirates
)

var cachedChunkX, cachedChunkY = -1, -1 // the chunk loadChunk loaded last

// loadChunk loads the world chunk (cy, cx) from WORLD.MAP, dropping the one loaded before.
func loadChunk(cy, cx int) {
	if cachedChunkX != -1 {
		loadedChunks[cachedChunkY][cachedChunkX] = nil
	}
	if loadedChunks[cy][cx] == nil {
		m := new(tMap32x32)
		chunk := worldMap[(cy*8+cx)*chunkSize:]
		for y := range m {
			copy(m[y][:], chunk[y*chunkTiles:])
		}
		loadedChunks[cy][cx] = m
	}
	cachedChunkX, cachedChunkY = cx, cy
}

// copyQuarter copies the world's quarter (qy, qx) into the window at tile (wy, wx).
func copyQuarter(qy, qx, wy, wx int) {
	loadChunk(qy/2, qx/2)
	chunk := loadedChunks[qy/2][qx/2]
	top, left := qy%2*quarterTiles, qx%2*quarterTiles
	for y := 0; y < quarterTiles; y++ {
		for x := 0; x < quarterTiles; x++ {
			*area.townTile(wy+y, wx+x) = chunk[top+y][left+x]
		}
	}
}

// nextQuarter is the quarter after q, wrapping round the world.
func nextQuarter(q uint8) int { return (int(q) + 1) % worldQuarters }

// previousQuarter is the quarter before q, wrapping round the world.
func previousQuarter(q uint8) uint8 { return (q + worldQuarters - 1) % worldQuarters }

func loadQuarterNorthWest() {
	copyQuarter(int(windowQuarter.y), int(windowQuarter.x), 0, 0)
}
func loadQuarterNorthEast() {
	copyQuarter(int(windowQuarter.y), nextQuarter(windowQuarter.x), 0, quarterTiles)
}
func loadQuarterSouthWest() {
	copyQuarter(nextQuarter(windowQuarter.y), int(windowQuarter.x), quarterTiles, 0)
}
func loadQuarterSouthEast() {
	copyQuarter(nextQuarter(windowQuarter.y), nextQuarter(windowQuarter.x), quarterTiles, quarterTiles)
}

// centerMap fills the window around the party (or, inside a place, around its entrance).
func centerMap() {
	x, y := party.x, party.y
	if party.loc >= locFirstDungeon {
		x, y = party.outX, party.outY
	}
	windowQuarter.x, windowPos.x = x/quarterTiles, x%quarterTiles
	if windowPos.x < quarterTiles/2 {
		windowPos.x += quarterTiles
		windowQuarter.x = previousQuarter(windowQuarter.x)
	}
	windowQuarter.y, windowPos.y = y/quarterTiles, y%quarterTiles
	if windowPos.y < quarterTiles/2 {
		windowPos.y += quarterTiles
		windowQuarter.y = previousQuarter(windowQuarter.y)
	}
	loadQuarterNorthWest()
	loadQuarterNorthEast()
	loadQuarterSouthWest()
	loadQuarterSouthEast()
}

// leaveTown takes the party out of a town, castle or village back to the world map.
func leaveTown() {
	printText("Leaving...\n")
	gameMode = modeVision
	LoadRec("OUTMONST.SAV", &area.people)
	talkFile = nil
	party.x, party.y = party.outX, party.outY
	centerMap()
	gameMode = modeOutdoors
	party.loc = 0
	party.flyingOrLight = 0
	lineOfSight = 1
	if party.x == hythlothX && party.y == hythlothY {
		const slot = 31 // the last of the world's people
		area.people.baseTile[slot] = tileBalloon
		area.people.tile[slot] = tileBalloon
		area.people.x[slot] = balloonX
		area.people.y[slot] = balloonY
	}
	flushKeys()
}

// spawnDaemons sets eight daemons on the party, unless the Silver Horn holds them off.
func spawnDaemons() {
	if activeSpell != effectHorn {
		for i := 7; i >= 0; i-- {
			area.people.baseTile[i] = tileDaemon
			area.people.tile[i] = tileDaemon
			area.people.x[i] = daemonX
			area.people.y[i] = party.y + 1
		}
	}
}

// The pirate fleet: where the eight ships appear and which way they face.
var (
	pirateX    = [8]uint8{224, 224, 226, 227, 228, 229, 229, 228}
	pirateY    = [8]uint8{220, 228, 220, 228, 227, 225, 223, 222}
	pirateTile = [8]uint8{tilePirateEast, tilePirateEast, tilePirateEast, tilePirateEast, tilePirateSouth, tilePirateSouth, tilePirateNorth, tilePirateNorth}
)

// spawnPirates sets the pirate fleet on the map.
func spawnPirates() {
	for i := 7; i >= 0; i-- {
		area.people.x[i] = pirateX[i]
		area.people.y[i] = pirateY[i]
		area.people.tile[i] = tilePirateWest
		area.people.baseTile[i] = pirateTile[i]
	}
}

// scrollNorth moves the party a tile north on the world map, scrolling the window when needed.
func scrollNorth() {
	party.y--
	windowPos.y--
	if windowPos.y < scrollLow {
		windowPos.y += quarterTiles
		windowQuarter.y = previousQuarter(windowQuarter.y)
		half := quarterTiles * windowTiles
		copy(area.tiles[half:2*half], area.tiles[:half])
		loadQuarterNorthWest()
		loadQuarterNorthEast()
	}
}

// scrollSouth moves the party a tile south on the world map.
func scrollSouth() {
	party.y++
	windowPos.y++
	if windowPos.y >= scrollHigh {
		windowPos.y -= quarterTiles
		windowQuarter.y = uint8(nextQuarter(windowQuarter.y))
		half := quarterTiles * windowTiles
		copy(area.tiles[:half], area.tiles[half:2*half])
		loadQuarterSouthWest()
		loadQuarterSouthEast()
	}
}

// scrollWest moves the party a tile west on the world map.
func scrollWest() {
	party.x--
	windowPos.x--
	if windowPos.x < scrollLow {
		windowPos.x += quarterTiles
		windowQuarter.x = previousQuarter(windowQuarter.x)
		for y := 0; y < windowTiles; y++ {
			row := area.tiles[y*windowTiles : (y+1)*windowTiles]
			copy(row[quarterTiles:], row[:quarterTiles])
		}
		loadQuarterNorthWest()
		loadQuarterSouthWest()
	}
}

// scrollEast moves the party a tile east on the world map.
func scrollEast() {
	party.x++
	windowPos.x++
	if windowPos.x >= scrollHigh {
		windowPos.x -= quarterTiles
		windowQuarter.x = uint8(nextQuarter(windowQuarter.x))
		for y := 0; y < windowTiles; y++ {
			row := area.tiles[y*windowTiles : (y+1)*windowTiles]
			copy(row[:quarterTiles], row[quarterTiles:])
		}
		loadQuarterNorthEast()
		loadQuarterSouthEast()
	}
}

// walkableTiles are the tiles the party can walk onto.
var walkableTiles = []uint8{
	tileSwamp, tileGrass, tileScrub, tileForest, tileHills, tileDungeon, tileTowne, tileCastle, tileVillage,
	tileShipWest, tileShipNorth, tileShipEast, tileShipSouth, tileHorseWest, tileHorseEast, tileTiledFloor, tileBridge, tileBalloon,
	tileBridgeNorth, tileBridgeSouth, tileLadderUp, tileLadderDown, tileRuins, tileShrine, tileChest, tileBrickFloor, tileWoodFloor,
	tileMoongate, tilePoisonField, tileFireField, tileSleepField, tileSecretDoor, tileAltar, tileLava, tileTwister, tileTwister2,
}

// isWalkable says whether the party can walk onto tile.
func isWalkable(tile uint8) bool {
	for _, t := range walkableTiles {
		if t == tile {
			return true
		}
	}
	return false
}

func sayBlocked() {
	printText("Blocked!\n")
	sound(soundBeep)
	flushKeys()
}

func sayDriftOnly() {
	printText("Drift Only!\n")
	flushKeys()
}

// slowProgress says whether rough terrain holds the party back this turn: swamp one turn in
// eight, scrub and forest one in four, hills and fire half the time.
func slowProgress(tile uint8) bool {
	switch tile {
	case tileSwamp:
		return randomBits(7) == 0
	case tileScrub, tileForest:
		return randomBits(3) == 0
	case tileHills, tileFireField:
		return randomBits(1) != 0
	}
	return false
}

// isSailable says whether a ship can sail onto tile: deep or medium water, or a whirlpool or
// twister.
func isSailable(tile uint8) bool {
	return tile < tileShallowWater || (tile >= tileWhirlpool && tile < tileRat)
}

// windAllows says whether a ship makes way toward dir this turn: sailing with the wind only
// every fourth turn, against it three turns in four.
func windAllows(dir uint16) bool {
	switch windDir {
	case dir:
		return party.moves%4 == 0
	case (dir + 2) % 4:
		return party.moves%4 != 0
	}
	return true
}

// enterMoongate takes the party through a moongate when tile is one: to the gate of the
// moon Felucca shows, or with both moons new into the Shrine of Spirituality.
func enterMoongate(tile uint8) bool {
	if tile != tileMoongate {
		return false
	}
	const newMoon = 4
	if party.trammel == newMoon && party.felucca == newMoon {
		redraw()
		teleportOut()
		party.loc = locShrineSpirituality
		sound(soundSweep, 160)
		enterShrine()
		return true
	}
	redraw()
	party.x = moongateX[party.felucca]
	party.y = moongateY[party.felucca]
	teleportOut()
	sound(soundSweep, 160)
	centerMap()
	redraw()
	teleportIn()
	return true
}

// leftTown says whether a step in a town to windowPos (which wraps below 0) left it,
// leaving it then.
func leftTown(pos uint8) bool {
	if pos < townTiles {
		return false
	}
	leaveTown()
	return true
}

// step moves the party one tile onto tile, scrolling outdoors or moving in a town; it
// reports whether the party may take another step (galloping).
func step(tile uint8, scroll func(), moveInTown func() uint8) bool {
	if !isWalkable(tile) {
		sayBlocked()
		return false
	}
	return stepOnto(tile, scroll, moveInTown)
}

// stepOnto is step past the walkability check.
func stepOnto(tile uint8, scroll func(), moveInTown func() uint8) bool {
	if slowProgress(tile) {
		saySlowProgress()
		return true
	}
	sound(soundStep)
	if gameMode == modeOutdoors {
		scroll()
		return !enterMoongate(tile)
	}
	return !leftTown(moveInTown())
}

func stepNorth() bool {
	// Lord British's castle: its middle wing can only be entered from the south.
	if hereTile == tileCastleBritanniaEntrance {
		sayBlocked()
		return false
	}
	if northTile != tileCastleBritanniaEntrance && !isWalkable(northTile) {
		sayBlocked()
		return false
	}
	return stepOnto(northTile, scrollNorth, func() uint8 { party.y--; windowPos.y = party.y; return windowPos.y })
}

func stepSouth() bool {
	return step(southTile, scrollSouth, func() uint8 { party.y++; windowPos.y = party.y; return windowPos.y })
}

func stepWest() bool {
	return step(westTile, scrollWest, func() uint8 { party.x--; windowPos.x = party.x; return windowPos.x })
}

func stepEast() bool {
	return step(eastTile, func() {
		scrollEast()
		if party.y == piratesTriggerY && party.x == piratesTriggerX {
			spawnPirates()
		}
	}, func() uint8 { party.x++; windowPos.x = party.x; return windowPos.x })
}

// walk takes a step, and on a galloping horse's turn a second one.
func walk(stepOnce func() bool) {
	sound(soundStep)
	if stepOnce() && gallopStep != 0 {
		redraw()
		sound(soundStep)
		stepOnce()
	}
}

// sail turns the ship to face (ship) or, facing that way already, sails a tile.
func sail(ship uint16, dir uint16, name string, ahead uint8, scroll func()) {
	if party.tile != ship {
		party.tile = ship
		printText("Turn " + name + "!\n")
		return
	}
	printText("Sail " + name + "!\n")
	switch {
	case !isSailable(ahead):
		sayBlocked()
	case !windAllows(dir):
		saySlowProgress()
	default:
		scroll()
	}
}

// isShip says whether tile is the party's ship (facing any way).
func isShip(tile uint16) bool { return tile >= tileShipWest && tile <= tileShipSouth }

func cmdNorth() {
	switch {
	case isShip(party.tile):
		sail(tileShipNorth, dirNorth, "North", northTile, scrollNorth)
	case party.tile == tileBalloon:
		sayDriftOnly()
	default:
		walk(stepNorth)
	}
}

func cmdSouth() {
	switch {
	case party.tile == tileBalloon:
		sayDriftOnly()
	case isShip(party.tile):
		sail(tileShipSouth, dirSouth, "South", southTile, scrollSouth)
	default:
		walk(stepSouth)
		if party.x >= daemonsWestX && party.x < daemonsEastX && party.y >= daemonsNorthY && party.y < daemonsSouthY {
			spawnDaemons()
		}
	}
}

func cmdWest() {
	switch {
	case party.tile == tileBalloon:
		sayDriftOnly()
	case isShip(party.tile):
		sail(tileShipWest, dirWest, "West", westTile, scrollWest)
	default:
		if party.tile == tileHorseEast {
			party.tile = tileHorseWest
		}
		walk(stepWest)
	}
}

func cmdEast() {
	switch {
	case party.tile == tileBalloon:
		sayDriftOnly()
	case isShip(party.tile):
		sail(tileShipEast, dirEast, "East", eastTile, func() {
			scrollEast()
			if party.y == piratesTriggerY && party.x == piratesTriggerX {
				spawnPirates()
			}
		})
	default:
		if party.tile == tileHorseWest {
			party.tile = tileHorseEast
		}
		walk(stepEast)
	}
}
