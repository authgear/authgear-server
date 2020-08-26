package template

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/authgear/authgear-server/pkg/util/fs"
)

var ErrUnknownURIScheme = errors.New("template: unknown URI scheme")

type URILoader struct {
	FileLoader *FileLoader
	DataLoader *DataLoader
}

func NewURILoader(fs fs.Fs) *URILoader {
	return &URILoader{
		FileLoader: &FileLoader{Fs: fs},
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
