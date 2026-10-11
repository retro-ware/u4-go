package host

// The modern layout's dungeon view: the dungeon level around the party drawn in 3D with the
// core of the BlobberRPG engine (textured quads in camera space, near-clipped, projected with
// perspective-correct texturing in a shader and painted far to near). It only reads what the
// game reports through DungeonView; the game itself moves the party a square at a time.

import (
	"bytes"
	"embed"
	"image"
	"image/color"
	"image/png"
	"math"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
)

// Dungeon is what the 3D view shows: one 8x8 level (it wraps), the party's square and facing.
type Dungeon struct {
	Map   [64]byte // the level's squares, y*8+x, in the game's DNG coding
	X, Y  int
	Dir   int  // 0 west, 1 north, 2 east, 3 south
	Lit   bool // a torch or light spell burns; else all is dark
	Frame int  // the monsters' animation frame, 0..3
	Tiles *[256][256]byte
}

// DungeonView returns the dungeon around the party, or nil when the map window shows
// something else.
var DungeonView func() *Dungeon

//go:embed dngtex/*.png
var dngFS embed.FS

const (
	near    = 0.05
	eyeBack = 0.3 // the eye sits behind the square's centre, so the side passages show
	reach   = 6   // squares a ray of sight goes
	fogDist = 5.0 // depth at which the torchlight is gone
)

// the DNG coding's high nibbles
const (
	dLadderUp = 0x10 + iota*0x10
	dLadderDown
	dLadderBoth
	dChest
	dHoleUp
	dHoleDown
	dOrb
	dTrap
	dFountain
	dField
	dAltar
	dDoor
	dRoom
	dSecret
	dWall
)

const texSrc = `//kage:unit pixels
package main

var Fog float

func Fragment(dstPos vec4, srcPos vec2, color vec4, custom vec4) vec4 {
	z := 1 / custom.z
	uv := clamp(custom.xy*z, 0.0005, 0.9995)
	c := imageSrc0At(imageSrc0Origin() + uv*imageSrc0Size())
	if c.a == 0 {
		discard()
	}
	k := clamp(1-z/Fog, 0, 1)
	return vec4(c.rgb*k, c.a) * color.a
}
`

type vert struct{ x, y, z, u, v float64 }

type face struct {
	v     [4]vert
	tex   *ebiten.Image
	dist  float64
	alpha float32
	pass  int // 0 floor and ceiling, 1 the rest, far to near
}

type cam struct{ x, z, fwx, fwz, rtx, rtz float64 }

var d3 struct {
	sh                                           *ebiten.Shader
	wall, door, floor, ceil, floorHole, ceilHole *ebiten.Image
	ladder                                       *ebiten.Image
	field                                        [4]*ebiten.Image
	tiles                                        map[int]*ebiten.Image
	lut                                          [256][4]byte
}

func loadPNG(name string) *image.RGBA {
	b, err := dngFS.ReadFile("dngtex/" + name + ".png")
	if err != nil {
		panic(err)
	}
	img, err := png.Decode(bytes.NewReader(b))
	if err != nil {
		panic(err)
	}
	r := image.NewRGBA(img.Bounds())
	for y := 0; y < r.Rect.Dy(); y++ {
		for x := 0; x < r.Rect.Dx(); x++ {
			r.Set(x, y, img.At(x, y))
		}
	}
	return r
}

// d3Init loads BlobberRPG's stone, floor and ceiling and makes the rest from them.
func d3Init() {
	sh, err := ebiten.NewShader([]byte(texSrc))
	if err != nil {
		panic(err)
	}
	d3.sh = sh
	stone, floor, ceil := loadPNG("stone"), loadPNG("floor"), loadPNG("ceiling")
	d3.wall = ebiten.NewImageFromImage(stone)
	d3.floor = ebiten.NewImageFromImage(floor)
	d3.ceil = ebiten.NewImageFromImage(ceil)
	d3.door = ebiten.NewImageFromImage(genDoor(stone))
	d3.floorHole = ebiten.NewImageFromImage(genHole(floor))
	d3.ceilHole = ebiten.NewImageFromImage(genHole(ceil))
	d3.ladder = ebiten.NewImageFromImage(genLadder())
	for i, c := range [4]color.RGBA{{60, 220, 60, 255}, {80, 140, 255, 255}, {255, 90, 30, 255}, {190, 90, 255, 255}} {
		d3.field[i] = ebiten.NewImageFromImage(genField(c, i))
	}
}

// genDoor puts a wooden door with iron bands into the stone wall.
func genDoor(stone *image.RGBA) *image.RGBA {
	img := image.NewRGBA(stone.Rect)
	copy(img.Pix, stone.Pix)
	n := img.Rect.Dx()
	x0, x1, y0 := n*7/32, n*25/32, n*6/32
	for y := y0; y < n; y++ {
		for x := x0; x < x1; x++ {
			c := color.RGBA{110, 70, 35, 255}
			switch {
			case x < x0+n/64 || x >= x1-n/64 || y < y0+n/64:
				c = color.RGBA{40, 30, 20, 255} // frame
			case (x-x0)%(n/8) < n/128+1:
				c = color.RGBA{70, 45, 22, 255} // gaps between the planks
			case (y-y0)%(n/4) < n/32:
				c = color.RGBA{60, 60, 66, 255} // iron bands
			case x > x1-n/8 && x < x1-n/8+n/32 && y > n*19/32 && y < n*21/32:
				c = color.RGBA{200, 170, 60, 255} // the handle
			default:
				k := uint8(int(c.R) - (x*7+y*3)%17)
				c.R, c.G = k, uint8(int(c.G)-(x*5+y)%11)
			}
			img.SetRGBA(x, y, c)
		}
	}
	return img
}

// genHole darkens a round opening in the middle of a floor or ceiling, a trapdoor's shaft.
func genHole(src *image.RGBA) *image.RGBA {
	img := image.NewRGBA(src.Rect)
	copy(img.Pix, src.Pix)
	n := float64(img.Rect.Dx())
	for y := 0; y < img.Rect.Dy(); y++ {
		for x := 0; x < img.Rect.Dx(); x++ {
			d := math.Hypot(float64(x)-n/2, float64(y)-n/2) / (n / 2)
			if d < 0.7 {
				k := math.Max(0, (d-0.55)/0.15) * 0.4
				o := img.PixOffset(x, y)
				for i := 0; i < 3; i++ {
					img.Pix[o+i] = uint8(float64(img.Pix[o+i]) * k)
				}
			}
		}
	}
	return img
}

// genLadder is two wooden rails and their rungs on transparency.
func genLadder() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	wood := color.RGBA{140, 95, 50, 255}
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			if (x >= 16 && x < 20) || (x >= 44 && x < 48) || (x > 16 && x < 48 && y%12 < 3) {
				c := wood
				c.R -= uint8((x + y*3) % 20)
				img.SetRGBA(x, y, c)
			}
		}
	}
	return img
}

// genField is a shimmering curtain of colour c, see-through in places.
func genField(c color.RGBA, seed int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			v := 0.5 + 0.25*math.Sin(float64(x)*0.4+float64(y)*0.15+float64(seed)) + 0.25*math.Sin(float64(y)*0.5-float64(x)*0.2)
			a := 0.35 + 0.5*v
			img.SetRGBA(x, y, color.RGBA{uint8(float64(c.R) * a), uint8(float64(c.G) * a), uint8(float64(c.B) * a), uint8(255 * a)})
		}
	}
	return img
}

// tile is the game's tile t as an image, colour 0 see-through.
func tile(g *game, d *Dungeon, t int) *ebiten.Image {
	if d3.tiles == nil || d3.lut != g.lut {
		for _, i := range d3.tiles {
			i.Deallocate()
		}
		d3.tiles, d3.lut = map[int]*ebiten.Image{}, g.lut
	}
	if i, ok := d3.tiles[t]; ok {
		return i
	}
	pix := make([]byte, 16*16*4)
	for p, c := range d.Tiles[t&0xff] {
		if c != 0 {
			copy(pix[p*4:p*4+4], g.lut[c][:])
		}
	}
	i := ebiten.NewImage(16, 16)
	i.WritePixels(pix)
	d3.tiles[t] = i
	return i
}

// solid squares stop the eye and the feet.
func solid(t byte) bool { return t >= dDoor }

// drawDungeon paints the 3D view over the whole of dst.
func (g *game) drawDungeon(dst *ebiten.Image, d *Dungeon) {
	dst.Fill(color.Black)
	if !d.Lit {
		return // "It's Dark!": the original shows nothing either
	}
	if d3.sh == nil {
		d3Init()
	}
	w, h := float64(dst.Bounds().Dx()), float64(dst.Bounds().Dy())
	focal := w / 2
	at := func(x, y int) byte { return d.Map[(y&7)*8+x&7] }

	ang := float64(d.Dir-1) * math.Pi / 2
	c := cam{fwx: math.Sin(ang), fwz: -math.Cos(ang), rtx: math.Cos(ang), rtz: math.Sin(ang)}
	c.x = float64(d.X) + 0.5 - c.fwx*eyeBack
	c.z = float64(d.Y) + 0.5 - c.fwz*eyeBack

	var faces []*face
	q := func(a, b, cc, e vert, t *ebiten.Image, pass int) *face {
		f := &face{v: [4]vert{a, b, cc, e}, tex: t, pass: pass, alpha: 1}
		mx, mz := (a.x+cc.x)/2-c.x, (a.z+cc.z)/2-c.z
		f.dist = mx*c.fwx + mz*c.fwz
		if f.dist > -1 {
			faces = append(faces, f)
		}
		return f
	}
	// bill stands a camera-facing picture wid wide and hgt high on the floor at (x, z)
	bill := func(x, z, wid, hgt float64, t *ebiten.Image) *face {
		hx, hz := c.rtx*wid/2, c.rtz*wid/2
		return q(vert{x - hx, 0.5 - hgt, z - hz, 0, 0}, vert{x + hx, 0.5 - hgt, z + hz, 1, 0},
			vert{x + hx, 0.5, z + hz, 1, 1}, vert{x - hx, 0.5, z - hz, 0, 1}, t, 1)
	}
	dirs := [4][2]int{{0, -1}, {1, 0}, {0, 1}, {-1, 0}}

	for cell := range visible(c, w, focal, at) {
		cx, cy := cell[0], cell[1]
		t := at(cx, cy)
		kind := t & 0xf0
		x0, z0, x1, z1 := float64(cx), float64(cy), float64(cx+1), float64(cy+1)
		fl, ce := d3.floor, d3.ceil
		switch kind {
		case dLadderUp, dHoleUp:
			ce = d3.ceilHole
		case dLadderDown, dHoleDown:
			fl = d3.floorHole
		case dLadderBoth:
			fl, ce = d3.floorHole, d3.ceilHole
		}
		q(vert{x0, 0.5, z0, 0, 0}, vert{x1, 0.5, z0, 1, 0}, vert{x1, 0.5, z1, 1, 1}, vert{x0, 0.5, z1, 0, 1}, fl, 0)
		q(vert{x0, -0.5, z0, 0, 0}, vert{x1, -0.5, z0, 1, 0}, vert{x1, -0.5, z1, 1, 1}, vert{x0, -0.5, z1, 0, 1}, ce, 0)
		type side struct{ ax, az, bx, bz float64 }
		for i, s := range []side{{x0, z0, x1, z0}, {x1, z0, x1, z1}, {x1, z1, x0, z1}, {x0, z1, x0, z0}} {
			n := at(cx+dirs[i][0], cy+dirs[i][1]) & 0xf0
			if !solid(n) {
				continue
			}
			tex := d3.wall
			if n == dDoor || n == dRoom {
				tex = d3.door
			}
			q(vert{s.ax, -0.5, s.az, 0, 0}, vert{s.bx, -0.5, s.bz, 1, 0}, vert{s.bx, 0.5, s.bz, 1, 1}, vert{s.ax, 0.5, s.az, 0, 1}, tex, 1)
		}
		mx, mz := x0+0.5, z0+0.5
		switch kind {
		case dLadderUp, dLadderDown, dLadderBoth:
			if cx != d.X || cy != d.Y { // our own shows as the hole above or below
				bill(mx, mz, 0.5, 1, d3.ladder)
			}
		case dChest:
			bill(mx, mz, 0.4, 0.4, tile(g, d, 0x3c))
		case dOrb:
			bill(mx, mz, 0.4, 0.4, tile(g, d, 0x4e))
		case dFountain:
			bill(mx, mz, 0.5, 0.5, tile(g, d, 0x02))
		case dAltar:
			bill(mx, mz, 0.5, 0.5, tile(g, d, 0x4a))
		case dField:
			f := bill(mx, mz, 1, 1, d3.field[t&3])
			f.alpha = 0.8
		}
		// monsters ride in the low nibble (C_ABE5), drawn as the game draws them (C_AD6A)
		if kind != dTrap && kind != dFountain && kind != dField && kind < dRoom && t&0x0f != 0 {
			f := bill(mx+c.fwx*eyeBack, mz+c.fwz*eyeBack, 0.7, 0.7, tile(g, d, 0x90+int(t&0xf)<<2-4+d.Frame&3))
			f.dist -= 0.4 // in front of what shares its square
		}
	}
	// ponytail: painter's sort by face centre, as BlobberRPG does; a z-buffer if overlaps show
	sort.Slice(faces, func(i, j int) bool {
		if faces[i].pass != faces[j].pass {
			return faces[i].pass < faces[j].pass
		}
		return faces[i].dist > faces[j].dist
	})
	for _, f := range faces {
		drawQuad(dst, c, f, w, h, focal)
	}
}

// visible casts a ray every few columns across the view; the open squares a ray crosses
// before a solid one are seen (a field is seen, but nothing behind it). Map coordinates
// are not wrapped here: the level repeats around the party.
func visible(c cam, w, focal float64, at func(x, y int) byte) map[[2]int]bool {
	cells := map[[2]int]bool{{int(math.Floor(c.x)), int(math.Floor(c.z))}: true}
	for sx := -w/2 - 8; sx <= w/2+8; sx += 4 {
		k := sx / focal
		dx, dz := c.fwx+c.rtx*k, c.fwz+c.rtz*k
		cx, cz := int(math.Floor(c.x)), int(math.Floor(c.z))
		stepX, stepZ := 1, 1
		tdx, tdz := math.Abs(1/dx), math.Abs(1/dz)
		tx, tz := (float64(cx+1)-c.x)*tdx, (float64(cz+1)-c.z)*tdz
		if dx < 0 {
			stepX, tx = -1, (c.x-float64(cx))*tdx
		}
		if dz < 0 {
			stepZ, tz = -1, (c.z-float64(cz))*tdz
		}
		for i := 0; i < reach; i++ {
			if tx < tz {
				cx, tx = cx+stepX, tx+tdx
			} else {
				cz, tz = cz+stepZ, tz+tdz
			}
			t := at(cx, cz)
			if solid(t & 0xf0) {
				break
			}
			cells[[2]int{cx, cz}] = true
			if t&0xf0 == dField {
				break
			}
		}
	}
	return cells
}

func lerp(a, b vert, t float64) vert {
	return vert{a.x + (b.x-a.x)*t, a.y + (b.y-a.y)*t, a.z + (b.z-a.z)*t, a.u + (b.u-a.u)*t, a.v + (b.v-a.v)*t}
}

// drawQuad transforms, near-clips and projects one textured quad (BlobberRPG's drawQuad,
// without its lights and subdivision: the shader's texturing is perspective-correct).
func drawQuad(dst *ebiten.Image, c cam, f *face, w, h, focal float64) {
	var poly [4]vert
	for i, v := range f.v {
		dx, dz := v.x-c.x, v.z-c.z
		v.x, v.z = dx*c.rtx+dz*c.rtz, dx*c.fwx+dz*c.fwz
		poly[i] = v
	}
	var out []vert
	for k := range poly {
		a, b := poly[k], poly[(k+1)%4]
		if a.z >= near {
			out = append(out, a)
		}
		if (a.z >= near) != (b.z >= near) {
			out = append(out, lerp(a, b, (near-a.z)/(b.z-a.z)))
		}
	}
	if len(out) < 3 {
		return
	}
	vs := make([]ebiten.Vertex, len(out))
	for k, v := range out {
		iz := 1 / v.z
		vs[k] = ebiten.Vertex{DstX: float32(w/2 + v.x*focal*iz), DstY: float32(h/2 + v.y*focal*iz),
			ColorR: 1, ColorG: 1, ColorB: 1, ColorA: f.alpha,
			Custom0: float32(v.u * iz), Custom1: float32(v.v * iz), Custom2: float32(iz)}
	}
	var idx []uint16
	for k := 1; k+1 < len(out); k++ {
		idx = append(idx, 0, uint16(k), uint16(k+1))
	}
	op := &ebiten.DrawTrianglesShaderOptions{Uniforms: map[string]any{"Fog": fogDist}}
	op.Images[0] = f.tex
	dst.DrawTrianglesShader(vs, idx, d3.sh, op)
}
