package u4

// Ports of U4_SRCH.C (searching the world and towns) and U4_SRCH2.C (dungeons).

const D_27A6 = "Nothing Here!\n"

// "You find..." & karma+=5
func C_8D4B() {
	u4_puts("You find...\n")
	karma_inc(&Party._honor, 5)
	Party.f_1e8 = uint16(Party._moves & 0xf0)
}

// get random amount of reagent (max 99)
func C_8D6D(bp04 int) {
	Party._reagents[bp04] += uint16(U4_RND1(7) + 2)
	if Party._reagents[bp04] > 99 {
		Party._reagents[bp04] = 99
		u4_puts("Dropped some!\n")
		sound(1)
	}
}

func moonsOrFound() bool {
	return Party._trammel|Party._felucca != 0 || Party._moves&0xf0 == uint32(Party.f_1e8)
}

func C_8DAA() {
	if moonsOrFound() {
		u4_puts(D_27A6)
		return
	}
	C_8D4B()
	u4_puts("Mandrake Root!\n")
	C_8D6D(7)
}

func C_8DE0() {
	if moonsOrFound() {
		u4_puts(D_27A6)
		return
	}
	C_8D4B()
	u4_puts("Nightshade!\n")
	C_8D6D(6)
}

// findItem gives the item with bit st once.
func findItem(st int, text string) {
	if TST_MSK(Party.mItems, st) {
		u4_puts(D_27A6)
		return
	}
	SET_MSK(&Party.mItems, st)
	C_8D4B()
	u4_puts(text)
	XP_inc(0, 400)
}

func C_8E16() { findItem(ST_BELL, "The Bell of Courage!\n") }
func C_8E46() { findItem(ST_HORN, "A Silver Horn!\n") }
func C_8E77() { findItem(ST_WHEEL, "The Wheel from the H.M.S. Cape!\n") }

func C_8EA8() {
	if TST_MSK(Party.mItems, ST_SKULL) || Party._trammel|Party._felucca != 0 || TST_MSK(Party.mItems, ST_CAST_SKULL) {
		u4_puts(D_27A6)
		return
	}
	SET_MSK(&Party.mItems, ST_SKULL)
	C_8D4B()
	u4_puts("The Skull of Mondain the Wizard!\n")
	XP_inc(0, 400)
}

func C_8EE8() {
	if TST_MSK8(Party.mStones, 7) || Party._trammel|Party._felucca != 0 {
		u4_puts(D_27A6)
		return
	}
	Party.mStones |= 1 << 7
	C_8D4B()
	u4_puts("The Black Stone!\n")
	XP_inc(0, 200)
}

func C_8F21() {
	if TST_MSK8(Party.mStones, 6) {
		u4_puts(D_27A6)
		return
	}
	Party.mStones |= 1 << 6
	C_8D4B()
	u4_puts("The White Stone!\n")
	XP_inc(0, 200)
}

func C_8F51() { findItem(ST_BOOK, "The Book of Truth!\n") }
func C_8F81() { findItem(ST_CANDLE, "The Candle of Love!\n") }

// use telescope
func C_8FB1() {
	u4_puts("You see a knob on the Telescope marked A-P\n")
	bp_02 := AskLetter("You Select:\x12\x12\b", 'A', 'P')
	if bp_02 < 0 {
		return
	}
	bp_02 -= 'A'
	Load(D_0824[bp_02], D_8742._map[:])
	C_B9EF(0)
	Load(D_0824[Party._loc-1], D_8742._map[:])
}

func anyKarma() uint16 {
	return Party._hones | Party._compa | Party._valor | Party._justi |
		Party._sacri | Party._honor | Party._spiri | Party._humil
}

func C_9027() {
	if Party._armors[7]|anyKarma() != 0 {
		u4_puts(D_27A6)
		return
	}
	Party._armors[7] = 8
	C_8D4B()
	u4_puts("Mystic Armour!\n")
	XP_inc(0, 400)
}

func C_9076() {
	if Party._weapons[15]|anyKarma() != 0 {
		u4_puts(D_27A6)
		return
	}
	Party._weapons[15] = 8
	C_8D4B()
	u4_puts("Mystic Weapons!\n")
	XP_inc(0, 400)
}

var D_2904 = [8][2]uint8{
	{0x05, 0x01}, {0x06, 0x02}, {0x07, 0x04}, {0x08, 0x08},
	{0x09, 0x10}, {0x0A, 0x20}, {0x01, 0x40}, {0x0D, 0x80},
}

func C_90C5() {
	si := 7
	for ; si > 0; si-- {
		if uint16(D_2904[si][0]) == Party._loc {
			break
		}
	}
	if D_2904[si][1]&Party.mRunes != 0 {
		u4_puts(D_27A6)
		return
	}
	Party.mRunes |= D_2904[si][1]
	C_8D4B()
	u4_puts("The rune of ")
	u4_puts(D_1E98[151+si])
	u4_puts("!\n")
	XP_inc(0, 100)
}

type t_05_local struct {
	_00, _01, _02 uint8
	_03           func()
}

var D_2920 = []t_05_local{
	{0x00, 0xB6, 0x36, C_8DAA}, // Mandrake root
	{0x00, 0x64, 0xA5, C_8DAA}, // Mandrake root
	{0x00, 0x2E, 0x95, C_8DE0}, // Nightshade
	{0x00, 0xCD, 0x2C, C_8DE0}, // Nightshade
	{0x00, 0xB0, 0xD0, C_8E16}, // The Bell of Courage!
	{0x00, 0x2D, 0xAD, C_8E46}, // A Silver Horn!
	{0x00, 0x60, 0xD7, C_8E77}, // The Wheel from the H.M.S. Cape!
	{0x00, 0xC5, 0xF5, C_8EA8}, // The Skull of Mondain the Wizard!
	{0x00, 0xE0, 0x85, C_8EE8}, // The Black Stone!
	{0x00, 0x40, 0x50, C_8F21}, // The White Stone!
	{0x02, 0x06, 0x06, C_8F51}, // The Book of Truth!
	{0x10, 0x16, 0x01, C_8F81}, // The Candle of Love!
	{0x02, 0x16, 0x03, C_8FB1}, // use Telescope
	{0x03, 0x16, 0x04, C_9027}, // Mystic Armour!
	{0x04, 0x08, 0x0F, C_9076}, // Mystic Weapons!
	{0x05, 0x08, 0x06, C_90C5}, // rune of Honesty
	{0x06, 0x19, 0x01, C_90C5}, // rune of Compassion
	{0x07, 0x1E, 0x1E, C_90C5}, // rune of Valor
	{0x08, 0x0D, 0x06, C_90C5}, // rune of Justice
	{0x09, 0x1C, 0x1E, C_90C5}, // rune of Sacrifice
	{0x0A, 0x02, 0x1D, C_90C5}, // rune of Honor
	{0x01, 0x11, 0x08, C_90C5}, // rune of Spirituality
	{0x0D, 0x1D, 0x1D, C_90C5}, // rune of Humility
}

func CMD_Search() { /*C_913A*/
	u4_puts("Search...\n")
	if CurMode <= MOD_BUILDING && Party.f_1dc != 0 {
		w_DriftOnly()
		return
	}
	for _, si := range D_2920 {
		if uint16(si._00) == Party._loc && si._01 == Party._x && si._02 == Party._y {
			si._03()
			return
		}
	}
	u4_puts(D_27A6)
}

// ---- U4_SRCH2.C: MODULE SEARCH [dungeon] ----

var D_2E04 = [1]string{"\nHmmm--No Effect!\n"}

func C_B730(bp06, bp04 int) {
	Gra_11(bp06)
	sound(7)
	sound(7)
	sound(7)
	Gra_11(bp06)
	hitChara(bp06, 100*bp04)
}

// increment characteristic (max 50)
func C_B770(bp06 *uint16, bp04 int) {
	*bp06 += uint16(bp04)
	if int16(*bp06) > 50 {
		*bp06 = 50
	}
}

var D_2E6E = [8]int{2, 2, 2, 4, 4, 4, 6, 6} // magical ball hits points
var D_2E7E = [3][8]int{                     // magical ball characteristic
	{0, 0, 5, 0, 5, 5, 5, 0},
	{0, 5, 0, 5, 5, 0, 5, 0},
	{5, 0, 0, 5, 0, 5, 5, 0},
}

// magical ball
func C_B795() {
	u4_puts("\nYou find a Magical Ball...\n")
	bp_02 := AskChara("Who touches?\x12\x12\b")
	if bp_02 < 0 {
		return
	}
	if !isCharaConscious(bp_02) {
		u4_puts("\nDisabled!\n")
		return
	}
	*D_8742.x8x8x8(int(Party._z), int(Party._y), int(Party._x)) = 0
	l := int(Party._loc) - 0x11
	C_B730(bp_02, D_2E6E[l])
	c := &Party.chara[bp_02]
	if D_2E7E[0][l] != 0 {
		C_B770(&c._str, 5)
		u4_puts("Strength + 5\n")
	}
	if D_2E7E[1][l] != 0 {
		C_B770(&c._dex, 5)
		u4_puts("Dexterity + 5\n")
	}
	if D_2E7E[2][l] != 0 {
		C_B770(&c._int, 5)
		u4_puts("Intelligence + 5\n")
	}
}

// fountain
func C_B863() {
	u4_puts("You find a Fountain.\n")
	bp_02 := AskChara("Who drinks?\x12\x12\b")
	if bp_02 < 0 {
		return
	}
	if !isCharaConscious(bp_02) {
		u4_puts("\nDisabled!\n")
		return
	}
	c := &Party.chara[bp_02]
	switch tile_cur & 0xf {
	case 0:
		u4_puts(D_2E04[0])
	case 1:
		if c._HP[0] == c._HP[1] {
			u4_puts(D_2E04[0])
		} else {
			u4_puts("\nAhh-Refreshing!\n")
			c._HP[0] = c._HP[1]
		}
	case 2:
		u4_puts("\nBleck--Nasty!\n")
		C_B730(bp_02, 1)
	case 3:
		if c._stat != 'P' {
			u4_puts(D_2E04[0])
		} else {
			c._stat = 'G'
			u4_puts("\nHmm--Delicious!\n")
		}
	case 4:
		if c._stat == 'P' {
			u4_puts(D_2E04[0])
		} else {
			c._stat = 'P'
			u4_puts("\nArgh-Choke-Gasp!\n")
			C_B730(bp_02, 1)
		}
	}
}

// found stone
func C_B93F() {
	l := int(Party._loc) - 0x11
	if Party._loc == 0x17 || Party._loc == 0x18 || TST_MSK8(Party.mStones, l) {
		u4_puts("\nYou find Nothing!\n")
		return
	}
	Party.mStones |= 1 << l
	u4_puts("\nYou find the ")
	u4_puts(D_0884[l])
	u4_puts(" stone!\n")
	karma_inc(&Party._honor, 5)
	XP_inc(0, 200)
}

// (S)earch [dungeon]
func C_B9B2() {
	u4_puts("Search...\n")
	switch tile_cur & 0xf0 {
	case 0x70:
		C_B795()
	case 0x90:
		C_B863()
	case 0xB0:
		C_B93F()
	default:
		u4_puts("\nYou find Nothing!\n")
	}
}
