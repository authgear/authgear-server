package web

import (
	"path"

	"github.com/authgear/authgear-server/pkg/util/resource"
)

type StaticAsset struct {
	Path string
	Data []byte
}

var AuthgearLightThemeCSS = resource.RegisterResource(CSSDescriptor{
	Path: path.Join(AppAssetsURLDirname, "authgear-light-theme.css"),
})

var AuthgearDarkThemeCSS = resource.RegisterResource(CSSDescriptor{
	Path: path.Join(AppAssetsURLDirname, "authgear-dark-theme.css"),
})

var AppLogo = resource.RegisterResource(LocaleAwareImageDescriptor{Name: "app_logo"})
var AppLogoDark = resource.RegisterResource(LocaleAwareImageDescriptor{Name: "app_logo_dark"})
var Favicon = resource.RegisterResource(LocaleAwareImageDescriptor{Name: "favicon"})
var AppBackgroundImage = resource.RegisterResource(NonLocaleAwareImageDescriptor{Name: "app_background_image", SizeLimit: 500 * 1024})
var AppBackgroundImageDark = resource.RegisterResource(NonLocaleAwareImageDescriptor{Name: "app_background_image_dark", SizeLimit: 500 * 1024})

var CSRFIOSInsturction = resource.RegisterResource(LocaleAwareImageDescriptor{Name: "csrf-ios-instruction"})
var CSRFAndroidInsturction = resource.RegisterResource(LocaleAwareImageDescriptor{Name: "csrf-android-instruction"})
var CSRFChromeInsturction = resource.RegisterResource(LocaleAwareImageDescriptor{Name: "csrf-chrome-instruction"})
var CSRFSamsungInsturction = resource.RegisterResource(LocaleAwareImageDescriptor{Name: "csrf-samsung-instruction"})

var AuthgearAuthflowV2LightThemeCSS = resource.RegisterResource(CSSDescriptor{
	Path: path.Join(AppAssetsURLDirname, "authgear-authflowv2-light-theme.css"),
})

var AuthgearAuthflowV2DarkThemeCSS = resource.RegisterResource(CSSDescriptor{
	Path: path.Join(AppAssetsURLDirname, "authgear-authflowv2-dark-theme.css"),
})
