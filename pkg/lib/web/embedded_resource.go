package web

import (
	"net/http"
)

const defaultResourceDir = "resources/authgear/generated"
const defaultManifestName = "manifest.json"

type GlobalEmbeddedResourceManagerImpl interface {
	AssetName(key string) (name string, err error)
	Open(name string) (http.File, error)
}

type GlobalEmbeddedResourceManager struct {
	Impl GlobalEmbeddedResourceManagerImpl
}

var _ GlobalEmbeddedResourceManagerImpl = (*GlobalEmbeddedResourceManager)(nil)

func (m *GlobalEmbeddedResourceManager) AssetName(key string) (name string, err error) {
	return m.Impl.AssetName(key)
}

func (m *GlobalEmbeddedResourceManager) Open(name string) (http.File, error) {
	return m.Impl.Open(name)
}
