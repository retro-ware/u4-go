package u4

import (
	"math"
	"math/rand/v2"

	"github.com/memmaker/u4-go/internal/ega"
)

// The moongate's pillar of light, in place of the original inverted-screen flashes: a beam
// drops onto the party, swells into a blue-white column and floods the map window white;
// on arrival the white draws back into the column, which thins out and dissolves into
// sparkles over the new map.

const (
	tpX0, tpY0 = 8, 8   // the map window
	tpSize     = 176    // its width and height
	tpCX       = 8 + 88 // the party's tile, centre
	tpCY       = 8 + 88
)

// tpRand is the effect's own generator, so the game's random numbers stay untouched.
// Its fixed seed (the letters of "Moongate") makes the sparkle the same every time.
var tpRand = rand.New(rand.NewPCG(1299148654, 1734440037))

// teleportOut plays the departure: beam, column, white-out (left on screen).
func teleportOut() {
	snap := tpSnapshot()
	const drop, swell = 8, 16
	for f := 1; f <= drop; f++ { // the beam falls from the top to the party
		bottom := tpY0 + (tpCY+8-tpY0)*f/drop
		tpPillar(snap, 3, bottom, 0)
		present()
	}
	for f := 1; f <= swell; f++ { // it swells into a column, growing ever brighter
		t := float64(f) / swell
		tpPillar(snap, 3+t*t*95, tpY0+tpSize, t*0.35)
		present()
	}
	tpFill(15)
	present()
}

// teleportIn plays the arrival over the map now drawn in the window, starting from white.
func teleportIn() {
	snap := tpSnapshot()
	tpFill(15)
	present()
	present()
	const shrink, fade = 16, 12
	for f := shrink; f >= 1; f-- {
		t := float64(f) / shrink
		tpPillar(snap, 3+t*t*95, tpY0+tpSize, t*0.35)
		present()
	}
	for f := 1; f <= fade; f++ { // the last of the light scatters into sparkles
		tpRestore(snap)
		left := 1 - float64(f)/fade
		n := int(left * 140)
		for i := 0; i < n; i++ {
			x := tpCX + int(tpRand.NormFloat64()*10*(1+2*(1-left)))
			y := tpY0 + tpRand.IntN(tpSize)
			tpSparkle(x, y, left)
		}
		present()
	}
	tpRestore(snap)
	present()
}

// tpPillar draws the column over snap: half width w (pixels), from the top down to bottom,
// with a glow over the rest of the window (0..1).
func tpPillar(snap []byte, w float64, bottom int, glow float64) {
	white, cyan, light, blue := screen.Color(15), screen.Color(11), screen.Color(9), screen.Color(1)
	for y := tpY0; y < tpY0+tpSize; y++ {
		row := y * ega.W
		for x := tpX0; x < tpX0+tpSize; x++ {
			i := row + x
			c := snap[(y-tpY0)*tpSize+x-tpX0]
			d := math.Abs(float64(x-tpCX)+0.5) / w
			// a slow shimmer along the column's edge
			d += 0.08 * math.Sin(float64(y)*0.35+float64(x)*0.1+tpRand.Float64())
			switch {
			case y > bottom:
			case d < 0.35:
				c = white
			case d < 0.6:
				c = cyan
			case d < 0.85:
				c = light
			case d < 1.1:
				if tpRand.IntN(2) == 0 {
					c = blue
				}
			case tpRand.Float64() < 0.12/d:
				c = light // flecks of light thrown off the column
			}
			if glow > 0 && c != white && tpRand.Float64() < glow {
				c = cyan
			}
			screen.Pix[i] = c
		}
	}
	// a halo of light where the beam meets the ground
	if bottom < tpY0+tpSize {
		for i := 0; i < 24; i++ {
			tpSparkle(tpCX+tpRand.IntN(17)-8, bottom-tpRand.IntN(6), 1)
		}
	}
}

// tpSparkle draws a small cross of light at (x, y), dimmer as bright falls to 0.
func tpSparkle(x, y int, bright float64) {
	c := screen.Color(15)
	switch {
	case bright < 0.33:
		c = screen.Color(9)
	case bright < 0.66:
		c = screen.Color(11)
	}
	for _, p := range [][2]int{{0, 0}, {-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
		px, py := x+p[0], y+p[1]
		if px >= tpX0 && px < tpX0+tpSize && py >= tpY0 && py < tpY0+tpSize {
			screen.Pix[py*ega.W+px] = c
		}
	}
}

func tpSnapshot() []byte {
	s := make([]byte, tpSize*tpSize)
	for y := 0; y < tpSize; y++ {
		copy(s[y*tpSize:], screen.Pix[(tpY0+y)*ega.W+tpX0:][:tpSize])
	}
	return s
}

func tpRestore(s []byte) {
	for y := 0; y < tpSize; y++ {
		copy(screen.Pix[(tpY0+y)*ega.W+tpX0:], s[y*tpSize:][:tpSize])
	}
}

func tpFill(c byte) {
	for y := tpY0; y < tpY0+tpSize; y++ {
		row := screen.Pix[y*ega.W+tpX0:][:tpSize]
		for i := range row {
			row[i] = screen.Color(c)
		}
	}
}
