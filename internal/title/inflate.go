package title

// Port of INFLATE.ASM: the LZW decoder (12-bit codes, hashed dictionary of 0F00h entries,
// reset after 0CCDh additions) behind Gra_inflate (*.EGA pictures) and C_217E (SHAPES.EGZ).
// Each decoded byte goes to a callback.

import (
	"errors"
	"os"

	"github.com/memmaker/u4-go/internal/gamefs"
)

var (
	D_3B20 uint16        // deflated bytes left
	D_3B22 []byte        // deflated file buffer (read position)
	D_3B24 uint8         // 12 bits flag
	D_3B25 uint8         // temporary stores 4 bits
	D_3B26 [0xf00]uint16 // prefix code | 2000h (in use), for codes 100h..FFFh
	D_5926 [0xf00]uint8  // last byte, for codes 100h..FFFh
	D_6826 [0x1000]uint8 // output stack (0x100 bytes in the original; a longer chain ran into the rest of _BSS)
)

func tab(code uint16) uint16 { return D_3B26[code-0x100] }

// C_21E5 finds the dictionary slot for (prefix bp04, byte bp06): the slot holding it, or the
// first free one along the probe sequence.
func C_21E5(bp04, bp06 uint16) uint16 {
	match := func(ax uint16) (stop bool) {
		cx := tab(ax)
		if cx == 0 {
			return true
		}
		return cx&^0x2000 == bp04 && D_5926[ax-0x100] == uint8(bp06)
	}
	ax := (bp06<<4 ^ bp04) & 0x0fff
	if ax>>8 != 0 && match(ax) {
		return ax
	}
	// C_221C: AX = ((byte(bp06)*2 + bp04) | 800h)^2, bits 10..21 taken through two rcl pairs.
	// MUL leaves the carry set (DX is never 0 for a square >= 800h^2): it enters the first rcl.
	v := (uint16(uint8(bp06))<<1 + bp04) | 0x800
	p := uint32(v) * uint32(v)
	lo, hi := uint16(p), uint16(p>>16)
	cf := uint16(1)
	rcl := func(x *uint16) {
		n := *x >> 15
		*x = *x<<1 | cf
		cf = n
	}
	rcl(&lo)
	rcl(&hi)
	rcl(&lo)
	rcl(&hi)
	ax = (hi&0xf)<<8 | lo>>8
	if ax>>8 != 0 && match(ax) {
		return ax
	}
	for { // C_2261
		ax = (ax + 0x1fd) & 0x0fff
		if ax>>8 == 0 {
			continue
		}
		if match(ax) {
			return ax
		}
	}
}

// C_2290 reads the next 12-bit code, FFFFh at the end of the buffer.
func C_2290() uint16 {
	var ax uint16
	D_3B20--
	if D_3B20 == 0xffff {
		ax = 0xffff
	} else {
		al := D_3B22[0]
		D_3B22 = D_3B22[1:]
		if D_3B24 == 0 {
			D_3B20--
			if D_3B20 == 0xffff {
				ax = 0xffff
			} else {
				dl := D_3B22[0]
				D_3B22 = D_3B22[1:]
				D_3B25 = dl
				ax = uint16(al)<<4 | uint16(dl>>4)
			}
		} else {
			ax = uint16(D_3B25&0xf)<<8 | uint16(al)
		}
	}
	D_3B24 ^= 1
	return ax
}

// C_22D5 decodes D_3B22 into cb.
func C_22D5(cb func(uint8)) {
	var code, old, fin, count uint16 // [bp-2], [bp-6], [bp-8], [bp-4]
	D_3B24 = 0
	clear(D_3B26[:])
	old = C_2290()
	fin = old
	cb(uint8(fin))
	for {
		code = C_2290()
		if code == 0xffff {
			return
		}
		if code>>8 != 0 {
			si := 0
			ax := code
			if tab(code) == 0 { // the code being defined: old's string + its first byte
				D_6826[si] = uint8(fin)
				si++
				ax = old
			}
			for ax>>8 != 0 { // C_232F
				D_6826[si] = D_5926[ax-0x100]
				si++
				ax = tab(ax) &^ 0x2000
			}
			fin = ax
			cb(uint8(ax))
			for si != 0 {
				si--
				cb(D_6826[si])
			}
		} else {
			cb(uint8(code))
			fin = fin&0xff00 | code&0xff // MOV [BP-08],AL: the high byte stays
		}
		// C_2362
		bx := C_21E5(old, fin)
		D_5926[bx-0x100] = uint8(fin)
		D_3B26[bx-0x100] = old | 0x2000
		count++
		if count > 0xccc {
			count = 0
			clear(D_3B26[:])
			code = C_2290()
			fin = code
			if code == 0xffff {
				return
			}
			cb(uint8(code))
		}
		old = code // C_23B3
	}
}

// C_23C3 loads a file and decodes it into cb: 0, or -1 when the file can't be opened.
func C_23C3(name string, cb func(uint8)) int {
	b, ok := gamefs.Read(name, SaveDir, GameDir)
	if !ok {
		return -1
	}
	D_3B20 = uint16(len(b))
	D_3B22 = b
	C_22D5(cb)
	return 0
}

// Picture decodes a 320x200 .EGA picture (LZW, two pixels a byte) to colour indexes, outside
// the game: goplay's cover. No cycle pacing, so it needs no host.
func Picture(path string) ([]byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	D_3B20, D_3B22 = uint16(len(b)), b
	pix := make([]byte, 0, 320*200)
	C_22D5(func(c uint8) {
		if len(pix) < 320*200 {
			pix = append(pix, c>>4, c&15)
		}
	})
	if len(pix) < 320*200 {
		return nil, errors.New(path + ": short picture")
	}
	return pix, nil
}
