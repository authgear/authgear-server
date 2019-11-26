package template

import (
	"net/url"

	"github.com/skygeario/skygear-server/pkg/core/errors"
)

var ErrUnknownURIScheme = errors.New("unknown URI scheme")

type URILoader struct {
	EnableFileLoader bool
	FileLoader       *FileLoader
	EnableDataLoader bool
	DataLoader       *DataLoader
	AssetGearLoader  *AssetGearLoader
}

func NewURILoader(assetGearLoader *AssetGearLoader) *URILoader {
	return &URILoader{
		FileLoader:      &FileLoader{},
		DataLoader:      &DataLoader{},
		AssetGearLoader: assetGearLoader,
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
		if !l.EnableFileLoader {
			err = &errNotFound{name: u.Path}
			return
		}
		templateContent, err = l.FileLoader.Load(u.Path)
		return
	case "data":
		if !l.EnableDataLoader {
			err = &errNotFound{name: uri}
			return
		}
		templateContent, err = l.DataLoader.Load(uri)
		return
	case "asset-gear":
		if l.AssetGearLoader == nil {
			err = &errNotFound{name: uri}
			return
		}
		templateContent, err = l.AssetGearLoader.Load(u)
		return
	default:
		err = ErrUnknownURIScheme
		return
	}
}
