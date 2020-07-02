package template

import (
	"net/url"

	"github.com/authgear/authgear-server/pkg/core/errors"
)

var ErrUnknownURIScheme = errors.New("unknown URI scheme")

type URILoader struct {
	FileLoader *FileLoader
	DataLoader *DataLoader
}

func NewURILoader() *URILoader {
	return &URILoader{
		FileLoader: &FileLoader{},
		DataLoader: &DataLoader{},
	}
}

func (l *URILoader) Load(uri string) (templateContent string, err error) {
	u, err := url.Parse(uri)
	if err != nil {
		err = errors.HandledWithMessage(err, "failed to parse template URI")
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
