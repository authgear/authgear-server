package template

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

type FileLoader struct {
	BaseDirectory string
}

func (l *FileLoader) Load(path string) (templateContent string, err error) {
	if l.BaseDirectory == "" {
		return "", errors.New("cannot resolve template file without base directory")
	}

	path = filepath.Join(l.BaseDirectory, path)

	relPath, err := filepath.Rel(l.BaseDirectory, path)
	if err != nil {
		return "", err
	} else if strings.HasPrefix(relPath, ".."+string(filepath.Separator)) || relPath == ".." {
		return "", fmt.Errorf("resolved template file outside base directory: %s", relPath)
	}

	f, err := os.Open(path)
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
