package title

// Port of LOW.ASM: start-up (equipment flags, random seeds, the speed calibration), clean-up
// and the graphics adapter test. The Tandy memory patch, the saved video mode and the saved
// interrupt vectors have no meaning on the host and are dropped; low_gra answers EGA.

import (
	"time"

	"github.com/memmaker/u4-go/internal/host"
)

var speed_info uint16 /*D_692E*/

func low_init() { /*C_244E*/
	t := host.Now()
	put := func(i int, v uint16) { seed_128[i], seed_128[i+1] = byte(v), byte(v>>8) }
	dx := uint16(t.Second())<<8 | uint16(t.Nanosecond()/10000000)
	cx := uint16(t.Hour())<<8 | uint16(t.Minute())
	put(0, dx)
	put(2, cx)
	put(8, dx)
	put(10, cx)
	dx = uint16(t.Month())<<8 | uint16(t.Day())
	cx = uint16(t.Year())
	put(4, dx)
	put(6, cx)
	put(12, dx)
	put(14, cx)
	speed_info = host.SpeedInfo
}

func low_clean() {} /*C_2551: video mode, Tandy memory, int 24h: nothing to restore*/

func low_gra() int16 { return 2 } /*C_2583: EGA*/

// ---- pacing ----

// Real-time lengths of the intro's steps. The calibrated waits (C_02A3, u_delay's cursor and
// monsters) last the same on any machine; the uncalibrated work (drawing, the moongate's empty
// loop) is timed as it ran on the era's 386 (host.SpeedInfo).
const (
	delayPass    = 8400 * time.Nanosecond // one pass of C_02A3's calibrated loop: 440 = a handwriting dot
	monsterFrame = 120 * time.Millisecond // t_callback: the two monsters move every 200 u_delay passes
	cursorPass   = 600 * time.Microsecond // the cursor turns every (1+cursor_rate) u_delay passes
	viewStep     = 20 * time.Millisecond  // one move of the view: 1+speed_info/2 redraws
	scrollStep   = 12 * time.Millisecond  // one step of the view's scroll-open (C_02A3 adds its own)
	dissolveStep = 10 * time.Millisecond  // one step of the "Ultima IV" dissolve (two Gra_B)
	moongateStep = 5 * time.Millisecond   // one frame of the moongate: 10000 passes of an empty loop
)

func pace(d time.Duration) { host.Wait(d) }
