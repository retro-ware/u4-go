// Package title is a port of Ultima IV's TITLE.EXE (SRC-TITLE in u4-decompiled): the Origin
// logo and title animation, the "view" of Britannia with the two animated monsters, the menu,
// and character creation (the intro story, the gypsy's questions, writing PARTY.SAV and
// MONSTERS.SAV for a new game). It reads the original data files (*.EGA pictures, SHAPES.EGZ,
// CHARSET.EGA, PARTY.NEW) from the game directory at run time.
//
// EGA only: the CGA (*.PIC, SHAPES.CGA, CHARSET.CGA) and Tandy paths of GRAPH.ASM, LOW.ASM and
// the C are dropped; D_7078 is always 2.
//
// One Go file per C/asm module, C names kept. Drawing goes to an ega.Driver's framebuffer,
// which the host shows (see Main). The original's CPU-timed busy loops become real-time
// waits (pacing in low.go, calibrated by host.SpeedInfo).
package title

// Port of TITLE_X.C: the monster animation counters.

var (
	D_0036 uint16 = 0
	D_0038 uint16 = 0
)
