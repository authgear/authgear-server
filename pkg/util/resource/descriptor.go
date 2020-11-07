package resource

import (
	"bytes"
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

type Descriptor interface {
	MatchResource(path string) bool
	FindResources(fs Fs) ([]Location, error)
	ViewResources(resources []ResourceFile, view View) (interface{}, error)
	UpdateResource(resource *ResourceFile, data []byte, view View) (*ResourceFile, error)
}

// SimpleDescriptor does not support view.
type SimpleDescriptor struct {
	Path string
}

var _ Descriptor = SimpleDescriptor{}

func (d SimpleDescriptor) MatchResource(path string) bool {
	return d.Path == path
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

func (d SimpleDescriptor) ViewResources(resources []ResourceFile, _ View) (interface{}, error) {
	last := resources[len(resources)-1]
	return last.Data, nil
}

func (d SimpleDescriptor) UpdateResource(resource *ResourceFile, data []byte, _ View) (*ResourceFile, error) {
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

func (d NewlineJoinedDescriptor) MatchResource(path string) bool {
	return d.Path == path
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

func (d NewlineJoinedDescriptor) ViewResources(resources []ResourceFile, _ View) (interface{}, error) {
	output := bytes.Buffer{}
	for _, resrc := range resources {
		output.Write(resrc.Data)
		output.WriteString("\n")
	}
	return d.Parse(output.Bytes())
}

func (d NewlineJoinedDescriptor) UpdateResource(resource *ResourceFile, data []byte, _ View) (*ResourceFile, error) {
	return &ResourceFile{
		Location: resource.Location,
		Data:     data,
	}, nil
}
