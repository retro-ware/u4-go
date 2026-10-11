package u4

// Port of U4_ANIM.C (unsigned char): the view redraw t_callback runs whenever the game
// waits, with the tile animations, wind, moons and the line-of-sight raycast.

import "github.com/memmaker/u4-go/internal/host"

var D_1630 uint16

// animate force fields, water, ...
func C_34EA() {
	D_1630--
	if D_1630+1 == 0 {
		D_1630 = speed_info - 1
		for _, t := range []int{TIL_00, TIL_01, TIL_02, TIL_4C, TIL_44, TIL_46, TIL_47, TIL_45} {
			Gra_animFlow(t)
		}
	}
	Gra_animSpit()
}

// display direction
func C_353D(bp06 uint16, bp04 string) {
	bp_02, bp_04 := txt_X, txt_Y
	txt_Y = 23
	txt_X = 7
	u4_puts(bp04)
	switch bp06 {
	case DIR_N:
		u4_puts(" North")
	case DIR_S:
		u4_puts(" South")
	case DIR_E:
		u4_puts("  East")
	case DIR_W:
		u4_puts("  West")
	}
	if CurMode == MOD_DUNGEON {
		txt_Y = 0
		txt_X = 11
		u4_putc('L')
		u4_putc(int(Party._z) + '1')
	}
	txt_X, txt_Y = bp_02, bp_04
}

var D_1654 uint16

// update/display wind
func C_35C7() {
	D_1654--
	if D_1654+1 == 0 {
		D_1654 = speed_info - 1
		if U4_RND1(0xfc) == 0 {
			WindDir = uint16(U4_RND1(2)+int(WindDir)-1) & 3
		}
	}
	C_353D(WindDir, "Wind")
}

var D_1656 = [7]int{192, 96, 60, 36, 20, 12, 8}

func animOdds() int {
	if speed_info < 8 {
		return D_1656[speed_info-1]
	}
	return 0xff
}

// animate sprites
func C_3605() {
	n := &D_8742._npc
	for si := 0; si < 32; si++ {
		di := int(n._tile[si])
		if (di >= TIL_84 && di <= TIL_8E) || (di >= TIL_20 && di <= TIL_2E) || (di >= TIL_50 && di <= TIL_5E) {
			if U4_RND2(1) != 0 {
				if animOdds() > U4_RND2(0xff) {
					n._gtile[si] = uint8((int(n._gtile[si])&1 | di) ^ 1)
				}
			}
		} else if di >= TIL_90 {
			if animOdds() > U4_RND2(0xff) {
				n._gtile[si] = uint8((int(n._gtile[si])+1)&3 | di)
			}
		} else if di != TIL_80 {
			n._gtile[si] = uint8(di)
		}
	}
	C_34EA()
	Gra_animFlag()
}

func u4_sign(a int16) int16 {
	switch {
	case a > 0:
		return 1
	case a < 0:
		return -1
	}
	return 0
}

func C_36C7() {
	i := int(hit_y)*11 + int(hit_x)
	si := D_96F9[i]
	if hit_tile != 0 {
		D_96F9[i] = uint8(hit_tile)
	}
	Gra_05(11, 11, D_96F9[:], 4*80+2)
	if activeCharaX != -1 {
		di := D_95C8
		D_95C8 -= u4_sign(D_95C8)
		if D_95C8 == 0 {
			if di < 0 {
				D_95C8 = 4
			} else {
				D_95C8 = -9
			}
		}
		if D_95C8 > 0 {
			Gra_00(int(activeCharaY), int(activeCharaX))
		}
	}
	D_96F9[i] = si
}

func bnot(v int) int { return b2i(v == 0) }

// cast one ray
func C_376F(bp0a, bp08 int, bp06, bp04 uint8) {
	if bp06 > 10 || bp04 > 10 {
		return
	}
	si := int(bp04)*11 + int(bp06)
	if D_96F9[si] != TIL_7E {
		return
	}
	bp_04 := Combat._map[si]
	D_96F9[si] = bp_04
	// opaque tiles
	if bp_04 == TIL_06 || bp_04 == TIL_08 || bp_04 == TIL_7E || bp_04 == TIL_49 || bp_04 == TIL_7F {
		return
	}
	bp06 += uint8(bp0a)
	bp04 += uint8(bp08)
	C_376F(bp0a, bp08, bp06, bp04)
	if bp0a&bp08 != 0 {
		C_376F(bp0a, bp08, bp06, bp04-uint8(bp08))
		C_376F(bp0a, bp08, bp06-uint8(bp0a), bp04)
	} else {
		C_376F(bnot(bp0a)*bp08+bp0a, bp08-bnot(bp08)*bp0a, uint8(bp08)+bp06, bp04-uint8(bp0a))
		C_376F(bp0a-bnot(bp0a)*bp08, bnot(bp08)*bp0a+bp08, bp06-uint8(bp08), uint8(bp0a)+bp04)
	}
}

// process visible area (raycast)
func C_38BB() {
	for si := 10*11 + 10; si >= 0; si-- {
		D_96F9[si] = TIL_7E
	}
	D_96F9[5*11+5] = *Combat_MAP(5, 5) // display character's icon
	C_376F(0, -1, 5, 4)                // UP
	C_376F(0, 1, 5, 6)                 // DOWN
	C_376F(-1, 0, 4, 5)                // LEFT
	C_376F(1, 0, 6, 5)                 // RIGHT
	C_376F(1, 1, 6, 6)                 // DOWN RIGHT
	C_376F(1, -1, 6, 4)                // UP RIGHT
	C_376F(-1, 1, 4, 6)                // DOWN LEFT
	C_376F(-1, -1, 4, 4)               // UP LEFT
}

func C_3960() {
	n := &D_8742._npc
	for b := 31; b >= 0; b-- {
		a := n._x[b] - Party._x + 5
		c := n._y[b] - Party._y + 5
		if n._gtile[b] != 0 && a < 11 && c < 11 {
			*Combat_MAP(int(c), int(a)) = n._gtile[b]
		}
	}
}

func C_39BA() {
	C_3960()
	if D_9141 != 0 {
		b := D_9445 - Party._x + 5
		c := D_9448 - Party._y + 5
		if b < 11 && c < 11 {
			*Combat_MAP(int(c), int(b)) = D_9141
		}
	}
	tile_cur = *Combat_MAP(5, 5)
	tile_north = *Combat_MAP(4, 5)
	tile_south = *Combat_MAP(6, 5)
	tile_east = *Combat_MAP(5, 6)
	tile_west = *Combat_MAP(5, 4)
	*Combat_MAP(5, 5) = uint8(Party._tile)
	if D_9440 != 0 {
		C_38BB()
	} else {
		D_96F9 = Combat._map
	}
	C_36C7()
}

var (
	D_1664 uint8
	D_1665 uint8 // trammel counter
	D_1666 uint8 // felucca counter
)

func C_3A4F() uint8 {
	D_9445 = D_0814[(D_1665&0xe0)>>5]
	D_9448 = D_081C[(D_1665&0xe0)>>5]
	return (D_1664 & 0xc0) >> 6
}

var D_1668 uint16

// display trammel and felucca
func C_3A80() {
	D_1668--
	if D_1668+1 == 0 {
		D_1668 = speed_info / 2
		D_1664 += 0x40
		if D_1664 == 0 {
			D_1665 += 2
			D_1666 += 6
		}
		Party._trammel = uint16(D_1665 >> 5)
		Party._felucca = uint16(D_1666 >> 5)
		if D_1665&0x1f == 0 {
			D_9141 = C_3A4F() + 0x40
		} else if D_1665&0x1f == 0x1e {
			D_9141 = (C_3A4F() ^ 3) + 0x40
		}
	}
	bp_02, bp_04 := txt_X, txt_Y
	txt_Y = 0
	txt_X = 11
	Gra_putchar((((D_1665 >> 5) - 1) & 7) + 0x14)
	txt_X++
	Gra_putchar((((D_1666 >> 5) - 1) & 7) + 0x14)
	txt_X, txt_Y = bp_02, bp_04
}

func C_3B35() {
	C_3605() // animate sprites
	C_3A80() // display trammel and felucca
	b := 10*11 + 10
	for c := 10; c >= 0; c-- {
		for a := 10; a >= 0; a-- {
			Combat._map[b] = *D_8742.x32x32((c+int(D_959C.y)-5)&0xff, (a+int(D_959C.x)-5)&0xff)
			b--
		}
	}
	C_39BA() // display 11x11 icons zone
}

// display inside towne/castle ...
func C_3B83() {
	C_3605()
	for di := -5; di <= 5; di++ {
		for si := -5; si <= 5; si++ {
			x, y := int(Party._x)+si, int(Party._y)+di
			if x < 0 || x > 0x1f || y < 0 || y > 0x1f {
				*Combat_MAP(di+5, si+5) = TIL_04
			} else {
				*Combat_MAP(di+5, si+5) = *D_8742.x32x32(y, x)
			}
		}
	}
	C_39BA() // display 11x11 icons zone
	C_35C7() // update/display wind
}

var D_166A uint16

// move balloon
func C_3C08() {
	D_166A--
	if D_166A+1 == 0 {
		D_166A = speed_info
		Party._moves++
		switch WindDir {
		case DIR_N:
			C_2891()
		case DIR_S:
			C_2839()
		case DIR_E:
			C_28E9()
		case DIR_W:
			C_2941()
		}
	}
}

var D_166C = [8]int{1, 3, 3, 7, 7, 7, 7, 15}

func C_3C54() {
	loc_B := 15
	if speed_info < 8 {
		loc_B = D_166C[speed_info-1]
	}
	C_34EA()
	D_96F9 = Combat._map
	if CurMode != MOD_SHRINE {
		for a := 15; a >= 0; a-- {
			e := Fighters._tile[a]
			if e == 0 {
				continue
			}
			d := &D_96F9[int(Combat._npcY[a])*11+int(Combat._npcX[a])]
			c := &Fighters._gtile[a]
			if Fighters._sleeping[a] != 0 {
				*d = *c
			} else if e >= TIL_90 {
				if e == TIL_AC && *c == TIL_3C || U4_RND1(loc_B) == 0 {
					*d = *c
				} else if U4_RND1(loc_B) == 0 {
					*c = (*c+1)&3 | e
					*d = *c
				} else {
					*d = *c
				}
			} else if e == TIL_38 {
				*c = TIL_38
				*d = TIL_38
			} else {
				*c ^= uint8(U4_RND1(U4_RND2(1)))
				*d = *c
			}
		}
		for a := int(Party.f_1d8) - 1; a >= 0; a-- {
			bp_10 := Fighters._chtile[a]
			if bp_10 == 0 {
				continue
			}
			c := &D_96F9[int(Combat._charaY[a])*11+int(Combat._charaX[a])]
			if bp_10 == TIL_38 {
				D_944A[a] = bp_10
				*c = bp_10
				continue
			}
			D_944A[a] ^= uint8(bnot(U4_RND1(loc_B)))
			*c = D_944A[a]
		}
	}
	C_36C7()
}

// t_callback displays the game zone (C_3DC8). It is the one place the game redraws while
// it waits, so the port lets the frame show and paces the game here.
func t_callback() {
	defer func() {
		for i := 0; i < CallbackTicks; i++ {
			host.Yield()
		}
	}()
	if CurMode == MOD_VISION {
		cursor_rate = 233
		return
	}
	cursor_rate = 0
	if CurMode == MOD_OUTDOORS {
		if Party._tile == TIL_18 && Party.f_1dc != 0 {
			C_3C08()
		}
		C_35C7() // update/display wind
		C_3B35()
	} else if CurMode == MOD_BUILDING {
		C_3B83()
	} else if CurMode == MOD_DUNGEON {
		C_353D(Party._dir, "DIR:")
		C_34EA()
		C_B677()
		C_ADEF()
	} else { // Combat
		C_3C54()
	}
}
