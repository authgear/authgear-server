package web

import (
	"embed"
	"encoding/json"
	"io"
	"io/fs"
	"net/http"
	"path"

	"github.com/authgear/authgear-server/pkg/util/resource"
)

type GlobalEmbeddedResourceManagerEmbed struct {
	EmbedFS                               embed.FS
	EmbedFSRoot                           string
	ManifestFilenameRelativeToEmbedFSRoot string
	Manifest                              map[string]string
}

var _ GlobalEmbeddedResourceManagerImpl = (*GlobalEmbeddedResourceManagerEmbed)(nil)

type NewGlobalEmbeddedResourceManagerEmbedOptions struct {
	EmbedFS                               embed.FS
	EmbedFSRoot                           string
	ManifestFilenameRelativeToEmbedFSRoot string
}

func NewGlobalEmbeddedResourceManagerEmbed(opts NewGlobalEmbeddedResourceManagerEmbedOptions) (*GlobalEmbeddedResourceManagerEmbed, error) {
	p := path.Join(opts.EmbedFSRoot, opts.ManifestFilenameRelativeToEmbedFSRoot)

	f, err := opts.EmbedFS.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	byteValue, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var manifest map[string]string
	_ = json.Unmarshal([]byte(byteValue), &manifest)

	return &GlobalEmbeddedResourceManagerEmbed{
		EmbedFS:                               opts.EmbedFS,
		EmbedFSRoot:                           opts.EmbedFSRoot,
		ManifestFilenameRelativeToEmbedFSRoot: opts.ManifestFilenameRelativeToEmbedFSRoot,
		Manifest:                              manifest,
	}, nil
}

func (m *GlobalEmbeddedResourceManagerEmbed) AssetName(key string) (name string, err error) {
	if val, ok := m.Manifest[key]; ok {
		return val, nil
	}
	return "", resource.ErrResourceNotFound
}

func (m *GlobalEmbeddedResourceManagerEmbed) Open(name string) (http.File, error) {
	fsFileSystem, err := fs.Sub(m.EmbedFS, m.EmbedFSRoot)
	if err != nil {
		return nil, err
	}
	httpFileSystem := http.FS(fsFileSystem)
	return httpFileSystem.Open(name)
}
