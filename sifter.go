package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/docopt/docopt.go"
	"github.com/nicksnyder/go-i18n/i18n/translation"
)

type visitor struct {
	allStrings []LocalizedString
	tFunc      string
	fileSet    *token.FileSet
}

type LocalizedString struct {
	String     string
	SourceFile string
	SourceLine int
}

func getAllFiles(siftParam, ext string) []string {
	var files []string
	if dir, err := isDir(siftParam); err == nil && dir {
		filepath.Walk(siftParam, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.Mode().IsRegular() && strings.HasSuffix(path, ext) {
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
		v.fileSet = token.NewFileSet()
		f, err := parser.ParseFile(v.fileSet, fileName, nil, 0)
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
						pos := v.fileSet.Position(b.Pos())
						unquoted, err := strconv.Unquote(b.Value)
						if err != nil {
							panic(err)
						}
						v.allStrings = append(v.allStrings, LocalizedString{
							String:     unquoted,
							SourceFile: pos.Filename,
							SourceLine: pos.Line,
						})
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

func parseTemplates(tmpls []string) []LocalizedString {
	var results []LocalizedString
	re := regexp.MustCompile(`{{L10n (".*")}}`)
	for _, template := range tmpls {
		data, err := ioutil.ReadFile(template)
		if err != nil {
			panic(err) // XXX: better error handling
		}
		m := re.FindAllSubmatch(data, -1)
		for _, i := range m {
			unquoted, err := strconv.Unquote(string(i[1]))
			if err != nil {
				panic(err)
			}
			results = append(results, LocalizedString{
				String:     unquoted,
				SourceFile: template,
				SourceLine: 0,
			})
		}
	}
	return results
}

func loadGoi18nJson(path string) (map[string]translation.Translation, error) {
	fileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var translationsData []map[string]interface{}
	if len(fileBytes) > 0 {
		if err := json.Unmarshal(fileBytes, &translationsData); err != nil {
			return nil, err
		}
	}
	translations := make([]translation.Translation, 0, len(translationsData))
	for i, translationData := range translationsData {
		t, err := translation.NewTranslation(translationData)
		if err != nil {
			return nil, fmt.Errorf("unable to parse translation #%d in %s because %s\n%v", i, path, err, translationData)
		}
		translations = append(translations, t)
	}
	result := make(map[string]translation.Translation)
	for _, t := range translations {
		result[t.ID()] = t
	}
	return result, nil
}

func mkUntranslatedName(json string) string {
	parts := strings.Split(json, ".")
	return fmt.Sprintf("%s.untranslated.json", parts[0])
}

func marshalInterface(translations []translation.Translation) []interface{} {
	mi := make([]interface{}, len(translations))
	for i, translation := range translations {
		mi[i] = translation.MarshalInterface()
	}
	return mi
}

func writeUntranslated(filename string, untranslated []translation.Translation) error {
	buf, err := json.MarshalIndent(marshalInterface(untranslated), "", "  ")
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(filename, buf, 0666); err != nil {
		return fmt.Errorf("failed to write %s because %s", filename, err)
	}
	return nil
}

func sift(goPath, templatesPath string) []LocalizedString {
	var v visitor
	allFiles := getAllFiles(goPath, ".go")
	fmt.Printf("%#v\n", allFiles)
	v.parseAllFiles(allFiles)
	if v.tFunc == "" {
		println("Error: no Tfunc found!")
		return nil
	}
	// Now when we have a tFunc, walk the files again, looking for the strings:
	v.allStrings = make([]LocalizedString, 0, 0)
	v.parseAllFiles(allFiles)
	// Also sift templates:
	allTemplates := getAllFiles(templatesPath, ".html")
	v.allStrings = append(v.allStrings, parseTemplates(allTemplates)...)
	for _, str := range v.allStrings {
		fmt.Printf("%s (%d): %q\n", str.SourceFile, str.SourceLine, str.String)
	}
	return v.allStrings
}

func filterUntranslated(translated map[string]translation.Translation,
	sifted []LocalizedString) []translation.Translation {
	untranslated := make([]translation.Translation, 0, 0)
	for _, xl := range sifted {
		_, ok := translated[xl.String]
		if !ok {
			fmt.Printf("%s\n", xl.String)
			t, err := translation.NewTranslation(map[string]interface{}{
				"id":          xl.String,
				"translation": "",
			})
			if err != nil {
				panic(err)
			}
			untranslated = append(untranslated, t)
		}
	}
	return untranslated
}

func main() {
	usage := `Sifter. Sifts code for untranslated strings.

Looks in all *.go files under path if path is a directory or treats path as
wildcard if it is not.

Usage:
  sift <path> <tmpl> <json>
  sift -h | --help
  sift --version

Options:
  -h --help     Show this screen.
  --version     Show version.`

	arguments, _ := docopt.Parse(usage, nil, true, "Sifter 0.1", false)
	allStrings := sift(arguments["<path>"].(string), arguments["<tmpl>"].(string))
	json := arguments["<json>"].(string)
	translated, err := loadGoi18nJson(json)
	if err != nil {
		panic(err) // XXX: better error handling
	}
	untranslated := filterUntranslated(translated, allStrings)
	filename := mkUntranslatedName(path.Base(json))
	err = writeUntranslated(filename, untranslated)
	if err != nil {
		panic(err)
	}
	return
}
