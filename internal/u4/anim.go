package u4

// The view's redraw, run whenever the game waits: the animated tiles and sprites, the wind,
// the moons and moongates, and the line of sight.

import "github.com/memmaker/u4-go/internal/host"

// The view is the 11x11 tiles around the party, which stands at its centre.
const (
	viewSize   = 11
	viewCenter = 5
)

// Where the wind (or the dungeon direction) and the moons show, in the map's frame.
const (
	frameBottomRow  = 23
	directionColumn = 7
	frameTopRow     = 0
	moonsColumn     = 11
	moonGlyph       = 20 // the first of the font's eight moon phases
)

// countDown counts *c down, and when it has run out restarts it at reload and reports true.
func countDown(c *uint16, reload uint16) bool {
	if *c == 0 {
		*c = reload
		return true
	}
	*c--
	return false
}

var flowCountdown uint16

// animateTiles lets the water, lava and force fields flow, and the flames flicker.
func animateTiles() {
	if countDown(&flowCountdown, speedInfo-1) {
		for _, t := range []int{tileDeepWater, tileMediumWater, tileShallowWater, tileLava, tilePoisonField, tileFireField, tileSleepField, tileEnergyField} {
			flowTile(t)
		}
	}
	flickerFlames()
}

// showDirection shows label and a direction under the map; in a dungeon also the level.
func showDirection(dir uint16, label string) {
	savedX, savedY := textCol, textRow
	textRow, textCol = frameBottomRow, directionColumn
	printText(label)
	switch dir {
	case dirNorth:
		printText(" North")
	case dirSouth:
		printText(" South")
	case dirEast:
		printText("  East")
	case dirWest:
		printText("  West")
	}
	if gameMode == modeDungeon {
		textRow, textCol = frameTopRow, moonsColumn
		printChar('L')
		printChar(int(party.z) + '1')
	}
	textCol, textRow = savedX, savedY
}

var windCountdown uint16

// updateWind now and then turns the wind a quarter left or right, and shows it.
func updateWind() {
	if countDown(&windCountdown, speedInfo-1) {
		const shiftOdds = 252 // a random byte masked with this is 0 one time in 64
		if randomBits(shiftOdds) == 0 {
			turn := randomBits(2) - 1 // the mask gives 0 or 2: a turn of -1 or +1
			windDir = uint16(int(windDir)+turn+4) % 4
		}
	}
	showDirection(windDir, "Wind")
}

// animOddsBySpeed are the odds (of 256) that a sprite changes frame at a redraw, by the
// machine's speed.
var animOddsBySpeed = [7]int{192, 96, 60, 36, 20, 12, 8}

// animOdds are the odds (of 256) that a sprite changes its frame at a redraw. The original
// gives machines past its table 255, a new frame nearly every redraw, far too fast at the
// port's 35 redraws a second; animSpeed is the table's entry used instead (about four
// frame changes a second, a 386's look). Only the threshold differs: the same random
// numbers are drawn.
func animOdds() int {
	return animOddsBySpeed[min(speedInfo, animSpeed)-1]
}

// animSpeed is the calibration knob for the sprites' animation rate (1 fast .. 7 slow).
const animSpeed = 4

// animateSprites steps the people's and monsters' frames: people and sea monsters flip
// between two, land monsters cycle through four.
func animateSprites() {
	people := &area.people
	for i := range people.tile {
		tile := int(people.tile[i])
		twoFrames := (tile >= tileNixie && tile <= tileTwister) || (tile >= tileMage && tile <= tileShepherd) || (tile >= tileGuard && tile <= tileLordBritish)
		switch {
		case twoFrames:
			if animRandomBits(1) != 0 && animOdds() > animRandomBits(255) {
				frame := int(people.baseTile[i]) % 2
				people.baseTile[i] = uint8((tile | frame) ^ 1)
			}
		case tile >= tileRat:
			if animOdds() > animRandomBits(255) {
				frame := (int(people.baseTile[i]) + 1) % 4
				people.baseTile[i] = uint8(tile | frame)
			}
		case tile != tilePirateWest:
			people.baseTile[i] = uint8(tile)
		}
	}
	animateTiles()
	waveFlags()
}

func signOf(a int16) int16 {
	switch {
	case a > 0:
		return 1
	case a < 0:
		return -1
	}
	return 0
}

// drawView draws the view, with a missile or hit (hitTile) and the blinking cursor box.
func drawView() {
	i := int(hitY)*viewSize + int(hitX)
	under := viewTiles[i]
	if hitTile != 0 {
		viewTiles[i] = uint8(hitTile)
	}
	const viewScreenPos = 4*80 + 2 // the view's top left, as the screen driver counts
	drawMap(viewSize, viewSize, viewTiles[:], viewScreenPos)
	if cursorBoxX != -1 {
		// The box shows while the blink counts down from 4, and hides while it counts up
		// from -9.
		was := cursorBlink
		cursorBlink -= signOf(cursorBlink)
		if cursorBlink == 0 {
			if was < 0 {
				cursorBlink = 4
			} else {
				cursorBlink = -9
			}
		}
		if cursorBlink > 0 {
			drawCursorBox(int(cursorBoxY), int(cursorBoxX))
		}
	}
	viewTiles[i] = under
}

// isOpaque says whether tile blocks the line of sight.
func isOpaque(tile uint8) bool {
	switch tile {
	case tileForest, tileMountains, tileBlank, tileSecretDoor, tileBrickWall:
		return true
	}
	return false
}

// castRay reveals the view from (x, y) on in direction (dx, dy), until an opaque tile; it
// spreads, a diagonal ray to its two straight neighbours, a straight ray to the diagonals
// ahead of it. Positions off the view (past 10, or below 0, which wraps) end it.
func castRay(dx, dy int, x, y uint8) {
	if x >= viewSize || y >= viewSize {
		return
	}
	i := int(y)*viewSize + int(x)
	if viewTiles[i] != tileBlank {
		return // seen already
	}
	tile := battlefield.tiles[i]
	viewTiles[i] = tile
	if isOpaque(tile) {
		return
	}
	x += uint8(dx)
	y += uint8(dy)
	castRay(dx, dy, x, y)
	switch {
	case dx != 0 && dy != 0:
		castRay(dx, dy, x, y-uint8(dy))
		castRay(dx, dy, x-uint8(dx), y)
	case dx == 0:
		castRay(dy, dy, x+uint8(dy), y)
		castRay(-dy, dy, x-uint8(dy), y)
	default:
		castRay(dx, -dx, x, y-uint8(dx))
		castRay(dx, dx, x, y+uint8(dx))
	}
}

// castRays fills the view with what the party sees: everything blank, then rays cast in the
// eight directions from the centre.
func castRays() {
	for i := range viewTiles {
		viewTiles[i] = tileBlank
	}
	viewTiles[viewCenter*viewSize+viewCenter] = *combatTile(viewCenter, viewCenter)
	const c = viewCenter
	castRay(0, -1, c, c-1)
	castRay(0, 1, c, c+1)
	castRay(-1, 0, c-1, c)
	castRay(1, 0, c+1, c)
	castRay(1, 1, c+1, c+1)
	castRay(1, -1, c+1, c-1)
	castRay(-1, 1, c-1, c+1)
	castRay(-1, -1, c-1, c-1)
}

// viewOffset is where world position p lands in the view centred on the party's center;
// past the view's edge (also below 0, which wraps) it is viewSize or more.
func viewOffset(p, center uint8) uint8 { return p - center + viewCenter }

// placePeople sets the people in view onto the area around the party.
func placePeople() {
	people := &area.people
	for i := len(people.tile) - 1; i >= 0; i-- {
		x := viewOffset(people.x[i], party.x)
		y := viewOffset(people.y[i], party.y)
		if people.baseTile[i] != 0 && x < viewSize && y < viewSize {
			*combatTile(int(y), int(x)) = people.baseTile[i]
		}
	}
}

// drawSurroundings adds the people and the moongate to the area around the party, notes
// the tiles next to it, sets it at the centre and draws the view.
func drawSurroundings() {
	placePeople()
	if moongateTile != 0 {
		x := viewOffset(moongateAtX, party.x)
		y := viewOffset(moongateAtY, party.y)
		if x < viewSize && y < viewSize {
			*combatTile(int(y), int(x)) = moongateTile
		}
	}
	const c = viewCenter
	hereTile = *combatTile(c, c)
	northTile = *combatTile(c-1, c)
	southTile = *combatTile(c+1, c)
	eastTile = *combatTile(c, c+1)
	westTile = *combatTile(c, c-1)
	*combatTile(c, c) = uint8(party.tile)
	if lineOfSight != 0 {
		castRays()
	} else {
		viewTiles = battlefield.tiles
	}
	drawView()
}

// The moons: each clock's top three bits are its moon's phase (of eight). moonClock ticks
// the moons on: when it wraps Trammel moves on two steps, Felucca six.
var (
	moonClock    uint8
	trammelClock uint8
	feluccaClock uint8
)

const (
	moonPhaseSteps = 32 // clock steps a phase lasts
	moonPhases     = 8
)

// showMoongate puts the moongate of Trammel's phase there, and returns how far it has grown.
func showMoongate() uint8 {
	gate := trammelClock / moonPhaseSteps
	moongateAtX = moongateX[gate]
	moongateAtY = moongateY[gate]
	return moonClock / 64 // 0 to 3
}

var moonCountdown uint16

// updateMoons moves the moons on, grows or shrinks the moongate at the start or end of a
// phase, and shows the moons' phases.
func updateMoons() {
	if countDown(&moonCountdown, speedInfo/2) {
		moonClock += 64
		if moonClock == 0 {
			trammelClock += 2
			feluccaClock += 6
		}
		party.trammel = uint16(trammelClock / moonPhaseSteps)
		party.felucca = uint16(feluccaClock / moonPhaseSteps)
		switch trammelClock % moonPhaseSteps {
		case 0: // the phase began: the gate grows
			moongateTile = showMoongate() + tileMoongateOpening
		case moonPhaseSteps - 2: // the phase ends: the gate shrinks
			moongateTile = (3 - showMoongate()) + tileMoongateOpening
		}
	}
	savedX, savedY := textCol, textRow
	textRow, textCol = frameTopRow, moonsColumn
	putGlyph((trammelClock/moonPhaseSteps+moonPhases-1)%moonPhases + moonGlyph)
	textCol++
	putGlyph((feluccaClock/moonPhaseSteps+moonPhases-1)%moonPhases + moonGlyph)
	textCol, textRow = savedX, savedY
}

// drawOutdoors draws the view of the world map.
func drawOutdoors() {
	animateSprites()
	updateMoons()
	for y := 0; y < viewSize; y++ {
		for x := 0; x < viewSize; x++ {
			*combatTile(y, x) = *area.townTile(y+int(windowPos.y)-viewCenter, x+int(windowPos.x)-viewCenter)
		}
	}
	drawSurroundings()
}

// drawTown draws the view of a town, grass around it.
func drawTown() {
	animateSprites()
	for dy := -viewCenter; dy <= viewCenter; dy++ {
		for dx := -viewCenter; dx <= viewCenter; dx++ {
			x, y := int(party.x)+dx, int(party.y)+dy
			tile := uint8(tileGrass)
			if x >= 0 && x < townTiles && y >= 0 && y < townTiles {
				tile = *area.townTile(y, x)
			}
			*combatTile(dy+viewCenter, dx+viewCenter) = tile
		}
	}
	drawSurroundings()
	updateWind()
}

var balloonCountdown uint16

// driftBalloon lets the wind carry the flying balloon a tile now and then.
func driftBalloon() {
	if countDown(&balloonCountdown, speedInfo) {
		party.moves++
		switch windDir { // the wind blows from its direction
		case dirNorth:
			scrollSouth()
		case dirSouth:
			scrollNorth()
		case dirEast:
			scrollWest()
		case dirWest:
			scrollEast()
		}
	}
}

// combatAnimMasks are the masks of the random numbers that, being 0, step a fighter's frame,
// by the machine's speed (past the table: 15).
var combatAnimMasks = [8]int{1, 3, 3, 7, 7, 7, 7, 15}

// drawCombat draws the battlefield with the monsters and party members animated.
func drawCombat() {
	mask := 15
	if speedInfo < 8 {
		mask = combatAnimMasks[speedInfo-1]
	}
	animateTiles()
	viewTiles = battlefield.tiles
	if gameMode != modeShrine {
		for i := len(fighters.tile) - 1; i >= 0; i-- {
			tile := fighters.tile[i]
			if tile == 0 {
				continue
			}
			frame := &fighters.baseTile[i]
			switch {
			case fighters.sleeping[i] != 0:
			case tile >= tileRat: // four frames; a mimic stays a chest
				hiding := tile == tileMimic && *frame == tileChest
				if !hiding && randomBits(mask) != 0 && randomBits(mask) == 0 {
					*frame = (*frame+1)%4 | tile
				}
			case tile == tileAsleep:
				*frame = tileAsleep
			default: // two frames
				*frame ^= uint8(randomBits(animRandomBits(1)))
			}
			viewTiles[int(battlefield.monsterY[i])*viewSize+int(battlefield.monsterX[i])] = *frame
		}
		for i := int(party.size) - 1; i >= 0; i-- {
			tile := fighters.memberTile[i]
			if tile == 0 {
				continue
			}
			if tile == tileAsleep {
				memberFrame[i] = tile
			} else {
				memberFrame[i] ^= uint8(b2i(randomBits(mask) == 0))
			}
			viewTiles[int(battlefield.memberY[i])*viewSize+int(battlefield.memberX[i])] = memberFrame[i]
		}
	}
	drawView()
}

// redraw draws the view for the current mode. It is the one place the game redraws while
// it waits, so the port lets the frame show and paces the game here.
func redraw() {
	defer func() {
		for i := 0; i < CallbackTicks; i++ {
			host.Yield()
		}
	}()
	switch gameMode {
	case modeVision:
		cursorRate = 233
		return
	case modeOutdoors:
		cursorRate = 0
		if party.tile == tileBalloon && party.flyingOrLight != 0 {
			driftBalloon()
		}
		updateWind()
		drawOutdoors()
	case modeTown:
		cursorRate = 0
		drawTown()
	case modeDungeon:
		cursorRate = 0
		showDirection(party.dir, "DIR:")
		animateTiles()
		renderDungeon()
		drawDungeonSprites()
	default:
		cursorRate = 0
		drawCombat()
	}
}
