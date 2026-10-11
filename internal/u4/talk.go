package u4

// Port of U4_TALK.C (signed char): talking to the citizens (.TLK records) and shop dispatch.

import (
	"strings"

	"github.com/memmaker/u4-go/internal/music"
)

var D_2A7A = "Funny, no response!\n"

var D_2A90 = [10]struct {
	_00 string
	_02 func()
}{
	{ /*D_2A3E*/ "bye", nil},
	{ /*D_2A42*/ "name", TLK_name},
	{ /*D_2A47*/ "look", TLK_look},
	{ /*D_2A4C*/ "job", TLK_job},
	{ /*D_2A50*/ "health", TLK_health},
	{"", TLK_special1},
	{"", TLK_special2},
	{ /*D_2A57*/ "join", TLK_join},
	{ /*D_2A5C*/ "give", TLK_give},
	{ /*D_2A61*/ "", nil},
}

/*
D_8CCE:
0.name
1.pronoun
2.look
3.job
4.health
5.personnal answer1
6.personnal answer2
7.special question
8.special answer Y
9.special answer N
10.personnal question1
11.personnal question2
*/
var D_8CCE [12]string

/*special question*/
func C_A163() {
	var bp_04 [4]byte

	Gra_CR()
	u_kbread()
	u4_puts(D_8CCE[7])
	u4_puts( /*D_2A62*/ "\n\nYou say: ")
	for {
		u4_gets(bp_04[:], 4)
		Gra_CR()
		if bp_04[0] == 0 {
			break
		}
		c := int(bp_04[0])
		u4_toupper(&c) // u4_toupper2
		bp_04[0] = byte(c)
		if bp_04[0] != 'N' && bp_04[0] != 'Y' {
			u4_puts( /*D_2A6E*/ "Yes or no!\n")
		}
		if bp_04[0] == 'N' || bp_04[0] == 'Y' {
			break
		}
	}
	if bp_04[0] == 0 {
		return
	}
	if bp_04[0] == 'Y' {
		if D_95CE[1] != 0 {
			karma_dec(&Party._humil, 5)
		}
		u4_puts(D_8CCE[8])
	} else {
		if D_95CE[1] != 0 && uint32(Party._moves>>4) != uint32(Party.f_1ec) {
			karma_inc(&Party._humil, 10)
		}
		u4_puts(D_8CCE[9])
	}
	Party.f_1ec = uint16(Party._moves >> 4)
	Gra_CR()
}

// C_A22D prints a string then CR; nil is the C's NULL pointer. (The C calls pass an extra,
// ignored, leading argument.)
func C_A22D(bp04 *string) {
	if bp04 == nil {
		return
	}
	u4_puts(*bp04)
	Gra_CR()
}

func TLK_special1() { /*C_A245*/
	C_A22D(&D_8CCE[5])
}

func TLK_special2() { /*C_A253*/
	C_A22D(&D_8CCE[6])
}

func TLK_name() { /*C_A261*/
	u4_puts(D_8CCE[1])
	u4_puts( /*D_2AB8*/ " says: I am ")
	u4_puts(D_8CCE[0])
	Gra_CR()
}

func TLK_look() { /*C_A280*/
	u4_puts( /*D_2AC5*/ "You see ")
	C_A22D(&D_8CCE[2])
}

func TLK_job() { /*C_A299*/
	C_A22D(&D_8CCE[3])
}

func TLK_health() { /*C_A2AB*/
	C_A22D(&D_8CCE[4])
}

var D_2BB2 = [8]string{
	/*D_2ACE*/ "honest",
	/*D_2AD5*/ "compassionate",
	/*D_2AE3*/ "valiant",
	/*D_2AEB*/ "just",
	/*D_2AF0*/ "sacrificial",
	/*D_2AFC*/ "honorable",
	/*D_2B06*/ "spiritual",
	/*D_2B10*/ "humble",
}

var D_8CE6 uint16 /*type?*/

func TLK_join() { /*C_A2BD*/
	var bp_02 int

	k := Party._loc - 0x05 // unsigned: the >= 8 test also catches non-towns
	if uint16(D_8742._npc._tlkidx[D_8CE6]) != 1 ||
		k >= 8 ||
		k == uint16(int16(sc(Party.chara[0]._class))) {
		u4_puts(D_8CCE[1])
		u4_puts( /*D_2B17*/ " says: I cannot join thee.\n")
		return
	}
	if *pKarmas[k] < 40 && *pKarmas[k] != 0 {
		u4_puts( /*D_2B33*/ "Thou art not ")
		u4_puts(D_2BB2[k])
		u4_puts( /*D_2B41*/ " enough for me to join thee.\n")
		return
	}
	if 100*Party.f_1d8+100 > Party.chara[0]._HP[1] { // 16-bit wrap
		u4_puts( /*D_2B5F*/ "Thou art not experienced enough for me to join thee.\n")
		return
	}
	u4_puts( /*D_2B95*/ "I am honored to join thee!\n")
	D_8742._npc._tile[31] = 0
	D_8742._npc._gtile[31] = 0
	D_8742._npc._var[31] = 0
	D_8742._npc._tlkidx[31] = 0
	t_callback()
	for bp_02 = 7; bp_02 >= 0; bp_02-- {
		if uint16(int16(sc(Party.chara[bp_02]._class))) == k {
			break
		}
	}
	if bp_02 != -1 {
		C_6FF9(&Party.chara[Party.f_1d8], &Party.chara[bp_02])
	}
	Party.f_1d8++
	Gra_13()
	dspl_Stats()
}

func TLK_give() { /*C_A3A2*/
	if D_8742._npc._tile[D_8CE6] != TIL_58 {
		u4_puts(D_8CCE[1])
		u4_puts( /*D_2BC2*/ " says: I do not need thy gold.  Keep it!\n")
		return
	}
	u4_puts( /*D_2BEC*/ "How much?\x12\x12\x12\b\b")
	bp_02 := int16(AskInt(2)) // int bp_02 = long
	if bp_02 > 0 {
		if Party._gold < uint16(bp_02) {
			u4_puts( /*D_2BFB*/ "Thou hast not that much gold!\n")
		} else {
			Party._gold -= uint16(bp_02)
			dspl_Stats()
			u4_puts(D_8CCE[1])
			u4_puts( /*D_2C1A*/ " says: Oh, Thank thee! I shall never forget thy kindness!\n")
			if uint32(Party._moves>>4) != uint32(Party.f_1ec) {
				karma_inc(&Party._compa, 2)
			}
			Party.f_1ec = uint16(Party._moves >> 4)
		}
	}
	C_A22D(nil)
}

/*parse talk info*/
func C_A443(si []byte) {
	p := 0
	for di := 0; di != 12; di++ {
		q := p
		for q < len(si) && si[q] != 0 { // C runs on; stay inside the record
			q++
		}
		D_8CCE[di] = string(si[min(p, len(si)):q])
		p = q + 1
	}
	if s := D_8CCE[2]; s != "" {
		c := int(s[0])
		u4_tolower(&c)
		D_8CCE[2] = string([]byte{byte(c)}) + s[1:]
	}
}

// C_A47F strips trailing blanks (the C does it in place on the TLK buffer).
func C_A47F(bp04 *string) {
	*bp04 = strings.TrimRight(*bp04, " ")
}

/*talk to citizen*/
func C_A4B4(bp04 int) {
	var bp_02, bp_04 int

	copy(D_95CE[:], at(File_TLK, int(uint16(int16((int(D_8742._npc._tlkidx[bp04])-1)*0x120)))))
	bp_04 = int(Party.f_1d8)
	D_9452 = D_8742._npc._tile[bp04]
	C_A443(D_95CE[3:])
	/*personnal question 1 & 2*/
	D_2A90[5]._00 = D_8CCE[10]
	D_2A90[6]._00 = D_8CCE[11]
	C_A47F(&D_2A90[5]._00)
	C_A47F(&D_2A90[6]._00)
	/* */
	u4_puts( /*D_2C55*/ "\nYou meet ")
	C_A22D(&D_8CCE[2])
	/*randomly says his name*/
	if U4_RND1(1) != 0 {
		Gra_CR()
		u4_puts(D_8CCE[1])
		u4_puts( /*D_2C60*/ " says: I am ")
		u4_puts(D_8CCE[0])
		Gra_CR()
	}
	bp_02 = 0
	for {
		var si int
		var bp_12 [12]byte

		u4_puts( /*D_2C6D*/ "\nYour Interest:\n")
		u4_gets(bp_12[:], 11)
		Gra_CR()
		if bp_12[0] == 0 {
			break
		}
		si = int(u_rand_a())
		if si < sc(D_95CE[2]) {
			if sc(D_95CE[2])-si >= 0x40 {
				/*he/she gets upset*/
				if strnicmp([]byte(D_8CCE[0]) /*D_2C7E*/, "a ", 2) != 0 && strnicmp([]byte(D_8CCE[0]) /*D_2C81*/, "the ", 4) != 0 {
					u4_puts(D_8CCE[0])
				} else {
					u4_puts(D_8CCE[1])
				}
				u4_puts( /*D_2C86*/ " says: On guard! Fool!\n")
				D_8742._npc._var[bp04] = 0xff
			} else {
				u4_puts(D_8CCE[1])
				u4_puts( /*D_2C9E*/ " turns away!\n\n")
			}
			return
		}
		for si = 0; D_2A90[si]._00 != ""; si++ {
			if strnicmp(bp_12[:], D_2A90[si]._00, 4) == 0 {
				if D_2A90[si]._02 != nil {
					Gra_CR()
					D_8CE6 = uint16(bp04)
					D_2A90[si]._02()
					if int(Party.f_1d8) != bp_04 {
						return /*looks like he/she has joined*/
					}
				} else {
					bp_02 = 1
				}
				if bp_02 == 0 {
					if sc(D_95CE[0]) == si {
						C_A163()
					}
				}
				break
			}
		}
		if D_2A90[si]._00 == "" {
			u4_puts( /*D_2CAD*/ "That I cannot help thee with.\n")
		}
		if bp_02 != 0 {
			break
		}
	}
	u4_puts( /*D_2CCC*/ "\nBye.\n")
}

/*shops'y positions by town*/
var D_2CD4 = [16][8]uint8{
	{0x00, 0x00, 0x00, 0x00, 0x00, 0x1A, 0x00, 0x00}, /*LB*/

	{0x00, 0x00, 0x00, 0x00, 0x00, 0x0C, 0x00, 0x00}, /*lycaeum*/
	{0x00, 0x00, 0x00, 0x00, 0x00, 0x0F, 0x00, 0x00}, /*empath*/
	{0x00, 0x00, 0x00, 0x00, 0x00, 0x0C, 0x00, 0x00}, /*serpent*/

	{0x00, 0x00, 0x0E, 0x00, 0x1A, 0x1B, 0x02, 0x00}, /*MOONGLOW*/
	{0x03, 0x07, 0x06, 0x02, 0x00, 0x1D, 0x0C, 0x00}, /*BRITAIN*/
	{0x09, 0x05, 0x00, 0x13, 0x00, 0x06, 0x1A, 0x00}, /*JHELOM*/
	{0x00, 0x00, 0x18, 0x00, 0x00, 0x19, 0x00, 0x00}, /*YEW.ULT*/
	{0x1C, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03, 0x00}, /*MINOC*/
	{0x14, 0x18, 0x00, 0x02, 0x00, 0x00, 0x03, 0x00}, /*TRINSIC*/
	{0x00, 0x00, 0x11, 0x00, 0x04, 0x1B, 0x0D, 0x00}, /*SKARA*/
	{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, /*MAGINCIA*/

	{0x00, 0x04, 0x1A, 0x05, 0x07, 0x00, 0x00, 0x00}, /*paws*/
	{0x0B, 0x11, 0x00, 0x19, 0x08, 0x00, 0x00, 0x07}, /*den*/
	{0x14, 0x00, 0x00, 0x16, 0x00, 0x00, 0x15, 0x1A}, /*vesper*/
	{0x00, 0x00, 0x00, 0x00, 0x00, 0x1A, 0x00, 0x00}, /*cove*/
}

/*shops functions callbacks*/
var D_2D54 = [10]func(){
	SHP_weapon,
	SHP_armor,
	SHP_food,
	SHP_pub,
	SHP_reagent,
	SHP_healer,
	SHP_inn,
	SHP_guild,

	SHP_horse,
	SHP_hawkwind,
}

/*shop at Y*/
func C_A686(bp04 uint8) {
	var si int

	for si = 7; si >= 0; si-- {
		if D_2CD4[Party._loc-1][si] == bp04 {
			break
		}
	}
	if si == -1 && bp04 == 0x18 && Party._loc == 0x0d {
		si = 8 /*patch! paws' horses*/
	}
	if si == -1 && bp04 == 0x19 && Party._loc == 0x01 {
		si = 9 /*patch! LB's hawkwind*/
	}
	if si == -1 {
		u4_puts(D_2A7A)
		return
	}
	music.Play(music.Shopping)
	D_2D54[si]()
}

func CMD_Talk() { /*C_A6F3*/
	var loc_A, loc_E uint8
	var loc_B, loc_C, loc_D int

	u4_puts( /*D_2D68*/ "Talk\n")
	if Party.f_1dc != 0 {
		w_DriftOnly()
		return
	}
	AskDir( /*D_2D6E*/ "Dir: ", &loc_B, &loc_D)
	if loc_B|loc_D == 0 {
		return
	}
	loc_E = uint8(int(Party._x) + loc_B)
	loc_A = uint8(int(Party._y) + loc_D)
	loc_C = C_0A8B(loc_E, loc_A)
	if loc_E > 31 || loc_A > 31 || Party._loc == 0 {
		u4_puts(D_2A7A)
		return
	}
	if *D_8742.x32x32(int(loc_A), int(loc_E)) >= TIL_60 &&
		*D_8742.x32x32(int(loc_A), int(loc_E)) <= TIL_7E {
		/*alphabet*/
		loc_E = uint8(int(loc_E) + loc_B)
		loc_A = uint8(int(loc_A) + loc_D)
		if loc_E > 31 || loc_A > 31 {
			u4_puts(D_2A7A)
			return
		}
		loc_C = C_0A8B(loc_E, loc_A)
		if loc_C == -1 || D_8742._npc._tile[loc_C] != TIL_52 {
			u4_puts(D_2A7A)
			return
		}
		C_A686(uint8(int(loc_A) - loc_D))
		return
	}
	if loc_C == -1 || D_8742._npc._tlkidx[loc_C] == 0 {
		u4_puts(D_2A7A)
		return
	}
	if D_8742._npc._tile[loc_C] == TIL_5E {
		music.Play(music.LordBritish)
		C_E59B()
		return
	}
	C_A4B4(loc_C)
}
