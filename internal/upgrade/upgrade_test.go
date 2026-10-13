package upgrade

import "testing"

// talkFileSize is a .TLK file's size: 16 people of 288 bytes.
const talkFileSize = 16 * 288

func TestFiles(t *testing.T) {
	if len(Shapes) != 256*256 || len(Charset) != 128*64 || len(Palette) != 768 {
		t.Fatalf("sizes: shapes %d charset %d palette %d", len(Shapes), len(Charset), len(Palette))
	}
	for _, n := range []string{"start", "rune_0", "rune_5", "rune_8", "stoncrcl", "key7", "honesty",
		"compassn", "valor", "justice", "sacrific", "honor", "spirit", "humility", "truth", "love", "courage"} {
		b, ok := Picture(n + ".ega")
		if !ok {
			t.Fatalf("no %s", n)
		}
		px := 0
		for i := 0; i < len(b); i++ {
			if b[i] == 2 {
				px += int(b[i+1])
				i += 2
			} else {
				px++
			}
		}
		if px != 320*200 {
			t.Errorf("%s: %d pixels", n, px)
		}
	}
	for _, n := range []string{"COVE.TLK", "lcb.tlk", "MINOC.TLK", "SKARA.TLK", "YEW.TLK"} {
		if b, ok := TLK(n); !ok || len(b) != talkFileSize {
			t.Errorf("%s: %v %d", n, ok, len(b))
		}
	}
	if _, ok := TLK("BRITAIN.TLK"); ok {
		t.Error("BRITAIN.TLK is not the patch's")
	}
}
