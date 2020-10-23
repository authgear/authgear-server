package resource

import (
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

func (m *Manager) Overlay(fs Fs) *Manager {
	newFs := make([]Fs, len(m.Fs)+1)
	copy(newFs, m.Fs)
	newFs[len(newFs)-1] = fs
	return NewManager(m.Registry, newFs)
}

func (m *Manager) Read(desc Descriptor, args map[string]interface{}) (*MergedFile, error) {
	var layers []LayerFile
	for _, fs := range m.Fs {
		files, err := desc.ReadResource(fs)
		if err != nil {
			return nil, err
		}
		layers = append(layers, files...)
	}
	if len(layers) == 0 {
		return nil, ErrResourceNotFound
	}

	merged, err := desc.Merge(layers, args)
	if err != nil {
		return nil, err
	}

	return merged, nil
}

func (m *Manager) Resolve(path string) (Descriptor, bool) {
	for _, desc := range m.Registry.Descriptors {
		if ok := desc.MatchResource(path); ok {
			return desc, true
		}
	}
	return nil, false
}

func (m *Manager) Filesystems() []Fs {
	return m.Fs
}
