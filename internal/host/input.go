package host

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// The keyboard is fed to the game as the BIOS keyboard buffer saw it (int 16h): one word per
// key press, the scan code in the high byte and the ASCII code in the low byte. The game reads
// it through bioskey (see the u4 package).

// KeyQueue holds key words the game has not read yet.
var KeyQueue []uint16

// Shift is the BIOS shift state (bioskey(2)): bit 0/1 right/left shift, bit 2 ctrl, bit 3 alt.
var Shift int

type bios struct{ scan, ascii, shifted, ctrl byte }

// BIOS scan codes of the special keys.
const (
	scanEscape    = 1
	scanBackspace = 14
	scanTab       = 15
	scanEnter     = 28
	scanSpace     = 57
	scanF1        = 59 // F1..F10 follow in order
	scanHome      = 71
	scanUp        = 72
	scanPageUp    = 73
	scanLeft      = 75
	scanRight     = 77
	scanEnd       = 79
	scanDown      = 80
	scanPageDown  = 81
	scanInsert    = 82
	scanDelete    = 83
)

// ASCII control characters the keys type.
const (
	asciiEscape   = 27
	asciiDelete   = 127
	asciiLineFeed = 10
)

// keys not delivered through AppendInputChars (or whose BIOS word carries a scan code the
// game tests): scan code and the ASCII byte plain / with shift / with ctrl
var specialKeys = map[ebiten.Key]bios{
	ebiten.KeyEscape:      {scanEscape, asciiEscape, asciiEscape, asciiEscape},
	ebiten.KeyBackspace:   {scanBackspace, '\b', '\b', asciiDelete},
	ebiten.KeyTab:         {scanTab, '\t', 0, 0},
	ebiten.KeyEnter:       {scanEnter, '\r', '\r', asciiLineFeed},
	ebiten.KeyNumpadEnter: {scanEnter, '\r', '\r', asciiLineFeed},
	ebiten.KeyF1:          {scanF1, 0, 0, 0},
	ebiten.KeyF3:          {scanF1 + 2, 0, 0, 0},
	ebiten.KeyF4:          {scanF1 + 3, 0, 0, 0},
	ebiten.KeyF5:          {scanF1 + 4, 0, 0, 0},
	ebiten.KeyF6:          {scanF1 + 5, 0, 0, 0},
	ebiten.KeyF7:          {scanF1 + 6, 0, 0, 0},
	ebiten.KeyF8:          {scanF1 + 7, 0, 0, 0},
	ebiten.KeyF9:          {scanF1 + 8, 0, 0, 0},
	ebiten.KeyF10:         {scanF1 + 9, 0, 0, 0},
	ebiten.KeyHome:        {scanHome, 0, '7', 0},
	ebiten.KeyArrowUp:     {scanUp, 0, '8', 0},
	ebiten.KeyPageUp:      {scanPageUp, 0, '9', 0},
	ebiten.KeyArrowLeft:   {scanLeft, 0, '4', 0},
	ebiten.KeyArrowRight:  {scanRight, 0, '6', 0},
	ebiten.KeyEnd:         {scanEnd, 0, '1', 0},
	ebiten.KeyArrowDown:   {scanDown, 0, '2', 0},
	ebiten.KeyPageDown:    {scanPageDown, 0, '3', 0},
	ebiten.KeyInsert:      {scanInsert, 0, '0', 0},
	ebiten.KeyDelete:      {scanDelete, 0, '.', 0},
}

// scan codes of the printable keys, by unshifted ASCII
var asciiScan = map[byte]byte{}

// KeyWord is the BIOS key word of typing c: its scan code and c.
func KeyWord(c byte) uint16 {
	l := c
	if c >= 'A' && c <= 'Z' {
		l = c + 'a' - 'A'
	}
	return uint16(asciiScan[l])<<8 | uint16(c)
}

func init() {
	rows := []struct {
		first byte
		keys  string
	}{{2, "1234567890-="}, {16, "qwertyuiop[]"}, {30, "asdfghjkl;'`"}, {43, "\\zxcvbnm,./"}}
	for _, r := range rows {
		for i := 0; i < len(r.keys); i++ {
			asciiScan[r.keys[i]] = r.first + byte(i)
		}
	}
	asciiScan[' '] = scanSpace
	for k, v := range map[byte]byte{'!': '1', '@': '2', '#': '3', '$': '4', '%': '5', '^': '6', '&': '7', '*': '8',
		'(': '9', ')': '0', '_': '-', '+': '=', '{': '[', '}': ']', ':': ';', '"': '\'', '~': '`', '|': '\\',
		'<': ',', '>': '.', '?': '/'} {
		asciiScan[k] = asciiScan[v]
	}
	for c := byte('A'); c <= 'Z'; c++ {
		asciiScan[c] = asciiScan[c+32]
	}
}

// PushChar queues the key word of a printable character.
func PushChar(c byte) { KeyQueue = append(KeyQueue, uint16(asciiScan[c])<<8|uint16(c)) }

var keysDown = map[ebiten.Key]bool{}

// ToggleGraphics, when set, is called on F2 (not passed to the game): VGA/EGA graphics.
var ToggleGraphics func()

func toggleGraphics() {
	if ToggleGraphics != nil {
		ToggleGraphics()
	}
}

// AltKey, when set, gets the host's own Alt hot keys: 'v', '-' and '+' (the upgrade patch's
// music toggle and volume keys).
var AltKey func(c byte)

var altKeys = map[ebiten.Key]byte{
	ebiten.KeyV: 'v', ebiten.KeyMinus: '-', ebiten.KeyNumpadSubtract: '-',
	ebiten.KeyEqual: '+', ebiten.KeyNumpadAdd: '+',
}

// Mouse state, in the game's 320x200 screen (top row 0).
var (
	MouseX, MouseY int
	MouseButtons   int // bit 0 left, bit 1 right
)

func pollInput(g *game) {
	if runScript() {
		uiInput(g)
		return
	}
	Shift = 0
	if ebiten.IsKeyPressed(ebiten.KeyShiftRight) {
		Shift |= 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyShiftLeft) {
		Shift |= 2
	}
	ctrl := ebiten.IsKeyPressed(ebiten.KeyControl) || ebiten.IsKeyPressed(ebiten.KeyMeta)
	if ctrl {
		Shift |= 4
	}
	if ebiten.IsKeyPressed(ebiten.KeyAlt) {
		Shift |= 8
	}
	if len(inpututil.AppendJustPressedKeys(nil)) > 0 {
		padLast = false
	}
	readPad()
	readPointer()
	altUsed := modernInput()
	uiInput(g)
	if Shift&8 != 0 && AltKey != nil {
		for k, c := range altKeys {
			if inpututil.IsKeyJustPressed(k) {
				AltKey(c)
				altUsed = true
			}
		}
	}
	if f2 := ebiten.IsKeyPressed(ebiten.KeyF2); f2 != keysDown[ebiten.KeyF2] {
		if keysDown[ebiten.KeyF2] = f2; f2 {
			toggleGraphics()
		}
	}
	for k, b := range specialKeys {
		down := ebiten.IsKeyPressed(k)
		if down && !keysDown[k] && !altUsed {
			a := b.ascii
			if Shift&3 != 0 {
				a = b.shifted
			} else if ctrl {
				a = b.ctrl
			}
			KeyQueue = append(KeyQueue, uint16(b.scan)<<8|uint16(a))
		}
		keysDown[k] = down
	}
	if ctrl { // ctrl+letter: the control code (the game's ^S, ^Q, ...)
		for k := ebiten.KeyA; k <= ebiten.KeyZ; k++ {
			down := ebiten.IsKeyPressed(k)
			if down && !keysDown[k] {
				c := byte('a' + k - ebiten.KeyA)
				KeyQueue = append(KeyQueue, uint16(asciiScan[c])<<8|uint16(c-'a'+1))
			}
			keysDown[k] = down
		}
	} else if !altUsed { // an Alt hot key types nothing (Alt+V would be the game's V)
		for _, r := range ebiten.AppendInputChars(nil) {
			if r >= ' ' && r <= asciiDelete { // ASCII only
				PushChar(byte(r))
			}
		}
	}
	if len(KeyQueue) > 16 { // the BIOS buffer holds 15 keys
		KeyQueue = KeyQueue[len(KeyQueue)-16:]
	}

	cx, cy := ebiten.CursorPosition()
	if g.scaleX > 0 {
		MouseX = min(max(int((float64(cx)-g.offX)/g.scaleX), 0), ScreenW-1)
		MouseY = min(max(int((float64(cy)-g.offY)/g.scaleY), 0), ScreenH-1)
	}
	b := 0
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		b |= 1
	}
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
		b |= 2
	}
	MouseButtons = b
}

// Timers are callbacks run every host tick before the game runs, as the DOS timer
// interrupt ran them.
var timers []func()

// AddTimer registers a callback run once per host tick.
func AddTimer(f func()) { timers = append(timers, f) }

func tickTimers() {
	for _, f := range timers {
		f()
	}
}
