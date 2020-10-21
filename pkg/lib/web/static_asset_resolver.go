package web

import (
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

var StaticAssetResources = map[string]resource.Descriptor{
	"web-js":     WebJS,
	"web-css":    WebCSS,
	"app-logo":   AppLogo,
	"app-banner": AppBanner,
	"zxcvbn-js":  ZxcvbnJS,
}

type ResourceManager interface {
	Read(desc resource.Descriptor, args map[string]interface{}) (*resource.MergedFile, error)
}

type StaticAssetResolver struct {
	Config             *config.HTTPConfig
	StaticAssetsPrefix config.StaticAssetURLPrefix
	Resources          ResourceManager
}

func (r *StaticAssetResolver) StaticAssetURL(id string) (string, error) {
	desc, ok := StaticAssetResources[id]
	if !ok {
		return "", fmt.Errorf("unknown static asset: %s", id)
	}

	merged, err := r.Resources.Read(desc, nil)
	if err != nil {
		return "", err
	}
	asset, err := desc.Parse(merged)
	if err != nil {
		return "", err
	}

	assetPath := strings.TrimPrefix(asset.(*StaticAsset).Path, StaticAssetResourcePrefix)
	origin, err := url.Parse(r.Config.PublicOrigin)
	if err != nil {
		return "", err
	}
	u, err := origin.Parse(string(r.StaticAssetsPrefix))
	if err != nil {
		return "", err
	}
	u.Path = path.Join(u.Path, assetPath)
	return u.String(), nil
}

func staticAssetURL(origin string, prefix string, assetPath string) (string, error) {
	o, err := url.Parse(origin)
	if err != nil {
		return "", err
	}
	u, err := o.Parse(prefix)
	if err != nil {
		return "", err
	}
	u.Path = path.Join(u.Path, assetPath)
	return u.String(), nil
}
