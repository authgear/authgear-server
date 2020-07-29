package template

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/authgear/authgear-server/pkg/auth/config"
)

type DefaultLoader interface {
	LoadDefault(t config.TemplateItemType) (string, error)
}

type DefaultLoaderFS struct {
	Directory string
}

func (l *DefaultLoaderFS) LoadDefault(t config.TemplateItemType) (string, error) {
	templatePath := path.Join(l.Directory, string(t))

	data, err := ioutil.ReadFile(templatePath)
	if os.IsNotExist(err) {
		return "", &errNotFound{name: string(t)}
	} else if err != nil {
		return "", err
	}

	return string(data), nil
}
