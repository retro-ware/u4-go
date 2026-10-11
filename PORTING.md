# Porting rules (read first)

Source: ~/Projects/u4-decompiled (SRC = AVATAR.EXE, SRC-TITLE = TITLE.EXE, MSC 4.0 small model).
Data files: ~/Projects/u4-decompiled/gog/inst (original game dir; read-only, never copy into repo).

- One Go file per C module in internal/u4 (lowercase module name, e.g. U4_SPELL.C -> spell.go).
  Header comment: "Port of U4_X.C (signed|unsigned char): what it does." (MAKEFILE says which modules use /J = unsigned char.)
- Functions keep their C names exactly (C_7D50, CMD_Cast, w_NotHere); keep the C's /*C_xxxx*/ address comments.
- Globals are typed Go vars in types.go (U4.H / U4_VAR.C). Initialised data (D_xxxx tables in a module) lives in
  the module's Go file. C int/unsigned = int16/uint16 where wrap/sign matters; function params/locals may be Go int
  when values are small, but reproduce 16-bit wrap and signed/unsigned compares where they can occur.
- char: sc(b) when the module is signed char.
- setjmp/longjmp: see cstd.go (jmpMain, catchJmp). exit(n): exit(n).
- Waiting/drawing: never busy-loop without u_delay / t_callback / host.Yield; existing helpers already yield.
- Before writing a helper, grep internal/u4 for it — it may exist. File-local helpers get a module prefix to avoid
  collisions with parallel porters.
- Check only your files: `go build -gcflags=-e ./internal/u4 2>&1 | grep -E 'yourfile.go'`. Missing functions owned
  by another module are fine while others work; don't stub them in your files.
- Keep quirks/bugs of the original; comment them.
