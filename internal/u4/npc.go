package u4

// Port of U4_NPC.C (signed char): NPC movement, outdoor creature spawning and their
// ranged attacks.
//
// u4_abs and u4_sign (C_4FBE, C_4FD7) are defined in d3.go / anim.go; isign wraps u4_sign for
// ints. MSC pushes call arguments right to left, so where the C passes two u_rand_a()-based
// arguments in one call the last argument draws the first random number: the Go computes
// them in that order.

var D_1FD8 = "\nAttacked by "

var npcDummy uint8

// npcMap is D_8742._map.x32x32[y][x] with the C's unchecked indexing: out of the 32x32 map
// (the C reads the NPC table or garbage; C_4E94 discards such values) it gives a scratch byte.
func npcMap(y, x int) *uint8 {
	i := y*32 + x
	if i < 0 || i >= 0x400 {
		npcDummy = 0
		return &npcDummy
	}
	return D_8742.x32x32(y, x)
}

// can npc move?
func C_4E94(bp0a int, bp08, bp06, bp04 uint8) bool { // bp08: new x, bp06: new y
	if CurMode == MOD_BUILDING {
		if bp08 >= 32 || bp06 >= 32 {
			return false
		}
	}
	di := int(D_8742._npc._tile[bp0a])
	if C_29EF(bp04) {
		return false
	}
	if di != TIL_8C && di != TIL_8E {
		for si := 31; si >= 0; si-- {
			if D_8742._npc._tile[si] != 0 &&
				D_8742._npc._x[si] == bp08 &&
				D_8742._npc._y[si] == bp06 &&
				D_8742._npc._tile[si] != TIL_3C {
				return false
			}
		}
		if Party._x == bp08 && Party._y == bp06 {
			return false
		}
	}
	if di != TIL_80 &&
		bp08 == D_8742._npc._old_x[bp0a] &&
		bp06 == D_8742._npc._old_y[bp0a] &&
		U4_RND1(3) == 0 {
		return false
	}
	if di < TIL_80 {
		return C_2999(bp04)
	}
	if di < TIL_8E {
		if di == TIL_80 {
			return bp04 <= TIL_01
		}
		return bp04 <= TIL_02
	}
	if di == TIL_8E || di == TIL_94 || di == TIL_B4 || di == TIL_F0 || di == TIL_F8 || di == TIL_FC {
		return bp04 <= TIL_02 || C_2999(bp04)
	}
	if di == TIL_EC || di == TIL_9C {
		return bp04 >= TIL_03 && bp04 != TIL_45
	}
	return C_2999(bp04)
}

func C_4FF8(si int, bp06, bp04 uint8) {
	// quirk: "||" (the move is accepted when either coordinate is in range)
	if bp06 < 32 || bp04 < 32 {
		D_8742._npc._old_x[si] = D_8742._npc._x[si]
		D_8742._npc._x[si] = bp06
		D_8742._npc._old_y[si] = D_8742._npc._y[si]
		D_8742._npc._y[si] = bp04
	}
}

func C_5036(si int, bp06, bp04 uint8) {
	D_8742._npc._old_x[si] = D_8742._npc._x[si]
	D_8742._npc._x[si] = bp06
	D_8742._npc._old_y[si] = D_8742._npc._y[si]
	D_8742._npc._y[si] = bp04
}

// NPC attempts move[MOD_OUTDOORS]
func C_5062(bp0a, bp08, bp06, bp04 int) {
	loc_B := D_8742._npc._x[bp0a]
	loc_A := uint8(bp08 + int(loc_B) - int(D_95A5.x)<<4)
	loc_D := D_8742._npc._y[bp0a]
	loc_C := uint8(bp06 + int(loc_D) - int(D_95A5.y)<<4)
	if loc_A > 31 || loc_C > 31 {
		if bp04 >= 0 {
			// right-to-left argument order; 2nd argument is u4_sign of the unsigned
			// u_rand_a() (never -1): quirk of the original
			y := isign(int(u_rand_a()))
			x := isign(sc(u_rand_a()))
			C_5062(bp0a, x, y, bp04-1)
		}
		return
	}
	if U4_RND1(1) != 0 && bp08 != 0 && C_4E94(bp0a, uint8(int(loc_B)+bp08), loc_D, *npcMap(int(loc_C)-bp06, int(loc_A))) {
		C_5036(bp0a, uint8(int(loc_B)+bp08), loc_D)
		return
	}
	if bp06 != 0 && C_4E94(bp0a, loc_B, uint8(int(loc_D)+bp06), *npcMap(int(loc_C), int(loc_A)-bp08)) {
		C_5036(bp0a, loc_B, uint8(int(loc_D)+bp06))
		return
	}
	if C_4E94(bp0a, uint8(int(loc_B)+bp08), loc_D, *npcMap(int(loc_C)-bp06, int(loc_A))) {
		C_5036(bp0a, uint8(int(loc_B)+bp08), loc_D)
		return
	}
	if bp04 >= 0 {
		y := isign(sc(u_rand_a()))
		x := isign(sc(u_rand_a()))
		C_5062(bp0a, x, y, bp04-1)
	}
}

// NPC attempts move[MOD_BUILDING]; bp04: # of tries
func C_51A7(bp0a, bp08, bp06, bp04 int) {
	bp_02 := int(D_8742._npc._x[bp0a])
	bp_04 := int(D_8742._npc._y[bp0a])
	// The C indexes the map without bounds control; C_4E94 rejects such moves anyway.
	if U4_RND1(1) != 0 && bp08 != 0 {
		if C_4E94(bp0a, uint8(bp08+bp_02), uint8(bp_04), *npcMap(bp_04, bp08+bp_02)) {
			C_4FF8(bp0a, uint8(bp08+bp_02), uint8(bp_04))
		}
		return
	}
	if bp06 != 0 {
		if C_4E94(bp0a, uint8(bp_02), uint8(bp06+bp_04), *npcMap(bp06+bp_04, bp_02)) {
			C_4FF8(bp0a, uint8(bp_02), uint8(bp06+bp_04))
		}
		return
	}
	if C_4E94(bp0a, uint8(bp08+bp_02), uint8(bp_04), *npcMap(bp_04, bp08+bp_02)) {
		C_4FF8(bp0a, uint8(bp08+bp_02), uint8(bp_04))
		return
	}
	if D_8742._npc._var[bp0a] != 0x80 && bp04 >= 0 {
		y := isign(sc(u_rand_a()))
		x := isign(sc(u_rand_a()))
		C_51A7(bp0a, x, y, bp04-1)
	}
}

// move NPCs[MOD_BUILDING]
func C_5293() {
	for si := 31; si >= 0; si-- {
		if D_8742._npc._tile[si] != 0 && D_8742._npc._var[si] != 0 {
			if sc(D_8742._npc._var[si]) >= 0 {
				// random mode
				if U4_RND1(1) != 0 {
					y := isign(sc(u_rand_a()))
					x := isign(sc(u_rand_a()))
					C_51A7(si, x, y, 2)
				}
			} else {
				// following mode
				dx := sc(Party._x - D_8742._npc._x[si])
				dy := sc(Party._y - D_8742._npc._y[si])
				bp_02 := isign(dx)
				bp_04 := isign(dy)
				if u4_abs(dx)+u4_abs(dy) >= 2 {
					C_51A7(si, bp_02, bp_04, 2)
				} else if D_8742._npc._var[si] == 0xff {
					// agressive mode
					// -- attacked by --
					t_callback()
					u4_puts(D_1FD8)
					u4_puts(C_1513(D_8742._npc._tile[si]))
					Gra_CR()
					C_7DFE(si)
				}
			}
		}
	}
}

// creature/party distance ?: x sign, y sign, x distance, y distance
func C_535B(bp0c int, bp0a, bp08, bp06, bp04 *int) int {
	*bp04 = sc(Party._y - D_8742._npc._y[bp0c])
	*bp08 = isign(*bp04)
	*bp06 = sc(Party._x - D_8742._npc._x[bp0c])
	*bp0a = isign(*bp06)
	return u4_abs(*bp06) + u4_abs(*bp04)
}

func C_53AF(si int, bp06, bp04 uint8) {
	if C_4E94(si, bp06, bp04, *npcMap(int(bp04)-int(D_95A5.y)*16, int(bp06)-int(D_95A5.x)*16)) {
		if C_2A5A(uint16(D_8742._npc._gtile[si] & 3)) {
			D_8742._npc._old_x[si] = D_8742._npc._x[si]
			D_8742._npc._x[si] = bp06
			D_8742._npc._old_y[si] = D_8742._npc._y[si]
			D_8742._npc._y[si] = bp04
		}
	} else {
		D_8742._npc._gtile[si] = uint8((isign(sc(u_rand_a()))+int(D_8742._npc._gtile[si]))&3) | TIL_80
	}
}

func C_5443(si int) {
	var loc_A, loc_B, loc_C, loc_D int
	n := &D_8742._npc

	if C_535B(si, &loc_A, &loc_B, &loc_C, &loc_D) > 5 || U4_RND1(3) == 0 {
		if u4_abs(loc_C) < u4_abs(loc_D) {
			if loc_D >= 0 {
				n._gtile[si] = TIL_83
			} else {
				n._gtile[si] = TIL_81
			}
		} else {
			if loc_C >= 0 {
				n._gtile[si] = TIL_82
			} else {
				n._gtile[si] = TIL_80
			}
		}
	} else if n._gtile[si] == TIL_80 {
		C_53AF(si, n._x[si]-1, n._y[si])
	} else if n._gtile[si] == TIL_81 {
		C_53AF(si, n._x[si], n._y[si]-1)
	} else if n._gtile[si] == TIL_82 {
		C_53AF(si, n._x[si]+1, n._y[si])
	} else {
		C_53AF(si, n._x[si], n._y[si]+1)
	}
}

func C_5500(si, bp06, bp04 int) {
	n := &D_8742._npc
	if n._gtile[si] == TIL_80 {
		if bp06 < 0 {
			C_53AF(si, uint8(int(n._x[si])+bp06), n._y[si])
			return
		}
	} else if n._gtile[si] == TIL_81 {
		if bp04 < 0 {
			C_53AF(si, n._x[si], uint8(int(n._y[si])+bp04))
			return
		}
	} else if n._gtile[si] == TIL_82 {
		if bp06 > 0 {
			C_53AF(si, uint8(int(n._x[si])+bp06), n._y[si])
			return
		}
	} else {
		if bp04 > 0 {
			C_53AF(si, n._x[si], uint8(int(n._y[si])+bp04))
			return
		}
	}
	C_5443(si)
}

// bp08, bp06: direction as the C's unsigned chars (-1 = 255)
func C_5569(bp0a int, bp08, bp06, bp04 uint8) int {
	loc_E := D_8742._npc._x[bp0a] + bp08
	loc_A := D_8742._npc._y[bp0a] + bp06
	sound(3)
	for loc_C := 1; loc_C <= 3; loc_C++ {
		if Party._x == loc_E && Party._y == loc_A {
			return 1
		}
		if loc_B := C_0A58(31, loc_E, loc_A); loc_B != -1 {
			sound(6)
			if loc_B >= 8 || U4_RND1(3) == 0 {
				D_8742._npc._tile[loc_B] = 0
				D_8742._npc._gtile[loc_B] = 0
			}
			return 0
		}
		si := npcMap(int(loc_A)-int(D_95A5.y)*16, int(loc_E)-int(D_95A5.x)*16)
		loc_D := *si
		*si = bp04
		t_callback()
		*si = loc_D
		loc_E += bp08
		loc_A += bp06
	}
	return 0
}

func C_564B(bp0a, bp08, bp06 int, bp04 uint8) {
	if C_5569(bp0a, uint8(bp08), uint8(bp06), bp04) == 0 {
		hit_tile = 0
		return
	}
	t_callback()
	hit_x = 5
	hit_y = 5
	hit_tile = TIL_4F
	C_36C7()
	C_1584()
	hit_tile = 0
	t_callback()
}

// (the C passes 3 arguments to the 4-argument C_564B; TIL_4D is the projectile)
func C_568F(bp08, bp06, bp04 int) {
	if D_8742._npc._gtile[bp08] == TIL_80 || D_8742._npc._gtile[bp08] == TIL_82 {
		if bp06 != 0 {
			C_5500(bp08, bp06, bp04)
		} else {
			C_564B(bp08, bp06, bp04, TIL_4D)
		}
	} else {
		if bp04 != 0 {
			C_5500(bp08, bp06, bp04)
		} else {
			C_564B(bp08, bp06, bp04, TIL_4D)
		}
	}
}

func C_56D3(bp04 int) bool {
	// signed int compares: a negative offset counts as in range
	return int(D_8742._npc._x[bp04])-int(D_95A5.x)*16 >= 32 ||
		int(D_8742._npc._y[bp04])-int(D_95A5.y)*16 >= 32
}

// move NPCs[MOD_OUTDOORS]
func C_5712() {
	var loc_A, loc_B, loc_C, loc_D int

	for loc_F := 7; loc_F >= 0; loc_F-- {
		var loc_G int

		loc_H := int(D_8742._npc._tile[loc_F])
		if loc_H == 0 {
			continue
		}
		// is out of range?
		if C_56D3(loc_F) {
			D_8742._npc._tile[loc_F] = 0
			D_8742._npc._gtile[loc_F] = 0
			continue
		}
		if loc_H != TIL_8C && loc_H != TIL_8E {
			loc_G = C_535B(loc_F, &loc_A, &loc_B, &loc_C, &loc_D)
			if loc_G < 2 {
				// -- attacked by --
				t_callback()
				u4_puts(D_1FD8)
				u4_puts(C_1513(uint8(loc_H)))
				Gra_CR()
				C_7DFE(loc_F)
				continue
			}
		}
		if loc_H == TIL_80 {
			if loc_G < 4 {
				C_568F(loc_F, loc_A, loc_B)
			} else {
				C_5500(loc_F, loc_A, loc_B)
			}
			continue
		}
		if loc_H == TIL_88 || loc_H == TIL_E8 || loc_H >= TIL_F4 {
			if u4_abs(loc_C) < 5 && u4_abs(loc_D) < 5 && U4_RND1(1) != 0 {
				C_564B(loc_F, loc_A, loc_B, TIL_4F)
			}
		}
		if loc_H == TIL_8C || loc_H == TIL_8E || (C_0AFE(uint8(loc_H)) != 0 && U4_RND1(3) != 0) {
			y := isign(sc(u_rand_a()))
			x := isign(sc(u_rand_a()))
			C_5062(loc_F, x, y, 1)
		} else {
			C_5062(loc_F, loc_A, loc_B, 2)
		}
	}
}

// move npc?
func C_5834() {
	if CurMode == MOD_BUILDING {
		C_5293()
	} else if CurMode == MOD_OUTDOORS {
		C_5712()
	}
}

// spawn creatures?
func C_5851() {
	if CurMode != MOD_OUTDOORS || U4_RND1(15) != 0 {
		return
	}
	n := &D_8742._npc
	for si := 3; si >= 0; si-- {
		if n._tile[si] != 0 {
			continue
		}
		loc_A := uint8(U4_RND1(31))
		loc_C := uint8(U4_RND1(31))
		loc_B := D_8742.x32x32(int(loc_C), int(loc_A))
		if u4_abs(int(loc_A)-int(D_959C.x)) >= 6 && u4_abs(int(loc_C)-int(D_959C.y)) >= 6 {
			if *loc_B < TIL_02 { // at sea
				if U4_RND1(7) == 0 {
					n._x[si] = uint8(int(loc_A) + int(D_95A5.x)*16)
					n._y[si] = uint8(int(loc_C) + int(D_95A5.y)*16)
					n._gtile[si] = uint8(TIL_80 + U4_RND1(7)*2)
					n._tile[si] = n._gtile[si]
					if n._gtile[si] == TIL_82 {
						n._gtile[si] = TIL_80
						n._tile[si] = TIL_80
					}
				}
			} else if *loc_B >= TIL_04 && *loc_B < TIL_08 { // on earth
				var bp_0a int

				n._x[si] = uint8(int(loc_A) + int(D_95A5.x)*16)
				n._y[si] = uint8(int(loc_C) + int(D_95A5.y)*16)
				if Party._moves > 100000 {
					bp_0a = 15
				} else if Party._moves < 10000 {
					bp_0a = 3
				} else if Party._moves < 20000 {
					bp_0a = 7
				} else if Party._moves >= 30000 {
					bp_0a = 15
				} else {
					bp_0a = 7
				}
				n._gtile[si] = uint8(U4_RND1(U4_RND1(bp_0a))*4 + TIL_C0)
				n._tile[si] = n._gtile[si]
			}
		}
	}
}
