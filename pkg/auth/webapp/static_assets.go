package webapp

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

var appLogoPaths = map[string]string{
	"image/png":  "static/app_logo.png",
	"image/jpeg": "static/app_logo.jpeg",
	"image/gif":  "static/app_logo.gif",
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

var AppLogo = resource.RegisterResource(appLogo{})

type appLogo struct{}

func (a appLogo) ReadResource(fs resource.Fs) ([]resource.LayerFile, error) {
	var files []resource.LayerFile
	for _, p := range appLogoPaths {
		data, err := resource.ReadFile(fs, p)
		if os.IsNotExist(err) {
			continue
		} else if err != nil {
			return nil, err
		}
		files = append(files, resource.LayerFile{Path: p, Data: data})
	}
	if len(files) >= 2 {
		return nil, fmt.Errorf("duplicated app logo files: %s, %s", files[0].Path, files[1].Path)
	}
	return files, nil
}

func (a appLogo) MatchResource(path string) bool {
	for _, p := range appLogoPaths {
		if p == path {
			return true
		}
	}
	return false
}

func (a appLogo) Merge(layers []resource.LayerFile, args map[string]interface{}) (*resource.MergedFile, error) {
	return &resource.MergedFile{Data: layers[len(layers)-1].Data}, nil
}

func (a appLogo) Parse(merged *resource.MergedFile) (interface{}, error) {
	mimeType := http.DetectContentType(merged.Data)
	path, ok := appLogoPaths[mimeType]
	if !ok {
		return nil, fmt.Errorf("invalid app logo format: %s", mimeType)
	}
	return &StaticAsset{
		Path: path,
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
