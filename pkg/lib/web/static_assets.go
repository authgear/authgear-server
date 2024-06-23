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

var AppLogo = resource.RegisterResource(ImageDescriptor{Name: "app_logo"})
var AppLogoDark = resource.RegisterResource(ImageDescriptor{Name: "app_logo_dark"})
var Favicon = resource.RegisterResource(ImageDescriptor{Name: "favicon"})
var AppBackgroundImage = resource.RegisterResource(ImageDescriptor{Name: "app_background_image"})
var AppBackgroundImageDark = resource.RegisterResource(ImageDescriptor{Name: "app_background_image_dark"})

var AuthgearAuthflowV2LightThemeCSS = resource.RegisterResource(CSSDescriptor{
	Path: path.Join(AppAssetsURLDirname, "authgear-authflowv2-light-theme.css"),
})

var AuthgearAuthflowV2DarkThemeCSS = resource.RegisterResource(CSSDescriptor{
	Path: path.Join(AppAssetsURLDirname, "authgear-authflowv2-dark-theme.css"),
})
