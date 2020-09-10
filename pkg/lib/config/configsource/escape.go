package configsource

import (
	"regexp"
	"strconv"
	"strings"
)

var escapeCharRegex = regexp.MustCompile("[^a-zA-Z-.]")

func EscapePath(path string) string {
	return escapeCharRegex.ReplaceAllStringFunc(path, func(s string) string {
		seq := ""
		for _, c := range s {
			seq += "_" + strconv.FormatInt(int64(c), 16) + "_"
		}
		return seq
	})
}

var unescapeCharRegex = regexp.MustCompile("_([0-9a-fA-F]+)_")

func UnescapePath(path string) (string, error) {
	var unErr error
	s := unescapeCharRegex.ReplaceAllStringFunc(path, func(s string) string {
		codePoint := strings.Trim(s, "_")
		r, err := strconv.ParseInt(codePoint, 16, 32)
		if err != nil {
			unErr = err
			return ""
		}
		return string(rune(r))
	})
	if unErr != nil {
		return "", unErr
	}
	return s, nil
}
