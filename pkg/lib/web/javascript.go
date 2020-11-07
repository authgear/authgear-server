package web

import (
	"bytes"
	"os"

	"github.com/authgear/authgear-server/pkg/util/resource"
)

type JavaScriptDescriptor struct {
	Path string
}

var _ resource.Descriptor = JavaScriptDescriptor{}

func (d JavaScriptDescriptor) MatchResource(path string) bool {
	return d.Path == path
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

func (d JavaScriptDescriptor) ViewResources(resources []resource.ResourceFile, _ resource.View) (interface{}, error) {
	output := bytes.Buffer{}
	for _, resrc := range resources {
		output.WriteString("(function(){")
		output.Write(resrc.Data)
		output.WriteString("})();")
	}
	return &StaticAsset{
		Path: d.Path,
		Data: output.Bytes(),
	}, nil
}

func (d JavaScriptDescriptor) UpdateResource(resrc *resource.ResourceFile, data []byte, _ resource.View) (*resource.ResourceFile, error) {
	return &resource.ResourceFile{
		Location: resrc.Location,
		Data:     data,
	}, nil
}
