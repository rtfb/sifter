package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/docopt/docopt.go"
)

type visitor struct {
	allStrings []string
	tFunc      string
}

func getAllFiles(siftParam string) []string {
	var files []string
	if dir, err := isDir(siftParam); err == nil && dir {
		filepath.Walk(siftParam, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.Mode().IsRegular() && strings.HasSuffix(path, ".go") {
				files = append(files, path)
			}
			return nil
		})
		return files
	}
	// TODO: add code to treat siftParam as a wildcard
	//if isGlob() {
		//return expandGlob()
	//}
	files = append(files, siftParam)
	return files
}

func isDir(file string) (bool, error) {
	f, err := os.Open(file)
	if err != nil {
		return false, err
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return false, err
	}
	switch mode := fi.Mode(); {
	case mode.IsDir():
		return true, nil
	case mode.IsRegular():
		return false, nil
	}
	return false, nil
}

func (v *visitor) parseAllFiles(files []string) {
	for _, fileName := range files {
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, fileName, nil, 0)
		if err != nil {
			panic(err) // XXX: better error handling
		}
		ast.Walk(v, f)
	}
}

func getTFuncName(stmt *ast.AssignStmt) (string, bool) {
	name := ""
	if len(stmt.Lhs) > 0 {
		if id, ok := stmt.Lhs[0].(*ast.Ident); ok {
			name = id.Name
		}
	}
	for _, exp := range stmt.Rhs {
		if call, ok := exp.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				//fmt.Printf("sel.X=%+v\n", sel.X)
				//fmt.Printf("sel.Sel=%+v\n", *sel.Sel)
				if fmt.Sprintf("%s.%s", sel.X, (*sel.Sel).Name) == "i18n.MustTfunc" {
					return name, true
				}
			}
		}
	}
	return "", false
}

func (v *visitor) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.CallExpr:
		if v.tFunc == "" || v.allStrings == nil {
			return v // Don't do anything until we have tFunc
		}
		call, _ := n.Fun.(*ast.Ident)
		if call != nil && call.Name == v.tFunc {
			for _, a := range n.Args {
				switch b := a.(type) {
				case *ast.BasicLit:
					if b.Kind == token.STRING {
						fmt.Printf("%+v\n", b.Value)
						v.allStrings = append(v.allStrings, b.Value)
					}
				default:
					fmt.Printf("%#v\n", b)
				}
			}
		}
	case *ast.AssignStmt:
		if v.tFunc != "" {
			return v // Don't redefine tFunc if we already have one
		}
		if tFunc, ok := getTFuncName(n); ok {
			fmt.Printf("OK, tfunc = %q\n", tFunc)
			v.tFunc = tFunc
		}
	}
	return v
}

func main() {
	usage := `Sifter. Sifts code for untranslated strings.

Looks in all *.go files under path if path is a directory or treats path as
wildcard if it is not.

Usage:
  sift <path>
  sift -h | --help
  sift --version

Options:
  -h --help     Show this screen.
  --version     Show version.`

	arguments, _ := docopt.Parse(usage, nil, true, "Sifter 0.1", false)
	var v visitor
	allFiles := getAllFiles(arguments["<path>"].(string))
	fmt.Printf("%#v\n", allFiles)
	v.parseAllFiles(allFiles)
	if v.tFunc != "" {
		// Now when we have a tFunc, walk the files again, looking for the
		// strings:
		v.allStrings = make([]string, 0, 0)
		v.parseAllFiles(allFiles)
	} else {
		println("Warning: no Tfunc found!")
	}
	for _, str := range v.allStrings {
		println(str)
	}
	return
}
