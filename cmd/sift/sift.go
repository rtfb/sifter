package main

import (
	"fmt"
	"io/ioutil"
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
	json, err := strings.ToJSON()
	if err != nil {
		fmt.Printf("Failed to serialize strings: %s\n", err)
		return
	}
	filename := mkUntranslatedName(path.Base(file))
	if err = ioutil.WriteFile(filename, json, 0666); err != nil {
		fmt.Printf("Failed to write %q because %s\n", filename, err)
		return
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
