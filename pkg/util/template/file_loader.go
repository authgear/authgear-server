package template

import (
	"fmt"
	"io/ioutil"
	"unicode/utf8"

	"github.com/authgear/authgear-server/pkg/util/fs"
)

type FileLoader struct {
	Fs fs.Fs
}

func (l *FileLoader) Load(path string) (templateContent string, err error) {
	f, err := l.Fs.Open(path)
	if err != nil {
		err = &errNotFound{name: path}
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
