package main

import (
	"fmt"

	"github.com/docopt/docopt.go"
)

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
	fmt.Println(arguments)
	return
}
