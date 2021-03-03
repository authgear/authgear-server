package resource

import (
	"github.com/spf13/afero"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var ErrResourceNotFound = apierrors.NotFound.WithReason("ResourceNotFound").
	New("specified resource is not configured")

type Manager struct {
	Registry *Registry
	Fs       []Fs
}

func NewManager(registry *Registry, fs []Fs) *Manager {
	return &Manager{Registry: registry, Fs: fs}
}

func NewManagerWithDir(registry *Registry, builtinResourceDir string, customResourceDir string) *Manager {
	var fs []Fs
	fs = append(fs,
		LeveledAferoFs{
			Fs:      afero.NewBasePathFs(afero.OsFs{}, builtinResourceDir),
			FsLevel: FsLevelBuiltin,
		},
	)
	if customResourceDir != "" {
		fs = append(fs,
			LeveledAferoFs{
				Fs:      afero.NewBasePathFs(afero.OsFs{}, customResourceDir),
				FsLevel: FsLevelCustom,
			},
		)
	}
	return &Manager{
		Registry: registry.Clone(),
		Fs:       fs,
	}
}

func (m *Manager) Overlay(fs Fs) *Manager {
	newFs := make([]Fs, len(m.Fs)+1)
	copy(newFs, m.Fs)
	newFs[len(newFs)-1] = fs
	return NewManager(m.Registry, newFs)
}

func (m *Manager) Read(desc Descriptor, view View) (interface{}, error) {
	var locations []Location
	for _, fs := range m.Fs {
		ls, err := desc.FindResources(fs)
		if err != nil {
			return nil, err
		}
		locations = append(locations, ls...)
	}
	if len(locations) == 0 {
		return nil, ErrResourceNotFound
	}

	files := make([]ResourceFile, len(locations))
	for idx, location := range locations {
		data, err := ReadLocation(location)
		if err != nil {
			return nil, err
		}
		files[idx] = ResourceFile{
			Location: location,
			Data:     data,
		}
	}

	return desc.ViewResources(files, view)
}

func (m *Manager) Resolve(path string) (Descriptor, bool) {
	for _, desc := range m.Registry.Descriptors {
		if _, ok := desc.MatchResource(path); ok {
			return desc, true
		}
	}
	return nil, false
}

func (m *Manager) Filesystems() []Fs {
	return m.Fs
}
