package title

// Port of TITLE_0.C: the intro (Origin logo, "Ultima IV" dissolve, the view opening like a
// scroll), the view of Britannia with its scripted tile animation, the two animated monsters,
// the menu, and main. Plus Main, the entry point that stands in for DOS.

import (
	"github.com/memmaker/u4-go/internal/ega"
	"github.com/memmaker/u4-go/internal/host"
)

var (
	D_00C0 int16 = 0
	D_00C2 int16 = 1
	D_00C4 int16 = -1
	D_00C6 int16 = 0
)

// base tiles for view
var D_00C8 = [...]uint8{0x40, 0x80, 0x10, 0x38, 0x38, 0xC8, 0xC8, 0x24, 0x20, 0x88, 0xF0, 0xF8, 0x4D, 0x4F, 0x4E}

var (
	D_3A24 [19 * 5]uint8
	D_3A84 uint16
)

// delayLoopCycles: one pass of C_02A3's while(--long) on an 8088.
const delayLoopCycles = 40

// Main runs TITLE.EXE with ULTIMA.COM's command tail (nparam = its length byte at 80h,
// param0/param1 = the bytes at 82h/84h) and returns the exit code: '0' (the port has no
// disks or drives; the original returned its drive flag or the player disk's drive letter),
// or an error code. ULTIMA.COM then runs AVATAR.EXE with the
// same tail, 84h replaced by this code's low byte and 80h set to 4, and starts over at
// TITLE.EXE when AVATAR.EXE exits with 4.
//
// It must run on the host's game goroutine; it shows the screen through host.ShowFrame.
func Main(gameDir, saveDir string, nparam, param0, param1 byte) (code int) {
	GameDir, SaveDir = gameDir, saveDir
	psp = [0x100]byte{}
	for i := 0x81; i < 0x100; i++ {
		psp[i] = ' '
	}
	psp[0x80], psp[0x82], psp[0x84] = nparam, param0, param1
	resetGlobals()
	defer func() {
		if r := recover(); r != nil {
			if c, ok := r.(exitCode); ok {
				code = int(c)
				return
			}
			panic(r)
		}
	}()
	main_()
	return 0
}

type exitCode int

func exit(n int) { panic(exitCode(n)) }

var frame = host.Frame{Pix: make([]byte, host.ScreenW*host.ScreenH), Palette: new([768]byte)}

// resetGlobals puts the initialised data back as a fresh load of TITLE.EXE has it (ULTIMA.COM
// runs it again after each game), and shows the EGA planes as the host's frame.
func resetGlobals() {
	D_0036, D_0038 = 0, 0
	D_00C0, D_00C2, D_00C4, D_00C6 = 0, 1, -1, 0
	D_3A84, D_6E80 = 0, 0
	D_31C0, cursor_rate, D_6E76, D_6E77 = 0xffff, 0, 0, 0
	D_03CC, D_03CD = 0x35, 0x9b
	D_3A86 = [0x10]uint8{}
	Party = [sizeofParty]byte{}
	D_6976 = t_500{}
	cpuCarry = 0

	drv = ega.New()
	for i, c := range ega.Palette {
		copy(frame.Palette[i*3:], c[:])
	}
	host.ShowFrame = func() *host.Frame {
		drv.Pixels(frame.Pix)
		return &frame
	}
}

// animate the 2 "monsters": callback for u_delay
func t_callback() { /*C_019A*/
	if D_3A84 != 0 {
		D_00C0 = 0
	}
	D_00C0--
	if D_00C0+1 != 0 {
		return
	}
	D_00C0 = int16(200 * speed_info)
	if D_6E80 != 0 {
		cursor_rate = 200
		bp_02 := D_3380[D_0036]
		D_0036 = (D_0036 + 1) & 0x7f
		Gra_3(6, 32, int(D_344A[bp_02]), int(D_3438[bp_02]), pAnim, 0, -1, 0)
		bp_02 = D_33F8[D_0038]
		D_0038 = (D_0038 + 1) & 0x3f
		Gra_3(6, 32, int(D_345C[bp_02]), int(D_3438[bp_02]), pAnim, 0, -1, 34)
	} else {
		cursor_rate = 233
	}
}

func u4_putc(bp04 uint8) { /*C_027A*/
	if bp04 == 8 {
		txt_X--
		Gra_putchar(' ')
	} else {
		Gra_putchar(bp04)
		txt_X++
	}
}

// C_02A3 busy-waits bp04*speed_info loop passes, unless a key is waiting.
func C_02A3(bp04 int32) {
	if u_kbhit() != 0 {
		return
	}
	bp04 *= int32(speed_info)
	cpu(int(bp04) * delayLoopCycles)
}

// handwriting: "Lord British", dot by dot (slowly when bp04 is set)
func C_02D1(bp04 int) {
	for si := 0; D_346E[si] != 0; {
		Gra_dot(191-int(D_346E[si+1]), int(D_346E[si])+20, 3)
		Gra_dot(191-int(D_346E[si+1]), int(D_346E[si])+21, 3)
		si += 2
		if bp04 != 0 {
			C_02A3(440)
		}
	}
	if bp04 != 0 {
		u_delay(1, 1)
	}
}

// tiles animation: the npcs' frames, then the water and the flags
func C_034D() {
	n := &D_6976._npc
	for si := 0; si < 32; si++ {
		di := n._060[si]
		if (di >= 0x84 && di <= 0x8e) || (di >= 0x20 && di <= 0x2e) || (di >= 0x50 && di <= 0x5e) {
			if u_rand_b() < 0xc0 {
				n._000[si] = ((n._000[si] & 1) | di) ^ 1
			}
		} else if di >= 0x90 {
			if u_rand_b() < 0xc0 {
				n._000[si] = ((n._000[si] + 1) & 3) | di
			}
		} else if di != 0x80 {
			n._000[si] = di
		}
	}
	D_00C2--
	if D_00C2 != 0 {
		return
	}
	if speed_info == 1 {
		D_00C2 = 1
	} else {
		D_00C2 = int16(speed_info - 1)
	}
	Gra_animFlow(0x00)
	Gra_animFlow(0x01)
	Gra_animFlow(0x02)
	Gra_animFlow(0x4c)
	Gra_animFlow(0x44)
	Gra_animFlow(0x46)
	Gra_animFlow(0x47)
	Gra_animFlow(0x45)
	Gra_animFlag()
}

// the view's script (D_36E2): place/move/remove npcs, then build the frame's map in D_3A24
func C_041A() {
	n := &D_6976._npc
	if D_00C6 != 0 {
		D_00C6--
		return
	}
	for {
		D_00C4++
		c := D_36E2[D_00C4]
		switch c & 0xf0 {
		case 0x00, 0x10, 0x20, 0x30, 0x40, 0x50, 0x60:
			id := c & 0xf
			nx := D_36E2[D_00C4+1]
			n._040[id] = c >> 4
			n._020[id] = nx & 0x1f
			n._000[id] = D_00C8[id] + nx>>5
			if id == 1 {
				n._060[id] = D_00C8[id]
			} else {
				n._060[id] = n._000[id]
			}
			D_00C4++
		case 0x70:
			n._000[c&0xf] = 0
			n._060[c&0xf] = 0
		case 0x80: /*end frame*/
			D_3A24 = D_3683
			for bp_02 := 0; bp_02 < 32; bp_02++ {
				if n._000[bp_02] != 0 {
					D_3A24[19*int(n._040[bp_02])+int(n._020[bp_02])] = n._000[bp_02]
				}
			}
			D_00C6 = int16(c & 0xf)
			return
		default: /*0x90..0xf0*/
			D_00C4 = -1
		}
	}
}

// the "view"
func C_05A4() {
	D_3A84 = 1
	D_31C0 = 0
	for u_kbhit() == 0 {
		C_034D()
		C_041A()
		D_6E80 = 0
		Gra_0(5, 19, D_3A24[:], 104, 0, 0, 2)
		for si := speed_info / 2; si != 0 && u_kbhit() == 0; si-- {
			Gra_0(5, 19, D_3A24[:], 104, 0, 0, 2)
		}
		C_034D()
		C_041A()
		Gra_0(5, 19, D_3A24[:], 104, 0, 0, 2)
		for si := speed_info / 2; si != 0 && u_kbhit() == 0; si-- {
			Gra_0(5, 19, D_3A24[:], 104, 0, 0, 2)
		}
		D_6E80 = 1
	}
	D_3A84 = 0
	u_kbread()
}

// logo + splash + animation
func C_068C() {
	var loc_A, loc_B uint16
	var loc_C int

	u_delay(1, 1)
	/*"lord british"*/
	C_02D1(1)
	/*"and"*/
	Gra_3(4, 4, 19, 17, pTitle, 17, -1, 19)
	u_delay(1, 1)
	/*draw line*/
	for loc_A = 86; loc_A < 238; loc_A++ {
		Gra_dot(31, int(loc_A), 2)
		C_02A3(320)
	}
	/*"origin systems inc."*/
	for loc_A = 1; loc_A < 10; loc_A++ {
		Gra_3(21, int(loc_A), 9, 21, pTitle, 30-int(loc_A), -1, 9)
		C_02A3(5000)
	}
	u_delay(1, 1)
	if u_kbhit() == 0 {
		/*"present"*/
		for loc_A = 1; loc_A <= 5; loc_A++ {
			Gra_3(15, int(loc_A), 14, 5-int(loc_A), pTitle, 33, -1, 14)
			C_02A3(5000)
		}
		u_delay(1, 1)
		/*"Ultima iv"*/
		for loc_A = 0; loc_A < 57; loc_A++ {
			if u_kbhit() != 0 {
				loc_A = 56
			}
			Gra_B(30, 45, 5, 34, pTitle, 34, int(loc_A), 5)
			if speed_info > 4 {
				Gra_B(30, 45, 5, 34, pTitle, 34, int(loc_A), 5)
			}
		}
		u_delay(1, 1)
	} else {
		Gra_3(30, 45, 5, 34, pTitle, 34, -1, 5)
		Gra_3(15, 5, 14, 0, pTitle, 33, -1, 14)
	}
	/*"quest of the avatar"*/
	for loc_A = 1; loc_A <= 6; loc_A++ {
		Gra_3(33, int(loc_A), 3, 81, pTitle, 86-int(loc_A), -1, 3)
		Gra_3(33, int(loc_A), 3, 93-int(loc_A), pTitle, 86, -1, 3)
		C_02A3(4000)
	}
	/*open view like scroll*/
	for loc_A = 1; loc_A <= 20; loc_A++ {
		if u_kbhit() != 0 {
			loc_A = 20
		}
		C_034D()
		Gra_C(5, 19, D_3683[:], 104, pTitle, 2)
		if speed_info > 4 {
			Gra_C(5, 19, D_3683[:], 104, pTitle, 2)
		}
		loc_C = 20 - int(loc_A)
		Gra_3(int(loc_A)*2, 96, loc_C, 96, pTitle, 96, -1, loc_C)
		C_02A3(680)
		C_034D()
		Gra_C(5, 19, D_3683[:], 104, pTitle, 2)
	}
	pTitle = nil

	/*the 2 "monsters"*/
	pAnim = make([]byte, 0x8000)
	Gra_inflate("ANIMATE.EGA", pAnim)
	for loc_A = 1; loc_A <= 32; loc_A++ {
		if u_kbhit() != 0 {
			loc_A = 32
		}
		a := int(loc_A)
		loc_C = int(D_3380[D_0036])
		D_0036 = (D_0036 + 1) & 0x7f
		Gra_3(6, a, int(D_344A[loc_C]), int(D_3438[loc_C])-a+32, pAnim, 0, -1, 0)
		loc_C = int(D_33F8[D_0038])
		D_0038 = (D_0038 + 1) & 0x3f
		Gra_3(6, a, int(D_345C[loc_C]), int(D_3438[loc_C])-a+32, pAnim, 0, -1, 34)
		/*the "view"*/
		C_034D()
		C_041A()
		Gra_0(5, 19, D_3A24[:], 104, 0, 0, 2)
		for loc_B = speed_info / 2; loc_B != 0 && u_kbhit() == 0; loc_B-- {
			Gra_0(5, 19, D_3A24[:], 104, 0, 0, 2)
		}
		C_034D()
		C_041A()
		Gra_0(5, 19, D_3A24[:], 104, 0, 0, 2)
		for loc_B = speed_info / 2; loc_B != 0 && u_kbhit() == 0; loc_B-- {
			Gra_0(5, 19, D_3A24[:], 104, 0, 0, 2)
		}
	}
}

func C_0B1E(bp08, bp06 uint16, si string) {
	txt_Y = bp08
	txt_X = bp06
	for i := 0; i < len(si); i++ {
		u4_putc(si[i])
	}
}

// the menu
func C_0B45() {
	Gra_2()
	C_0B1E(14, 2, "In another world, in a time to come.")
	C_0B1E(16, 15, "Options:")
	C_0B1E(17, 11, "Return to the view")
	C_0B1E(18, 11, "Journey Onward")
	C_0B1E(19, 11, "Initiate New Game")
	C_0B1E(21, 3, "Conversion by James Van Artsdalen")
	C_0B1E(22, 5, "\x09 Copyright 1987 Lord British")
	txt_X = 24
	txt_Y = 16
	D_31C0 = 1
}

// redraw the title screen without the intro (back from an abandoned character creation)
func C_0BCA() {
	Gra_clrscr()
	pShapes = make([]byte, 0x8000)
	C_217E("SHAPES.EGZ", pShapes)
	sizzleShapes()

	pTitle = make([]byte, 0x8000)
	Gra_inflate("TITLE.EGA", pTitle)
	C_02D1(0)
	Gra_3(4, 4, 19, 17, pTitle, 17, -1, 19)
	for bp_02 := 86; bp_02 < 238; bp_02++ {
		Gra_dot(31, bp_02, 2)
	}
	Gra_3(21, 9, 9, 21, pTitle, 21, -1, 9)
	Gra_3(30, 45, 5, 34, pTitle, 34, -1, 5)
	Gra_3(15, 5, 14, 0, pTitle, 33, -1, 14)
	Gra_3(33, 6, 3, 81, pTitle, 80, -1, 3)
	Gra_3(33, 6, 3, 87, pTitle, 86, -1, 3)
	Gra_3(40, 96, 0, 96, pTitle, 96, -1, 0)
	pTitle = nil

	pAnim = make([]byte, 0x8000)
	Gra_inflate("ANIMATE.EGA", pAnim)

	bp_02 := D_3380[D_0036]
	D_0036 = (D_0036 + 1) & 0x7f
	Gra_3(6, 32, int(D_344A[bp_02]), int(D_3438[bp_02]), pAnim, 0, -1, 0)
	bp_02 = D_33F8[D_0038]
	D_0038 = (D_0038 + 1) & 0x3f
	Gra_3(6, 32, int(D_345C[bp_02]), int(D_3438[bp_02]), pAnim, 0, -1, 34)

	D_6E80 = 1
	C_0B45()
}

func main_() { /*C_0EAA*/
	D_31C0 = 0
	low_init()

	u4_toupper(&psp[0x84])
	u4_toupper(&psp[0x82])
	/*set drive number: dropped, the port has no drives*/
	/*set graphic type: the letter (C/E/T) or low_gra(); the port is EGA only*/
	D_7078 = uint16(low_gra())

	D_6E80 = 0
	pCharset = make([]byte, 0x1000)
	if Load("CHARSET.EGA", 0x1000, pCharset) == -1 {
		exit(6)
	}
	pShapes = make([]byte, 0x8000)
	C_217E("SHAPES.EGZ", pShapes)
	Gra_init()

	pTitle = make([]byte, 0x8000)
	Gra_inflate("TITLE.EGA", pTitle)
	D_3A84 = 0
	C_068C() /*logo + splash + animation*/

	D_6E80 = 1
	C_05A4()
	C_0B45()
	for {
		bp_02 := u_kbread()
		if u4_isupper(uint8(bp_02)) {
			bp_02 = bp_02&0xff00 | int(uint8(bp_02)+'a'-'A')
		}
		if uint8(bp_02) > ' ' && uint8(bp_02)&0x80 == 0 {
			Gra_putchar(uint8(bp_02))
		}
		switch bp_02 {
		case KBD_R: /*(R)eturn*/
			C_05A4()
			C_0B45()
		case KBD_I: /*(I)nitiate*/
			C_3030()
			C_0BCA()
		case KBD_J: /*(J)ourney*/
			pAnim, pShapes, pCharset = nil, nil, nil
			low_clean()
			exit('0')
		default:
			sound_1()
		}
	}
}

// ---- TITLE.H / U4_SCAN.H ----

func u4_isupper(c uint8) bool { return c >= 'A' && c <= 'Z' }
func u4_islower(c uint8) bool { return c >= 'a' && c <= 'z' }
func u4_toupper(c *uint8) {
	if u4_islower(*c) {
		*c -= 'a' - 'A'
	}
}

func b2u(b bool) uint16 {
	if b {
		return 1
	}
	return 0
}

const (
	KBD_BS    = 0x0e08
	KBD_0e7f  = 0x0e7f
	KBD_SPACE = 0x3920
	KBD_ENTER = 0x1c0d
	KBD_LEFT  = 0x4B00
	KBD_A     = 0x1E61
	KBD_B     = 0x3062
	KBD_I     = 0x1769
	KBD_J     = 0x246A
	KBD_R     = 0x1372
)
