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

func Validate(appID string, appFs resource.Fs, manager *resource.Manager, secretKeyAllowlist []string, updates []Update) error {
	// Validate file size.
	for _, f := range updates {
		if len(f.Data) > ConfigFileMaxSize {
			return fmt.Errorf("invalid resource '%s': too large (%v > %v)", f.Path, len(f.Data), ConfigFileMaxSize)
		}
	}

	// Construct new resource manager.
	newManager, err := applyUpdates(manager, appFs, secretKeyAllowlist, updates)
	if err != nil {
		return err
	}

	// Validate resource FS by viewing EffectiveResource.
	for _, desc := range newManager.Registry.Descriptors {
		_, err := newManager.Read(desc, resource.EffectiveResource{
			// The values using in here does not really matter.
			PreferredTags: []string{"en"},
			DefaultTag:    "en",
		})
		if err != nil {
			return fmt.Errorf("invalid resource: %w", err)
		}
	}

	// Validate configuration.
	cfg, err := configsource.LoadConfig(newManager)
	if err != nil {
		return err
	}

	if string(cfg.AppConfig.ID) != appID {
		return fmt.Errorf("invalid resource '%s': incorrect app ID", configsource.AuthgearYAML)
	}

	return nil
}

func applyUpdates(manager *resource.Manager, appFs resource.Fs, secretKeyAllowlist []string, updates []Update) (*resource.Manager, error) {
	newFs, err := cloneFS(appFs)
	if err != nil {
		return nil, err
	}

	for _, u := range updates {
		// Retrieve the original file.
		resrc, err := func() (*resource.ResourceFile, error) {
			f, err := newFs.Open(u.Path)
			if os.IsNotExist(err) {
				return &resource.ResourceFile{
					Location: resource.Location{
						Fs:   resource.AferoFs{Fs: newFs},
						Path: u.Path,
					},
					Data: nil,
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
				Location: resource.Location{
					Fs:   resource.AferoFs{Fs: newFs},
					Path: u.Path,
				},
				Data: data,
			}, nil
		}()
		if err != nil {
			return nil, err
		}

		desc, ok := manager.Resolve(u.Path)
		if !ok {
			err = fmt.Errorf("invalid resource '%s': unknown resource path", resrc.Location.Path)
			return nil, err
		}

		resrc, err = desc.UpdateResource(resrc, u.Data, resource.AppFile{
			Path:              u.Path,
			AllowedSecretKeys: secretKeyAllowlist,
		})
		if err != nil {
			return nil, err
		}

		if resrc.Data == nil {
			_ = newFs.Remove(resrc.Location.Path)
		} else {
			_ = newFs.MkdirAll(path.Dir(resrc.Location.Path), 0666)
			_ = afero.WriteFile(newFs, resrc.Location.Path, resrc.Data, 0666)
		}
	}

	newAppFs := resource.AferoFs{Fs: newFs}
	var newResFs []resource.Fs
	for _, fs := range manager.Fs {
		if fs == appFs {
			newResFs = append(newResFs, newAppFs)
		} else {
			newResFs = append(newResFs, fs)
		}
	}
	return resource.NewManager(manager.Registry, newResFs), nil
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
