package web

import (
	"context"
	"fmt"
	"os"

	"github.com/authgear/authgear-server/pkg/util/resource"
)

type SourceMapDescriptor struct {
	Path string
}

var _ resource.Descriptor = SourceMapDescriptor{}

func (d SourceMapDescriptor) MatchResource(path string) (*resource.Match, bool) {
	if path == d.Path {
		return &resource.Match{}, true
	}
	return nil, false
}

func (d SourceMapDescriptor) FindResources(fs resource.Fs) ([]resource.Location, error) {
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

func (d SourceMapDescriptor) ViewResources(resources []resource.ResourceFile, rawView resource.View) (interface{}, error) {
	switch rawView.(type) {
	case resource.AppFileView:
		var appResources []resource.ResourceFile
		for _, resrc := range resources {
			if resrc.Location.Fs.GetFsLevel() == resource.FsLevelApp {
				s := resrc
				appResources = append(appResources, s)
			}
		}
		return d.viewResources(appResources)
	case resource.EffectiveFileView:
		return d.viewResources(resources)
	case resource.EffectiveResourceView:
		b, err := d.viewResources(resources)
		if err != nil {
			return nil, err
		}
		return &StaticAsset{
			Path: d.Path,
			Data: b,
		}, nil
	case resource.ValidateResourceView:
		b, err := d.viewResources(resources)
		if err != nil {
			return nil, err
		}
		return &StaticAsset{
			Path: d.Path,
			Data: b,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported view: %T", rawView)
	}
}

func (d SourceMapDescriptor) viewResources(resources []resource.ResourceFile) ([]byte, error) {
	if len(resources) == 0 {
		return nil, resource.ErrResourceNotFound
	}
	last := resources[len(resources)-1]
	return last.Data, nil
}

func (d SourceMapDescriptor) UpdateResource(_ context.Context, _ []resource.ResourceFile, resrc *resource.ResourceFile, data []byte) (*resource.ResourceFile, error) {
	return &resource.ResourceFile{
		Location: resrc.Location,
		Data:     data,
	}, nil
}
