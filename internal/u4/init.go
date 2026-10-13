package u4

// startGame loads the font, the tiles, the saved party and the world, and draws the screen.
// With no party saved it goes back to the title.
func startGame() {
	charset := make([]byte, 4096)
	Load("CHARSET.EGA", charset)
	checkDriverFile("EGA.DRV")
	initDriver()
	screen.SetGlyphs(charset)
	clearScreen()
	shapes := make([]byte, 32768)
	Load("SHAPES.EGA", shapes)
	screen.SetTiles(shapes)
	LoadRec("PARTY.SAV", &party)
	if party.size == 0 {
		clearScreen()
		textRow, textCol = 8, 12
		printText("No party formed!")
		sound(soundFailure)
		waitSeconds(3, false)
		flushKeys()
		backToTitle()
	}
	LoadRec("MONSTERS.SAV", &area.people)
	if worldMap = readFile("WORLD.MAP"); len(worldMap) < worldChunks*chunkSize {
		fatalf("WORLD.MAP: file wrong size")
	}
	talkFile = nil
	moongateAtX = moongateX[party.trammel]
	moongateAtY = moongateY[party.felucca]
	drawFrame()
	newLine()
	activeSpell = effectNone
	spellTurns = 0
	lineOfSight = uint16(b2i(party.flyingOrLight == 0))
	gallopStep = 0
	gameMode = modeOutdoors
	windDir = dirNorth
	soundOn = 1
	showStats()
	centerMap()
}
