package resources

import (
	"sort"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

func List(r *resource.Manager) ([]string, error) {
	// Find the union all known paths in all FSs.
	filePaths := make(map[string]struct{})
	for _, fs := range r.Fs {
		locations, err := resource.EnumerateAllLocations(fs)
		if err != nil {
			return nil, err
		}
		for _, location := range locations {
			filePaths[location.Path] = struct{}{}
		}
	}

	// Omit paths that are not resources.
	for p := range filePaths {
		found := false
		for _, desc := range r.Registry.Descriptors {
			if _, ok := desc.MatchResource(p); !ok {
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

type DescriptedPath struct {
	Descriptor resource.Descriptor
	Path       string
}

func AssociateDescriptor(r *resource.Manager, paths ...string) ([]DescriptedPath, error) {
	var matches []DescriptedPath
	for _, p := range paths {
		found := false
		for _, desc := range r.Registry.Descriptors {
			if _, ok := desc.MatchResource(p); !ok {
				continue
			}
			matches = append(matches, DescriptedPath{
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
	return matches, nil
}
