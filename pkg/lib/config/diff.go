package config

import (
	"encoding/json"

	"github.com/yudai/gojsondiff"
	diffformatter "github.com/yudai/gojsondiff/formatter"
)

func DiffAppConfig(originalConfig *AppConfig, newConfig *AppConfig) (string, error) {
	oBytes, err := json.Marshal(originalConfig)
	if err != nil {
		return "", err
	}
	nBytes, err := json.Marshal(newConfig)
	if err != nil {
		return "", err
	}
	diff, err := gojsondiff.New().Compare(oBytes, nBytes)
	if err != nil {
		return "", err
	}
	if !diff.Modified() {
		return "", nil
	}
	config := diffformatter.AsciiFormatterConfig{
		ShowArrayIndex: true,
		Coloring:       false,
	}
	var oMap map[string]interface{}
	err = json.Unmarshal(oBytes, &oMap)
	if err != nil {
		return "", err
	}
	formatter := diffformatter.NewAsciiFormatter(oMap, config)
	formattedDiff, err := formatter.Format(diff)
	if err != nil {
		return "", err
	}
	return formattedDiff, nil
}
