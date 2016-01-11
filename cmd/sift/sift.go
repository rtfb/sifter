package main

import (
	"fmt"
	"io/ioutil"

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

func writeResult(filename string, strings sifter.StringMap) {
	json, err := strings.ToJSON()
	if err != nil {
		fmt.Printf("Failed to serialize strings: %s\n", err)
		return
	}
	if err = ioutil.WriteFile(filename, json, 0666); err != nil {
		fmt.Printf("Failed to write %q because %s\n", filename, err)
		return
	}
}

func main() {
	arguments, _ := docopt.Parse(usage, nil, true, "Sifter 0.1", false)
	goPath := arguments["<path>"].(string)
	tmplPath := arguments["<tmpl>"].(string)
	json := arguments["<json>"].(string)
	sifter := sifter.NewGoI18nSifter(goPath, tmplPath, json)
	allStrings := sifter.Run()
	translated, err := sifter.Load()
	if err != nil {
		panic(err) // XXX: better error handling
	}
	writeResult(sifter.OutputFileName(), sifter.Filter(translated, allStrings))
	return
}
