package web

import (
	"errors"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path"
	"regexp"
	"sort"

	"github.com/authgear/authgear-server/pkg/util/intlresource"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

type languageImage struct {
	languageTag string
	data        []byte
}

func (i languageImage) GetLanguageTag() string {
	return i.languageTag
}

var preferredExtensions = map[string]string{
	"image/png":  ".png",
	"image/jpeg": ".jpeg",
	"image/gif":  ".gif",
}

var imageRegex = regexp.MustCompile(`^static/([a-zA-Z0-9-]+)/(.+)\.(png|jpe|jpeg|jpg|gif)$`)

type ImageDescriptor struct {
	Name string
}

var _ resource.Descriptor = ImageDescriptor{}

func (a ImageDescriptor) MatchResource(path string) (*resource.Match, bool) {
	matches := imageRegex.FindStringSubmatch(path)
	if len(matches) != 4 {
		return nil, false
	}
	languageTag := matches[1]
	name := matches[2]

	if name != a.Name {
		return nil, false
	}
	return &resource.Match{LanguageTag: languageTag}, true
}

func (a ImageDescriptor) FindResources(fs resource.Fs) ([]resource.Location, error) {
	staticDir, err := fs.Open("static")
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	defer staticDir.Close()

	langTagDirs, err := staticDir.Readdirnames(0)
	if err != nil {
		return nil, err
	}

	var locations []resource.Location
	for _, langTag := range langTagDirs {
		stat, err := fs.Stat(path.Join("static", langTag))
		if err != nil {
			return nil, err
		}
		if !stat.IsDir() {
			continue
		}

		for mediaType := range preferredExtensions {
			exts, _ := mime.ExtensionsByType(mediaType)
			for _, ext := range exts {
				p := path.Join("static", langTag, a.Name+ext)
				location := resource.Location{
					Fs:   fs,
					Path: p,
				}
				_, err := resource.ReadLocation(location)
				if os.IsNotExist(err) {
					continue
				} else if err != nil {
					return nil, err
				}
				locations = append(locations, location)
			}
		}
	}

	return locations, nil
}

func (a ImageDescriptor) ViewResources(resources []resource.ResourceFile, rawView resource.View) (interface{}, error) {
	switch view := rawView.(type) {
	case resource.AppFileView:
		return a.viewAppFile(resources, view)
	case resource.EffectiveFileView:
		return a.viewEffectiveFile(resources, view)
	case resource.EffectiveResourceView:
		return a.viewEffectiveResource(resources, view)
	case resource.ValidateResourceView:
		return a.viewValidateResource(resources, view)
	default:
		return nil, fmt.Errorf("unsupported view: %T", rawView)
	}
}

func (a ImageDescriptor) UpdateResource(resrc *resource.ResourceFile, data []byte, view resource.View) (*resource.ResourceFile, error) {
	return &resource.ResourceFile{
		Location: resrc.Location,
		Data:     data,
	}, nil
}

func (a ImageDescriptor) viewValidateResource(resources []resource.ResourceFile, view resource.ValidateResourceView) (interface{}, error) {
	// Ensure there is at most one resource
	// For each Fs and for each locale, remember how many paths we have seen.
	seen := make(map[resource.Fs]map[string][]string)
	for _, resrc := range resources {
		languageTag := imageRegex.FindStringSubmatch(resrc.Location.Path)[1]
		m, ok := seen[resrc.Location.Fs]
		if !ok {
			m = make(map[string][]string)
			seen[resrc.Location.Fs] = m
		}
		paths := m[languageTag]
		paths = append(paths, resrc.Location.Path)
		m[languageTag] = paths
	}
	for _, m := range seen {
		for _, paths := range m {
			if len(paths) > 1 {
				sort.Strings(paths)
				return nil, fmt.Errorf("duplicate resource: %v", paths)
			}
		}
	}

	return nil, nil
}

func (a ImageDescriptor) viewEffectiveResource(resources []resource.ResourceFile, view resource.EffectiveResourceView) (interface{}, error) {
	preferredLanguageTags := view.PreferredLanguageTags()
	defaultLanguageTag := view.DefaultLanguageTag()

	images := make(map[string]intlresource.LanguageItem)
	add := func(langTag string, resrc resource.ResourceFile) error {
		images[langTag] = languageImage{
			languageTag: langTag,
			data:        resrc.Data,
		}
		return nil
	}
	extractLanguageTag := func(resrc resource.ResourceFile) string {
		langTag := imageRegex.FindStringSubmatch(resrc.Location.Path)[1]
		return langTag
	}

	err := intlresource.Prepare(resources, view, extractLanguageTag, add)
	if err != nil {
		return nil, err
	}

	var items []intlresource.LanguageItem
	for _, i := range images {
		items = append(items, i)
	}

	matched, err := intlresource.Match(preferredLanguageTags, defaultLanguageTag, items)
	if errors.Is(err, intlresource.ErrNoLanguageMatch) {
		if len(items) > 0 {
			// Use first item in case of no match, to ensure resolution always succeed
			matched = items[0]
		} else {
			// If no configured translation, fail the resolution process
			return nil, resource.ErrResourceNotFound
		}
	} else if err != nil {
		return nil, err
	}

	tagger := matched.(languageImage)
	resolvedLanguageTag := tagger.languageTag

	mimeType := http.DetectContentType(tagger.data)
	ext, ok := preferredExtensions[mimeType]
	if !ok {
		return nil, fmt.Errorf("invalid image format: %s", mimeType)
	}

	path := fmt.Sprintf("%s%s/%s%s", StaticAssetResourcePrefix, resolvedLanguageTag, a.Name, ext)
	return &StaticAsset{
		Path: path,
		Data: tagger.data,
	}, nil
}

func (a ImageDescriptor) viewAppFile(resources []resource.ResourceFile, view resource.AppFileView) (interface{}, error) {
	path := view.AppFilePath()
	var appResources []resource.ResourceFile
	for _, resrc := range resources {
		if resrc.Location.Fs.GetFsLevel() == resource.FsLevelApp {
			appResources = append(appResources, resrc)
		}
	}
	asset, err := a.viewByPath(appResources, path)
	if err != nil {
		return nil, err
	}
	return asset.Data, nil
}

func (a ImageDescriptor) viewEffectiveFile(resources []resource.ResourceFile, view resource.EffectiveFileView) (interface{}, error) {
	path := view.EffectiveFilePath()
	asset, err := a.viewByPath(resources, path)
	if err != nil {
		return nil, err
	}
	return asset.Data, nil
}

func (a ImageDescriptor) viewByPath(resources []resource.ResourceFile, path string) (*StaticAsset, error) {
	matches := imageRegex.FindStringSubmatch(path)
	if len(matches) < 4 {
		return nil, resource.ErrResourceNotFound
	}
	requestedLangTag := matches[1]
	requestedExtension := matches[3]

	var found bool
	var bytes []byte
	for _, resrc := range resources {
		m := imageRegex.FindStringSubmatch(resrc.Location.Path)
		langTag := m[1]
		extension := m[3]
		if langTag == requestedLangTag && extension == requestedExtension {
			found = true
			bytes = resrc.Data
		}
	}

	if !found {
		return nil, resource.ErrResourceNotFound
	}

	mimeType := http.DetectContentType(bytes)
	ext, ok := preferredExtensions[mimeType]
	if !ok {
		return nil, fmt.Errorf("invalid image format: %s", mimeType)
	}

	p := fmt.Sprintf("%s%s/%s%s", StaticAssetResourcePrefix, requestedLangTag, a.Name, ext)
	return &StaticAsset{
		Path: p,
		Data: bytes,
	}, nil
}
