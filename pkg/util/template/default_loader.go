package template

import (
	"io/ioutil"
	"os"
	"path"
)

type DefaultLoader interface {
	LoadDefault(templateType string) (string, error)
}

type DefaultLoaderFS struct {
	Directory string
}

func (l *DefaultLoaderFS) LoadDefault(templateType string) (string, error) {
	templatePath := path.Join(l.Directory, templateType)

	data, err := ioutil.ReadFile(templatePath)
	if os.IsNotExist(err) {
		return "", &errNotFound{name: templateType}
	} else if err != nil {
		return "", err
	}

	return string(data), nil
}
