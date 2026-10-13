package title

// Port of INFLATE.ASM: the LZW decoder (12-bit codes, a hashed dictionary of 3840 entries,
// reset after 3277 additions) behind loadPicture (*.EGA pictures) and loadPacked
// (SHAPES.EGZ). Each decoded byte goes to a callback.

import (
	"errors"
	"os"

	"github.com/memmaker/u4-go/internal/gamefs"
)

// The dictionary: codes 0..255 are bytes; codes 256..4095 are a string's prefix code plus
// one more byte. Codes are 12 bits: two codes in three bytes.
const (
	firstCode  = 256
	codeCount  = 4096
	codeMask   = codeCount - 1
	inUse      = 8192 // marks a defined code's prefix
	resetAfter = 3276 // the dictionary starts over once it has more new codes than this
	probeStep  = 509  // the next slot to try after a collision
	endOfData  = 65535
	hashBias   = 2048
)

// lzw decodes one file.
type lzw struct {
	input    []byte                        // what is left to read
	left     uint16                        // bytes left
	oddCode  bool                          // the next code starts halfway into a byte
	nibble   uint8                         // that byte
	prefix   [codeCount - firstCode]uint16 // a code's prefix | inUse, 0 for a free slot
	lastByte [codeCount - firstCode]uint8
	stack    [codeCount]uint8 // a string's bytes, last first
}

// slot finds the dictionary slot for (prefix, b): the one holding it, or the first free one
// along the probe sequence.
func (d *lzw) slot(prefix, b uint16) uint16 { /*C_21E5*/
	match := func(code uint16) bool {
		p := d.prefix[code-firstCode]
		return p == 0 || p&^inUse == prefix && d.lastByte[code-firstCode] == uint8(b)
	}
	// b's high byte takes part in the first hash, though only its low byte is stored (the
	// decoder's quirk: see decode)
	code := (b<<4 ^ prefix) & codeMask
	if code >= firstCode && match(code) {
		return code
	}
	// then bits 6..17 of the square of (2b + prefix) | hashBias
	v := uint32((uint16(uint8(b))<<1 + prefix) | hashBias)
	code = uint16(v*v>>6) & codeMask
	if code >= firstCode && match(code) {
		return code
	}
	for {
		code = (code + probeStep) & codeMask
		if code >= firstCode && match(code) {
			return code
		}
	}
}

// readCode reads the next 12-bit code, endOfData at the end of the input.
func (d *lzw) readCode() uint16 { /*C_2290*/
	var code uint16
	d.left--
	if d.left == endOfData {
		code = endOfData
	} else {
		b := d.input[0]
		d.input = d.input[1:]
		if !d.oddCode {
			d.left--
			if d.left == endOfData {
				code = endOfData
			} else {
				next := d.input[0]
				d.input = d.input[1:]
				d.nibble = next
				code = uint16(b)<<4 | uint16(next>>4)
			}
		} else {
			code = uint16(d.nibble%16)<<8 | uint16(b)
		}
	}
	d.oddCode = !d.oddCode
	return code
}

// decode decodes the input into emit.
func (d *lzw) decode(emit func(uint8)) { /*C_22D5*/
	d.oddCode = false
	clear(d.prefix[:])
	old := d.readCode()
	first := old // the last string's first byte (see below)
	emit(uint8(first))
	count := 0
	for {
		code := d.readCode()
		if code == endOfData {
			return
		}
		if code >= firstCode {
			n := 0
			c := code
			if d.prefix[code-firstCode] == 0 { // the code being defined: old's string + its first byte
				d.stack[n] = uint8(first)
				n++
				c = old
			}
			for c >= firstCode {
				d.stack[n] = d.lastByte[c-firstCode]
				n++
				c = d.prefix[c-firstCode] &^ inUse
			}
			first = c
			emit(uint8(c))
			for n != 0 {
				n--
				emit(d.stack[n])
			}
		} else {
			emit(uint8(code))
			// quirk: only the low byte is replaced, a high byte left from a code stays
			first = first&^255 | code
		}
		s := d.slot(old, first)
		d.lastByte[s-firstCode] = uint8(first)
		d.prefix[s-firstCode] = old | inUse
		count++
		if count > resetAfter {
			count = 0
			clear(d.prefix[:])
			code = d.readCode()
			first = code
			if code == endOfData {
				return
			}
			emit(uint8(code))
		}
		old = code
	}
}

func decodeBytes(b []byte, emit func(uint8)) {
	d := &lzw{input: b, left: uint16(len(b))}
	d.decode(emit)
}

// decodeFile loads a file and decodes it into emit; false when the file can't be read.
func decodeFile(name string, emit func(uint8)) bool { /*C_23C3*/
	b, ok := gamefs.Read(name, SaveDir, GameDir)
	if !ok {
		return false
	}
	decodeBytes(b, emit)
	return true
}

// Picture decodes a 320x200 .EGA picture (LZW, two pixels a byte) to colour indexes, outside
// the game: goplay's cover. No cycle pacing, so it needs no host.
func Picture(path string) ([]byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	pix := make([]byte, 0, picSize)
	decodeBytes(b, func(c uint8) {
		if len(pix) < picSize {
			pix = append(pix, c>>4, c%16)
		}
	})
	if len(pix) < picSize {
		return nil, errors.New(path + ": short picture")
	}
	return pix, nil
}
