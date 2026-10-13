// Package music plays the MIDI soundtrack of the "Ultima 4 VGA upgrade patch" v1.3 (Aridindae
// Dragon, freeware): its LARGE.XMI, nine General MIDI songs, synthesized with meltysynth from a
// SoundFont (GeneralUser GS, embedded in desktop builds when tools/fetch-soundfont.sh fetched
// it; on js/wasm read from the file $GOPLAY_SF2 names) into one Ebiten audio stream, mixed by
// Ebiten with the PC speaker effects.
package music

//go:generate sh ../../tools/fetch-soundfont.sh

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"math"
	"os"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/memmaker/u4-go/internal/meltysynth"
)

// The songs, numbered as the patch's driver asks MIDPAK for them (sequence n-1 of LARGE.XMI).
const (
	Stop        = 0
	Towns       = 1 // also the title
	Overworld   = 2
	Dungeon     = 3
	Castle      = 4
	LordBritish = 5
	Shopping    = 6
	Shrine      = 7
	Combat      = 8
	Fanfare     = 9
)

//go:embed large.xmi
var largeXMI []byte

const rate = 44100

// seq renders one song at a time through a synthesizer, looping it.
type seq struct {
	synth  *meltysynth.Synthesizer
	events []event
	pos    int // next event
	at     int // samples since the song started
}

// start switches to song n (Stop: silence), letting the old notes release.
func (s *seq) start(songs [][]event, n int) {
	s.synth.NoteOffAll(false)
	s.synth.ResetAllControllers()
	s.events, s.pos, s.at = nil, 0, 0
	if n > 0 && n <= len(songs) {
		s.events = songs[n-1]
	}
}

// render fills l and r, playing the events due on the way.
func (s *seq) render(l, r []float32) {
	for len(l) > 0 {
		n := len(l)
		if len(s.events) > 0 {
			for s.pos < len(s.events) && s.events[s.pos].tick*rate/tickHz <= s.at {
				e := s.events[s.pos]
				if e.status != 0 {
					s.synth.ProcessMidiMessage(int32(midiChannel(e.status)), int32(midiMessage(e.status)<<4), int32(e.d1), int32(e.d2))
				}
				s.pos++
			}
			if s.pos == len(s.events) { // loop
				s.pos, s.at = 0, 0
				continue
			}
			n = min(n, s.events[s.pos].tick*rate/tickHz-s.at)
		}
		s.synth.Render(l[:n], r[:n])
		s.at += n
		l, r = l[n:], r[n:]
	}
}

func newSeq(sf2 []byte) (*seq, error) {
	sf, err := meltysynth.NewSoundFont(bytes.NewReader(sf2))
	if err != nil {
		return nil, err
	}
	syn, err := meltysynth.NewSynthesizer(sf, meltysynth.NewSynthesizerSettings(rate))
	if err != nil {
		return nil, err
	}
	return &seq{synth: syn}, nil
}

var (
	mu      sync.Mutex
	songs   [][]event
	player  *seq // nil until the SoundFont is loaded
	song    int  // what the game asked for
	playing int  // what player plays
	muted   bool
	volume  = 8 // 0..10, 10 is the SoundFont's own level
	started bool
)

// Start loads the songs and the SoundFont in the background and opens the stream, once.
// Scripted test runs (U4GO_INPUT) stay silent; without a SoundFont there is no music.
func Start() {
	if started || os.Getenv("U4GO_INPUT") != "" {
		return
	}
	started = true
	sf2 := embeddedSF2()
	if p := os.Getenv("GOPLAY_SF2"); p != "" {
		if b, err := os.ReadFile(p); err == nil {
			sf2 = b
		}
	}
	if sf2 == nil {
		return
	}
	go func() {
		ss, err := parseXMI(largeXMI)
		if err != nil {
			return
		}
		s, err := newSeq(sf2)
		if err != nil {
			return
		}
		mu.Lock()
		songs, player, playing = ss, s, Stop
		mu.Unlock()
		ctx := audio.CurrentContext()
		if ctx == nil {
			ctx = audio.NewContext(rate)
		}
		p, err := ctx.NewPlayerF32(stream{})
		if err != nil {
			return
		}
		p.SetBufferSize(100 * time.Millisecond)
		p.Play()
	}()
}

// Play asks for song n (Stop for none); asking for the song already playing changes nothing.
func Play(n int) {
	mu.Lock()
	song = n
	mu.Unlock()
}

// Current is the song last asked for.
func Current() int {
	mu.Lock()
	defer mu.Unlock()
	return song
}

// Toggle mutes or unmutes the music (the patch's Alt-V).
func Toggle() { mu.Lock(); muted = !muted; mu.Unlock() }

// Louder and Softer step the volume (the patch's Alt-'+' and Alt-'-').
func Louder() { mu.Lock(); volume = min(volume+1, 10); mu.Unlock() }
func Softer() { mu.Lock(); volume = max(volume-1, 0); mu.Unlock() }

type stream struct{}

var buf [2][]float32

func (stream) Read(b []byte) (int, error) {
	frames := len(b) / 8
	mu.Lock()
	defer mu.Unlock()
	if cap(buf[0]) < frames {
		buf[0], buf[1] = make([]float32, frames), make([]float32, frames)
	}
	l, r := buf[0][:frames], buf[1][:frames]
	if song != playing {
		player.start(songs, song)
		playing = song
	}
	player.render(l, r)
	g := float32(volume) / 10 * float32(fader("music")) // and the goplay desk's
	if muted {
		g = 0
	}
	for i := range frames {
		putF32(b[i*8:], l[i]*g)
		putF32(b[i*8+4:], r[i]*g)
	}
	return frames * 8, nil
}

func putF32(b []byte, v float32) { binary.LittleEndian.PutUint32(b, math.Float32bits(v)) }
