// Command localrename renames the parameters and local variables of functions, one
// "Func old new" line per rename in a map file (# starts a comment; Func is a function name,
// or Type.Method). Every object named old declared inside Func is renamed, closures
// included. It refuses, changing nothing, when new is already visible at a use or
// declaration of old, or when a name is not found.
//
//	go run ./tools/localrename MAP DIR
package main

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type rename struct{ fn, old, new string }

type edit struct {
	off  int
	old  string
	new  string
	file string
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: localrename MAP DIR")
		os.Exit(2)
	}
	renames, err := readMap(os.Args[1])
	check(err)
	dir := os.Args[2]
	fset := token.NewFileSet()
	paths, _ := filepath.Glob(filepath.Join(dir, "*.go"))
	var files []*ast.File
	for _, p := range paths {
		if strings.HasSuffix(p, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, p, nil, parser.ParseComments)
		check(err)
		files = append(files, f)
	}
	info := &types.Info{Defs: map[*ast.Ident]types.Object{}, Uses: map[*ast.Ident]types.Object{}, Scopes: map[ast.Node]*types.Scope{}}
	conf := types.Config{Importer: importer.ForCompiler(fset, "source", nil)}
	_, err = conf.Check(files[0].Name.Name, fset, files, info)
	check(err)

	funcs := map[string]*ast.FuncDecl{}
	for _, f := range files {
		for _, d := range f.Decls {
			if fd, ok := d.(*ast.FuncDecl); ok {
				name := fd.Name.Name
				if fd.Recv != nil && len(fd.Recv.List) == 1 {
					t := fd.Recv.List[0].Type
					if s, ok := t.(*ast.StarExpr); ok {
						t = s.X
					}
					if id, ok := t.(*ast.Ident); ok {
						name = id.Name + "." + name
					}
				}
				funcs[name] = fd
			}
		}
	}
	var edits []edit
	bad := false
	for _, r := range renames {
		fd := funcs[r.fn]
		if fd == nil {
			fmt.Fprintf(os.Stderr, "no function %s\n", r.fn)
			bad = true
			continue
		}
		objs := map[types.Object]bool{}
		ast.Inspect(fd, func(n ast.Node) bool {
			if id, ok := n.(*ast.Ident); ok && id.Name == r.old {
				if o := info.Defs[id]; o != nil {
					if _, isVar := o.(*types.Var); isVar && !o.(*types.Var).IsField() {
						objs[o] = true
					}
				}
			}
			return true
		})
		if len(objs) == 0 {
			fmt.Fprintf(os.Stderr, "%s: no local %s\n", r.fn, r.old)
			bad = true
			continue
		}
		ast.Inspect(fd, func(n ast.Node) bool {
			id, ok := n.(*ast.Ident)
			if !ok || id.Name != r.old {
				return true
			}
			o := info.Defs[id]
			if o == nil {
				o = info.Uses[id]
			}
			if !objs[o] {
				return true
			}
			// new must not resolve to anything at this point
			if s := innermost(info, fd, id.Pos()); s != nil {
				if _, hit := s.LookupParent(r.new, id.Pos()); hit != nil {
					fmt.Fprintf(os.Stderr, "%s: %s -> %s: %s already visible at %s\n", r.fn, r.old, r.new, r.new, fset.Position(id.Pos()))
					bad = true
					return false
				}
			}
			p := fset.Position(id.Pos())
			edits = append(edits, edit{p.Offset, r.old, r.new, p.Filename})
			return true
		})
	}
	// two renames in one function to the same new name
	seen := map[string]string{}
	for _, r := range renames {
		k := r.fn + " " + r.new
		if o, dup := seen[k]; dup && o != r.old {
			fmt.Fprintf(os.Stderr, "%s: both %s and %s renamed to %s\n", r.fn, o, r.old, r.new)
			bad = true
		}
		seen[k] = r.old
	}
	if bad {
		os.Exit(1)
	}
	byFile := map[string][]edit{}
	for _, e := range edits {
		byFile[e.file] = append(byFile[e.file], e)
	}
	for file, es := range byFile {
		sort.Slice(es, func(i, j int) bool { return es[i].off > es[j].off })
		b, err := os.ReadFile(file)
		check(err)
		last := -1
		for _, e := range es {
			if e.off == last {
				continue
			}
			last = e.off
			if string(b[e.off:e.off+len(e.old)]) != e.old {
				check(fmt.Errorf("%s:%d: expected %s", file, e.off, e.old))
			}
			b = append(b[:e.off], append([]byte(e.new), b[e.off+len(e.old):]...)...)
		}
		check(os.WriteFile(file, b, 0o644))
	}
}

// innermost is the innermost scope of fd containing pos.
func innermost(info *types.Info, fd *ast.FuncDecl, pos token.Pos) *types.Scope {
	s := info.Scopes[fd.Type]
	if s == nil {
		return nil
	}
	return s.Innermost(pos)
}

func readMap(path string) ([]rename, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var rs []rename
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line, _, _ := strings.Cut(sc.Text(), "#")
		fs := strings.Fields(line)
		if len(fs) == 0 {
			continue
		}
		if len(fs) != 3 {
			return nil, fmt.Errorf("bad line: %q", sc.Text())
		}
		rs = append(rs, rename{fs[0], fs[1], fs[2]})
	}
	return rs, sc.Err()
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
