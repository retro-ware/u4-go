// Package gamefs resolves DOS file names case-insensitively in the game directory (the
// original files, never written) and the save directory (where every write goes; reads look
// there first). Shared by the TITLE.EXE and AVATAR.EXE ports.
package gamefs

import (
	"os"
	"path/filepath"
	"strings"
)

// Find returns the path of name in dir, matching case-insensitively.
func Find(dir, name string) (string, bool) {
	if dir == "" {
		return "", false
	}
	p := filepath.Join(dir, name)
	if _, err := os.Stat(p); err == nil {
		return p, true
	}
	ents, err := os.ReadDir(dir)
	if err != nil {
		return "", false
	}
	for _, e := range ents {
		if strings.EqualFold(e.Name(), name) && !e.IsDir() {
			return filepath.Join(dir, e.Name()), true
		}
	}
	return "", false
}

// Read returns a file's contents from the first of dirs that has it.
func Read(name string, dirs ...string) ([]byte, bool) {
	for _, d := range dirs {
		if p, ok := Find(d, name); ok {
			if b, err := os.ReadFile(p); err == nil {
				return b, true
			}
		}
	}
	return nil, false
}

// Write stores data as name in dir, over an existing file of any case, else as a new
// upper-case name.
func Write(dir, name string, data []byte) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	p, ok := Find(dir, name)
	if !ok {
		p = filepath.Join(dir, strings.ToUpper(name))
	}
	return os.WriteFile(p, data, 0o644)
}
