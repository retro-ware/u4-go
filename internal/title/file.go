package title

// Port of FILE.C: Load and Save. Files resolve case-insensitively (internal/gamefs) in
// SaveDir, then GameDir. Writes go to SaveDir only.

import (
	"fmt"

	"github.com/memmaker/u4-go/internal/gamefs"
)

var (
	// GameDir holds the original game files (never written).
	GameDir = "."
	// SaveDir is where PARTY.SAV and MONSTERS.SAV are written.
	SaveDir = "."
)

// fail unwinds the title goroutine to Main with err.
type fail struct{ err error }

func fatal(err error)                { panic(fail{err}) }
func fatalf(format string, a ...any) { fatal(fmt.Errorf(format, a...)) }
func check(err error) {
	if err != nil {
		fatal(err)
	}
}

// Load fills dst from the start of a file.
func Load(name string, dst []byte) { check(gamefs.Load(name, dst, SaveDir, GameDir)) } /*C_0051*/

// LoadRec fills a record from the start of a file.
func LoadRec(name string, r gamefs.Record) { check(gamefs.LoadRecord(name, r, SaveDir, GameDir)) }

// Save writes a record as name in SaveDir.
func Save(name string, r gamefs.Record) { /*C_0110*/
	if err := gamefs.Write(SaveDir, name, gamefs.Encode(r)); err != nil {
		fatalf("writing %s: %v", name, err)
	}
}
