package web

import (
	"github.com/authgear/authgear-server/pkg/util/resource"
)

const StaticAssetResourcePrefix = "static/"
const StaticAssetFontResourcePrefix = "static/fonts/"

type StaticAsset struct {
	Path string
	Data []byte
}

var WebJS = resource.RegisterResource(JavaScriptDescriptor{
	Path: StaticAssetResourcePrefix + "authgear.js",
})

var PasswordPolicyJS = resource.RegisterResource(JavaScriptDescriptor{
	Path: StaticAssetResourcePrefix + "password-policy.js",
})

var AuthgearLightThemeCSS = resource.RegisterResource(CSSDescriptor{
	Path: StaticAssetResourcePrefix + "authgear-light-theme.css",
})

var AuthgearDarkThemeCSS = resource.RegisterResource(CSSDescriptor{
	Path: StaticAssetResourcePrefix + "authgear-dark-theme.css",
})

var AuthgearCSS = resource.RegisterResource(CSSDescriptor{
	Path: StaticAssetResourcePrefix + "authgear.css",
})

// NormalizeCSS - normalize.css v8.0.1
var NormalizeCSS = resource.RegisterResource(CSSDescriptor{
	Path: StaticAssetResourcePrefix + "normalize.min.css",
})

var NormalizeCSSMap = resource.RegisterResource(resource.SimpleDescriptor{
	Path: StaticAssetResourcePrefix + "normalize.min.css.map",
})

// IconsCSS - Tabler Icons 1.41.1 by tabler - https://tabler.io
var IconsCSS = resource.RegisterResource(CSSDescriptor{
	Path: StaticAssetResourcePrefix + "tabler-icons.min.css",
})

var IconsFontEOT = resource.RegisterResource(resource.SimpleDescriptor{
	Path: StaticAssetFontResourcePrefix + "tabler-icons.eot",
})

var IconsFontSVG = resource.RegisterResource(resource.SimpleDescriptor{
	Path: StaticAssetFontResourcePrefix + "tabler-icons.svg",
})

var IconsFontTTF = resource.RegisterResource(resource.SimpleDescriptor{
	Path: StaticAssetFontResourcePrefix + "tabler-icons.ttf",
})

var IconsFontWOFF = resource.RegisterResource(resource.SimpleDescriptor{
	Path: StaticAssetFontResourcePrefix + "tabler-icons.woff",
})

var IconsFontWOFF2 = resource.RegisterResource(resource.SimpleDescriptor{
	Path: StaticAssetFontResourcePrefix + "tabler-icons.woff2",
})

var AppLogo = resource.RegisterResource(ImageDescriptor{Name: "app_logo"})
var AppLogoDark = resource.RegisterResource(ImageDescriptor{Name: "app_logo_dark"})
var Favicon = resource.RegisterResource(ImageDescriptor{Name: "favicon"})
