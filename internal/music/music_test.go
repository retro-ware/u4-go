package music

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"
)

func TestParseXMI(t *testing.T) {
	ss, err := parseXMI(largeXMI)
	if err != nil {
		t.Fatal(err)
	}
	if len(ss) != 9 {
		t.Fatalf("%d songs, want 9", len(ss))
	}
	for i, s := range ss {
		ons := 0
		for j, e := range s {
			if j > 0 && e.tick < s[j-1].tick {
				t.Fatalf("song %d: unsorted at %d", i+1, j)
			}
			if e.status>>4 == 9 {
				ons++
			}
		}
		if ons == 0 {
			t.Errorf("song %d has no notes", i+1)
		}
	}
}

// TestRender plays each song once through the SoundFont and checks it is audible and does not
// clip. $MUSIC_WAV_DIR gets a WAV of each.
func TestRender(t *testing.T) {
	sf2 := embeddedSF2()
	if sf2 == nil {
		t.Skip("no SoundFont: run tools/fetch-soundfont.sh")
	}
	ss, _ := parseXMI(largeXMI)
	s, err := newSeq(sf2)
	if err != nil {
		t.Fatal(err)
	}
	for n := 1; n <= len(ss); n++ {
		s.start(ss, n)
		frames := ss[n-1][len(ss[n-1])-1].tick * rate / tickHz
		l, r := make([]float32, frames), make([]float32, frames)
		s.render(l, r)
		var peak, sum float64
		for i := range l {
			peak = max(peak, math.Abs(float64(l[i])), math.Abs(float64(r[i])))
			sum += float64(l[i]*l[i] + r[i]*r[i])
		}
		rms := math.Sqrt(sum / float64(2*frames))
		t.Logf("song %d: %.1f s, peak %.3f, rms %.4f", n, float64(frames)/rate, peak, rms)
		if frames < rate*5 || rms < 0.005 || peak >= 1 {
			t.Errorf("song %d: %d frames, peak %.3f, rms %.4f", n, frames, peak, rms)
		}
		if dir := os.Getenv("MUSIC_WAV_DIR"); dir != "" {
			writeWAV(t, filepath.Join(dir, fmt.Sprintf("song%d.wav", n)), l, r)
		}
	}
}

func writeWAV(t *testing.T, path string, l, r []float32) {
	b := make([]byte, 44+len(l)*4)
	copy(b, "RIFF")
	binary.LittleEndian.PutUint32(b[4:], uint32(len(b)-8))
	copy(b[8:], "WAVEfmt ")
	binary.LittleEndian.PutUint32(b[16:], 16)
	binary.LittleEndian.PutUint16(b[20:], 1)
	binary.LittleEndian.PutUint16(b[22:], 2)
	binary.LittleEndian.PutUint32(b[24:], rate)
	binary.LittleEndian.PutUint32(b[28:], rate*4)
	binary.LittleEndian.PutUint16(b[32:], 4)
	binary.LittleEndian.PutUint16(b[34:], 16)
	copy(b[36:], "data")
	binary.LittleEndian.PutUint32(b[40:], uint32(len(l)*4))
	for i := range l {
		binary.LittleEndian.PutUint16(b[44+i*4:], uint16(int16(max(-1, min(1, l[i]))*32767)))
		binary.LittleEndian.PutUint16(b[46+i*4:], uint16(int16(max(-1, min(1, r[i]))*32767)))
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatal(err)
	}
}
