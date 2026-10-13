package u4

// Peering at a gem (the map of the town or dungeon level), getting chests (and their traps),
// firing the ship's cannon and lighting torches.

// ---- Peering at a gem ----

// How a town's tiles look on the gem's map: each a pattern of dots in a 4x4 square.
const (
	gemNothing = iota
	gemGrass
	gemForest
	gemLava
	gemWall // castles, bridges, walls
	gemBuilding
	gemShrine
	gemMountain
	gemHill
	gemScrub
	gemShallowWater
	gemWater
	gemDeepWater
)

// gemLook is the pattern of each of the first 128 tiles; people and monsters show as
// gemBuilding.
var gemLook = [128]uint8{
	12, 11, 10, 1, 1, 9, 2, 8, 7, 5, 5, 5, 5, 4, 4, 4,
	5, 5, 5, 5, 5, 5, 5, 4, 5, 4, 4, 5, 5, 6, 6, 5,
	6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6,
	7, 7, 7, 7, 7, 7, 7, 5, 5, 7, 6, 6, 5, 0, 3, 4,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 7, 3, 3, 3, 3, 3, 3,
	5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5,
	4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
	4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 6, 0, 7,
}

// gemDot is a dot of a pattern: its row and column in the 4x4 square.
type gemDot struct{ row, col int }

// gemPatterns are the patterns' colours and dots (deep water is drawn by itself).
var gemPatterns = map[uint8]struct {
	color int
	dots  []gemDot
}{
	gemGrass:        {1, []gemDot{{0, 1}, {1, 3}, {2, 1}, {3, 3}}},
	gemForest:       {1, []gemDot{{0, 3}, {1, 1}, {2, 3}, {3, 1}, {0, 1}, {1, 3}, {2, 1}, {3, 3}}},
	gemLava:         {2, []gemDot{{0, 0}, {1, 0}, {2, 0}, {3, 0}, {0, 2}, {1, 2}, {2, 2}, {3, 2}}},
	gemWall:         {3, []gemDot{{0, 0}, {0, 1}, {0, 2}, {0, 3}, {3, 0}, {3, 1}, {3, 2}, {3, 3}}},
	gemBuilding:     {3, []gemDot{{1, 1}, {2, 1}, {1, 2}, {2, 2}}},
	gemShrine:       {3, []gemDot{{1, 0}, {2, 0}, {1, 3}, {2, 3}, {0, 0}, {0, 1}, {0, 2}, {0, 3}, {3, 0}, {3, 1}, {3, 2}, {3, 3}}},
	gemMountain:     {3, []gemDot{{1, 1}, {2, 1}, {1, 2}, {2, 2}, {1, 0}, {2, 0}, {1, 3}, {2, 3}, {0, 0}, {0, 1}, {0, 2}, {0, 3}, {3, 0}, {3, 1}, {3, 2}, {3, 3}}},
	gemHill:         {3, []gemDot{{0, 0}, {1, 0}, {0, 1}, {1, 1}, {2, 2}, {3, 2}, {2, 3}, {3, 3}}},
	gemScrub:        {1, []gemDot{{0, 1}, {1, 1}, {2, 1}, {0, 3}, {2, 3}, {3, 3}}},
	gemShallowWater: {2, []gemDot{{0, 0}, {2, 0}, {1, 2}, {3, 2}}},
	gemWater:        {2, []gemDot{{0, 0}, {2, 2}}},
}

// gemPixel is the pixel of dot d of the town map's square (row, col): the map's 32x32
// squares start eight squares in.
func gemPixel(square, d int) int { return (square+8)*4 + d }

// showTownMap shows the town's map until a key, with the party's position blinking when
// withParty.
func showTownMap(withParty bool) {
	savedMode := gameMode
	gameMode = modeVision
	clearMap()
	for row := 0; row < townTiles; row++ {
		for col := 0; col < townTiles; col++ {
			dot := func(r, c, color int) { plot(gemPixel(row, r), gemPixel(col, c), color) }
			tile := *area.townTile(row, col)
			look := uint8(gemBuilding)
			if tile < tilePirateWest {
				look = gemLook[tile]
			}
			if look == gemDeepWater {
				dot(2*(col%2), 2, 3)
				continue
			}
			if p, ok := gemPatterns[look]; ok {
				for _, d := range p.dots {
					dot(d.row, d.col, p.color)
				}
			}
		}
	}
	flushKeys()
	y, x := int(windowPos.y), int(windowPos.x)
	for blink := 0; !keyWaiting(); blink++ {
		if withParty {
			for r := 0; r < 4; r++ {
				plotXOR(gemPixel(y, r), gemPixel(x, blink%4), 3)
			}
		}
	}
	readKey()
	gameMode = savedMode
}

// dungeonMapGlyphs are the font's characters for each kind of dungeon room on the gem's map.
var dungeonMapGlyphs = [16]uint8{32, 6, 5, 4, 36, 32, 32, 15, 84, 70, 94, 0, 14, 14, 2, 3}

// The dungeon map shows the level around the party on the map window's 22x22 characters.
const (
	gemMapSize   = 22
	gemMapCenter = 11
	avatarGlyph  = 1
	mapStackSize = 150 // the original's stack of rooms to look round, which wrapped
)

// showDungeonMap shows the rooms of the party's level that can be seen from it (through any
// but walls) until a key, the level wrapping round as often as fits.
func showDungeonMap() {
	savedMode := gameMode
	gameMode = modeVision
	clearMap()
	var seen [gemMapSize][gemMapSize]bool
	type room struct{ row, col, y, x uint8 }
	var stack [mapStackSize]room
	top := 0
	textRow, textCol = gemMapCenter, gemMapCenter
	stack[top] = room{gemMapCenter, gemMapCenter, party.y, party.x}
	top++
	seen[gemMapCenter-1][gemMapCenter-1] = true
	putGlyph(avatarGlyph)
	for {
		top--
		r := stack[top]
		for dx := -1; dx <= 1; dx++ {
			for dy := -1; dy <= 1; dy++ {
				if dx == 0 && dy == 0 {
					continue
				}
				row, col := int(r.row)+dy, int(r.col)+dx
				if row < 1 || row > gemMapSize || col < 1 || col > gemMapSize || seen[row-1][col-1] {
					continue
				}
				x, y := (int(r.x)+dx)&(levelSize-1), (int(r.y)+dy)&(levelSize-1)
				textRow, textCol = int16(row), int16(col)
				contents := dngTile(y, x)
				putGlyph(dungeonMapGlyphs[contents/16])
				if dngKind(contents) != dngWall {
					stack[top] = room{uint8(row), uint8(col), uint8(y), uint8(x)}
					top = (top + 1) % mapStackSize
					seen[row-1][col-1] = true
				}
			}
		}
		if top == 0 {
			break
		}
	}
	textRow, textCol = textBottom, textLeft
	flushKeys()
	readKey()
	gameMode = savedMode
	forceRender = 1
}

// peerAtGem shows the map of the dungeon level or the town.
func peerAtGem() {
	if party.loc >= locFirstDungeon {
		showDungeonMap()
	} else {
		showTownMap(true)
	}
}

func cmdPeer() {
	printText("Peer at ")
	if party.gems == 0 {
		sayWhat()
		return
	}
	printText("a Gem!\n")
	party.gems--
	peerAtGem()
}

// ---- Chests ----

func acidTrap(member int) {
	flashMember(member)
	damageMember(member, randomBelow(30))
}

const maxGold = 9999

// chestGold gives the party the chest's 10 to 96 gold.
func chestGold() {
	printText("The Chest Holds:\n")
	gold := randomBelow(80) + randomBits(7) + 10
	printNumber(int32(gold), 1, ' ')
	printText(" Gold\n")
	party.gold = uint16(min(int(party.gold)+gold, maxGold))
}

// openChest lets member (-1: by magic, who always evades) open a chest: half of them are
// trapped, which the nimble may evade.
func openChest(member int) {
	if trap := randomBits(3); trap%2 == 0 {
		// The trap is 0 or 2 masked with a random byte: Sleep and Bomb never come.
		trap &= int(randomByte())
		switch trap {
		case 0:
			printText("Acid")
		case 1:
			printText("Sleep")
		case 2:
			printText("Poison")
		default:
			printText("Bomb")
		}
		printText(" Trap!\n")
		switch {
		case member == -1 || randomBelow(100) <= int(party.members[member].dexterity)+25:
			printText("Evaded!\n")
			sound(soundFailure)
		case trap == 0:
			acidTrap(member)
		case trap == 1:
			party.members[member].status = statusAsleep
			if gameMode >= modeCombat {
				fighters.memberTile[member] = tileAsleep
			}
			flashMember(member)
		case trap == 2:
			party.members[member].status = statusPoisoned
			showStats()
			flashMember(member)
		default:
			damageParty()
		}
	}
	chestGold()
}

// getChest lets member open the chest the party stands on. Taking a town's chest is theft:
// honesty, justice and honor suffer.
func getChest(member int) {
	if gameMode == modeDungeon {
		if hereTile != dngChest {
			sayNotHere()
			return
		}
		*area.dungeonRoom(int(party.z), int(party.y), int(party.x)) = dngCorridor
	} else {
		if hereTile != tileChest {
			sayNotHere()
			return
		}
		if party.loc != 0 && *area.townTile(int(party.y), int(party.x)) == tileChest {
			*area.townTile(int(party.y), int(party.x)) = tileBrickFloor
			lowerKarma(&party.karmaHonesty, 1)
			lowerKarma(&party.karmaJustice, 1)
			lowerKarma(&party.karmaHonor, 1)
		} else {
			for i := len(area.people.tile) - 1; i >= 0; i-- {
				if area.people.tile[i] == tileChest && party.x == area.people.x[i] && party.y == area.people.y[i] {
					area.people.baseTile[i] = 0
					area.people.tile[i] = 0
					break
				}
			}
		}
	}
	openChest(member)
}

func cmdGet() {
	if party.flyingOrLight != 0 && party.tile == tileBalloon {
		printText("Get chest\n")
		sayDriftOnly()
		return
	}
	printText("Get Chest!\n")
	member := askMember("Who opens?" + answerRoom(2))
	if member < 0 {
		return
	}
	if !isConscious(member) {
		sayDisabled()
		return
	}
	getChest(member)
}

// getChestInCombat lets the active member open the chest on the battlefield they stand on.
func getChestInCombat() {
	a := int(activeMember)
	tile := combatTile(int(battlefield.memberY[a]), int(battlefield.memberX[a]))
	if *tile != tileChest {
		sayNotHere()
		return
	}
	*tile = tileTiledFloor
	if room := area.dungeonRoom(int(party.z), int(party.y), int(party.x)); *room == dngChest {
		*room = dngCorridor
	}
	openChest(a)
}

func cmdGetInCombat() {
	printText("Get Chest!\n")
	getChestInCombat()
}

// ---- The ship's cannon ----

const broadsidesOnly = "Broadsides Only!\n"

// cmdFire fires the ship's cannon broadside, up to three tiles: it sinks a ship or kills a
// person, a monster only one time in four; Lord British survives it.
func cmdFire() {
	printText("Fire ")
	if party.tile < tileShipWest || party.tile > tileShipSouth {
		sayWhat()
		return
	}
	printText("Cannon!\n")
	dx, dy := askDirection("Dir: ")
	if dx == 0 && dy == 0 {
		return
	}
	facesAlong := party.tile == tileShipNorth || party.tile == tileShipSouth
	if (dx != 0) != facesAlong {
		printText(broadsidesOnly)
		sound(soundBadCommand)
		return
	}
	sound(soundMissile)
	x, y := int(party.x)+dx, int(party.y)+dy
	hitX, hitY = uint8(viewCenter+dx), uint8(viewCenter+dy)
	hitTile = tileMissile
	target := -1
	for flown := 1; ; flown++ {
		if flown > 3 {
			hitTile = 0
			return
		}
		target = personAt(31, uint8(x), uint8(y))
		if target != -1 && area.people.tile[target] != tileWhirlpool && area.people.tile[target] != tileBalloon {
			break
		}
		drawView()
		hitX += uint8(dx)
		hitY += uint8(dy)
		x += dx
		y += dy
	}
	hitTile = tileHitFlash
	drawView()
	hitTile = 0
	sound(soundMonsterHit)
	if target < 8 && randomBits(3) != 0 {
		return
	}
	if area.people.tile[target] == tileLordBritish {
		return
	}
	area.people.baseTile[target] = 0
	area.people.tile[target] = 0
}

// ---- Torches ----

// torchTurns is how long a torch burns.
const torchTurns = 100

func cmdIgnite() {
	printText("Ignite Torch!\n")
	if gameMode != modeDungeon {
		sayNotHere()
		return
	}
	if party.torches == 0 {
		sayNoneLeft()
		return
	}
	party.torches--
	party.flyingOrLight += torchTurns
	lookAround()
}
