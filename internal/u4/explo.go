package u4

// Entering and leaving places, boarding and leaving vehicles, doors, klimbing and
// descending.

// talkFiles are the conversations of the castles, townes and villages, by location.
var talkFiles = [16]string{
	"LCB.TLK", "LYCAEUM.TLK", "EMPATH.TLK", "SERPENT.TLK",
	"MOONGLOW.TLK", "BRITAIN.TLK", "JHELOM.TLK", "YEW.TLK",
	"MINOC.TLK", "TRINSIC.TLK", "SKARA.TLK", "MAGINCIA.TLK",
	"PAWS.TLK", "DEN.TLK", "VESPER.TLK", "COVE.TLK",
}

// Where the party starts in a place.
const (
	towneGateX, towneGateY         = 1, 15  // a towne's, village's or ruin's west gate
	castleGateX, castleGateY       = 15, 30 // a castle's south gate
	dungeonStartX, dungeonStartY   = 1, 1
	hythlothStartX, hythlothStartY = 5, 5 // down the ladder from Lord British's cellar

	abyssX, abyssY = 233, 233 // the Abyss's entrance in the lava
	cellarLadderY  = 2        // the ladder in Lord British's castle that leads into Hythloth
)

// placeName is the name of location loc.
func placeName(loc uint16) string { return placeNames[int(loc)-1] }

// loadTown loads a castle's, towne's or village's map, people and conversations.
func loadTown(loc int) {
	Save("OUTMONST.SAV", &area.people)
	LoadRec(townFiles[loc-1], &area)
	talkFile = readFile(talkFiles[party.loc-1])
	gameMode = modeTown
}

// loadDungeon loads the dungeon's map, and clears the people.
func loadDungeon() {
	file := dungeonFiles[party.loc-locFirstDungeon]
	Load(file, area.tiles[:dungeonMapSize])
	dungeonFile = readFile(file)
	area.people.tile = [32]uint8{}
}

// startDungeon takes the party into the dungeon, on its first level facing east.
func startDungeon() {
	loadDungeon()
	gameMode = modeDungeon
	party.dir = dirEast
	party.z = 0
	dungeonLoop()
}

// enterDungeon takes the party, if on foot, into a dungeon.
func enterDungeon() {
	newLine()
	newLine()
	if party.tile != tileOnFoot {
		party.loc = 0
		sayOnlyOnFoot()
		return
	}
	party.x, party.y = dungeonStartX, dungeonStartY
	Save("OUTMONST.SAV", &area.people)
	startDungeon()
}

// StartInTown, when set, is asked where the party stands on entering town loc (1..16)
// instead of the west gate; the replay test uses it to start at a shop counter.
var StartInTown func(loc int) (x, y uint8, ok bool)

// enterTowne takes the party through a towne's west gate. In a class's home towne the one
// waiting to join (the last of its people) is gone when they are in the party already.
func enterTowne() {
	printText(placeName(party.loc))
	newLine()
	newLine()
	party.x, party.y = towneGateX, towneGateY
	windowPos.x, windowPos.y = towneGateX, towneGateY
	loadTown(int(party.loc))
	if StartInTown != nil {
		if x, y, ok := StartInTown(int(party.loc)); ok {
			party.x, party.y = x, y
			windowPos.x, windowPos.y = x, y
		}
	}
	if party.loc >= locFirstTowne && party.loc < locFirstTowne+8 {
		class := party.loc - locFirstTowne // the townes are in the classes' order
		const companion = 31
		for i := int(party.size) - 1; i >= 1; i-- {
			if uint16(party.members[i].class) == class {
				area.people.behaviour[companion] = 0
				area.people.baseTile[companion] = 0
				area.people.tile[companion] = 0
			}
		}
	}
}

// enterAbyss lets the party into the Abyss once the Bell, Book and Candle have opened it.
func enterAbyss() {
	opened := true
	for _, item := range []int{itemBook, itemBell, itemCandle, itemBookUsed, itemBellUsed, itemCandleUsed} {
		opened = opened && hasBit(party.items, item)
	}
	if !opened {
		party.loc = 0
		sayCant()
		return
	}
	printText(placeName(party.loc))
	enterDungeon()
}

func cmdEnter() {
	printText("Enter ")
	if party.loc != 0 || party.tile == tileBalloon {
		sayWhat()
		return
	}
	place := -1
	for i := len(placeX) - 1; i >= 0; i-- {
		if party.x == placeX[i] && party.y == placeY[i] {
			place = i
			break
		}
	}
	if place == -1 {
		sayWhat()
		return
	}
	party.loc = uint16(place + 1)
	party.outX, party.outY = party.x, party.y
	switch *area.townTile(int(windowPos.y), int(windowPos.x)) {
	case tileDungeon:
		printText("dungeon!\n\n")
		printText(placeName(party.loc))
		enterDungeon()
	case tileTowne:
		printText("towne!\n\n")
		enterTowne()
	case tileCastle, tileCastleBritanniaEntrance:
		printText("castle!\n\n")
		printText(placeName(party.loc))
		newLine()
		newLine()
		party.x, party.y = castleGateX, castleGateY
		windowPos.x, windowPos.y = castleGateX, castleGateY
		party.z = 0
		loadTown(int(party.loc))
	case tileVillage:
		printText("village!\n\n")
		enterTowne()
	case tileRuins:
		printText("ruin!\n\n")
		enterTowne()
	case tileLava:
		if party.x != abyssX || party.y != abyssY {
			sayWhat()
			return
		}
		enterAbyss()
	case tileFireField:
		enterAbyss()
	case tileShrine:
		printText("the Shrine of\n")
		printText(virtueNames[int(party.loc)-locFirstShrine])
		newLine()
		enterShrine()
	default:
		sayWhat()
	}
}

// leaveVehicle leaves the party's ship, horse or balloon where it is, as one of the world's
// people (the first eight are monsters): in a free place, or if none is free over one of
// the last sixteen.
func leaveVehicle() {
	people := &area.people
	slot := -1
	for i := len(people.tile) - 1; i >= 8; i-- {
		if people.tile[i] == 0 {
			slot = i
			break
		}
	}
	if slot == -1 {
		slot = 16 + randomBits(15)
	}
	people.tile[slot] = uint8(party.tile)
	people.baseTile[slot] = uint8(party.tile)
	people.x[slot] = party.x
	people.y[slot] = party.y
	people.talkIndex[slot] = 0
	people.behaviour[slot] = 0
	party.tile = tileOnFoot
	newLine()
}

func cmdExit() {
	printText("X-it ")
	switch {
	case party.tile <= tileShipSouth:
		shipX, shipY = party.x, party.y
		leaveVehicle()
	case party.tile <= tileHorseEast:
		gallopStep = 0
		leaveVehicle()
	case party.tile == tileBalloon && party.flyingOrLight == 0:
		leaveVehicle()
	default:
		sayWhat()
	}
}

// takeVehicle takes the vehicle the party stands on: kind is the first of its tiles (its
// tiles differ in the two lowest bits only).
func takeVehicle(kind int) {
	people := &area.people
	for i := len(people.tile) - 1; i >= 8; i-- {
		if party.x == people.x[i] && party.y == people.y[i] && kind == int(people.tile[i]&^3) {
			people.tile[i] = 0
			people.baseTile[i] = 0
			break
		}
	}
	party.tile = uint16(hereTile)
}

func cmdBoard() {
	if party.tile != tileOnFoot {
		printText("Board: ")
		sayCant()
		return
	}
	if hereTile == tileHorseWest || hereTile == tileHorseEast {
		takeVehicle(tileHorseWest)
		printText("Mount Horse!\n")
		return
	}
	printText("Board ")
	if hereTile == tileBalloon {
		takeVehicle(tileBalloon)
		printText("Balloon!\n")
		party.flyingOrLight = 0
		return
	}
	if hereTile < tileShipWest || hereTile > tileShipSouth {
		sayWhat()
		return
	}
	takeVehicle(tileShipWest)
	printText("Frigate!\n")
	if shipX != party.x || shipY != party.y {
		party.hull = 50 // not the ship left last: a new one
	}
}

func cmdYell() {
	printText("Yell ")
	if party.tile != tileHorseWest && party.tile != tileHorseEast {
		sayWhat()
		return
	}
	gallopStep ^= 1
	if gallopStep != 0 {
		printText("Giddyup!\n")
	} else {
		printText("Whoa!\n")
	}
}

// An opened door closes again after doorOpenTurns turns.
const doorOpenTurns = 5

var (
	openDoorX, openDoorY uint16
	doorCloseTurns       uint16 // turns until the open door closes; 0: none open
)

// closeDoorInTime counts down the open door's turns and closes it at the end.
func closeDoorInTime() {
	if gameMode != modeTown || doorCloseTurns == 0 {
		doorCloseTurns = 0
		return
	}
	doorCloseTurns--
	if doorCloseTurns == 0 {
		*area.townTile(int(openDoorY), int(openDoorX)) = tileDoor
	}
}

func cmdOpen() {
	for doorCloseTurns != 0 {
		closeDoorInTime()
	}
	if gameMode != modeTown {
		printText("Open; ")
		sayNotHere()
		return
	}
	dx, dy := askDirection("Open: ")
	if dx == 0 && dy == 0 {
		return
	}
	x := dx + int(party.x)
	if x < 0 || x >= townTiles || dy != 0 { // doors open only east or west
		sayNotHere()
		return
	}
	door := area.townTile(int(party.y), x)
	switch *door {
	case tileLockedDoor:
		sayCant()
	case tileDoor:
		openDoorX, openDoorY = uint16(x), uint16(party.y)
		doorCloseTurns = doorOpenTurns
		*door = tileBrickFloor
		printText("\nOpened!\n")
	default:
		sayNotHere()
	}
}

func cmdJimmy() {
	printText("Jimmy lock!\n")
	dx, dy := askDirection("Dir: ")
	if dx == 0 && dy == 0 {
		return
	}
	// Unlike Open this lets x be -1 at the town's west edge (the tile before the row).
	x := dx + int(party.x)
	if gameMode != modeTown || x >= townTiles || dy != 0 {
		sayNotHere()
		return
	}
	door := area.townTile(int(party.y), x)
	if *door != tileLockedDoor {
		sayNotHere()
		return
	}
	if party.keys == 0 {
		printText("No keys left!\n")
		return
	}
	party.keys--
	*door = tileDoor
}

func cmdKlimb() {
	printText("Klimb ")
	if party.loc == 0 {
		if party.tile != tileBalloon {
			sayWhat()
			return
		}
		printText("altitude\n")
		party.flyingOrLight = 1
		lineOfSight = 0
		return
	}
	if party.loc != locCastleBritannia || hereTile != tileLadderUp {
		sayWhat()
		return
	}
	if party.tile != tileOnFoot {
		newLine()
		sayOnlyOnFoot()
		return
	}
	printText("to second floor!\n")
	LoadRec("LCB_2.ULT", &area)
}

func cmdDescend() {
	if party.tile == tileBalloon {
		printText("Land Balloon\n")
		if hereTile != tileGrass {
			sound(soundBeep)
			sayNotHere()
			return
		}
		if party.flyingOrLight == 0 {
			printText("Already Landed!\n")
			return
		}
		party.flyingOrLight = 0
		lineOfSight = 1
		return
	}
	if party.loc != locCastleBritannia {
		printText("Descend ")
		sayWhat()
		return
	}
	if party.tile != tileOnFoot {
		sayOnlyOnFoot()
		return
	}
	printText("Descend ")
	if hereTile != tileLadderDown {
		sayWhat()
		return
	}
	if party.y == cellarLadderY {
		printText("into the depths!\n")
		party.outX, party.outY = hythlothX, hythlothY
		party.x, party.y = hythlothStartX, hythlothStartY
		party.loc = locHythloth
		startDungeon()
		return
	}
	printText("to first floor!\n")
	LoadRec("LCB_1.ULT", &area)
}
