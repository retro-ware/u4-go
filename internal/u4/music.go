package u4

import "github.com/memmaker/u4-go/internal/music"

// musicPrompt picks the song for where the party is. It runs at each command prompt, where
// the upgrade patch's driver (seeing the prompt drawn) also ended the songs of conversations,
// shops and camping.
func musicPrompt() {
	switch CurMode {
	case MOD_OUTDOORS:
		music.Play(music.Overworld)
	case MOD_BUILDING:
		if Party._loc <= 4 { // the castles
			music.Play(music.Castle)
		} else {
			music.Play(music.Towns)
		}
	case MOD_DUNGEON:
		music.Play(music.Dungeon)
	case MOD_COMBAT, MOD_COM_CAMP, MOD_COM_ROOM:
		music.Play(music.Combat)
	}
}
