package title

// Port of BEEP.ASM: sound_1, the error beep (AVATAR.EXE's sound(1)): 16 toggles of the PC
// speaker, each held 0CAh*speed_info passes of a delay loop. The toggles go to
// internal/speaker as timed segments; the game waits for them as the CPU loop did.

import "github.com/memmaker/u4-go/internal/speaker"

// beepLoopCycles: dec/pushf/push/pop/popf/jnz on an 8088.
const beepLoopCycles = 67

func sound_1() { /*C_21BF*/
	n := 0xca * int(speed_info)
	us := float64(n*beepLoopCycles) / (4.77 * float64(speed_info))
	var segs []speaker.Seg
	on := false // "and al,not 3" then toggle bit 1
	for cx := 0x10; cx != 0; cx-- {
		segs = append(segs, speaker.Seg{On: on, Us: us})
		on = !on
	}
	speaker.Play(segs)
	cpu(0x10 * n * beepLoopCycles)
}
