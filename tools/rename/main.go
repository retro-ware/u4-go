// Command rename renames identifiers across Go packages, one "old new" pair per line of a map
// file (# starts a comment), whole words only, comments included but string literals left alone. It refuses,
// changing nothing, when a new name is already an identifier in the packages.
//
//	go run ./tools/rename MAP DIR...
package main

import (
	"bufio"
	"fmt"
	"go/scanner"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: rename MAP DIR...")
		os.Exit(2)
	}
	names, err := readMap(os.Args[1])
	if err != nil {
		fail(err)
	}
	src := map[string]string{}
	for _, dir := range os.Args[2:] {
		files, _ := filepath.Glob(filepath.Join(dir, "*.go"))
		for _, f := range files {
			b, err := os.ReadFile(f)
			if err != nil {
				fail(err)
			}
			src[f] = string(b)
		}
	}
	idents := map[string]bool{}
	for _, s := range src {
		fset := token.NewFileSet()
		var sc scanner.Scanner
		sc.Init(fset.AddFile("", fset.Base(), len(s)), []byte(s), nil, 0)
		for {
			_, tok, lit := sc.Scan()
			if tok == token.EOF {
				break
			}
			if tok == token.IDENT {
				idents[lit] = true
			}
		}
	}
	bad := false
	for old, n := range names {
		if idents[n] {
			fmt.Fprintf(os.Stderr, "name in use: %s (for %s)\n", n, old)
			bad = true
		}
	}
	if bad {
		os.Exit(1)
	}
	olds := make([]string, 0, len(names))
	for o := range names {
		olds = append(olds, regexp.QuoteMeta(o))
	}
	sort.Slice(olds, func(i, j int) bool { return len(olds[i]) > len(olds[j]) })
	re := regexp.MustCompile(`\b(` + strings.Join(olds, "|") + `)\b`)
	for f, s := range src {
		// Rewrite identifiers and comments only; string and rune literals are game text.
		var out strings.Builder
		fset := token.NewFileSet()
		file := fset.AddFile("", fset.Base(), len(s))
		var sc scanner.Scanner
		sc.Init(file, []byte(s), nil, scanner.ScanComments)
		last := 0
		for {
			pos, tok, lit := sc.Scan()
			if tok == token.EOF {
				break
			}
			if tok != token.IDENT && tok != token.COMMENT {
				continue
			}
			off := file.Offset(pos)
			text := s[off : off+len(lit)]
			out.WriteString(s[last:off])
			if tok == token.IDENT {
				if n, ok := names[text]; ok {
					text = n
				}
			} else {
				text = re.ReplaceAllStringFunc(text, func(m string) string { return names[m] })
			}
			out.WriteString(text)
			last = off + len(lit)
		}
		out.WriteString(s[last:])
		r := out.String()
		if r != s {
			if err := os.WriteFile(f, []byte(r), 0o644); err != nil {
				fail(err)
			}
		}
	}
}

func readMap(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	m := map[string]string{}
	seen := map[string]string{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line, _, _ := strings.Cut(sc.Text(), "#")
		fs := strings.Fields(line)
		if len(fs) == 0 {
			continue
		}
		if len(fs) != 2 {
			return nil, fmt.Errorf("bad line: %q", sc.Text())
		}
		if o, dup := seen[fs[1]]; dup {
			return nil, fmt.Errorf("%s is the new name of both %s and %s", fs[1], o, fs[0])
		}
		seen[fs[1]] = fs[0]
		m[fs[0]] = fs[1]
	}
	return m, sc.Err()
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
