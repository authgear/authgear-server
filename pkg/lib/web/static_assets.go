package web

import (
	"github.com/authgear/authgear-server/pkg/util/resource"
)

const StaticAssetResourcePrefix = "static/"

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

var AppLogo = resource.RegisterResource(ImageDescriptor{Name: "app_logo"})
var AppLogoDark = resource.RegisterResource(ImageDescriptor{Name: "app_logo_dark"})
var Favicon = resource.RegisterResource(ImageDescriptor{Name: "favicon"})
