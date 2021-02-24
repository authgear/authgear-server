package resources

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/spf13/afero"

	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

type Update struct {
	Path string
	Data []byte
}

const ConfigFileMaxSize = 100 * 1024

func ApplyUpdates(appID string, appFs resource.Fs, manager *resource.Manager, secretKeyAllowlist []string, updates []Update) ([]*resource.ResourceFile, error) {
	// Validate file size.
	for _, f := range updates {
		if len(f.Data) > ConfigFileMaxSize {
			return nil, fmt.Errorf("invalid resource '%s': too large (%v > %v)", f.Path, len(f.Data), ConfigFileMaxSize)
		}
	}

	// Construct new resource manager.
	newManager, files, err := applyUpdates(manager, appFs, secretKeyAllowlist, updates)
	if err != nil {
		return nil, err
	}

	// Validate resource FS by viewing ValidateResource.
	for _, desc := range newManager.Registry.Descriptors {
		_, err := newManager.Read(desc, resource.ValidateResource{})
		if err != nil {
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

func applyUpdates(manager *resource.Manager, appFs resource.Fs, secretKeyAllowlist []string, updates []Update) (*resource.Manager, []*resource.ResourceFile, error) {
	newFs, err := cloneFS(appFs)
	if err != nil {
		return nil, nil, err
	}

	newAppFs := resource.AferoFs{Fs: newFs, IsAppFs: true}

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

func cloneFS(fs resource.Fs) (afero.Fs, error) {
	memory := afero.NewMemMapFs()
	locations, err := resource.EnumerateAllLocations(fs)
	if err != nil {
		return nil, err
	}

	for _, location := range locations {
		err := func() error {
			f, err := fs.Open(location.Path)
			if err != nil {
				return err
			}
			defer f.Close()

			data, err := ioutil.ReadAll(f)
			if err != nil {
				return err
			}

			_ = memory.MkdirAll(path.Dir(location.Path), 0666)
			_ = afero.WriteFile(memory, location.Path, data, 0666)
			return nil
		}()
		if err != nil {
			return nil, err
		}
	}

	return memory, nil
}
