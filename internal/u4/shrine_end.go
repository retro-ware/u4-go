package u4

// Ports of U4_SHRIN.C (meditating at the shrines) and U4_END.C (the Abyss chamber and the
// ending).

import (
	"github.com/memmaker/u4-go/internal/host"
	"github.com/memmaker/u4-go/internal/music"
)

// strnicmp compares the C string in a with b, ignoring case, for at most n characters.
func strnicmp(a []byte, b string, n int) int {
	lc := func(c byte) byte {
		if c >= 'A' && c <= 'Z' {
			return c + 'a' - 'A'
		}
		return c
	}
	at := func(i int) byte {
		if i < len(a) {
			return a[i]
		}
		return 0
	}
	for i := 0; i < n; i++ {
		x, y := lc(at(i)), byte(0)
		if i < len(b) {
			y = lc(b[i])
		}
		if x != y {
			return int(x) - int(y)
		}
		if x == 0 {
			return 0
		}
	}
	return 0
}

// ---- U4_SHRIN.C ----

// mantras
var D_8322 = [8]string{"ahm", "mu", "ra", "beh", "cah", "summ", "om", "lum"}

// visions
var D_8332 = [8][3]string{
	{
		"Take not the gold of others found in towns and castles for yours it is not!\n",
		"Cheat not the merchants and peddlers for tis an evil thing to do!\n",
		"Second, read the Book of Truth at the entrance to the Great Stygian Abyss!\n",
	},
	{
		"Kill not the non-evil beasts of the land, and do not attack the fair people!\n",
		"Give of thy purse to those who beg and thy deed shall not be forgotten!\n",
		"Third, light the Candle of Love at the entrance to the Great Stygian Abyss!\n",
	},
	{
		"Victories scored over evil creatures help to build a valorous soul!\n",
		"To flee from battle with less than grievous wounds often shows a coward!\n",
		"First, ring the Bell of Courage at the entrance to the Great Stygian Abyss!\n",
	},
	{
		"To take the gold of others is injustice not soon forgotten. Take only thy due!\n",
		"Attack not a peaceful citizen for that action deserves strict punishment!\n",
		"Kill not a non-evil beast for they deserve not death, even if in hunger they attack thee!",
	},
	{
		"To give thy last gold piece unto the needy shows good measure of self-sacrifice!\n",
		"For thee to flee and leave thy companions is a self-serving action to be avoided!\n",
		"To give of thy life's blood so that others may live is a virtue of great praise!\n",
	},
	{
		"Take not the gold of others for this shall bring dishonor upon thee!\n",
		"To strike first a non-evil being is by no means an honorable deed!\n",
		"Seek ye to solve the many Quests before thee, and honor shall be a reward!\n",
	},
	{
		"Seek ye to know thyself.  Visit the seer often for he can see into thy inner being!\n",
		"Meditation leads to enlightenment Seek ye all Wisdom and Knowledge!\n",
		"If thou dost seek the White Stone, search ye not under the ground, but in Serpent's Spine!\n",
	},
	{
		"Claim not to be that which thou art not.  Humble actions speak well of thee!\n",
		"Strive not to wield the Great Force of Evil for its power will overcome thee!\n",
		"If thou dost seek the Black Stone, search ye at the Time and Place of the Gate on the darkest of all nights!\n",
	},
}

// visions 2
var D_8362 = [8]string{"rune_1", "rune_2", "rune_0", "rune_1", "rune_2", "rune_1", "rune_3", "rune_4"} // I N F I N I T Y

// display avatarhood vision I.N.F.I.N.I.T.Y
func C_E6DF(bp04 int) {
	CurMode = MOD_VISION
	Gra_10()
	Gra_16(D_8362[bp04]+".ega", 0, nil)
}

// enter a shrine
func C_E72C() {
	var loc_D [16]byte
	loc_A := int(Party._loc) - 0x19 // shrine "number"
	leave := func() {
		CurMode = MOD_OUTDOORS
		Party._loc = 0
	}
	if !TST_MSK8(Party.mRunes, loc_A) {
		u4_puts("\nThou dost not bear the rune of entry!  A strange force keeps you out!\n")
		leave()
		return
	}
	Load("SHRINE.CON", Combat._map[:])
	CurMode = MOD_SHRINE
	music.Play(music.Shrine)
	u4_puts("\nYou enter the ancient shrine and sit before the altar...\nUpon what virtue dost thou meditate?\n")
	u4_gets(loc_D[:], 15)
	Gra_CR()
	u4_puts("\nFor how many\n")
	loc_E := AskLetter("Cycles (0-3)?\x12\x12\b", '0', '3')
	if loc_E < 0 {
		leave()
		return
	}
	loc_E -= '0'
	if loc_E == 0 || strnicmp(loc_D[:], D_1E98[151+loc_A], 16) != 0 {
		u4_puts("\nThou art unable to focus thy thoughts on this subject!\n")
		leave()
		return
	}
	if Party._moves/100 == uint32(Party.f_1ea) {
		u4_puts("\nThy mind is still weary from thy last Meditation!\n")
		leave()
		return
	}
	Party.f_1ea = uint16(Party._moves / 100)
	u4_puts("Begin Meditation\n")
	for loc_B := 0; loc_B < loc_E; loc_B++ {
		for loc_C := 0; loc_C < 0x10; loc_C++ {
			u_delay(1, 0)
			u4_putc('.')
		}
		u_kbflush()
		u4_puts("\nMantra: ")
		u4_gets(loc_D[:], 15)
		Gra_CR()
		if strnicmp(loc_D[:], D_8322[loc_A], 0x10) != 0 {
			u4_puts("\nThou art not able to focus thy thoughts with that Mantra!\n")
			karma_dec(&Party._spiri, 3)
			leave()
			return
		}
	}
	if loc_E != 3 || *pKarmas[loc_A] != 99 {
		u4_puts("\nThy thoughts are pure. Thou art granted a vision!\n")
		karma_inc(&Party._spiri, loc_E*3)
		u_kbflush()
		u_kbread()
		Gra_CR()
		u4_puts(D_8332[loc_A][loc_E-1])
	} else {
		u4_puts("\nThou hast achieved partial Avatarhood in the Virtue of\n")
		u4_puts(D_1E98[151+loc_A])
		Gra_09()
		sound(9, 0xff)
		Gra_09()
		Gra_CR()
		*pKarmas[loc_A] = 0
		dspl_Stats()
		u_kbflush()
		u_kbread()
		u4_puts("\n\nThou art granted a vision!\n")
		C_E6DF(loc_A)
		u_kbflush()
	}
	u_kbread()
	Gra_CR()
	leave()
}

// ---- U4_END.C ----

var D_0BDA = [11]string{
	"What dost thou possess if all may rely upon your every word?",
	"What quality compels one to share in the journeys of others?",
	"What answers when great deeds are called for?",
	"What should be the same for Lord and Serf alike?",
	"What is loath to place the self above aught else?",
	"What shirks no duty?",
	"What, in knowing the true self, knows all?",
	"What is that which Serfs are born with but Nobles must strive to obtain?",
	"If all else is imaginary, this is real...",
	"What plunges to the depths, while soaring on the heights?",
	"What turns not away from any peril?",
}

func C_2F7E() {
	u_delay(15, 1)
	if u_kbhit() != 0 {
		u_kbread()
	}
}

var D_0BF0 = [13]uint8{0xE7, 0x53, 0x23, 0x3B, 0x9E, 0x69, 0x17, 0xBA, 0xD8, 0x1D, 0x91, 0x59, 0xE9}
var D_0BFE = [13]uint8{0x88, 0x69, 0xDD, 0x2C, 0x15, 0xB7, 0x81, 0xAC, 0x6A, 0x30, 0xF3, 0x6A, 0xE9}

var D_8CCA uint16

// return to surface
func C_2F9D(bp04 int) {
	u_delay(5, 0)
	Gra_CR()
	Party._x = D_0BF0[bp04]
	Party._y = D_0BFE[bp04]
	Party.f_1d8 = D_8CCA
	Party._loc = 0
	CurMode = MOD_OUTDOORS
	spell_cnt = 0
	hit_tile = 0
	spell_sta = 7
	D_9440 = 1
	D_95C6 = 0
	Party.f_1dc = 0
	C_26B6()
	dspl_Stats()
	u_kbflush()
	longjmp_D_9458()
}

// passage is not granted
func C_3010() {
	u4_puts("\nPassage is not granted.\n")
	C_2F9D(12)
}

// Victory !
func C_3025() {
	u_delay(2, 0)
	shakefx()
	shakefx()
	shakefx()
	shakefx()
	u_delay(3, 0)
	music.Play(music.Fanfare)
	Gra_17() // infinity symbol
	u_delay(3, 0)
	for _, s := range []string{
		"\n\nThe boundless knowledge of the Codex of Ultimate Wisdom is revealed unto thee.",
		"\n\nThe voice says: Thou hast proven thyself to be truly good in nature.",
		"\n\nThou must know that thy quest to become an Avatar is the endless quest of a lifetime.",
		"\n\nAvatarhood is a living gift.  It must always and forever be nurtured to flourish.",
		"\n\nFor if thou dost stray from the paths of virtue, thy way may be lost forever.",
		"\n\nReturn now unto thine own world. Live there as an example to thy people, as our memory of thy gallant deeds serves us.",
	} {
		u4_puts(s)
		C_2F7E()
	}
	Gra_10()
	u4_puts("\n\nAs the sound of the voice trails off, darkness seems to rise around you. There is a moment of intense, wrenching vertigo.")
	C_2F7E()
	Gra_16("stoncrcl.ega", 0, nil)
	for _, s := range []string{
		"\n\nYou open your eyes to a familiar circle of stones.  You wonder of your recent adventures.",
		"\n\nIt seems a time and place very distant.  You wonder if it really happened. Then you realize that in your hand you hold The Ankh.",
		"\n\nYou walk away from the circle, knowing that you can always return from whence you came, since you now know the secret of the gates.",
	} {
		u4_puts(s)
		C_2F7E()
	}
	u4_puts("\n\nCONGRATULATIONS!\n   Thou hast\n   completed\n   ULTIMA IV\n  Quest of the\n    AVATAR\n  in ")
	u4_putl(int32(Party._moves), 0, '0')
	u4_puts("\n turns! Report\n thy feat unto\nLord British at\nOrigin Systems!")
	for { // while(1);
		host.Yield()
	}
}

func C_310F(bp06, bp04 string) bool {
	var bp_10 [16]byte
	for try := 0; try < 3; try++ {
		if try > 0 {
			u_delay(1, 0)
			u4_puts("\nThy thoughts are not pure.\nI ask again.\n")
			u_delay(2, 0)
		}
		u4_puts(bp06)
		Gra_CR()
		Gra_CR()
		u4_gets(bp_10[:], 15)
		Gra_CR()
		if strnicmp(bp_10[:], bp04, 15) == 0 {
			return true
		}
	}
	return false
}

var D_161A = [3]string{"truth", "love", "courage"}
var D_1620 = [8]string{"honesty", "compassn", "valor", "justice", "sacrific", "honor", "spirit", "humility"}

// last phase of game ?
func C_31F4() {
	var bp_12 [16]byte
	D_8CCA = Party.f_1d8
	Party.f_1d8 = 1
	CurMode = MOD_VISION
	Gra_13()
	dspl_Stats()
	Gra_10()
	music.Play(music.Stop)
	u4_puts("\n\n\nThere is a sudden darkness, and you find yourself alone in an empty chamber.\n")
	u_delay(4, 0)
	Gra_16("key7.ega", 0, nil)
	if !TST_MSK(Party.mItems, ST_KEY_C) || !TST_MSK(Party.mItems, ST_KEY_T) || !TST_MSK(Party.mItems, ST_KEY_L) {
		u4_puts("\nThou dost not have the Key of Three Parts.\n")
		C_2F9D(12)
	}
	u4_puts("\nYou use your Key of Three Parts.\n")
	u_delay(3, 0)
	u4_puts("\nA voice rings out:\n")
	if !C_310F("\"What is the Word of Passage?\"", "veramocor") {
		C_3010()
	}
	if D_8CCA != 8 {
		u4_puts("\nThou hast not proved thy leadership in all eight virtues.\n")
		u_delay(8, 0)
		C_3010()
	}
	for bp_02 := 7; bp_02 >= 0; bp_02-- {
		if *pKarmas[bp_02] != 0 {
			u4_puts("\nThou art not ready.\n")
			C_3010()
		}
	}
	u4_puts("\nPassage is granted.\n")
	u_delay(5, 0)
	Gra_10()
	for bp_02 := 0; bp_02 < 11; bp_02++ {
		u_delay(2, 0)
		u4_puts("\n\nThe voice asks:\n\n")
		u_delay(2, 0)
		ans := ""
		if bp_02 <= 7 {
			ans = D_1E98[151+bp_02]
		} else {
			ans = D_161A[bp_02-8]
		}
		if !C_310F(D_0BDA[bp_02], ans) {
			u_delay(1, 0)
			u4_puts("\nThy quest is not yet complete.\n")
			C_2F9D(bp_02)
		}
		pic := ""
		if bp_02 >= 8 {
			pic = D_161A[bp_02-8]
		} else {
			pic = D_1620[bp_02]
		}
		Gra_16(pic+".ega", 0, nil)
		if bp_02 == 7 {
			u_delay(3, 0)
			u4_puts("\nThou art well versed in the virtues of the Avatar.\n")
			u_delay(5, 0)
		}
	}
	u4_puts("\n\nThe floor rumbles beneath your feet.\n")
	shakefx()
	shakefx()
	u_delay(5, 0)
	u4_puts("\nAbove the din, the voice asks:\n\nIf all eight virtues of the Avatar combine into and are derived from the Three Principles of Truth, Love and Courage...")
	C_2F7E()
	u4_puts("\n\nThen what is the one thing which encompasses and is the whole of all undeniable Truth, unending Love, and unyielding Courage?\n\n")
	u4_gets(bp_12[:], 15)
	Gra_CR()
	if strnicmp(bp_12[:], "infinity", 16) != 0 {
		ok := false
		for try := 0; try < 2 && !ok; try++ {
			u4_puts("\nThy thoughts are not pure.\nI ask again.\n")
			u_delay(3, 0)
			u4_puts("\n\nWhat is the one thing which encompasses and is the whole of all undeniable Truth, unending Love, and unyielding Courage?\n\n")
			u4_gets(bp_12[:], 15)
			Gra_CR()
			ok = strnicmp(bp_12[:], "infinity", 16) == 0
		}
		if !ok {
			u_delay(1, 0)
			u4_puts("\nThou dost not know the true nature of the Universe.\n")
			C_2F9D(11)
		}
	}
	C_3025()
}
