package hook

import (
	"bytes"
	"context"
	"fmt"
	"regexp"

	"github.com/authgear/authgear-server/pkg/util/resource"
)

var DenoFileFilenameRegexp = regexp.MustCompile(`^deno/[0-9a-zA-Z]+\.ts$`)

type ResourceManager interface {
	Read(desc resource.Descriptor, view resource.View) (interface{}, error)
}

type DenoFileDescriptor struct{}

func (DenoFileDescriptor) MatchResource(path string) (*resource.Match, bool) {
	if DenoFileFilenameRegexp.MatchString(path) {
		return &resource.Match{}, true
	}
	return nil, false
}

func (d DenoFileDescriptor) FindResources(fs resource.Fs) ([]resource.Location, error) {
	allLocations, err := resource.EnumerateAllLocations(fs)
	if err != nil {
		return nil, err
	}

	var locations []resource.Location
	for _, location := range allLocations {
		_, ok := d.MatchResource(location.Path)
		if ok {
			l := location
			locations = append(locations, l)
		}
	}

	return locations, nil
}

func (DenoFileDescriptor) ViewResources(resources []resource.ResourceFile, rawView resource.View) (interface{}, error) {
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

	switch rawView.(type) {
	case resource.AppFileView:
		err := app()
		if err != nil {
			return nil, err
		}
		return output.Bytes(), nil
	default:
		return nil, fmt.Errorf("unsupported view: %T", rawView)
	}
}

func (DenoFileDescriptor) UpdateResource(_ context.Context, _ []resource.ResourceFile, _ *resource.ResourceFile, _ []byte) (*resource.ResourceFile, error) {
	return nil, fmt.Errorf("not yet implemented")
}

var DenoFile = resource.RegisterResource(DenoFileDescriptor{})
