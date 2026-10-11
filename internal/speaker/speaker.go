// Package speaker is the PC speaker as a plain synth: a sound is a list of notes (a square
// wave of a frequency for a duration, or silence), rendered into one Ebiten audio stream.
package speaker

import (
	"encoding/binary"
	"os"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

const rate = 44100

// Note is a square wave of Hz for Dur; Hz 0 is silence.
type Note struct {
	Hz  float64
	Dur time.Duration
}

// Half is one half period of a square wave lasting us microseconds: sweeps and noise are
// sequences of such halves.
func Half(us float64) Note {
	return Note{Hz: 1e6 / (2 * us), Dur: time.Duration(us * float64(time.Microsecond))}
}

// Beep is the error beep (sound_1 in both programs): 16 half periods of 0CAh*3 calibrated
// delay-loop passes (6.875 us each), a 120 Hz buzz of 67 ms.
var Beep = []Note{{Hz: 1e6 / (2 * 0xca * 3 * 6.875), Dur: 16 * time.Duration(0xca*3*6875) * time.Nanosecond}}

var (
	mu      sync.Mutex
	pcm     []int16 // queued samples
	started bool
	frac    float64 // sample time owed between notes
	phase   float64 // of the square wave, in periods
	lp      float64 // one-pole low-pass state (the speaker is no ideal square)
)

// Start opens the audio stream, once. Scripted test runs (U4GO_INPUT) stay silent.
func Start() {
	if started || os.Getenv("U4GO_INPUT") != "" {
		return
	}
	started = true
	p, err := audio.NewContext(rate).NewPlayer(stream{})
	if err != nil {
		return
	}
	p.SetBufferSize(50 * time.Millisecond)
	p.Play()
}

// Play queues notes and returns their total length.
func Play(notes []Note) time.Duration {
	mu.Lock()
	defer mu.Unlock()
	pcm = append(pcm, Render(notes)...)
	if len(pcm) > rate*4 { // never let the queue run away
		pcm = pcm[len(pcm)-rate*4:]
	}
	var total time.Duration
	for _, n := range notes {
		total += n.Dur
	}
	return total
}

// Render turns notes into mono 44.1 kHz samples, carrying the phase across calls.
func Render(notes []Note) []int16 {
	var out []int16
	for _, n := range notes {
		f := n.Dur.Seconds()*rate + frac
		k := int(f)
		frac = f - float64(k)
		for i := 0; i < k; i++ {
			target := 0.0
			if n.Hz > 0 {
				target = 6000
				if phase >= 0.5 {
					target = -6000
				}
				phase += n.Hz / rate
				phase -= float64(int(phase))
			}
			lp += (target - lp) * 0.5
			out = append(out, int16(lp))
		}
	}
	return out
}

type stream struct{}

func (stream) Read(b []byte) (int, error) {
	frames := len(b) / 4
	mu.Lock()
	defer mu.Unlock()
	for f := 0; f < frames; f++ {
		v := int16(0)
		if len(pcm) > 0 {
			v = pcm[0]
			pcm = pcm[1:]
		} else {
			lp *= 0.99 // relax to the middle when silent
			v = int16(lp)
		}
		binary.LittleEndian.PutUint16(b[f*4:], uint16(v))
		binary.LittleEndian.PutUint16(b[f*4+2:], uint16(v))
	}
	return frames * 4, nil
}
