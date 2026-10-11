GeneralUser GS v2.0.3 by S. Christian Collins (https://www.schristiancollins.com,
https://github.com/mrbumpy409/GeneralUser-GS), the SoundFont the music is synthesized with.
Free to use and redistribute: see LICENSE-GeneralUser-GS.txt.

GeneralUser-GS.sf2 (~32 MB) is not in git: run tools/fetch-soundfont.sh (or go generate
./internal/music) to fetch it here. Desktop builds embed whatever .sf2 is here; js/wasm builds
never embed it and read the file $GOPLAY_SF2 names instead.
