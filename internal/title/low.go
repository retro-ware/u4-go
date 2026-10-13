package title

// Port of LOW.ASM: start-up (the random seed and the speed). The equipment flags, the Tandy
// memory patch, the adapter test, the saved video mode and interrupt vectors, and the clean-up
// that restored them have no meaning on the host and are dropped.

import (
	"time"

	"github.com/memmaker/u4-go/internal/host"
)

// speedInfo is the machine's speed (host.SpeedInfo): how many times some steps repeat.
var speedInfo uint16

// seedRandom seeds the random generator from the date and time, and sets the speed.
func seedRandom() { /*C_244E*/
	t := host.Now()
	put := func(i int, v uint16) {
		randomSeed[i], randomSeed[i+1] = byte(v), byte(v>>8)
		randomSeed[i+8], randomSeed[i+9] = byte(v), byte(v>>8)
	}
	put(0, uint16(t.Second())<<8|uint16(t.Nanosecond()/10000000))
	put(2, uint16(t.Hour())<<8|uint16(t.Minute()))
	put(4, uint16(t.Month())<<8|uint16(t.Day()))
	put(6, uint16(t.Year()))
	speedInfo = host.SpeedInfo
}

// ---- pacing ----

// Real-time lengths of the intro's steps. The calibrated waits (pause, the delay's cursor and
// the monsters) last the same on any machine; the uncalibrated work (drawing, the moongate's empty
// loop) is timed as it ran on the era's 386 (host.SpeedInfo).
const (
	delayPass    = 8400 * time.Nanosecond // one pass of pause's calibrated loop: 440 = a handwriting dot
	monsterFrame = 120 * time.Millisecond // animateMonsters: the two monsters move every 200 delay passes
	cursorPass   = 600 * time.Microsecond // the cursor turns every (1+cursorRate) delay passes
	viewStep     = 20 * time.Millisecond  // one move of the view: 1+speedInfo/2 redraws
	scrollStep   = 12 * time.Millisecond  // one step of the view's scroll-open (pause adds its own)
	dissolveStep = 10 * time.Millisecond  // one step of the "Ultima IV" dissolve (two dissolvePic)
	moongateStep = 5 * time.Millisecond   // one frame of the moongate: 10000 passes of an empty loop
)

func pace(d time.Duration) { host.Wait(d) }
