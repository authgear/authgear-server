package web

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"sync/atomic"

	"gopkg.in/fsnotify.v1"

	"github.com/authgear/authgear-server/pkg/util/resource"
)

const DefaultResourceDir = "resources/authgear/"
const DefaultResourcePrefix = "static/generated/"
const DefaultManifestName = "manifest.json"

type Manifest struct {
	ResourceDir    string
	ResourcePrefix string
	Name           string
	content        atomic.Value
}

type GlobalEmbeddedResourceManager struct {
	Manifest *Manifest
	watcher  *fsnotify.Watcher
}

type ManifestContext struct {
	Content map[string]string
}

func NewDefaultGlobalEmbeddedResourceManager() (*GlobalEmbeddedResourceManager, error) {
	return NewGlobalEmbeddedResourceManager(&Manifest{
		ResourceDir:    DefaultResourceDir,
		ResourcePrefix: DefaultResourcePrefix,
		Name:           DefaultManifestName,
	})
}

func NewGlobalEmbeddedResourceManager(manifest *Manifest) (*GlobalEmbeddedResourceManager, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	err = watcher.Add(manifest.ResourceDir + manifest.ResourcePrefix)
	if err != nil {
		return nil, err
	}

	m := &GlobalEmbeddedResourceManager{
		Manifest: &Manifest{
			ResourceDir:    manifest.ResourceDir,
			ResourcePrefix: manifest.ResourcePrefix,
			Name:           manifest.Name,
		},
		watcher: watcher,
	}

	err = m.reload()
	if err != nil {
		return nil, err
	}

	go m.watch()

	return m, nil
}

func (m *GlobalEmbeddedResourceManager) loadManifest() (map[string]string, error) {
	jsonFile, err := os.Open(m.Manifest.ResourceDir + m.Manifest.ResourcePrefix + m.Manifest.Name)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	byteValue, _ := io.ReadAll(jsonFile)

	var result map[string]string
	_ = json.Unmarshal([]byte(byteValue), &result)

	return result, nil
}

func (m *GlobalEmbeddedResourceManager) watch() {
	for {
		select {
		case event, ok := <-m.watcher.Events:
			if !ok {
				return
			}

			if event.Name != m.Manifest.ResourceDir+m.Manifest.ResourcePrefix+m.Manifest.Name {
				break
			}

			if event.Op&fsnotify.Write != fsnotify.Write && event.Op&fsnotify.Create != fsnotify.Create {
				break
			}

			_ = m.reload()

		case _, ok := <-m.watcher.Errors:
			if !ok {
				return
			}
		}
	}
}

func (m *GlobalEmbeddedResourceManager) reload() error {
	newManifest, err := m.loadManifest()
	if err != nil {
		return err
	}

	manifestCtx := &ManifestContext{
		Content: newManifest,
	}
	m.Manifest.content.Store(manifestCtx)
	return nil
}

func (m *GlobalEmbeddedResourceManager) GetManifestContext() *ManifestContext {
	return m.Manifest.content.Load().(*ManifestContext)
}

func (m *GlobalEmbeddedResourceManager) Close() error {
	return m.watcher.Close()
}

func (m *GlobalEmbeddedResourceManager) HTTPFileSystem() http.FileSystem {
	return http.Dir(m.Manifest.ResourceDir + m.Manifest.ResourcePrefix)
}

func (m *GlobalEmbeddedResourceManager) AssetPath(key string) (prefix string, name string, err error) {
	manifest := m.GetManifestContext().Content
	if val, ok := manifest[key]; ok {
		return m.Manifest.ResourcePrefix, val, nil
	}
	return "", "", resource.ErrResourceNotFound
}

func (m *GlobalEmbeddedResourceManager) Resolve(resourcePath string) (string, bool) {
	manifest := m.GetManifestContext().Content

	key := strings.TrimPrefix(resourcePath, m.Manifest.ResourcePrefix)
	if IsSourceMapPath(key) {
		key = strings.TrimSuffix(key, ".map")
	}

	if assetFileName, ok := manifest[key]; ok {
		// Add source map extension to the file name if resourcePath is a source map path
		ae := path.Ext(assetFileName)
		ve := path.Ext(resourcePath)
		if ae != ve {
			assetFileName += ve
		}
		return assetFileName, true
	}
	return "", false
}

func (m *GlobalEmbeddedResourceManager) Open(assetPath string) (http.File, error) {
	fs := m.HTTPFileSystem()
	return fs.Open(assetPath)
}
