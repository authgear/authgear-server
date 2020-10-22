package resources

import (
	"sort"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

func List(r *resource.Manager) ([]string, error) {
	filePaths := make(map[string]struct{})
	for _, fs := range r.Fs {
		paths, err := resource.ListFiles(fs)
		if err != nil {
			return nil, err
		}
		for _, p := range paths {
			filePaths[p] = struct{}{}
		}
	}

	// Filter out non-resource file paths in case of local FS
	for p := range filePaths {
		found := false
		for _, desc := range r.Registry.Descriptors {
			if !desc.MatchResource(p) {
				continue
			}
			found = true
			break
		}

		if !found {
			delete(filePaths, p)
		}
	}

	var paths []string
	for p := range filePaths {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	return paths, nil
}

type Resource struct {
	Descriptor resource.Descriptor
	Path       string
	Files      []File
}

type File struct {
	Fs   resource.Fs
	Data []byte
}

func Load(r *resource.Manager, paths ...string) ([]Resource, error) {
	type matchedResource struct {
		Path       string
		Descriptor resource.Descriptor
	}

	// Match input path with corresponding resource descriptors
	var matches []matchedResource
	for _, p := range paths {
		found := false
		for _, desc := range r.Registry.Descriptors {
			if !desc.MatchResource(p) {
				continue
			}
			matches = append(matches, matchedResource{
				Path:       p,
				Descriptor: desc,
			})
			found = true
			break
		}
		if !found {
			return nil, apierrors.NewInvalid("unknown resource: " + p)
		}
	}

	// Load resource file layers for each match
	var resources []Resource
	for _, match := range matches {
		var files []File
		for _, fs := range r.Fs {
			layers, err := match.Descriptor.ReadResource(fs)
			if err != nil {
				return nil, err
			}

			for _, l := range layers {
				if l.Path != match.Path {
					continue
				}
				files = append(files, File{Fs: fs, Data: l.Data})
			}
		}

		resources = append(resources, Resource{
			Descriptor: match.Descriptor,
			Path:       match.Path,
			Files:      files,
		})
	}

	return resources, nil
}
