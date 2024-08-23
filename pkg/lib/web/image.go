package web

import (
	"context"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"os"
	stdlibpath "path"
	"regexp"
	"sort"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/util/intlresource"
	"github.com/authgear/authgear-server/pkg/util/libmagic"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

var imageResolveFsLevelPriority = []resource.FsLevel{
	resource.FsLevelApp, resource.FsLevelCustom, resource.FsLevelBuiltin,
}

var preferredExtensions = map[string]string{
	"image/png":  ".png",
	"image/jpeg": ".jpeg",
	"image/gif":  ".gif",
}

const defaultSizeLimit = 100 * 1024

var nonLocaleAwareImageRegex = regexp.MustCompile(`^static/(.+)\.(png|jpe|jpeg|jpg|gif)$`)

type NonLocaleAwareImageDescriptor struct {
	Name      string
	SizeLimit int
}

var _ resource.Descriptor = NonLocaleAwareImageDescriptor{}
var _ resource.SizeLimitDescriptor = NonLocaleAwareImageDescriptor{}

func (a NonLocaleAwareImageDescriptor) MatchResource(path string) (*resource.Match, bool) {
	matches := nonLocaleAwareImageRegex.FindStringSubmatch(path)
	if len(matches) != 3 {
		return nil, false
	}
	name := matches[1]

	if name != a.Name {
		return nil, false
	}
	return &resource.Match{}, true
}

func (a NonLocaleAwareImageDescriptor) FindResources(fs resource.Fs) ([]resource.Location, error) {
	staticDir, err := fs.Open(AppAssetsURLDirname)
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	defer staticDir.Close()

	var locations []resource.Location

	for mediaType := range preferredExtensions {
		exts, _ := mime.ExtensionsByType(mediaType)
		for _, ext := range exts {
			p := stdlibpath.Join(AppAssetsURLDirname, a.Name+ext)
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

func (a NonLocaleAwareImageDescriptor) ViewResources(resources []resource.ResourceFile, rawView resource.View) (interface{}, error) {
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

func (a NonLocaleAwareImageDescriptor) viewAppFile(resources []resource.ResourceFile, view resource.AppFileView) (interface{}, error) {
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

func (a NonLocaleAwareImageDescriptor) viewEffectiveFile(resources []resource.ResourceFile, view resource.EffectiveFileView) (interface{}, error) {
	path := view.EffectiveFilePath()
	asset, err := a.viewByPath(resources, path)
	if err != nil {
		return nil, err
	}
	return asset.Data, nil
}

func (a NonLocaleAwareImageDescriptor) viewEffectiveResource(resources []resource.ResourceFile, view resource.EffectiveResourceView) (interface{}, error) {
	for _, fsLevel := range imageResolveFsLevelPriority {
		for _, resrc := range resources {
			if resrc.Location.Fs.GetFsLevel() == fsLevel {
				mimeType := http.DetectContentType(resrc.Data)
				ext, ok := preferredExtensions[mimeType]
				if !ok {
					return nil, fmt.Errorf("invalid image format: %s", mimeType)
				}
				path := stdlibpath.Join(AppAssetsURLDirname, a.Name+ext)
				return &StaticAsset{
					Path: path,
					Data: resrc.Data,
				}, nil
			}
		}
	}
	return nil, resource.ErrResourceNotFound
}

func (a NonLocaleAwareImageDescriptor) viewValidateResource(resources []resource.ResourceFile, view resource.ValidateResourceView) (interface{}, error) {
	// Ensure there is at most one resource
	// For each Fs, remember how many paths we have seen.
	seen := make(map[resource.Fs][]string)
	for _, resrc := range resources {
		paths, ok := seen[resrc.Location.Fs]
		if !ok {
			paths = []string{}
			seen[resrc.Location.Fs] = paths
		}
		paths = append(paths, resrc.Location.Path)
		seen[resrc.Location.Fs] = paths
	}
	for _, paths := range seen {
		if len(paths) > 1 {
			sort.Strings(paths)
			return nil, fmt.Errorf("duplicate resource: %v", paths)
		}
	}

	return nil, nil
}

func (a NonLocaleAwareImageDescriptor) UpdateResource(_ context.Context, _ []resource.ResourceFile, resrc *resource.ResourceFile, data []byte) (*resource.ResourceFile, error) {
	if len(data) > 0 {
		typ := libmagic.MimeFromBytes(data)
		_, ok := preferredExtensions[typ]
		if !ok {
			return nil, UnsupportedImageFile.NewWithDetails("unsupported image file", apierrors.Details{
				"type": apierrors.APIErrorDetail.Value(typ),
			})
		}
	}

	return &resource.ResourceFile{
		Location: resrc.Location,
		Data:     data,
	}, nil
}

func (a NonLocaleAwareImageDescriptor) viewByPath(resources []resource.ResourceFile, path string) (*StaticAsset, error) {
	matches := nonLocaleAwareImageRegex.FindStringSubmatch(path)
	if len(matches) < 3 {
		return nil, resource.ErrResourceNotFound
	}
	requestedExtension := matches[2]

	var found bool
	var bytes []byte
	for _, resrc := range resources {
		m := nonLocaleAwareImageRegex.FindStringSubmatch(resrc.Location.Path)
		extension := m[2]
		if extension == requestedExtension {
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

	p := stdlibpath.Join(AppAssetsURLDirname, a.Name+ext)
	return &StaticAsset{
		Path: p,
		Data: bytes,
	}, nil
}

func (a NonLocaleAwareImageDescriptor) GetSizeLimit() int {
	if a.SizeLimit == 0 {
		return defaultSizeLimit
	}
	return a.SizeLimit
}

type languageImage struct {
	LanguageTag     string
	RealLanguageTag string
	Data            []byte
}

func (i languageImage) GetLanguageTag() string {
	return i.LanguageTag
}

var localeAwareImageRegex = regexp.MustCompile(`^static/([a-zA-Z0-9-]+)/(.+)\.(png|jpe|jpeg|jpg|gif)$`)

type LocaleAwareImageDescriptor struct {
	Name      string
	SizeLimit int
}

var _ resource.Descriptor = LocaleAwareImageDescriptor{}
var _ resource.SizeLimitDescriptor = LocaleAwareImageDescriptor{}

func (a LocaleAwareImageDescriptor) MatchResource(path string) (*resource.Match, bool) {
	matches := localeAwareImageRegex.FindStringSubmatch(path)
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

func (a LocaleAwareImageDescriptor) FindResources(fs resource.Fs) ([]resource.Location, error) {
	staticDir, err := fs.Open(AppAssetsURLDirname)
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
		stat, err := fs.Stat(stdlibpath.Join(AppAssetsURLDirname, langTag))
		if err != nil {
			return nil, err
		}
		if !stat.IsDir() {
			continue
		}

		for mediaType := range preferredExtensions {
			exts, _ := mime.ExtensionsByType(mediaType)
			for _, ext := range exts {
				p := stdlibpath.Join(AppAssetsURLDirname, langTag, a.Name+ext)
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

func (a LocaleAwareImageDescriptor) ViewResources(resources []resource.ResourceFile, rawView resource.View) (interface{}, error) {
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

func (a LocaleAwareImageDescriptor) UpdateResource(_ context.Context, _ []resource.ResourceFile, resrc *resource.ResourceFile, data []byte) (*resource.ResourceFile, error) {
	if len(data) > 0 {
		typ := libmagic.MimeFromBytes(data)
		_, ok := preferredExtensions[typ]
		if !ok {
			return nil, UnsupportedImageFile.NewWithDetails("unsupported image file", apierrors.Details{
				"type": apierrors.APIErrorDetail.Value(typ),
			})
		}
	}

	return &resource.ResourceFile{
		Location: resrc.Location,
		Data:     data,
	}, nil
}

func (a LocaleAwareImageDescriptor) viewValidateResource(resources []resource.ResourceFile, view resource.ValidateResourceView) (interface{}, error) {
	// Ensure there is at most one resource
	// For each Fs and for each locale, remember how many paths we have seen.
	seen := make(map[resource.Fs]map[string][]string)
	for _, resrc := range resources {
		languageTag := localeAwareImageRegex.FindStringSubmatch(resrc.Location.Path)[1]
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

func (a LocaleAwareImageDescriptor) viewEffectiveResource(resources []resource.ResourceFile, view resource.EffectiveResourceView) (interface{}, error) {
	preferredLanguageTags := view.PreferredLanguageTags()
	defaultLanguageTag := view.DefaultLanguageTag()

	var fallbackImage *languageImage
	images := make(map[resource.FsLevel]map[string]intlresource.LanguageItem)
	extractLanguageTag := func(resrc resource.ResourceFile) string {
		langTag := localeAwareImageRegex.FindStringSubmatch(resrc.Location.Path)[1]
		return langTag
	}
	add := func(langTag string, resrc resource.ResourceFile) error {
		fsLevel := resrc.Location.Fs.GetFsLevel()
		i := languageImage{
			LanguageTag:     langTag,
			RealLanguageTag: extractLanguageTag(resrc),
			Data:            resrc.Data,
		}
		if images[fsLevel] == nil {
			images[fsLevel] = make(map[string]intlresource.LanguageItem)
		}
		images[fsLevel][langTag] = i
		if fallbackImage == nil {
			fallbackImage = &i
		}
		return nil
	}

	err := intlresource.Prepare(resources, view, extractLanguageTag, add)
	if err != nil {
		return nil, err
	}

	var matched intlresource.LanguageItem
	for _, fsLevel := range imageResolveFsLevelPriority {
		var items []intlresource.LanguageItem
		imagesInFsLevel, ok := images[fsLevel]
		if !ok {
			continue
		}

		for _, i := range imagesInFsLevel {
			items = append(items, i)
		}

		matched, err = intlresource.Match(preferredLanguageTags, defaultLanguageTag, items)
		if err == nil {
			break
		} else if errors.Is(err, intlresource.ErrNoLanguageMatch) {
			continue
		} else {
			return nil, err
		}
	}

	if matched == nil {
		if fallbackImage != nil {
			// Use first item in case of no match, to ensure resolution always succeed
			matched = *fallbackImage
		} else {
			// If no configured translation, fail the resolution process
			return nil, resource.ErrResourceNotFound
		}
	}

	tagger := matched.(languageImage)

	mimeType := http.DetectContentType(tagger.Data)
	ext, ok := preferredExtensions[mimeType]
	if !ok {
		return nil, fmt.Errorf("invalid image format: %s", mimeType)
	}

	path := stdlibpath.Join(AppAssetsURLDirname, tagger.RealLanguageTag, a.Name+ext)
	return &StaticAsset{
		Path: path,
		Data: tagger.Data,
	}, nil
}

func (a LocaleAwareImageDescriptor) viewAppFile(resources []resource.ResourceFile, view resource.AppFileView) (interface{}, error) {
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

func (a LocaleAwareImageDescriptor) viewEffectiveFile(resources []resource.ResourceFile, view resource.EffectiveFileView) (interface{}, error) {
	path := view.EffectiveFilePath()
	asset, err := a.viewByPath(resources, path)
	if err != nil {
		return nil, err
	}
	return asset.Data, nil
}

func (a LocaleAwareImageDescriptor) viewByPath(resources []resource.ResourceFile, path string) (*StaticAsset, error) {
	matches := localeAwareImageRegex.FindStringSubmatch(path)
	if len(matches) < 4 {
		return nil, resource.ErrResourceNotFound
	}
	requestedLangTag := matches[1]
	requestedExtension := matches[3]

	var found bool
	var bytes []byte
	for _, resrc := range resources {
		m := localeAwareImageRegex.FindStringSubmatch(resrc.Location.Path)
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

	p := stdlibpath.Join(AppAssetsURLDirname, requestedLangTag, a.Name+ext)
	return &StaticAsset{
		Path: p,
		Data: bytes,
	}, nil
}

func (a LocaleAwareImageDescriptor) GetSizeLimit() int {
	if a.SizeLimit == 0 {
		return defaultSizeLimit
	}
	return a.SizeLimit
}

var staticImageRegex = regexp.MustCompile(`^static/(.+)\.(png|jpe|jpeg|jpg|gif)$`)

type StaticImageDescriptor struct {
	Name string
}

var _ resource.Descriptor = StaticImageDescriptor{}

func (a StaticImageDescriptor) MatchResource(path string) (*resource.Match, bool) {
	matches := staticImageRegex.FindStringSubmatch(path)
	if len(matches) != 3 {
		return nil, false
	}
	name := matches[1]

	if name != a.Name {
		return nil, false
	}
	return &resource.Match{}, true
}

func (a StaticImageDescriptor) FindResources(fs resource.Fs) ([]resource.Location, error) {
	if fs.GetFsLevel() != resource.FsLevelBuiltin {
		return []resource.Location{}, nil
	}

	staticDir, err := fs.Open(AppAssetsURLDirname)
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	defer staticDir.Close()

	var locations []resource.Location

	for mediaType := range preferredExtensions {
		exts, _ := mime.ExtensionsByType(mediaType)
		for _, ext := range exts {
			p := stdlibpath.Join(AppAssetsURLDirname, a.Name+ext)
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

func (a StaticImageDescriptor) ViewResources(resources []resource.ResourceFile, rawView resource.View) (interface{}, error) {
	switch view := rawView.(type) {
	case resource.AppFileView:
		return nil, nil
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

func (a StaticImageDescriptor) viewEffectiveFile(resources []resource.ResourceFile, view resource.EffectiveFileView) (interface{}, error) {
	path := view.EffectiveFilePath()
	asset, err := a.viewByPath(resources, path)
	if err != nil {
		return nil, err
	}
	return asset.Data, nil
}

func (a StaticImageDescriptor) viewEffectiveResource(resources []resource.ResourceFile, view resource.EffectiveResourceView) (interface{}, error) {
	for _, resrc := range resources {
		if resrc.Location.Fs.GetFsLevel() == resource.FsLevelBuiltin {
			mimeType := http.DetectContentType(resrc.Data)
			ext, ok := preferredExtensions[mimeType]
			if !ok {
				return nil, fmt.Errorf("invalid image format: %s", mimeType)
			}
			path := stdlibpath.Join(AppAssetsURLDirname, a.Name+ext)
			return &StaticAsset{
				Path: path,
				Data: resrc.Data,
			}, nil
		}
	}
	return nil, resource.ErrResourceNotFound
}

func (a StaticImageDescriptor) viewValidateResource(resources []resource.ResourceFile, view resource.ValidateResourceView) (interface{}, error) {
	return nil, nil
}

func (a StaticImageDescriptor) UpdateResource(_ context.Context, _ []resource.ResourceFile, resrc *resource.ResourceFile, data []byte) (*resource.ResourceFile, error) {
	return nil, fmt.Errorf("Static image resource cannot be updated. Use locale aware image or non locale aware image resource instead.")
}

func (a StaticImageDescriptor) viewByPath(resources []resource.ResourceFile, path string) (*StaticAsset, error) {
	matches := staticImageRegex.FindStringSubmatch(path)
	if len(matches) < 3 {
		return nil, resource.ErrResourceNotFound
	}
	requestedExtension := matches[2]

	var found bool
	var bytes []byte
	for _, resrc := range resources {
		m := staticImageRegex.FindStringSubmatch(resrc.Location.Path)
		extension := m[2]
		if extension == requestedExtension {
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

	p := stdlibpath.Join(AppAssetsURLDirname, a.Name+ext)
	return &StaticAsset{
		Path: p,
		Data: bytes,
	}, nil
}
