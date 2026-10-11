package u4

// The modern layout's wider map view: the original shows 11x11 tiles; this draws more of
// the map around them by the original's rules (C_3B35/C_3B83 for the terrain, C_3960 for
// the people and monsters, C_38BB's rays for the line of sight). It only reads the game's
// state; the 11x11 middle is still the original's own picture, drawn over it by the host.

import "github.com/memmaker/u4-go/internal/host"

func init() { host.WideView = wideView }

// wideTile is the terrain at (dx, dy) from the party.
func wideTile(dx, dy int) uint8 {
	if CurMode == MOD_BUILDING {
		x, y := int(Party._x)+dx, int(Party._y)+dy
		if x < 0 || x > 31 || y < 0 || y > 31 {
			return TIL_04
		}
		return *D_8742.x32x32(y, x)
	}
	// outdoors: the loaded 32x32 window, else WORLD.MAP itself (8x8 chunks of 32x32)
	if x, y := int(D_959C.x)+dx, int(D_959C.y)+dy; x >= 0 && x < 32 && y >= 0 && y < 32 {
		return *D_8742.x32x32(y, x)
	}
	x, y := (int(Party._x)+dx)&0xff, (int(Party._y)+dy)&0xff
	return File_MAP[((y>>5)*8+(x>>5))*0x400+(y&31)*32+x&31]
}

// wideView returns the view cols x rows tiles around the party (both odd) as palette
// indexes, 16 pixels a tile; nil when the screen is not the world or the town map.
func wideView(cols, rows int) []byte {
	if CurMode != MOD_OUTDOORS && CurMode != MOD_BUILDING {
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
	n := &D_8742._npc
	for b := 31; b >= 0; b-- {
		if n._gtile[b] != 0 {
			put(int(int8(n._x[b]-Party._x)), int(int8(n._y[b]-Party._y)), n._gtile[b])
		}
	}
	if D_9141 != 0 {
		put(int(int8(D_9445-Party._x)), int(int8(D_9448-Party._y)), D_9141)
	}
	put(0, 0, uint8(Party._tile))
	if D_9440 != 0 {
		grid = sight(grid, cols, rows)
	}
	pix := make([]byte, cols*16*rows*16)
	w := cols * 16
	for i, t := range grid {
		tile := &drv.Tiles[t]
		o := (i/cols)*16*w + (i%cols)*16
		for r := 0; r < 16; r++ {
			copy(pix[o+r*w:o+r*w+16], tile[r*16:r*16+16])
		}
	}
	return pix
}

// sight is C_38BB on a cols x rows grid: what the rays from the middle do not reach is
// darkness.
func sight(grid []uint8, cols, rows int) []uint8 {
	seen := make([]uint8, len(grid))
	for i := range seen {
		seen[i] = TIL_7E
	}
	cx, cy := cols/2, rows/2
	seen[cy*cols+cx] = grid[cy*cols+cx]
	var ray func(dx, dy, x, y int)
	ray = func(dx, dy, x, y int) { // C_376F
		if x < 0 || x >= cols || y < 0 || y >= rows {
			return
		}
		i := y*cols + x
		if seen[i] != TIL_7E {
			return
		}
		t := grid[i]
		seen[i] = t
		if t == TIL_06 || t == TIL_08 || t == TIL_7E || t == TIL_49 || t == TIL_7F {
			return
		}
		x += dx
		y += dy
		ray(dx, dy, x, y)
		if dx&dy != 0 {
			ray(dx, dy, x, y-dy)
			ray(dx, dy, x-dx, y)
		} else {
			ray(bnot(dx)*dy+dx, dy-bnot(dy)*dx, dy+x, y-dx)
			ray(dx-bnot(dx)*dy, bnot(dy)*dx+dy, x-dy, dx+y)
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
