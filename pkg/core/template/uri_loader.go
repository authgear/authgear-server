package template

import (
	"net/url"

	"github.com/skygeario/skygear-server/pkg/core/errors"
)

var ErrUnknownURIScheme = errors.New("unknown URI scheme")

type URILoader struct {
	FileLoader      *FileLoader
	AssetGearLoader *AssetGearLoader
}

func NewURILoader(enabledFileLoader bool) *URILoader {
	return &URILoader{
		FileLoader: &FileLoader{
			Enabled: enabledFileLoader,
		},
		AssetGearLoader: &AssetGearLoader{},
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
	case "asset-gear":
		templateContent, err = l.AssetGearLoader.Load(u)
		return
	default:
		err = ErrUnknownURIScheme
		return
	}
}
