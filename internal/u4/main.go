package u4

// The game's entry point and its command loop.

// Main runs the game (AVATAR.EXE) on the host's game goroutine. It returns nil when the game
// wants the title back (no party formed), else the error that stopped it.
func Main() (err error) {
	defer func() {
		if r := recover(); r != nil {
			s, ok := r.(stop)
			if !ok {
				panic(r)
			}
			err = s.err
		}
	}()
	resetGlobals()
	setHooks()
	run()
	return nil
}

// catchBackToCommands runs f, returning normally when it calls backToCommands.
func catchBackToCommands(f func()) (jumped bool) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(backToCommandsPanic); ok {
				jumped = true
				return
			}
			panic(r)
		}
	}()
	f()
	return false
}

// run starts the game where the party was saved (in a dungeon, or else on the world map)
// and runs the command loop for good.
func run() {
	seedRandom()
	startGame()
	if party.loc >= locFirstDungeon && party.loc <= locAbyss {
		gameMode = modeDungeon
		Load("DNGMAP.SAV", area.tiles[:dungeonMapSize])
		dungeonFile = readFile(dungeonFiles[party.loc-locFirstDungeon])
		catchBackToCommands(dungeonLoop)
	} else {
		party.loc = 0
	}
	flushKeys()
	var keys commandKeys
	for {
		catchBackToCommands(func() { commandLoop(&keys) })
	}
}

// commandKeys are the key of the command read last and the one before (they differ when a
// command went back to the command loop).
type commandKeys struct{ read, previous int }

// commandWait is how long the game waits for a command before the party passes.
const commandWait = 25

// commandLoop reads and runs commands, a turn passing after each (backToCommands returns
// here).
func commandLoop(keys *commandKeys) {
	for {
		redraw()
		groundEvent()
		if checkPartyAlive() {
			printChar(promptChar)
			ask(pCmd)
			waitSeconds(commandWait, true)
			key := int(keySpace)
			if keyWaiting() {
				key = readKey()
			}
			key = lowerKey(key)
			keys.read = key
			runCommand(key, keys.previous)
		} else {
			printText("Zzzzz\n")
		}
		if turnPasses {
			endTurn()
		} else {
			redraw()
			turnPasses = true
		}
		keys.previous = keys.read
	}
}

// runCommand runs the command of key; Alt-Z, Ctrl-S shows the karmas.
func runCommand(key, lastKey int) {
	switch key {
	case keySpace:
		sayPass()
	case keyUpAlt, keyUp:
		cmdNorth()
	case keyLeftAlt, keyLeft:
		cmdWest()
	case keyDownAlt, keyDown:
		cmdSouth()
	case keyRightAlt, keyRight:
		cmdEast()
	case keyA:
		cmdAttack()
	case keyB:
		cmdBoard()
	case keyC:
		cmdCast()
	case keyD:
		cmdDescend()
	case keyE:
		cmdEnter()
	case keyF:
		cmdFire()
	case keyG:
		cmdGet()
	case keyH:
		cmdHoleUp()
	case keyI:
		cmdIgnite()
	case keyJ:
		cmdJimmy()
	case keyK:
		cmdKlimb()
	case keyL:
		cmdLocate()
	case keyM:
		cmdMix()
	case keyN:
		cmdNewOrder()
	case keyO:
		cmdOpen()
	case keyP:
		cmdPeer()
	case keyQ:
		cmdQuit()
	case keyR:
		cmdReady()
	case keyS:
		cmdSearch()
	case keyT:
		cmdTalk()
	case keyU:
		cmdUse()
	case keyV:
		cmdVolume()
	case keyW:
		cmdWear()
	case keyX:
		cmdExit()
	case keyY:
		cmdYell()
	case keyZ:
		cmdZtats()
	case keyCtrlS:
		if lastKey == keyAltZ {
			showKarmas()
			return
		}
		badCommand()
	default:
		badCommand()
	}
}

func badCommand() {
	printText("Bad command!\n")
	sound(soundBadCommand)
	turnPasses = false
}

// showKarmas is the cheat that shows the eight karmas.
func showKarmas() {
	newLine()
	for _, k := range karmas {
		printNumber(int32(*k), 2, '0')
	}
	newLine()
}

// endTurn passes a turn: the ship mends, the asleep may wake, poison and hunger hurt,
// magic returns, whirlpools, twisters and creatures move, spells wear off, doors close.
func endTurn() {
	checkPartyAlive()
	if party.hull < 50 && randomBits(3) == 0 {
		party.hull++
	}
	for i := 0; i < int(party.size); i++ {
		m := &party.members[i]
		if m.status == statusAsleep && randomBits(7) == 0 {
			m.status = statusGood
		} else if m.status == statusPoisoned {
			damageMember(i, 2)
			invertStatusLine(i)
			sound(soundMonsterHit)
			invertStatusLine(i)
		}
	}
	if eatFood(int(party.size)) {
		printText("\nStarving!!!\n")
		for i := 0; i < int(party.size); i++ {
			if isAlive(i) {
				damageMember(i, 2)
			}
		}
		for i := int(party.size) - 1; i >= 0; i-- {
			invertStatusLine(i)
		}
		sound(soundMonsterHit)
		for i := int(party.size) - 1; i >= 0; i-- {
			invertStatusLine(i)
		}
	}
	regainMP()
	if party.flyingOrLight == 0 {
		whirlpoolsAndTwisters()
		movePeople()
		spawnMonsters()
		whirlpoolsAndTwisters()
	}
	party.moves++
	if spellTurns != 0 {
		spellTurns--
		if spellTurns == 0 {
			activeSpell = effectNone
		}
	}
	showStats()
	closeDoorInTime()
}

// turnPasses is cleared by a command that takes no time.
var turnPasses = true

// karmas are the eight virtues' karmas, in the order of the virtues.
var karmas = [8]*uint16{
	&party.karmaHonesty, &party.karmaCompassion, &party.karmaValor, &party.karmaJustice,
	&party.karmaSacrifice, &party.karmaHonor, &party.karmaSpirituality, &party.karmaHumility,
}

// cursorBoxX is the column of the box around the active member in combat; -1 hides it.
var cursorBoxX int16 = -1

// dirDX and dirDY are the steps west, north, east and south.
var (
	dirDX = [4]int{-1, 0, 1, 0}
	dirDY = [4]int{0, -1, 0, 1}
)

// moongateX and moongateY are where the moongates stand: Trammel's phase picks the
// moongate, Felucca's where it leads.
var (
	moongateX = [8]uint8{224, 96, 38, 50, 166, 104, 23, 187}
	moongateY = [8]uint8{133, 102, 224, 37, 19, 194, 126, 167}
)

// townFiles are the maps of the castles, townes and villages, by location.
var townFiles = [16]string{
	"LCB_1.ULT", "LYCAEUM.ULT", "EMPATH.ULT", "SERPENT.ULT",
	"MOONGLOW.ULT", "BRITAIN.ULT", "JHELOM.ULT", "YEW.ULT",
	"MINOC.ULT", "TRINSIC.ULT", "SKARA.ULT", "MAGINCIA.ULT",
	"PAWS.ULT", "DEN.ULT", "VESPER.ULT", "COVE.ULT",
}

// placeX and placeY are where the locations lie on the world map.
var (
	placeX = [32]uint8{86, 218, 28, 146, 232, 82, 36, 58, 159, 106, 22, 187, 98, 136, 201, 136,
		240, 91, 72, 126, 156, 58, 239, 233, 233, 128, 36, 73, 205, 81, 231, 231}
	placeY = [32]uint8{107, 107, 50, 241, 135, 106, 222, 43, 20, 184, 128, 169, 145, 158, 59, 90,
		73, 67, 168, 20, 27, 102, 240, 233, 66, 92, 229, 11, 45, 207, 216, 216}
)

// colorNames are the colours of the stones.
var colorNames = [8]string{"Blue", "Yellow", "Red", "Green", "Orange", "Purple", "White", "Black"}

// dungeonFiles are the dungeons' maps, by location.
var dungeonFiles = [8]string{
	"Deceit.Dng", "Despise.Dng", "Destard.Dng", "Wrong.Dng",
	"Covetous.Dng", "Shame.Dng", "Hythloth.Dng", "Abyss.Dng",
}
