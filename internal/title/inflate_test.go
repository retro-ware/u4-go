package title

import (
	"bytes"
	"os"
	"testing"
)

// SHAPES.EGZ must inflate to SHAPES.EGA byte for byte (needs the game files).
func TestInflate(t *testing.T) {
	GameDir = os.Getenv("U4_GAMEDIR")
	if GameDir == "" {
		GameDir = os.ExpandEnv("$HOME/Projects/u4-decompiled/gog/inst")
	}
	SaveDir = t.TempDir()
	want, err := os.ReadFile(GameDir + "/SHAPES.EGA")
	if err != nil {
		t.Skip("no game files")
	}
	speed_info = 0xffff // no pacing (cpu would yield to a host that isn't running)
	got := make([]byte, 0x8000)
	C_217E("SHAPES.EGZ", got)
	if !bytes.Equal(got, want) {
		t.Fatal("SHAPES.EGZ does not inflate to SHAPES.EGA")
	}
	pic := make([]byte, 0x8000)
	Gra_inflate("TITLE.EGA", pic)
}
