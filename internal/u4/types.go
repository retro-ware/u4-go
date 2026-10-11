package u4

// Port of U4.H, U4_SCAN.H, U4_LOC.H and U4_VAR.C: the game's types and its uninitialised
// globals.
//
// Types: C int / unsigned are 16 bits, so globals keep int16 / uint16 (wrapping and
// signedness as in the C). A C "char" is stored as a byte; MSC's /J made plain char unsigned
// in some modules (u4_use, u4_file, u4_main, u4_main2, u4_map, u4_end, u4_anim, u4_comba,
// u4_spell, u4_q_n_v, u4_combb) and left it signed in the others, so each ported module reads
// a char global the way its own C did: sc() where the C sign-extended.
//
// Records are fixed-size Go structs whose fields are only (arrays of) uint8/uint16/uint32,
// read and written little-endian with encoding/binary, so they are the C layout byte for byte
// (save files are raw dumps of them).

import (
	"encoding/binary"
	"fmt"
)

// sc reads a C char the way a signed-char module did.
func sc(b uint8) int { return int(int8(b)) }

const (
	DIR_W = 0
	DIR_N = 1
	DIR_E = 2
	DIR_S = 3
)

// constants for "struct tParty.mItems" masking
const (
	ST_USE_BELL   = 12
	ST_USE_BOOK   = 11
	ST_USE_CANDLE = 10
	ST_WHEEL      = 9
	ST_HORN       = 8
	ST_KEY_T      = 7
	ST_KEY_L      = 6
	ST_KEY_C      = 5
	ST_BELL       = 4
	ST_BOOK       = 3
	ST_CANDLE     = 2
	ST_CAST_SKULL = 1
	ST_SKULL      = 0
)

func DNG_X_p(x, dir int) int { return (x + sc(D_080C[dir])) & 7 }
func DNG_X_m(x, dir int) int { return (x - sc(D_080C[dir])) & 7 }
func DNG_Y_p(y, dir int) int { return (y + sc(D_0810[dir])) & 7 }
func DNG_Y_m(y, dir int) int { return (y - sc(D_0810[dir])) & 7 }

func u4_lower(c int) int     { return c + 'a' - 'A' }
func u4_isupper(c int) bool  { return c >= 'A' && c <= 'Z' }
func u4_islower(c int) bool  { return c >= 'a' && c <= 'z' }
func u4_toupper(c *int) {
	if u4_islower(*c) {
		*c -= 'a' - 'A'
	}
}
func u4_tolower(c *int) {
	if u4_isupper(*c) {
		*c += 'a' - 'A'
	}
}

func u4_max(a, b int) int {
	if a >= b {
		return a
	}
	return b
}
func u4_min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

type tMap32x32 [32][32]uint8
type tMap8x8x8 [8][8][8]uint8

type tNPC struct { // size 0x100
	_gtile  [32]uint8
	_x, _y  [32]uint8
	_tile   [32]uint8
	_old_x  [32]uint8
	_old_y  [32]uint8
	_var    [32]uint8 // agressivity (or _z in dungeon)
	_tlkidx [32]uint8
}

// t_500 is *.ULT / *.DNG: a map (32x32 town or 8x8x8 dungeon, a union) and the NPCs.
type t_500 struct {
	_map [0x400]uint8
	_npc tNPC
}

func (t *t_500) x32x32(y, x int) *uint8   { return &t._map[y*32+x] }
func (t *t_500) x8x8x8(z, y, x int) *uint8 { return &t._map[z*64+y*8+x] }

type tChara struct { // size 0x27
	_HP     [2]uint16
	_XP     uint16
	_str    uint16
	_dex    uint16
	_int    uint16
	_MP     uint16
	__0e    [2]uint8
	_weapon uint16
	_armor  uint16
	_name   [16]uint8
	p_24    uint8 // sex char
	_class  uint8
	_stat   uint8
}

type tParty struct { // size 0x1f6
	f_000  uint32 // a counter
	_moves uint32
	chara  [8]tChara
	_food  int32
	_gold  uint16
	// karmas
	_hones, _compa, _valor, _justi, _sacri, _honor, _spiri, _humil uint16

	_torches  uint16
	_gems     uint16
	_keys     uint16
	_sextants uint16
	_armors   [8]uint16
	_weapons  [16]uint16
	_reagents [8]uint16
	_mixtures [26]uint16
	mItems    uint16
	_x, _y    uint8
	mStones   uint8
	mRunes    uint8
	f_1d8     uint16 // characters #
	_tile     uint16
	f_1dc     uint16 // isFlying or light[dungeon]
	_trammel  uint16 // moons phase
	_felucca  uint16
	_ship     uint16 // hull integrity
	f_1e4     uint16 // did met with LB
	f_1e6     uint16 // last hole up&camp
	f_1e8     uint16 // last found
	f_1ea     uint16 // last meditation/Hawkwind
	f_1ec     uint16 // last karma-conversation
	out_x     uint8
	out_y     uint8
	_dir      uint16 // [dungeon]
	_z        int16  // [dungeon]
	_loc      uint16
}

func TST_MSK(mask uint16, bit int) bool { return mask&(1<<bit) != 0 }
func SET_MSK(mask *uint16, bit int)     { *mask |= 1 << bit }
func RST_MSK(mask *uint16, bit int)     { *mask &^= 1 << bit }
func TST_MSK8(mask uint8, bit int) bool { return mask&(1<<bit) != 0 }

// tCombat is a CON file.
type tCombat struct { // size 0xc0
	_npcX, _npcY     [16]uint8
	_charaX, _charaY [8]uint8
	_030             [16]uint8 // padding?
	_map             [11 * 11]uint8
	_0b9             [7]uint8 // padding?
}

func Combat_MAP(a, b int) *uint8 { return &Combat._map[a*11+b] }

type t_68 struct {
	_x, _y    [16]uint8
	_HP       [16]uint8
	_tile     [16]uint8
	_gtile    [16]uint8
	_sleeping [16]uint8
	_chtile   [8]uint8
}

type aStruct struct{ x, y uint8 }
type aStruct3 struct{ x, y, z uint8 }

func init() {
	for _, c := range []struct {
		v    any
		size int
	}{{tNPC{}, 0x100}, {t_500{}, 0x500}, {tChara{}, 0x27}, {tParty{}, 0x1f6}, {tCombat{}, 0xc0}, {t_68{}, 0x68}} {
		if n := binary.Size(c.v); n != c.size {
			panic(fmt.Sprintf("%T is %#x bytes, want %#x", c.v, n, c.size))
		}
	}
}

// Game modes (CurMode).
const (
	MOD_VISION   = 0
	MOD_OUTDOORS = 1
	MOD_BUILDING = 2
	MOD_DUNGEON  = 3
	MOD_COMBAT   = 4
	MOD_COM_CAMP = 5
	MOD_COM_ROOM = 6
	MOD_SHRINE   = 7
)

// Locations (U4_LOC.H).
const (
	LOC_BUILDINGS    = 0x01
	LOC_CASTLES      = 0x01
	LOC_LCB          = 0x01
	LOC_LYCAEUM      = 0x02
	LOC_EMPATH       = 0x03
	LOC_SERPENT      = 0x04
	LOC_TOWNS        = 0x05
	LOC_MOONGLOW     = 0x05
	LOC_BRITAIN      = 0x06
	LOC_JHELOM       = 0x07
	LOC_YEW          = 0x08
	LOC_MINOC        = 0x09
	LOC_TRINSIC      = 0x0a
	LOC_SKARA        = 0x0b
	LOC_MAGINCIA     = 0x0c
	LOC_PAWS         = 0x0d
	LOC_DEN          = 0x0e
	LOC_VESPER       = 0x0f
	LOC_COVE         = 0x10
	LOC_DUNGEONS     = 0x11
	LOC_DECEIT       = 0x11
	LOC_DESPISE      = 0x12
	LOC_DESTARD      = 0x13
	LOC_WRONG        = 0x14
	LOC_COVETOUS     = 0x15
	LOC_SHAME        = 0x16
	LOC_HYTHLOTH     = 0x17
	LOC_ABYSS        = 0x18
	LOC_SHRINES      = 0x19
	LOC_HONESTY      = 0x19
	LOC_COMPASSION   = 0x1a
	LOC_VALOR        = 0x1b
	LOC_JUSTICE      = 0x1c
	LOC_SACRIFICE    = 0x1d
	LOC_HONOR        = 0x1e
	LOC_SPIRITUALITY = 0x1f
	LOC_HUMILITY     = 0x20
)

// BIOS key words (U4_SCAN.H): scan code << 8 | ASCII.
const (
	KBD_ESC    = 0x011b
	KBD_BS     = 0x0e08
	KBD_0e7f   = 0x0e7f
	KBD_SPACE  = 0x3920
	KBD_ENTER  = 0x1c0d
	KBD_HOME   = 0x4700
	KBD_UP     = 0x4800
	KBD_UP_    = 0x487e
	KBD_PGUP   = 0x4900
	KBD_LEFT   = 0x4B00
	KBD_LEFT_  = 0x4b7c
	KBD_RIGHT  = 0x4D00
	KBD_RIGHT_ = 0xf400
	KBD_END    = 0x4f00
	KBD_DOWN   = 0x5000
	KBD_DOWN_  = 0x5060
	KBD_PGDN   = 0x5100
	KBD_1      = 0x0231
	KBD_2      = 0x0332
	KBD_3      = 0x0433
	KBD_4      = 0x0534
	KBD_5      = 0x0635
	KBD_6      = 0x0736
	KBD_7      = 0x0837
	KBD_8      = 0x0938
	KBD_0      = 0x0b30
	KBD_A      = 0x1E61
	KBD_B      = 0x3062
	KBD_C      = 0x2E63
	KBD_D      = 0x2064
	KBD_E      = 0x1265
	KBD_F      = 0x2166
	KBD_G      = 0x2267
	KBD_H      = 0x2368
	KBD_I      = 0x1769
	KBD_J      = 0x246A
	KBD_K      = 0x256B
	KBD_L      = 0x266C
	KBD_M      = 0x326D
	KBD_N      = 0x316E
	KBD_O      = 0x186F
	KBD_P      = 0x1970
	KBD_Q      = 0x1071
	KBD_R      = 0x1372
	KBD_S      = 0x1F73
	KBD_T      = 0x1474
	KBD_U      = 0x1675
	KBD_V      = 0x2F76
	KBD_W      = 0x1177
	KBD_X      = 0x2D78
	KBD_Y      = 0x1579
	KBD_Z      = 0x2c7a
	KBD_CTRL_S = 0x1f13
	KBD_ALT_Z  = 0x2c00
)

// ---- uninitialised globals (U4_VAR.C, the c_common communals) ----

var (
	D_9140       uint8
	D_9141       uint8
	D_9142       int16
	Party        tParty
	D_933A       [8][8]*tMap32x32 // the world map's loaded 32x32 chunks
	D_943A       int16            // graphic adaptater type
	tile_forward uint8
	D_943E       int16
	D_9440       uint16
	ship_x       uint8
	ship_y       uint8
	tile_cur     uint8
	D_9445       uint8 // char
	D_9446       uint16
	D_9448       uint8 // char
	D_944A       [8]uint8
	D_9452       uint8
	pShapes      []byte
	CurMode      uint16
	D_946C       uint8
	spell_cnt    uint16
	Combat       tCombat
	Fighters     t_68
	File_DNG     int16
	activeChara  int16
	D_959C       aStruct
	tile_east    uint8
	pCharset     []byte
	spell_sta    uint8 // char
	D_95A5       aStruct3
	txt_X        int16
	tile_north   uint8 // char
	activeCharaY uint16
	hit_tile     uint16
	hit_x, hit_y uint8
	D_95B2       [16]uint8
	tile_west    uint8
	tile_south   uint8
	txt_Y        int16
	D_95C6       uint16
	D_95C8       int16
	tile_back    uint8
	D_95CC       int16
	D_95CE       [0x120]uint8 // current .tlk data
	D_96EE       uint8        // char: room exit x[dungeon]
	File_MAP     int16
	WindDir      uint16
	D_96F4       uint8 // char: room exit y[dungeon]
	File_TLK     int16
	D_96F8       uint8
	D_96F9       [11 * 11]uint8
	D_9772       uint8

	D_8742 t_500 // UTIL2.ASM's BSS: the current map and its NPCs
)

// ---- LOW.ASM ----
var (
	speed_info  uint16
	equip_flags uint16
)

// ---- U4_SOUND.ASM ----
var SoundFlag uint8 = 1
