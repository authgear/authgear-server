package template

import (
	"errors"
	"net/url"
)

type AssetGearLoader struct {
}

func (l *AssetGearLoader) Load(u *url.URL) (templateContent string, err error) {
	// TODO(template): AssetGearLoader
	err = errors.New("not yet implemented")
	return
}
