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

var StaticAssetResources = map[string]resource.Descriptor{
	"web-js":                   WebJS,
	"password-policy-js":       PasswordPolicyJS,
	"app-logo":                 AppLogo,
	"app-logo-dark":            AppLogoDark,
	"favicon":                  Favicon,
	"authgear.css":             AuthgearCSS,
	"authgear-light-theme.css": AuthgearLightThemeCSS,
	"authgear-dark-theme.css":  AuthgearDarkThemeCSS,
}

type ResourceManager interface {
	Read(desc resource.Descriptor, view resource.View) (interface{}, error)
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
	result, err := r.Resources.Read(desc, resource.EffectiveResource{
		PreferredTags: preferredLanguageTags,
		DefaultTag:    r.Localization.FallbackLanguage,
	})
	if err != nil {
		return "", err
	}

	asset := result.(*StaticAsset)
	assetPath := strings.TrimPrefix(asset.Path, StaticAssetResourcePrefix)

	return staticAssetURL(r.Config.PublicOrigin, string(r.StaticAssetsPrefix), assetPath)
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
