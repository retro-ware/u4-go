package u4

// Talking to the citizens (their .TLK records), and the shops behind the counters.

import (
	"strings"

	"github.com/memmaker/u4-go/internal/music"
)

func sayNoResponse() { printText("Funny, no response!\n") }

// talkRecordSize is the size of a citizen's record in a .TLK file.
const talkRecordSize = 288

// citizen is a citizen's talk record.
type citizen struct {
	askAfter int  // the topic (index into topics) after which the citizen asks a question
	humble   bool // the question tests humility
	turnAway int  // how likely the citizen turns away (out of 256); a lot more: attacks
	name     string
	pronoun  string
	look     string
	job      string
	health   string
	answer   [2]string // the answers to the two keywords of their own
	question string
	yes, no  string // the replies to the answers to the question
	keyword  [2]string
}

// talking is the citizen being talked to, talkingTo their people slot.
var (
	talking   citizen
	talkingTo int
)

// readCitizen reads a citizen's record: three numbers, then twelve zero-terminated texts.
func readCitizen(record []uint8) citizen {
	c := citizen{askAfter: signedByte(record[0]), humble: record[1] != 0, turnAway: signedByte(record[2])}
	texts := make([]string, 12)
	rest := record[3:]
	for i := range texts {
		end := 0
		for end < len(rest) && rest[end] != 0 {
			end++
		}
		texts[i] = string(rest[:end])
		rest = rest[min(end+1, len(rest)):]
	}
	c.name, c.pronoun, c.look, c.job, c.health = texts[0], texts[1], texts[2], texts[3], texts[4]
	c.answer = [2]string{texts[5], texts[6]}
	c.question, c.yes, c.no = texts[7], texts[8], texts[9]
	c.keyword = [2]string{strings.TrimRight(texts[10], " "), strings.TrimRight(texts[11], " ")}
	if c.look != "" { // "You see a ..." from "A ..."
		if first := c.look[0]; first >= 'A' && first <= 'Z' {
			c.look = string(first+'a'-'A') + c.look[1:]
		}
	}
	return c
}

// The topics, in the order a citizen's askAfter counts them.
const (
	topicBye = iota
	topicName
	topicLook
	topicJob
	topicHealth
	topicKeyword1
	topicKeyword2
	topicJoin
	topicGive
)

// topicWord is the word that asks after a topic: four letters of it are enough.
func topicWord(topic int) string {
	switch topic {
	case topicBye:
		return "bye"
	case topicName:
		return "name"
	case topicLook:
		return "look"
	case topicJob:
		return "job"
	case topicHealth:
		return "health"
	case topicKeyword1:
		return talking.keyword[0]
	case topicKeyword2:
		return talking.keyword[1]
	case topicJoin:
		return "join"
	case topicGive:
		return "give"
	}
	return ""
}

// The karma a citizen's question moves (once per talkKarmaMoves moves).
const (
	humilityCost   = 5
	humilityReward = 10
	charityReward  = 2
	talkKarmaMoves = 16
)

// karmaTalkWindow is the stretch of moves the party is in: karma from talk rises once in it.
func karmaTalkWindow() uint32 { return party.moves / talkKarmaMoves }

// askQuestion asks the citizen's question, which wants a yes or a no.
func askQuestion() {
	newLine()
	readKey()
	printText(talking.question)
	printText("\n\nYou say: ")
	var answer byte
	for {
		word := readWord(3)
		newLine()
		if word == "" {
			return
		}
		answer = strings.ToUpper(word)[0]
		if answer == 'Y' || answer == 'N' {
			break
		}
		printText("Yes or no!\n")
	}
	if answer == 'Y' {
		if talking.humble {
			lowerKarma(&party.karmaHumility, humilityCost)
		}
		printText(talking.yes)
	} else {
		if talking.humble && karmaTalkWindow() != uint32(party.lastKarmaTalk) {
			raiseKarma(&party.karmaHumility, humilityReward)
		}
		printText(talking.no)
	}
	party.lastKarmaTalk = uint16(karmaTalkWindow())
	newLine()
}

func printLine(text string) {
	printText(text)
	newLine()
}

func talkName() {
	printText(talking.pronoun)
	printText(" says: I am ")
	printLine(talking.name)
}

func talkLook() {
	printText("You see ")
	printLine(talking.look)
}

// recruitTownes are the townes whose citizens may join: each of a class, mage to shepherd.
const recruitTownes = locFirstTowne

// virtueAdjectives are the virtues as said of a person.
var virtueAdjectives = [8]string{
	"honest",
	"compassionate",
	"valiant",
	"just",
	"sacrificial",
	"honorable",
	"spiritual",
	"humble",
}

// The joining: a companion asks for half the virtue's karma (or the virtue done) and
// 100 more maximum hit points of the leader per member.
const (
	joinKarma       = 40
	joinHPPerMember = 100
	recruitTalkIdx  = 1 // only the citizen with the first talk record of a towne may join
	lastPersonSlot  = 31
)

func talkJoin() {
	class := int(party.loc) - recruitTownes
	if area.people.talkIndex[talkingTo] != recruitTalkIdx ||
		class < 0 || class >= len(karmas) ||
		class == int(signedByte(party.members[0].class)) {
		printText(talking.pronoun)
		printText(" says: I cannot join thee.\n")
		return
	}
	if *karmas[class] < joinKarma && *karmas[class] != 0 {
		printText("Thou art not ")
		printText(virtueAdjectives[class])
		printText(" enough for me to join thee.\n")
		return
	}
	if joinHPPerMember*party.size+joinHPPerMember > party.members[0].maxHP {
		printText("Thou art not experienced enough for me to join thee.\n")
		return
	}
	printText("I am honored to join thee!\n")
	// Quirk: the original clears the last people slot, not the citizen who joined.
	area.people.tile[lastPersonSlot] = 0
	area.people.baseTile[lastPersonSlot] = 0
	area.people.behaviour[lastPersonSlot] = 0
	area.people.talkIndex[lastPersonSlot] = 0
	redraw()
	for m := len(party.members) - 1; m >= 0; m-- {
		if signedByte(party.members[m].class) == class {
			swapMembers(int(party.size), m)
			break
		}
	}
	party.size++
	clearStatus()
	showStats()
}

func talkGive() {
	if area.people.tile[talkingTo] != tileBeggar {
		printText(talking.pronoun)
		printText(" says: I do not need thy gold.  Keep it!\n")
		return
	}
	printText("How much?" + answerRoom(3))
	amount := int16(askNumber(2))
	if amount <= 0 {
		return
	}
	if party.gold < uint16(amount) {
		printText("Thou hast not that much gold!\n")
		return
	}
	party.gold -= uint16(amount)
	showStats()
	printText(talking.pronoun)
	printText(" says: Oh, Thank thee! I shall never forget thy kindness!\n")
	if karmaTalkWindow() != uint32(party.lastKarmaTalk) {
		raiseKarma(&party.karmaCompassion, charityReward)
	}
	party.lastKarmaTalk = uint16(karmaTalkWindow())
}

// answerTopic answers a topic; false to end the talk.
func answerTopic(topic int) bool {
	switch topic {
	case topicBye:
		return false
	case topicName:
		talkName()
	case topicLook:
		talkLook()
	case topicJob:
		printLine(talking.job)
	case topicHealth:
		printLine(talking.health)
	case topicKeyword1:
		printLine(talking.answer[0])
	case topicKeyword2:
		printLine(talking.answer[1])
	case topicJoin:
		talkJoin()
	case topicGive:
		talkGive()
	}
	return true
}

// attackOffence is how far past the turning-away chance a citizen attacks instead.
const attackOffence = 64

// talkTo talks to the citizen in people slot person.
func talkTo(person int) {
	record := (int(area.people.talkIndex[person]) - 1) * talkRecordSize
	talking = readCitizen(at(talkFile, int(uint16(record))))
	talkingTo = person
	members := int(party.size)
	foeTile = area.people.tile[person]
	printText("\nYou meet ")
	printLine(talking.look)
	if randomBits(1) != 0 {
		newLine()
		talkName()
	}
	for {
		printText("\nYour Interest:\n")
		getsWords = talkWords
		asked := readWord(10)
		newLine()
		if asked == "" {
			break
		}
		if chance := int(randomByte()); chance < talking.turnAway {
			if talking.turnAway-chance >= attackOffence {
				// A titled citizen ("a ...", "the ...") is named by pronoun.
				lower := strings.ToLower(talking.name)
				if !strings.HasPrefix(lower, "a ") && !strings.HasPrefix(lower, "the ") {
					printText(talking.name)
				} else {
					printText(talking.pronoun)
				}
				printText(" says: On guard! Fool!\n")
				area.people.behaviour[person] = behaviourAttack
			} else {
				printText(talking.pronoun)
				printText(" turns away!\n\n")
			}
			return
		}
		// The topics are tried in order; an empty keyword of their own ends the list.
		topic := topicBye
		for ; topic <= topicGive && topicWord(topic) != ""; topic++ {
			if sameWord(asked, topicWord(topic), 4) {
				break
			}
		}
		if topic > topicGive || topicWord(topic) == "" {
			printText("That I cannot help thee with.\n")
			continue
		}
		if topic == topicBye {
			break
		}
		newLine()
		answerTopic(topic)
		if int(party.size) != members {
			return // they joined
		}
		if talking.askAfter == topic {
			askQuestion()
		}
	}
	printText("\nBye.\n")
}

// The shops, in the order of the towns' counters.
const (
	shopWeapons = iota
	shopArmour
	shopFood
	shopPub
	shopReagents
	shopHealer
	shopInn
	shopGuild
	shopHorses
	shopHawkwind
)

// shopCounters is the row of each shop's counter in each town (0: none).
var shopCounters = [16][8]uint8{
	locCastleBritannia - 1: {shopHealer: 26},
	locLycaeum - 1:         {shopHealer: 12},
	locEmpathAbbey - 1:     {shopHealer: 15},
	locSerpentsHold - 1:    {shopHealer: 12},
	locMoonglow - 1:        {shopFood: 14, shopReagents: 26, shopHealer: 27, shopInn: 2},
	locBritain - 1:         {3, 7, 6, 2, 0, 29, 12, 0},
	locJhelom - 1:          {9, 5, 0, 19, 0, 6, 26, 0},
	locYew - 1:             {shopFood: 24, shopHealer: 25},
	locMinoc - 1:           {shopWeapons: 28, shopInn: 3},
	locTrinsic - 1:         {shopWeapons: 20, shopArmour: 24, shopPub: 2, shopInn: 3},
	locSkaraBrae - 1:       {shopFood: 17, shopReagents: 4, shopHealer: 27, shopInn: 13},
	locMagincia - 1:        {},
	locPaws - 1:            {shopArmour: 4, shopFood: 26, shopPub: 5, shopReagents: 7},
	locBuccaneersDen - 1:   {11, 17, 0, 25, 8, 0, 0, 7},
	locVesper - 1:          {shopWeapons: 20, shopPub: 22, shopInn: 21, shopGuild: 26},
	locCove - 1:            {shopHealer: 26},
}

// The two shops not in the table: the horses of Paws and Hawkwind at the castle.
const (
	pawsHorsesRow  = 24
	britainHawkRow = 25
)

var shops = [10]func(){
	shopWeapons:  visitWeaponShop,
	shopArmour:   visitArmourShop,
	shopFood:     visitFoodShop,
	shopPub:      visitPub,
	shopReagents: visitReagentShop,
	shopHealer:   visitHealer,
	shopInn:      visitInn,
	shopGuild:    visitGuild,
	shopHorses:   visitHorseSeller,
	shopHawkwind: visitHawkwind,
}

// visitShop opens the shop whose counter is on row.
func visitShop(row uint8) {
	shop := len(shopCounters[0]) - 1
	for shop >= 0 && shopCounters[party.loc-1][shop] != row {
		shop--
	}
	if shop == -1 && row == pawsHorsesRow && party.loc == locPaws {
		shop = shopHorses
	}
	if shop == -1 && row == britainHawkRow && party.loc == locCastleBritannia {
		shop = shopHawkwind
	}
	if shop == -1 {
		sayNoResponse()
		return
	}
	music.Play(music.Shopping)
	shops[shop]()
}

// isLetterTile says whether the tile is one of the letters of a shop's counter sign.
func isLetterTile(t uint8) bool { return t >= tileLetterA && t <= tileBlank }

func cmdTalk() {
	printText("Talk\n")
	if party.flyingOrLight != 0 {
		sayDriftOnly()
		return
	}
	dx, dy := askDirection("Dir: ")
	if dx|dy == 0 {
		return
	}
	x := uint8(int(party.x) + dx)
	y := uint8(int(party.y) + dy)
	person := objectAt(x, y)
	if x >= townTiles || y >= townTiles || party.loc == 0 {
		sayNoResponse()
		return
	}
	if isLetterTile(*area.townTile(int(y), int(x))) { // talk across the counter
		x = uint8(int(x) + dx)
		y = uint8(int(y) + dy)
		if x >= townTiles || y >= townTiles {
			sayNoResponse()
			return
		}
		person = objectAt(x, y)
		if person == -1 || area.people.tile[person] != tileMerchant {
			sayNoResponse()
			return
		}
		visitShop(uint8(int(y) - dy))
		return
	}
	if person == -1 || area.people.talkIndex[person] == 0 {
		sayNoResponse()
		return
	}
	if area.people.tile[person] == tileLordBritish {
		music.Play(music.LordBritish)
		talkToLordBritish()
		return
	}
	talkTo(person)
}
