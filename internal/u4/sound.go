package u4

// Port of U4_SOUND.ASM: thirteen PC speaker effects. Each is described as square-wave half
// periods (internal/speaker); their lengths come from the original's delay-loop counts: a
// count of n passes weighing w simple dec/jnz passes each lasts n*w*passUs/SpeedInfo
// microseconds (the calibration made speed_info*8000 passes last one 55 ms timer tick). The
// game waits for the effect, as the original did.

import (
	"time"

	"github.com/memmaker/u4-go/internal/host"
	"github.com/memmaker/u4-go/internal/speaker"
)

// passUs is one calibrated delay-loop pass at speed 1: 55 ms / 8000.
const passUs = 55000.0 / 8000

// fx collects an effect's notes.
type fx []speaker.Note

// half adds a half period of n passes (0 means 65536, as dec/jnz from 0 does) of weight w,
// merging equal halves into one longer note.
func (e *fx) half(n uint16, w float64) {
	c := float64(n)
	if n == 0 {
		c = 65536
	}
	h := speaker.Half(c * w * passUs / host.SpeedInfo)
	if l := len(*e); l > 0 && (*e)[l-1].Hz == h.Hz {
		(*e)[l-1].Dur += h.Dur
		return
	}
	*e = append(*e, h)
}

// waitFor lets d pass on the host clock.
func waitFor(d time.Duration) { host.Wait(d) }

func mul8(a, b int) uint16 { return uint16(byte(a)) * uint16(byte(b)) }

// sound plays effect n (cdecl: the effect, then the parameter some effects take).
func sound(n int, args ...int) { /*C_1D47*/
	cl := 0
	if len(args) > 0 {
		cl = args[0] & 0xff
	}
	if SoundFlag == 0 || uint(n)&0xff >= 0xd {
		return
	}
	waitFor(speaker.Play(effect(n, cl)))
}

// effect describes effect n with parameter cl.
func effect(n, cl int) fx {
	S := host.SpeedInfo
	e := &fx{}
	switch n & 0xff {
	case 0: // C_1D69: a click (an uncalibrated 32h passes)
		e.half(0x32, 1)
	case 1: // sound_1: the beep (TITLE.EXE's too)
		*e = append(*e, speaker.Beep...)
	case 2: // C_1DA8: 48 halves of 0E0h passes (325 Hz, 74 ms), then the beep
		for i := 0; i < 0x30; i++ {
			e.half(uint16(0xe0*S), 1)
		}
		*e = append(*e, speaker.Beep...)
	case 3: // C_1EB3: falling
		for bl := 5; bl < 0x100; bl++ {
			e.half(mul8(S, bl)>>1, 1.3)
		}
	case 4: // C_1EFD
		bl := 0
		for {
			e.half(mul8(S, bl)>>1, 1.3)
			bl = (bl - 1) & 0xff
			if bl < 0x80 {
				break
			}
		}
	case 5: // C_1F4A
		for bl := 0x80; bl < 0x100; bl++ {
			e.half(mul8(S, bl)>>1, 1.6)
		}
	case 6: // C_1E7F: noise
		for bh := 0; bh < 0xff; bh++ {
			r := uint16(u_rand_a()&0x7f | 1)
			cx := uint16(int(r*3) * S)
			e.half(cx>>2|1, 1)
		}
	case 7: // C_1E53: noise
		for bh := 0; bh < 0xff; bh++ {
			r := int(u_rand_a()&0x7f | 0x40)
			e.half(mul8(r, S)>>1|1, 1.3)
		}
	case 8: // C_1F22
		for bl := 0x80; bl > 0; bl-- {
			ax := mul8(S, bl)
			e.half(ax>>1+ax&1, 1.3)
		}
	case 9: // C_1DCD: a sweep up and down, cl sets the pitch; uneven halves
		sp := S>>1 + S&1
		end := uint16(0x1b * sp)
		dx := mul8(sp, cl)
		cx := uint16(sp)
		phase := func() {
			for ah := 0; ah < 0x30; ah++ {
				e.half(dx, 1)
				e.half(cx, 1)
			}
		}
		for {
			phase()
			dx -= uint16(sp)
			cx += uint16(sp)
			if cx == end {
				break
			}
		}
		for {
			phase()
			dx += uint16(sp)
			cx -= uint16(sp)
			if cx == 0 {
				break
			}
		}
	case 10: // C_1F73: cl random tones of 40 halves
		for {
			dl := int(u_rand_a()&0x3f) + 0x40
			dx := mul8(S, dl) << 1
			for c := 0; c < 0x28; c++ {
				e.half(dx, 1)
			}
			cl = (cl - 1) & 0xff
			if cl == 0 {
				break
			}
		}
	case 11: // C_1FA3: rising
		for cx := 0x40; cx < 0xc0; cx++ {
			for ah := 0; ah < 0x14; ah++ {
				e.half(uint16(cx*S), 1.15)
			}
		}
	case 12: // C_1FCC: falling
		for cx := 0xc0; cx > 0x40; cx-- {
			for ah := 0; ah < 0x14; ah++ {
				e.half(uint16(cx*S), 1.15)
			}
		}
	}
	return *e
}
