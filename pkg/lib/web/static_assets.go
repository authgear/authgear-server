package web

import (
	"bytes"

	"github.com/authgear/authgear-server/pkg/util/resource"
)

const StaticAssetResourcePrefix = "static/"

const (
	webJSName        = "authgear.js"
	passwordPolicyJS = "password-policy.js"
	webCSSName       = "authgear.css"
)

const (
	appLogoNamePrefix   = "app_logo"
	appBannerNamePrefix = "app_banner"
)

type StaticAsset struct {
	Path string
	Data []byte
}

func makeJSResource(name string) resource.Descriptor {
	return resource.SimpleFile{
		Name: StaticAssetResourcePrefix + name,
		MergeFn: func(layers []resource.LayerFile) ([]byte, error) {
			// Concat JS by wrapping each one in an IIFE
			output := bytes.Buffer{}
			for _, layer := range layers {
				output.WriteString("(function(){\n")
				output.Write(layer.Data)
				output.WriteString("\n})();\n")
			}
			return output.Bytes(), nil
		},
		ParseFn: func(data []byte) (interface{}, error) {
			return &StaticAsset{
				Path: StaticAssetResourcePrefix + name,
				Data: data,
			}, nil
		},
	}
}

var WebJS = resource.RegisterResource(makeJSResource(webJSName))
var PasswordPolicyJS = resource.RegisterResource(makeJSResource(passwordPolicyJS))

var WebCSS = resource.RegisterResource(resource.SimpleFile{
	Name: StaticAssetResourcePrefix + webCSSName,
	MergeFn: func(layers []resource.LayerFile) ([]byte, error) {
		// Concat CSS by simply joining together
		output := bytes.Buffer{}
		for _, layer := range layers {
			output.Write(layer.Data)
			output.WriteString("\n")
		}
		return output.Bytes(), nil
	},
	ParseFn: func(data []byte) (interface{}, error) {
		return &StaticAsset{
			Path: StaticAssetResourcePrefix + webCSSName,
			Data: data,
		}, nil
	},
})

var AppLogo = resource.RegisterResource(imageAsset{Name: appLogoNamePrefix})
var AppBanner = resource.RegisterResource(imageAsset{Name: appBannerNamePrefix})
