//go:build authgeardev
// +build authgeardev

package web

func NewDefaultGlobalEmbeddedResourceManager() (*GlobalEmbeddedResourceManager, error) {
	impl, err := NewGlobalEmbeddedResourceManagerWorkdir(&globalEmbeddedResourceManagerManifest{
		ResourceDir: defaultResourceDir,
		Name:        defaultManifestName,
	})
	if err != nil {
		return nil, err
	}

	return &GlobalEmbeddedResourceManager{
		Impl: impl,
	}, nil
}
