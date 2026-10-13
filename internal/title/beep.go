package title

// Port of BEEP.ASM: the error beep (the game's sound(1)), and the clicks of the "Ultima IV"
// dissolve.

import (
	"time"

	"github.com/memmaker/u4-go/internal/speaker"
)

func beep() { /*C_21BF*/
	pace(speaker.Play(speaker.Beep))
}

// dissolveClicks counts the speaker clicks dissolvePic made since the last clicks call.
var dissolveClicks int

// clicks turns the counted clicks into a buzz spread over d: n toggles over d sound like a
// square wave of n/2 periods.
func clicks(d time.Duration) []speaker.Note {
	n := dissolveClicks
	dissolveClicks = 0
	if n == 0 {
		return []speaker.Note{{Dur: d}}
	}
	return []speaker.Note{{Hz: float64(n) / 2 / d.Seconds(), Dur: d}}
}
