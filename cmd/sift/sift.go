package main

import (
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

func main() {
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
