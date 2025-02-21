package resource

import (
	"bytes"
	"context"
	"fmt"
	"os"
)

type Location struct {
	Fs   Fs
	Path string
}

// nolint: golint
type ResourceFile struct {
	Location Location
	Data     []byte
}

type Match struct {
	LanguageTag string
}

type DescriptedPath struct {
	Descriptor Descriptor
	Path       string
}

type Descriptor interface {
	MatchResource(path string) (*Match, bool)
	FindResources(fs Fs) ([]Location, error)
	ViewResources(ctx context.Context, resources []ResourceFile, view View) (interface{}, error)
	UpdateResource(ctx context.Context, resourcesInAllFss []ResourceFile, resourceInTargetFs *ResourceFile, data []byte) (*ResourceFile, error)
}

// SimpleDescriptor does not support view.
type SimpleDescriptor struct {
	Path string
}

var _ Descriptor = SimpleDescriptor{}

func (d SimpleDescriptor) MatchResource(path string) (*Match, bool) {
	if path == d.Path {
		return &Match{}, true
	}
	return nil, false
}

func (d SimpleDescriptor) FindResources(fs Fs) ([]Location, error) {
	location := Location{
		Fs:   fs,
		Path: d.Path,
	}
	_, err := ReadLocation(location)
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return []Location{location}, nil
}

func (d SimpleDescriptor) ViewResources(ctx context.Context, resources []ResourceFile, rawView View) (interface{}, error) {
	switch rawView.(type) {
	case AppFileView:
		var appResources []ResourceFile
		for _, resrc := range resources {
			if resrc.Location.Fs.GetFsLevel() == FsLevelApp {
				s := resrc
				appResources = append(appResources, s)
			}
		}
		return d.viewResources(appResources)
	case EffectiveFileView:
		return d.viewResources(resources)
	case EffectiveResourceView:
		return d.viewResources(resources)
	case ValidateResourceView:
		return d.viewResources(resources)
	default:
		return nil, fmt.Errorf("unsupported view: %T", rawView)
	}
}

func (d SimpleDescriptor) viewResources(resources []ResourceFile) (interface{}, error) {
	if len(resources) == 0 {
		return nil, ErrResourceNotFound
	}
	last := resources[len(resources)-1]
	return last.Data, nil
}

func (d SimpleDescriptor) UpdateResource(_ context.Context, _ []ResourceFile, resource *ResourceFile, data []byte) (*ResourceFile, error) {
	return &ResourceFile{
		Location: resource.Location,
		Data:     data,
	}, nil
}

type NewlineJoinedDescriptor struct {
	Path  string
	Parse func([]byte) (interface{}, error)
}

var _ Descriptor = NewlineJoinedDescriptor{}

func (d NewlineJoinedDescriptor) MatchResource(path string) (*Match, bool) {
	if path == d.Path {
		return &Match{}, true
	}
	return nil, false
}

func (d NewlineJoinedDescriptor) FindResources(fs Fs) ([]Location, error) {
	location := Location{
		Fs:   fs,
		Path: d.Path,
	}
	_, err := ReadLocation(location)
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return []Location{location}, nil
}

func (d NewlineJoinedDescriptor) ViewResources(ctx context.Context, resources []ResourceFile, rawView View) (interface{}, error) {
	switch rawView.(type) {
	case AppFileView:
		var appResources []ResourceFile
		for _, resrc := range resources {
			if resrc.Location.Fs.GetFsLevel() == FsLevelApp {
				s := resrc
				appResources = append(appResources, s)
			}
		}
		return d.viewResources(appResources)
	case EffectiveFileView:
		return d.viewResources(resources)
	case EffectiveResourceView:
		bytes, err := d.viewResources(resources)
		if err != nil {
			return nil, err
		}
		if d.Parse == nil {
			return bytes, nil
		}
		return d.Parse(bytes)
	case ValidateResourceView:
		bytes, err := d.viewResources(resources)
		if err != nil {
			return nil, err
		}
		if d.Parse == nil {
			return bytes, nil
		}
		return d.Parse(bytes)
	default:
		return nil, fmt.Errorf("unsupported view: %T", rawView)
	}
}

func (d NewlineJoinedDescriptor) viewResources(resources []ResourceFile) ([]byte, error) {
	if len(resources) == 0 {
		return nil, ErrResourceNotFound
	}

	output := bytes.Buffer{}
	for _, resrc := range resources {
		output.Write(resrc.Data)
		output.WriteString("\n")
	}
	return output.Bytes(), nil
}

func (d NewlineJoinedDescriptor) UpdateResource(_ context.Context, _ []ResourceFile, resource *ResourceFile, data []byte) (*ResourceFile, error) {
	return &ResourceFile{
		Location: resource.Location,
		Data:     data,
	}, nil
}
