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
func driverExit(code int)            { fatalf("EGA driver error %d", code) }
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

func (n *peopleTable) Fields(c *gamefs.Codec) {
	for _, a := range []*[32]uint8{&n.baseTile, &n.x, &n.y, &n.tile, &n.oldX, &n.oldY, &n.behaviour, &n.talkIndex} {
		c.Bytes(a[:])
	}
}

func (t *areaRecord) Fields(c *gamefs.Codec) {
	c.Bytes(t.tiles[:])
	t.people.Fields(c)
}

func (h *character) Fields(c *gamefs.Codec) {
	for _, p := range []*uint16{&h.hp, &h.maxHP, &h.xp, &h.strength, &h.dexterity, &h.intelligence, &h.mp} {
		c.U16(p)
	}
	c.Bytes(h.unused[:])
	c.U16(&h.weapon)
	c.U16(&h.armor)
	c.Bytes(h.name[:])
	c.U8(&h.sex)
	c.U8(&h.class)
	c.U8(&h.status)
}

func (p *partyRecord) Fields(c *gamefs.Codec) {
	c.U32(&p.counter)
	c.U32(&p.moves)
	for i := range p.members {
		p.members[i].Fields(c)
	}
	c.I32(&p.food)
	for _, f := range []*uint16{&p.gold, &p.karmaHonesty, &p.karmaCompassion, &p.karmaValor, &p.karmaJustice, &p.karmaSacrifice,
		&p.karmaHonor, &p.karmaSpirituality, &p.karmaHumility, &p.torches, &p.gems, &p.keys, &p.sextants} {
		c.U16(f)
	}
	c.U16s(p.armors[:])
	c.U16s(p.weapons[:])
	c.U16s(p.reagents[:])
	c.U16s(p.mixtures[:])
	c.U16(&p.items)
	for _, f := range []*uint8{&p.x, &p.y, &p.stones, &p.runes} {
		c.U8(f)
	}
	for _, f := range []*uint16{&p.size, &p.tile, &p.flyingOrLight, &p.trammel, &p.felucca, &p.hull,
		&p.metLordBritish, &p.lastCamp, &p.lastFound, &p.lastMeditation, &p.lastKarmaTalk} {
		c.U16(f)
	}
	c.U8(&p.outX)
	c.U8(&p.outY)
	c.U16(&p.dir)
	c.I16(&p.z)
	c.U16(&p.loc)
}

func (m *battlefieldRecord) Fields(c *gamefs.Codec) {
	for _, a := range [][]uint8{m.monsterX[:], m.monsterY[:], m.memberX[:], m.memberY[:], m.unused1[:], m.tiles[:], m.unused2[:]} {
		c.Bytes(a)
	}
}
