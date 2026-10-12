package title

// Port of UTIL2.ASM: the delay with the spinning cursor, the random generators, the keyboard,
// Not ported: the disk and drive calls and the critical (disk) error handler, the port has no
// disks; C_32CB, a floppy track 9 piracy check, is never called.

import (
	"time"

	"github.com/memmaker/u4-go/internal/host"
)

var (
	cursor_rate uint16            /*D_31BC*/
	D_31BE      uint16            /*D_31BE*/
	D_31C0      uint16   = 0xffff /*display cursor flag*/
	seed_128    [16]byte          /*D_6966 random seeds*/
	D_6E76      uint8             /*cursor state {0, 1, 2, 3}*/
)

func dosSeconds() uint16 { return uint16(host.Clock/256) % 60 }

// now is the host clock as a duration.
func now() time.Duration { return time.Duration(host.Clock) * time.Second / 256 }

var cursorDue time.Duration // when the cursor turns next (D_6E77 counted u_delay passes)

// u_delay waits bp04 seconds; with bp06 set a key ends the wait, and with bp06 & D_31C0 the
// cursor spins at txt_X/txt_Y.
func u_delay(bp04, bp06 uint16) { /*C_3149*/
	cursorDue = 0
	D_31BE = cursor_rate
	si := dosSeconds() + bp04
	if si >= 60 {
		si -= 60
		if bp06 == 0 {
			goto C_318C
		}
		if u_kbhit() != 0 {
			return // C_31E8: no cursor to delete
		}
	}
C_317D:
	if bp06 != 0 && u_kbhit() != 0 {
		goto C_31D9
	}
C_318C:
	t_callback()
	if bp06&D_31C0 != 0 {
		if D_31BE == cursor_rate {
			if now() < cursorDue {
				goto C_31CD
			}
			D_6E76 = (D_6E76 - 1) & 3
			Gra_putchar(D_6E76 + 0x1c)
		}
		// C_31B9: reload the cursor counter
		cursorDue = now() + time.Duration(1+cursor_rate)*cursorPass
		D_31BE = cursor_rate
	}
C_31CD:
	host.Yield()
	if si != dosSeconds() {
		goto C_317D
	}
C_31D9:
	if bp06&D_31C0 != 0 {
		Gra_putchar(' ') // delete cursor
	}
}

// u_rand_a (C_31ED): an add-with-carry chain over the 128-bit seed, then a 128-bit increment.
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

// u_rand_b (C_321E). Its adcs take the caller's carry; the compiled C calls it with the carry
// clear (after a compare that fails or an and), so they are adds.
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

// u_kbhit (C_3241) counts Party.f_000 and reports a waiting key (dropping ^C), running
// t_callback when there is none.
func u_kbhit() int {
	host.KeyPolled = true
	u_rand_a()
	Party.f_000++
	if len(host.KeyQueue) > 0 {
		k := host.KeyQueue[0]
		if k != 0x2E03 && k != 0 {
			return 1
		}
		host.KeyQueue = host.KeyQueue[1:] // remove <CTRL C>
	}
	t_callback()
	return 0
}

// u_kbread (C_3270) waits for a key (a very long u_delay a key interrupts) and returns its
// BIOS word.
func u_kbread() int {
	u_delay(0x8081, 0x8081)
	for len(host.KeyQueue) == 0 {
		host.KeyPolled = true
		host.Yield()
	}
	k := host.KeyQueue[0]
	host.KeyQueue = host.KeyQueue[1:]
	return int(k)
}

func u_kbflush() { /*C_3280*/
	for u_kbhit() != 0 {
		u_kbread()
	}
}
