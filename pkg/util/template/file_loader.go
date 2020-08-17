package template

import (
	"fmt"
	"io/ioutil"
	"os"
	"unicode/utf8"
)

type FileLoader struct{}

func (l *FileLoader) Load(absolutePath string) (templateContent string, err error) {
	f, err := os.Open(absolutePath)
	if err != nil {
		err = &errNotFound{name: absolutePath}
		return
	}
	defer f.Close()

	content, err := ioutil.ReadAll(f)
	if err != nil {
		err = fmt.Errorf("template: failed to read template: %w", err)
		return
	}

	if !utf8.Valid(content) {
		err = ErrInvalidUTF8
		return
	}

	templateContent = string(content)
	return
}
