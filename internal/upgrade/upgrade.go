// Package upgrade carries the freeware "Ultima 4 VGA upgrade patch v1.3" (VGA 1.15; see
// files/Readme.txt, which allows free redistribution): the 256-colour tile set, font and
// palette, the VGA game screen frame and cutscene/rune pictures, and the corrected .TLK
// files. Its MIDPAK music drivers are not included.
//
// Formats: shapes.vga is 256 tiles of 16x16 and charset.vga 128 glyphs of 8x8, one byte a
// pixel; u4vga.pal is 256 6-bit VGA DAC triples; the pictures (*.old, rune_6..8.ega) are the
// game's RLE (byte 02h, count, value; else a literal) over 320x200 bytes, one a pixel. The
// .old names are those the patch's switch.bat swaps in.
package upgrade

import (
	"embed"
	"path"
	"strings"
)

//go:embed files
var files embed.FS

func file(name string) []byte {
	b, err := files.ReadFile("files/" + name)
	if err != nil {
		return nil
	}
	return b
}

var (
	Shapes  = file("shapes.vga")  // 256 tiles x 256 bytes
	Charset = file("charset.vga") // 128 glyphs x 64 bytes
	Palette = file("u4vga.pal")   // 768 bytes, 6-bit
)

// Picture returns the VGA version of a game picture (name as the game passes it, e.g.
// "rune_5.ega"), RLE-packed; ok false when the patch has none.
func Picture(name string) ([]byte, bool) {
	base := strings.ToLower(strings.TrimSuffix(name, path.Ext(name)))
	for _, n := range []string{base + ".old", base + ".ega"} {
		if b := file(n); b != nil {
			return b, true
		}
	}
	return nil, false
}

// TLK returns the patch's corrected conversation file for name (COVE, LCB, MINOC, SKARA,
// YEW .TLK); ok false for the others.
func TLK(name string) ([]byte, bool) {
	if !strings.EqualFold(path.Ext(name), ".tlk") {
		return nil, false
	}
	b := file(strings.ToUpper(strings.TrimSuffix(name, path.Ext(name))) + ".old")
	return b, b != nil
}
