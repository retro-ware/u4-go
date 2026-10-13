package u4

// The dungeon's first person view: walls, doors, ladders and holes drawn as lines and
// filled quads by depth (0, the party's room, to 3), and the monsters' and objects' sprites
// scaled by distance.

import (
	"time"

	"github.com/memmaker/u4-go/internal/host"
)

// The view is 176 pixels square, from 8 to 183; its centre is between 95 and 96.
const (
	viewMiddle = 95
	lineColor  = 3
)

// The square frame of the room at each depth (and of the vanishing point past it), the
// same for x and y: its near (left or top) and far (right or bottom) edge.
var (
	edgeNear = [6]int{8, 40, 72, 88, 95, 0}
	edgeFar  = [6]int{183, 151, 119, 103, 96, 0}
)

// A side door by depth: its two x edges (left wall's, right wall's), and its top and
// bottom, at its near and far edge.
var (
	leftDoorNearX  = [4]int{8, 48, 76, 90}
	leftDoorFarX   = [4]int{24, 64, 84, 92}
	rightDoorNearX = [4]int{183, 143, 114, 100}
	rightDoorFarX  = [4]int{167, 127, 106, 98}
	doorTopNear    = [4]int{52, 72, 86, 90}
	doorTopFar     = [4]int{61, 81, 91, 95}
	doorBottomNear = [4]int{183, 143, 115, 101}
	doorBottomFar  = [4]int{167, 127, 107, 98}
)

// A hole (or a ladder's opening) by depth: in the ceiling its near and far y, in the floor
// its near and far y, and its corners' x.
var (
	ceilingNearY  = [4]int{8, 48, 76, 90}
	ceilingFarY   = [4]int{24, 64, 84, 93}
	floorNearY    = [4]int{183, 143, 115, 101}
	floorFarY     = [4]int{167, 127, 107, 98}
	holeLeftNear  = [4]int{53, 72, 85, 92}
	holeLeftFar   = [4]int{60, 80, 90, 93}
	holeRightNear = [4]int{139, 120, 106, 98}
	holeRightFar  = [4]int{131, 112, 101, 97}
)

// A ladder by depth: its two rails' x, and its three rungs' y going up and going down.
var (
	ladderLeftX  = [4]int{72, 84, 92, 95}
	ladderRightX = [4]int{119, 107, 99, 96}
	upRungY      = [3][4]int{{16, 56, 80, 93}, {48, 72, 86, 94}, {80, 88, 92, 95}}
	downRungY    = [3][4]int{{175, 135, 111, 98}, {143, 119, 105, 97}, {111, 103, 99, 96}}
)

var trapCountdown, frameCountdown uint16

// The view as rendered last, to skip rendering it again when nothing changed.
var rendered struct {
	z          int16
	x, y       uint8
	dir        uint16
	dark       bool
	here       uint8
	ahead      uint8
	twoAhead   uint8
	threeAhead uint8
}

var (
	twoAheadTile, threeAheadTile uint8 // the rooms two and three ahead
	depth                        int   // the depth being drawn
	monsterFrame                 int   // the dungeon monsters' animation frame
)

// drawField draws the force field of kind field filling the room's frame at the depth.
func drawField(field uint8) {
	near, far := edgeNear[depth], edgeFar[depth]
	tiles := (far - near + 1) / 16
	fillTiles(tiles, tiles, tilePoisonField+int(field%4), near, near/4)
}

// drawHole draws a hole's outline at the depth, between near and far y.
func drawHole(nearY, farY [4]int) {
	e := depth
	leftNear, leftFar, rightNear, rightFar := holeLeftNear[e], holeLeftFar[e], holeRightNear[e], holeRightFar[e]
	drawLine(leftNear, nearY[e], leftFar, farY[e], lineColor)
	drawLine(leftNear+1, nearY[e], leftFar+1, farY[e], lineColor)
	drawLine(leftFar+1, farY[e], rightFar, farY[e], lineColor)
	drawLine(rightFar, farY[e], rightNear, nearY[e], lineColor)
	drawLine(rightFar-1, farY[e], rightNear-1, nearY[e], lineColor)
	drawLine(rightNear-1, nearY[e], leftNear+1, nearY[e], lineColor)
}

func drawCeilingHole() { drawHole(ceilingNearY, ceilingFarY) }
func drawFloorHole()   { drawHole(floorNearY, floorFarY) }

// contentsSprite is the sprite to draw for a room's contents: 0 for none (traps show as a
// hole now and then), -1 for a door, room or wall, which blocks the view. A field is drawn
// here, and slows the cursor by its distance.
func contentsSprite(room uint8) int {
	countDown(&trapCountdown, speedInfo-1)
	switch dngKind(room) {
	case dngChest:
		return tileChest
	case dngOrb:
		return tileMagicSphere
	case dngTrap:
		if trapCountdown == 1 && randomBits(63) == 0 && dngDetail(room) != 0 {
			if dngDetail(room) >= 8 {
				drawFloorHole()
			} else {
				drawCeilingHole()
			}
		}
		return 0
	case dngFountain:
		return tileShallowWater
	case dngField:
		drawField(room)
		cursorRate = [4]uint16{0, 1, 6, 10}[depth]
		return tileBarrier
	case dngAltar:
		return tileAltar
	}
	if room < dngDoor {
		return 0
	}
	return -1
}

// hasMonster says whether a monster is in the room (not one of the kinds whose lower bits
// mean something else).
func hasMonster(room uint8) bool {
	switch dngKind(room) {
	case dngTrap, dngFountain, dngField:
		return false
	}
	return room < dngRoom && dngDetail(room) != 0
}

// spritePause is the pause after drawing a sprite at each depth: the original counted down
// 1000, 1200, 2400 and 3000 passes per speed step (about 3 delay-loop passes of 55ms/8000).
var spritePause = [4]time.Duration{20625 * time.Microsecond, 24750 * time.Microsecond, 49500 * time.Microsecond, 61875 * time.Microsecond}

// drawSprite draws tile on the room's floor at the depth: at 3x, 2x, 1x and half size.
func drawSprite(tile int) {
	cursorRate = 0
	switch depth {
	case 0:
		for row := 0; row < 48; row++ {
			drawTileRow(tile, row/3, 112+row, 20, 2)
		}
	case 1:
		for row := 0; row < 32; row++ {
			drawTileRow(tile, row/2, 104+row, 20, 2)
		}
	case 2:
		fillTiles(1, 1, tile, 96, 22)
	case 3:
		for row := 0; row < 8; row++ {
			drawTileRow(tile, row*2, 96+row, 22, 1)
		}
	}
	host.Wait(spritePause[depth%4])
}

// drawRoomContents draws what room holds at depth d; it reports whether the view goes on
// past it.
func drawRoomContents(room uint8, d int) bool {
	depth = d
	if countDown(&frameCountdown, speedInfo-1) {
		monsterFrame = randomBits(3)
	}
	if hasMonster(room) {
		drawSprite(tileRat + int(dngDetail(room)-1)*4 + monsterFrame)
		return false
	}
	sprite := contentsSprite(room)
	if sprite > 0 {
		if sprite != tileBarrier {
			drawSprite(sprite)
		}
		return false
	}
	if sprite != 0 && (depth != 0 || room >= dngRoom) {
		return false
	}
	return true
}

// drawDungeonSprites draws the contents of the party's room and the three ahead, up to the
// first that blocks the view; in the dark, nothing.
func drawDungeonSprites() {
	if party.flyingOrLight == 0 {
		clearMap()
		return
	}
	_ = drawRoomContents(hereTile, 0) &&
		drawRoomContents(aheadTile, 1) &&
		drawRoomContents(twoAheadTile, 2) &&
		drawRoomContents(threeAheadTile, 3)
}

// dngTile is the room (x, y) on the party's level. Above the first level (Y-up cast there,
// the dungeon left at the next turn) the original read the bytes before the map, which the
// turn never uses: empty here.
func dngTile(y, x int) uint8 {
	if party.z < 0 {
		return 0
	}
	return *area.dungeonRoom(int(party.z), y, x)
}

// lookAround notes the rooms behind the party, its own and the three ahead.
func lookAround() {
	d := int(party.dir)
	behindTile = dngTile(behindY(int(party.y), d), behindX(int(party.x), d))
	hereTile = dngTile(int(party.y), int(party.x))
	x, y := aheadX(int(party.x), d), aheadY(int(party.y), d)
	aheadTile = dngTile(y, x)
	x, y = aheadX(x, d), aheadY(y, d)
	twoAheadTile = dngTile(y, x)
	x, y = aheadX(x, d), aheadY(y, d)
	threeAheadTile = dngTile(y, x)
}

// roomAt is the position of the room at the depth ahead of the party and column (-1 left,
// 0 ahead, +1 right).
func roomAt(column int) (x, y int) {
	x, y = int(party.x), int(party.y)
	switch party.dir {
	case dirNorth:
		y -= depth
		x += column
	case dirSouth:
		y += depth
		x -= column
	case dirEast:
		x += depth
		y += column
	case dirWest:
		x -= depth
		y -= column
	}
	return x & (levelSize - 1), y & (levelSize - 1)
}

func vertical(x, y0, y1 int)   { drawLine(x, y0, x, y1, lineColor) }
func horizontal(x0, y, x1 int) { drawLine(x0, y, x1, y, lineColor) }

// drawLadder draws a ladder through a hole at the depth: going up from the ceiling's near
// edge, down from the floor's; seen from its side (facing north or south) just a pole.
func drawLadder(up bool) {
	e := depth
	end, rungs := floorNearY[e], downRungY
	if up {
		end, rungs = ceilingNearY[e], upRungY
		drawCeilingHole()
	} else {
		drawFloorHole()
	}
	if party.dir == dirNorth || party.dir == dirSouth {
		vertical(viewMiddle, end, viewMiddle)
		vertical(viewMiddle+1, end, viewMiddle)
		return
	}
	from, to := end, viewMiddle-1
	if !up {
		from, to = viewMiddle, end
	}
	left, right := ladderLeftX[e], ladderRightX[e]
	vertical(left, from, to)
	vertical(left+1, from, to)
	vertical(right, from, to)
	vertical(right-1, from, to)
	for _, rung := range rungs {
		horizontal(left, rung[e], right)
	}
}

// The walls' colours: facing the party, and to its sides, differ by the way it faces.
func frontWallColor() int { return int(party.dir%2) + 1 }
func sideWallColor() int  { return b2i(party.dir%2 == 0) + 1 }

// drawFrontWall fills the wall ahead at the depth.
func drawFrontWall() {
	e := depth
	fillQuad(edgeNear[e], edgeFar[e]+1, edgeNear[e], edgeFar[e], edgeNear[e], edgeFar[e], frontWallColor(), 0)
}

// drawLeftFace and drawRightFace fill the wall facing the party in the room to the left or
// right one depth on (seen past an open side).
func drawLeftFace() {
	e := depth
	fillQuad(edgeNear[e], edgeNear[e+1], edgeNear[e+1], edgeFar[e+1], edgeNear[e+1], edgeFar[e+1], frontWallColor(), 0)
}

func drawRightFace() {
	e := depth
	fillQuad(edgeFar[e], edgeFar[e+1], edgeNear[e+1], edgeFar[e+1], edgeNear[e+1], edgeFar[e+1], frontWallColor(), 0)
}

// drawLeftWall and drawRightWall fill the side walls at the depth, in perspective.
func drawLeftWall() {
	e := depth
	fillQuad(edgeNear[e], edgeNear[e+1], edgeNear[e], edgeFar[e], edgeNear[e+1], edgeFar[e+1], sideWallColor(), 0)
}

func drawRightWall() {
	e := depth
	fillQuad(edgeFar[e], edgeFar[e+1], edgeNear[e], edgeFar[e], edgeNear[e+1], edgeFar[e+1], sideWallColor(), 0)
}

// drawFrontDoor draws a door in the wall ahead (in the party's own room drawn inverted).
func drawFrontDoor() {
	e := depth
	color, erase := lineColor, 0
	if e == 0 {
		color, erase = 0, lineColor
	}
	fillQuad(edgeNear[e+1], edgeFar[e+1], edgeNear[e+1], edgeFar[e], edgeNear[e+1], edgeFar[e], color, erase)
}

func drawLeftDoor() {
	e := depth
	fillQuad(leftDoorNearX[e], leftDoorFarX[e], doorTopNear[e], doorBottomNear[e], doorTopFar[e], doorBottomFar[e], lineColor, 0)
}

func drawRightDoor() {
	e := depth
	fillQuad(rightDoorNearX[e], rightDoorFarX[e], doorTopNear[e], doorBottomNear[e], doorTopFar[e], doorBottomFar[e], lineColor, 0)
}

// drawSide draws one side at the depth: a door in a wall, a wall, or the face of the open
// room beside.
func drawSide(column int, wall, door, face func()) {
	x, y := roomAt(column)
	switch dngKind(dngTile(y, x)) {
	case dngDoor, dngRoom:
		wall()
		door()
	case dngSecretDoor, dngWall:
		wall()
	default:
		face()
	}
}

// drawSides draws the left side then the right at the depth.
func drawSides() {
	drawSide(-1, drawLeftWall, drawLeftDoor, drawLeftFace)
	drawSide(1, drawRightWall, drawRightDoor, drawRightFace)
}

// drawCorridor draws the view ahead, depth by depth, up to the first door or wall.
func drawCorridor() {
	for depth = 0; depth <= 3; depth++ {
		x, y := roomAt(0)
		room := dngTile(y, x)
		switch dngKind(room) {
		case dngLadderUpDown:
			drawLadder(false)
			drawLadder(true)
			drawSides()
		case dngLadderUp:
			drawLadder(true)
			drawSides()
		case dngCorridor, dngChest, dngOrb, dngTrap, dngFountain, dngAltar:
			drawSides()
		case dngLadderDown:
			drawLadder(false)
			drawSides()
		case dngCeilingHole:
			drawCeilingHole()
			drawSides()
		case dngFloorHole:
			drawFloorHole()
			drawSides()
		case dngDoor:
			drawFrontWall()
			drawFrontDoor()
			if depth != 0 {
				return
			}
		case dngRoom:
			drawFrontWall()
			drawFrontDoor()
			return
		case dngField:
			drawField(room)
			return
		case dngSecretDoor, dngWall:
			drawFrontWall()
			return
		}
	}
	depth = 4
	x, y := roomAt(0)
	if kind := dngKind(dngTile(y, x)); kind == dngWall || kind == dngSecretDoor {
		drawFrontWall()
	}
}

// renderView notes what it renders and renders the view (in the dark, nothing).
func renderView() {
	rendered.z, rendered.x, rendered.y, rendered.dir = party.z, party.x, party.y, party.dir
	rendered.dark = party.flyingOrLight == 0
	rendered.here, rendered.ahead = hereTile, aheadTile
	rendered.twoAhead, rendered.threeAhead = twoAheadTile, threeAheadTile
	forceRender = 0
	clearMap()
	if party.flyingOrLight != 0 {
		drawCorridor()
	}
}

// renderDungeon renders the view, unless nothing changed and no sprite or field in view
// needs drawing again.
func renderDungeon() {
	if party.flyingOrLight != 0 {
		cursorRate = 13
	} else {
		cursorRate = 1
	}
	lookAround()
	same := forceRender == 0 && rendered.z == party.z && rendered.x == party.x && rendered.y == party.y &&
		rendered.dir == party.dir && rendered.here == hereTile && rendered.dark == (party.flyingOrLight == 0)
	if same {
		if k := dngKind(hereTile); k == dngField || hereTile >= dngRoom {
			return
		}
		if aheadTile == rendered.ahead {
			if k := dngKind(aheadTile); k == dngField || aheadTile >= dngDoor {
				return
			}
			if twoAheadTile == rendered.twoAhead {
				return // (the original compared the third room with itself)
			}
		}
	}
	renderView()
}

// The modern layout's 3D view reads the level, the party and the light from here.
func init() { host.DungeonView = dungeonView }

func dungeonView() *host.Dungeon {
	if gameMode != modeDungeon {
		return nil
	}
	d := &host.Dungeon{X: int(party.x), Y: int(party.y), Dir: int(party.dir),
		Lit: party.flyingOrLight != 0, Frame: monsterFrame, Tiles: &screen.Tiles}
	level := int(party.z) & (dungeonLevels - 1)
	copy(d.Map[:], area.tiles[level*levelSize*levelSize:])
	return d
}
