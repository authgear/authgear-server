package web

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"sync/atomic"

	"gopkg.in/fsnotify.v1"

	"github.com/authgear/authgear-server/pkg/util/resource"
)

const GeneratedAssetPath = "resources/authgear/static/generated"
const GeneratedAssetManifest = "resources/authgear/static/generated/manifest.json"

type GeneratedAssetDescriptor struct {
	manifest atomic.Value
	watcher  *fsnotify.Watcher
}

type ManifestContext struct {
	content map[string]string
}

var _ resource.Descriptor = &GeneratedAssetDescriptor{}

func NewGeneratedAssetDescriptor() *GeneratedAssetDescriptor {
	watcher, _ := fsnotify.NewWatcher()

	var manifest atomic.Value
	manifestContent, _ := LoadManifest()

	manifest.Store(&ManifestContext{
		content: manifestContent,
	})

	s := &GeneratedAssetDescriptor{
		manifest: manifest,
		watcher:  watcher,
	}

	go s.watch()

	_ = watcher.Add(GeneratedAssetPath)

	return s
}

func LoadManifest() (map[string]string, error) {
	jsonFile, err := os.Open(GeneratedAssetManifest)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var result map[string]string
	_ = json.Unmarshal([]byte(byteValue), &result)

	return result, nil
}

func (d *GeneratedAssetDescriptor) watch() {
	for {
		select {
		case event, ok := <-d.watcher.Events:
			if !ok {
				return
			}

			if event.Name != GeneratedAssetManifest {
				break
			}

			if event.Op&fsnotify.Write != fsnotify.Write && event.Op&fsnotify.Create != fsnotify.Create {
				break
			}

			_ = d.reload()

		case _, ok := <-d.watcher.Errors:
			if !ok {
				return
			}
		}
	}
}

func (d *GeneratedAssetDescriptor) reload() error {
	newManifest, err := LoadManifest()
	if err != nil {
		return err
	}

	manifestCtx := &ManifestContext{
		content: newManifest,
	}
	d.manifest.Store(manifestCtx)
	return nil
}

func (d *GeneratedAssetDescriptor) GetAssetPathForKey(key string) (string, error) {
	manifest := d.manifest.Load().(*ManifestContext).content
	if val, ok := manifest[key]; ok {
		return val, nil
	}
	return "", resource.ErrResourceNotFound
}

func (d *GeneratedAssetDescriptor) MatchResource(resourcePath string) (*resource.Match, bool) {
	manifest := d.manifest.Load().(*ManifestContext).content

	key := strings.TrimPrefix(resourcePath, GeneratedStaticAssetResourcePrefix)
	if IsSourceMapPath(key) {
		key = strings.TrimSuffix(key, ".map")
	}

	if _, ok := manifest[key]; ok {
		return &resource.Match{}, true
	}
	return nil, false
}

func (d *GeneratedAssetDescriptor) FindResources(fs resource.Fs) ([]resource.Location, error) {
	manifest := d.manifest.Load().(*ManifestContext).content
	locations := make([]resource.Location, 0)

	for _, value := range manifest {
		location := resource.Location{
			Fs:   fs,
			Path: path.Join(GeneratedStaticAssetResourcePrefix, value),
		}
		_, err := resource.ReadLocation(location)

		if os.IsNotExist(err) || err != nil {
			continue
		}
		locations = append(locations, location)

		mapValue := fmt.Sprintf("%s.map", value)
		locationMap := resource.Location{
			Fs:   fs,
			Path: path.Join(GeneratedStaticAssetResourcePrefix, mapValue),
		}
		_, err = resource.ReadLocation(locationMap)

		if os.IsNotExist(err) || err != nil {
			continue
		}
		locations = append(locations, locationMap)
	}
	return locations, nil
}

func (d *GeneratedAssetDescriptor) ViewResources(resources []resource.ResourceFile, rawView resource.View) (interface{}, error) {
	switch rawView.(type) {
	case resource.AppFileView:
		return d.viewResources(resources, rawView)
	case resource.EffectiveFileView:
		return d.viewResources(resources, rawView)
	case resource.EffectiveResourceView:
		return d.viewResources(resources, rawView)
	case resource.ValidateResourceView:
		return d.viewResources(resources, rawView)
	default:
		return nil, fmt.Errorf("unsupported view: %T", rawView)
	}
}

func (d *GeneratedAssetDescriptor) viewResources(resources []resource.ResourceFile, rawView resource.View) ([]byte, error) {
	switch rawView.(type) {
	case resource.AppFileView:
		break
	case resource.EffectiveFileView:
		break
	case resource.EffectiveResourceView:
		return nil, resource.ErrResourceNotFound
	case resource.ValidateResourceView:
		return nil, resource.ErrResourceNotFound
	default:
		return nil, fmt.Errorf("unsupported view: %T", rawView)
	}

	if len(resources) == 0 {
		return nil, resource.ErrResourceNotFound
	}

	manifest := d.manifest.Load().(*ManifestContext).content
	viewPath := strings.TrimPrefix(rawView.(resource.EffectiveFile).Path, GeneratedStaticAssetResourcePrefix)
	key := viewPath

	if IsSourceMapPath(viewPath) {
		key = strings.TrimSuffix(key, ".map")
	}
	assetPath := path.Join(GeneratedStaticAssetResourcePrefix, manifest[key])

	ae := path.Ext(assetPath)
	ve := path.Ext(viewPath)
	if ae != ve {
		assetPath += ve
	}

	for _, r := range resources {
		if r.Location.Path == assetPath {
			return r.Data, nil
		}
	}

	return nil, resource.ErrResourceNotFound
}

func (d *GeneratedAssetDescriptor) UpdateResource(_ context.Context, _ []resource.ResourceFile, resrc *resource.ResourceFile, data []byte) (*resource.ResourceFile, error) {
	return nil, fmt.Errorf("unsupported resource update")
}
