package u4

// Port of U4_SOUND.ASM: thirteen PC speaker effects, each a CPU-timed loop toggling port
// 61h bit 1 with delays scaled by speed_info. The port replays the toggles as timed square
// wave segments (internal/speaker) and blocks the game for the effect's length, as the
// original did.

import (
	"time"

	"github.com/memmaker/u4-go/internal/host"
	"github.com/memmaker/u4-go/internal/speaker"
)

// loopUs is the time of one "dec reg / jnz" delay iteration. LOW.ASM's calibration loop
// (speed_info*1000 iterations a 55 ms tick) costs about 8 such iterations, so one is
// 55 ms / (speed_info * 8000).
func loopUs() float64 { return 55000.0 / (float64(speed_info) * 8000) }

// spk collects an effect's toggles.
type spk struct {
	segs []speaker.Seg
	on   bool
	cur  float64
}

// out writes the current level (al) to the port.
func (s *spk) out(on bool) {
	if s.cur > 0 {
		s.segs = append(s.segs, speaker.Seg{On: s.on, Us: s.cur})
	}
	s.on, s.cur = on, 0
}

// wait spends n iterations (0 means 65536, as dec/jnz from 0 does) of a loop costing
// weight simple iterations each.
func (s *spk) wait(n uint16, weight float64) {
	c := float64(n)
	if n == 0 {
		c = 65536
	}
	s.cur += c * weight * loopUs()
}

func (s *spk) play() {
	s.out(false)
	d := speaker.Play(s.segs)
	waitFor(d)
}

// waitFor lets d pass on the host clock.
func waitFor(d time.Duration) {
	ticks := int(d * host.TPS / time.Second)
	for i := 0; i < ticks; i++ {
		host.Yield()
	}
}

func mul8(a, b int) uint16 { return uint16(byte(a)) * uint16(byte(b)) }

// sound plays effect n (cdecl: the effect, then the parameters some effects take).
func sound(n int, args ...int) { /*C_1D47*/
	cl, ch := 0, 0
	if len(args) > 0 {
		cl = args[0] & 0xff
	}
	if len(args) > 1 {
		ch = args[1] & 0xff
	}
	_ = ch
	if SoundFlag == 0 || uint(n)&0xff >= 0xd {
		return
	}
	S := int(speed_info)
	s := &spk{}
	al := false // port 61h bit 1, after "and al,not 3"
	toggle := func() { s.out(al); al = !al }
	switch n & 0xff {
	case 0: // C_1D69: a click (not scaled by speed)
		s.out(true)
		s.wait(0x32, 1)
	case 1: // sound_1: the beep (TITLE.EXE's too)
		sound_1(s, S)
	case 2: // C_1DA8
		dx := uint16(0xe0 * S)
		for i := 0; i < 0x30; i++ {
			toggle()
			s.wait(dx, 1)
		}
		s.out(false)
		sound_1(s, S)
	case 3: // C_1EB3: falling
		for bl := 5; bl < 0x100; bl++ {
			toggle()
			s.wait(mul8(S, bl)>>1, 1.3)
		}
	case 4: // C_1EFD
		bl := 0
		for {
			toggle()
			s.wait(mul8(S, bl)>>1, 1.3)
			bl = (bl - 1) & 0xff
			if bl < 0x80 {
				break
			}
		}
	case 5: // C_1F4A
		for bl := 0x80; bl < 0x100; bl++ {
			toggle()
			s.wait(mul8(S, bl)>>1, 1.6)
		}
	case 6: // C_1E7F: noise
		for bh := 0; bh < 0xff; bh++ {
			r := uint16(u_rand_a()&0x7f | 1)
			cx := uint16(int(r*3) * S)
			cx = cx>>2 | 1
			s.wait(cx, 1)
			toggle()
		}
	case 7: // C_1E53: noise
		for bh := 0; bh < 0xff; bh++ {
			r := int(u_rand_a()&0x7f | 0x40)
			dx := mul8(r, S)>>1 | 1
			s.wait(dx, 1.3)
			toggle()
		}
	case 8: // C_1F22
		for bl := 0x80; bl > 0; bl-- {
			toggle()
			ax := mul8(S, bl)
			ax = ax>>1 + ax&1
			s.wait(ax, 1.3)
		}
	case 9: // C_1DCD: a sweep up and down, cl sets the pitch
		sp := S>>1 + S&1
		end := uint16(0x1b * sp)
		dx := mul8(sp, cl)
		cx := uint16(sp)
		phase := func() {
			for ah := 0; ah < 0x30; ah++ {
				s.wait(dx, 1)
				toggle()
				s.wait(cx, 1)
				toggle()
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
	case 10: // C_1F73: cl random tones
		for {
			dl := int(u_rand_a()&0x3f) + 0x40
			dx := mul8(S, dl) << 1
			for c := 0; c < 0x28; c++ {
				s.wait(dx, 1)
				toggle()
			}
			cl = (cl - 1) & 0xff
			if cl == 0 {
				break
			}
		}
	case 11: // C_1FA3: rising
		for cx := 0x40; cx < 0xc0; cx++ {
			for ah := 0; ah < 0x14; ah++ {
				s.wait(uint16(cx*S), 1.15)
				toggle()
			}
		}
	case 12: // C_1FCC: falling
		for cx := 0xc0; cx > 0x40; cx-- {
			for ah := 0; ah < 0x14; ah++ {
				s.wait(uint16(cx*S), 1.15)
				toggle()
			}
		}
	}
	s.play()
}

func sound_1(s *spk, S int) {
	dx := uint16(0xca * S)
	al := false
	for i := 0; i < 0x10; i++ {
		s.out(al)
		al = !al
		s.wait(dx, 3)
	}
	s.out(false)
}

// waitUs lets us microseconds pass, carrying the fraction of a tick over.
func waitUs(us float64) {
	waitCarry += us
	tick := 1e6 / float64(host.TPS)
	for waitCarry >= tick {
		waitCarry -= tick
		host.Yield()
	}
}

var waitCarry float64
