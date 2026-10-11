package u4

// Port of U4_DOS.ASM and U4_FILE.C: the DOS file calls on the game's files.
//
// Files resolve case-insensitively in GameDir (the original data files, never written) and
// SaveDir (the saved game: PARTY.SAV, MONSTERS.SAV, OUTMONST.SAV, DNGMAP.SAV, ...). A file
// opened for writing (DOS open mode 1: an existing file, written over in place) is read from
// wherever it is and written back to SaveDir on close. Reads prefer SaveDir.

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"unsafe"
)

var (
	// GameDir holds the original game files.
	GameDir = "."
	// SaveDir is where the game's writes go (and reads look first).
	SaveDir = "."
)

func findFile(dir, name string) (string, bool) {
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

// readGameFile returns a file's contents, from SaveDir or GameDir.
func readGameFile(name string) ([]byte, bool) {
	for _, d := range []string{SaveDir, GameDir} {
		if p, ok := findFile(d, name); ok {
			b, err := os.ReadFile(p)
			if err == nil {
				return b, true
			}
		}
	}
	return nil, false
}

type dosFile struct {
	name  string
	data  []byte
	pos   int
	write bool
	dirty bool
}

var files = map[int16]*dosFile{}

func dalloc(n int) []byte { return make([]byte, n) } /*C_1FF4*/
func dfree(p []byte)      {}                         /*C_2023*/

func dopen(name string, mode int) int16 { /*C_2033*/
	b, ok := readGameFile(name)
	if !ok {
		return -1
	}
	h := int16(5)
	for files[h] != nil {
		h++
	}
	files[h] = &dosFile{name: strings.ToUpper(name), data: b, write: mode&3 != 0}
	return h
}

func dlseek(fd int16, ofs uint32) int32 { /*C_2049*/
	f := files[fd]
	if f == nil {
		return -1
	}
	f.pos = int(ofs)
	return int32(ofs)
}

// dread reads up to n bytes into dst (a byte slice, or a pointer to a fixed-size value,
// whose first n bytes are replaced). It returns the count read, -1 on error.
func dread(fd int16, dst any, n int) int { /*C_2067*/
	f := files[fd]
	if f == nil {
		return -1
	}
	end := min(f.pos+n, len(f.data))
	var got []byte
	if f.pos < end {
		got = f.data[f.pos:end]
	}
	f.pos += len(got)
	storeBytes(dst, got)
	return len(got)
}

func dwrite(fd int16, src any, n int) int { /*C_208B*/
	b := loadBytes(src)
	if n < len(b) {
		b = b[:n]
	}
	if fd == 1 || fd == 2 {
		os.Stderr.Write(b)
		return len(b)
	}
	f := files[fd]
	if f == nil || !f.write {
		return -1
	}
	if need := f.pos + len(b); need > len(f.data) {
		f.data = append(f.data, make([]byte, need-len(f.data))...)
	}
	copy(f.data[f.pos:], b)
	f.pos += len(b)
	f.dirty = true
	return len(b)
}

func dclose(fd int16) int { /*C_20AB*/
	f := files[fd]
	if f == nil {
		return -1
	}
	delete(files, fd)
	if f.dirty {
		if err := os.MkdirAll(SaveDir, 0o755); err != nil {
			return -1
		}
		p, ok := findFile(SaveDir, f.name)
		if !ok {
			p = filepath.Join(SaveDir, f.name)
		}
		if err := os.WriteFile(p, f.data, 0o644); err != nil {
			return -1
		}
	}
	return 0
}

// loadBytes returns the memory image of v: a byte slice, or a pointer to a fixed-size value.
func loadBytes(v any) []byte {
	if b, ok := v.([]byte); ok {
		return b
	}
	var buf bytes.Buffer
	if err := binary.Write(&buf, binary.LittleEndian, v); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

// storeBytes replaces the first len(b) bytes of v's memory image with b.
func storeBytes(v any, b []byte) {
	if d, ok := v.([]byte); ok {
		copy(d, b)
		return
	}
	img := loadBytes(v)
	copy(img, b)
	setLE(reflect.ValueOf(v).Elem(), img)
}

// setLE fills v from little-endian img in field order, like binary.Read, but also into the
// unexported fields the records use (binary.Read refuses those). Returns the rest of img.
func setLE(v reflect.Value, img []byte) []byte {
	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			img = setLE(v.Field(i), img)
		}
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			img = setLE(v.Index(i), img)
		}
	default:
		n := int(v.Type().Size())
		var u uint64
		for i := n - 1; i >= 0; i-- {
			u = u<<8 | uint64(img[i])
		}
		w := reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem()
		if w.CanInt() {
			w.SetInt(int64(u) << (64 - 8*n) >> (64 - 8*n))
		} else {
			w.SetUint(u)
		}
		img = img[n:]
	}
	return img
}

// ---- U4_FILE.C ----

func Console(s string) { /*C_0812*/
	if dwrite(1, []byte(s), len(s)) != len(s) {
		exit(3)
	}
}

// Load reads size bytes of a file into v (C: Load(name, size, far ptr)).
func Load(name string, size int, v any) int { /*C_0855*/
	fd := dopen(name, 0)
	if fd == -1 {
		Console("\r\nOpening ")
		Console(name)
		return -1
	}
	n := dread(fd, v, size)
	if n == -1 || n != size {
		Console("\r\nReading ")
		Console(name)
		if n != -1 {
			Console(": File wrong size\r\n")
		}
		return -1
	}
	if dclose(fd) == -1 {
		Console("\r\nClosing ")
		Console(name)
		return -1
	}
	return 0
}

// Save writes size bytes of v over a file (which must exist).
func Save(name string, size int, v any) int { /*C_08D1*/
	fd := dopen(name, 1)
	if fd == -1 {
		Console("\r\nOpening ")
		Console(name)
		return -1
	}
	n := dwrite(fd, v, size)
	if n == -1 || n != size {
		Console("\r\nWriting ")
		Console(name)
		if n != -1 {
			Console(": File wrong size\r\n")
		}
		exit(1)
	}
	if dclose(fd) == -1 {
		Console("\r\nClosing ")
		Console(name)
		return -1
	}
	return 0
}

// sizeof is C's sizeof for the records Load/Save take.
func sizeof(v any) int { return binary.Size(v) }

var _ = fmt.Sprint
