package hook

import (
	"bytes"
	"context"
	"fmt"
	"regexp"

	"github.com/authgear/authgear-server/pkg/util/resource"
)

var DenoFileFilenameRegexp = regexp.MustCompile(`^deno/[.0-9a-zA-Z]+\.ts$`)

//go:generate go tool mockgen -source=resources.go -destination=resources_mock_test.go -package hook

type denoClientContextKeyType struct{}

var ContextKeyDenoClient = denoClientContextKeyType{}

type ResourceManager interface {
	Read(ctx context.Context, desc resource.Descriptor, view resource.View) (interface{}, error)
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

func (DenoFileDescriptor) ViewResources(ctx context.Context, resources []resource.ResourceFile, rawView resource.View) (interface{}, error) {
	output := bytes.Buffer{}

	app := func(p string) error {
		var target *resource.ResourceFile
		for _, resrc := range resources {
			if resrc.Location.Fs.GetFsLevel() == resource.FsLevelApp && resrc.Location.Path == p {
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

	// We have to support AppFileView and EffectiveFileView
	// because the portal assumes every editable resources support these two views.
	switch view := rawView.(type) {
	case resource.AppFileView:
		err := app(view.AppFilePath())
		if err != nil {
			return nil, err
		}
		return output.Bytes(), nil
	case resource.EffectiveFileView:
		err := app(view.EffectiveFilePath())
		if err != nil {
			return nil, err
		}
		return output.Bytes(), nil
	case resource.ValidateResourceView:
		// Actual validation happens in UpdateResource.
		return nil, nil
	default:
		return nil, fmt.Errorf("unsupported view: %T", rawView)
	}
}

func (DenoFileDescriptor) UpdateResource(ctx context.Context, _ []resource.ResourceFile, resrc *resource.ResourceFile, data []byte) (*resource.ResourceFile, error) {
	denoClient := ctx.Value(ContextKeyDenoClient).(*DenoClientImpl)
	err := denoClient.Check(ctx, string(data))
	if err != nil {
		return nil, err
	}
	return &resource.ResourceFile{
		Location: resrc.Location,
		Data:     data,
	}, nil
}

var DenoFile = resource.RegisterResource(DenoFileDescriptor{})
