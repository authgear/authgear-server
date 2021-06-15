package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const DevelopmentLanguage = "en"

type errKeys struct {
	MissingKeysByLang map[string][]string
	ExtraKeysByLang   map[string][]string
}

func (e errKeys) Error() string {
	var buf strings.Builder

	if len(e.MissingKeysByLang) > 0 {
		fmt.Fprintf(&buf, "The following languages have missing keys:\n")

		for lang, keys := range e.MissingKeysByLang {
			fmt.Fprintf(&buf, "  %v:\n", lang)
			for _, key := range keys {
				fmt.Fprintf(&buf, "    %v\n", key)
			}
		}
	}

	if len(e.ExtraKeysByLang) > 0 {
		fmt.Fprintf(&buf, "The following languages have extra keys:\n")

		for lang, keys := range e.ExtraKeysByLang {
			fmt.Fprintf(&buf, "  %v:\n", lang)
			for _, key := range keys {
				fmt.Fprintf(&buf, "    %v\n", key)
			}
		}
	}

	return buf.String()
}

func check(expectedKeys map[string]struct{}, jsonObject map[string]interface{}) (missingKeys []string, extraKeys []string) {
	for missing := range expectedKeys {
		_, ok := jsonObject[missing]
		if !ok {
			missingKeys = append(missingKeys, missing)
		}
	}

	for extra := range jsonObject {
		_, ok := expectedKeys[extra]
		if !ok {
			extraKeys = append(extraKeys, extra)
		}
	}

	return
}

func doMain() (err error) {
	matches, err := filepath.Glob("./resources/authgear/templates/*/translation.json")
	if err != nil {
		return
	}

	expectedKeys := make(map[string]struct{})
	for _, match := range matches {
		langTag := filepath.Base(filepath.Dir(match))
		if langTag != DevelopmentLanguage {
			continue
		}

		var f *os.File
		f, err = os.Open(match)
		if err != nil {
			return
		}
		defer f.Close()

		var jsonObject map[string]interface{}
		err = json.NewDecoder(f).Decode(&jsonObject)
		if err != nil {
			return
		}

		for key := range jsonObject {
			expectedKeys[key] = struct{}{}
		}
	}

	missingKeysByLang := make(map[string][]string)
	extraKeysByLang := make(map[string][]string)
	for _, match := range matches {
		langTag := filepath.Base(filepath.Dir(match))
		if langTag == DevelopmentLanguage {
			continue
		}

		var f *os.File
		f, err = os.Open(match)
		if err != nil {
			return
		}
		defer f.Close()

		var jsonObject map[string]interface{}
		err = json.NewDecoder(f).Decode(&jsonObject)
		if err != nil {
			return
		}

		missingKeys, extraKeys := check(expectedKeys, jsonObject)
		if len(missingKeys) > 0 {
			missingKeysByLang[langTag] = missingKeys
		}
		if len(extraKeys) > 0 {
			extraKeysByLang[langTag] = extraKeys
		}
	}

	if len(missingKeysByLang) > 0 || len(extraKeysByLang) > 0 {
		err = errKeys{
			MissingKeysByLang: missingKeysByLang,
			ExtraKeysByLang:   extraKeysByLang,
		}
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
