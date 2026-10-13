package u4

// The keyboard, waits with the blinking cursor, and the random numbers.

import (
	"github.com/memmaker/u4-go/internal/host"
)

// CallbackTicks is the host ticks one redraw (one redraw of the view, the unit the
// game's animations count in) takes: 35 redraws a second, what a 386 (host.SpeedInfo 12)
// managed; the moon cycle shows the rate grows as speedInfo/2+1 from 5.3 on a 4.77 MHz PC.
const CallbackTicks = 2

// untilKey is a wait in seconds that no clock reaches, so only a key ends it.
const untilKey = 32897

// The cursor is four characters of the font shown in turn, from cursorGlyph on.
const (
	cursorGlyph  = 28
	cursorPhases = 4
)

var (
	cursorRate      uint16   // redraws between cursor steps, less one (set by the view)
	cursorRateSeen  uint16   // the rate the countdown was started with
	cursorPhase     uint8    // which of the cursor's characters shows
	cursorCountdown uint16   // passes until the cursor's next step
	randomSeed      [16]byte // the random generator's state
)

// dosSeconds is DOS's clock (int 21h/2Ch) seconds, from the host clock.
func dosSeconds() int { return int(host.Clock/256) % 60 }

// waitSeconds lets seconds pass (the clock's seconds changing that often), running the
// view. With keyEnds set a key ends the wait early, and the cursor blinks meanwhile.
func waitSeconds(seconds int, keyEnds bool) {
	cursorCountdown = 1
	cursorRateSeen = cursorRate
	end := dosSeconds() + seconds
	if end >= 60 {
		end -= 60
		if keyEnds && keyWaiting() {
			return // before the cursor was drawn: nothing to erase
		}
	}
	for !(keyEnds && keyWaiting()) {
		redraw()
		if keyEnds {
			stepCursor()
		}
		if dosSeconds() == end {
			break
		}
	}
	if keyEnds {
		putGlyph(' ') // erase the cursor
	}
}

// stepCursor counts down to the cursor's next step and draws it; a new rate restarts the
// count.
func stepCursor() {
	if cursorRateSeen == cursorRate {
		cursorCountdown--
		if cursorCountdown != 0 {
			return
		}
		cursorPhase = (cursorPhase - 1) % cursorPhases
		putGlyph(cursorPhase + cursorGlyph)
	}
	cursorCountdown = (1 + cursorRate) * speedInfo
	cursorRateSeen = cursorRate
}

// randomByte is the game's random generator: a 128-bit additive generator, an add-with-carry
// chain over the seed, then a 128-bit increment.
func randomByte() uint8 {
	s := &randomSeed
	sum := s[15]
	carry := uint16(0)
	for i := 14; i >= 1; i-- {
		v := uint16(sum) + uint16(s[i]) + carry
		sum = byte(v)
		carry = v >> 8
		s[i] = sum
	}
	s[0] = byte(uint16(s[0]) + uint16(sum) + carry)
	for i := 15; i >= 0; i-- {
		s[i]++
		if s[i] != 0 {
			break
		}
	}
	return s[0]
}

// animRandomByte is the animations' cheaper generator, over the same seed.
func animRandomByte() uint8 {
	s := &randomSeed
	a := s[3] + s[2]
	s[2] = a
	s[1] ^= a
	a = s[1] + s[0]
	a = a>>1 | a<<7
	s[0] = a
	s[3] = a
	return a
}

// randomBits is a random byte masked with mask (mask 7: 0..7).
func randomBits(mask int) int { return int(randomByte()) & mask }

// randomBelow is a random number from 0 to n-1.
func randomBelow(n int) int { return int(randomByte()) % n }

// randomBelowSafe is randomBelow that gives 0 for n == 0.
func randomBelowSafe(n int) int { return safeMod(int(randomByte()), n) }

// animRandomBits is an animation's random byte masked with mask.
func animRandomBits(mask int) int { return int(animRandomByte()) & mask }

// ctrlC is the BIOS key word of Ctrl-C, which the keyboard routines drop.
const ctrlC = 11779

// keyWaiting says whether a key is waiting. Every poll draws a random number and counts on
// the party's counter; when no key is waiting the view runs.
func keyWaiting() bool {
	host.KeyPolled = true
	randomByte()
	party.counter++
	if len(host.KeyQueue) > 0 {
		k := host.KeyQueue[0]
		if k != ctrlC && k != 0 {
			return true
		}
		host.KeyQueue = host.KeyQueue[1:]
	}
	redraw()
	return false
}

// readKey waits for a key (blinking the cursor) and returns its BIOS word.
func readKey() int {
	waitSeconds(untilKey, true)
	waiting = true
	for len(host.KeyQueue) == 0 {
		host.KeyPolled = true
		host.Yield()
	}
	waiting = false
	k := host.KeyQueue[0]
	host.KeyQueue = host.KeyQueue[1:]
	if prompt != pNone {
		ask(pNone)
	}
	return int(k)
}

// flushKeys drops the waiting keys.
func flushKeys() {
	for keyWaiting() {
		readKey()
	}
}

// seedRandom seeds the random generator from the time of day, and sets the speed.
func seedRandom() {
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
