//go:build !authgeardev
// +build !authgeardev

package web

import (
	"github.com/authgear/authgear-server"
)

func NewDefaultGlobalEmbeddedResourceManager() (*GlobalEmbeddedResourceManager, error) {
	impl, err := NewGlobalEmbeddedResourceManagerEmbed(NewGlobalEmbeddedResourceManagerEmbedOptions{
		EmbedFS:                               runtimeresource.EmbedFS_resources_authgear,
		EmbedFSRoot:                           defaultResourceDir,
		ManifestFilenameRelativeToEmbedFSRoot: defaultManifestName,
	})
	if err != nil {
		return nil, err
	}

	return &GlobalEmbeddedResourceManager{
		Impl: impl,
	}, nil
}
