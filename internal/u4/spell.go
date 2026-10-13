package u4

// Casting spells.

// The spells, A to Z.
const (
	spellAwaken = iota
	spellBlink
	spellCure
	spellDispel
	spellEnergy
	spellFireball
	spellGate
	spellHeal
	spellIceball
	spellJinx
	spellKill
	spellLight
	spellMagicMissile
	spellNegate
	spellOpen
	spellProtection
	spellQuickness
	spellResurrect
	spellSleep
	spellTremor
	spellUndead
	spellView
	spellWinds
	spellXit
	spellYup
	spellZdown
	spellCount
)

// spellCost is the magic points each spell takes.
var spellCost = [spellCount]uint8{
	spellAwaken:       5,
	spellBlink:        15,
	spellCure:         5,
	spellDispel:       20,
	spellEnergy:       10,
	spellFireball:     15,
	spellGate:         40,
	spellHeal:         10,
	spellIceball:      20,
	spellJinx:         30,
	spellKill:         25,
	spellLight:        5,
	spellMagicMissile: 5,
	spellNegate:       20,
	spellOpen:         5,
	spellProtection:   15,
	spellQuickness:    20,
	spellResurrect:    45,
	spellSleep:        15,
	spellTremor:       30,
	spellUndead:       15,
	spellView:         15,
	spellWinds:        10,
	spellXit:          15,
	spellYup:          10,
	spellZdown:        5,
}

// castingSpell is the spell being cast.
var castingSpell int

// castPitch is added to the spell's number for the pitch of its casting sound.
const castPitch = 96

// Lasting spells (jinx, negate, protection, quickness) hold this many turns; light lasts
// lightTurns.
const (
	lastingTurns = 10
	lightTurns   = 100
)

// noTile is what an off-map look finds: never a field, never walkable.
var noTile uint8

// mapTileFlat is the town or world window's tile at (y, x) as the original found it, by a
// flat index into the 32x32 map: a step off the left or right edge lands on the row above
// or below. Off the top or bottom it read memory that is never a field; that reads noTile.
func mapTileFlat(y, x int) *uint8 {
	i := y*townTiles + x
	if i < 0 || i >= len(area.tiles) {
		noTile = 0
		return &noTile
	}
	return &area.tiles[i]
}

// combatTileFlat is the battlefield's tile at (y, x) by a flat index, the same way.
func combatTileFlat(y, x int) *uint8 {
	if y*viewSize+x < 0 {
		noTile = 0
		return &noTile
	}
	return combatTile(y, x)
}

// isUndead says whether the monster tile is one of the undead.
func isUndead(tile uint8) bool {
	return tile == tileGhost || tile == tilePhantom || tile == tileSkeleton || tile == tileLiche
}

func spellFailed() {
	printText("Failed!\n")
	hitTile = 0
	sound(soundFailure)
}

// castIt spends the spell's magic points, with its sound and flash; false (and failure)
// under a negate spell.
func castIt() bool {
	cost := spellCost[castingSpell]
	party.members[activeMember].mp -= uint16(cost)
	sound(soundMagic, int(cost))
	invertMap()
	sound(soundSweep, castingSpell+castPitch)
	invertMap()
	if activeSpell == effectNegate {
		spellFailed()
		return false
	}
	return true
}

// onlyIn fails the spell with complaint unless where holds.
func onlyIn(where bool, complaint string) bool {
	if !where {
		printText(complaint)
		spellFailed()
	}
	return where
}

func outdoorsOnly() bool { return onlyIn(gameMode == modeOutdoors, "Outdoors Only!\n") }
func combatOnly() bool   { return onlyIn(gameMode >= modeCombat, "Combat Only!\n") }
func dungeonOnly() bool  { return onlyIn(gameMode == modeDungeon, "Dungeon Only!\n") }

// The damage of the missile spells: a random part, with bits always set.
func missileDamage(tile uint16) int {
	switch tile {
	case tileMissile:
		return randomBelow(64) | 16
	case tileMagicSphere:
		return randomBelow(224) | 32
	case tileHitFlash:
		return randomBelow(128) | 24
	default: // kill
		return 232
	}
}

// castMissile flies a spell's missile from the active member in a direction until it hits
// a monster.
func castMissile(tile uint16) {
	if !combatOnly() {
		return
	}
	dx, dy := askDirection("Dir: ")
	if dx|dy == 0 || !castIt() {
		return
	}
	hitTile = tile
	hitX = battlefield.memberX[activeMember]
	hitY = battlefield.memberY[activeMember]
	monster := -1
	for advanceHit(uint8(dx), uint8(dy)) {
		drawCombat()
		if monster = monsterAt(hitX, hitY); monster != -1 {
			break
		}
	}
	if monster == -1 {
		spellFailed()
		return
	}
	sound(soundMonsterHit)
	damage := missileDamage(hitTile)
	hitTile = 0
	damageMonster(monster, int(activeMember), damage)
}

// changeStatus sets a member's status from one to another, failing if it was not from.
func changeStatus(member int, from, to uint8) bool {
	if party.members[member].status != from {
		spellFailed()
		return false
	}
	party.members[member].status = to
	return true
}

func castAwaken() {
	m := askMember("Who:" + answerRoom(2))
	if m < 0 || !castIt() || !changeStatus(m, statusAsleep, statusGood) {
		return
	}
	if gameMode >= modeCombat {
		fighters.memberTile[m] = uint8(memberShape(m))
		memberFrame[m] = fighters.memberTile[m]
	}
}

// blinkCorner: the original's test lets no blink start where x and y share both top bits
// (the south-east corner of the world).
const blinkCorner = 192

func castBlink() {
	if party.tile < tileHorseWest || party.tile == tileBalloon { // only on foot or horse
		spellFailed()
		return
	}
	if !outdoorsOnly() {
		return
	}
	dx, dy := askDirection("Dir: ")
	if dx|dy == 0 || !castIt() {
		return
	}
	if party.x&party.y >= blinkCorner {
		spellFailed()
		return
	}
	// Out to the edge of the window, then back to the farthest walkable place.
	x, y := int(windowPos.x), int(windowPos.y)
	for x >= 0 && x < windowTiles && y >= 0 && y < windowTiles {
		x += dx
		y += dy
	}
	for {
		x -= dx
		y -= dy
		if x == int(windowPos.x) && y == int(windowPos.y) || isWalkable(*area.townTile(y, x)) {
			break
		}
	}
	if x == int(windowPos.x) && y == int(windowPos.y) {
		spellFailed()
		return
	}
	party.x += uint8(x) - windowPos.x
	party.y += uint8(y) - windowPos.y
	centerMap()
}

func castCure() {
	m := askMember("Who:" + answerRoom(2))
	if m >= 0 && castIt() {
		changeStatus(m, statusPoisoned, statusGood)
	}
}

// roomAhead is the dungeon room in front of the party.
func roomAhead() *uint8 {
	return area.dungeonRoom(int(party.z), aheadY(int(party.y), int(party.dir)), aheadX(int(party.x), int(party.dir)))
}

func isField(tile uint8) bool { return tile >= tilePoisonField && tile <= tileSleepField }

// memberSpot is the active member's place on the battlefield.
func memberSpot() (x, y int) {
	return int(battlefield.memberX[activeMember]), int(battlefield.memberY[activeMember])
}

func castDispel() {
	if gameMode == modeDungeon {
		if !castIt() {
			return
		}
		if room := roomAhead(); dngKind(*room) == dngField {
			*room = dngCorridor
		} else {
			spellFailed()
		}
		return
	}
	if gameMode <= modeTown && party.flyingOrLight != 0 {
		spellFailed()
		return
	}
	dx, dy := askDirection("Dir: ")
	if dx|dy == 0 || !castIt() {
		return
	}
	var target *uint8
	ground := hereTile
	switch {
	case gameMode >= modeCombat:
		x, y := memberSpot()
		if y+dy >= viewSize || x+dx >= viewSize {
			spellFailed()
			return
		}
		target = combatTileFlat(y+dy, x+dx)
		ground = *combatTile(y, x)
	case gameMode == modeOutdoors:
		target = mapTileFlat(int(windowPos.y)+dy, int(windowPos.x)+dx)
	default:
		if int(party.y)+dy >= townTiles || int(party.x)+dx >= townTiles {
			spellFailed()
			return
		}
		target = mapTileFlat(int(party.y)+dy, int(party.x)+dx)
	}
	if !isField(*target) {
		spellFailed()
		return
	}
	*target = ground // the field gives way to the ground the caster stands on
}

func castEnergy() {
	printText("Energy type? ")
	askKeys("Energy type", item("Poison", key('p')), item("Lightning", key('l')), item("Fire", key('f')), item("Sleep", key('s')))
	letter := keyChar(readKey())
	if letter >= 'a' && letter <= 'z' {
		letter -= 'a' - 'A'
	}
	var field uint8
	switch letter {
	case 'P':
		field = tilePoisonField
	case 'L':
		field = tileEnergyField
	case 'F':
		field = tileFireField
	case 'S':
		field = tileSleepField
	default:
		spellFailed()
		return
	}
	printChar(letter)
	newLine()
	if gameMode == modeDungeon {
		if !castIt() {
			return
		}
		room := roomAhead()
		if *room != dngCorridor {
			spellFailed() // quirk: the field is laid anyway
		}
		*room = dngField + field - tilePoisonField
		return
	}
	if gameMode >= modeCombat {
		dx, dy := askDirection("Dir: ")
		if dx|dy == 0 || !castIt() {
			return
		}
		x, y := memberSpot()
		if y+dy < viewSize && x+dx < viewSize {
			if target := combatTileFlat(y+dy, x+dx); isWalkable(*target) {
				*target = field
				return
			}
		}
	}
	spellFailed()
}

func castFireball() { castMissile(tileHitFlash) }

func castGate() {
	if party.tile < tileHorseWest || party.tile == tileBalloon {
		spellFailed()
		return
	}
	if !outdoorsOnly() {
		return
	}
	phase := askLetter("To Phase:"+answerRoom(2), '0', '8')
	if phase < 0 || phase == '0' || !castIt() {
		return
	}
	gate := phase - '1'
	party.x = moongateX[gate]
	party.y = moongateY[gate]
	centerMap()
}

func castHeal() {
	m := askMember("Who?" + answerRoom(2))
	if m < 0 || !castIt() {
		return
	}
	if !isAlive(m) {
		spellFailed()
		return
	}
	healMember(m, randomBelow(25)+75)
}

func castIceball() { castMissile(tileMagicSphere) }

// castLasting starts a spell that holds some turns.
func castLasting(spell uint8) {
	if castIt() {
		activeSpell = spell
		spellTurns = lastingTurns
	}
}

func castJinx() { castLasting(effectJinx) }
func castKill() { castMissile(tileWhirlpool) }

func castLight() {
	if !dungeonOnly() || !castIt() {
		return
	}
	party.flyingOrLight += lightTurns
	if gameMode == modeDungeon {
		lookAround()
	}
}

func castMagicMissile() { castMissile(tileMissile) }
func castNegate()       { castLasting(effectNegate) }

func castOpen() {
	if !castIt() {
		return
	}
	if gameMode == modeOutdoors && party.tile == tileBalloon {
		spellFailed()
		return
	}
	if gameMode >= modeCombat {
		getChestInCombat()
	} else {
		getChest(-1) // a sure escape from the trap
	}
}

func castProtection() { castLasting(effectProtection) }
func castQuickness()  { castLasting(effectQuickness) }

func castResurrect() {
	if gameMode >= modeCombat {
		spellFailed()
		return
	}
	m := askMember("Who:" + answerRoom(2))
	if m >= 0 && castIt() {
		changeStatus(m, statusDead, statusGood)
	}
}

// eachMonster calls do for each monster slot, last first.
func eachMonster(do func(m int)) {
	for m := len(fighters.tile) - 1; m >= 0; m-- {
		do(m)
	}
}

func castSleep() {
	if !combatOnly() || !castIt() {
		return
	}
	eachMonster(func(m int) {
		t := fighters.tile[m]
		if t != 0 && !isUndead(t) && t != tileBalron && fighters.hp[m] < randomByte() {
			fighters.sleeping[m] = 1
		}
	})
}

// The undead spell and the tremor leave a monster this many hit points (it flees below
// fleeingHP).
const shakenHP = 23

// tremorSpares are the hit points from which a monster stands the tremor.
const tremorSpares = 192

func castTremor() {
	if !combatOnly() || !castIt() {
		return
	}
	shakeScreen()
	hitTile = tileHitFlash
	eachMonster(func(m int) {
		if fighters.tile[m] == 0 || fighters.hp[m] >= tremorSpares {
			return
		}
		if randomBits(1) != 0 { // missed; half of these are shaken
			if randomBits(1) != 0 {
				fighters.hp[m] = shakenHP
			}
			return
		}
		hitX = battlefield.monsterX[m]
		hitY = battlefield.monsterY[m]
		drawCombat()
		sound(soundMonsterHit)
		damageMonster(m, int(activeMember), 255)
	})
	hitTile = 0
}

func castUndead() {
	if !combatOnly() || !castIt() {
		return
	}
	eachMonster(func(m int) {
		if isUndead(fighters.tile[m]) && randomBits(1) != 0 && fighters.hp[m] > shakenHP {
			fighters.hp[m] = shakenHP
		}
	})
}

func castView() {
	if castIt() {
		peerAtGem()
	}
}

func castWinds() {
	if !outdoorsOnly() {
		return
	}
	dx, dy := askDirection("From Dir: ")
	if dx|dy == 0 || !castIt() {
		return
	}
	switch {
	case dx == 1:
		windDir = dirEast
	case dx != 0:
		windDir = dirWest
	case dy == 1:
		windDir = dirSouth
	default:
		windDir = dirNorth
	}
}

func castXit() {
	if dungeonOnly() && castIt() {
		party.z = -1 // out of the dungeon
	}
}

// landTries is how many random rooms the Y-up and Z-down spells try for an empty one.
const landTries = 32

// landOnLevel moves the party to an empty room of the level it is on; false if none was
// found.
func landOnLevel() bool {
	var x, y uint8
	tries := landTries
	for ; tries > 0; tries-- {
		x = uint8(randomBits(levelSize - 1))
		y = uint8(randomBits(levelSize - 1))
		if *area.dungeonRoom(int(party.z), int(y), int(x)) == dngCorridor {
			break
		}
	}
	if tries == 0 {
		return false
	}
	party.x, party.y = x, y
	return true
}

func castYup() {
	if !dungeonOnly() || !castIt() {
		return
	}
	if party.loc == locAbyss {
		spellFailed()
		return
	}
	party.z--
	if party.z < 0 {
		return // up out of the dungeon
	}
	if !landOnLevel() {
		party.z++
		spellFailed()
	}
}

func castZdown() {
	if !dungeonOnly() || !castIt() {
		return
	}
	if party.loc == locAbyss || party.z == dungeonLevels-1 {
		spellFailed()
		return
	}
	party.z++
	if !landOnLevel() {
		party.z--
		spellFailed()
	}
}

var spells = [spellCount]func(){
	spellAwaken:       castAwaken,
	spellBlink:        castBlink,
	spellCure:         castCure,
	spellDispel:       castDispel,
	spellEnergy:       castEnergy,
	spellFireball:     castFireball,
	spellGate:         castGate,
	spellHeal:         castHeal,
	spellIceball:      castIceball,
	spellJinx:         castJinx,
	spellKill:         castKill,
	spellLight:        castLight,
	spellMagicMissile: castMagicMissile,
	spellNegate:       castNegate,
	spellOpen:         castOpen,
	spellProtection:   castProtection,
	spellQuickness:    castQuickness,
	spellResurrect:    castResurrect,
	spellSleep:        castSleep,
	spellTremor:       castTremor,
	spellUndead:       castUndead,
	spellView:         castView,
	spellWinds:        castWinds,
	spellXit:          castXit,
	spellYup:          castYup,
	spellZdown:        castZdown,
}

func cmdCast() {
	printText("Cast Spell!\n")
	if gameMode < modeCombat {
		activeMember = int16(askMember("Player:" + answerRoom(2)))
		if activeMember < 0 {
			return
		}
		if !isConscious(int(activeMember)) {
			sayDisabled()
			return
		}
	}
	clearStatus()
	showMixtures()
	textRow, textCol = promptRow, promptCol
	castingSpell = askLetter("Spell:"+answerRoom(2), 'A', 'Z')
	afterChoosing()
	if castingSpell < 0 {
		return
	}
	castingSpell -= 'A'
	printText(spellNames[castingSpell])
	printText("!\n")
	if party.mixtures[castingSpell] == 0 {
		sayNoneLeft()
		return
	}
	party.mixtures[castingSpell]--
	if party.members[activeMember].mp < uint16(spellCost[castingSpell]) {
		printText("M.P. too low!\n")
		spellFailed()
		return
	}
	spells[castingSpell]()
	showStats()
}
