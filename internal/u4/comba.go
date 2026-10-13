package u4

// The battle: the party members' turns, their attacks, and the damage done to monsters.

// monsterAt is the monster at (x, y) on the battlefield, or -1.
func monsterAt(x, y uint8) int {
	for i := len(fighters.tile) - 1; i >= 0; i-- {
		if fighters.tile[i] != 0 && battlefield.monsterX[i] == x && battlefield.monsterY[i] == y {
			return i
		}
	}
	return -1
}

// partyWipedOut says whether no party member lives.
func partyWipedOut() bool {
	for i := int(party.size) - 1; i >= 0; i-- {
		if isAlive(i) {
			return false
		}
	}
	return true
}

// combatOver says whether the battle has ended: no party member is left on the field, or
// (but in a dungeon room, which the party must leave) no monster.
func combatOver() bool {
	membersLeft := false
	for i := 7; i >= 0; i-- {
		if fighters.memberTile[i] != 0 {
			membersLeft = true
			break
		}
	}
	if !membersLeft {
		return true
	}
	if gameMode == modeCombatRoom {
		return false
	}
	for i := len(fighters.tile) - 1; i >= 0; i-- {
		if fighters.tile[i] != 0 {
			return false
		}
	}
	return true
}

// fight runs the battle: each party member's command in turn, then the monsters', until it
// ends.
func fight() {
	var key, lastKey uint16
	flushKeys()
	fleeDirX, fleeDirY = 0, 0
	for {
		for activeMember = 0; uint16(activeMember) < party.size && !combatOver(); activeMember++ {
			a := int(activeMember)
			if fighters.memberTile[a] != 0 && isConscious(a) {
				cursorBlink = 4
				invertStatusLine(a)
				cursorBoxX = int16(battlefield.memberX[a])
				cursorBoxY = uint16(battlefield.memberY[a])
				printText(cstr(party.members[a].name[:]))
				printText(" with ")
				printText(weaponNames[int(party.members[a].weapon)])
				newLine()
				printChar(promptChar)
				cursorBlink = 2
				drawCombat()
				ask(pCmd)
				const commandSeconds = 25 // a member who waits this long passes
				waitSeconds(commandSeconds, true)
				key = keySpace
				if keyWaiting() {
					key = uint16(readKey())
				}
				key = uint16(lowerKey(int(key)))
				switch key {
				case keySpace:
					sayPass()
				case keyUpAlt, keyUp:
					cmdNorthInCombat()
				case keyLeftAlt, keyLeft:
					cmdWestInCombat()
				case keyDownAlt, keyDown:
					cmdSouthInCombat()
				case keyRightAlt, keyRight:
					cmdEastInCombat()
				case keyA:
					cmdAttackInCombat()
				case keyC:
					cmdCast()
				case keyG:
					cmdGetInCombat()
				case keyR:
					cmdReady()
				case keyU:
					cmdUse()
				case keyV:
					cmdVolume()
				case keyZ:
					cmdZtats()
				case keyB, keyD, keyE, keyF, keyH, keyI, keyJ, keyK, keyL, keyM, keyN,
					keyO, keyP, keyQ, keyS, keyT, keyW, keyX, keyY:
					sayNotHere()
				case keyCtrlS:
					if lastKey == keyAltZ {
						showKarmas()
						break
					}
					fallthrough
				default:
					printText("Bad command\n")
					sound(soundBadCommand)
					invertStatusLine(a)
					activeMember-- // the same member again
					continue
				}
				invertStatusLine(a)
				cursorBoxX = -1
				showStats()
			}
			if !turnPasses {
				turnPasses = true
				activeMember-- // the command took no time: the same member again
			} else {
				endMemberTurn()
			}
		}
		monstersTurn()
		lastKey = key
		if combatOver() {
			break
		}
	}
	if !partyWipedOut() {
		endCombat()
	}
	turnPasses = true
}

// endMemberTurn ends the active member's turn: a random event, Quickness's second turn, a
// sleeper waking (one time in eight), and a ration eaten.
func endMemberTurn() {
	a := int(activeMember)
	if fighters.memberTile[a] == 0 {
		return
	}
	if isConscious(a) {
		groundEvent()
	}
	member := &party.members[a]
	switch {
	case !isAlive(a):
		fighters.memberTile[a] = 0
	case isConscious(a) && activeSpell == effectQuickness && randomBits(1) != 0:
		activeMember-- // another turn
		return
	case member.status == statusAsleep && randomBits(7) == 0:
		member.status = statusGood
		fighters.memberTile[a] = uint8(memberShape(a))
		memberFrame[a] = fighters.memberTile[a]
		showStats()
	}
	eatFood(1)
}

// damageMonster hurts monster by damage (taken as a byte), dealt by member (-1: none, who
// then earns experience for a kill), and tells how it fares. Lord British can't be hurt.
func damageMonster(monster, member, damage int) {
	hurt := uint8(damage)
	printText(creatureName(fighters.tile[monster]))
	printChar(' ')
	killed := false
	if fighters.tile[monster] != tileLordBritish {
		killed = fighters.hp[monster] < hurt
		fighters.hp[monster] -= hurt
	}
	strength := int(monsterStrength[monsterIndex(fighters.tile[monster])])
	if killed {
		printText("Killed!\n")
		if member != -1 {
			xp := strength/16 + 1
			addXP(member, xp)
			printText("Exp. ")
			printNumber(int32(xp), 1, ' ')
			newLine()
		}
		fighters.tile[monster] = 0
		return
	}
	half, quarter := strength/2, strength/4
	hp := int(fighters.hp[monster])
	switch {
	case hp < fleeingHP:
		printText("Fleeing!\n")
	case hp < quarter:
		printText("Critical!\n")
	default:
		newLine()
		switch {
		case hp < half:
			printText("Heavily ")
		case hp < half+quarter:
			printText("Lightly ")
		default:
			printText("Barely ")
		}
		printText("Wounded!\n")
	}
}

// advanceHit moves the missile (hitX, hitY) on by (dx, dy), unless that leaves the field or
// meets something that stops missiles; it reports whether it moved.
func advanceHit(dx, dy uint8) bool {
	x, y := hitX+dx, hitY+dy
	if x >= viewSize || y >= viewSize {
		return false
	}
	tile := *combatTile(int(y), int(x))
	if tile == tileSecretDoor {
		return false
	}
	// Water, shores, the barrier and monsters let missiles pass; other tiles if walkable.
	isTerrain := tile <= tilePirateWest && tile > tileShallowWater && tile != tileBarrier
	isShore := tile >= tileShoreSouthWest && tile <= tileShoreNorthEast
	if isTerrain && !isShore && !isWalkable(tile) {
		return false
	}
	hitX, hitY = x, y
	return true
}

// returnAxe flies the magic axe back to its thrower, over reach tiles.
func returnAxe(reach int) {
	if party.members[activeMember].weapon != weaponMagicAxe || reach < 2 {
		return
	}
	dx := signOf(int16(battlefield.memberX[activeMember]) - int16(hitX))
	dy := signOf(int16(battlefield.memberY[activeMember]) - int16(hitY))
	hitTile = tileMissile
	for ; reach != 0; reach-- {
		advanceHit(uint8(dx), uint8(dy))
		drawCombat()
	}
	hitTile = 0
}

// missed reports a miss (reach tiles out); burning oil sets a fire where it lands.
func missed(reach int) {
	if reach == 0 {
		sound(soundMiss)
	}
	hitTile = 0
	printText("Missed!\n")
	if party.members[activeMember].weapon == weaponOil && hitX < viewSize && hitY < viewSize && *combatTile(int(hitY), int(hitX)) >= tileSwamp {
		*combatTile(int(hitY), int(hitX)) = tileFireField
	}
	drawCombat()
	returnAxe(reach)
}

// tryHit is the active member's blow at monster, reach tiles away: it may miss (the clumsy
// do; and in the Abyss only magic weapons hit), else does damage up to the member's
// strength plus the weapon's.
func tryHit(monster, reach int) {
	member := &party.members[activeMember]
	if party.loc == locAbyss && member.weapon < weaponMagicAxe {
		missed(reach)
		return
	}
	if member.weapon == weaponMagicWand {
		hitTile = tileMagicSphere
	} else {
		hitTile = tileHitFlash
	}
	if member.dexterity < 40 && uint16(randomBits(255)) > member.dexterity+128 {
		missed(reach)
		return
	}
	under := combatTile(int(battlefield.monsterY[monster]), int(battlefield.monsterX[monster]))
	if member.weapon == weaponOil && *under >= tileSwamp {
		*under = tileFireField
	}
	most := int(int16(int(member.strength) + int(weaponDamage[member.weapon])))
	damage := safeMod(randomBits(255), min(most, 255))
	drawCombat()
	sound(soundMonsterHit)
	hitTile = 0
	damageMonster(monster, int(activeMember), damage)
	returnAxe(reach)
}

// shoot fires the active member's missile weapon toward (dx, dy): up to 11 tiles, oil as far
// as asked. Daggers and oil are thrown, and used up.
func shoot(dx, dy int) {
	member := &party.members[activeMember]
	reach := viewSize
	if member.weapon == weaponOil {
		reach = askLetter("Range:"+answerRoom(2), '0', '9') - '0'
		if reach < 0 {
			return
		}
	}
	if member.weapon == weaponDagger || member.weapon == weaponOil {
		if party.weapons[member.weapon] == 0 {
			member.weapon = weaponHands
			printText("Last one!\n")
		} else {
			party.weapons[member.weapon]--
		}
	}
	if member.weapon == weaponMagicWand {
		hitTile = tileMagicSphere
	} else {
		hitTile = tileMissile
	}
	sound(soundMiss)
	flown := 0
	target := 0 // (oil thrown 0 tiles hits monster 0, as the original's unset variable did)
	for ; reach != 0; reach-- {
		if !advanceHit(uint8(dx), uint8(dy)) {
			missed(flown)
			return
		}
		flown++
		drawCombat()
		if target = monsterAt(hitX, hitY); target != -1 {
			break
		}
	}
	hitTile = 0
	if target == -1 {
		missed(flown)
		return
	}
	tryHit(target, flown)
}

// cmdAttackInCombat attacks in a direction: a melee weapon the next tile (a halberd the one
// beyond), a missile weapon, or a dagger thrown when nothing is near.
func cmdAttackInCombat() {
	member := &party.members[activeMember]
	dx, dy := askDirection("Dir: ")
	if dx == 0 && dy == 0 {
		return
	}
	hitX, hitY = battlefield.memberX[activeMember], battlefield.memberY[activeMember]
	if weaponRanged[member.weapon] {
		shoot(dx, dy)
		return
	}
	if member.weapon == weaponHalberd {
		hitX += uint8(dx)
		hitY += uint8(dy)
	}
	if advanceHit(uint8(dx), uint8(dy)) {
		hitTile = tileMissile
		drawCombat()
		hitTile = 0
		sound(soundMiss)
		if monster := monsterAt(hitX, hitY); monster != -1 {
			tryHit(monster, 1)
			return
		}
		if member.weapon == weaponDagger {
			shoot(dx, dy)
			return
		}
	}
	missed(0)
}

// cmdAttack attacks whoever is in a direction. In a town that sets the guards and Lord
// British on the party; attacking there, or a good creature, costs compassion, justice and
// honor.
func cmdAttack() {
	if gameMode <= modeTown && party.flyingOrLight != 0 {
		printText("Attack\n")
		sayDriftOnly()
		return
	}
	dx, dy := askDirection("Attack: ")
	if dx == 0 && dy == 0 {
		return
	}
	target := objectAt(uint8(int(party.x)+dx), uint8(int(party.y)+dy))
	if target == -1 {
		printText("Nothing to Attack!\n")
		return
	}
	switch area.people.tile[target] {
	case tileWhirlpool, tileShallowWater, tileTwister:
		printText("Nothing to Attack!\n")
		return
	}
	if gameMode == modeTown {
		for i := len(area.people.tile) - 1; i >= 0; i-- {
			if area.people.tile[i] == tileGuard || area.people.tile[i] == tileLordBritish {
				area.people.behaviour[i] = behaviourAttack
				area.people.talkIndex[i] = 0
			}
		}
	}
	if gameMode == modeTown || isGoodCreature(area.people.tile[target]) {
		lowerKarma(&party.karmaCompassion, 5)
		lowerKarma(&party.karmaJustice, 5)
		lowerKarma(&party.karmaHonor, 5)
	}
	startCombat(target)
}
