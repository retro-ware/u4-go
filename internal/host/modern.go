package host

// The modern layout: on the game's play screen the map window fills the window, the status
// box and the text window float over it as overlays that can be toggled, and menus offer the
// commands and answers that the original wanted typed. A menu only types the original's keys
// for the player, so the game itself runs exactly as before.

import (
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// MenuItem is one choice: the key words it types for the player.
type MenuItem struct {
	Label string
	Keys  []uint16
}

// Menu is a list of choices shown over the map.
type Menu struct {
	Title string
	Items []MenuItem
	Sel   int
	// Pause holds the game while the menu is open (the command prompt passes the turn
	// when left waiting).
	Pause bool
	id    int
}

var (
	// Modern selects the modern layout (F6 switches to the original one).
	Modern = true
	// InWorld is set by the game while the frame shows the play screen.
	InWorld bool
	// ShowStats and ShowText toggle the status box (F3) and text window (F4) overlays.
	ShowStats = true
	ShowText  = true

	// PromptMenu returns the menu for what the game waits for: asked is set when the
	// player asked for it (Enter or the right button); otherwise it is about to open by
	// itself. nil: none.
	PromptMenu func(asked bool) *Menu
	// MapClick returns the keys a click on the map types, the tile given relative to the
	// party; nil keys with handled false open the menu instead.
	MapClick func(dx, dy int) (keys []uint16, handled bool)
	// WideView returns cols x rows tiles of map around the party (16x16 pixels each,
	// palette indexes) for the wider view, or nil when the map window shows something else.
	WideView func(cols, rows int) []byte
	// WideRows is the number of map rows the wider view shows (the original shows 11).
	WideRows = 15
	// Glyphs returns the game's font, 8x8 palette indexes per character.
	Glyphs func() [][64]byte

	// History is every line the text window showed, oldest first (F5 shows it).
	History []string

	promptID, autoID int
	menu             *Menu
	modal            int // 0 none, modalHistory, modalHelp
	histBack         int // lines scrolled back in the history
)

const (
	modalHistory = 1 + iota
	modalHelp
)

// NewPrompt tells the host the game now waits for something else: an open menu for the
// last prompt closes.
func NewPrompt() { promptID++ }

// AddHistory appends a finished text line to the history.
func AddHistory(s string) {
	History = append(History, s)
	if len(History) > 2000 {
		History = History[len(History)-1000:]
	}
}

var helpText = []string{
	"Enter / right click   commands & choices",
	"Click the map         walk / pick a tile",
	"Arrows, letters       as in the original",
	"F3   party overlay",
	"F4   message overlay",
	"F5   message history",
	"F6   original layout",
	"F2   VGA / EGA graphics",
	"F1   this help",
}

func modern() bool { return Modern && InWorld }

func openMenu(m *Menu) {
	if m == nil || len(m.Items) == 0 {
		return
	}
	m.id = promptID
	menu = m
}

func choose(i int) {
	m := menu
	menu = nil
	if i >= 0 && i < len(m.Items) {
		KeyQueue = append(KeyQueue, m.Items[i].Keys...)
	}
}

// updateModern runs before the game's tick: it opens and closes menus. It reports whether
// the game is held this tick.
func updateModern() bool {
	if menu != nil && menu.id != promptID {
		menu = nil
	}
	if menu == nil && modern() && len(KeyQueue) == 0 && autoID != promptID && PromptMenu != nil {
		autoID = promptID
		openMenu(PromptMenu(false))
	}
	return (menu != nil && menu.Pause) || modal != 0
}

// pressed reports a key that went down this tick.
func pressed(k ebiten.Key) bool { return inpututil.IsKeyJustPressed(k) }

// repeating is pressed with typematic repeat while k is held (after a quarter second, 15/s).
func repeating(k ebiten.Key) bool {
	d := inpututil.KeyPressDuration(k)
	return d == 1 || d >= TPS/4 && (d-TPS/4)%(TPS/15) == 0
}

// modernInput handles the modern layout's keys and mouse. It reports whether it used the
// tick's keyboard input, which then does not reach the game.
func modernInput(g *game) bool {
	if pressed(ebiten.KeyF6) {
		Modern = !Modern
		menu, modal = nil, 0
		return true
	}
	if !modern() {
		menu, modal = nil, 0
		return false
	}
	switch {
	case pressed(ebiten.KeyF1):
		modal = modalHelp
		return true
	case pressed(ebiten.KeyF3):
		ShowStats = !ShowStats
		return true
	case pressed(ebiten.KeyF4):
		ShowText = !ShowText
		return true
	case pressed(ebiten.KeyF5):
		modal, histBack = modalHistory, 0
		return true
	}
	if modal != 0 {
		rows := 20
		switch {
		case repeating(ebiten.KeyArrowUp):
			histBack++
		case repeating(ebiten.KeyArrowDown):
			histBack--
		case pressed(ebiten.KeyPageUp):
			histBack += rows
		case pressed(ebiten.KeyPageDown):
			histBack -= rows
		case pressed(ebiten.KeyEscape), pressed(ebiten.KeyEnter), pressed(ebiten.KeyF5),
			inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft),
			inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight):
			modal = 0
		}
		_, wy := ebiten.Wheel()
		histBack += int(wy * 3)
		histBack = max(0, min(histBack, len(History)-rows))
		ebiten.AppendInputChars(nil)
		return true
	}

	mx, my := ebiten.CursorPosition()
	if menu != nil {
		moved := mx != g.lastMouseX || my != g.lastMouseY
		g.lastMouseX, g.lastMouseY = mx, my
		if i, ok := g.menuRow(mx, my); ok {
			if moved {
				menu.Sel = i
			}
			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
				choose(i)
				return true
			}
		} else if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			menu = nil
			return true
		}
		n := len(menu.Items)
		switch {
		case repeating(ebiten.KeyArrowUp):
			menu.Sel = (menu.Sel + n - 1) % n
		case repeating(ebiten.KeyArrowDown):
			menu.Sel = (menu.Sel + 1) % n
		case pressed(ebiten.KeyPageUp), pressed(ebiten.KeyHome):
			menu.Sel = 0
		case pressed(ebiten.KeyPageDown), pressed(ebiten.KeyEnd):
			menu.Sel = n - 1
		case pressed(ebiten.KeyEnter), pressed(ebiten.KeyNumpadEnter):
			choose(menu.Sel)
		case pressed(ebiten.KeyEscape), inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight):
			menu = nil
		default:
			// typing answers the prompt as in the original: the menu steps aside
			if len(ebiten.AppendInputChars(nil)) > 0 {
				menu = nil
				return false
			}
		}
		return true
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) ||
		pressed(ebiten.KeyEnter) || pressed(ebiten.KeyNumpadEnter) {
		if m := PromptMenu(true); m != nil {
			openMenu(m)
			return true
		}
		return false
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && g.world.Dx() > 0 {
		if MapClick != nil {
			t := float64(g.world.Dx()) / 11
			dx := int(math.Floor(float64(mx-g.world.Min.X)/t)) - 5
			dy := int(math.Floor(float64(my-g.world.Min.Y)/t)) - 5
			keys, ok := MapClick(dx, dy)
			if ok {
				KeyQueue = append(KeyQueue, keys...)
			} else if dx == 0 && dy == 0 {
				openMenu(PromptMenu(true))
			}
		}
	}
	return false
}

// text paints lines in the game's font into an image, row sel inverted. pad cells of
// border on every side.
func (g *game) text(lines []string, sel int, title string) *ebiten.Image {
	glyphs := Glyphs()
	w := len(title)
	for _, l := range lines {
		w = max(w, len(l))
	}
	w += 2
	top := 1
	if title != "" {
		top = 2
	}
	h := len(lines) + top + 1
	pix := make([]byte, w*8*h*8*4)
	put := func(cx, cy int, c byte, inv bool) {
		var gl [64]byte
		if int(c) < len(glyphs) {
			gl = glyphs[c]
		}
		for r := 0; r < 8; r++ {
			for k := 0; k < 8; k++ {
				v := gl[r*8+k]
				rgb := g.lut[v]
				if inv && v == 0 { // the selection: the letters on the colour of the border's pipe
					rgb = g.lut[g.pix[3*ScreenW+20]]
				}
				o := ((cy*8+r)*w*8 + cx*8 + k) * 4
				copy(pix[o:o+4], rgb[:])
			}
		}
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			put(x, y, ' ', false)
		}
	}
	for i := 0; i < len(title); i++ {
		put(1+i, 0, title[i], false)
	}
	for j, l := range lines {
		for x := 0; x < w-2; x++ {
			c := byte(' ')
			if x < len(l) {
				c = l[x]
			}
			put(1+x, top+j, c, j == sel)
		}
	}
	img := ebiten.NewImage(w*8, h*8)
	img.WritePixels(pix)
	return img
}

// The pieces of the original screen's border (8x8 cells of the play screen) that frame
// every overlay, menu and window: the corner stud and the pipes.
var (
	cellCorner = image.Rect(0, 0, 8, 8)
	cellTop    = image.Rect(16, 0, 24, 8)
	cellLeft   = image.Rect(0, 40, 8, 48)
	cellRight  = image.Rect(312, 40, 320, 48)
)

// piece draws the border cell c stretched over w x h window pixels at (x, y).
func (g *game) piece(screen *ebiten.Image, c image.Rectangle, x, y, w, h float64) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(w/8, h/8)
	op.GeoM.Translate(x, y)
	screen.DrawImage(g.img.SubImage(c).(*ebiten.Image), op)
}

// drawBox draws img scaled by k with its top left at (x, y), in the original's border
// (one cell, 8k, around it). It returns the content's place.
func (g *game) drawBox(screen, img *ebiten.Image, x, y, k float64) image.Rectangle {
	b := img.Bounds()
	w, h, e := float64(b.Dx())*k, float64(b.Dy())*k, 8*k
	r := image.Rect(int(x), int(y), int(x+w), int(y+h))
	screen.SubImage(r).(*ebiten.Image).Fill(color.Black)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(k, k)
	op.GeoM.Translate(x, y)
	screen.DrawImage(img, op)
	g.piece(screen, cellTop, x, y-e, w, e)
	g.piece(screen, cellTop, x, y+h, w, e)
	g.piece(screen, cellLeft, x-e, y, e, h)
	g.piece(screen, cellRight, x+w, y, e, h)
	for _, c := range [][2]float64{{x - e, y - e}, {x + w, y - e}, {x - e, y + h}, {x + w, y + h}} {
		g.piece(screen, cellCorner, c[0], c[1], e, e)
	}
	return r
}

// drawPart draws a part of the play screen that has its own border as it is.
func (g *game) drawPart(screen *ebiten.Image, part image.Rectangle, x, y, k float64) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(k, k)
	op.GeoM.Translate(x, y)
	screen.DrawImage(g.img.SubImage(part).(*ebiten.Image), op)
}

// drawModern draws the play screen in the modern layout.
func (g *game) drawModern(screen *ebiten.Image) {
	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()
	// the map window (x, y 8..183) as large as fits, centred
	k := float64(min(sw, sh)) / 176
	k = float64(int(k*16)) / 16
	var wide []byte
	cols, rows := 0, WideRows
	if WideView != nil {
		k = float64(sh) / float64(rows*16)
		cols = int(float64(sw)/(16*k)) | 1
		cols += 2 // a partly shown column on each side
		if wide = WideView(cols, rows); wide != nil {
			img := ebiten.NewImage(cols*16, rows*16)
			pix := make([]byte, len(wide)*4)
			for i, c := range wide {
				copy(pix[i*4:i*4+4], g.lut[c][:])
			}
			img.WritePixels(pix)
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(-float64(cols*16)/2, -float64(rows*16)/2)
			op.GeoM.Scale(k, k)
			op.GeoM.Translate(float64(sw)/2, float64(sh)/2)
			screen.DrawImage(img, op)
			img.Deallocate()
		} else {
			k = float64(int(float64(min(sw, sh))/176*16)) / 16
		}
	}
	size := int(176 * k)
	wx, wy := (sw-size)/2, (sh-size)/2
	g.world = image.Rect(wx, wy, wx+size, wy+size)
	if d := dungeon(); d != nil {
		g.drawDungeon(screen, d)
	} else {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(k, k)
		op.GeoM.Translate(float64(wx), float64(wy))
		screen.DrawImage(g.img.SubImage(image.Rect(8, 8, 184, 184)).(*ebiten.Image), op)
	}

	// overlays: about a third of the map's scale, at least 2; the parts of the play screen
	// that have their own border are shown with it
	ok := max(2, float64(int(k/1.6)))
	fw, fh := float64(sw), float64(sh)
	if ShowStats {
		// the party box with food and gold, and the moons and the wind bars of the map
		g.drawPart(screen, image.Rect(184, 0, 320, 96), fw-136*ok, 0, ok)
		g.drawPart(screen, image.Rect(0, 0, 192, 8), 0, 0, ok)
		g.drawPart(screen, image.Rect(0, 184, 192, 200), 0, fh-16*ok, ok)
	}
	if ShowText {
		g.drawBox(screen, g.img.SubImage(image.Rect(192, 96, 320, 192)).(*ebiten.Image),
			fw-136*ok, fh-104*ok, ok)
	}

	g.menuRect = image.Rectangle{}
	switch {
	case modal == modalHelp:
		g.centred(screen, g.text(helpText, -1, "Keys"), ok, sw, sh)
	case modal == modalHistory:
		rows := 20
		end := len(History) - histBack
		lines := History[max(0, end-rows):max(0, end)]
		g.centred(screen, g.text(lines, -1, "Messages (arrows, wheel; Esc)"), ok, sw, sh)
	case menu != nil:
		img := g.text(labels(menu), menu.Sel, menu.Title)
		g.menuK = min(ok, float64(sh)*0.9/float64(img.Bounds().Dy()))
		g.menuTop = 2
		if menu.Title == "" {
			g.menuTop = 1
		}
		g.menuRect = g.centred(screen, img, g.menuK, sw, sh)
	}
}

func labels(m *Menu) []string {
	s := make([]string, len(m.Items))
	for i, it := range m.Items {
		s[i] = it.Label
	}
	return s
}

func (g *game) centred(screen, img *ebiten.Image, k float64, sw, sh int) image.Rectangle {
	b := img.Bounds()
	x := (float64(sw) - float64(b.Dx())*k) / 2
	y := (float64(sh) - float64(b.Dy())*k) / 2
	return g.drawBox(screen, img, x, y, k)
}

// menuRow is the menu row under window pixel (x, y).
func (g *game) menuRow(x, y int) (int, bool) {
	if menu == nil || !image.Pt(x, y).In(g.menuRect) {
		return 0, false
	}
	i := int(float64(y-g.menuRect.Min.Y)/(8*g.menuK)) - g.menuTop
	return i, i >= 0 && i < len(menu.Items)
}

func dungeon() *Dungeon {
	if DungeonView == nil {
		return nil
	}
	return DungeonView()
}
