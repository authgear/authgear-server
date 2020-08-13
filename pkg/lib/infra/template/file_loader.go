package template

import (
	"io/ioutil"
	"os"
	"unicode/utf8"

	"github.com/authgear/authgear-server/pkg/util/errorutil"
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
		err = errorutil.HandledWithMessage(err, "failed to read template")
		return
	}

	if !utf8.Valid(content) {
		err = errorutil.New("expected content to be UTF-8 encoded")
		return
	}

	templateContent = string(content)
	return
}
