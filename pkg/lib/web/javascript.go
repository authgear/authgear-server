package web

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/authgear/authgear-server/pkg/util/readcloserthunk"
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
	_, err := resource.StatLocation(location)
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return []resource.Location{location}, nil
}

func (d JavaScriptDescriptor) ViewResources(resources []resource.ResourceFile, rawView resource.View) (interface{}, error) {
	var thunks []readcloserthunk.ReadCloserThunk

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

		thunks = append(thunks, target.ReadCloserThunk)
		return nil
	}

	concat := func() {
		for _, resrc := range resources {
			thunks = append(
				thunks,
				readcloserthunk.Reader(strings.NewReader("(function(){")),
				resrc.ReadCloserThunk,
				readcloserthunk.Reader(strings.NewReader("})();")),
			)
		}
	}

	switch rawView.(type) {
	case resource.AppFileView:
		err := app()
		if err != nil {
			return nil, err
		}
		return readcloserthunk.Performance_Bytes(readcloserthunk.MultiReadCloserThunk(thunks...))
	case resource.EffectiveFileView:
		concat()
		return readcloserthunk.Performance_Bytes(readcloserthunk.MultiReadCloserThunk(thunks...))
	case resource.EffectiveResourceView:
		concat()
		return &StaticAsset{
			Path:            d.Path,
			ReadCloserThunk: readcloserthunk.MultiReadCloserThunk(thunks...),
		}, nil
	case resource.ValidateResourceView:
		concat()
		return &StaticAsset{
			Path:            d.Path,
			ReadCloserThunk: readcloserthunk.MultiReadCloserThunk(thunks...),
		}, nil
	default:
		return nil, fmt.Errorf("unsupported view: %T", rawView)
	}
}

func (d JavaScriptDescriptor) UpdateResource(_ context.Context, _ []resource.ResourceFile, resrc *resource.ResourceFile, data []byte) (*resource.ResourceFile, error) {
	return &resource.ResourceFile{
		Location:        resrc.Location,
		ReadCloserThunk: readcloserthunk.Reader(bytes.NewReader(data)),
	}, nil
}
