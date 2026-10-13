package title

// Port of DATA.C / DATA.ASM: the title's animation tables.

// The monsters' frames in ANIMATE.EGA: frame f is 6 columns by 32 rows, in row f%6 of the
// picture; the first monster's frames are columns 0, 7 and 14 (by f/6), the second's 34, 28
// and 22.
const (
	monsterCols   = 6
	monsterHeight = 32
	monsterBCol   = 34 // where the second monster stands, in columns
)

func monsterFrameY(f uint8) int    { return int(f%6) * monsterHeight }
func monsterAFrameCol(f uint8) int { return int(f/6) * 7 }
func monsterBFrameCol(f uint8) int { return monsterBCol - int(f/6)*6 }

// monsterASteps is the first monster's sequence of frames. It is 128 steps long but the
// original's table holds 120: the last 8 run on into the second monster's.
var monsterASteps = append(monsterAOwnSteps[:], monsterBSteps[:8]...)

var monsterAOwnSteps = [...]uint8{
	1, 1, 1, 0, 0, 1, 1, 1, 0, 0, 1, 1, 2, 2, 3, 3,
	4, 4, 1, 2, 3, 4, 1, 2, 5, 6, 7, 8, 5, 6, 7, 8,
	5, 6, 7, 8, 5, 6, 7, 8, 5, 6, 7, 8, 5, 6, 7, 8,
	9, 10, 9, 10, 9, 10, 11, 11, 11, 11, 12, 12, 13, 13, 12, 13,
	12, 13, 12, 11, 11, 11, 0, 0, 1, 2, 3, 4, 1, 2, 5, 6,
	7, 8, 5, 6, 7, 8, 9, 10, 11, 11, 11, 0, 0, 14, 14, 14,
	15, 16, 16, 16, 17, 17, 17, 16, 16, 16, 17, 17, 17, 16, 16, 16,
	15, 14, 14, 0, 0, 11, 11, 11,
}

// monsterBSteps is the second monster's sequence of frames (64 steps).
var monsterBSteps = [...]uint8{
	1, 0, 1, 2, 3, 4, 3, 2, 1, 0, 1, 2, 3, 4, 5, 6,
	5, 6, 5, 6, 4, 7, 8, 9, 10, 9, 8, 7, 8, 9, 10, 11,
	12, 11, 12, 13, 11, 12, 13, 1, 13, 1, 14, 1, 15, 1, 14, 1,
	15, 10, 9, 8, 16, 17, 16, 17, 16, 17, 9, 8, 7, 4, 3, 2,
}

// signature is "Lord British" handwritten, dot by dot: each dot's x (from x 20) and its
// height (up from y 191).
var signature = [...][2]uint8{
	{84, 189}, {85, 188}, {87, 188}, {89, 188}, {91, 188}, {92, 189}, {92, 190}, {91, 191},
	{89, 191}, {88, 190}, {88, 189}, {87, 187}, {86, 186}, {86, 185}, {85, 184}, {85, 183},
	{84, 182}, {84, 181}, {83, 180}, {83, 179}, {82, 178}, {82, 177}, {81, 176}, {79, 176},
	{78, 176}, {77, 177}, {77, 178}, {78, 179}, {80, 179}, {81, 178}, {84, 178}, {85, 177},
	{86, 177}, {87, 176}, {89, 176}, {91, 176}, {93, 176}, {95, 176}, {97, 176}, {99, 176},
	{101, 176}, {103, 176}, {105, 176}, {107, 176}, {109, 176}, {111, 176}, {113, 176}, {115, 176},
	{117, 176}, {119, 176}, {121, 176}, {123, 176}, {125, 176}, {127, 176}, {129, 176}, {131, 176},
	{133, 176}, {135, 176}, {137, 176}, {139, 176}, {141, 176}, {143, 176}, {145, 176}, {147, 176},
	{149, 176}, {151, 176}, {153, 176}, {155, 176}, {157, 176}, {159, 176}, {161, 176}, {163, 176},
	{165, 176}, {167, 176}, {169, 176}, {171, 176}, {173, 176}, {175, 176}, {177, 176}, {179, 176},
	{181, 176}, {183, 176}, {185, 176}, {187, 176}, {189, 176}, {191, 176}, {193, 176}, {195, 176},
	{197, 176}, {199, 176}, {201, 176}, {202, 176}, {203, 177}, {204, 177}, {205, 178}, {94, 184},
	{94, 183}, {93, 182}, {93, 181}, {92, 180}, {92, 179}, {93, 178}, {95, 178}, {96, 178},
	{97, 179}, {97, 180}, {98, 181}, {98, 182}, {99, 183}, {99, 184}, {98, 185}, {96, 185},
	{95, 185}, {105, 185}, {106, 184}, {106, 183}, {105, 182}, {105, 181}, {104, 180}, {104, 179},
	{103, 178}, {107, 184}, {108, 185}, {110, 185}, {119, 185}, {117, 185}, {116, 185}, {115, 184},
	{115, 183}, {114, 182}, {114, 181}, {113, 180}, {113, 179}, {114, 178}, {116, 178}, {117, 179},
	{118, 180}, {119, 181}, {119, 182}, {120, 183}, {120, 184}, {121, 185}, {121, 186}, {122, 187},
	{122, 188}, {123, 189}, {123, 190}, {118, 179}, {119, 178}, {139, 190}, {139, 189}, {138, 188},
	{138, 187}, {137, 186}, {137, 185}, {136, 184}, {136, 183}, {135, 182}, {135, 181}, {134, 180},
	{134, 179}, {133, 178}, {140, 191}, {142, 191}, {143, 191}, {144, 190}, {144, 189}, {143, 188},
	{143, 187}, {142, 186}, {142, 185}, {140, 185}, {143, 184}, {143, 183}, {142, 182}, {142, 181},
	{141, 180}, {141, 179}, {140, 178}, {138, 178}, {136, 178}, {135, 179}, {150, 185}, {151, 184},
	{151, 183}, {150, 182}, {150, 181}, {149, 180}, {149, 179}, {148, 178}, {152, 184}, {153, 185},
	{155, 185}, {161, 185}, {160, 184}, {160, 183}, {159, 182}, {159, 181}, {158, 180}, {158, 179},
	{157, 178}, {162, 188}, {162, 187}, {169, 188}, {169, 187}, {168, 186}, {168, 185}, {167, 184},
	{167, 183}, {166, 182}, {166, 181}, {165, 180}, {165, 179}, {166, 178}, {168, 178}, {169, 178},
	{170, 179}, {165, 185}, {166, 185}, {170, 185}, {178, 185}, {177, 184}, {177, 183}, {176, 182},
	{176, 181}, {175, 180}, {175, 179}, {174, 178}, {179, 188}, {179, 187}, {187, 184}, {186, 185},
	{184, 185}, {183, 185}, {182, 184}, {182, 183}, {183, 182}, {184, 181}, {185, 180}, {185, 179},
	{184, 178}, {182, 178}, {181, 178}, {180, 179}, {197, 190}, {197, 189}, {196, 188}, {196, 187},
	{195, 186}, {195, 185}, {194, 184}, {194, 183}, {193, 182}, {193, 181}, {192, 180}, {192, 179},
	{191, 178}, {197, 185}, {198, 185}, {199, 184}, {199, 183}, {198, 182}, {198, 181}, {197, 180},
	{197, 179}, {198, 178},
}

// viewBackground is the view's map before its actors: 5 rows of 19 tiles.
var viewBackground = [viewTiles]uint8{
	6, 6, 6, 4, 4, 4, 1, 1, 0, 0, 0, 0, 1, 4, 4, 13, 14, 15, 4,
	6, 6, 4, 4, 4, 1, 1, 1, 1, 0, 0, 1, 1, 4, 4, 4, 4, 4, 4,
	6, 4, 4, 1, 1, 1, 2, 2, 1, 1, 1, 1, 1, 1, 10, 4, 4, 4, 6,
	6, 4, 4, 1, 1, 2, 2, 1, 1, 9, 8, 1, 1, 1, 1, 4, 6, 6, 6,
	4, 4, 4, 4, 1, 1, 1, 1, 4, 4, 8, 8, 1, 1, 1, 1, 1, 6, 6,
}

// The view's script, a byte stream of commands:
//   - place(row, actor), at(frame, col): put an actor on the view with a frame of its tile
//   - vanish(actor): take it off
//   - endFrame(n): draw the view, then wait n frames before reading on
//   - scriptRestart: start again from the beginning
const (
	scriptVanish   = 112
	scriptEndFrame = 128
	scriptRestart  = 255
)

func place(row, actor int) uint8 { return uint8(row<<4 | actor) }
func at(frame, col int) uint8    { return uint8(frame<<5 | col) }
func vanish(actor int) uint8     { return uint8(scriptVanish | actor) }
func endFrame(n int) uint8       { return uint8(scriptEndFrame | n) }

// The view's actors and the tile each is drawn from.
const (
	actorMoongate = iota
	actorPirateShip
	actorShip
	actorSleeper1
	actorSleeper2
	actorRogue1
	actorRogue2
	actorFighter
	actorMage
	actorSeaSerpent
	actorDaemon
	actorDragon
	actorMissile
	actorHitFlash
	actorMagicSphere
)

var actorTiles = [...]uint8{
	tileMoongateOpening, tilePirateWest, tileShipWest, tileAsleep, tileAsleep, tileRogue, tileRogue,
	tileFighter, tileMage, tileSeaSerpent, tileDaemon, tileDragon, tileMissile, tileHitFlash, tileMagicSphere,
}

// Tile numbers (SHAPES.EGA) the title uses.
const (
	tileDeepWater       = 0
	tileMediumWater     = 1
	tileShallowWater    = 2
	tileShipWest        = 16
	tileMage            = 32
	tileFighter         = 36
	tileShepherd        = 46
	tileAsleep          = 56
	tileMoongateOpening = 64
	tilePoisonField     = 68
	tileEnergyField     = 69
	tileFireField       = 70
	tileSleepField      = 71
	tileLava            = 76
	tileMissile         = 77
	tileMagicSphere     = 78
	tileHitFlash        = 79
	tileGuard           = 80
	tileLordBritish     = 94
	tilePirateWest      = 128
	tileNixie           = 132
	tileTwister         = 142
	tileRat             = 144
	tileSeaSerpent      = 136
	tileRogue           = 200
	tileDaemon          = 240
	tileDragon          = 248
)

var viewScript = [...]uint8{
	endFrame(8),
	// moongate appears
	place(2, actorMoongate), at(0, 17), endFrame(0),
	place(2, actorMoongate), at(1, 17), endFrame(0),
	place(2, actorMoongate), at(2, 17), endFrame(0),
	place(2, actorMoongate), at(3, 17), endFrame(8),
	// fighter goes out, enters the castle
	place(2, actorFighter), at(0, 17), endFrame(4),
	place(2, actorFighter), at(0, 16), endFrame(4),
	place(1, actorFighter), at(0, 16), endFrame(4),
	place(0, actorFighter), at(0, 16), endFrame(4),
	vanish(actorFighter), endFrame(8),
	// moongate disappears
	place(2, actorMoongate), at(2, 17), endFrame(0),
	place(2, actorMoongate), at(1, 17), endFrame(0),
	place(2, actorMoongate), at(0, 17), endFrame(0),
	vanish(actorMoongate), endFrame(8),

	place(0, actorPirateShip), at(3, 6), endFrame(4), // boat
	place(1, actorPirateShip), at(3, 6), endFrame(4), // boat
	place(0, actorFighter), at(0, 16), endFrame(4), // fighter
	place(1, actorFighter), at(0, 16), endFrame(2), // fighter
	place(1, actorPirateShip), at(2, 6), endFrame(2), // boat
	place(1, actorFighter), at(0, 15), endFrame(2), // fighter
	place(1, actorPirateShip), at(2, 7), endFrame(2), // boat
	place(1, actorFighter), at(0, 14), endFrame(2), // fighter
	place(1, actorPirateShip), at(2, 8), place(2, actorFighter), at(0, 14), endFrame(4),
	place(1, actorPirateShip), at(2, 9), vanish(actorFighter), endFrame(4),
	place(1, actorPirateShip), at(2, 10), endFrame(8),
	place(2, actorFighter), at(0, 14), endFrame(4),
	place(1, actorFighter), at(0, 14), endFrame(4),
	place(1, actorFighter), at(0, 13), endFrame(4),
	place(2, actorMage), at(0, 14), endFrame(4),
	place(1, actorMage), at(0, 14), endFrame(4),
	place(1, actorPirateShip), at(1, 10), endFrame(0),
	place(1, actorMissile), at(0, 11), endFrame(0),
	place(1, actorMissile), at(0, 12), endFrame(0),
	vanish(actorMissile), place(1, actorHitFlash), at(0, 13), endFrame(1),
	vanish(actorHitFlash), endFrame(4),
	place(1, actorMagicSphere), at(0, 13), endFrame(0),
	place(1, actorMagicSphere), at(0, 12), endFrame(0),
	place(1, actorMagicSphere), at(0, 11), endFrame(0),
	place(1, actorMagicSphere), at(0, 10), endFrame(1),
	vanish(actorMagicSphere), endFrame(4),
	place(1, actorPirateShip), at(2, 10), endFrame(4),
	place(1, actorPirateShip), at(2, 11), place(0, actorFighter), at(0, 13), endFrame(4),
	place(1, actorPirateShip), at(2, 12), place(0, actorFighter), at(0, 14), endFrame(4),
	place(1, actorRogue1), at(0, 12), endFrame(4),
	place(1, actorRogue1), at(0, 13), endFrame(4),
	place(0, actorRogue1), at(0, 13), place(1, actorRogue2), at(0, 12), endFrame(4),
	place(1, actorRogue2), at(0, 13), endFrame(4),
	place(0, actorHitFlash), at(0, 14), endFrame(1),
	vanish(actorHitFlash), endFrame(4),
	place(1, actorHitFlash), at(0, 14), endFrame(1),
	vanish(actorHitFlash), endFrame(4),
	place(0, actorHitFlash), at(0, 13), endFrame(1),
	vanish(actorHitFlash), endFrame(4),
	place(1, actorMissile), at(0, 13), endFrame(1),
	vanish(actorMissile), endFrame(4),
	place(0, actorHitFlash), at(0, 14), endFrame(1),
	vanish(actorHitFlash), endFrame(4),
	place(1, actorHitFlash), at(0, 14), endFrame(1),
	vanish(actorHitFlash), endFrame(4),
	place(0, actorHitFlash), at(0, 13), endFrame(1),
	vanish(actorHitFlash), vanish(actorRogue1), place(0, actorSleeper1), at(0, 13), endFrame(4),
	place(1, actorHitFlash), at(0, 13), endFrame(1),
	vanish(actorHitFlash), endFrame(4),
	place(1, actorMissile), at(0, 14), endFrame(1),
	vanish(actorMissile), endFrame(4),
	place(1, actorHitFlash), at(0, 13), endFrame(1),
	vanish(actorHitFlash), vanish(actorRogue2), place(1, actorSleeper2), at(0, 13), endFrame(8),
	place(1, actorMage), at(0, 13), endFrame(4),
	place(1, actorMage), at(0, 12), place(1, actorFighter), at(0, 14), endFrame(4),
	vanish(actorMage), place(1, actorFighter), at(0, 13), endFrame(4),
	place(1, actorFighter), at(0, 12), endFrame(4),
	vanish(actorFighter), endFrame(8),
	vanish(actorPirateShip), place(1, actorShip), at(2, 12), endFrame(4),
	place(1, actorShip), at(3, 12), endFrame(4),
	place(1, actorShip), at(0, 12), endFrame(4),
	place(4, actorDragon), at(0, 0), vanish(actorSleeper1), endFrame(2),
	place(3, actorDragon), at(0, 1), endFrame(2),
	place(2, actorDragon), at(0, 2), place(1, actorShip), at(0, 11), endFrame(2),
	place(2, actorDragon), at(0, 3), endFrame(2),
	place(1, actorShip), at(0, 10), place(1, actorDragon), at(0, 4), vanish(actorSleeper2), endFrame(2),
	place(1, actorDragon), at(0, 5), endFrame(2),
	place(1, actorDragon), at(0, 6), place(1, actorShip), at(3, 10), endFrame(2),
	place(1, actorDragon), at(0, 7), endFrame(2),
	place(1, actorMissile), at(0, 9), endFrame(0),
	place(1, actorMissile), at(0, 8), endFrame(0),
	vanish(actorMissile), place(1, actorHitFlash), at(0, 7), endFrame(1),
	vanish(actorHitFlash), endFrame(8),
	place(1, actorDragon), at(0, 8), endFrame(2),
	place(1, actorMissile), at(0, 9), endFrame(0),
	vanish(actorMissile), place(1, actorHitFlash), at(0, 8), endFrame(1),
	vanish(actorHitFlash), vanish(actorDragon), endFrame(4),
	place(3, actorDaemon), at(0, 9), endFrame(4),
	place(4, actorDaemon), at(0, 9), endFrame(4),
	place(4, actorDaemon), at(0, 8), place(1, actorShip), at(0, 10), endFrame(4),
	place(1, actorShip), at(0, 9), endFrame(4),
	place(1, actorShip), at(0, 8), endFrame(4),
	place(2, actorMissile), at(0, 8), endFrame(0),
	place(3, actorMissile), at(0, 8), endFrame(0),
	vanish(actorMissile), place(4, actorHitFlash), at(0, 8), endFrame(1),
	vanish(actorHitFlash), vanish(actorDaemon), endFrame(8),
	place(1, actorShip), at(3, 8), endFrame(4),
	place(2, actorShip), at(3, 8), endFrame(4),
	place(3, actorShip), at(3, 8), endFrame(4),
	place(3, actorShip), at(0, 8), endFrame(4),
	place(3, actorShip), at(0, 7), endFrame(4),
	place(3, actorShip), at(3, 7), endFrame(4),
	place(4, actorShip), at(3, 7), endFrame(8),
	place(4, actorFighter), at(0, 7), endFrame(4),
	place(4, actorFighter), at(0, 8), endFrame(4),
	place(4, actorFighter), at(0, 9), place(4, actorMage), at(0, 7), endFrame(4),
	place(3, actorFighter), at(0, 9), place(4, actorMage), at(0, 8), endFrame(4),
	vanish(actorFighter), place(4, actorMage), at(0, 9), endFrame(4),
	place(3, actorMage), at(0, 9), endFrame(4),
	vanish(actorMage), endFrame(8),
	place(3, actorFighter), at(0, 9), endFrame(4),
	place(4, actorFighter), at(0, 9), endFrame(4),
	place(4, actorFighter), at(0, 8), place(3, actorMage), at(0, 9), endFrame(4),
	place(4, actorFighter), at(0, 7), place(4, actorMage), at(0, 9), endFrame(4),
	vanish(actorFighter), place(4, actorMage), at(0, 8), endFrame(4),
	place(4, actorMage), at(0, 7), endFrame(4),
	vanish(actorMage), endFrame(8),
	place(4, actorShip), at(0, 7), endFrame(4),
	place(4, actorShip), at(0, 6), endFrame(4),
	place(4, actorShip), at(0, 5), endFrame(4),
	place(4, actorShip), at(0, 4), endFrame(4),
	place(4, actorShip), at(1, 4), endFrame(4),
	place(3, actorShip), at(1, 4), endFrame(4),
	place(3, actorShip), at(0, 4), endFrame(4),
	place(3, actorShip), at(0, 3), endFrame(8),
	place(3, actorFighter), at(0, 3), endFrame(4),
	place(3, actorFighter), at(0, 2), endFrame(4),
	place(3, actorFighter), at(0, 1), place(3, actorMage), at(0, 3), endFrame(4),
	place(4, actorFighter), at(0, 1), place(3, actorMage), at(0, 2), endFrame(2),
	place(0, actorSeaSerpent), at(0, 9), endFrame(2),
	place(4, actorFighter), at(0, 0), place(4, actorMage), at(0, 2), place(1, actorSeaSerpent), at(0, 8), endFrame(2),
	place(2, actorSeaSerpent), at(0, 7), endFrame(2),
	place(4, actorFighter), at(0, 1), place(3, actorSeaSerpent), at(0, 6), endFrame(2),
	place(3, actorMagicSphere), at(0, 5), endFrame(0),
	place(3, actorMagicSphere), at(0, 4), endFrame(0),
	place(3, actorMagicSphere), at(0, 3), endFrame(1),
	vanish(actorMagicSphere), vanish(actorShip), endFrame(2),
	place(3, actorSeaSerpent), at(0, 5), endFrame(2),
	place(3, actorSeaSerpent), at(0, 4), place(4, actorMage), at(0, 3), place(3, actorFighter), at(0, 1), endFrame(2),
	place(3, actorSeaSerpent), at(0, 3), endFrame(2),
	place(3, actorMagicSphere), at(0, 3), place(3, actorFighter), at(0, 2), endFrame(1),
	vanish(actorMagicSphere), endFrame(3),
	place(4, actorHitFlash), at(0, 3), endFrame(1),
	vanish(actorHitFlash), vanish(actorMage), place(4, actorSleeper2), at(0, 3), endFrame(4),
	place(3, actorHitFlash), at(0, 3), endFrame(1),
	vanish(actorHitFlash), endFrame(2),
	place(3, actorHitFlash), at(0, 2), endFrame(1),
	vanish(actorHitFlash), endFrame(2),
	place(3, actorHitFlash), at(0, 3), endFrame(1),
	vanish(actorHitFlash), vanish(actorSeaSerpent), endFrame(8),
	place(4, actorFighter), at(0, 2), endFrame(4),
	place(4, actorFighter), at(0, 3), vanish(actorSleeper2), endFrame(8),
	place(2, actorMoongate), at(0, 1), endFrame(0),
	place(2, actorMoongate), at(1, 1), endFrame(0),
	place(2, actorMoongate), at(2, 1), endFrame(0),
	place(2, actorMoongate), at(3, 1), endFrame(5),
	place(4, actorFighter), at(0, 2), endFrame(4),
	place(3, actorFighter), at(0, 2), endFrame(4),
	place(2, actorFighter), at(0, 2), endFrame(4),
	place(2, actorFighter), at(0, 1), endFrame(4),
	vanish(actorFighter), endFrame(8),
	place(2, actorMoongate), at(3, 1), endFrame(0),
	place(2, actorMoongate), at(2, 1), endFrame(0),
	place(2, actorMoongate), at(1, 1), endFrame(0),
	place(2, actorMoongate), at(0, 1), endFrame(0),
	vanish(actorMoongate), endFrame(8),
	scriptRestart,
}
