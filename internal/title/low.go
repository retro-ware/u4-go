package title

// Port of LOW.ASM: start-up (equipment flags, random seeds, the speed calibration), clean-up
// and the graphics adapter test. The Tandy memory patch, the saved video mode and the saved
// interrupt vectors have no meaning on the host and are dropped; low_gra answers EGA.

import (
	"time"

	"github.com/memmaker/u4-go/internal/host"
)

// SpeedInfo is what LOW.ASM's calibration loop measures: thousands of loop passes per 55 ms
// timer tick, about 1 on a 4.77 MHz PC. TITLE.EXE multiplies it into its delay loops and
// draws the view speed_info/2 extra times a frame; the port's cycle model (cpu) divides the
// cost of the work by it, so the pacing stays close to the original at any value. 10 matches
// AVATAR.EXE's port.
var SpeedInfo = 10

var (
	speed_info uint16      /*D_692E*/
	psp        [0x100]byte // far_psp (D_041C): the command line at 80h
)

func low_init() { /*C_244E*/
	t := time.Now()
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
	speed_info = uint16(SpeedInfo)
	if speed_info == 0 {
		speed_info = 1
	}
}

func low_clean() {} /*C_2551: video mode, Tandy memory, int 24h: nothing to restore*/

func low_gra() int16 { return 2 } /*C_2583: EGA*/

// ---- the cycle model ----

// cpu lets the time pass that the given 4.77 MHz-PC cycles of work took on a machine of
// speed speed_info. DOS drew straight to video memory; here the host shows the planes at
// each yield, so pacing the work is also what shows it.
func cpu(cycles int) {
	cpuCarry += float64(cycles) / (4.77 * float64(speed_info))
	tick := 1e6 / float64(host.TPS)
	for cpuCarry >= tick {
		cpuCarry -= tick
		host.Yield()
	}
}

var cpuCarry float64 // microseconds owed
