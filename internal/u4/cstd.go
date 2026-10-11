package u4

// The C library and runtime pieces the game uses: setjmp/longjmp on D_9458, string helpers
// on zero-terminated byte buffers.

import "strconv"

// jmpMain is longjmp(D_9458, -1): back to the setjmp in main's loop (or around DNG_main).
type jmpMain struct{}

func longjmp_D_9458() { panic(jmpMain{}) }

// strlen of a C string in a byte buffer.
func strlen(b []byte) int {
	for i, c := range b {
		if c == 0 {
			return i
		}
	}
	return len(b)
}

// cstr is the C string held in b.
func cstr(b []byte) string { return string(b[:strlen(b)]) }

// atol parses the leading integer of a C string, as the C library's atol.
func atol(b []byte) int32 {
	s := cstr(b)
	i := 0
	for i < len(s) && (s[i] == ' ' || s[i] == '\t') {
		i++
	}
	j := i
	if j < len(s) && (s[j] == '-' || s[j] == '+') {
		j++
	}
	for j < len(s) && s[j] >= '0' && s[j] <= '9' {
		j++
	}
	v, _ := strconv.ParseInt(s[i:j], 10, 64)
	return int32(v)
}

// strcpy into a byte buffer.
func strcpy(dst []byte, src string) {
	n := copy(dst, src)
	if n < len(dst) {
		dst[n] = 0
	}
}

// strcat onto a C string in a byte buffer.
func strcat(dst []byte, src string) { strcpy(dst[strlen(dst):], src) }

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}
