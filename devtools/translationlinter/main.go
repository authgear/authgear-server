package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	gomessageformat "github.com/iawaknahc/gomessageformat"
	"golang.org/x/text/language"
)

const DevelopmentLanguage = "en"

type errMissingKeys struct {
	MissingKeysByLang map[string][]string
}

func (e errMissingKeys) Error() string {
	var buf strings.Builder

	fmt.Fprintf(&buf, "The following languages have missing keys:\n")
	for lang, keys := range e.MissingKeysByLang {
		fmt.Fprintf(&buf, "  %v:\n", lang)
		for _, key := range keys {
			fmt.Fprintf(&buf, "    %v\n", key)
		}
	}

	return buf.String()
}

type errExtraKeys struct {
	ExtraKeysByLang map[string][]string
}

func (e errExtraKeys) Error() string {
	var buf strings.Builder

	fmt.Fprintf(&buf, "The following languages have extra keys:\n")
	for lang, keys := range e.ExtraKeysByLang {
		fmt.Fprintf(&buf, "  %v:\n", lang)
		for _, key := range keys {
			fmt.Fprintf(&buf, "    %v\n", key)
		}
	}

	return buf.String()
}

type InvalidKey struct {
	Key   string
	Error error
}

type errInvalidKeys struct {
	InvalidKeysByLang map[string][]InvalidKey
}

func (e errInvalidKeys) Error() string {
	var buf strings.Builder

	fmt.Fprintf(&buf, "The following languages have invalid keys:\n")
	for lang, invalidKeys := range e.InvalidKeysByLang {
		fmt.Fprintf(&buf, "  %v\n", lang)
		for _, invalidKey := range invalidKeys {
			fmt.Fprintf(&buf, "    %v: %v\n", invalidKey.Key, invalidKey.Error)
		}
	}

	return buf.String()
}

func check(expectedKeys map[string]struct{}, jsonObject map[string]string) (missingKeys []string, extraKeys []string) {
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

	// Prepare the expected keys.
	expectedKeys := make(map[string]struct{})
	for _, match := range matches {
		langTag := filepath.Base(filepath.Dir(match))
		// The keys appearing in DevelopmentLanguage are the expected keys.
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
	invalidKeysByLang := make(map[string][]InvalidKey)

	for _, match := range matches {
		langTag := filepath.Base(filepath.Dir(match))
		tag := language.Make(langTag)

		var f *os.File
		f, err = os.Open(match)
		if err != nil {
			return
		}
		defer f.Close()

		var jsonObject map[string]string
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

		for key, pattern := range jsonObject {
			_, parseErr := gomessageformat.FormatTemplateParseTree(tag, pattern)
			if parseErr != nil {
				invalidKeysByLang[langTag] = append(invalidKeysByLang[langTag], InvalidKey{
					Key:   key,
					Error: parseErr,
				})
			}
		}
	}

	if len(missingKeysByLang) > 0 {
		err = errors.Join(err, errMissingKeys{
			MissingKeysByLang: missingKeysByLang,
		})
	}

	if len(extraKeysByLang) > 0 {
		err = errors.Join(err, errExtraKeys{
			ExtraKeysByLang: extraKeysByLang,
		})
	}
	if len(invalidKeysByLang) > 0 {
		err = errors.Join(err, errInvalidKeys{
			InvalidKeysByLang: invalidKeysByLang,
		})
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
