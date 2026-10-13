package replay

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/memmaker/u4-go/internal/ega"
	"github.com/memmaker/u4-go/internal/host"
	"github.com/memmaker/u4-go/internal/speaker"
	"github.com/memmaker/u4-go/internal/title"
	"github.com/memmaker/u4-go/internal/u4"
)

// The replay test pins the game's observable behaviour: each scenario feeds keys tick by tick
// to a headless game (title, then the game) and records every new screen, every speaker note
// and the files in the save folder at the end. The traces in testdata were recorded before
// the rewrite into readable Go; every rewrite must reproduce them exactly.
//
//	U4_DATA=~/Projects/u4-decompiled/gog/inst go test ./internal/replay
//	U4_UPDATE=1 ...   re-records the traces
//	U4_PNG=dir ...    (with U4_SCENARIO=name) also writes each new screen as a PNG

type scenario struct {
	name   string
	ticks  int
	ega    bool   // the original EGA graphics instead of the VGA upgrade
	script []step // keys at fixed ticks
	seed   int64  // after the script: random keys from this seed (0: none)
	keys   string // the random keys (default randomKeys)
	place  int    // 1..32: a rich party (richParty) starts outside place place-1 and enters it
	shop   int    // 1..: in a town, the party starts at counter shop-1 (counters)
}

type step struct {
	tick int
	keys string // printable characters; '\r' Enter, '\033' Esc; "↑↓←→" the arrow keys
}

// typed is keys pressed one every 35 ticks (half a second) from tick start.
func typed(start int, keys string) []step {
	var s []step
	for i, k := range []rune(keys) {
		s = append(s, step{start + 35*i, string(k)})
	}
	return s
}

// randomKeys are what the random player presses: the commands, menu answers and walking.
const randomKeys = "abcdefghijklmnopqrstuvwxyz0123456789\r\033 ↑↓←→" +
	"\r\r\r123123yyynnn↑↑↑↑↑↑↓↓↓↓↓↓←←←←←←→→→→→→eeekkk"

// newGame starts a new game: Initiate, a name and a sex, then the gypsy's story and her
// questions, every key of which is answer (the questions flush the keys typed ahead).
func newGame(answer string) []step {
	s := []step{{300, " "}, {400, "i"}}
	s = append(s, typed(500, "AVATAR\rm")...)
	return append(s, typed(1000, strings.Repeat(answer, 150))...)
}

var scenarios = []scenario{
	{name: "title", ticks: 6000},
	{name: "title-ega", ticks: 6000, ega: true},
	{name: "newgame", ticks: 7000, script: newGame("a")},
}

// journey loads the saved party (Journey Onward) and enters the place it stands at.
var journey = []step{{300, " "}, {400, "j"}, {1500, "e"}}

// placeKeys add talking (words the people answer to), buying, casting and the dungeon's
// commands to the random player's keys.
const placeKeys = randomKeys + "tttt↑↓←→name\rjob\rhealth\rbye\r\rbbbsss" +
	"cccccc1234abcdefghijklmnopqrstuvwxyz\r\rdddkkkkuuuhhhppppggggooommmyyy"

// shopKeys talk to the merchant (across the counter, the arrow added per counter) and answer.
const shopKeys = "tttttabcdefghijklmnop0123456789yyynnnbbbsss\r\r\r\r\033 "

func init() {
	for i := 1; i <= 49; i++ { // the towns have 49 counters
		scenarios = append(scenarios, scenario{name: fmt.Sprintf("shop%02d", i-1), ticks: 40000, script: journey, seed: int64(i) * 7, keys: shopKeys, shop: i})
	}
	for p := 1; p <= 32; p++ {
		for seed := int64(1); seed <= 2; seed++ {
			scenarios = append(scenarios, scenario{name: fmt.Sprintf("place%02d-%d", p-1, seed), ticks: 60000, script: journey, seed: seed + int64(p)*100, keys: placeKeys, place: p})
		}
	}
	for seed := int64(1); seed <= 16; seed++ {
		answer := string("ab"[seed%2])
		scenarios = append(scenarios, scenario{name: fmt.Sprint("play", seed), ticks: 100000, ega: seed%4 == 0, script: newGame(answer), seed: seed})
	}
}

func TestReplay(t *testing.T) {
	if os.Getenv("U4_DATA") == "" {
		t.Skip("U4_DATA (the original game folder) not set")
	}
	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			t.Parallel()
			cmd := exec.Command(os.Args[0], "-test.run=^TestReplayChild$")
			cmd.Env = append(os.Environ(), "U4_SCENARIO="+sc.name)
			out, err := cmd.Output()
			if err != nil {
				t.Fatalf("child: %v\n%s", err, out)
			}
			_, trace, _ := strings.Cut(string(out), "--trace--\n")
			trace, _, _ = strings.Cut(trace, "--end--\n")
			golden := filepath.Join("testdata", sc.name+".trace")
			if os.Getenv("U4_UPDATE") != "" {
				os.MkdirAll(filepath.Dir(golden), 0o755)
				if err := os.WriteFile(golden, []byte(trace), 0o644); err != nil {
					t.Fatal(err)
				}
				return
			}
			want, err := os.ReadFile(golden)
			if err != nil {
				t.Fatal(err)
			}
			if trace != string(want) {
				g, w := strings.Split(trace, "\n"), strings.Split(string(want), "\n")
				for i := 0; i < len(g) && i < len(w); i++ {
					if g[i] != w[i] {
						t.Fatalf("diverges at line %d:\n got  %s\n want %s", i+1, g[i], w[i])
					}
				}
				t.Fatalf("trace length %d lines, want %d", len(g), len(w))
			}
		})
	}
}

func TestReplayChild(t *testing.T) {
	name := os.Getenv("U4_SCENARIO")
	if name == "" {
		t.Skip("run by TestReplay")
	}
	var sc scenario
	for _, s := range scenarios {
		if s.name == name {
			sc = s
		}
	}
	if sc.name == "" {
		t.Fatalf("no scenario %q", name)
	}
	game := os.Getenv("U4_DATA")
	save := t.TempDir()
	if err := blankSaves(game, save); err != nil {
		t.Fatal(err)
	}
	if sc.place != 0 {
		if err := richParty(game, save, sc.place-1); err != nil {
			t.Fatal(err)
		}
	}
	if sc.shop != 0 {
		cs, err := counters(game)
		if err != nil {
			t.Fatal(err)
		}
		if sc.shop > len(cs) {
			t.Skipf("only %d counters", len(cs))
		}
		c := cs[sc.shop-1]
		if err := richParty(game, save, c.town); err != nil {
			t.Fatal(err)
		}
		u4.StartInTown = func(int) (byte, byte, bool) { return c.x, c.y, true }
		keys := []rune(sc.keys)
		for i := 0; i < 8; i++ {
			keys = append(keys, c.toward)
		}
		sc.keys = string(keys)
	}
	u4.GameDir, u4.SaveDir = game, save
	ega.VGA = !sc.ega
	host.Modern = false
	pngDir := os.Getenv("U4_PNG")

	// Every second of play is one line: a digest of the screens shown and the notes played
	// in it, tick by tick.
	const ticksPerLine = host.TPS
	var out bytes.Buffer
	tick := 0
	second, changes := sha1.New(), 0
	speaker.NoteHook = func(n speaker.Note) {
		fmt.Fprintf(second, "%d note %.1f %d\n", tick, n.Hz, n.Dur.Microseconds())
		changes++
	}
	rng := rand.New(rand.NewSource(sc.seed))
	keys := []rune(randomKeys)
	if sc.keys != "" {
		keys = []rune(sc.keys)
	}
	lastScripted := 0
	if len(sc.script) > 0 {
		lastScripted = sc.script[len(sc.script)-1].tick
	}
	next := 0 // tick of the next random key
	var lastFrame [20]byte

	host.StartHeadless(func() {
		for {
			host.InWorld = false
			if title.Main(game, save) != nil || u4.Main() != nil {
				return
			}
		}
	})
	for ; tick < sc.ticks; tick++ {
		for _, s := range sc.script {
			if s.tick == tick {
				for _, k := range s.keys {
					pushKey(k)
				}
			}
		}
		if sc.seed != 0 && tick > lastScripted && tick >= next {
			pushKey(keys[rng.Intn(len(keys))])
			next = tick + 5 + rng.Intn(40)
		}
		if !host.Step() {
			fmt.Fprintf(&out, "%d exit\n", tick)
			break
		}
		f := host.ShowFrame()
		h := sha1.New()
		h.Write(f.Pix)
		h.Write(f.Palette[:])
		var sum [20]byte
		copy(sum[:], h.Sum(nil))
		if sum != lastFrame {
			lastFrame = sum
			fmt.Fprintf(second, "%d frame %x\n", tick, sum)
			changes++
			if pngDir != "" {
				writePNG(t, filepath.Join(pngDir, fmt.Sprintf("%06d.png", tick)), f)
			}
		}
		if (tick+1)%ticksPerLine == 0 && changes > 0 {
			fmt.Fprintf(&out, "%d %d %x\n", tick+1-ticksPerLine, changes, second.Sum(nil)[:8])
			second.Reset()
			changes = 0
		}
	}
	var names []string
	filepath.Walk(save, func(p string, fi os.FileInfo, err error) error {
		if err == nil && !fi.IsDir() {
			names = append(names, p)
		}
		return nil
	})
	sort.Strings(names)
	for _, p := range names {
		b, _ := os.ReadFile(p)
		rel, _ := filepath.Rel(save, p)
		fmt.Fprintf(&out, "file %s %x\n", rel, sha1.Sum(b))
	}
	fmt.Printf("--trace--\n%s--end--\n", out.String())
}

// blankSaves gives the save folder the files a new game starts from (as cmd/u4 does).
func blankSaves(game, save string) error {
	party, err := os.ReadFile(filepath.Join(game, "PARTY.NEW"))
	if err != nil {
		return err
	}
	for name, b := range map[string][]byte{
		"PARTY.SAV": party, "MONSTERS.SAV": make([]byte, 256),
		"OUTMONST.SAV": make([]byte, 256), "DNGMAP.SAV": make([]byte, 512),
	} {
		if err := os.WriteFile(filepath.Join(save, name), b, 0o644); err != nil {
			return err
		}
	}
	return nil
}

// pushKey queues key k as the BIOS key word the game reads.
func pushKey(k rune) {
	const (
		scanEnter, scanEsc                    = 28, 1
		scanUp, scanDown, scanLeft, scanRight = 72, 80, 75, 77
	)
	scan := map[rune]uint16{'↑': scanUp, '↓': scanDown, '←': scanLeft, '→': scanRight}
	switch {
	case scan[k] != 0:
		host.KeyQueue = append(host.KeyQueue, scan[k]<<8)
	case k == '\r':
		host.KeyQueue = append(host.KeyQueue, scanEnter<<8|'\r')
	case k == '\033':
		host.KeyQueue = append(host.KeyQueue, scanEsc<<8|27)
	default:
		host.PushChar(byte(k))
	}
}

func writePNG(t *testing.T, path string, f *host.Frame) {
	img := image.NewPaletted(image.Rect(0, 0, host.ScreenW, host.ScreenH), nil)
	for i := 0; i < 256; i++ {
		p := f.Palette[i*3 : i*3+3]
		img.Palette = append(img.Palette, color.RGBA{p[0]<<2 | p[0]>>4, p[1]<<2 | p[1]>>4, p[2]<<2 | p[2]>>4, 255})
	}
	copy(img.Pix, f.Pix)
	var b bytes.Buffer
	png.Encode(&b, img)
	os.MkdirAll(filepath.Dir(path), 0o755)
	if err := os.WriteFile(path, b.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
}
