// Package gamefs resolves DOS file names case-insensitively in the game directory (the
// original files, never written) and the save directory (where every write goes; reads look
// there first). Shared by the TITLE.EXE and AVATAR.EXE ports.
package gamefs

import (
	"encoding/binary"
	"fmt"
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

// Load fills dst from the start of name (found in the first of dirs that has it).
func Load(name string, dst []byte, dirs ...string) error {
	b, ok := Read(name, dirs...)
	if !ok {
		return fmt.Errorf("can't open %s", name)
	}
	if len(b) < len(dst) {
		return fmt.Errorf("%s: file wrong size (%d bytes, want %d)", name, len(b), len(dst))
	}
	copy(dst, b)
	return nil
}

// A Record is a fixed-size file record: Fields walks its fields in file order.
type Record interface{ Fields(*Codec) }

// Codec decodes a Record from little-endian B, or (Enc) encodes it by appending to B.
type Codec struct {
	B   []byte
	Enc bool
}

func (c *Codec) Bytes(p []byte) {
	if c.Enc {
		c.B = append(c.B, p...)
	} else {
		c.B = c.B[copy(p, c.B):]
	}
}

func (c *Codec) U8(p *uint8) {
	if c.Enc {
		c.B = append(c.B, *p)
	} else {
		*p, c.B = c.B[0], c.B[1:]
	}
}

func (c *Codec) U16(p *uint16) {
	if c.Enc {
		c.B = binary.LittleEndian.AppendUint16(c.B, *p)
	} else {
		*p, c.B = binary.LittleEndian.Uint16(c.B), c.B[2:]
	}
}

func (c *Codec) U32(p *uint32) {
	if c.Enc {
		c.B = binary.LittleEndian.AppendUint32(c.B, *p)
	} else {
		*p, c.B = binary.LittleEndian.Uint32(c.B), c.B[4:]
	}
}

func (c *Codec) I16(p *int16) {
	u := uint16(*p)
	c.U16(&u)
	*p = int16(u)
}

func (c *Codec) I32(p *int32) {
	u := uint32(*p)
	c.U32(&u)
	*p = int32(u)
}

func (c *Codec) U16s(p []uint16) {
	for i := range p {
		c.U16(&p[i])
	}
}

// Encode returns r's file image.
func Encode(r Record) []byte {
	c := Codec{Enc: true}
	r.Fields(&c)
	return c.B
}

// Decode fills r from the start of b, which must hold a whole record.
func Decode(r Record, b []byte) error {
	if n := len(Encode(r)); len(b) < n {
		return fmt.Errorf("record needs %d bytes, have %d", n, len(b))
	}
	r.Fields(&Codec{B: b})
	return nil
}

// LoadRecord fills r from the start of name.
func LoadRecord(name string, r Record, dirs ...string) error {
	b := make([]byte, len(Encode(r)))
	if err := Load(name, b, dirs...); err != nil {
		return err
	}
	return Decode(r, b)
}
