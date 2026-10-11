package title

// Port of FILE.C: Console, Load and Save.
//
// Files resolve case-insensitively (internal/gamefs) in SaveDir, then GameDir. Writes go to
// SaveDir only.

import (
	"os"

	"github.com/memmaker/u4-go/internal/gamefs"
)

var (
	// GameDir holds the original game files (never written).
	GameDir = "."
	// SaveDir is where PARTY.SAV and MONSTERS.SAV are written.
	SaveDir = "."
)

func Console(s string) { /*C_0010*/
	if n, err := os.Stderr.WriteString(s); err != nil || n != len(s) {
		exit(3)
	}
}

// Load reads size bytes of a file into dst.
func Load(name string, size int, dst []byte) int { /*C_0051*/
	b, ok := gamefs.Read(name, SaveDir, GameDir)
	if !ok {
		Console("\r\nOpening ")
		Console(name)
		return -1
	}
	if len(b) < size { // read() came back short
		Console("\r\nReading ")
		Console(name)
		Console(": File wrong size\r\n")
		return -1
	}
	copy(dst[:size], b)
	return 0
}

// Save writes size bytes of src over the start of a file (open O_WRONLY, no truncation).
// The original needs the file to exist already: the port creates it in SaveDir instead.
func Save(name string, size int, src []byte) int { /*C_0110*/
	b, _ := gamefs.Read(name, SaveDir, GameDir)
	if len(b) < size {
		b = append(b, make([]byte, size-len(b))...)
	}
	copy(b, src[:size])
	if gamefs.Write(SaveDir, name, b) != nil {
		Console("\r\nWriting ")
		Console(name)
		exit(1)
	}
	return 0
}
