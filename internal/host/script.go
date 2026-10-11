package host

import (
	"github.com/hajimehoshi/ebiten/v2"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Scripted input for testing, from U4GO_INPUT: comma-separated events, each
// "tick:what:args" with the tick counted in host ticks from the start:
//
//	k:HEX        a BIOS key word (scan code << 8 | ASCII), e.g. k:1c0d for Enter
//	t:TEXT       typed characters
//	m:X:Y        move the pointer to (X, Y), screen pixels from the top left
//	c:X:Y        move there and click the left button
//	r:X:Y        the same with the right button
//	q            quit
//	s:NAME       save the screen as NAME.png in U4GO_SHOT
//	g            toggle VGA/EGA graphics (F2)
//	h:WHAT       the modern layout's input: enter, up, down, esc, f3, f4, f5, f6, or
//	             tile:DX:DY (a click on the map, relative to the centre tile)
//
// While a script is set, the real mouse and keyboard are ignored.

type scriptEvent struct {
	tick int
	what string
	args []int
	raw  []string
}

var (
	script     []scriptEvent
	scriptTick int
	scriptOn   bool
	scriptHold int // ticks left with the scripted button held
	scriptBtn  int
	scriptQuit bool
)

func init() {
	s := os.Getenv("U4GO_INPUT")
	if s == "" {
		return
	}
	scriptOn = true
	for _, e := range strings.Split(s, ",") {
		f := strings.Split(strings.TrimSpace(e), ":")
		if len(f) < 2 {
			continue
		}
		t, _ := strconv.Atoi(f[0])
		ev := scriptEvent{tick: t, what: f[1], raw: f[2:]}
		for _, a := range f[2:] {
			base := 10
			if ev.what == "k" {
				base = 16
			}
			v, _ := strconv.ParseInt(a, base, 32)
			ev.args = append(ev.args, int(v))
		}
		script = append(script, ev)
	}
}

// runScript feeds the events due this tick; it reports whether a script is running.
func runScript() bool {
	if !scriptOn {
		return false
	}
	scriptTick++
	if scriptHold > 0 {
		scriptHold--
		if scriptHold == 0 {
			MouseButtons = 0
		}
	}
	for len(script) > 0 && script[0].tick <= scriptTick {
		ev := script[0]
		script = script[1:]
		switch ev.what {
		case "k":
			if len(ev.args) > 0 {
				KeyQueue = append(KeyQueue, uint16(ev.args[0]))
			}
		case "t":
			for _, r := range ev.raw {
				for i := 0; i < len(r); i++ {
					PushChar(r[i])
				}
			}
		case "m", "c", "r":
			if len(ev.args) >= 2 {
				MouseX, MouseY = ev.args[0], ev.args[1]
			}
			if ev.what != "m" {
				MouseButtons = 1
				if ev.what == "r" {
					MouseButtons = 2
				}
				scriptHold = 4
			}
		case "q":
			scriptQuit = true
		case "g":
			toggleGraphics()
		case "h":
			scriptHost(ev.raw)
		case "s":
			if len(ev.raw) > 0 {
				saveShot(ev.raw[0])
			}
		}
	}
	return true
}

var shotName string

func saveShot(name string) { shotName = name }

// takeShot saves the window as drawn when a shot is due.
func takeShot(screen *ebiten.Image) {
	dir := os.Getenv("U4GO_SHOT")
	name := shotName
	shotName = ""
	if dir == "" || name == "" {
		return
	}
	b := screen.Bounds()
	img := image.NewRGBA(b)
	screen.ReadPixels(img.Pix)
	out, err := os.Create(filepath.Join(dir, name+".png"))
	if err != nil {
		return
	}
	png.Encode(out, img)
	out.Close()
}

// scriptHost plays the modern layout's input for a script.
func scriptHost(raw []string) {
	if len(raw) == 0 {
		return
	}
	switch raw[0] {
	case "enter":
		if menu != nil {
			choose(menu.Sel)
		} else if modal != 0 {
			modal = 0
		} else {
			openMenu(PromptMenu(true))
		}
	case "up", "down":
		if menu != nil {
			n := len(menu.Items)
			d := 1
			if raw[0] == "up" {
				d = n - 1
			}
			menu.Sel = (menu.Sel + d) % n
		}
	case "esc":
		menu, modal = nil, 0
	case "f3":
		ShowStats = !ShowStats
	case "f4":
		ShowText = !ShowText
	case "f5":
		modal = modalHistory
	case "f6":
		Modern = !Modern
	case "tile":
		if len(raw) == 3 && MapClick != nil {
			dx, _ := strconv.Atoi(raw[1])
			dy, _ := strconv.Atoi(raw[2])
			if keys, ok := MapClick(dx, dy); ok {
				KeyQueue = append(KeyQueue, keys...)
			} else {
				openMenu(PromptMenu(true))
			}
		}
	}
}
