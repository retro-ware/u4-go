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
	pDir               // askDirection
	pLetter            // askLetter
	pChara             // askMember
	pYN                // askYesNo
	pWords             // readLine with known answers (getsWords)
	pText              // readLine: free text
	pKeys              // a key read with its own answers (askKeys)
)

var (
	prompt     promptKind
	promptText string
	promptLo   int
	promptHi   int
	// getsWords are the answers offered for the next readLine: set by its caller.
	getsWords   []string
	askingChara bool // askMember is asking through askLetter
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

// waiting is set while readKey waits for a key.
var waiting bool

func init() { setHooks() }

// setHooks points the host's menus, clicks and font at the game (the title sets its own).
func setHooks() {
	host.Hover = nil
	host.PromptMenu = promptMenu
	host.MapClick = mapClick
	host.Glyphs = func() [][64]byte { return screen.Glyphs[:] }
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
		return keyEnter
	case pCmd:
		return 0
	}
	return keyEsc
}

// screenClick turns a click on the original layout's map window into a mapClick.
func screenClick(x, y int) ([]uint16, bool) {
	if !host.InWorld || x < 8 || x >= 184 || y < 8 || y >= 184 || gameMode == modeDungeon {
		return nil, false
	}
	keys, ok := mapClick((x-8)/16-5, (y-8)/16-5)
	return keys, ok && keys != nil
}

var talkWords = []string{"Name", "Job", "Health", "Look", "Join", "Give", "Bye"}

// usableItems are the Use command's words for what the party carries.
func usableItems() []string {
	var w []string
	if party.stones != 0 {
		w = append(w, "Stone")
	}
	for _, it := range []struct {
		bit  int
		name string
	}{{itemBell, "Bell"}, {itemBook, "Book"}, {itemCandle, "Candle"}, {itemHorn, "Horn"}, {itemWheel, "Wheel"}, {itemSkull, "Skull"}} {
		if hasBit(party.items, it.bit) {
			w = append(w, it.name)
		}
	}
	if party.items&(1<<itemKeyTruth|1<<itemKeyLove|1<<itemKeyCourage) != 0 {
		w = append(w, "Key")
	}
	return w
}

// stoneColors are the colours of the stones the party carries.
func stoneColors() []string {
	var w []string
	for i, c := range colorNames {
		if hasBit(party.stones, i) {
			w = append(w, c)
		}
	}
	return w
}

// histPut follows printChar into the history line.
func histPut(c byte) {
	if textRow < 12 || textCol < 24 { // only the text window, not the status box or the borders
		return
	}
	switch {
	case c == '\b':
		if len(histLine) > 0 {
			histLine = histLine[:len(histLine)-1]
		}
	case c >= ' ' && c <= lastPrintable:
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
		return keyUp
	case dy > 0:
		return keyDown
	case dx < 0:
		return keyLeft
	}
	return keyRight
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
		for i := 0; i < int(party.size); i++ {
			m.Items = append(m.Items, item(fmt.Sprintf("%d %s", i+1, cstr(party.members[i].name[:])), key(byte('1'+i))))
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
			m.Items = append(m.Items, item(w, append(keys, keyEnter)...))
		}
		m.Items = append(m.Items, host.MenuItem{Label: "Type...", Keyboard: true}, item("(nothing)", keyEnter))
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

// letterMenu offers askLetter's letters, named after the prompt's subject, else after the
// list the game just printed ("A-Dagger" lines of a shop).
func letterMenu() *host.Menu {
	m := &host.Menu{Title: strings.TrimRight(promptText, string(rune(glyphBlank))+"\b :?")}
	name := func(c int) string { return "" }
	named := promptLo >= 'A' // only the named letters are offered
	switch {
	case strings.HasPrefix(promptText, "Spell") || strings.HasPrefix(promptText, "For Spell"):
		name = func(c int) string {
			return fmt.Sprintf("%-10s %2d", spellNames[c-'A'], party.mixtures[c-'A'])
		}
	case strings.HasPrefix(promptText, "Weapon"):
		name = func(c int) string {
			i := c - 'A'
			if i == 0 {
				return weaponNames[0]
			}
			if party.weapons[i] == 0 {
				return ""
			}
			return fmt.Sprintf("%-12s %2d", weaponNames[i], party.weapons[i])
		}
	case strings.HasPrefix(promptText, "Armour"):
		name = func(c int) string {
			i := c - 'A'
			if i == 0 {
				return armorNames[0]
			}
			if party.armors[i] == 0 {
				return ""
			}
			return fmt.Sprintf("%-12s %2d", armorNames[i], party.armors[i])
		}
	case strings.HasPrefix(promptText, "Reagent"):
		name = func(c int) string {
			i := c - 'A'
			if party.reagents[i] == 0 {
				return ""
			}
			return fmt.Sprintf("%-12s %2d", reagentNames[i], party.reagents[i])
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
	if promptText == "Reagent:"+answerRoom(2) {
		m.Items = append(m.Items, item("(done)", keyEnter))
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
	for _, n := range party.mixtures {
		mixtures = mixtures || n != 0
	}
	switch {
	case gameMode >= modeCombat:
		m = append(m, item("Attack", key('a')))
		when(mixtures, "Cast", 'c')
		when(true, "Ready weapon", 'r')
		when(len(usableItems()) > 0, "Use item", 'u')
		when(true, "Ztats", 'z')
	case gameMode == modeDungeon:
		kind := dngKind(hereTile)
		when(kind == dngLadderUp || kind == dngLadderUpDown, "Klimb ladder", 'k')
		when(kind == dngLadderDown || kind == dngLadderUpDown, "Descend ladder", 'd')
		when(kind == dngChest, "Get chest", 'g')
		when(mixtures, "Cast", 'c')
		when(true, "Hole up & camp", 'h')
		when(party.torches != 0, "Ignite torch", 'i')
		when(true, "Mix reagents", 'm')
		when(party.size > 1, "New order", 'n')
		when(party.gems != 0, "Peer at gem", 'p')
		when(true, "Save game", 'q')
		when(true, "Ready weapon", 'r')
		when(true, "Search", 's')
		when(len(usableItems()) > 0, "Use item", 'u')
		when(true, "Wear armour", 'w')
		when(true, "Ztats", 'z')
	default:
		m = worldContext()
		when(party.tile == tileBalloon && party.flyingOrLight == 0, "Klimb (lift off)", 'k')
		when(party.tile == tileBalloon && party.flyingOrLight != 0, "Descend (land)", 'd')
		when(mixtures, "Cast", 'c')
		when(party.loc == 0, "Hole up & camp", 'h')
		when(party.sextants != 0 && party.loc == 0, "Locate", 'l')
		when(true, "Mix reagents", 'm')
		when(party.size > 1, "New order", 'n')
		when(party.gems != 0, "Peer at gem", 'p')
		when(party.loc == 0, "Save game", 'q')
		when(true, "Ready weapon", 'r')
		when(true, "Search", 's')
		when(len(usableItems()) > 0, "Use item", 'u')
		when(true, "Wear armour", 'w')
		when(party.tile == tileHorseWest || party.tile == tileHorseEast, "Yell", 'y')
		when(true, "Ztats", 'z')
	}
	m = append(m, item("Volume", key('v')), item("Pass", keySpace))
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
		x, y := int(party.x)+d.dx, int(party.y)+d.dy
		dk := dirKey(d.dx, d.dy)
		if i := objectAt(uint8(x), uint8(y)); i >= 0 {
			t := area.people.tile[i]
			if party.loc != 0 && area.people.talkIndex[i] != 0 {
				m = append(m, item("Talk "+d.name+": "+creatureName(t), key('t'), dk))
			} else if party.loc == 0 || t >= tilePirateWest {
				m = append(m, item("Attack "+d.name+": "+creatureName(t), key('a'), dk))
			}
		}
		if party.loc == 0 || x < 0 || x > 31 || y < 0 || y > 31 {
			continue
		}
		t := *area.townTile(y, x)
		switch {
		case t >= tileLetterA && t <= tileBlank: // a counter: the merchant behind it
			x2, y2 := x+d.dx, y+d.dy
			if i := objectAt(uint8(x2), uint8(y2)); x2 >= 0 && x2 < 32 && y2 >= 0 && y2 < 32 && i >= 0 &&
				area.people.tile[i] == tileMerchant {
				m = append(m, item("Talk "+d.name+": merchant", key('t'), dk))
			}
		case t == tileDoor && d.dy == 0:
			m = append(m, item("Open door "+d.name, key('o'), dk))
		case t == tileLockedDoor && d.dy == 0:
			m = append(m, item("Jimmy lock "+d.name, key('j'), dk))
		}
	}
	switch {
	case hereTile == tileChest && party.tile == tileOnFoot:
		m = append(m, item("Get chest", key('g')))
	case party.tile == tileOnFoot && (hereTile == tileHorseWest || hereTile == tileHorseEast || hereTile == tileBalloon ||
		(hereTile >= tileShipWest && hereTile <= tileShipSouth)):
		m = append(m, item("Board", key('b')))
	}
	if party.tile != tileOnFoot {
		if party.tile >= tileShipWest && party.tile <= tileShipSouth {
			m = append(m, item("Fire cannon", key('f')))
		}
		if party.tile != tileBalloon || party.flyingOrLight == 0 {
			m = append(m, item("X-it craft", key('x')))
		}
	}
	if party.loc == 0 && party.tile != tileBalloon {
		for i := range placeX {
			if party.x == placeX[i] && party.y == placeY[i] {
				m = append(m, item("Enter "+placeNames[i], key('e')))
			}
		}
	}
	if party.loc == 1 {
		switch hereTile {
		case tileLadderUp:
			m = append(m, item("Klimb ladder", key('k')))
		case tileLadderDown:
			m = append(m, item("Descend ladder", key('d')))
		}
	}
	return m
}

// mapClick turns a click on the map into keys: a step or a direction toward the tile, or
// the key a "press a key" wait wants. The party's own tile opens the command menu.
func mapClick(dx, dy int) ([]uint16, bool) {
	if gameMode >= modeCombat { // the combat map is the view: relative to the active fighter
		a := int(activeMember)
		dx, dy = dx+5-int(battlefield.memberX[a]), dy+5-int(battlefield.memberY[a])
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
		return []uint16{keySpace}, true
	}
	return nil, true
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}
