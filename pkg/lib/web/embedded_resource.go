package web

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path"
	"sync/atomic"

	"gopkg.in/fsnotify.v1"

	"github.com/authgear/authgear-server/pkg/util/resource"
)

const defaultResourceDir = "resources/authgear/generated"
const defaultManifestName = "manifest.json"

type globalEmbeddedResourceManagerManifest struct {
	ResourceDir string
	Name        string
	content     atomic.Value
}

type GlobalEmbeddedResourceManager struct {
	Manifest *globalEmbeddedResourceManagerManifest
	watcher  *fsnotify.Watcher
}

type globalEmbeddedResourceManagerManifestContext struct {
	Content map[string]string
}

func NewDefaultGlobalEmbeddedResourceManager() (*GlobalEmbeddedResourceManager, error) {
	return NewGlobalEmbeddedResourceManager(&globalEmbeddedResourceManagerManifest{
		ResourceDir: defaultResourceDir,
		Name:        defaultManifestName,
	})
}

func NewGlobalEmbeddedResourceManager(manifest *globalEmbeddedResourceManagerManifest) (*GlobalEmbeddedResourceManager, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	m := &GlobalEmbeddedResourceManager{
		Manifest: &globalEmbeddedResourceManagerManifest{
			ResourceDir: manifest.ResourceDir,
			Name:        manifest.Name,
		},
		watcher: watcher,
	}

	err = m.setupWatch(nil)
	if err != nil {
		return nil, err
	}

	err = m.reload()
	if err != nil {
		return nil, err
	}

	go m.watch()

	return m, nil
}

func (m *GlobalEmbeddedResourceManager) loadManifest() (map[string]string, error) {
	jsonFile, err := os.Open(m.manifestFilePath())
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	byteValue, _ := io.ReadAll(jsonFile)

	var result map[string]string
	_ = json.Unmarshal([]byte(byteValue), &result)

	return result, nil
}

func (m *GlobalEmbeddedResourceManager) setupWatch(event *fsnotify.Event) (err error) {
	if event == nil {
		err = m.watcher.Add(m.manifestFilePath())
		if os.IsNotExist(err) {
			err = m.watcher.Add(m.Manifest.ResourceDir)
		}
		return
	}

	switch event.Op {
	case fsnotify.Create, fsnotify.Write:
		_ = m.watcher.Remove(m.Manifest.ResourceDir)
		err = m.watcher.Add(m.manifestFilePath())
	case fsnotify.Remove:
		err = m.watcher.Add(m.Manifest.ResourceDir)
	}

	return
}

func (m *GlobalEmbeddedResourceManager) watch() {
	for {
		select {
		case event, ok := <-m.watcher.Events:
			if !ok {
				return
			}

			if event.Name != m.manifestFilePath() {
				break
			}

			_ = m.setupWatch(&event)
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

	manifestCtx := &globalEmbeddedResourceManagerManifestContext{
		Content: newManifest,
	}
	m.Manifest.content.Store(manifestCtx)
	return nil
}

func (m *GlobalEmbeddedResourceManager) manifestFilePath() string {
	return path.Join(m.Manifest.ResourceDir, m.Manifest.Name)
}

func (m *GlobalEmbeddedResourceManager) getManifestContext() *globalEmbeddedResourceManagerManifestContext {
	return m.Manifest.content.Load().(*globalEmbeddedResourceManagerManifestContext)
}

func (m *GlobalEmbeddedResourceManager) close() error {
	return m.watcher.Close()
}

func (m *GlobalEmbeddedResourceManager) AssetName(key string) (name string, err error) {
	manifest := m.getManifestContext().Content
	if val, ok := manifest[key]; ok {
		return val, nil
	}
	return "", resource.ErrResourceNotFound
}

func (m *GlobalEmbeddedResourceManager) Open(name string) (http.File, error) {
	fs := http.Dir(m.Manifest.ResourceDir)
	return fs.Open(name)
}
