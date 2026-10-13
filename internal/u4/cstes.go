package u4

// The game's names.

// seaMonsterNames are the sea monsters, two tiles each from tilePirateWest.
var seaMonsterNames = [8]string{
	"Pirate",
	"Pirate",
	"Nixie",
	"Squid",
	"Serpent",
	"Seahorse",
	"Whirlpool",
	"Twister",
}

// landMonsterNames are the land monsters, four tiles each from tileRat.
var landMonsterNames = [28]string{
	"Rat",
	"Bat",
	"Spider",
	"Ghost",
	"Slime",
	"Troll",
	"Gremlin",
	"Mimic",
	"Reaper",
	"Insects",
	"Gazer",
	"Phantom",
	"Orc",
	"Skeleton",
	"Rogue",
	"Python",
	"Ettin",
	"Headless",
	"Cyclops",
	"Wisp",
	"Mage",
	"Liche",
	"Lava Lizard",
	"Zorn",
	"Daemon",
	"Hydra",
	"Dragon",
	"Balron",
}

var weaponNames = [16]string{
	"Hands",
	"Staff",
	"Dagger",
	"Sling",
	"Mace",
	"Axe",
	"Sword",
	"Bow",
	"Crossbow",
	"Flaming Oil",
	"Halberd",
	"Magic Axe",
	"Magic Sword",
	"Magic Bow",
	"Magic Wand",
	"Mystic Sword",
}

var armorNames = [8]string{
	"Skin",
	"Cloth",
	"Leather",
	"Chain Mail",
	"Plate Mail",
	"Magic Chain",
	"Magic Plate",
	"Mystic Robe",
}

// weaponCodes are the weapons' short names, for the party's status.
var weaponCodes = [16]string{
	"HND",
	"STF",
	"DAG",
	"SLN",
	"MAC",
	"AXE",
	"SWD",
	"BOW",
	"XBO",
	"OIL",
	"HAL",
	"+AX",
	"+SW",
	"+BO",
	"WND",
	"^SW",
}

// peopleNames are the classes, then the other people: two tiles each from tileMage, then from tileGuard.
var peopleNames = [16]string{
	"Mage",
	"Bard",
	"Fighter",
	"Druid",
	"Tinker",
	"Paladin",
	"Ranger",
	"Shepherd",
	"Guard",
	"Merchant",
	"Bard",
	"Jester",
	"Beggar",
	"Child",
	"Bull",
	"Lord British",
}

var reagentNames = [8]string{
	"Sulfur Ash",
	"Ginseng",
	"Garlic",
	"Spider Silk",
	"Blood Moss",
	"Black Pearl",
	"Nightshade",
	"Mandrake",
}

// spellNames are the spells, A to Z.
var spellNames = [26]string{
	"Awaken",
	"Blink",
	"Cure",
	"Dispell",
	"Energy",
	"Fireball",
	"Gate",
	"Heal",
	"Iceball",
	"Jinx",
	"Kill",
	"Light",
	"Magic Misl",
	"Negate",
	"Open",
	"Protection",
	"Quickness",
	"Resurrect",
	"Sleep",
	"Tremor",
	"Undead",
	"View",
	"Winds",
	"X-it",
	"Y-up",
	"Z-down",
}

// placeNames are the places, by location from 1.
var placeNames = [24]string{
	"Britannia",
	"The Lycaeum",
	"Empath Abbey",
	"Serpents Hold",
	"Moonglow",
	"Britain",
	"Jhelom",
	"Yew",
	"Minoc",
	"Trinsic",
	"Skara Brae",
	"Magincia",
	"Paws",
	"Buccaneers Den",
	"Vesper",
	"Cove",
	"Deceit",
	"Despise",
	"Destard",
	"Wrong",
	"Covetous",
	"Shame",
	"Hythloth",
	"The Great Stygian Abyss!",
}

var virtueNames = [8]string{
	"Honesty",
	"Compassion",
	"Valor",
	"Justice",
	"Sacrifice",
	"Honor",
	"Spirituality",
	"Humility",
}

// villagerName is the name of whoever is in a town but not in peopleNames.
const villagerName = "Villager"
