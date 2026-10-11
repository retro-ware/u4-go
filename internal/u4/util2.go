package u4

// Port of UTIL2.ASM (keyboard, delays with the blinking cursor, the random generators) and
// LOW.ASM (start-up and the speed calibration).

import (
	"time"

	"github.com/memmaker/u4-go/internal/host"
)

// SpeedInfo is the speed_info LOW.ASM measured: the machine's speed in units of 1000 loops
// per timer tick (55 ms). The game multiplies it into its CPU-timed loops and divides its
// per-frame animations by it. The port runs t_callback (one redraw of the view, the unit the
// game's animations count in) at a fixed rate, CallbackTicks host ticks each, and picks the
// speed that makes the moons and the animations run at the original pace: a 4.77 MHz PC
// measured 1 and redrew about 5 times a second; the moon cycle shows the rate grows as
// speed_info/2+1.
var (
	SpeedInfo     = 10
	CallbackTicks = 2 // 70/2 = 35 redraws a second, 5.3*(10/2+1) as on a machine of speed 10
)

var (
	cursor_rate uint16 /*D_0672*/
	D_0674      uint16
	D_0676      uint16 = 0xffff // display cursor flag
	seed_128    [16]byte        // random seeds (D_8732)
	D_8C42      uint8           // cursor state {0, 1, 2, 3}
	D_8C43      uint16          // cursor update counter
)

// dosSeconds is DOS's clock (int 21h/2Ch) seconds, from the host clock.
func dosSeconds() int { return int(host.Clock/256) % 60 }

// u_delay waits bp04 seconds; with bp06 set a key ends the wait and the cursor blinks.
func u_delay(bp04, bp06 int) { /*C_16CD*/
	D_8C43 = 1
	D_0674 = cursor_rate
	si := dosSeconds() + bp04&0xffff
	check := func() bool { // C_1701: true when a key ended the wait
		return bp06 != 0 && u_kbhit() != 0
	}
	if si >= 60 {
		si -= 60
		if bp06 != 0 && u_kbhit() != 0 {
			return // C_176C: no cursor to delete
		}
	}
	for {
		if check() {
			break
		}
		t_callback()
		// cursor
		if uint16(bp06)&D_0676 != 0 {
			reload := false
			if D_0674 != cursor_rate {
				reload = true
			} else {
				D_8C43--
				if D_8C43 == 0 {
					D_8C42 = (D_8C42 - 1) & 3
					Gra_putchar(D_8C42 + 0x1c)
					reload = true
				}
			}
			if reload {
				D_8C43 = (1 + cursor_rate) * speed_info
				D_0674 = cursor_rate
			}
		}
		host.Busy()
		if dosSeconds() == si {
			break
		}
	}
	if uint16(bp06)&D_0676 != 0 {
		// delete cursor
		Gra_putchar(' ')
	}
}

// u_rand_a (C_1771): a 128-bit additive generator: an add-with-carry chain over the seed,
// then a 128-bit increment.
func u_rand_a() uint8 {
	s := &seed_128
	al := s[15]
	cf := uint16(0)
	for i := 14; i >= 1; i-- {
		v := uint16(al) + uint16(s[i]) + cf
		al = byte(v)
		cf = v >> 8
		s[i] = al
	}
	s[0] = byte(uint16(s[0]) + uint16(al) + cf)
	for i := 15; i >= 0; i-- {
		s[i]++
		if s[i] != 0 {
			break
		}
	}
	return s[0]
}

// u_rand_b (C_17A2). Its first adc takes the caller's carry; compiled C calls it after
// instructions that leave the carry clear, so it is an add.
func u_rand_b() uint8 {
	s := &seed_128
	a := s[3] + s[2]
	s[2] = a
	s[1] ^= a
	a = s[1] + s[0]
	a = a>>1 | a<<7
	s[0] = a
	s[3] = a
	return a
}

func U4_RND1(a int) int { return int(u_rand_a()) & a }
func U4_RND3(a int) int { return int(u_rand_a()) % a }
func U4_RND4(a int) int { return SafeModulo(int(u_rand_a()), a) }
func U4_RND2(a int) int { return int(u_rand_b()) & a }

// u_kbhit counts Party.f_000 and reports a waiting key (dropping ^C), running the view's
// animation when there is none (C_17C5).
func u_kbhit() int {
	u_rand_a()
	Party.f_000++
	if len(host.KeyQueue) > 0 {
		k := host.KeyQueue[0]
		if k != 0x2E03 && k != 0 {
			return 1
		}
		host.KeyQueue = host.KeyQueue[1:]
	}
	t_callback()
	host.Busy()
	return 0
}

// u_kbread waits for a key (blinking the cursor) and returns its BIOS word (C_17F4).
func u_kbread() int {
	u_delay(0x8081, 0x8081)
	for len(host.KeyQueue) == 0 {
		host.Yield()
	}
	k := host.KeyQueue[0]
	host.KeyQueue = host.KeyQueue[1:]
	return int(k)
}

// u_kbflush drops the waiting keys (C_1804).
func u_kbflush() {
	for u_kbhit() != 0 {
		u_kbread()
	}
}

// get current default drive (1 = A:)
func C_1814() int { return 3 }

// select disk
func C_181D(d int) {}

// test file
func C_182F(name string) int {
	if _, ok := readGameFile(name); ok {
		return 1
	}
	return 0
}

// piracy check (reads track 9 of the floppy): success
func C_184F() int { return 0 }

// set critical error handler
func C_18A2() {}

// ---- LOW.ASM ----

// psp holds the command line: _80[0] the length, _80[2] the graphics letter, _80[4] the
// drive digit ULTIMA.COM got from TITLE.EXE.
var psp [0x100]byte

func low_init() { /*C_0012*/
	equip_flags = 0x0001 // one floppy drive
	t := time.Now()
	dx := uint16(t.Second())<<8 | uint16(t.Nanosecond()/10000000)
	cx := uint16(t.Hour())<<8 | uint16(t.Minute())
	put := func(i int, v uint16) { seed_128[i], seed_128[i+1] = byte(v), byte(v>>8) }
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

func low_clean() {} /*C_0115*/

// determine graphic adapter type: EGA
func low_gra() int16 { return 2 } /*C_0147*/
