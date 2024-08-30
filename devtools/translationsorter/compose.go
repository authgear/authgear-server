package main

import (
	"fmt"
	"os"
	"strings"
)

func ComposeJSON(kvPairs []KeyValuePair, path string) (err error) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	f.Write([]byte("{\n"))

	for i, kvPair := range kvPairs {
		lineStart := ",\n"
		if i == 0 {
			lineStart = ""
		}
		if _, err = f.Write([]byte(fmt.Sprintf(lineStart + "  "))); err != nil {
			return fmt.Errorf("failed to write: %w", err)
		}
		value := escapeSpecialChars(kvPair.Value)
		byteArray := []byte(fmt.Sprintf("\"%s\": \"%s\"", kvPair.Key, value))
		_, err = f.Write(byteArray)
		if err != nil {
			return fmt.Errorf("failed to write: %w", err)
		}

	}

	f.Write([]byte("\n}\n"))
	return
}

func escapeSpecialChars(s string) string {
	// source - https://yourbasic.org/golang/multiline-string/#all-escape-characters
	out := strings.ReplaceAll(s, "\\", "\\\\")
	out = strings.ReplaceAll(out, "\a", "\\a")
	out = strings.ReplaceAll(out, "\b", "\\b")
	out = strings.ReplaceAll(out, "\t", "\\t")
	out = strings.ReplaceAll(out, "\n", "\\n")
	out = strings.ReplaceAll(out, "\f", "\\f")
	out = strings.ReplaceAll(out, "\r", "\\r")
	out = strings.ReplaceAll(out, "\v", "\\v")
	out = strings.ReplaceAll(out, "\"", "\\\"")
	out = strings.ReplaceAll(out, "\n", "\\n")

	return out
}
