package web

import (
	"bytes"

	"github.com/authgear/authgear-server/pkg/util/resource"
)

const StaticAssetResourcePrefix = "static/"

const (
	webJSPath    = "static/authgear.js"
	webCSSPath   = "static/authgear.css"
	zxcvbnJSPath = "static/zxcvbn.js"
)

const (
	appLogoFilename   = "static/app_logo"
	appBannerFilename = "static/app_banner"
)

type StaticAsset struct {
	Path string
	Data []byte
}

var WebJS = resource.RegisterResource(resource.SimpleFile{
	Name: webJSPath,
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
			Path: webJSPath,
			Data: data,
		}, nil
	},
})

var WebCSS = resource.RegisterResource(resource.SimpleFile{
	Name: webCSSPath,
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
			Path: webCSSPath,
			Data: data,
		}, nil
	},
})

var AppLogo = resource.RegisterResource(imageAsset{Name: appLogoFilename})
var AppBanner = resource.RegisterResource(imageAsset{Name: appBannerFilename})

// zxcvbn version commit hash: 67c4ece9efc40c9d0a1d7d995b2b22a91be500c2

var ZxcvbnJS = resource.RegisterResource(resource.SimpleFile{
	Name: zxcvbnJSPath,
	ParseFn: func(data []byte) (interface{}, error) {
		return &StaticAsset{
			Path: zxcvbnJSPath,
			Data: data,
		}, nil
	},
})
