package main

import (
	"fmt"
	"os"
	"path/filepath"
)

type KeyValuePair struct {
	Key   string
	Value string
}

func doMain() (err error) {
	matches, err := filepath.Glob("./resources/authgear/templates/*/translation.json")
	if err != nil {
		return
	}

	for _, match := range matches {
		var f *os.File
		f, err = os.Open(match)
		if err != nil {
			return
		}
		defer f.Close()

		keyValuePairs := GetKeyValuePairs(f)
		// TODO: sort the pairs
		fmt.Printf("%v\n", keyValuePairs)
	}

	return
}

func main() {
	err := doMain()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
