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
	"tabler-icons.min.css":     IconsCSS,
	"fonts/tabler-icons.eot":   IconsFontEOT,
	"fonts/tabler-icons.svg":   IconsFontSVG,
	"fonts/tabler-icons.ttf":   IconsFontTTF,
	"fonts/tabler-icons.woff":  IconsFontWOFF,
	"fonts/tabler-icons.woff2": IconsFontWOFF2,
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
		SupportedTags: r.Localization.SupportedLanguages,
		DefaultTag:    *r.Localization.FallbackLanguage,
		PreferredTags: preferredLanguageTags,
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
