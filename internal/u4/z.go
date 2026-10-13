package u4

// The Ztats pages in the status area: one per party member, then the weapons, armour,
// equipment, items, reagents and mixtures.

// The status area: columns statusLeft to statusRight of rows 0 (the title) to statusBottom.
const (
	statusLeft   = 24
	statusRight  = 39
	statusWidth  = statusRight - statusLeft
	statusBottom = 8
)

// Glyphs of the font that frame the status area's title.
const (
	glyphTitleBar   = 13 // the line across the top of the status area
	glyphTitleLeft  = 16
	glyphTitleRight = 17
)

// The pages, in the order the arrow keys step through them: one per member, then the
// party's belongings.
const (
	pageWeapons = 8 + iota
	pageArmour
	pageEquipment
	pageItems
	pageReagents
	pageMixtures
	pageCount
)

// The letters of the stones (by virtue) and runes shown on the items page.
const (
	stoneLetters = "BYRGOPWB"
	runeLetters  = "HCVJSHSH"
)

// blankStatusBelow blanks the status rows from the cursor's row down.
func blankStatusBelow() {
	for ; textRow <= statusBottom; textRow++ {
		textCol = statusLeft
		printText("               ")
	}
}

// printStatusTitle prints text centred in the status area (at most 12 letters), nudged by
// shift columns; on the title row it is framed by the title glyphs.
func printStatusTitle(text string, shift int) {
	width := min(len(text), 12)
	width -= width % 2
	col := (12-width)/2 + shift + statusLeft + 1
	if width >= 8 {
		col++
	}
	textCol = int16(col)
	if textRow == 0 {
		textCol--
		printChar(glyphTitleLeft)
	}
	printText(text)
	if textRow == 0 {
		printChar(glyphTitleRight)
	}
}

// clearStatusTitle redraws the bare line across the top of the status area.
func clearStatusTitle() {
	row, col := textRow, textCol
	textRow = 0
	for textCol = statusLeft; textCol < statusRight; textCol++ {
		putGlyph(glyphTitleBar)
	}
	textRow, textCol = row, col
}

// printAt moves the cursor to row and col and prints text there.
func printAt(row, col int, text string) {
	textRow, textCol = int16(row), int16(col)
	printText(text)
}

// showMemberPage shows a party member's name, class, stats and equipment.
func showMemberPage(member int) {
	m := &party.members[member]
	textRow = 0
	printStatusTitle(cstr(m.name[:]), 0)
	textCol = statusLeft
	textRow = 1
	printChar(int(m.sex))
	printStatusTitle(peopleNames[int(m.class)], 0)
	textCol = statusRight - 1
	printChar(int(m.status))
	printAt(3, 25, "MP:")
	printNumber(int32(m.mp), 2, '0')
	printAt(3, 32, "LV:")
	printChar(int(m.maxHP/100) + '0')
	printAt(4, statusLeft, "STR:")
	printNumber(int32(m.strength), 2, '0')
	printAt(4, 32, "HP:")
	printNumber(int32(m.hp), 4, '0')
	printAt(5, statusLeft, "DEX:")
	printNumber(int32(m.dexterity), 2, '0')
	printAt(5, 32, "HM:")
	printNumber(int32(m.maxHP), 4, '0')
	printAt(6, statusLeft, "INT:")
	printNumber(int32(m.intelligence), 2, '0')
	printAt(6, 32, "EX:")
	printNumber(int32(m.xp), 4, '0')
	printAt(7, statusLeft, "W:")
	printText(weaponNames[int(m.weapon)])
	printAt(8, statusLeft, "A:")
	printText(armorNames[int(m.armor)])
}

// nextWeaponSlot moves the cursor to the next of the weapons page's slots: down a column
// of eight rows, then on to the next column. Every slot starts on a multiple of eight.
func nextWeaponSlot() {
	textCol = (textCol - 1) / 8 * 8
	textRow++
	if textRow == statusBottom+1 {
		textRow = 1
		textCol += 8
	}
}

// printCount prints an item's letter and count, as "B-12-".
func printCount(index int, count uint16, pad int) {
	printChar(index + 'A')
	printNumber(int32(count), 2, pad)
	printChar('-')
}

// showWeapons lists the party's weapons, by their letters, in two columns.
func showWeapons() {
	textRow = 0
	printStatusTitle("Weapons", 0)
	printAt(1, statusLeft, "A-Hands")
	textRow = 2
	textCol = statusLeft
	for w := weaponStaff; w < len(party.weapons); w++ {
		if party.weapons[w] != 0 {
			printCount(w, party.weapons[w], '-')
			printText(weaponCodes[w])
			nextWeaponSlot()
		}
	}
	for textCol < 40 {
		printText("       ")
		nextWeaponSlot()
	}
}

// printToEndOfLine blanks the rest of the status line and moves to the next.
func printToEndOfLine() {
	for textCol != statusRight {
		printChar(' ')
	}
	textRow++
}

// showArmour lists the party's armour.
func showArmour() {
	textRow = 0
	printStatusTitle("Armour", 0)
	printAt(1, statusLeft, "A  -No Armour")
	textRow = 2
	for a := 1; a < len(party.armors); a++ {
		if party.armors[a] != 0 {
			textCol = statusLeft
			printCount(a, party.armors[a], '-')
			printText(armorNames[a])
			printToEndOfLine()
		}
	}
	blankStatusBelow()
}

// showEquipment shows the party's torches, gems, keys and sextants.
func showEquipment() {
	textRow = 0
	printStatusTitle("Equipment", -1)
	line := func(row int, count uint16, what string) {
		textCol, textRow = statusLeft, int16(row)
		printNumber(int32(count), 2, ' ')
		printText(what)
	}
	line(1, party.torches, " Torches")
	line(2, party.gems, " Gems")
	line(3, party.keys, " Keys")
	if party.sextants != 0 {
		line(4, party.sextants, " Sextants")
	}
	textRow++
	blankStatusBelow()
}

// showItems shows the stones, runes and quest items the party holds.
func showItems() {
	it := party.items
	textRow = 0
	printStatusTitle("Items", 0)
	textRow = 1
	if party.stones != 0 {
		textCol = statusLeft
		printText("Stones:")
		for v := range virtueNames {
			if hasBit(party.stones, v) {
				printChar(int(stoneLetters[v]))
			}
		}
		textRow++
	}
	if party.runes != 0 {
		textCol = statusLeft
		printText("Runes:")
		for v := range virtueNames {
			if hasBit(party.runes, v) {
				printChar(int(runeLetters[v]))
			}
		}
		textRow++
	}
	if hasBit(it, itemBell) || hasBit(it, itemBook) || hasBit(it, itemCandle) {
		textCol = statusLeft
		if hasBit(it, itemBell) {
			printText("Bell ")
		}
		if hasBit(it, itemBook) {
			printText("Book ")
		}
		if hasBit(it, itemCandle) {
			printText("Candl")
			if !hasBit(it, itemBook) || !hasBit(it, itemBell) {
				printChar('e') // with all three the line has no room for it
			}
		}
		textRow++
	}
	if hasBit(it, itemKeyCourage) || hasBit(it, itemKeyLove) || hasBit(it, itemKeyTruth) {
		textCol = statusLeft
		printText("3 Part Key:")
		if hasBit(it, itemKeyTruth) {
			printChar('T')
		}
		if hasBit(it, itemKeyLove) {
			printChar('L')
		}
		if hasBit(it, itemKeyCourage) {
			printChar('C')
		}
		textRow++
	}
	for _, item := range []struct {
		bit  int
		name string
	}{{itemHorn, "Horn"}, {itemWheel, "Wheel"}, {itemSkull, "Skull"}} {
		if hasBit(it, item.bit) {
			printAt(int(textRow), statusLeft, item.name)
			if item.bit != itemSkull {
				textRow++
			}
		}
	}
}

// showReagents lists the party's reagents.
func showReagents() {
	textRow = 0
	printStatusTitle("Reagents", 0)
	textRow = 1
	for r := range party.reagents {
		if party.reagents[r] != 0 {
			textCol = statusLeft
			printCount(r, party.reagents[r], '-')
			printText(reagentNames[r])
			printToEndOfLine()
		}
	}
	blankStatusBelow()
}

// showMixtures lists the mixed spells, in columns of eight, as many as fit.
func showMixtures() {
	textRow = 0
	printStatusTitle("Mixtures", 0)
	textRow = 1
	textCol = statusLeft
	for s := range party.mixtures {
		if party.mixtures[s] == 0 {
			continue
		}
		printChar(s + 'A')
		printChar('-')
		printNumber(int32(party.mixtures[s]), 2, '0')
		textCol -= 4
		textRow++
		if textRow == statusBottom+1 {
			textRow = 1
			textCol += 5
			if textCol >= statusRight {
				return
			}
		}
	}
}

// showPage shows one of the Ztats pages.
func showPage(page int) {
	switch page {
	case pageWeapons:
		showWeapons()
	case pageArmour:
		showArmour()
	case pageEquipment:
		showEquipment()
	case pageItems:
		showItems()
	case pageReagents:
		showReagents()
	case pageMixtures:
		showMixtures()
	default:
		showMemberPage(page)
	}
}

// browsePages shows the Ztats pages from page on: the arrows, Home, End, PgUp and PgDn step
// between them, a digit picks a member, and any other key ends.
func browsePages(page int) {
	row, col := textRow, textCol
	members := int(party.size)
	lastMember := members - 1
	for page != -1 {
		clearStatus()
		clearStatusTitle()
		showPage(page)
		for !keyWaiting() {
		}
		switch key := readKey(); key {
		case key1, key2, key3, key4, key5, key6, key7, key8:
			if n := keyChar(key) - '0'; n <= members {
				page = n - 1
			} else {
				sound(soundBeep)
			}
		case keyRight, keyDown:
			page = (page + 1) % pageCount
			if page == members {
				page = pageWeapons
			}
		case keyUp, keyLeft:
			if page == pageWeapons {
				page = lastMember
			} else {
				page = (page + pageCount - 1) % pageCount
			}
		case keyPageUp: // the first page of the group, or the group before
			if page >= 1 && page <= pageWeapons {
				page = 0
			} else if page > pageWeapons && page < pageCount {
				page = pageWeapons
			} else {
				page = pageMixtures
			}
		case keyHome:
			page = 0
		case key0:
			page = pageWeapons
		case keyPageDown: // the next group's start, or the last page
			if page >= 0 && page < pageWeapons {
				page = pageWeapons
			} else if page >= pageWeapons && page < pageMixtures {
				page = pageMixtures
			} else {
				page = 0
			}
		case keyEnd:
			page = pageMixtures
		default:
			page = -1
		}
	}
	clearStatusTitle()
	clearStatus()
	showStats()
	textRow, textCol = row, col
}

func cmdZtats() {
	if gameMode >= modeCombat {
		printText("Ztats\n")
		invertStatusLine(int(activeMember))
		browsePages(int(activeMember))
		invertStatusLine(int(activeMember))
		return
	}
	member := askMember("Ztats for:" + answerRoom(2))
	if member == -1 {
		return
	}
	if member == -2 { // no member named: the party's belongings
		member = pageWeapons
	}
	browsePages(member)
}
