package web

import (
	"context"
	// nolint:gosec
	"crypto/md5"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/intl"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

var StaticAssetResources = map[string]resource.Descriptor{
	"web-js":             WebJS,
	"password-policy-js": PasswordPolicyJS,
	"app-logo":           AppLogo,
	"app-logo-dark":      AppLogoDark,
	"favicon":            Favicon,
	"avatar-placeholder": AvatarPlaceholder,

	"authgear.css":             AuthgearCSS,
	"authgear-light-theme.css": AuthgearLightThemeCSS,
	"authgear-dark-theme.css":  AuthgearDarkThemeCSS,

	"normalize.min.css":     NormalizeCSS,
	"normalize.min.css.map": NormalizeCSSMap,

	"tabler-icons.min.css":     IconsCSS,
	"fonts/tabler-icons.eot":   IconsFontEOT,
	"fonts/tabler-icons.svg":   IconsFontSVG,
	"fonts/tabler-icons.ttf":   IconsFontTTF,
	"fonts/tabler-icons.woff":  IconsFontWOFF,
	"fonts/tabler-icons.woff2": IconsFontWOFF2,

	"intl-tel-input/css/intlTelInput.min.css": IntlTelInputCSS,
	"intl-tel-input/img/flags.png":            IntlTelInputImage,
	"intl-tel-input/img/flags@2x.png":         IntlTelInputImage2X,
	"intl-tel-input/js/intlTelInput.min.js":   IntlTelInputRuntime,
	"intl-tel-input/js/utils.js":              IntlTelInputRealRuntime,
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
	// md5 is used to compute the hash in the filename for caching purpose only
	// nolint:gosec
	hash := md5.Sum(asset.Data)

	hashPath := PathWithHash(assetPath, fmt.Sprintf("%x", hash))
	return staticAssetURL(r.Config.PublicOrigin, string(r.StaticAssetsPrefix), hashPath)
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

func PathWithHash(filePath string, hash string) string {
	extension := path.Ext(filePath)
	nameOnly := strings.TrimSuffix(filePath, extension)
	return fmt.Sprintf("%s.%s%s", nameOnly, hash, extension)
}

func ParsePathWithHash(hashedPath string) (filePath string, hash string) {
	extension := path.Ext(hashedPath)
	if extension == "" {
		return "", ""
	}

	nameWithHash := strings.TrimSuffix(hashedPath, extension)
	dotIdx := strings.LastIndex(nameWithHash, ".")
	if dotIdx == -1 {
		// hashedPath doesn't have extension, e.g. filename.hash
		// so the extension is the hashed
		filePath = nameWithHash
		hash = strings.TrimPrefix(extension, ".")
		return
	}

	nameOnly := nameWithHash[:dotIdx]

	hash = nameWithHash[dotIdx+1:]
	filePath = fmt.Sprintf("%s%s", nameOnly, extension)

	return
}
