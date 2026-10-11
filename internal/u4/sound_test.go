package u4

import (
	"encoding/binary"
	"os"
	"testing"
	"time"

	"github.com/memmaker/u4-go/internal/speaker"
)

// TestSoundEffects checks the beep's pitch and length and, with U4GO_WAV set to a directory,
// writes each effect there as a WAV to inspect.
func TestSoundEffects(t *testing.T) {
	b := effect(1, 0)
	if len(b) != 1 || int(b[0].Hz) != 120 || b[0].Dur.Round(time.Millisecond) != 67*time.Millisecond {
		t.Fatalf("beep %+v", b)
	}
	if e := effect(2, 0); e[0].Hz < 320 || e[0].Hz > 330 || e[0].Dur.Round(time.Millisecond) != 74*time.Millisecond {
		t.Fatalf("effect 2 %+v", e[0])
	}
	for n := 0; n < 13; n++ {
		e := effect(n, 0x20)
		var d time.Duration
		lo, hi := 1e9, 0.0
		for _, x := range e {
			d += x.Dur
			lo, hi = min(lo, x.Hz), max(hi, x.Hz)
		}
		t.Logf("effect %2d: %4d ms, %5.0f..%5.0f Hz", n, d.Milliseconds(), lo, hi)
		if dir := os.Getenv("U4GO_WAV"); dir != "" {
			pcm := speaker.Render(e)
			h := make([]byte, 44)
			copy(h, "RIFF")
			binary.LittleEndian.PutUint32(h[4:], uint32(36+2*len(pcm)))
			copy(h[8:], "WAVEfmt ")
			binary.LittleEndian.PutUint32(h[16:], 16)
			binary.LittleEndian.PutUint16(h[20:], 1)
			binary.LittleEndian.PutUint16(h[22:], 1)
			binary.LittleEndian.PutUint32(h[24:], 44100)
			binary.LittleEndian.PutUint32(h[28:], 88200)
			binary.LittleEndian.PutUint16(h[32:], 2)
			binary.LittleEndian.PutUint16(h[34:], 16)
			copy(h[36:], "data")
			binary.LittleEndian.PutUint32(h[40:], uint32(2*len(pcm)))
			for _, v := range pcm {
				h = binary.LittleEndian.AppendUint16(h, uint16(v))
			}
			os.WriteFile(dir+"/fx"+string(rune('a'+n))+".wav", h, 0o644)
		}
	}
}
