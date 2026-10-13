// Command fieldpos prints the position (file:line:col) of a declaration in a package
// directory, for gopls rename: a struct field as Type.field, or a top-level name.
//
//	go run ./tools/fieldpos DIR Type.field|name
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

func main() {
	dir, want := os.Args[1], os.Args[2]
	typ, field, isField := strings.Cut(want, ".")
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, nil, 0)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, p := range pkgs {
		for _, f := range p.Files {
			for _, d := range f.Decls {
				switch d := d.(type) {
				case *ast.GenDecl:
					for _, s := range d.Specs {
						switch s := s.(type) {
						case *ast.TypeSpec:
							if !isField && s.Name.Name == want {
								out(fset, s.Name.Pos())
							}
							st, ok := s.Type.(*ast.StructType)
							if !isField || s.Name.Name != typ || !ok {
								continue
							}
							for _, fl := range st.Fields.List {
								for _, n := range fl.Names {
									if n.Name == field {
										out(fset, n.Pos())
									}
								}
							}
						case *ast.ValueSpec:
							for _, n := range s.Names {
								if !isField && n.Name == want {
									out(fset, n.Pos())
								}
							}
						}
					}
				case *ast.FuncDecl:
					if !isField && d.Recv == nil && d.Name.Name == want {
						out(fset, d.Name.Pos())
					}
				}
			}
		}
	}
	fmt.Fprintln(os.Stderr, "not found:", want)
	os.Exit(1)
}

func out(fset *token.FileSet, pos token.Pos) {
	p := fset.Position(pos)
	fmt.Printf("%s:%d:%d\n", p.Filename, p.Line, p.Column)
	os.Exit(0)
}
