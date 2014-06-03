package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/docopt/docopt.go"
)

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
	allFiles := getAllFiles(arguments["<path>"].(string))
	fmt.Printf("%#v\n", allFiles)
	return
}
