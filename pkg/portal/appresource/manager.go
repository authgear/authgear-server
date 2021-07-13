package appresource

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sort"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/spf13/afero"
)

const ConfigFileMaxSize = 100 * 1024

type Manager struct {
	AppResourceManager *resource.Manager
	AppFS              resource.Fs
}

func (m *Manager) List() ([]string, error) {
	r := m.AppResourceManager

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

func (m *Manager) AssociateDescriptor(paths ...string) ([]DescriptedPath, error) {
	r := m.AppResourceManager

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

func (m *Manager) ApplyUpdates(appID string, secretKeyAllowlist []string, updates []Update) ([]*resource.ResourceFile, error) {
	// Validate file size.
	for _, f := range updates {
		if len(f.Data) > ConfigFileMaxSize {
			return nil, fmt.Errorf("invalid resource '%s': too large (%v > %v)", f.Path, len(f.Data), ConfigFileMaxSize)
		}
	}

	// Construct new resource manager.
	newManager, files, err := m.applyUpdates(m.AppFS, secretKeyAllowlist, updates)
	if err != nil {
		return nil, err
	}

	// Validate resource FS by viewing ValidateResource.
	for _, desc := range newManager.Registry.Descriptors {
		_, err := newManager.Read(desc, resource.ValidateResource{})
		// Some resource may not have builtin value, e.g. app_logo_dark.
		if errors.Is(err, resource.ErrResourceNotFound) {
			continue
		} else if err != nil {
			return nil, fmt.Errorf("invalid resource: %w", err)
		}
	}

	// Validate configuration.
	cfg, err := configsource.LoadConfig(newManager)
	if err != nil {
		return nil, err
	}

	if string(cfg.AppConfig.ID) != appID {
		return nil, fmt.Errorf("invalid resource '%s': incorrect app ID", configsource.AuthgearYAML)
	}

	return files, nil
}

func (m *Manager) applyUpdates(appFs resource.Fs, secretKeyAllowlist []string, updates []Update) (*resource.Manager, []*resource.ResourceFile, error) {
	manager := m.AppResourceManager

	newFs, err := cloneFS(appFs)
	if err != nil {
		return nil, nil, err
	}

	newAppFs := resource.LeveledAferoFs{Fs: newFs, FsLevel: resource.FsLevelApp}

	var files []*resource.ResourceFile
	for _, u := range updates {
		location := resource.Location{
			Fs:   newAppFs,
			Path: u.Path,
		}

		// Retrieve the original file.
		resrc, err := func() (*resource.ResourceFile, error) {
			f, err := newFs.Open(u.Path)
			if os.IsNotExist(err) {
				return &resource.ResourceFile{
					Location: location,
					Data:     nil,
				}, nil
			} else if err != nil {
				return nil, err
			}
			defer f.Close()

			data, err := ioutil.ReadAll(f)
			if err != nil {
				return nil, err
			}

			return &resource.ResourceFile{
				Location: location,
				Data:     data,
			}, nil
		}()
		if err != nil {
			return nil, nil, err
		}

		desc, ok := manager.Resolve(u.Path)
		if !ok {
			err = fmt.Errorf("invalid resource '%s': unknown resource path", resrc.Location.Path)
			return nil, nil, err
		}

		resrc, err = desc.UpdateResource(resrc, u.Data, resource.AppFile{
			Path:              u.Path,
			AllowedSecretKeys: secretKeyAllowlist,
		})
		if err != nil {
			return nil, nil, err
		}

		if resrc.Data == nil {
			_ = newFs.Remove(resrc.Location.Path)
		} else {
			_ = newFs.MkdirAll(path.Dir(resrc.Location.Path), 0666)
			_ = afero.WriteFile(newFs, resrc.Location.Path, resrc.Data, 0666)
		}

		files = append(files, resrc)
	}

	var newResFs []resource.Fs
	for _, fs := range manager.Fs {
		if fs == appFs {
			newResFs = append(newResFs, newAppFs)
		} else {
			newResFs = append(newResFs, fs)
		}
	}
	return resource.NewManager(manager.Registry, newResFs), files, nil
}
