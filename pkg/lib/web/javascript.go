package web

import (
	"bytes"
	"fmt"
	"os"

	"github.com/authgear/authgear-server/pkg/util/resource"
)

type JavaScriptDescriptor struct {
	Path string
}

var _ resource.Descriptor = JavaScriptDescriptor{}

func (d JavaScriptDescriptor) MatchResource(path string) (*resource.Match, bool) {
	if path == d.Path {
		return &resource.Match{}, true
	}
	return nil, false
}

func (d JavaScriptDescriptor) FindResources(fs resource.Fs) ([]resource.Location, error) {
	location := resource.Location{
		Fs:   fs,
		Path: d.Path,
	}
	_, err := resource.ReadLocation(location)
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return []resource.Location{location}, nil
}

func (d JavaScriptDescriptor) ViewResources(resources []resource.ResourceFile, rawView resource.View) (interface{}, error) {
	output := bytes.Buffer{}

	app := func() error {
		var target *resource.ResourceFile
		for _, resrc := range resources {
			if resrc.Location.Fs.GetFsLevel() == resource.FsLevelApp {
				s := resrc
				target = &s
			}
		}
		if target == nil {
			return resource.ErrResourceNotFound
		}

		output.Write(target.Data)
		return nil
	}

	concat := func() {
		for _, resrc := range resources {
			output.WriteString("(function(){")
			output.Write(resrc.Data)
			output.WriteString("})();")
		}
	}

	switch rawView.(type) {
	case resource.AppFileView:
		err := app()
		if err != nil {
			return nil, err
		}
		return output.Bytes(), nil
	case resource.EffectiveFileView:
		concat()
		return output.Bytes(), nil
	case resource.EffectiveResourceView:
		concat()
		return &StaticAsset{
			Path: d.Path,
			Data: output.Bytes(),
		}, nil
	case resource.ValidateResourceView:
		concat()
		return &StaticAsset{
			Path: d.Path,
			Data: output.Bytes(),
		}, nil
	default:
		return nil, fmt.Errorf("unsupported view: %T", rawView)
	}
}

func (d JavaScriptDescriptor) UpdateResource(resrc *resource.ResourceFile, data []byte, _ resource.View) (*resource.ResourceFile, error) {
	return &resource.ResourceFile{
		Location: resrc.Location,
		Data:     data,
	}, nil
}
