package web

import (
	"bytes"

	"github.com/authgear/authgear-server/pkg/util/resource"
)

const StaticAssetResourcePrefix = "static/"

const (
	webJSName    = "authgear.js"
	webCSSName   = "authgear.css"
	zxcvbnJSName = "zxcvbn.js"
)

const (
	appLogoNamePrefix   = "app_logo"
	appBannerNamePrefix = "app_banner"
)

type StaticAsset struct {
	Path string
	Data []byte
}

var WebJS = resource.RegisterResource(resource.SimpleFile{
	Name: StaticAssetResourcePrefix + webJSName,
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
			Path: StaticAssetResourcePrefix + webJSName,
			Data: data,
		}, nil
	},
})

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

// zxcvbn version commit hash: 67c4ece9efc40c9d0a1d7d995b2b22a91be500c2

var ZxcvbnJS = resource.RegisterResource(resource.SimpleFile{
	Name: StaticAssetResourcePrefix + zxcvbnJSName,
	ParseFn: func(data []byte) (interface{}, error) {
		return &StaticAsset{
			Path: StaticAssetResourcePrefix + zxcvbnJSName,
			Data: data,
		}, nil
	},
})
