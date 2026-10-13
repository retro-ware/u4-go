package u4

// Port of U4_SOUND.ASM: thirteen PC speaker effects. Each is described as square-wave half
// periods (internal/speaker); their lengths come from the original's delay-loop counts: a
// count of n passes weighing w simple dec/jnz passes each lasts n*w*passUs/SpeedInfo
// microseconds (the calibration made speedInfo*8000 passes last one 55 ms timer tick). The
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

// The thirteen speaker effects, named for where the game plays them.
const (
	soundStep        = iota // a click: a step, a member moving in combat
	soundBeep               // the beep: a bump, a refusal, a shop's ring
	soundBadCommand         // a buzz, then the beep
	soundMissile            // a falling whistle: something shot or thrown
	soundMiss               // a short fall: an attack that misses
	soundMonsterBlow        // a rising whistle: a monster's blow lands
	soundMonsterHit         // noise: a monster or thing is hit
	soundMemberHit          // higher noise: a party member is hurt
	soundFailure            // a falling whine: no party, food stolen, a spell fails
	soundSweep              // a warble down and up; its parameter sets the pitch
	soundMagic              // random tones; its parameter sets how many
	soundRising             // a slow rise: the whirlpool
	soundFalling            // a slow fall: the twister
	soundCount
)

// sound plays an effect and waits for it to finish; soundSweep and soundMagic take a
// parameter, used as a byte.
func sound(effectID int, param ...int) { /*C_1D47*/
	p := 0
	if len(param) > 0 {
		p = int(byte(param[0]))
	}
	if soundOn == 0 || effectID < 0 || effectID >= soundCount {
		return
	}
	waitFor(speaker.Play(effect(effectID, p)))
}

// effect describes an effect with its parameter p.
func effect(effectID, p int) fx {
	S := host.SpeedInfo
	e := &fx{}
	switch effectID {
	case soundStep: // C_1D69: uncalibrated, 50 passes
		e.half(50, 1)
	case soundBeep: // sound_1: the beep (TITLE.EXE's too)
		*e = append(*e, speaker.Beep...)
	case soundBadCommand: // C_1DA8: 48 halves of 224 passes (325 Hz, 74 ms), then the beep
		for range 48 {
			e.half(uint16(224*S), 1)
		}
		*e = append(*e, speaker.Beep...)
	case soundMissile: // C_1EB3
		for step := 5; step < 256; step++ {
			e.half(mul8(S, step)>>1, 1.3)
		}
	case soundMiss: // C_1EFD: starts at step 0, which is the longest half (0 means 65536)
		e.half(0, 1.3)
		for step := 255; step >= 128; step-- {
			e.half(mul8(S, step)>>1, 1.3)
		}
	case soundMonsterBlow: // C_1F4A
		for step := 128; step < 256; step++ {
			e.half(mul8(S, step)>>1, 1.6)
		}
	case soundMonsterHit: // C_1E7F
		for range 255 {
			r := uint16(randomByte()%128 | 1)
			passes := uint16(int(r*3) * S)
			e.half(passes>>2|1, 1)
		}
	case soundMemberHit: // C_1E53
		for range 255 {
			r := int(randomByte()%128 | 64)
			e.half(mul8(r, S)>>1|1, 1.3)
		}
	case soundFailure: // C_1F22: halves rounded up
		for step := 128; step > 0; step-- {
			passes := mul8(S, step)
			e.half(passes>>1+passes&1, 1.3)
		}
	case soundSweep: // C_1DCD: pairs of uneven halves, the long one shortening while the
		// short one grows, then back
		unit := uint16(S>>1 + S&1)
		turn := 27 * unit
		long := mul8(int(unit), p)
		short := unit
		phase := func() {
			for range 48 {
				e.half(long, 1)
				e.half(short, 1)
			}
		}
		for {
			phase()
			long -= unit
			short += unit
			if short == turn {
				break
			}
		}
		for {
			phase()
			long += unit
			short -= unit
			if short == 0 {
				break
			}
		}
	case soundMagic: // C_1F73: p random tones of 40 halves (0 counts as 256)
		for {
			step := int(randomByte()%64) + 64
			passes := mul8(S, step) << 1
			for range 40 {
				e.half(passes, 1)
			}
			p = int(byte(p - 1))
			if p == 0 {
				break
			}
		}
	case soundRising: // C_1FA3
		for step := 64; step < 192; step++ {
			for range 20 {
				e.half(uint16(step*S), 1.15)
			}
		}
	case soundFalling: // C_1FCC
		for step := 192; step > 64; step-- {
			for range 20 {
				e.half(uint16(step*S), 1.15)
			}
		}
	}
	return *e
}
