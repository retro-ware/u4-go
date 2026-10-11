// Package speaker plays the PC speaker: the game toggles port 61h bit 1 in CPU-timed loops,
// which the port turns into a list of (level, duration) segments rendered as a square wave
// into one Ebiten audio stream.
package speaker

import (
	"encoding/binary"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

const rate = 44100

// Seg is one stretch of the speaker cone held in or out.
type Seg struct {
	On bool
	Us float64 // microseconds
}

var (
	mu      sync.Mutex
	pcm     []int16 // queued samples
	started bool
	frac    float64
	lp      float64 // one-pole low-pass state (the speaker is no ideal square)
)

// Start opens the audio stream, once.
func Start() {
	if started {
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

// Play queues segments and returns their total length.
func Play(segs []Seg) time.Duration {
	mu.Lock()
	defer mu.Unlock()
	total := 0.0
	for _, s := range segs {
		total += s.Us
		n := s.Us*rate/1e6 + frac
		k := int(n)
		frac = n - float64(k)
		target := -6000.0
		if s.On {
			target = 6000
		}
		for i := 0; i < k; i++ {
			lp += (target - lp) * 0.5
			pcm = append(pcm, int16(lp))
		}
	}
	if len(pcm) > rate*4 { // never let the queue run away
		pcm = pcm[len(pcm)-rate*4:]
	}
	return time.Duration(total * float64(time.Microsecond))
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
