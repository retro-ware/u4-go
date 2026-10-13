package u4

import (
	"strings"

	"github.com/memmaker/u4-go/internal/music"
)

// Helpers shared by the whole game: the party's numbers, the text window, prompts.

const (
	maxXP    = 9999
	maxKarma = 99
	maxMP    = 99
)

// shakeScreen shakes the map window four times, with noise.
func shakeScreen() {
	for range 4 {
		shakeUp()
		shakeDown()
	}
}

// addXP gives member experience, up to maxXP.
func addXP(member, amount int) {
	m := &party.members[member]
	m.xp += uint16(amount)
	if m.xp > maxXP {
		m.xp = maxXP
	}
}

func sayPass() {
	printText("Pass\n")
}

// healMember gives member hit points, up to their maximum.
func healMember(member, amount int) {
	m := &party.members[member]
	m.hp += uint16(amount)
	if m.hp > m.maxHP {
		m.hp = m.maxHP
	}
}

// flashMember flashes member's status line, with a sound.
func flashMember(member int) {
	invertStatusLine(member)
	sound(soundMemberHit)
	invertStatusLine(member)
}

// raiseKarma raises a virtue's karma by amount, up to maxKarma. A karma of 0 is a virtue
// the Avatar has mastered (an eighth of Avatarhood); it stays there.
func raiseKarma(karma *uint16, amount int) {
	if *karma == 0 {
		return
	}
	*karma += uint16(amount)
	if *karma > maxKarma {
		*karma = maxKarma
	}
}

// lowerKarma lowers a virtue's karma by amount, to no less than 1. A mastered virtue is
// lost first: its karma starts again from maxKarma.
func lowerKarma(karma *uint16, amount int) {
	if *karma == 0 {
		*karma = maxKarma
		updateAnkh()
		textCol = statsSpellColumn
		textRow = statsRowFood
		printChar(int(activeSpell))
		printText("\n Thou hast lost\n  an Eighth!\n")
	}
	*karma -= uint16(amount)
	if *karma == 0 {
		*karma = 1
	}
}

// personAt is the index of the last person (from last down to 0) standing at (x, y) on the
// current map, or -1.
func personAt(last int, x, y uint8) int {
	p := &area.people
	i := last
	for ; i >= 0; i-- {
		if p.tile[i] != 0 && p.x[i] == x && p.y[i] == y {
			break
		}
	}
	return i
}

// objectAt is the index of the person or creature at (x, y): outdoors only the first 8
// entries move about, in towns all 32.
func objectAt(x, y uint8) int {
	if party.loc == 0 {
		return personAt(7, x, y)
	}
	return personAt(31, x, y)
}

// safeMod is a % b in 16-bit arithmetic, 0 for b == 0.
func safeMod(a, b int) int {
	if b == 0 {
		return 0
	}
	return int(int16(a) % int16(b))
}

// memberShape is the tile that shows member: their class, or asleep.
func memberShape(member int) int {
	if party.members[member].status == statusAsleep {
		return tileAsleep
	}
	return int(party.members[member].class)*2 + tileMage
}

// isGoodCreature says whether a creature is no evil one, so killing it costs karma: people,
// and the animals.
func isGoodCreature(tile uint8) bool {
	switch tile {
	case tileSeahorse, tileRat, tileBat, tileSpider, tileInsects, tilePython:
		return true
	}
	return tile < tilePirateWest
}

// printText prints s in the text window, wrapping before a word that would not fit. After a
// glyphBlank is the font's blank glyph: blank like a space, but not ending a word.
const glyphBlank = 18

// answerRoom follows a prompt with room for an answer of n characters: blank glyphs, which
// keep the prompt and its room together on one line and clear the room, then back to just
// after the first.
func answerRoom(n int) string {
	return strings.Repeat(string(rune(glyphBlank)), n) + strings.Repeat("\b", n-1)
}

// page of lines it waits for a key. It returns the number of characters printed.
func printText(s string) int {
	lines := 0    // lines printed since the last wait
	wordLeft := 0 // characters until the next word starts
	nextLine := func() {
		if lines != textPageLines {
			lines++
			return
		}
		flushKeys()
		if textCol == textLastColumn {
			for !keyWaiting() {
			}
		}
		readKey()
		lines = 0
	}
	i := 0
	for ; i < len(s) && s[i] != 0; i++ {
		if wordLeft == 0 {
			for wordLeft = 0; i+wordLeft < len(s); wordLeft++ {
				if c := s[i+wordLeft]; c == 0 || c == '\n' || c == ' ' {
					break
				}
			}
			if wordLeft+int(textCol) > textWidth && textCol != textLeft {
				nextLine()
				newLine()
			}
		} else {
			wordLeft--
		}
		if s[i] == '\n' {
			nextLine()
		}
		printChar(int(s[i]))
	}
	return i
}

// printNumber prints n right-aligned in width characters, padded with pad.
func printNumber(n int32, width int, pad int) {
	if n < 0 {
		printChar('-')
		printNumber(-n, width-1, pad)
	}
	switch {
	case n >= 10:
		printNumber(n/10, width-1, pad)
		printChar(int(n%10 + '0'))
	case width > 1:
		printChar(pad)
		printNumber(n, width-1, pad)
	default:
		printChar(int(n + '0'))
	}
}

// printChar prints c at the text cursor: backspace, newline and wrapping included.
func printChar(c int) {
	histPut(byte(c))
	if c == promptChar && textRow != 0 {
		musicPrompt()
	}
	ch := uint8(c)
	switch ch {
	case '\b':
		if textCol > 1 {
			textCol--
			putGlyph(' ')
		}
	case '\n':
		newLine()
	case ' ':
		if textCol <= textLastColumn {
			putGlyph(ch)
			textCol++
		}
	default:
		if textCol > textLastColumn {
			newLine()
		}
		putGlyph(ch)
		textCol++
	}
}

// showStats shows the party's status lines: each member's name, hit points and status,
// then the food, the active spell, and the ship's hull or the gold.
func showStats() {
	const nameLength = 8
	updateAnkh()
	y, x := textRow, textCol
	for i := 0; i < int(party.size); i++ {
		m := &party.members[i]
		textRow = int16(i + 1)
		textCol = textLeft
		printChar(i + '1')
		printChar('-')
		for c := 0; c < nameLength && m.name[c] != 0; c++ {
			printChar(int(m.name[c]))
		}
		textCol = statsHPColumn
		printNumber(int32(m.hp), 3, ' ')
		printChar(int(m.status))
	}
	textCol = textLeft
	textRow = statsRowFood
	printText("F:")
	printNumber(party.food/100, 4, '0')
	printChar(' ')
	printChar(int(activeSpell))
	if party.tile < tileHorseWest {
		printText(" SHP:")
		printNumber(int32(party.hull), 2, '0')
	} else {
		printText(" G:")
		printNumber(int32(party.gold), 4, '0')
	}
	textRow, textCol = y, x
}

// showGold updates the gold on the status lines.
func showGold() {
	x, y := textCol, textRow
	textCol = statsGoldColumn
	textRow = statsRowFood
	printNumber(int32(party.gold), 4, '0')
	textRow, textCol = y, x
}

// isAlive says whether member is in the party and alive (good, poisoned or asleep).
func isAlive(member int) bool {
	if member >= int(party.size) {
		return false
	}
	switch party.members[member].status {
	case statusGood, statusPoisoned, statusAsleep:
		return true
	}
	return false
}

// isConscious says whether member is in the party and can act (good or poisoned).
func isConscious(member int) bool {
	if member >= int(party.size) {
		return false
	}
	switch party.members[member].status {
	case statusGood, statusPoisoned:
		return true
	}
	return false
}

// partyDies is the party's death: Lord British brings everyone back in his castle, with
// full hit points, some food and gold, and no weapons or armour. It does not return.
func partyDies() {
	music.Play(music.Stop)
	for i := int(party.size) - 1; i >= 0; i-- {
		party.members[i].status = statusDead
		party.members[i].hp = 0
	}
	clearStatus()
	showStats()
	waitSeconds(5, false)
	clearMap()
	gameMode = modeVision
	for _, s := range []string{"\n\n\nAll is Dark...\n", "\nBut wait...\n", "Where am I?...\n",
		"Am I dead?...\n", "Afterlife?...\n"} {
		printText(s)
		waitSeconds(5, false)
	}
	printText("You hear:\n")
	const nameWidth = 12 // the name is centred in the text window
	textCol = int16(uint16(nameWidth-uint16(strlen(party.members[0].name[:])))/2 + textLeft + 1)
	printText(cstr(party.members[0].name[:]))
	newLine()
	waitSeconds(5, false)
	printText("I feel motion...\n")
	waitSeconds(5, false)

	// Lord British's throne room, on the castle's upper floor
	party.tile = tileOnFoot
	party.x, party.y = 19, 8
	party.outX, party.outY = 86, 108
	talkFile = nil
	if party.loc == 0 {
		Save("OUTMONST.SAV", &area.people)
	}
	LoadRec("LCB_2.ULT", &area)
	talkFile = readFile("LCB.TLK")
	gameMode = modeTown
	party.loc = locCastleBritannia
	party.flyingOrLight = 0
	modeBeforeCombat = modeOutdoors
	spellTurns = 0
	hitTile = 0
	activeSpell = effectNone
	lineOfSight = 1
	gallopStep = 0
	redraw()
	updateMoons()
	printText("\n\nLord British says: I have pulled thy spirit and some possessions from the void.  Be more careful in the future!\n")
	for i := int(party.size) - 1; i >= 0; i-- {
		party.members[i].status = statusGood
		party.members[i].hp = party.members[i].maxHP
	}
	clear(party.weapons[:])
	clear(party.armors[:])
	party.food = 20099
	party.gold = 200
	showStats()
	flushKeys()
	backToCommands()
}

// checkPartyAlive says whether anyone in the party can act. When no one is left alive the
// party dies (partyDies, which does not return).
func checkPartyAlive() bool {
	anyAlive := false
	for i := int(party.size) - 1; i >= 0; i-- {
		if isConscious(i) {
			return true
		}
		anyAlive = anyAlive || isAlive(i)
	}
	if !anyAlive {
		partyDies()
	}
	return false
}

// damageMember takes damage from member's hit points; it says whether they died.
func damageMember(member int, damage int) bool {
	m := &party.members[member]
	if m.hp >= uint16(damage) {
		m.hp -= uint16(damage)
		showStats()
		return false
	}
	m.hp = 0
	m.status = statusDead
	if gameMode >= modeCombat {
		fighters.memberTile[member] = tileAsleep
		memberFrame[member] = tileAsleep
	}
	showStats()
	return true
}

func sayWhat() {
	printText("What?\n")
	flushKeys()
}

func sayCant() {
	printText("Can't!\n")
	flushKeys()
}

func sayNotHere() {
	printText("Not Here!\n")
	flushKeys()
}

func sayDisabled() {
	printText("Disabled!\n")
	flushKeys()
}

func sayOnlyOnFoot() {
	printText("Only on foot!\n")
	flushKeys()
}

func saySlowProgress() {
	printText("Slow progress!\n")
}

func sayNoneLeft() {
	printText("None left!\n")
}

// askLetter asks for a letter (or digit) from lo to hi, beeping at others. Letters come
// back in upper case; Enter or Space give -1, Esc -2.
func askLetter(prompt string, lo, hi int) int {
	for tries := 0; ; tries++ {
		if tries > 0 {
			sound(soundBeep)
		}
		printText(prompt)
		promptText, promptLo, promptHi = prompt, lo, hi
		if askingChara {
			ask(pChara)
		} else {
			ask(pLetter)
		}
		key := readKey()
		switch key {
		case keyEnter, keySpace:
			newLine()
			return -1
		case keyEsc:
			newLine()
			return -2
		}
		c := toUpper(keyChar(key))
		if c > ' ' && c <= lastPrintable {
			printChar(c)
		}
		newLine()
		if lo <= c && c <= hi {
			return c
		}
	}
}

// askMember asks for a party member: their index, -1 for none, -2 for 0 (the whole party,
// where a prompt allows it).
func askMember(prompt string) int {
	if party.size == 1 {
		printText(prompt)
		printText("1\n")
		return 0
	}
	askingChara = true
	c := askLetter(prompt, '0', '0'+int(party.size))
	askingChara = false
	switch {
	case c == '0':
		return -2
	case c < 0:
		return -1
	}
	return c - '1'
}

// askDirection asks for a direction with the arrow keys: the step (dx, dy), or (0, 0) when
// none was given.
func askDirection(prompt string) (dx, dy int) {
	for {
		printText(prompt)
		ask(pDir)
		key := readKey()
		switch key {
		case keySpace, keyEsc, keyBackspace, keyEnter:
		case keyUp:
			dy = -1
			printText("North")
		case keyDown:
			dy = 1
			printText("South")
		case keyRight:
			dx = 1
			printText("East")
		case keyLeft:
			dx = -1
			printText("West")
		default:
			if c := keyChar(key); c >= ' ' && c <= lastPrintable {
				printChar(c)
			}
			newLine()
			sound(soundBeep)
			continue
		}
		newLine()
		return dx, dy
	}
}

// eatFood takes food from the party; it says whether they ran out.
func eatFood(amount int) bool {
	party.food -= int32(amount)
	if party.food >= 0 {
		return false
	}
	party.food = 0
	return true
}

// regainMP gives every living member a magic point, up to what their class and
// intelligence allow.
func regainMP() {
	for i := int(party.size) - 1; i >= 0; i-- {
		if !isAlive(i) {
			continue
		}
		m := &party.members[i]
		in := int(m.intelligence)
		limit := 0
		switch m.class {
		case classMage:
			limit = in * 2
		case classBard, classPaladin, classRanger:
			limit = in
		case classDruid:
			limit = in/2 + in
		case classTinker:
			limit = in / 2
		}
		limit = min(limit, maxMP)
		m.mp++
		if m.mp > uint16(limit) {
			m.mp = uint16(limit)
		}
	}
}

// readLine reads a line typed into buf (at most size-1 characters, zero-terminated),
// without leading and trailing blanks.
func readLine(buf []byte, size int) {
	n := 0
	words := getsWords
	getsWords = nil
	for {
		if words != nil && n == 0 {
			ask(pWords)
			getsWords = words
		} else {
			ask(pText)
		}
		key := readKey()
		getsWords = nil
		switch key {
		case keyBackspace, keyDelete, keyLeft:
			if n == 0 {
				sound(soundBeep)
			} else {
				printChar('\b')
				n--
				buf[n] = ' '
			}
		case keyEnter:
			buf[n] = 0
		default:
			c := keyChar(key)
			if n == size-1 || c < ' ' || c >= 128 { // full, a control key or not ASCII
				sound(soundBeep)
			} else {
				buf[n] = byte(c)
				printChar(c)
				n++
			}
		}
		if key == keyEnter {
			break
		}
	}
	start := 0
	for buf[start] == ' ' {
		start++
	}
	if buf[start] != 0 {
		n = copy(buf, buf[start:start+strlen(buf[start:])])
	} else {
		n = start
	}
	for {
		buf[n] = 0
		n--
		if n < 0 || buf[n] != ' ' {
			break
		}
	}
}

// readWord reads a line of at most maxLen characters typed in, without the blanks around it.
func readWord(maxLen int) string {
	buf := make([]byte, maxLen+1)
	readLine(buf, maxLen+1)
	return cstr(buf)
}

// sameWord says whether a and b agree, ignoring case, in their first n letters.
func sameWord(a, b string, n int) bool {
	return strings.EqualFold(a[:min(n, len(a))], b[:min(n, len(b))])
}

// creatureName is the name of the creature or person shown by tile.
func creatureName(tile uint8) string {
	switch {
	case tile >= tileRat: // land monsters: four tiles each
		return landMonsterNames[int(tile-tileRat)/4]
	case tile >= tilePirateWest: // sea monsters: two tiles each
		return seaMonsterNames[int(tile-tilePirateWest)/2]
	case tile == tileAsleep:
		return villagerName
	case tile < tileMage || tile >= tileLetterA || (tile >= tileColumn && tile < tileGuard):
		return landMonsterNames[11] // "Phantom"
	}
	return peopleNames[int(tile&31)/2] // the classes and the townspeople
}

// damageParty hits the whole party (a trap, an earthquake): each living member, at random,
// takes 10 to 24 points. At sea the ship takes the blow.
func damageParty() {
	for i := int(party.size) - 1; i >= 0; i-- {
		invertStatusLine(i)
	}
	sound(soundMonsterHit)
	shakeScreen()
	for i := int(party.size) - 1; i >= 0; i-- {
		invertStatusLine(i)
	}
	if gameMode >= modeCombat || party.tile > tileShipSouth {
		for i := int(party.size) - 1; i >= 0; i-- {
			if randomBits(1) != 0 && isAlive(i) {
				damage := randomBelow(15) + 10
				if gameMode < modeCombat || fighters.memberTile[i] != 0 {
					damageMember(i, damage)
				}
			}
		}
	} else {
		party.hull -= 10 // unsigned: the original's "below 0" test never held
	}
	showStats()
}

// askYesNo asks for Y or N (Space, Esc and Enter end it too) and returns the key.
func askYesNo() int {
	key := -1
	for {
		if key != -1 {
			sound(soundBeep)
		}
		ask(pYN)
		key = toUpper(readKey() % 128)
		if key == 'N' || key == 'Y' || key == ' ' || key == 27 || key == '\r' {
			break
		}
	}
	if key == 'Y' || key == 'N' {
		printChar(key)
	}
	newLine()
	return key
}

// askNumber reads a number of up to digits digits; -1 when nothing was typed.
func askNumber(digits int) int32 {
	var buf [20]byte
	readLine(buf[:], digits+1)
	newLine()
	if buf[0] != 0 {
		return atol(buf[:])
	}
	return -1
}

// signInt is -1, 0 or 1 by a's sign.
func signInt(a int) int { return int(signOf(int16(a))) }
