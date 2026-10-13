package u4

// Port of U4.H, U4_SCAN.H, U4_LOC.H and U4_VAR.C: the game's types and its uninitialised
// globals.
//
// Types: C int / unsigned are 16 bits, so globals keep int16 / uint16 (wrapping and
// signedness as in the C). A C "char" is stored as a byte; MSC's /J made plain char unsigned
// in some modules (u4_use, u4_file, u4_main, u4_main2, u4_map, u4_end, u4_anim, u4_comba,
// u4_spell, u4_q_n_v, u4_combb) and left it signed in the others, so each ported module reads
// a char global the way its own C did: signedByte() where the C sign-extended.
//
// Records are fixed-size Go structs; the ones in files encode and decode field by field
// (file.go), the C layout byte for byte (save files are raw dumps of them).

import (
	"fmt"

	"github.com/memmaker/u4-go/internal/gamefs"
)

// signedByte reads a C char the way a signed-char module did.
func signedByte(b uint8) int { return int(int8(b)) }

const (
	dirWest  = 0
	dirNorth = 1
	dirEast  = 2
	dirSouth = 3
)

// constants for "struct partyRecord.mItems" masking
const (
	itemBellUsed       = 12
	itemBookUsed       = 11
	itemCandleUsed     = 10
	itemWheel          = 9
	itemHorn           = 8
	itemKeyTruth       = 7
	itemKeyLove        = 6
	itemKeyCourage     = 5
	itemBell           = 4
	itemBook           = 3
	itemCandle         = 2
	itemSkullDestroyed = 1
	itemSkull          = 0
)

// The weapons (character.weapon, partyRecord.weapons).
const (
	weaponHands = iota
	weaponStaff
	weaponDagger
	weaponSling
	weaponMace
	weaponAxe
	weaponSword
	weaponBow
	weaponCrossbow
	weaponOil
	weaponHalberd
	weaponMagicAxe
	weaponMagicSword
	weaponMagicBow
	weaponMagicWand
	weaponMysticSword
)

// The armours (character.armor, partyRecord.armors).
const (
	armorSkin = iota
	armorCloth
	armorLeather
	armorChainMail
	armorPlateMail
	armorMagicChain
	armorMagicPlate
	armorMysticRobe
)

// The virtues, in the order of their stones, runes and karma.
const (
	virtueHonesty = iota
	virtueCompassion
	virtueValor
	virtueJustice
	virtueSacrifice
	virtueHonor
	virtueSpirituality
	virtueHumility
)

// The spell (or item) in effect, shown by its letter on the status line.
const (
	effectNone       = 7 // shows as nothing
	effectHorn       = 1 // the Silver Horn keeps the daemons away
	effectJinx       = 'J'
	effectProtection = 'P'
	effectNegate     = 'N'
	effectQuickness  = 'Q'
)

// What a dungeon room holds: its kind in the upper four bits, in the lower four a monster
// (its number plus one), or a variant of the kind.
const (
	dngCorridor     = 0
	dngLadderUp     = 16
	dngLadderDown   = 32
	dngLadderUpDown = 48
	dngChest        = 64
	dngCeilingHole  = 80
	dngFloorHole    = 96
	dngOrb          = 112
	dngTrap         = 128
	dngFountain     = 144
	dngField        = 160
	dngAltar        = 176
	dngDoor         = 192
	dngRoom         = 208
	dngSecretDoor   = 224
	dngWall         = 240
)

// dngKind is the kind of what a dungeon room holds; dngDetail its lower four bits.
func dngKind(room uint8) uint8   { return room &^ 15 }
func dngDetail(room uint8) uint8 { return room % 16 }

// A dungeon is 8 levels of 8x8 rooms, each level wrapping around at its edges.
const (
	levelSize      = 8
	dungeonLevels  = 8
	dungeonMapSize = dungeonLevels * levelSize * levelSize
)

// aheadX, aheadY, behindX and behindY step from (x, y) on a dungeon level, facing dir.
func aheadX(x, dir int) int  { return (x + dirDX[dir]) & (levelSize - 1) }
func behindX(x, dir int) int { return (x - dirDX[dir]) & (levelSize - 1) }
func aheadY(y, dir int) int  { return (y + dirDY[dir]) & (levelSize - 1) }
func behindY(y, dir int) int { return (y - dirDY[dir]) & (levelSize - 1) }

// A key is a BIOS key word: the scan code in the high byte, the character in the low.
func keyChar(key int) int { return key & 255 }

// lastPrintable is the last printable ASCII character, '~'.
const lastPrintable = 126

func toUpper(c int) int {
	if c >= 'a' && c <= 'z' {
		return c - 'a' + 'A'
	}
	return c
}

func toLower(c int) int {
	if c >= 'A' && c <= 'Z' {
		return c - 'A' + 'a'
	}
	return c
}

// lowerKey is key with its character lowered.
func lowerKey(key int) int { return key - keyChar(key) + toLower(keyChar(key)) }

// WORLD.MAP is the world, 256x256 tiles, as 8x8 chunks of 32x32 tiles, a chunk after another.
const (
	chunkTiles  = 32
	chunkSize   = chunkTiles * chunkTiles
	worldChunks = 64
)

type tMap32x32 [chunkTiles][chunkTiles]uint8
type tMap8x8x8 [8][8][8]uint8

// peopleTable is the people and monsters about: 32 slots, the world map using the first 8.
type peopleTable struct { // 256 bytes
	baseTile  [32]uint8
	x, y      [32]uint8
	tile      [32]uint8
	oldX      [32]uint8
	oldY      [32]uint8
	behaviour [32]uint8 // agressivity (or _z in dungeon)
	talkIndex [32]uint8
}

// areaRecord is *.ULT / *.DNG: a map (32x32 town or 8x8x8 dungeon, a union) and the NPCs.
type areaRecord struct {
	tiles  [chunkSize]uint8
	people peopleTable
}

func (t *areaRecord) townTile(y, x int) *uint8 { return &t.tiles[y*townTiles+x] }
func (t *areaRecord) dungeonRoom(z, y, x int) *uint8 {
	return &t.tiles[(z*levelSize+y)*levelSize+x]
}

// character is a party member, as PARTY.SAV holds them.
type character struct { // 39 bytes
	hp, maxHP    uint16
	xp           uint16
	strength     uint16
	dexterity    uint16
	intelligence uint16
	mp           uint16
	unused       [2]uint8
	weapon       uint16
	armor        uint16
	name         [16]uint8
	sex          uint8 // sex char
	class        uint8
	status       uint8
}

// partyRecord is the party: PARTY.SAV.
type partyRecord struct { // 502 bytes
	counter uint32 // a counter
	moves   uint32
	members [8]character
	food    int32
	gold    uint16
	// karmas
	karmaHonesty, karmaCompassion, karmaValor, karmaJustice, karmaSacrifice, karmaHonor, karmaSpirituality, karmaHumility uint16

	torches        uint16
	gems           uint16
	keys           uint16
	sextants       uint16
	armors         [8]uint16
	weapons        [16]uint16
	reagents       [8]uint16
	mixtures       [26]uint16
	items          uint16
	x, y           uint8
	stones         uint8
	runes          uint8
	size           uint16 // characters #
	tile           uint16
	flyingOrLight  uint16 // isFlying or light[dungeon]
	trammel        uint16 // moons phase
	felucca        uint16
	hull           uint16 // hull integrity
	metLordBritish uint16 // did met with LB
	lastCamp       uint16 // last hole up&camp
	lastFound      uint16 // last found
	lastMeditation uint16 // last meditation/Hawkwind
	lastKarmaTalk  uint16 // last karma-conversation
	outX           uint8
	outY           uint8
	dir            uint16 // [dungeon]
	z              int16  // [dungeon]
	loc            uint16
}

// bitMask is a byte or word of flags.
type bitMask interface{ ~uint8 | ~uint16 }

func hasBit[T bitMask](mask T, bit int) bool { return mask&(1<<bit) != 0 }
func setBit[T bitMask](mask *T, bit int)     { *mask |= 1 << bit }
func clearBit[T bitMask](mask *T, bit int)   { *mask &^= 1 << bit }

// battlefieldRecord is a CON file.
type battlefieldRecord struct { // 192 bytes
	monsterX, monsterY [16]uint8
	memberX, memberY   [8]uint8
	unused1            [16]uint8 // padding?
	tiles              [viewSize * viewSize]uint8
	unused2            [7]uint8 // padding?
}

// combatTile is the battlefield's tile at (x, y).
func combatTile(y, x int) *uint8 { return &battlefield.tiles[y*viewSize+x] }

type fighterTable struct {
	x, y       [16]uint8
	hp         [16]uint8
	tile       [16]uint8
	baseTile   [16]uint8
	sleeping   [16]uint8
	memberTile [8]uint8
}

type spot struct{ x, y uint8 }
type spot3 struct{ x, y, z uint8 }

func init() {
	for _, c := range []struct {
		v    gamefs.Record
		size int
	}{{&peopleTable{}, 256}, {&areaRecord{}, 1280}, {&character{}, 39}, {&partyRecord{}, 502}, {&battlefieldRecord{}, 192}} {
		if n := len(gamefs.Encode(c.v)); n != c.size {
			panic(fmt.Sprintf("%T is %d bytes, want %d", c.v, n, c.size))
		}
	}
}

// Game modes (gameMode).
const (
	modeVision     = 0 // the screen shows a picture, not the map
	modeOutdoors   = 1
	modeTown       = 2 // towns and castles
	modeDungeon    = 3
	modeCombat     = 4
	modeCombatCamp = 5 // attacked while camping
	modeCombatRoom = 6 // a dungeon room
	modeShrine     = 7
)

// Locations: 0 is the world map, then the towns and castles, the dungeons and the shrines.
const (
	locFirstTown          = 1
	locFirstCastle        = 1
	locCastleBritannia    = 1
	locLycaeum            = 2
	locEmpathAbbey        = 3
	locSerpentsHold       = 4
	locFirstTowne         = 5
	locMoonglow           = 5
	locBritain            = 6
	locJhelom             = 7
	locYew                = 8
	locMinoc              = 9
	locTrinsic            = 10
	locSkaraBrae          = 11
	locMagincia           = 12
	locPaws               = 13
	locBuccaneersDen      = 14
	locVesper             = 15
	locCove               = 16
	locFirstDungeon       = 17
	locDeceit             = 17
	locDespise            = 18
	locDestard            = 19
	locWrong              = 20
	locCovetous           = 21
	locShame              = 22
	locHythloth           = 23
	locAbyss              = 24
	locFirstShrine        = 25
	locShrineHonesty      = 25
	locShrineCompassion   = 26
	locShrineValor        = 27
	locShrineJustice      = 28
	locShrineSacrifice    = 29
	locShrineHonor        = 30
	locShrineSpirituality = 31
	locShrineHumility     = 32
)

// BIOS key words: the key's scan code << 8 | its character.
const (
	keyEsc       = 1<<8 | 27
	keyBackspace = 14<<8 | '\b'
	keyDelete    = 14<<8 | 127
	keySpace     = 57<<8 | ' '
	keyEnter     = 28<<8 | '\r'
	keyHome      = 71<<8 | 0
	keyUp        = 72<<8 | 0
	keyUpAlt     = 72<<8 | '~'
	keyPageUp    = 73<<8 | 0
	keyLeft      = 75<<8 | 0
	keyLeftAlt   = 75<<8 | '|'
	keyRight     = 77<<8 | 0
	keyRightAlt  = 244<<8 | 0
	keyEnd       = 79<<8 | 0
	keyDown      = 80<<8 | 0
	keyDownAlt   = 80<<8 | '`'
	keyPageDown  = 81<<8 | 0
	key1         = 2<<8 | '1'
	key2         = 3<<8 | '2'
	key3         = 4<<8 | '3'
	key4         = 5<<8 | '4'
	key5         = 6<<8 | '5'
	key6         = 7<<8 | '6'
	key7         = 8<<8 | '7'
	key8         = 9<<8 | '8'
	key0         = 11<<8 | '0'
	keyA         = 30<<8 | 'a'
	keyB         = 48<<8 | 'b'
	keyC         = 46<<8 | 'c'
	keyD         = 32<<8 | 'd'
	keyE         = 18<<8 | 'e'
	keyF         = 33<<8 | 'f'
	keyG         = 34<<8 | 'g'
	keyH         = 35<<8 | 'h'
	keyI         = 23<<8 | 'i'
	keyJ         = 36<<8 | 'j'
	keyK         = 37<<8 | 'k'
	keyL         = 38<<8 | 'l'
	keyM         = 50<<8 | 'm'
	keyN         = 49<<8 | 'n'
	keyO         = 24<<8 | 'o'
	keyP         = 25<<8 | 'p'
	keyQ         = 16<<8 | 'q'
	keyR         = 19<<8 | 'r'
	keyS         = 31<<8 | 's'
	keyT         = 20<<8 | 't'
	keyU         = 22<<8 | 'u'
	keyV         = 47<<8 | 'v'
	keyW         = 17<<8 | 'w'
	keyX         = 45<<8 | 'x'
	keyY         = 21<<8 | 'y'
	keyZ         = 44<<8 | 'z'
	keyCtrlS     = 31<<8 | 19
	keyAltZ      = 44<<8 | 0
)

// The game's state.
var (
	party            partyRecord
	area             areaRecord // the current town or dungeon map and its people
	battlefield      battlefieldRecord
	fighters         fighterTable
	gameMode         uint16
	modeBeforeCombat int16

	worldMap     []byte           // WORLD.MAP
	loadedChunks [8][8]*tMap32x32 // the world map's loaded 32x32 chunks
	dungeonFile  []byte           // the current dungeon's file (nil: none)
	talkFile     []byte           // the current town's .TLK file (nil: none)

	// The map window and the view.
	windowPos     spot
	windowQuarter spot3
	viewTiles     [viewSize * viewSize]uint8
	lineOfSight   uint16
	forceRender   uint16
	windDir       uint16
	gallopStep    uint16

	// The tiles around the party.
	hereTile, northTile, southTile, eastTile, westTile uint8
	aheadTile, behindTile                              uint8

	// The moongate on the map.
	moongateTile             uint8
	moongateAtX, moongateAtY uint8

	// The ship the pirates sail toward.
	shipX, shipY uint8

	// The coming or running fight.
	foeX, foeY, foeGround  uint8
	foeTile, encounterTile uint8
	activeMember           int16
	memberFrame            [8]uint8
	hitTile                uint16
	hitX, hitY             uint8
	altarRoom              int16
	roomTriggers           [16]uint8
	fleeDirX, fleeDirY     uint8 // the way out of a dungeon room

	// A lasting spell (or the horn), and its turns left.
	activeSpell uint8
	spellTurns  uint16

	// The text cursor.
	textCol, textRow int16
	cursorBoxY       uint16
	cursorBlink      int16

	speedInfo uint16     // the machine's speed, which the delays count in
	soundOn   uint8  = 1 // sound effects on (the Volume command)
)

// A party member's status, the letter shown after their hit points.
const (
	statusGood     = 'G'
	statusPoisoned = 'P'
	statusAsleep   = 'S'
	statusDead     = 'D'
)

// The classes, in the order of their tiles (from tileMage on, two tiles each).
const (
	classMage = iota
	classBard
	classFighter
	classDruid
	classTinker
	classPaladin
	classRanger
	classShepherd
)
