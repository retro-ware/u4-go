package title

// The host's menus and clicks for the title: what each of its key reads takes, so the
// gamepad and the mouse can answer them. They only type the original's key words.

import (
	"image"
	"image/color"
	"strings"

	"github.com/memmaker/u4-go/internal/ega"
	"github.com/memmaker/u4-go/internal/host"
)

const (
	askNone  = iota
	askMain  // the main menu (its cursor is the port's)
	askName  // the name
	askSex   // Male or Female
	askGypsy // the gypsy's A or B
)

var asking int

// askThen sets what the next key reads are for.
func askThen(k int) {
	asking = k
	host.NewPrompt()
}

func item(label string, c byte) host.MenuItem {
	return host.MenuItem{Label: label, Keys: []uint16{host.KeyWord(c)}}
}

func setHooks() {
	host.Glyphs = func() [][64]byte { return screen.Glyphs[:] }
	host.MapClick = nil
	host.PromptMenu = func(bool) *host.Menu {
		switch asking {
		case askName:
			return &host.Menu{Items: []host.MenuItem{{Label: "Type...", Keyboard: true}}}
		case askSex:
			return &host.Menu{Items: []host.MenuItem{item("Male", 'm'), item("Female", 'f')}}
		}
		return nil
	}
	host.BackKey = func() uint16 {
		if asking == askName {
			return keyEnter // an empty name goes back to the menu
		}
		return keyEscape
	}
	host.Hover = func(x, y int) {
		if k := answerAt(x, y); asking == askGypsy && k >= 0 && k != answerSel {
			answerSel = k
			drawAnswers()
		}
	}
	host.ScreenClick = func(x, y int) ([]uint16, bool) {
		if k := answerAt(x, y); asking == askGypsy && k >= 0 {
			return []uint16{[]uint16{keyA, keyB}[k]}, true
		}
		// the main menu's lines
		if row, col := y/cellSize, x/cellSize; asking == askMain && row >= menuFirstRow && row < menuFirstRow+menuItems && col >= menuCol && col < menuEndCol {
			return []uint16{uint16(menuKeys[row-menuFirstRow])}, true
		}
		return nil, false
	}
}

// The gypsy's answers are picked in her text: "A) ..." and "B) ..." are drawn in gray, the
// selected one bright; the arrows switch, Enter takes it, the mouse points and clicks.

var (
	textCells   [textRows][textCols]byte // what the title printed where
	answers     [2][2]int                // each answer's first and last cell, counted from textTop
	answerSel   int                      // the selected answer
	answerFound bool                     // the question shows both answers
	// answerImg recolours the answers over the text area (rows textTop..24)
	answerImg = image.NewRGBA(image.Rect(0, 0, ega.W, ega.H-textTop*cellSize))
)

// findAnswers locates "A) ..." and "B) ..." in the text area.
func findAnswers() {
	var b strings.Builder
	for _, row := range textCells[textTop:] {
		for _, c := range row {
			b.WriteByte(max(c, ' '))
		}
	}
	s := b.String()
	ia, ib := strings.Index(s, "A)"), strings.Index(s, "B)")
	answerFound, answerSel = ia >= 0 && ib > ia, 0
	if !answerFound {
		return
	}
	endA := strings.TrimSuffix(strings.TrimRight(s[:ib], " "), "or")
	endA = strings.TrimRight(endA, " ,")
	answers = [2][2]int{{ia, len(endA) - 1}, {ib, len(strings.TrimRight(s, " ")) - 1}}
	for _, o := range frame.Overlays {
		if o.Img == answerImg {
			return
		}
	}
	frame.Overlays = append(frame.Overlays, host.Overlay{Y: textTop * cellSize, Img: answerImg})
}

// forCell calls f with each pixel (x, y) of text cell i, y counted from textTop.
func forCell(i int, f func(x, y int)) {
	x0, y0 := i%textCols*cellSize, i/textCols*cellSize
	for r := 0; r < cellSize; r++ {
		for k := 0; k < cellSize; k++ {
			f(x0+k, y0+r)
		}
	}
}

// clearAnswers takes the recolouring away (clearTextArea).
func clearAnswers() {
	clear(answerImg.Pix)
	answerFound = false
}

// drawAnswers recolours the answers' letters: a slightly dimmed white, the selected one
// a soft periwinkle (tints multiplying the glyphs' colours).
func drawAnswers() {
	if !answerFound {
		return
	}
	for n, a := range answers {
		c := color.RGBA{200, 200, 200, 255}
		if n == answerSel {
			c = color.RGBA{150, 140, 210, 255}
		}
		for i := a[0]; i <= a[1]; i++ {
			forCell(i, func(x, y int) {
				if v := frame.Pix[(textTop*cellSize+y)*host.ScreenW+x]; v != 0 {
					// tint the glyph's own colour, keeping the font's shading
					p := frame.Palette[int(v)*3:]
					// the palette is 6-bit VGA levels: widen to 8 bits, then tint
					t := func(i int, k uint8) uint8 {
						v := int(p[i] % 64)
						return uint8((v<<2 | v>>4) * int(k) / 255)
					}
					answerImg.SetRGBA(x, y, color.RGBA{t(0, c.R), t(1, c.G), t(2, c.B), 255})
				}
			})
		}
	}
}

// answerAt is the answer at screen pixel (x, y), or -1.
func answerAt(x, y int) int {
	i := (y/cellSize-textTop)*textCols + x/cellSize
	for n, a := range answers {
		if answerFound && y >= textTop*cellSize && x >= 0 && x < ega.W && i >= a[0] && i <= a[1] {
			return n
		}
	}
	return -1
}
