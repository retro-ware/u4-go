package title

// Port of VAR.C (the shared globals) and the record layouts of TITLE.H.

import "github.com/memmaker/u4-go/internal/gamefs"

// character is a party member as PARTY.NEW holds it (39 bytes): the fields the title sets are
// named, the rest is kept as read.
type character struct {
	hpAndXP                           [6]uint8 // hit points, their maximum, experience
	strength, dexterity, intelligence uint16
	magicAndGear                      [8]uint8 // magic points, padding, weapon, armour
	name                              [16]uint8
	sex                               uint8 // glyphMale or glyphFemale
	classAndStatus                    [2]uint8
}

// partyRecord is PARTY.NEW / PARTY.SAV (502 bytes).
type partyRecord struct {
	counter     uint32 // counts key polls
	moves       [4]uint8
	members     [8]character
	foodAndGold [6]uint8
	karma       [8]uint16 // honesty .. humility
	inventory   [126]uint8
	x, y        uint8
	stonesRunes [2]uint8
	size        uint16 // how many members
	rest        [28]uint8
}

const partyRecordSize = 502

func (m *character) Fields(c *gamefs.Codec) {
	c.Bytes(m.hpAndXP[:])
	c.U16(&m.strength)
	c.U16(&m.dexterity)
	c.U16(&m.intelligence)
	c.Bytes(m.magicAndGear[:])
	c.Bytes(m.name[:])
	c.U8(&m.sex)
	c.Bytes(m.classAndStatus[:])
}

func (p *partyRecord) Fields(c *gamefs.Codec) {
	c.U32(&p.counter)
	c.Bytes(p.moves[:])
	for i := range p.members {
		p.members[i].Fields(c)
	}
	c.Bytes(p.foodAndGold[:])
	c.U16s(p.karma[:])
	c.Bytes(p.inventory[:])
	c.U8(&p.x)
	c.U8(&p.y)
	c.Bytes(p.stonesRunes[:])
	c.U16(&p.size)
	c.Bytes(p.rest[:])
}

// peopleTable is the people of a map (MONSTERS.SAV, 256 bytes): eight arrays of 32, as in
// the game. The title's view uses it for its actors.
type peopleTable struct {
	baseTile  [32]uint8 // the tile drawn (tile plus animation frame)
	x, y      [32]uint8
	tile      [32]uint8
	oldX      [32]uint8
	oldY      [32]uint8
	behaviour [32]uint8
	talkIndex [32]uint8
}

func (t *peopleTable) Fields(c *gamefs.Codec) {
	for _, a := range []*[32]uint8{&t.baseTile, &t.x, &t.y, &t.tile, &t.oldX, &t.oldY, &t.behaviour, &t.talkIndex} {
		c.Bytes(a[:])
	}
}

var (
	monstersShown bool        // the two monsters animate beside the view
	party         partyRecord // the new party
	people        peopleTable // the view's actors, saved empty as MONSTERS.SAV
	shapes        []byte      // SHAPES.EGA, the tiles
	animPic       []byte      // ANIMATE.EGA, the monsters' frames
	charset       []byte      // CHARSET.EGA, the font
	titlePic      []byte      // TITLE.EGA
	textCol       int         // the text cursor
	textRow       int
)

func init() {
	if n := len(gamefs.Encode(&party)); n != partyRecordSize {
		panic("partyRecord layout")
	}
}
