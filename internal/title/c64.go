package title

// The introduction's pictures from Per Olofsson's C64 "Ultima IV Remastered" (Apache 2.0):
// from the first full-screen picture on (the seven story pictures, the moongate, the virtue
// cards and the abacus beads) in the C64's colours (Pepto's palette); the title and menu stay
// the DOS ones. The DOS drawing still runs underneath (same timing, same random numbers), and
// each copy from a story picture is mirrored onto a true-colour overlay (c64Scr) from the
// picture's C64 counterpart. Where the overlay is clear (the text area) the DOS text shows.

import (
	_ "embed"
	"image"

	"github.com/memmaker/u4-go/internal/ega"
	"github.com/memmaker/u4-go/internal/host"
)

var (
	//go:embed c64/cards1.koa
	cards1 []byte
	//go:embed c64/cards2.koa
	cards2 []byte
	//go:embed c64/145.koa
	koaTree []byte
	//go:embed c64/146.koa
	koaPortal []byte
	//go:embed c64/147.koa
	koaOutside []byte
	//go:embed c64/148.koa
	koaInside []byte
	//go:embed c64/149.koa
	koaWagon []byte
	//go:embed c64/14a.koa
	koaGypsy []byte
	//go:embed c64/14b.koa
	koaAbacus []byte
)

// pepto is the C64 palette.
var pepto = [16][3]byte{
	{0, 0, 0}, {255, 255, 255}, {104, 55, 43}, {112, 164, 178},
	{111, 61, 134}, {88, 141, 67}, {53, 40, 121}, {184, 199, 111},
	{111, 79, 37}, {67, 57, 0}, {154, 103, 89}, {68, 68, 68},
	{108, 108, 108}, {154, 210, 132}, {108, 94, 181}, {149, 149, 149},
}

func setC64(img *image.RGBA, x, y int, c byte) {
	if !image.Pt(x, y).In(img.Rect) {
		return
	}
	p := pepto[c&15]
	copy(img.Pix[img.PixOffset(x, y):], []byte{p[0], p[1], p[2], 255})
}

// A Koala picture: a 2-byte load address, the 8000-byte bitmap (40x25 cells of 8 bytes),
// 1000 bytes of screen RAM and 1000 of colour RAM (a cell's colours), the background colour.
const (
	koalaBitmap     = 2
	koalaScreenRAM  = koalaBitmap + 8000
	koalaColorRAM   = koalaScreenRAM + 1000
	koalaBackground = koalaColorRAM + 1000
)

// koala decodes the w x rows cells from cell column cx0 of a Koala picture (multicolour:
// each pixel two wide) into img at (dx, dy).
func koala(img *image.RGBA, k []byte, cx0, w, rows, dx, dy int) {
	bitmap, screenRAM, colorRAM := k[koalaBitmap:koalaScreenRAM], k[koalaScreenRAM:koalaColorRAM], k[koalaColorRAM:koalaBackground]
	background := k[koalaBackground]
	for cy := 0; cy < rows; cy++ {
		for cx := 0; cx < w; cx++ {
			cell := cy*40 + cx0 + cx
			cols := [4]byte{background, screenRAM[cell] >> 4, screenRAM[cell], colorRAM[cell]}
			for r := 0; r < 8; r++ {
				b := bitmap[cell*8+r]
				for p := 0; p < 8; p++ {
					setC64(img, dx+cx*8+p, dy+cy*8+r, cols[b>>(6-p/2*2)&3])
				}
			}
		}
	}
}

// koalaStoryPic decodes a story picture: 32204, black below to the text area (y 152), clear after.
func koalaStoryPic(k []byte) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, ega.W, ega.H))
	koala(img, k, 0, 40, 18, 0, 0)
	for y := 144; y < 152; y++ {
		for x := 0; x < ega.W; x++ {
			setC64(img, x, y, 0)
		}
	}
	return img
}

var (
	c64Scr = image.NewRGBA(image.Rect(0, 0, ega.W, ega.H)) // the overlay
	c64Of  = map[*byte]*image.RGBA{}                       // a picture buffer's C64 counterpart
)

func c64Reset() {
	clear(c64Scr.Pix)
	clear(c64Of)
	frame.Overlays = []host.Overlay{{Img: c64Scr}}
}

// c64Blit mirrors copyPic: a block of cols x rows of pic from (srcCol, srcY) to
// (dstCol, dstY).
func c64Blit(pic []byte, srcCol, srcY, cols, rows, dstCol, dstY int) {
	if len(pic) == 0 || c64Of[&pic[0]] == nil {
		return
	}
	src := c64Of[&pic[0]]
	for y := 0; y < rows; y++ {
		for x := 0; x < cols*cellSize; x++ {
			c64Copy(src, srcCol*cellSize+x, srcY+y, dstCol*cellSize+x, dstY+y)
		}
	}
}

// c64Inflate makes name's C64 counterpart for the picture buffer pic.
func c64Inflate(name string, pic []byte) {
	var img *image.RGBA
	switch name {
	case "tree.ega":
		img = koalaStoryPic(koaTree)
	case "portal.ega":
		img = koalaStoryPic(koaPortal)
	case "outside.ega":
		img = koalaStoryPic(koaOutside)
	case "inside.ega":
		img = koalaStoryPic(koaInside)
	case "wagon.ega":
		img = koalaStoryPic(koaWagon)
	case "gypsy.ega":
		img = koalaStoryPic(koaGypsy)
	case "abacus.ega":
		img = koalaStoryPic(koaAbacus)
	}
	c64Of[&pic[0]] = img
}

// c64Copy copies pixel (sx, sy) of src to (dx, dy) of the overlay, unless src is clear there.
func c64Copy(src *image.RGBA, sx, sy, dx, dy int) {
	if !image.Pt(sx, sy).In(src.Rect) || !image.Pt(dx, dy).In(c64Scr.Rect) {
		return
	}
	s := src.Pix[src.PixOffset(sx, sy):][:4]
	if s[3] != 0 {
		copy(c64Scr.Pix[c64Scr.PixOffset(dx, dy):], s)
	}
}

// c64Clear makes the overlay clear in a rectangle (the DOS text shows there).
func c64Clear(y0, y1, x, w int) {
	for y := y0; y < y1; y++ {
		clear(c64Scr.Pix[c64Scr.PixOffset(x, y):c64Scr.PixOffset(x+w, y)])
	}
}

// c64Gate draws the C64's moongate (a shimmering blue sprite, 24x27 at (78, 51)) over the
// picture: its rows y0..y1, on the even or odd pixels by phase.
func c64Gate(y0, y1, phase int) {
	pic := c64Of[&storyPic[0]]
	if pic == nil {
		return
	}
	for y := 51; y < 78; y++ {
		for x := 78; x < 102; x++ {
			c64Copy(pic, x, y, x, y)
			if y >= 51+y0 && y < 51+y1 && (x+phase)&1 == 0 {
				setC64(c64Scr, x, y, 6)
			}
		}
	}
}

// showCard puts virtue card v on the table, left (side 0) or right, where the C64 has it
// (columns 3 and 27, from row 2), the C64 table around it in the DOS card's place (column
// col, 96x124 from y 12).
func showCard(v, side, col int) {
	if pic := c64Of[&storyPic[0]]; pic != nil {
		for y := 12; y < 136; y++ {
			for x := col * 8; x < col*8+96; x++ {
				c64Copy(pic, x, y, x, y)
			}
		}
	}
	k := cards1
	if v >= 4 {
		k = cards2
	}
	koala(c64Scr, k, v%4*10, 10, 14, []int{24, 216}[side], 16)
}

// c64Bead puts a bead on the C64 abacus (the remaster's own bead is in Origin's code, not in
// its sources: this one is drawn): white chosen, black discarded, in virtue v's column
// between the rungs of line `line`.
func c64Bead(chosen bool, v, line int) {
	c := byte(0)
	if chosen {
		c = 1
	}
	x0, y0 := 128+8*v, 26+16*line
	for y := 0; y < 12; y++ {
		for x := 0; x < 8; x++ {
			if (y == 0 || y == 11) && (x < 2 || x > 5) {
				continue
			}
			setC64(c64Scr, x0+x, y0+y, c)
		}
	}
}
