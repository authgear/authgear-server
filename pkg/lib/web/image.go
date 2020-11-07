package web

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"regexp"

	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/template"
)

type languageImage struct {
	languageTag string
	data        []byte
}

func (i languageImage) GetLanguageTag() string {
	return i.languageTag
}

var imageExtensions = map[string]string{
	"image/png":  ".png",
	"image/jpeg": ".jpeg",
	"image/gif":  ".gif",
}

var imageRegex = regexp.MustCompile(`^static/([a-zA-Z0-9-]+)/(.+)\.(png|jpeg|gif)$`)

type ImageDescriptor struct {
	Name string
}

var _ resource.Descriptor = ImageDescriptor{}

func (a ImageDescriptor) MatchResource(path string) bool {
	matches := imageRegex.FindStringSubmatch(path)
	if len(matches) != 4 {
		return false
	}
	return matches[2] == a.Name
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

		for _, ext := range imageExtensions {
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

	return locations, nil
}

func (a ImageDescriptor) ViewResources(resources []resource.ResourceFile, rawView resource.View) (interface{}, error) {
	switch view := rawView.(type) {
	case resource.EffectiveResourceView:
		return a.viewEffectiveResource(resources, view)
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

func (a ImageDescriptor) viewEffectiveResource(resources []resource.ResourceFile, view resource.EffectiveResourceView) (interface{}, error) {
	preferredLanguageTags := view.PreferredLanguageTags()
	defaultLanguageTag := view.DefaultLanguageTag()

	images := make(map[string]template.LanguageItem)
	for _, resrc := range resources {
		languageTag := imageRegex.FindStringSubmatch(resrc.Location.Path)[1]
		images[languageTag] = languageImage{
			languageTag: languageTag,
			data:        resrc.Data,
		}
	}

	var items []template.LanguageItem
	for _, i := range images {
		items = append(items, i)
	}

	matched, err := template.MatchLanguage(preferredLanguageTags, defaultLanguageTag, items)
	if errors.Is(err, template.ErrNoLanguageMatch) {
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
	ext, ok := imageExtensions[mimeType]
	if !ok {
		return nil, fmt.Errorf("invalid image format: %s", mimeType)
	}

	path := fmt.Sprintf("%s%s/%s%s", StaticAssetResourcePrefix, resolvedLanguageTag, a.Name, ext)
	return &StaticAsset{
		Path: path,
		Data: tagger.data,
	}, nil
}
