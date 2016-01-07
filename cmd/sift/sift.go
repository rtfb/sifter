package main

import (
	"fmt"
	"path"
	"strings"

	"github.com/docopt/docopt-go"
	"github.com/rtfb/sifter/sifter"
)

const (
	usage = `Sifter. Sifts code for untranslated strings.

Looks in all *.go files under path if path is a directory or treats path as
wildcard if it is not.

Usage:
  sift <path> <tmpl> <json>
  sift -h | --help
  sift --version

Options:
  -h --help     Show this screen.
  --version     Show version.`
)

func mkUntranslatedName(json string) string {
	parts := strings.Split(json, ".")
	return fmt.Sprintf("%s.untranslated.json", parts[0])
}

func writeResult(file string, strings sifter.StringMap) {
	filename := mkUntranslatedName(path.Base(file))
	err := sifter.WriteUntranslated(filename, strings)
	if err != nil {
		panic(err)
	}
}

func main() {
	arguments, _ := docopt.Parse(usage, nil, true, "Sifter 0.1", false)
	allStrings := sifter.Sift(arguments["<path>"].(string), arguments["<tmpl>"].(string))
	json := arguments["<json>"].(string)
	translated, err := sifter.LoadGoi18nJson(json)
	if err != nil {
		panic(err) // XXX: better error handling
	}
	writeResult(json, sifter.FilterUntranslated(translated, allStrings))
	return
}
