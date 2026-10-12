package u4

// The modern layout's knowledge of the game: what the game is waiting for, the menus that
// answer it, the commands that make sense where the party stands, and the text window's
// lines for the message history. Menus only type the keys a player would have typed: the
// game's rules are untouched.

import (
	"fmt"
	"strings"

	"github.com/memmaker/u4-go/internal/host"
)

type promptKind int

const (
	pNone   promptKind = iota
	pCmd               // a command loop's key (world, dungeon or combat)
	pDir               // AskDir
	pLetter            // AskLetter
	pChara             // AskChara
	pYN                // AskY_N
	pWords             // u4_gets with known answers (getsWords)
	pText              // u4_gets: free text
	pKeys              // a key read with its own answers (askKeys)
)

var (
	prompt     promptKind
	promptText string
	promptLo   int
	promptHi   int
	// getsWords are the answers offered for the next u4_gets: set by its caller.
	getsWords   []string
	askingChara bool // AskChara is asking through AskLetter
	// promptItems are pKeys' answers.
	promptItems []host.MenuItem
	// histLine is the text window's line being written, for the history.
	histLine []byte
)

// ask tells the host what the game is about to wait for.
func ask(k promptKind) {
	prompt = k
	if k == pCmd {
		host.InWorld = true
	}
	host.NewPrompt()
}

// waiting is set while u_kbread waits for a key.
var waiting bool

func init() { setHooks() }

// setHooks points the host's menus, clicks and font at the game (the title sets its own).
func setHooks() {
	host.Hover = nil
	host.PromptMenu = promptMenu
	host.MapClick = mapClick
	host.Glyphs = func() [][64]byte { return drv.Glyphs[:] }
	host.BackKey = backKey
	host.ScreenClick = screenClick
}

// askKeys tells the host the next key read takes one of items.
func askKeys(title string, items ...host.MenuItem) {
	promptText, promptItems = title, items
	ask(pKeys)
}

// backKey is what B types: Enter where an empty line cancels, nothing at a command, else Esc.
func backKey() uint16 {
	switch prompt {
	case pWords, pText:
		return KBD_ENTER
	case pCmd:
		return 0
	}
	return KBD_ESC
}

// screenClick turns a click on the original layout's map window into a mapClick.
func screenClick(x, y int) ([]uint16, bool) {
	if !host.InWorld || x < 8 || x >= 184 || y < 8 || y >= 184 || CurMode == MOD_DUNGEON {
		return nil, false
	}
	keys, ok := mapClick((x-8)/16-5, (y-8)/16-5)
	return keys, ok && keys != nil
}

var talkWords = []string{"Name", "Job", "Health", "Look", "Join", "Give", "Bye"}

// usableItems are the Use command's words for what the party carries.
func usableItems() []string {
	var w []string
	if Party.mStones != 0 {
		w = append(w, "Stone")
	}
	for _, it := range []struct {
		bit  int
		name string
	}{{ST_BELL, "Bell"}, {ST_BOOK, "Book"}, {ST_CANDLE, "Candle"}, {ST_HORN, "Horn"}, {ST_WHEEL, "Wheel"}, {ST_SKULL, "Skull"}} {
		if TST_MSK(Party.mItems, it.bit) {
			w = append(w, it.name)
		}
	}
	if Party.mItems&(1<<ST_KEY_T|1<<ST_KEY_L|1<<ST_KEY_C) != 0 {
		w = append(w, "Key")
	}
	return w
}

// stoneColors are the colours of the stones the party carries.
func stoneColors() []string {
	var w []string
	for i, c := range D_0884 {
		if TST_MSK8(Party.mStones, i) {
			w = append(w, c)
		}
	}
	return w
}

// histPut follows u4_putc into the history line.
func histPut(c byte) {
	if txt_Y < 12 || txt_X < 24 { // only the text window, not the status box or the borders
		return
	}
	switch {
	case c == '\b':
		if len(histLine) > 0 {
			histLine = histLine[:len(histLine)-1]
		}
	case c >= ' ' && c < 0x7f:
		histLine = append(histLine, c)
	}
}

// histCR ends the history line at a carriage return of the text window.
func histCR() {
	host.AddHistory(string(histLine))
	histLine = histLine[:0]
}

func dirKey(dx, dy int) uint16 {
	switch {
	case dy < 0:
		return KBD_UP
	case dy > 0:
		return KBD_DOWN
	case dx < 0:
		return KBD_LEFT
	}
	return KBD_RIGHT
}

var dirNames = []struct {
	dx, dy int
	name   string
}{{0, -1, "North"}, {0, 1, "South"}, {1, 0, "East"}, {-1, 0, "West"}}

func key(c byte) uint16 { return host.KeyWord(c) }

func item(label string, keys ...uint16) host.MenuItem {
	return host.MenuItem{Label: label, Keys: keys}
}

func promptMenu(asked bool) *host.Menu {
	switch prompt {
	case pCmd:
		if !asked {
			return nil
		}
		return &host.Menu{Title: "Command", Items: commandItems(), Pause: true}
	case pDir:
		m := &host.Menu{Title: "Direction"}
		for _, d := range dirNames {
			m.Items = append(m.Items, item(d.name, dirKey(d.dx, d.dy)))
		}
		return m
	case pChara:
		m := &host.Menu{Title: "Who"}
		for i := 0; i < int(Party.f_1d8); i++ {
			m.Items = append(m.Items, item(fmt.Sprintf("%d %s", i+1, cstr(Party.chara[i]._name[:])), key(byte('1'+i))))
		}
		return m
	case pYN:
		return &host.Menu{Items: []host.MenuItem{item("Yes", key('y')), item("No", key('n'))}}
	case pLetter:
		return letterMenu()
	case pWords:
		m := &host.Menu{Title: "Say (or type)"}
		for _, w := range getsWords {
			keys := make([]uint16, 0, len(w)+1)
			for i := 0; i < len(w); i++ {
				keys = append(keys, key(w[i]))
			}
			m.Items = append(m.Items, item(w, append(keys, KBD_ENTER)...))
		}
		m.Items = append(m.Items, host.MenuItem{Label: "Type...", Keyboard: true}, item("(nothing)", KBD_ENTER))
		return m
	case pText:
		if !asked {
			return nil
		}
		return &host.Menu{Items: []host.MenuItem{{Label: "Type...", Keyboard: true}}}
	case pKeys:
		return &host.Menu{Title: promptText, Items: promptItems}
	}
	return nil
}

// letterMenu offers AskLetter's letters, named after the prompt's subject, else after the
// list the game just printed ("A-Dagger" lines of a shop).
func letterMenu() *host.Menu {
	m := &host.Menu{Title: strings.TrimRight(promptText, "\x12\b :?")}
	name := func(c int) string { return "" }
	named := promptLo >= 'A' // only the named letters are offered
	switch {
	case strings.HasPrefix(promptText, "Spell") || strings.HasPrefix(promptText, "For Spell"):
		name = func(c int) string {
			return fmt.Sprintf("%-10s %2d", D_1E98[101+c-'A'], Party._mixtures[c-'A'])
		}
	case strings.HasPrefix(promptText, "Weapon"):
		name = func(c int) string {
			i := c - 'A'
			if i == 0 {
				return D_1E98[37]
			}
			if Party._weapons[i] == 0 {
				return ""
			}
			return fmt.Sprintf("%-12s %2d", D_1E98[37+i], Party._weapons[i])
		}
	case strings.HasPrefix(promptText, "Armour"):
		name = func(c int) string {
			i := c - 'A'
			if i == 0 {
				return D_1E98[53]
			}
			if Party._armors[i] == 0 {
				return ""
			}
			return fmt.Sprintf("%-12s %2d", D_1E98[53+i], Party._armors[i])
		}
	case strings.HasPrefix(promptText, "Reagent"):
		name = func(c int) string {
			i := c - 'A'
			if Party._reagents[i] == 0 {
				return ""
			}
			return fmt.Sprintf("%-12s %2d", D_1E98[93+i], Party._reagents[i])
		}
	case promptLo >= 'A':
		listed := printedChoices()
		name = func(c int) string { return listed[byte(c)] }
		named = false // nothing listed: every letter
		for c := promptLo; c <= promptHi; c++ {
			if listed[byte(c)] != "" {
				named = true
			}
		}
	}
	for c := promptLo; c <= promptHi; c++ {
		label := name(c)
		if named && label == "" {
			continue
		}
		if label == "" {
			label = string(rune(c))
		} else {
			label = string(rune(c)) + " " + label
		}
		m.Items = append(m.Items, item(label, key(byte(c))))
	}
	if promptText == "Reagent:\x12\x12\b" {
		m.Items = append(m.Items, item("(done)", KBD_ENTER))
	}
	return m
}

// printedChoices reads "A-Name" or "A: Name" lines from the recent text.
func printedChoices() map[byte]string {
	got := map[byte]string{}
	h := host.History
	for i := len(h) - 1; i >= 0 && i >= len(h)-24; i-- {
		l := strings.TrimSpace(h[i])
		if len(l) > 2 && l[0] >= 'A' && l[0] <= 'Z' && (l[1] == '-' || l[1] == ':' || l[1] == ')' ||
			l[1] == ' ' && l[2] >= 'A' && l[2] <= 'Z') { // "B Cloth" (the armourer)
			if _, ok := got[l[0]]; !ok {
				got[l[0]] = strings.TrimSpace(l[2:])
			}
		}
	}
	return got
}

// commandItems are the commands that apply where the party is: first what it can do with
// its neighbours and the ground it stands on, then what it can do anywhere with what it has.
func commandItems() []host.MenuItem {
	var m []host.MenuItem
	when := func(ok bool, label string, c byte) {
		if ok {
			m = append(m, item(label, key(c)))
		}
	}
	mixtures := false
	for _, n := range Party._mixtures {
		mixtures = mixtures || n != 0
	}
	switch {
	case CurMode >= MOD_COMBAT:
		m = append(m, item("Attack", key('a')))
		when(mixtures, "Cast", 'c')
		when(true, "Ready weapon", 'r')
		when(len(usableItems()) > 0, "Use item", 'u')
		when(true, "Ztats", 'z')
	case CurMode == MOD_DUNGEON:
		t := tile_cur & 0xf0
		when(t == 0x10 || t == 0x30, "Klimb ladder", 'k')
		when(t == 0x20 || t == 0x30, "Descend ladder", 'd')
		when(t == 0x40, "Get chest", 'g')
		when(mixtures, "Cast", 'c')
		when(true, "Hole up & camp", 'h')
		when(Party._torches != 0, "Ignite torch", 'i')
		when(true, "Mix reagents", 'm')
		when(Party.f_1d8 > 1, "New order", 'n')
		when(Party._gems != 0, "Peer at gem", 'p')
		when(true, "Save game", 'q')
		when(true, "Ready weapon", 'r')
		when(true, "Search", 's')
		when(len(usableItems()) > 0, "Use item", 'u')
		when(true, "Wear armour", 'w')
		when(true, "Ztats", 'z')
	default:
		m = worldContext()
		when(Party._tile == TIL_18 && Party.f_1dc == 0, "Klimb (lift off)", 'k')
		when(Party._tile == TIL_18 && Party.f_1dc != 0, "Descend (land)", 'd')
		when(mixtures, "Cast", 'c')
		when(Party._loc == 0, "Hole up & camp", 'h')
		when(Party._sextants != 0 && Party._loc == 0, "Locate", 'l')
		when(true, "Mix reagents", 'm')
		when(Party.f_1d8 > 1, "New order", 'n')
		when(Party._gems != 0, "Peer at gem", 'p')
		when(Party._loc == 0, "Save game", 'q')
		when(true, "Ready weapon", 'r')
		when(true, "Search", 's')
		when(len(usableItems()) > 0, "Use item", 'u')
		when(true, "Wear armour", 'w')
		when(Party._tile == TIL_14 || Party._tile == TIL_15, "Yell", 'y')
		when(true, "Ztats", 'z')
	}
	m = append(m, item("Volume", key('v')), item("Pass", KBD_SPACE))
	all := &host.Menu{Title: "All commands", Pause: true}
	for i, name := range allCommands {
		all.Items = append(all.Items, item(name, key(byte('a'+i))))
	}
	return append(m, host.MenuItem{Label: "All commands...", Sub: all})
}

// allCommands are the original's commands, A to Z, for what the context list leaves out.
var allCommands = []string{"Attack", "Board", "Cast", "Descend", "Enter", "Fire", "Get chest",
	"Hole up & camp", "Ignite torch", "Jimmy lock", "Klimb", "Locate", "Mix reagents", "New order",
	"Open door", "Peer at gem", "Quit & save", "Ready weapon", "Search", "Talk", "Use item",
	"Volume", "Wear armour", "X-it", "Yell", "Ztats"}

// worldContext lists what the party can do with its neighbours and the tile it is on.
func worldContext() []host.MenuItem {
	var m []host.MenuItem
	for _, d := range dirNames {
		x, y := int(Party._x)+d.dx, int(Party._y)+d.dy
		dk := dirKey(d.dx, d.dy)
		if i := C_0A8B(uint8(x), uint8(y)); i >= 0 {
			t := D_8742._npc._tile[i]
			if Party._loc != 0 && D_8742._npc._tlkidx[i] != 0 {
				m = append(m, item("Talk "+d.name+": "+C_1513(t), key('t'), dk))
			} else if Party._loc == 0 || t >= TIL_80 {
				m = append(m, item("Attack "+d.name+": "+C_1513(t), key('a'), dk))
			}
		}
		if Party._loc == 0 || x < 0 || x > 31 || y < 0 || y > 31 {
			continue
		}
		t := *D_8742.x32x32(y, x)
		switch {
		case t >= TIL_60 && t <= TIL_7E: // a counter: the merchant behind it
			x2, y2 := x+d.dx, y+d.dy
			if i := C_0A8B(uint8(x2), uint8(y2)); x2 >= 0 && x2 < 32 && y2 >= 0 && y2 < 32 && i >= 0 &&
				D_8742._npc._tile[i] == TIL_52 {
				m = append(m, item("Talk "+d.name+": merchant", key('t'), dk))
			}
		case t == TIL_3B && d.dy == 0:
			m = append(m, item("Open door "+d.name, key('o'), dk))
		case t == TIL_3A && d.dy == 0:
			m = append(m, item("Jimmy lock "+d.name, key('j'), dk))
		}
	}
	switch {
	case tile_cur == TIL_3C && Party._tile == TIL_1F:
		m = append(m, item("Get chest", key('g')))
	case Party._tile == TIL_1F && (tile_cur == TIL_14 || tile_cur == TIL_15 || tile_cur == TIL_18 ||
		(tile_cur >= TIL_10 && tile_cur <= TIL_13)):
		m = append(m, item("Board", key('b')))
	}
	if Party._tile != TIL_1F {
		if Party._tile >= TIL_10 && Party._tile <= TIL_13 {
			m = append(m, item("Fire cannon", key('f')))
		}
		if Party._tile != TIL_18 || Party.f_1dc == 0 {
			m = append(m, item("X-it craft", key('x')))
		}
	}
	if Party._loc == 0 && Party._tile != TIL_18 {
		for i := range D_0844 {
			if Party._x == D_0844[i] && Party._y == D_0864[i] {
				m = append(m, item("Enter "+D_1E98[127+i], key('e')))
			}
		}
	}
	if Party._loc == 1 {
		switch tile_cur {
		case TIL_1B:
			m = append(m, item("Klimb ladder", key('k')))
		case TIL_1C:
			m = append(m, item("Descend ladder", key('d')))
		}
	}
	return m
}

// mapClick turns a click on the map into keys: a step or a direction toward the tile, or
// the key a "press a key" wait wants. The party's own tile opens the command menu.
func mapClick(dx, dy int) ([]uint16, bool) {
	if CurMode >= MOD_COMBAT { // the combat map is the view: relative to the active fighter
		a := int(activeChara)
		dx, dy = dx+5-int(Combat._charaX[a]), dy+5-int(Combat._charaY[a])
	}
	if abs(dx) >= abs(dy) {
		dy = 0
	} else {
		dx = 0
	}
	switch {
	case prompt == pCmd && dx == 0 && dy == 0:
		return nil, false
	case prompt == pCmd || prompt == pDir && (dx != 0 || dy != 0):
		return []uint16{dirKey(dx, dy)}, true
	case waiting && prompt == pNone:
		return []uint16{KBD_SPACE}, true
	}
	return nil, true
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}
