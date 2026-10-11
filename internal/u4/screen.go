package u4

// The port's own glue between the game and the host: the EGA driver instance, showing its
// video memory as the host's frame, and the per-run state reset.

import (
	"github.com/memmaker/u4-go/internal/ega"
	"github.com/memmaker/u4-go/internal/host"
)

var frame = host.Frame{Pix: make([]byte, host.ScreenW*host.ScreenH), Palette: new([768]byte)}

// SetVGA switches the game to VGA (or EGA) graphics.
func SetVGA(on bool) {
	if drv != nil {
		drv.SetMode(on)
	}
}

// present shows the screen: DOS drew to video memory directly, the port lets Ebiten draw
// what the driver left in the framebuffer at the next tick.
func present() { host.Yield() }

// resetGlobals puts in place what DOS gave a freshly loaded AVATAR.EXE: the EGA driver
// (Gra_init only sets it up) shown as the host's frame.
func resetGlobals() {
	drv = ega.New(frame.Pix)
	drv.Palette = frame.Palette
	drv.SetMode(ega.VGA)
	drv.ReadFile = readGameFile
	drv.Present = present
	drv.Speaker = func(bool) {} // ponytail: shake clicks silent; route to internal/speaker if missed
	host.ShowFrame = func() *host.Frame { return &frame }
}
