package web

import (
	"github.com/authgear/authgear-server/pkg/util/resource"
)

const StaticAssetResourcePrefix = "static/"

const (
	webJSName        = "authgear.js"
	passwordPolicyJS = "password-policy.js"
	webCSSName       = "authgear.css"
)

const (
	appLogoNamePrefix = "app_logo"
	faviconNamePrefix = "favicon"
)

type StaticAsset struct {
	Path string
	Data []byte
}

var WebJS = resource.RegisterResource(JavaScriptDescriptor{
	Path: StaticAssetResourcePrefix + webJSName,
})

var PasswordPolicyJS = resource.RegisterResource(JavaScriptDescriptor{
	Path: StaticAssetResourcePrefix + passwordPolicyJS,
})

var WebCSS = resource.RegisterResource(CSSDescriptor{
	Path: StaticAssetResourcePrefix + webCSSName,
})

var AppLogo = resource.RegisterResource(ImageDescriptor{Name: appLogoNamePrefix})
var Favicon = resource.RegisterResource(ImageDescriptor{Name: faviconNamePrefix})
