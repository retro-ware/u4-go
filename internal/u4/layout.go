package u4

// The screen's layout in character cells (8x8 pixels, 40 columns by 25 rows): the map
// window on the left, the party's status at the top right, the text window below it.
const (
	textLeft       = 24 // the first column of the text window and of the status lines
	textLastColumn = 39 // the last column of the screen
	textWidth      = 40 // a word that would end past this column wraps
	textBottom     = 23 // the text window's bottom row, where new lines appear

	statsRowFood     = 10 // food, gold (or the ship's hull) and the active spell
	statsHPColumn    = 35 // a party member's hit points, then their status letter
	statsGoldColumn  = 35
	statsSpellColumn = 31 // the active spell's letter, or the virtue just lost

	textPageLines = 12 // lines shown before a long text waits for a key
)

// promptChar is the character printed before a prompt for a command.
const promptChar = 16
