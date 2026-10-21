// Command u4 runs Ultima IV (the DOS EGA version) from the original data files: ULTIMA.COM's
// job, TITLE.EXE then AVATAR.EXE, each ported to Go.
//
//	u4                        play with the game files in ./data, saving to ./save (goplay's
//	                          layout: run from the binary's directory; -data and -save override)
//	                          (VGA upgrade graphics; -ega for the original EGA, F2 switches)
//	u4 -import SRC [-data DIR] copy the game files out of SRC (folder, zip, ISO/BIN/CUE, GOG
//	                          installer, loose files) into DIR, write DIR/COVER.PNG and exit
//
// Music needs the GeneralUser GS SoundFont at ../../system/sf2/GeneralUser-GS.sf2. Volume
// is the game's own: Alt-+/Alt-- and Alt-V for the music, the Volume command for the effects.
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/memmaker/u4-go/internal/ega"
	"github.com/memmaker/u4-go/internal/host"
	"github.com/memmaker/u4-go/internal/music"
	"github.com/memmaker/u4-go/internal/speaker"
	"github.com/memmaker/u4-go/internal/title"
	"github.com/memmaker/u4-go/internal/u4"
	"github.com/retro-ware/gameimport"
)

// gameFile says whether name is one of the game's data files.

// The blank saves' sizes: a people table (eight arrays of 32), a dungeon level (8x8 squares
// on 8 levels).
const (
	peopleTableSize  = 8 * 32
	dungeonLevelSize = 8 * 8 * 8
)

func gameFile(name string) bool {
	switch strings.ToUpper(filepath.Ext(name)) {
	case ".EGA", ".PIC", ".CON", ".DNG", ".TLK", ".ULT", ".MAP", ".NEW", ".DRV", ".CGA", ".EGZ":
		return true
	}
	return false
}

func main() {
	game := flag.String("data", "data", "directory with the original game files")
	save := flag.String("save", "save", "directory for saved games")
	imp := flag.String("import", "", "import the game files from this folder/zip/ISO/installer into -data and exit")
	scale := flag.Int("scale", 3, "window scale")
	egaMode := flag.Bool("ega", false, "start with the original EGA graphics instead of the VGA upgrade (F2 toggles)")
	noMusic := flag.Bool("nomusic", false, "no MIDI music (Alt-V toggles it, Alt-+/Alt-- set its volume)")
	flag.Parse()

	if *imp != "" {
		files, err := gameimport.Find(*imp, gameFile)
		if err == nil && !hasAll(files, "WORLD.MAP", "SHAPES.EGA") {
			err = fmt.Errorf("no Ultima IV files (WORLD.MAP) found in %s", *imp)
		}
		if err == nil {
			err = os.MkdirAll(*game, 0o755)
		}
		for name, b := range files {
			if err == nil {
				err = os.WriteFile(filepath.Join(*game, strings.ToUpper(filepath.Base(name))), b, 0o644)
			}
		}
		if err == nil {
			err = writeCover(*game)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, "import:", err)
			os.Exit(1)
		}
		return
	}

	if _, err := os.Stat(filepath.Join(*game, "WORLD.MAP")); err != nil {
		fmt.Fprintf(os.Stderr, "no game files in %s; use -import SRC or -data DIR\n", *game)
		os.Exit(1)
	}
	if err := blankSaves(*game, *save); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	u4.GameDir, u4.SaveDir = *game, *save

	ega.VGA = !*egaMode
	host.ToggleGraphics = func() {
		ega.VGA = !ega.VGA
		title.SetVGA(ega.VGA)
		u4.SetVGA(ega.VGA)
	}
	speaker.Start()
	if !*noMusic {
		music.Start()
		host.AltKey = func(c byte) {
			map[byte]func(){'v': music.Toggle, '+': music.Louder, '-': music.Softer}[c]()
		}
	}
	var gameErr error
	err := host.Run("Ultima IV", *scale, func() {
		// The title, then the game, which comes back to the title when no party is formed.
		for gameErr == nil {
			host.InWorld = false
			if gameErr = title.Main(*game, *save); gameErr == nil {
				gameErr = u4.Main()
			}
		}
	})
	if err == nil {
		err = gameErr
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func hasAll(files map[string][]byte, names ...string) bool {
	have := map[string]bool{}
	for n := range files {
		have[strings.ToUpper(filepath.Base(n))] = true
	}
	for _, n := range names {
		if !have[n] {
			return false
		}
	}
	return true
}

// writeCover decodes TITLE.EGA into dir/COVER.PNG for goplay's floppy label: the
// "Ultima IV / Quest of the Avatar" logo alone, centred on black.
func writeCover(dir string) error {
	pix, err := title.Picture(filepath.Join(dir, "TITLE.EGA"))
	if err != nil {
		return err
	}
	var pal color.Palette
	for _, c := range ega.Palette {
		pal = append(pal, color.RGBA{c[0]<<2 | c[0]>>4, c[1]<<2 | c[1]>>4, c[2]<<2 | c[2]>>4, 255})
	}
	src := &image.Paletted{Pix: pix, Stride: 320, Rect: image.Rect(0, 0, 320, 200), Palette: pal}
	logo := image.Rect(41, 33, 281, 93)
	img := image.NewPaletted(image.Rect(0, 0, 320, 200), pal) // index 0: EGA black
	at := image.Pt((320-logo.Dx())/2, (200-logo.Dy())/2)
	draw.Draw(img, logo.Sub(logo.Min).Add(at), src, logo.Min, draw.Src)
	f, err := os.Create(filepath.Join(dir, "COVER.PNG"))
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

// blankSaves puts the blank saves in save if they are missing: AVATAR.EXE only opens
// existing files for writing. They are PARTY.NEW and zero-filled files.
func blankSaves(game, save string) error {
	if err := os.MkdirAll(save, 0o755); err != nil {
		return err
	}
	party, err := os.ReadFile(filepath.Join(game, "PARTY.NEW"))
	if err != nil {
		return err
	}
	for name, b := range map[string][]byte{
		"PARTY.SAV": party, "MONSTERS.SAV": make([]byte, peopleTableSize),
		"OUTMONST.SAV": make([]byte, peopleTableSize), "DNGMAP.SAV": make([]byte, dungeonLevelSize),
	} {
		p := filepath.Join(save, name)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			if err := os.WriteFile(p, b, 0o644); err != nil {
				return err
			}
		}
	}
	return nil
}
