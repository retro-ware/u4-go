package replay

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
)

// saveMember and saveParty are PARTY.SAV's layout (little-endian, packed): the save format,
// not the code's names, so the scenarios survive the code's rewrite.
type saveMember struct {
	HP, MaxHP, XP      uint16
	Str, Dex, Int, MP  uint16
	Unused             uint16
	Weapon, Armor      uint16
	Name               [16]byte
	Sex, Class, Status byte
}

type saveParty struct {
	Counter, Moves                                                     uint32
	Members                                                            [8]saveMember
	Food                                                               int32
	Gold                                                               uint16
	Karma                                                              [8]uint16
	Torches, Gems, Keys, Sextants                                      uint16
	Armors                                                             [8]uint16
	Weapons                                                            [16]uint16
	Reagents                                                           [8]uint16
	Mixtures                                                           [26]uint16
	Items                                                              uint16
	X, Y                                                               byte
	Stones, Runes                                                      byte
	Size, Tile, Flying, Trammel, Felucca, Ship                         uint16
	MetLordBritish, LastCamp, LastFound, LastMeditation, LastKarmaTalk uint16
	OutX, OutY                                                         byte
	Dir                                                                uint16
	Z                                                                  int16
	Location                                                           uint16
}

// places are the world coordinates of the towns and castles (0..15), dungeons (16..23) and
// shrines (24..31), in the game's order.
var places = [32][2]byte{
	{86, 107}, {218, 107}, {28, 50}, {146, 241}, {232, 135}, {82, 106}, {36, 222}, {58, 43},
	{159, 20}, {106, 184}, {22, 128}, {187, 169}, {98, 145}, {136, 158}, {201, 59}, {136, 90},
	{240, 73}, {91, 67}, {72, 168}, {126, 20}, {156, 27}, {58, 102}, {239, 240}, {233, 233},
	{233, 66}, {128, 92}, {36, 229}, {73, 11}, {205, 45}, {81, 207}, {231, 216}, {231, 216},
}

// richParty writes a PARTY.SAV from PARTY.NEW's eight companions, all of them in the party
// and well supplied, standing outside at place p.
func richParty(game, save string, p int) error {
	b, err := os.ReadFile(filepath.Join(game, "PARTY.NEW"))
	if err != nil {
		return err
	}
	var party saveParty
	if err := binary.Read(bytes.NewReader(b), binary.LittleEndian, &party); err != nil {
		return err
	}
	const lots = 99
	for i := range party.Members {
		m := &party.Members[i]
		m.MaxHP, m.HP, m.XP = 400, 400, 800
		m.Str, m.Dex, m.Int, m.MP = 30, 30, 30, 50
		m.Status = 'G'
	}
	party.Size = 8
	party.Food, party.Gold = 9000, 9000
	party.Torches, party.Gems, party.Keys, party.Sextants = lots, lots, lots, 1
	for i := range party.Armors {
		party.Armors[i] = 8
	}
	for i := range party.Weapons {
		party.Weapons[i] = 8
	}
	for i := range party.Reagents {
		party.Reagents[i] = lots
	}
	for i := range party.Mixtures {
		party.Mixtures[i] = lots
	}
	for i := range party.Karma {
		party.Karma[i] = 50
	}
	party.X, party.Y = places[p][0], places[p][1]
	party.Location = 0
	var out bytes.Buffer
	binary.Write(&out, binary.LittleEndian, &party)
	return os.WriteFile(filepath.Join(save, "PARTY.SAV"), out.Bytes(), 0o644)
}

// townFiles are the towns' and castles' maps, by location number - 1.
var townFiles = [16]string{
	"LCB_1.ULT", "LYCAEUM.ULT", "EMPATH.ULT", "SERPENT.ULT",
	"MOONGLOW.ULT", "BRITAIN.ULT", "JHELOM.ULT", "YEW.ULT",
	"MINOC.ULT", "TRINSIC.ULT", "SKARA.ULT", "MAGINCIA.ULT",
	"PAWS.ULT", "DEN.ULT", "VESPER.ULT", "COVE.ULT",
}

// A town file is its 32x32 map, then the people's records, a field of 32 bytes after another.
const (
	townSize        = 32
	peopleX         = townSize*townSize + 32
	peopleY         = peopleX + 32
	peopleTile      = peopleY + 32
	merchantTile    = 82
	firstLetterTile = 96
	lastLetterTile  = 126
	floorTile       = 62
)

// counter is a place to stand at a shop: across a counter (a letter tile) from its merchant.
type counter struct {
	town   int // 0..15
	x, y   byte
	toward rune // the arrow key from the party to the merchant
}

// counters lists every shop counter of the towns, read from their maps.
func counters(game string) ([]counter, error) {
	var cs []counter
	arrows := map[[2]int]rune{{1, 0}: '→', {-1, 0}: '←', {0, 1}: '↓', {0, -1}: '↑'}
	for town, name := range townFiles {
		b, err := os.ReadFile(filepath.Join(game, name))
		if err != nil {
			return nil, err
		}
		at := func(x, y int) byte {
			if x < 0 || y < 0 || x >= townSize || y >= townSize {
				return 0
			}
			return b[y*townSize+x]
		}
		for i := 0; i < 32; i++ {
			if b[peopleTile+i] != merchantTile {
				continue
			}
			for d, arrow := range arrows {
				cx, cy := int(b[peopleX+i])-d[0], int(b[peopleY+i])-d[1]
				sx, sy := cx-d[0], cy-d[1]
				if c := at(cx, cy); c >= firstLetterTile && c <= lastLetterTile && at(sx, sy) == floorTile {
					cs = append(cs, counter{town, byte(sx), byte(sy), arrow})
				}
			}
		}
	}
	return cs, nil
}
