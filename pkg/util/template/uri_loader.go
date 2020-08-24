package template

import (
	"errors"
	"fmt"
	"net/url"
)

var ErrUnknownURIScheme = errors.New("template: unknown URI scheme")

type URILoader struct {
	FileLoader *FileLoader
	DataLoader *DataLoader
}

func NewURILoader(baseDirectory string) *URILoader {
	return &URILoader{
		FileLoader: &FileLoader{
			BaseDirectory: baseDirectory,
		},
		DataLoader: &DataLoader{},
	}
}

func (l *URILoader) Load(uri string) (templateContent string, err error) {
	u, err := url.Parse(uri)
	if err != nil {
		err = fmt.Errorf("template: failed to parse URI: %w", err)
		return
	}

	switch u.Scheme {
	case "file":
		templateContent, err = l.FileLoader.Load(u.Path)
		return
	case "data":
		templateContent, err = l.DataLoader.Load(uri)
		return
	default:
		err = ErrUnknownURIScheme
		return
	}
}
