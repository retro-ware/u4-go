package title

// Port of UTIL2.ASM: the delay with the spinning cursor, the random generators and the
// keyboard. Not ported: the disk and drive calls and the critical (disk) error handler, the
// port has no disks; a floppy track 9 piracy check is never called.

import (
	"time"

	"github.com/memmaker/u4-go/internal/host"
)

// BIOS key words: the scan code in the high byte, the character in the low.
const (
	keyBackspace     = 14<<8 | '\b'
	keyCtrlBackspace = 14<<8 | 127
	keyCtrlC         = 46<<8 | 3
	keyEscape        = 1<<8 | 27
	keySpace         = 57<<8 | ' '
	keyEnter         = 28<<8 | '\r'
	keyLeft          = 75 << 8
	keyRight         = 77 << 8
	keyUp            = 72 << 8
	keyDown          = 80 << 8
	keyA             = 30<<8 | 'a'
	keyB             = 48<<8 | 'b'
	keyI             = 23<<8 | 'i'
	keyJ             = 36<<8 | 'j'
	keyR             = 19<<8 | 'r'
)

// The text cursor spins through four glyphs.
const (
	cursorGlyph  = 28 // the first of the four
	cursorPhases = 4
)

var (
	cursorRate  uint16        // delay passes between turns of the cursor (set by animateMonsters)
	cursorCount uint16        // cursorRate when the count started
	cursorShown bool          // the delay shows the spinning cursor
	cursorPhase uint8         // which of the cursor glyphs shows
	randomSeed  [16]byte      // the 128-bit random state
	cursorDue   time.Duration // when the cursor turns next
)

// forever is a delay that never ends by itself: its end second is never reached.
const forever = 32897

func dosSeconds() uint16 { return uint16(host.Clock/256) % 60 }

// now is the host clock as a duration.
func now() time.Duration { return time.Duration(host.Clock) * time.Second / 256 }

// delay waits until the clock's second has advanced by seconds, animating the monsters. With
// byKey a waiting key ends it early and the cursor (when shown) spins at the text position.
func delay(seconds uint16, byKey bool) { /*C_3149*/
	cursorDue = 0
	cursorCount = cursorRate
	end := dosSeconds() + seconds
	if end >= 60 {
		end -= 60
		if byKey && keyWaiting() {
			return // no cursor drawn yet to delete
		}
	}
	for {
		if byKey && keyWaiting() {
			break
		}
		animateMonsters()
		if byKey && cursorShown && (cursorCount != cursorRate || now() >= cursorDue) {
			if cursorCount == cursorRate {
				cursorPhase = (cursorPhase + cursorPhases - 1) % cursorPhases
				putGlyph(cursorGlyph + cursorPhase)
			}
			cursorDue = now() + time.Duration(1+cursorRate)*cursorPass
			cursorCount = cursorRate
		}
		host.Yield()
		if dosSeconds() == end {
			break
		}
	}
	if byKey && cursorShown {
		putGlyph(' ') // delete the cursor
	}
}

// randomByte (C_31ED) adds the 128-bit seed's bytes into each other from the end with
// carries, then counts the seed up by one.
func randomByte() uint8 {
	s := &randomSeed
	sum := s[15]
	carry := 0
	for i := 14; i >= 1; i-- {
		v := int(sum) + int(s[i]) + carry
		sum, carry = byte(v), v>>8
		s[i] = sum
	}
	s[0] = byte(int(s[0]) + int(sum) + carry)
	for i := 15; i >= 0; i-- {
		s[i]++
		if s[i] != 0 {
			break
		}
	}
	return s[0]
}

// quickRandom (C_321E) is a cheaper generator on the seed's first four bytes. (Its
// add-with-carry instructions take the caller's carry; the compiled C always calls it with
// the carry clear.)
func quickRandom() uint8 {
	s := &randomSeed
	a := s[3] + s[2]
	s[2] = a
	s[1] ^= a
	a = s[1] + s[0]
	a = a>>1 | a<<7 // rotate right
	s[0] = a
	s[3] = a
	return a
}

// keyWaiting (C_3241) counts the poll in the party record, stirs the random generator and
// reports a waiting key; Ctrl-C is dropped. Without a key the monsters get their turn.
func keyWaiting() bool {
	host.KeyPolled = true
	randomByte()
	party.counter++
	if len(host.KeyQueue) > 0 {
		k := host.KeyQueue[0]
		if k != keyCtrlC && k != 0 {
			return true
		}
		host.KeyQueue = host.KeyQueue[1:]
	}
	animateMonsters()
	return false
}

// readKey (C_3270) waits for a key (spinning the cursor) and returns its BIOS key word.
func readKey() int {
	delay(forever, true)
	for len(host.KeyQueue) == 0 {
		host.KeyPolled = true
		host.Yield()
	}
	k := host.KeyQueue[0]
	host.KeyQueue = host.KeyQueue[1:]
	return int(k)
}

func flushKeys() { /*C_3280*/
	for keyWaiting() {
		readKey()
	}
}
