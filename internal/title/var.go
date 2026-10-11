package title

// Port of VAR.C (the communals) and the record layouts of TITLE.H.

import "github.com/memmaker/u4-go/internal/gamefs"

// tChara (0x27 bytes): the fields TITLE.EXE sets are named, the rest kept as read.
type tChara struct {
	_000          [6]uint8 // HP, max HP, XP
	str, dex, int uint16
	_00c          [8]uint8 // MP, padding, weapon, armor
	name          [16]uint8
	p_24          uint8    // sex char
	_025          [2]uint8 // class, status
}

// tParty (0x1f6 bytes, PARTY.NEW / PARTY.SAV).
type tParty struct {
	f_000  uint32 // a counter
	_004   [4]uint8
	chara  [8]tChara
	_140   [6]uint8  // food, gold
	karmas [8]uint16 // honesty .. humility
	_156   [0x7e]uint8
	x, y   uint8
	_1d6   [2]uint8 // stones, runes
	f_1d8  uint16   // characters #
	_1da   [0x1c]uint8
}

func (h *tChara) Fields(c *gamefs.Codec) {
	c.Bytes(h._000[:])
	c.U16(&h.str)
	c.U16(&h.dex)
	c.U16(&h.int)
	c.Bytes(h._00c[:])
	c.Bytes(h.name[:])
	c.U8(&h.p_24)
	c.Bytes(h._025[:])
}

func (p *tParty) Fields(c *gamefs.Codec) {
	c.U32(&p.f_000)
	c.Bytes(p._004[:])
	for i := range p.chara {
		p.chara[i].Fields(c)
	}
	c.Bytes(p._140[:])
	c.U16s(p.karmas[:])
	c.Bytes(p._156[:])
	c.U8(&p.x)
	c.U8(&p.y)
	c.Bytes(p._1d6[:])
	c.U16(&p.f_1d8)
	c.Bytes(p._1da[:])
}

// tNPC (0x100 bytes): eight arrays of 32.
type tNPC struct {
	_000, _020, _040, _060, _080, _0A0, _0C0, _0E0 [32]uint8
}

func (n *tNPC) Fields(c *gamefs.Codec) {
	for _, a := range []*[32]uint8{&n._000, &n._020, &n._040, &n._060, &n._080, &n._0A0, &n._0C0, &n._0E0} {
		c.Bytes(a[:])
	}
}

// t_500: the map (unused by TITLE.EXE) and the npc table.
type t_500 struct {
	_map [0x400]uint8
	_npc tNPC
}

var (
	D_6E80   uint16 /*D_6E80*/
	Party    tParty /*D_6E82*/
	D_7078   uint16 /*D_7078 graphic adapter: always 2 (EGA)*/
	pShapes  []byte /*D_707A*/
	pAnim    []byte /*D_707E*/
	pCharset []byte /*D_7084*/
	txt_X    uint16 /*D_7088*/
	txt_Y    uint16 /*D_708A*/
	pTitle   []byte /*D_708C*/

	D_6976 t_500 // UTIL2.ASM's _BSS
)

func init() {
	if n := len(gamefs.Encode(&Party)); n != 0x1f6 {
		panic("tParty layout")
	}
}
