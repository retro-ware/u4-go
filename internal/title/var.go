package title

// Port of VAR.C (the communals) and the record layouts of TITLE.H.

import "unsafe"

// tParty is kept as its raw 0x1f6 bytes (tChara is 0x27 bytes, odd, so a Go struct would
// pad): the fields TITLE.EXE touches are offsets into it.
const (
	sizeofParty = 0x1f6
	sizeofChara = 0x27

	party_f_000 = 0x000 // unsigned long counter
	party_chara = 0x008 // struct tChara chara[8]
	party_hones = 0x146 // the eight karmas, U16 each
	party_x     = 0x1d4
	party_y     = 0x1d5
	party_f_1d8 = 0x1d8 // characters #

	chara_str  = 0x06
	chara_dex  = 0x08
	chara_int  = 0x0a
	chara_name = 0x14 // char[16]
	chara_p_24 = 0x24
)

// tNPC (0x100 bytes): eight arrays of 32.
type tNPC struct {
	_000, _020, _040, _060, _080, _0A0, _0C0, _0E0 [32]uint8
}

// t_500: the map (unused by TITLE.EXE) and the npc table.
type t_500 struct {
	_map [0x400]uint8
	_npc tNPC
}

func (n *tNPC) bytes() []byte { return (*[0x100]byte)(unsafe.Pointer(n))[:] }

var (
	D_6E80   uint16            /*D_6E80*/
	Party    [sizeofParty]byte /*D_6E82*/
	D_7078   uint16            /*D_7078 graphic adapter: always 2 (EGA)*/
	pShapes  []byte            /*D_707A*/
	pAnim    []byte            /*D_707E*/
	pCharset []byte            /*D_7084*/
	txt_X    uint16            /*D_7088*/
	txt_Y    uint16            /*D_708A*/
	pTitle   []byte            /*D_708C*/

	D_6976 t_500 // UTIL2.ASM's _BSS
)

func put16(b []byte, o int, v uint16) { b[o], b[o+1] = byte(v), byte(v>>8) }
