package u4

// The shops: Hawkwind the seer, the reagent, weapon and armour shops, the horses, the
// guilds, the inns, the healers, the pubs and the food shops.

import (
	"slices"
	"strings"
)

// shopFront is a shop's town, name and keeper.
type shopFront struct {
	loc    uint16
	name   string
	keeper string
}

func (s shopFront) town() uint16 { return s.loc }

// findShop is the index of the shop of the party's town among shops.
func findShop[S interface{ town() uint16 }](shops []S) int {
	for i, s := range shops {
		if s.town() == party.loc {
			return i
		}
	}
	panic("no shop of this kind in this town")
}

// The most of anything the party can carry.
const (
	maxStock = 99
	maxFood  = 999900 // 9999 rations, in hundredths
	ration   = 100
)

// leaderName is the party leader's name.
func leaderName() string { return cstr(party.members[0].name[:]) }

// askAmount reads an amount of up to digits digits; -1 when nothing was typed.
func askAmount(digits int) int { return int(askNumber(digits)) }

// ---- Hawkwind the seer ----

// hawkwindInsights are the seer's words on each virtue, by how far along its path the
// party is: karma below 20, 40, 60 and 99, and 99.
var hawkwindInsights = [5][8]string{
	{ /* < 20*/
		"Thou art a thief and a scoundrel. Thou may not ever become an Avatar!",
		"Thou art a cold and cruel brute.  Thou shouldst go to prison for thy crimes!",
		"Thou art a coward, thou dost flee from the hint of danger!",
		"Thou art an unjust wretch. Thou are a fulsome meddler!",
		"Thou art a self-serving Tufthunter. Thou deservest not my help, yet I grant it!",
		"Thou art a cad and a bounder. Thy presence is an affront. Thou art low as a slug!",
		"Thy spirit is weak and feeble. Thou dost not strive for Perfection!",
		"Thou art proud and vain. All other virtue in thee is a loss!",
	}, { /* < 40*/
		"Thou art not an honest soul. Thou must live a more honest life to be an Avatar!",
		"Thou dost kill where there is no need and give too little unto others!",
		"Thou dost not display a great deal of Valor. Thou dost flee before the need!",
		"Thou art cruel and unjust. In time thou will suffer for thy crimes!",
		"Thou dost need to think more of the life of others and less of thy own!",
		"Thou dost not fight with honor but with malice and deceit!",
		"Thou dost not take time to care about thy inner being, a must to be an Avatar!",
		"Thou art too proud of thy little deeds. Humility is the root of all Virtue!",
	}, { /* < 60*/
		"Thou hast made little progress on the paths of Honesty. Strive to prove thy worth!",
		"Thou hast not shown thy compassion well. Be more kind unto others!",
		"Thou art not yet a valiant warrior.  Fight to defeat evil and prove thyself!",
		"Thou hast not proven thyself to be just. Strive to do justice unto all things!",
		"Thy sacrifice is small. Give of thy life's blood so that others may live.",
		"Thou dost need to show thyself to be more honorable.  The path lies before thee!",
		"Strive to know and master more of thine inner being. Meditation lights the path!",
		"Thy progress on this path is most uncertain. Without Humility thou art empty!",
	}, { /* < 99*/
		"Thou dost seem to be an honest soul.  Continued honesty will reward thee!",
		"Thou dost show thy compassion well.  Continued goodwill should be thy guide!",
		"Thou art showing Valor in the face of danger. Strive to become yet more so!",
		"Thou dost seem fair and just. Strive to uphold Justice even more sternly!",
		"Thou art giving of thyself in some ways. Seek ye now to find yet more!",
		"Thou dost seem to be Honorable in nature.  Seek to bring Honor upon others as well!",
		"Thou art doing well on the path to inner sight continue to seek the inner light!",
		"Thou dost seem a humble soul.  Thou art setting strong stones to build virtues upon!",
	}, { /* >= 99*/
		"Thou art truly an honest soul. Seek ye now to reach Elevation!",
		"Compassion is a virtue that thou hast shown well.  Seek ye now Elevation!",
		"Thou art a truly valiant warrior. Seek ye now Elevation in the virtue of valor!",
		"Thou art just and fair.  Seek ye now the Elevation!",
		"Thou art giving and good.  Thy self-sacrifice is great.  Seek now Elevation!",
		"Thou hast proven thyself to be Honorable. Seek ye now for the Elevation!",
		"Spirituality is in thy nature. Seek ye now the Elevation!",
		"Thy Humility shines bright upon thy being. Seek ye now for Elevation!",
	},
}

// insightBand is how far along a virtue's path karma is, 0 to 4.
func insightBand(karma uint16) int {
	switch {
	case karma < 20:
		return 0
	case karma < 40:
		return 1
	case karma < 60:
		return 2
	case karma < avatarhoodKarma:
		return 3
	}
	return 4
}

// The seer's visit raises spirituality, once per meditation's while.
const seerKarma = 3

func visitHawkwind() {
	if !isConscious(0) {
		printText("The Seer says: I will speak only with ")
		printText(leaderName())
		printText(".\nReturn when ")
		printText(leaderName())
		printText(" is revived!\n")
		return
	}
	printText("\n\nWelcome, ")
	printText(leaderName())
	printText("\nI am Hawkwind, Seer of Souls. I see that which is within thee and drives thee to deeds of good or evil...\n")
	flushKeys()
	readKey()
	asked := ""
	for first := true; ; first = false {
		if first || asked == "" {
			printText("\nFor what path dost thou seek enlightenment?\n")
		} else {
			printText("\nHawkwind asks: What other path seeks clarity?\n")
		}
		asked = readWord(14)
		if asked == "" || strings.EqualFold(asked, "none") || strings.EqualFold(asked, "bye") {
			break
		}
		virtue := len(virtueNames) - 1
		for virtue >= 0 && !sameWord(asked, virtueNames[virtue], 4) {
			virtue--
		}
		if virtue == -1 {
			printText("\nHe says: That is not a subject for enlightenment.\n")
			continue
		}
		newLine()
		karma := *karmas[virtue]
		if karma == 0 {
			printText("He says:\nThou hast become a partial Avatar in that attribute. Thou need not my insights.\n")
			continue
		}
		band := insightBand(karma)
		newLine()
		printText(hawkwindInsights[band][virtue])
		newLine()
		if band == 4 {
			printText("\nGo to the Shrine and meditate for three Cycles!\n")
			flushKeys()
			readKey()
		}
	}
	printText("\nHawkwind says: Fare thee well and may thou complete the Quest of the Avatar!\n")
	if party.moves/meditationMoves != uint32(party.lastMeditation) {
		raiseKarma(&party.karmaSpirituality, seerKarma)
		party.lastMeditation = uint16(party.moves / meditationMoves)
	}
}

// ---- Reagents ----

type reagentShop struct {
	shopFront
	prices [6]int // of the first six reagents: nightshade and mandrake are not sold
}

var reagentShops = []reagentShop{
	{shopFront{locMoonglow, "Magical Herbs", "Margot"}, [6]int{2, 5, 6, 3, 6, 9}},
	{shopFront{locSkaraBrae, "Herbs and Spice", "Sasha"}, [6]int{2, 4, 9, 6, 4, 8}},
	{shopFront{locPaws, "The Magics", "Shiela"}, [6]int{3, 4, 2, 9, 6, 7}},
	{shopFront{locBuccaneersDen, "Magic Mentar", "Shannon"}, [6]int{6, 7, 9, 9, 9, 1}},
}

// Paying the blind woman less than the price costs honesty, justice and honor: a third of
// the shortfall, at least cheatMinimum; paying in full raises them.
const (
	cheatMinimum   = 4
	cheatThreshold = 12
	fairPayKarma   = 2
)

// cheatKarma is the karma lost by paying short of the price.
func cheatKarma(short int) int {
	if short < cheatThreshold {
		return cheatMinimum
	}
	return short / 3
}

func visitReagentShop() {
	shop := reagentShops[findShop(reagentShops)]
	printText("A blind woman turns to you and says: Welcome to ")
	printText(shop.name)
	clearStatus()
	showReagents()
	printText("\n\nI am ")
	printText(shop.keeper)
	printText("\nAre you in need of Reagents?" + answerRoom(2))
	if askYesNo() == 'Y' {
		printText("Very well,\n")
		for {
			printText("I have\nA-Sulfurous Ash\nB-Ginseng\nC-Garlic\nD-Spider Silk\nE-Blood Moss\nF-Black Pearl\n")
			reagent := askLetter("Your Interest:"+answerRoom(2), 'A', 'F') - 'A'
			if reagent < 0 {
				break
			}
			price := shop.prices[reagent]
			printText("\nVery well, we sell ")
			printText(reagentNames[reagent])
			printText(" for ")
			printNumber(int32(price), 1, '0')
			printText("gp. How many would you like?" + answerRoom(3))
			if count := askAmount(2); count > 0 {
				buyReagents(reagent, count, price)
			} else {
				printText("\nI see, then\n")
			}
			printText("Anything else?" + answerRoom(2))
			if askYesNo() != 'Y' {
				break
			}
		}
	}
	clearStatus()
	clearStatusTitle()
	showStats()
	newLine()
	printText(shop.keeper)
	printText(" says: Perhaps another time then.... and slowly turns away.\n")
}

// buyReagents: the blind woman takes what the party says it pays.
func buyReagents(reagent, count, price int) {
	printText("\nVery good, that will be ")
	printNumber(int32(count)*int32(price), 1, '0')
	printText("gp. You pay:" + answerRoom(4))
	paid := askAmount(3)
	if paid <= 0 {
		if paid < 0 {
			sound(soundBeep)
		}
		return
	}
	if paid < count*price {
		lost := cheatKarma(count*price - paid)
		lowerKarma(&party.karmaHonesty, lost)
		lowerKarma(&party.karmaJustice, lost)
		lowerKarma(&party.karmaHonor, lost)
	}
	if uint16(paid) > party.gold {
		printText("It seems you have not the gold!\n")
		return
	}
	raiseKarma(&party.karmaHonesty, fairPayKarma)
	raiseKarma(&party.karmaJustice, fairPayKarma)
	raiseKarma(&party.karmaHonor, fairPayKarma)
	party.gold -= uint16(paid)
	showGold()
	party.reagents[reagent] = min(party.reagents[reagent]+uint16(count), maxStock)
	showReagents()
	printText("\nVery good.\n")
}

// ---- Weapons and armour ----

var weaponPrices = [16]int{
	weaponHands:       0,
	weaponStaff:       20,
	weaponDagger:      2,
	weaponSling:       25,
	weaponMace:        100,
	weaponAxe:         225,
	weaponSword:       300,
	weaponBow:         250,
	weaponCrossbow:    600,
	weaponOil:         5,
	weaponHalberd:     350,
	weaponMagicAxe:    1500,
	weaponMagicSword:  2500,
	weaponMagicBow:    2000,
	weaponMagicWand:   5000,
	weaponMysticSword: 7000, // never sold
}

var armorPrices = [8]int{
	armorSkin:       0,
	armorCloth:      50,
	armorLeather:    200,
	armorChainMail:  600,
	armorPlateMail:  2000,
	armorMagicChain: 4000,
	armorMagicPlate: 7000,
	armorMysticRobe: 9000, // never sold
}

// weaponPitches are what the weapon smiths say of each weapon.
var weaponPitches = [16]string{
	weaponStaff:      "We are the only staff makers in Britannia, yet sell them for only 20gp.\n",
	weaponDagger:     "We sell the most deadly of daggers, a bargain at only 2gp each.\n",
	weaponSling:      "Our slings are made from only the finest gut and leather, 'Tis yours for 25gp.\n",
	weaponMace:       "These maces have a hardened shaft and a 5lb head fairly priced at 100gp.\n",
	weaponAxe:        "Notice the fine workmanship on this axe, you'll agree 225gp is a good price.\n",
	weaponSword:      "The fine work on these swords will be the dread of thy foes, for 300gp.\n",
	weaponBow:        "Our bows are made of finest yew, and the arrows willow, a steal at 250gp.\n",
	weaponCrossbow:   "Crossbows made by Iolo the Bard are the finest in the world, yours for 600gp.\n",
	weaponOil:        "Flasks of oil make great weapons and creates a wall of flame too. 5gp each.\n",
	weaponHalberd:    "A Halberd is a mighty weapon to attack over obstacles; a must and only 350gp.\n",
	weaponMagicAxe:   "This magical axe can be thrown at thy enemy and will then return all for 1500gp.\n",
	weaponMagicSword: "Magical swords such as these are rare indeed I will part with one for 2500gp.\n",
	weaponMagicBow:   "A magical bow will keep thy enemies far away or dead! A must for 2000gp!\n",
	weaponMagicWand:  "This magic wand casts mighty blue bolts to strike down thy foes, 5000gp.\n",
}

// armorPitches are what the armourers say of each armour.
var armorPitches = [8]string{
	armorCloth:      "Cloth Armour is good for a tight budget, Fairly priced at 50gp.\n",
	armorLeather:    "Leather Armour is both supple and strong, and costs a mere 200gp.\nA Bargain!\n",
	armorChainMail:  "Chain Mail is the armour used by more warriors than all others. Ours costs 600gp.\n",
	armorPlateMail:  "Full Plate armour is the ultimate in non-magical armour. Get yours for 2000gp.\n",
	armorMagicChain: "Magic Armour is rare and expensive.  This chain sells for 4000gp.\n",
	armorMagicPlate: "Magic Plate Armour is the best known protection. Only we have it. Cost: 7000gp.\n",
}

type weaponShop struct {
	shopFront
	stock [4]int
}

var weaponShops = []weaponShop{
	{shopFront{locBritain, "Windsor Weaponry", "Winston"}, [4]int{weaponStaff, weaponDagger, weaponSling, weaponSword}},
	{shopFront{locJhelom, "Willard's Weaponry", "Willard"}, [4]int{weaponAxe, weaponSword, weaponCrossbow, weaponHalberd}},
	{shopFront{locMinoc, "The Iron Works", "Peter"}, [4]int{weaponMace, weaponHalberd, weaponMagicAxe, weaponMagicSword}},
	{shopFront{locTrinsic, "Duelling Weapons", "Jumar"}, [4]int{weaponMace, weaponAxe, weaponSword, weaponBow}},
	{shopFront{locBuccaneersDen, "Hook's Arms", "Hook"}, [4]int{weaponCrossbow, weaponOil, weaponMagicBow, weaponMagicWand}},
	{shopFront{locVesper, "Village Arms", "Wendy"}, [4]int{weaponDagger, weaponSling, weaponBow, weaponOil}},
}

type armorShop struct {
	shopFront
	stock [4]int // 0: an empty slot
}

var armorShops = []armorShop{
	{shopFront{locBritain, "Windsor Armour", "Winston"}, [4]int{armorCloth, armorLeather, armorChainMail}},
	{shopFront{locJhelom, "Valiant's Armour", "Valiant"}, [4]int{armorChainMail, armorPlateMail, armorMagicChain, armorMagicPlate}},
	{shopFront{locTrinsic, "Duelling Armour", "Jean"}, [4]int{armorCloth, armorChainMail, armorMagicChain}},
	{shopFront{locPaws, "Light Armour", "Pierre"}, [4]int{armorCloth, armorLeather}},
	{shopFront{locBuccaneersDen, "Basic Armour", "Limpy"}, [4]int{armorCloth, armorLeather, armorChainMail}},
}

// halfPrice is what a shop gives back for count items of a price.
func halfPrice(count, price int) int32 { return int32(count) * int32(price) / 2 }

// takeGold adds gold to the party's, up to maxGold.
func takeGold(amount int32) { party.gold = uint16(min(int32(party.gold)+amount, maxGold)) }

// askStocked asks for one of the shop's items by letter (from B to last), sounding for
// any it does not have; the item, or -1.
func askStocked(question string, last int, stock [4]int) int {
	for {
		item := askLetter(question, 'B', last) - 'A'
		if item <= 0 || slices.Contains(stock[:], item) {
			return item
		}
		sound(soundBeep)
	}
}

func buyWeapons(keeper string, stock [4]int) {
	newLine()
	printText("\n\nVery Good!\n\n")
	for {
		printText("We have:\n")
		for _, w := range stock {
			printChar(w + 'A')
			printChar('-')
			printText(weaponNames[w])
			printChar('s')
			newLine()
		}
		w := askStocked("Your Interest?"+answerRoom(2), 'O', stock)
		if w < 0 {
			break
		}
		if uint16(weaponPrices[w]) > party.gold {
			printText("You have not the funds for even one!\n")
		} else {
			newLine()
			printText(weaponPitches[w])
			newLine()
			printText("How many would you like?" + answerRoom(3))
			switch count := askAmount(2); {
			case count > 0:
				sellWeapon(keeper, w, count)
			case count == 0:
				printText("\nToo bad.\n")
			default:
				sound(soundBeep)
			}
		}
		printText("\nAnything else?" + answerRoom(2))
		if askYesNo() != 'Y' {
			break
		}
	}
	newLine()
}

// sellWeapon sells count of the weapon to the party, if it can pay for them all.
func sellWeapon(keeper string, w, count int) {
	cost := uint16(weaponPrices[w] * count) // 16 bits, as the original: big orders wrap
	if cost > party.gold {
		printText("I fear you have not the funds, perhaps something else.\n")
		return
	}
	party.gold -= cost
	showGold()
	party.weapons[w] = min(party.weapons[w]+uint16(count), maxStock)
	showWeapons()
	newLine()
	printText(keeper)
	printText(" says: A fine choice!\n")
}

// buyBackWeapons buys the party's weapons back at half price.
func buyBackWeapons() {
	printText("\nExcellent! Which wouldst\n")
	answer := 0 // quirk: kept from the last deal when the party offers what it lacks
	for {
		textRow, textCol = promptRow, promptCol
		w := askLetter("You sell:"+answerRoom(2), 'B', 'P') - 'A'
		newLine()
		if w <= 0 {
			break
		}
		if party.weapons[w] == 0 {
			printText("Thou dost not own that. What else might\n")
		} else {
			count := 1
			if party.weapons[w] > 1 {
				printText("How many ")
				printText(weaponNames[w])
				printText("s would you wish to sell?" + answerRoom(3))
				count = askAmount(2)
				if count <= 0 {
					if count == 0 {
						printText("\nToo bad.\n")
					} else {
						sound(soundBeep)
					}
					break
				}
				if party.weapons[w] < uint16(count) {
					printText("You don't have that many swine!\n")
					break
				}
				printText("I will give you ")
				printNumber(halfPrice(count, weaponPrices[w]), 1, '0')
				printText("gp for ")
				printText(themOrIt(count))
			} else {
				printText("I will give you ")
				printNumber(halfPrice(1, weaponPrices[w]), 1, '0')
				printText("gp for that ")
				printText(weaponNames[w])
			}
			printText("\nDeal?" + answerRoom(2))
			answer = askYesNo()
			if answer < 'N' { // a blank or escape
				break
			}
			if answer == 'N' {
				printText("Hmmph. What else would\n")
			} else {
				party.weapons[w] -= uint16(count)
				takeGold(halfPrice(count, weaponPrices[w]))
				showGold()
				showWeapons()
				printText("\nFine! What else?\n")
			}
		}
		if answer != 'Y' && answer != 'N' {
			break
		}
	}
}

func themOrIt(count int) string {
	if count > 1 {
		return "them."
	}
	return "it."
}

// askBuyOrSell reads the Buy or Sell answer: 'b', 's' or 0 to leave.
func askBuyOrSell() int {
	for {
		askKeys("", item("Buy", key('b')), item("Sell", key('s')), item("Leave", keyEsc))
		switch int(int8(readKey())) {
		case 'B', 'b':
			return 'b'
		case 'S', 's':
			return 's'
		case ' ', '\n', '\r', 27:
			newLine()
			return 0
		}
		sound(soundBeep)
	}
}

func visitWeaponShop() {
	shop := weaponShops[findShop(weaponShops)]
	clearStatus()
	showWeapons()
	printText("\n\n\n\nWelcome to ")
	printText(shop.name)
	printText("\n\n")
	printText(shop.keeper)
	printText(" says: Welcome friend! Art thou here to Buy or Sell? ")
	if choice := askBuyOrSell(); choice != 0 {
		printChar(choice)
		newLine()
		if choice == 'b' {
			buyWeapons(shop.keeper, shop.stock)
		} else {
			buyBackWeapons()
		}
	}
	clearStatusTitle()
	clearStatus()
	showStats()
	newLine()
	printText(shop.keeper)
	printText(" says: Fare thee well!\n")
}

func buyArmour(keeper string, stock [4]int) {
	printText("\n\n\nWell then,\n")
	for {
		printText("We've got:\n")
		for _, a := range stock {
			if a != armorSkin {
				printChar(a + 'A')
				printChar(' ')
				printText(armorNames[a])
				newLine()
			}
		}
		newLine()
		a := askStocked("What'll it be?"+answerRoom(2), 'G', stock)
		if a < 0 {
			break
		}
		if uint16(armorPrices[a]) > party.gold {
			printText("What you try'n to pull? You can't pay!\n")
		} else {
			newLine()
			printText(armorPitches[a])
			newLine()
			printText("How many would you like?" + answerRoom(3))
			if count := askAmount(2); count > 0 {
				sellArmour(keeper, a, count)
			} else {
				printText("\nToo bad.\n")
			}
		}
		printText("\nAnything else?" + answerRoom(2))
		if askYesNo() != 'Y' {
			break
		}
	}
}

// sellArmour sells count of the armour to the party, if it can pay for them all.
func sellArmour(keeper string, a, count int) {
	cost := uint16(armorPrices[a] * count) // 16 bits, as the original: big orders wrap
	if cost > party.gold {
		printText("You don't have enough gold.\nMaybe something cheaper?\n")
		return
	}
	party.gold -= cost
	showGold()
	party.armors[a] = min(party.armors[a]+uint16(count), maxStock)
	showArmour()
	newLine()
	printText(keeper)
	printText(" says: Good choice!\n")
}

// buyBackArmour buys the party's armour back at half price.
func buyBackArmour() {
	printText("\nWhat will\n")
	answer := 0 // quirk: kept from the last deal when the party offers what it lacks
	for {
		textRow, textCol = promptRow, promptCol
		a := askLetter("You sell:"+answerRoom(2), 'B', 'H') - 'A'
		newLine()
		if a <= 0 {
			break
		}
		if party.armors[a] == 0 {
			printText("Come on, you don't own any.\n")
		} else {
			count := 1
			if party.armors[a] > 1 {
				printText("How many ")
				printText(armorNames[a])
				printText("s would you like to sell?" + answerRoom(3))
				count = askAmount(2)
				if count <= 0 {
					printText("\nToo bad.\n")
					break
				}
				if party.armors[a] < uint16(count) {
					printText("You don't have that many swine!\n")
					break
				}
				printText("I'll give ya\n")
				printNumber(halfPrice(count, armorPrices[a]), 1, '0')
				printText("gp for ")
				printText(themOrIt(count))
			} else {
				printText("I will give you\n")
				printNumber(halfPrice(1, armorPrices[a]), 1, '0')
				printText("gp for that ")
				printText(armorNames[a])
			}
			printText("\nDeal?" + answerRoom(2))
			answer = askYesNo()
			if answer < 'N' { // a blank or escape
				break
			}
			if answer == 'N' {
				printText("Harumph. What else would\n")
			} else {
				party.armors[a] -= uint16(count)
				takeGold(halfPrice(count, armorPrices[a]))
				showGold()
				showArmour()
				printText("\nFine! What else?\n")
			}
		}
		if answer != 'Y' && answer != 'N' {
			break
		}
	}
}

func visitArmourShop() {
	shop := armorShops[findShop(armorShops)]
	clearStatus()
	showArmour()
	printText("\n\n\n\nWelcome to\n")
	printText(shop.name)
	printText("\n\n")
	printText(shop.keeper)
	printText(" says: Welcome friend! Want to Buy or Sell? ")
	if choice := askBuyOrSell(); choice != 0 {
		printChar(choice)
		newLine()
		if choice == 'b' {
			buyArmour(shop.keeper, shop.stock)
		} else {
			buyBackArmour()
		}
	}
	clearStatusTitle()
	clearStatus()
	showStats()
	newLine()
	printText(shop.keeper)
	printText(" says: Good Bye.\n")
}

// ---- Horses ----

// horsePrice is the price of a horse for each member.
const horsePrice = 100

func visitHorseSeller() {
	printText("\nWelcome friend! Can I interest thee in horses?" + answerRoom(2))
	if answer := askYesNo(); answer != 'Y' {
		if answer == 'N' {
			printText("\nA shame, thou looks like thou could use a good horse!\n")
		}
		return
	}
	printText("\nFor only ")
	printNumber(int32(party.size), 1, '0')
	printText("00g.p. Thou can have the best!  Wilt thou buy?" + answerRoom(2))
	if answer := askYesNo(); answer != 'Y' {
		if answer == 'N' {
			printText("\nA shame, thou looks like thou could use a good horse!\n")
		}
		return
	}
	if horsePrice*party.size > party.gold {
		printText("\nIt seems thou hast not gold enough to pay!\n")
		return
	}
	party.gold -= horsePrice * party.size
	showGold()
	printText("\nHere, a better breed thou shalt not find ever!\n")
	party.tile = tileHorseWest
}

// ---- The guilds ----

var guilds = []shopFront{
	{locBuccaneersDen, "Pirate's Guild", "One Eyed Willey"},
	{locVesper, "The Guild Shop", "Long John Leary"},
}

// guildGoods are the guild's goods: torches, gems, keys and (not on the list) a sextant.
var guildGoods = [4]struct {
	price, count int
	pitch        string
}{
	{50, 5, "I can give ya 5 long lasting Torches for a mere 50gp.\n"},
	{60, 5, "I've got magical mapping Gems, 5 for only 60gp.\n"},
	{60, 6, "Magical Keys 1 use each, a fair price at 60gp for 6.\n"},
	{900, 1, "So...Ya want a Sextant...Well I gots one which I might part with fer 900 gold!\n"},
}

// guildStock is the party's stock of guild good g.
func guildStock(g int) *uint16 {
	return [...]*uint16{&party.torches, &party.gems, &party.keys, &party.sextants}[g]
}

func visitGuild() {
	guild := guilds[findShop(guilds)]
	clearStatus()
	showEquipment()
	printText("\n\n\n\nAvast ye mate! Shure ye wishes to buy from ol' ")
	printText(guild.keeper)
	printText("?\n\n\n")
	printText(guild.keeper)
	printText(" says: Welcome to ")
	printText(guild.name)
	printText("\nLike to see my goods?" + answerRoom(2))
	for askYesNo() == 'Y' {
		newLine()
		printText(guild.keeper)
		printText(" says: Good Mate! Ya see I gots:\nA-Torches\nB-Magic Gems\nC-Magic Keys\n")
		g := askLetter("Wat'l it be?"+answerRoom(2), 'A', 'D') - 'A'
		if g < 0 {
			break
		}
		goods := guildGoods[g]
		newLine()
		printText(goods.pitch)
		printText("\nWill ya buy?" + answerRoom(2))
		if answer := askYesNo(); answer != 'Y' {
			if answer == 'N' {
				printText("\nHmmm...Grmbl...\n")
			}
			break
		}
		if party.gold < uint16(goods.price) {
			clearStatus()
			clearStatusTitle()
			showStats()
			printText("\nWhat? Can't pay! Buzz off swine!\n")
			return
		}
		stock := guildStock(g)
		*stock = min(*stock+uint16(goods.count), maxStock)
		party.gold -= uint16(goods.price)
		printText("\nFine...fine...\n\n")
		showGold()
		showEquipment()
		newLine()
		printText(guild.keeper)
		printText(" says: See more?" + answerRoom(2))
	}
	clearStatus()
	clearStatusTitle()
	showStats()
	newLine()
	printText(guild.keeper)
	printText(" says:  See ya matie!\n")
}

// ---- The inns ----

// room is an inn's room: where its bed is, and its price.
type room struct {
	x, y  uint8
	price uint8
}

type inn struct {
	shopFront
	offer string
	room  room
}

var inns = []inn{
	{shopFront{locMoonglow, "The Honest Inn", "Scatu"}, "We have a room with 2 beds that rents for 20gp.", room{28, 6, 20}},
	{shopFront{locBritain, "Britannia Manor", "Jason"}, "We have a modest sized room with 1 bed for 15gp.", room{29, 6, 15}},
	{shopFront{locJhelom, "The Inn of Ends", "Smirk"}, "We have a very secure room of modest size and 1 bed for 10gp.", room{10, 26, 10}},
	{shopFront{locMinoc, "Wayfarer's Inn", "Estro"}, "We have three rooms available, a 1, 2 and 3 bed room for 30, 60 and 90gp each.", room{}},
	{shopFront{locTrinsic, "Honorable Inn", "Zajac"}, "We have a single bed room with a back door for 15g.", room{29, 2, 15}},
	{shopFront{locSkaraBrae, "The Inn of the Spirits", "Tyrone"}, "Unfortunately, I have but only a very small room with 1 bed: worse yet, it is haunted!  If you do wish to stay it costs 5gp.", room{28, 11, 5}},
	{shopFront{locVesper, "The Sleep Shop", "Tymus"}, "All we have is that cot over there. But it is comfortable, and only 1gp.", room{25, 23, 1}},
}

// minocRooms are the three rooms of Minoc's inn, of one, two and three beds.
var minocRooms = [3]room{{2, 6, 30}, {2, 2, 60}, {8, 2, 90}}

// The night at an inn: the healing, and the chance of rogues (one in innAmbushOdds), of
// Isaac the ghost at Skara Brae (one in ghostOdds) and of the talk of rats.
const (
	innHealBase   = 100
	innHealSteps  = 50 // and up to 49 steps of two more
	innAmbushOdds = 8
	ghostOdds     = 4
	ratsOdds      = 4
	isaacTalk     = 16 // Isaac's talk record
	isaacSlot     = 0
)

// innAmbush: rogues fall on the party in the night.
func innAmbush() {
	printText("\nIn the middle of the night, while out for a stroll...\n")
	mode := gameMode
	gameMode = modeCombatCamp
	foeTile = tileRogue
	encounterTile = tileRogue
	startInnFight()
	gameMode = mode
}

// sleepAtInn puts the party to bed in the room and heals it overnight.
func sleepAtInn(r room) {
	waitSeconds(1, false)
	party.x = r.x
	party.y = r.y
	waitSeconds(1, false)
	for m := int(party.size) - 1; m >= 0; m-- {
		if isAlive(m) {
			party.members[m].status = statusAsleep
		}
	}
	showStats()
	party.tile = tileAsleep
	waitSeconds(5, false)
	party.tile = tileOnFoot
	for m := int(party.size) - 1; m >= 0; m-- {
		if party.members[m].status == statusAsleep {
			party.members[m].status = statusGood
			healMember(m, randomBelow(innHealSteps)*2+innHealBase)
		}
	}
	regainMP()
	showStats()
	if isAlive(0) && randomBits(innAmbushOdds-1) == 0 {
		innAmbush()
	} else if party.loc == locSkaraBrae && randomBits(ghostOdds-1) == 0 {
		people := &area.people
		people.tile[isaacSlot] = tileGhost
		people.baseTile[isaacSlot] = tileGhost
		people.x[isaacSlot] = party.x - 1
		people.y[isaacSlot] = party.y
		people.behaviour[isaacSlot] = 1
		people.talkIndex[isaacSlot] = isaacTalk
	}
	printText("\nMorning!\n")
}

func visitInn() {
	if party.tile != tileOnFoot {
		printText("The Innkeeper says: Get that horse out of here!!!\n")
		return
	}
	printText("The Innkeeper says: Welcome to ")
	inn := inns[findShop(inns)]
	printText(inn.name)
	printText("\n\nI am ")
	printText(inn.keeper)
	printText(".\n\nAre you in need of lodging?" + answerRoom(2))
	if askYesNo() != 'Y' {
		newLine()
		printText(inn.keeper)
		printText(" says: Then you have come to the wrong place! Good Day.\n")
		return
	}
	newLine()
	printText(inn.offer)
	newLine()
	r := inn.room
	if inn.loc == locMinoc {
		beds := askLetter("1, 2 or 3 beds?"+answerRoom(2), '1', '3') - '1'
		if beds < 0 {
			return
		}
		r = minocRooms[beds]
	} else {
		printText("\nTake it?" + answerRoom(2))
		if answer := askYesNo(); answer != 'Y' {
			if answer == 'N' {
				printText("\nYou won't find a better deal in this towne!\n")
			}
			return
		}
	}
	if uint16(r.price) > party.gold {
		printText("\nIf you can't pay, you can't stay! Good Bye.\n")
		return
	}
	party.gold -= uint16(r.price)
	showGold()
	printText("\n\nVery good.  Have a pleasant night.\n")
	if randomBits(ratsOdds-1) == 0 {
		printText("\nOh, and don't mind the strange noises, it's only rats!\n")
	}
	sleepAtInn(r)
}

// ---- The healers ----

var healers = []shopFront{
	{locCastleBritannia, "The Royal Healer", "Pendragon"},
	{locLycaeum, "The Truth Healer", "Starfire"},
	{locEmpathAbbey, "The Love Healer", "Salle'"},
	{locSerpentsHold, "Courage Healer", "Windwalker"},
	{locMoonglow, "The Healer", "Harmony"},
	{locBritain, "Wound Healing", "Celest"},
	{locJhelom, "Heal and Health", "Triplet"},
	{locYew, "Just Healing", "Justin"},
	{locSkaraBrae, "The Mystic Heal", "Spiran"},
	{locCove, "The Healer Shop", "Quat"},
}

// The healers' prices, and the blood they ask of a leader of at least donorHP.
const (
	curePrice        = 100
	healPrice        = 200
	resurrectPrice   = 300
	donorHP          = 400
	bloodGiven       = 100
	bloodKarma       = 5
	healerSoundPitch = 192
)

// askPatient asks who needs help; -1 for no one.
func askPatient(healer shopFront) int {
	newLine()
	printText(healer.keeper)
	printText(" asks:\n")
	m := askMember("Who is in need?" + answerRoom(2))
	if m < 0 {
		printText("No one?\n")
	}
	return m
}

// payHealer asks the party to pay the price; true if it did.
func payHealer(price int) bool {
	printText("\nWilt thou pay?" + answerRoom(2))
	switch askYesNo() {
	case 'Y':
		party.gold -= uint16(price)
		showGold()
		return true
	case 'N':
		printText("\nThen I cannot aid thee.\n")
	}
	return false
}

// healerGlow flashes the map and the patient's status line.
func healerGlow(m int) {
	invertMap()
	invertStatusLine(m)
	sound(soundSweep, healerSoundPitch)
	invertStatusLine(m)
	invertMap()
}

func sayPurseTooLight() {
	printText("\nI see by thy purse that thou hast not enough gold. I cannot aid thee.\n")
}

func cure(healer shopFront) {
	m := askPatient(healer)
	if m == -1 {
		return
	}
	if party.members[m].status != statusPoisoned {
		printText("Thou suffers not from Poison!\n")
		return
	}
	printText("A curing will cost thee 100gp.\n")
	if party.gold < curePrice {
		printText("\nI see by thy purse that thou hast not enough gold.")
		flushKeys()
		waitSeconds(5, true)
		if keyWaiting() {
			readKey()
		}
		printText("\n\nI will cure thee for free, but give unto others whenever thou may!\n")
	} else if !payHealer(curePrice) {
		return
	}
	party.members[m].status = statusGood
	healerGlow(m)
	showStats()
}

func heal(healer shopFront) {
	m := askPatient(healer)
	if m == -1 {
		return
	}
	if party.members[m].hp == party.members[m].maxHP {
		printText("Thou art already quite healthy!\n")
		return
	}
	printText("\nA healing will cost thee 200gp.\n")
	if party.gold < healPrice {
		sayPurseTooLight()
		return
	}
	if payHealer(healPrice) {
		party.members[m].hp = party.members[m].maxHP
		healerGlow(m)
		showStats()
	}
}

func resurrect(healer shopFront) {
	m := askPatient(healer)
	if m == -1 {
		return
	}
	if party.members[m].status != statusDead {
		printText("Thou art not dead fool!\n")
		return
	}
	printText("Resurrection will cost thee 300gp.\n")
	if party.gold < resurrectPrice {
		sayPurseTooLight()
		return
	}
	if payHealer(resurrectPrice) {
		party.members[m].status = statusGood
		healerGlow(m)
		showStats()
	}
}

// askForBlood: the healer asks a strong leader for blood, for sacrifice.
func askForBlood() {
	printText("Art thou willing to give 100pts of thy blood to aid others?" + answerRoom(2))
	switch askYesNo() {
	case 'Y':
		raiseKarma(&party.karmaSacrifice, bloodKarma)
		printText("Thou art a great help.  We are in dire need!\n")
		party.members[0].hp -= bloodGiven
		showStats()
		waitSeconds(1, false)
	case 'N':
		lowerKarma(&party.karmaSacrifice, bloodKarma)
		showStats()
	}
}

func visitHealer() {
	healer := healers[findShop(healers)]
	printText("Welcome unto\n")
	printText(healer.name)
	newLine()
	printText(healer.keeper)
	printText(" says: Peace and Joy be with you friend. Are you in need of help?" + answerRoom(2))
	for askYesNo() == 'Y' {
		newLine()
		printText(healer.keeper)
		printText(" says: We can perform:\nA-Curing\nB-Healing\nC-Resurrection\n")
		need := askLetter("Your need:"+answerRoom(2), 'A', 'C')
		if need < 'A' {
			break
		}
		switch need {
		case 'A':
			cure(healer)
		case 'B':
			heal(healer)
		case 'C':
			resurrect(healer)
		}
		newLine()
		printText(healer.keeper)
		printText(" asks: Do you need more help?" + answerRoom(2))
	}
	if party.members[0].hp >= donorHP {
		askForBlood()
	}
	showStats()
	newLine()
	printText(healer.keeper)
	printText(" says: May thy life be guarded by the powers of good.\n")
}

// ---- The pubs ----

type pub struct {
	shopFront
	dish      string
	dishPrice int
}

var pubs = []pub{
	{shopFront{locBritain, "Jolly Spirits", "Sam"}, "Lamb Chops", 4},
	{shopFront{locJhelom, "The Bloody Pub", "Celestial"}, "Dragon Tartar", 2},
	{shopFront{locTrinsic, "The Keg Tap", "Terran"}, "Brown Beans", 3},
	{shopFront{locPaws, "Folley Tavern", "Greg 'n Rob"}, "Folley Filet", 2},
	{shopFront{locBuccaneersDen, "Captain Black Tavern", "The Cap'n"}, "Dog Meat Pie", 4},
	{shopFront{locVesper, "Axe 'n Ale", "Arron"}, "Green Granukit", 2},
}

// pubTopics are what a tipped barkeep knows of, with the tip each wants. A pub knows only
// the topics from its own number in the list of pubs on.
var pubTopics = []struct {
	word string
	tip  int
	tale string
}{
	{"black stone", 20, "Ah, the Black Stone.  Yes I've heard of it. But, the only one who knows where it lies is the wizard Merlin.\n"},
	{"sextant", 30, "For navigation a Sextant is vital...  Ask for item \"D\" in the Guild shops!\n"},
	{"white stone", 10, "Hmmm, Now let me see... Yes, it was the old Hermit... Sloven!  He is tough to find, lives near Lock Lake I hear.\n"},
	{"mandrake", 40, "The last person I knew that had any Mandrake was an old alchemist named Calumny.\n"},
	{"skull", 99, "If thou must know of that evilest of all things... find the beggar Jude.  He is very very poor!\n"},
	{"nightshade", 25, "Of Nightshade I know but this... Seek out Virgil or thou shalt miss! Try in Trinsic!\n"},
	{"mandrake root", 0, ""}, // asked as "mandrake"
}

// mandrakeRoot is asked as the last word, and answered as "mandrake".
const (
	topicMandrake     = 3
	topicMandrakeRoot = 6
)

// The ale: a mug costs alePrice; three mugs are the limit, and a tip above the price
// buys a question.
const (
	alePrice  = 2
	maxDrinks = 3
)

var drinksServed int

// orderFood sells plates of the pub's dish, a ration each; true to leave.
func orderFood(p pub) bool {
	printText("Our specialty is ")
	printText(p.dish)
	printText(", which costs ")
	printNumber(int32(p.dishPrice), 1, '0')
	printText("gp.  How many plates would you like?" + answerRoom(3))
	plates := askAmount(2)
	if plates <= 0 {
		if plates < 0 {
			sound(soundBeep)
		}
		return true
	}
	for served := 0; plates != 0; served++ {
		if party.gold < uint16(p.dishPrice) {
			if served == 0 {
				printText("\nYa cannot afford any!\n")
				return true
			}
			printText("\nYa can only afford ")
			printNumber(int32(served), 1, '0')
			printText(" plate")
			if served != 1 {
				printChar('s')
			}
			printChar('.')
			newLine()
			break
		}
		party.gold -= uint16(p.dishPrice)
		party.food = min(party.food+ration, maxFood)
		showStats()
		plates--
	}
	return false
}

// orderAle sells a mug of ale, and with a tip a tale; true to leave.
func orderAle(pubNumber int, p pub) bool {
	if drinksServed == maxDrinks {
		newLine()
		printText(p.keeper)
		printText(" says: Sorry, you seem to have too many.  Bye!\n")
		return true
	}
	drinksServed++
	printText("\nHere's a mug of our best. That'll be 2gp. You pay?" + answerRoom(3))
	paid := askAmount(2)
	if paid <= 0 {
		if paid < 0 {
			sound(soundBeep)
		}
		return true
	}
	if paid < alePrice {
		printText("\nWon't pay, eh. Ya scum, be gone fore ey call the guards!\n")
		return true
	}
	if party.gold < uint16(paid) {
		printText("\nIt seems that you have not the gold. Good Day!\n")
		return true
	}
	party.gold -= uint16(paid)
	showGold()
	if paid > alePrice && askBarkeep(pubNumber, p, paid) {
		return true
	}
	printText("\nAnythin' else?" + answerRoom(2))
	return askYesNo() != 'Y'
}

// askBarkeep: a tipped barkeep tells what they know; true to leave.
func askBarkeep(pubNumber int, p pub, paid int) bool {
	printText("\nWhat'd ya like to know friend?\n")
	asked := readWord(14)
	topic := len(pubTopics) - 1
	for topic >= 0 && !strings.EqualFold(asked, pubTopics[topic].word) {
		topic--
	}
	if topic == topicMandrakeRoot {
		topic = topicMandrake
	}
	if topic < pubNumber {
		printText("\n'fraid I can't help ya there friend!\n")
		return false
	}
	for pubTopics[topic].tip > paid {
		printText("\nThat subject is a bit foggy, perhaps more gold will refresh my memory. You give:" + answerRoom(3))
		more := askAmount(2)
		if more <= 0 {
			// Quirk: the original tests the first payment (never negative here), so no bell.
			printText("\nSorry, I could not hep ya mate!\n")
			return true
		}
		if party.gold < uint16(more) {
			printText("\nYer don't have that mate!\n")
			printText("\nSorry, I could not hep ya mate!\n")
			return true
		}
		party.gold -= uint16(more)
		showGold()
		paid += more
	}
	newLine()
	printText(p.keeper)
	printText(" says: ")
	printText(pubTopics[topic].tale)
	newLine()
	return false
}

func visitPub() {
	drinksServed = 0
	number := findShop(pubs)
	p := pubs[number]
	newLine()
	printText(p.keeper)
	printText(" says:\nWelcome to ")
	printText(p.name)
	newLine()
	for {
		printText(p.keeper)
		printText(" says: What'll it be, Food er Ale?" + answerRoom(2))
		var order int
		for {
			askKeys("", item("Food", key('f')), item("Ale", key('a')), item("Leave", keyEsc))
			order = int(int8(readKey()))
			if order >= 'a' && order <= 'z' {
				order -= 'a' - 'A'
			}
			if order == ' ' || order == 27 || order == '\r' {
				newLine()
				return
			}
			if order == 'F' || order == 'A' {
				break
			}
			sound(soundBeep)
		}
		printChar(order)
		newLine()
		leave := false
		if order == 'F' {
			leave = orderFood(p)
		} else {
			leave = orderAle(number, p)
		}
		if leave {
			break
		}
		showStats()
		printText("Here ye arr. Somethin' else?" + answerRoom(2))
		if askYesNo() != 'Y' {
			break
		}
	}
	printText("See ya mate!\n")
}

// ---- Food ----

type foodShop struct {
	shopFront
	packPrice int
}

var foodShops = []foodShop{
	{shopFront{locMoonglow, "The Sage Deli", "Shaman"}, 25},
	{shopFront{locBritain, "Adventure Food", "Windrick"}, 40},
	{shopFront{locYew, "The Dry Goods", "Donnar"}, 35},
	{shopFront{locSkaraBrae, "Food for Thought", "Mintol"}, 20},
	{shopFront{locPaws, "The Market", "Max"}, 30},
}

// pack is a pack of 25 rations.
const pack = 25 * ration

func visitFoodShop() {
	shop := foodShops[findShop(foodShops)]
	printText("\n\n\n\nWelcome to ")
	printText(shop.name)
	printText("\n\n")
	printText(shop.keeper)
	printText(" says: Good day, and Welcome friend.\n\nMay I interest you in some rations?" + answerRoom(2))
	if answer := askYesNo(); answer != 'Y' {
		if answer == 'N' {
			printText("Goodbye, Come again!\n")
		}
		return
	}
	printText("We have the best adventure rations, 25 for only ")
	printNumber(int32(shop.packPrice), 1, '0')
	printText("gp.")
	for {
		printText("\nHow many packs of 25 would you like?" + answerRoom(4))
		packs := int(int8(askAmount(3))) // quirk: kept in a byte, so 128 packs and more wrap
		answer := 'N'
		switch {
		case packs <= 0:
			printText("\nToo bad. Maybe next time.\n")
		case uint16(shop.packPrice*packs) > party.gold:
			if uint16(shop.packPrice) > party.gold {
				printText("You cannot afford any!\n")
				printText("\nGoodbye. Come again!\n")
				return
			}
			printText("You can only afford ")
			printNumber(int32(party.gold)/int32(shop.packPrice), 1, '0')
			printText(" packs.\n")
			answer = 'Y'
		default:
			party.food = min(party.food+pack*int32(packs), maxFood)
			party.gold -= uint16(shop.packPrice * packs)
			showStats()
			printText("\nThank you. Anything else?" + answerRoom(2))
			answer = rune(askYesNo())
		}
		if answer != 'Y' {
			break
		}
	}
	printText("\nGoodbye. Come again!\n")
}
