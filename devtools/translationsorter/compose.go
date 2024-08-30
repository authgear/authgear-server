package main

import (
	"fmt"
	"os"
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
		byteArray := []byte(fmt.Sprintf("%q: %q", kvPair.Key, kvPair.Value))
		_, err = f.Write(byteArray)
		if err != nil {
			return fmt.Errorf("failed to write: %w", err)
		}

	}

	f.Write([]byte("\n}\n"))
	return
}
