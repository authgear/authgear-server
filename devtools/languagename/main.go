package main

// This program takes an installation of cldr-localenames-modern.
// It prints the native name of found languages.
// For languages that do not have their native name found, they are skipped.

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/authgear/authgear-server/pkg/util/intl"
)

func Do(cldrLocalenamesModernDir string) error {
	available := make(map[string]struct{})
	for _, lang := range intl.AvailableLanguages {
		available[lang] = struct{}{}
	}

	// Special cases
	delete(available, "zh-CN")
	delete(available, "zh-HK")
	delete(available, "zh-TW")
	out := map[string]interface{}{
		"language-zh-CN": "简体中文",
		"language-zh-HK": "繁體中文(香港)",
		"language-zh-TW": "繁體中文(台灣)",
	}

	mainDir := filepath.Join(cldrLocalenamesModernDir, "main")
	infos, err := ioutil.ReadDir(mainDir)
	if err != nil {
		return err
	}
	for _, info := range infos {
		lang := info.Name()
		jsonFile, err := os.Open(filepath.Join(mainDir, lang, "languages.json"))
		if err != nil {
			return err
		}
		jsonBytes, err := ioutil.ReadAll(jsonFile)
		if err != nil {
			return err
		}
		var jsonObject map[string]interface{}
		err = json.Unmarshal(jsonBytes, &jsonObject)
		if err != nil {
			return err
		}
		localized, ok := jsonObject["main"].(map[string]interface{})[lang].(map[string]interface{})["localeDisplayNames"].(map[string]interface{})["languages"].(map[string]interface{})[lang].(string)
		if ok {
			_, isAvailable := available[lang]
			if isAvailable {
				key := fmt.Sprintf("language-%s", lang)
				out[key] = localized
				delete(available, lang)
			}
		}
	}

	if len(available) > 0 {
		return fmt.Errorf("No localization found for: %#v", available)
	}

	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}

	fmt.Printf("%v\n", string(b))
	return nil
}

func main() {
	cldrLocalenamesModernDir := os.Args[1]
	err := Do(cldrLocalenamesModernDir)
	if err != nil {
		panic(err)
	}
}
