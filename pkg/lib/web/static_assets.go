package web

import (
	"bytes"
	"fmt"
	"net/http"
	"os"

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

var imageExtensions = map[string]string{
	"image/png":  ".png",
	"image/jpeg": ".jpeg",
	"image/gif":  ".gif",
}

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

type imageAsset struct {
	Name string
}

func (a imageAsset) ReadResource(fs resource.Fs) ([]resource.LayerFile, error) {
	var files []resource.LayerFile
	for _, ext := range imageExtensions {
		path := a.Name + ext
		data, err := resource.ReadFile(fs, path)
		if os.IsNotExist(err) {
			continue
		} else if err != nil {
			return nil, err
		}
		files = append(files, resource.LayerFile{Path: path, Data: data})
	}
	if len(files) >= 2 {
		return nil, fmt.Errorf("duplicated image files: %s, %s", files[0].Path, files[1].Path)
	}
	return files, nil
}

func (a imageAsset) MatchResource(path string) bool {
	for _, ext := range imageExtensions {
		if path == a.Name+ext {
			return true
		}
	}
	return false
}

func (a imageAsset) Merge(layers []resource.LayerFile, args map[string]interface{}) (*resource.MergedFile, error) {
	return &resource.MergedFile{Data: layers[len(layers)-1].Data}, nil
}

func (a imageAsset) Parse(merged *resource.MergedFile) (interface{}, error) {
	mimeType := http.DetectContentType(merged.Data)
	ext, ok := imageExtensions[mimeType]
	if !ok {
		return nil, fmt.Errorf("invalid image format: %s", mimeType)
	}
	return &StaticAsset{
		Path: a.Name + ext,
		Data: merged.Data,
	}, nil
}

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
