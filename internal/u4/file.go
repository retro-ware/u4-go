package u4

// Port of U4_FILE.C: the game's files, read whole from GameDir (the original data files,
// never written) or SaveDir (the saved game, written there; reads look there first), file
// names matched case-insensitively. The records the files hold encode and decode field by
// field (internal/gamefs.Codec), the C layout byte for byte.

import (
	"errors"
	"fmt"

	"github.com/memmaker/u4-go/internal/gamefs"
	"github.com/memmaker/u4-go/internal/upgrade"
)

var (
	// GameDir holds the original game files.
	GameDir = "."
	// SaveDir is where the saved game goes (and reads look first).
	SaveDir = "."
)

// stop unwinds the game goroutine to Main: err nil means back to the title.
type stop struct{ err error }

func fatal(err error)                { panic(stop{err}) }
func fatalf(format string, a ...any) { fatal(fmt.Errorf(format, a...)) }
func driverExit(code int)            { fatalf("EGA driver error %#x", code) }
func backToTitle()                   { panic(stop{}) }
func check(err error) {
	if err != nil {
		fatal(err)
	}
}

// readGameFile returns a file's contents, from SaveDir or GameDir.
// The VGA upgrade patch's corrected .TLK files win over the game's, in either mode.
func readGameFile(name string) ([]byte, bool) {
	if b, ok := upgrade.TLK(name); ok {
		return b, true
	}
	return gamefs.Read(name, SaveDir, GameDir)
}

// readFile returns a whole file, which must exist.
func readFile(name string) []byte {
	b, ok := readGameFile(name)
	if !ok {
		fatal(errors.New("can't open " + name))
	}
	return b
}

// at is b from ofs on (nil past its end): a seek and a read that may come back short.
func at(b []byte, ofs int) []byte {
	if ofs > len(b) {
		return nil
	}
	return b[ofs:]
}

// Load fills dst from the start of name.
func Load(name string, dst []byte) { check(gamefs.Load(name, dst, SaveDir, GameDir)) }

// LoadRec fills a record from the start of name.
func LoadRec(name string, r gamefs.Record) { check(gamefs.LoadRecord(name, r, SaveDir, GameDir)) }

// Save writes a record as name in SaveDir.
func Save(name string, r gamefs.Record) { SaveBytes(name, gamefs.Encode(r)) }

func SaveBytes(name string, b []byte) {
	if err := gamefs.Write(SaveDir, name, b); err != nil {
		fatalf("writing %s: %v", name, err)
	}
}

// ---- record layouts ----

func (n *tNPC) Fields(c *gamefs.Codec) {
	for _, a := range []*[32]uint8{&n._gtile, &n._x, &n._y, &n._tile, &n._old_x, &n._old_y, &n._var, &n._tlkidx} {
		c.Bytes(a[:])
	}
}

func (t *t_500) Fields(c *gamefs.Codec) {
	c.Bytes(t._map[:])
	t._npc.Fields(c)
}

func (h *tChara) Fields(c *gamefs.Codec) {
	c.U16s(h._HP[:])
	for _, p := range []*uint16{&h._XP, &h._str, &h._dex, &h._int, &h._MP} {
		c.U16(p)
	}
	c.Bytes(h.__0e[:])
	c.U16(&h._weapon)
	c.U16(&h._armor)
	c.Bytes(h._name[:])
	c.U8(&h.p_24)
	c.U8(&h._class)
	c.U8(&h._stat)
}

func (p *tParty) Fields(c *gamefs.Codec) {
	c.U32(&p.f_000)
	c.U32(&p._moves)
	for i := range p.chara {
		p.chara[i].Fields(c)
	}
	c.I32(&p._food)
	for _, f := range []*uint16{&p._gold, &p._hones, &p._compa, &p._valor, &p._justi, &p._sacri,
		&p._honor, &p._spiri, &p._humil, &p._torches, &p._gems, &p._keys, &p._sextants} {
		c.U16(f)
	}
	c.U16s(p._armors[:])
	c.U16s(p._weapons[:])
	c.U16s(p._reagents[:])
	c.U16s(p._mixtures[:])
	c.U16(&p.mItems)
	for _, f := range []*uint8{&p._x, &p._y, &p.mStones, &p.mRunes} {
		c.U8(f)
	}
	for _, f := range []*uint16{&p.f_1d8, &p._tile, &p.f_1dc, &p._trammel, &p._felucca, &p._ship,
		&p.f_1e4, &p.f_1e6, &p.f_1e8, &p.f_1ea, &p.f_1ec} {
		c.U16(f)
	}
	c.U8(&p.out_x)
	c.U8(&p.out_y)
	c.U16(&p._dir)
	c.I16(&p._z)
	c.U16(&p._loc)
}

func (m *tCombat) Fields(c *gamefs.Codec) {
	for _, a := range [][]uint8{m._npcX[:], m._npcY[:], m._charaX[:], m._charaY[:], m._030[:], m._map[:], m._0b9[:]} {
		c.Bytes(a)
	}
}
