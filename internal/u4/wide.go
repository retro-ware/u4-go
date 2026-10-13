package u4

// The modern layout's wider map view: the original shows 11x11 tiles; this draws more of
// the map around them by the original's rules (drawOutdoors/drawTown for the terrain, placePeople for
// the people and monsters, castRays's rays for the line of sight). It only reads the game's
// state; the 11x11 middle is still the original's own picture, drawn over it by the host.

import "github.com/memmaker/u4-go/internal/host"

func init() { host.WideView = wideView }

// wideTile is the terrain at (dx, dy) from the party.
func wideTile(dx, dy int) uint8 {
	if gameMode == modeTown {
		x, y := int(party.x)+dx, int(party.y)+dy
		if x < 0 || x >= townTiles || y < 0 || y >= townTiles {
			return tileGrass
		}
		return *area.townTile(y, x)
	}
	// outdoors: the loaded 32x32 window, else WORLD.MAP itself (8x8 chunks of 32x32)
	if x, y := int(windowPos.x)+dx, int(windowPos.y)+dy; x >= 0 && x < windowTiles && y >= 0 && y < windowTiles {
		return *area.townTile(y, x)
	}
	return worldTile(uint8(int(party.x)+dx), uint8(int(party.y)+dy))
}

// chunksAcross is the world's width in chunks.
const chunksAcross = 8

// worldTile is WORLD.MAP's tile at (x, y), the world wrapping around.
func worldTile(x, y uint8) uint8 {
	chunk := int(y)/chunkTiles*chunksAcross + int(x)/chunkTiles
	return worldMap[chunk*chunkSize+int(y)%chunkTiles*chunkTiles+int(x)%chunkTiles]
}

// tilePixels is a tile's width and height in pixels.
const tilePixels = 16

// wideView returns the view cols x rows tiles around the party (both odd) as palette
// indexes, tilePixels a tile; nil when the screen is not the world or the town map.
func wideView(cols, rows int) []byte {
	if gameMode != modeOutdoors && gameMode != modeTown {
		return nil
	}
	cx, cy := cols/2, rows/2
	grid := make([]uint8, cols*rows)
	for y := 0; y < rows; y++ {
		for x := 0; x < cols; x++ {
			grid[y*cols+x] = wideTile(x-cx, y-cy)
		}
	}
	put := func(dx, dy int, t uint8) {
		if x, y := dx+cx, dy+cy; x >= 0 && x < cols && y >= 0 && y < rows {
			grid[y*cols+x] = t
		}
	}
	people := &area.people
	for p := len(people.baseTile) - 1; p >= 0; p-- {
		if people.baseTile[p] != 0 {
			put(signedByte(people.x[p]-party.x), signedByte(people.y[p]-party.y), people.baseTile[p])
		}
	}
	if moongateTile != 0 {
		put(signedByte(moongateAtX-party.x), signedByte(moongateAtY-party.y), moongateTile)
	}
	put(0, 0, uint8(party.tile))
	if lineOfSight != 0 {
		grid = sight(grid, cols, rows)
	}
	pix := make([]byte, cols*tilePixels*rows*tilePixels)
	w := cols * tilePixels
	for i, t := range grid {
		tile := &screen.Tiles[t]
		o := (i/cols)*tilePixels*w + (i%cols)*tilePixels
		for r := range tilePixels {
			copy(pix[o+r*w:o+r*w+tilePixels], tile[r*tilePixels:(r+1)*tilePixels])
		}
	}
	return pix
}

// sight is castRays on a cols x rows grid: what the rays from the middle do not reach is
// darkness.
func sight(grid []uint8, cols, rows int) []uint8 {
	seen := make([]uint8, len(grid))
	for i := range seen {
		seen[i] = tileBlank
	}
	cx, cy := cols/2, rows/2
	seen[cy*cols+cx] = grid[cy*cols+cx]
	var ray func(dx, dy, x, y int)
	ray = func(dx, dy, x, y int) { // castRay
		if x < 0 || x >= cols || y < 0 || y >= rows {
			return
		}
		i := y*cols + x
		if seen[i] != tileBlank {
			return
		}
		t := grid[i]
		seen[i] = t
		if t == tileForest || t == tileMountains || t == tileBlank || t == tileSecretDoor || t == tileBrickWall {
			return
		}
		x += dx
		y += dy
		ray(dx, dy, x, y)
		if dx&dy != 0 {
			ray(dx, dy, x, y-dy)
			ray(dx, dy, x-dx, y)
		} else {
			ray(isZero(dx)*dy+dx, dy-isZero(dy)*dx, dy+x, y-dx)
			ray(dx-isZero(dx)*dy, isZero(dy)*dx+dy, x-dy, dx+y)
		}
	}
	ray(0, -1, cx, cy-1)
	ray(0, 1, cx, cy+1)
	ray(-1, 0, cx-1, cy)
	ray(1, 0, cx+1, cy)
	ray(1, 1, cx+1, cy+1)
	ray(1, -1, cx+1, cy-1)
	ray(-1, 1, cx-1, cy+1)
	ray(-1, -1, cx-1, cy-1)
	return seen
}

// isZero is 1 for v == 0, else 0.
func isZero(v int) int { return b2i(v == 0) }
