package web

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/intl"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

const ResourceArgPreferredLanguageTag = "preferred_language_tag"
const ResourceArgDefaultLanguageTag = "default_language_tag"
const ResourceArgRequestedPath = "requested_path"

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
	Context            context.Context
	Config             *config.HTTPConfig
	Localization       *config.LocalizationConfig
	StaticAssetsPrefix config.StaticAssetURLPrefix
	Resources          ResourceManager
}

func (r *StaticAssetResolver) StaticAssetURL(id string) (string, error) {
	desc, ok := StaticAssetResources[id]
	if !ok {
		return "", fmt.Errorf("unknown static asset: %s", id)
	}

	preferredLanguageTags := intl.GetPreferredLanguageTags(r.Context)
	merged, err := r.Resources.Read(desc, map[string]interface{}{
		ResourceArgPreferredLanguageTag: preferredLanguageTags,
		ResourceArgDefaultLanguageTag:   r.Localization.FallbackLanguage,
	})
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
