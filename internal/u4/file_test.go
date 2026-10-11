package u4

import (
	"bytes"
	"math/rand"
	"testing"

	"github.com/memmaker/u4-go/internal/gamefs"
)

// Records decode and re-encode any file image byte for byte.
func TestRecordRoundTrip(t *testing.T) {
	for _, r := range []gamefs.Record{&tParty{}, &tNPC{}, &t_500{}, &tCombat{}} {
		img := make([]byte, len(gamefs.Encode(r)))
		rand.Read(img)
		if err := gamefs.Decode(r, img); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(gamefs.Encode(r), img) {
			t.Errorf("%T does not round-trip", r)
		}
	}
}
