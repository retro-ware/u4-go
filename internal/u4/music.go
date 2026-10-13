package u4

import "github.com/memmaker/u4-go/internal/music"

// musicPrompt picks the song for where the party is. It runs at each command prompt, where
// the upgrade patch's driver (seeing the prompt drawn) also ended the songs of conversations,
// shops and camping.
func musicPrompt() {
	switch gameMode {
	case modeOutdoors:
		music.Play(music.Overworld)
	case modeTown:
		if party.loc <= 4 { // the castles
			music.Play(music.Castle)
		} else {
			music.Play(music.Towns)
		}
	case modeDungeon:
		music.Play(music.Dungeon)
	case modeCombat, modeCombatCamp, modeCombatRoom:
		music.Play(music.Combat)
	}
}
