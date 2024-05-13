package appresource

import (
	"io/ioutil"
	"path"

	"github.com/spf13/afero"

	"github.com/authgear/authgear-server/pkg/util/resource"
)

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
