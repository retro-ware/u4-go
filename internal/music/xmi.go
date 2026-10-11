package music

import (
	"encoding/binary"
	"errors"
	"sort"
)

// XMI (Miles' extended MIDI) is IFF: FORM XDIR { INFO } then CAT XMID { FORM XMID { TIMB? EVNT } ... },
// chunk lengths big-endian and padded to even. An EVNT chunk is a MIDI track with two twists:
// a delay is a run of bytes below 80h that add up (no variable-length quantity), and a note-on
// carries its duration as a variable-length quantity instead of a matching note-off. Time is in
// 120 Hz ticks; MIDIFORM already folded the tempo into the delays.

// tickHz is XMI's fixed tick rate.
const tickHz = 120

// event is a MIDI channel message due at tick; status 0 is a no-op.
type event struct {
	tick    int
	status  byte
	d1, d2  byte
	noteOff bool // sorts before note-ons of the same tick
}

// parseXMI returns the sequences of an XMI file, in order.
func parseXMI(b []byte) ([][]event, error) {
	var songs [][]event
	var walk func(b []byte) error
	walk = func(b []byte) error {
		for len(b) >= 8 {
			id, n := string(b[:4]), int(binary.BigEndian.Uint32(b[4:8]))
			if 8+n > len(b) {
				return errors.New("xmi: truncated chunk " + id)
			}
			body := b[8 : 8+n]
			switch id {
			case "FORM", "CAT ":
				if len(body) < 4 {
					return errors.New("xmi: short " + id)
				}
				if err := walk(body[4:]); err != nil {
					return err
				}
			case "EVNT":
				ev, err := parseEVNT(body)
				if err != nil {
					return err
				}
				songs = append(songs, ev)
			}
			b = b[8+n+n&1:]
		}
		return nil
	}
	if err := walk(b); err != nil {
		return nil, err
	}
	if len(songs) == 0 {
		return nil, errors.New("xmi: no sequences")
	}
	return songs, nil
}

func parseEVNT(d []byte) ([]event, error) {
	var out []event
	tick, p := 0, 0
	errShort := errors.New("xmi: truncated event")
	vlq := func() (int, bool) {
		v := 0
		for p < len(d) {
			c := d[p]
			p++
			v = v<<7 | int(c&0x7f)
			if c < 0x80 {
				return v, true
			}
		}
		return 0, false
	}
	for p < len(d) {
		s := d[p]
		if s < 0x80 { // delay
			tick += int(s)
			p++
			continue
		}
		p++
		switch s >> 4 {
		case 0x9:
			if p+2 > len(d) {
				return nil, errShort
			}
			key, vel := d[p], d[p+1]
			p += 2
			dur, ok := vlq()
			if !ok {
				return nil, errShort
			}
			out = append(out, event{tick: tick, status: s, d1: key, d2: vel},
				event{tick: tick + dur, status: 0x80 | s&0xf, d1: key, noteOff: true})
		case 0x8, 0xa, 0xb, 0xe:
			if p+2 > len(d) {
				return nil, errShort
			}
			// Controllers 110-120 are AIL's own (channel locks, loops, callbacks).
			if s>>4 != 0xb || d[p] < 110 || d[p] > 120 {
				out = append(out, event{tick: tick, status: s, d1: d[p], d2: d[p+1]})
			}
			p += 2
		case 0xc, 0xd:
			if p+1 > len(d) {
				return nil, errShort
			}
			out = append(out, event{tick: tick, status: s, d1: d[p]})
			p++
		default: // F0/F7 sysex, FF meta
			if s == 0xff {
				if p >= len(d) {
					return nil, errShort
				}
				typ := d[p]
				p++
				if typ == 0x2f { // end of track: a no-op event marks where the loop restarts
					out = append(out, event{tick: tick})
					p = len(d)
					continue
				}
			}
			n, ok := vlq()
			if !ok || p+n > len(d) {
				return nil, errShort
			}
			p += n
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].tick != out[j].tick {
			return out[i].tick < out[j].tick
		}
		return out[i].noteOff && !out[j].noteOff
	})
	return out, nil
}
