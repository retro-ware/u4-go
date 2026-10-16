# u4-go

A Go port of *Ultima IV: Quest of the Avatar* (1985, DOS), built on Ebiten. It is a
function-by-function translation of the decompiled C source of AVATAR.EXE and TITLE.EXE
([u4-decompiled](https://github.com/memmaker/u4-decompiled)), so the game rules, quirks and bugs
of the original are kept. See `PORTING.md` for the porting rules.

No original game data is included. Bring your own copy of the game (for example the GOG
release); the port reads its files at runtime. Nothing is ever written to the game directory.

## Build and run

Needs Go 1.26 or newer.

    go build -o u4 ./cmd/u4
    ./u4 -import /path/to/your/game     # once: copies the game files into ./game
    ./u4                                # play

`-import SRC` takes a plain folder, a zip, a CD image (ISO/BIN/CUE), a GOG installer
(`setup_*.exe`/`.bin`) or loose files, finds the game files in it and exits.

Flags:

| Flag | Meaning |
|------|---------|
| `-data DIR` | game files directory (default `$GOPLAY_DATA`, else `game`) |
| `-save DIR` | saved games directory (default: the game directory) |
| `-import SRC` | import the game from SRC into `-data`, then exit |
| `-scale N` | window scale (default 3) |
| `-ega` | start with the original EGA graphics instead of the VGA upgrade |
| `-nomusic` | no MIDI music |

Music is synthesised from an embedded SoundFont, which is not in git. Run
`tools/fetch-soundfont.sh` before building to include it.

## Controls

The original keys work as in the DOS game. In addition:

| Key | Action |
|-----|--------|
| Enter / right click | command menu (pauses the game) |
| Click on the map | walk there |
| F1 | help |
| F2 | toggle EGA / VGA graphics |
| F3, F4 | overlays |
| F5 | message history |
| F6 | original screen layout |
| Alt-V, Alt-+/Alt-- | music on/off, volume |

The modern view is wider than the original 11x11 map; game rules are unchanged, and menu
entries only queue the original keys.

## Layout

- `cmd/u4`: entry point
- `internal/u4`: AVATAR.EXE port, one Go file per C module
- `internal/title`: TITLE.EXE port
- `internal/host`, `internal/ega`, `internal/speaker`, `internal/music`: display, input and sound
- `internal/gamefs`, `internal/gameimport`: game file lookup and `-import`

## Status

| Part | State |
|------|-------|
| Title / intro | ported |
| Avatar (world, towns, dungeons, combat) | ported |
| Sound and MIDI music | working |
| Wide view, menus, VGA upgrade | working |

Run the tests with `go test ./...`.

## Legal

Ultima is a trademark of Electronic Arts. This project is unaffiliated and ships no game assets.
