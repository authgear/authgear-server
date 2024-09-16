package template

import (
	"context"
	"errors"
	"fmt"
	htmltemplate "html/template"
	"os"
	"path"
	"regexp"
	texttemplate "text/template"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/util/intl"
	"github.com/authgear/authgear-server/pkg/util/intlresource"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

type Resource interface {
	templateResource()
}

func isTemplateUpdateAllowed(ctx context.Context) (bool, error) {
	fc, ok := ctx.Value(configsource.ContextKeyFeatureConfig).(*config.FeatureConfig)
	if !ok || fc == nil {
		return false, ErrMissingFeatureFlagInCtx
	}
	if fc.Messaging.TemplateCustomizationDisabled {
		return false, ErrUpdateDisallowed
	}
	return true, nil
}

// HTML defines a HTML template that is non-customizable
type HTML struct {
	// Name is the name of template
	Name string
	// ComponentDependencies is the HTML component templates this template depends on.
	ComponentDependencies []*HTML
}

// MessageHTML defines a HTML template that is customizable
type MessageHTML struct {
	// Name is the name of template
	Name string
}

var _ resource.Descriptor = &HTML{}
var _ resource.Descriptor = &MessageHTML{}

func (t *HTML) templateResource()        {}
func (t *MessageHTML) templateResource() {}

func (t *HTML) MatchResource(path string) (*resource.Match, bool) {
	return matchTemplatePath(path, t.Name)
}

func (t *MessageHTML) MatchResource(path string) (*resource.Match, bool) {
	return matchTemplatePath(path, t.Name)
}

func (t *HTML) FindResources(fs resource.Fs) ([]resource.Location, error) {
	// Exclude App Fs
	if fs.GetFsLevel() == resource.FsLevelApp {
		return []resource.Location{}, nil
	}
	return readTemplates(fs, t.Name)
}

func (t *MessageHTML) FindResources(fs resource.Fs) ([]resource.Location, error) {
	// Any Fs
	return readTemplates(fs, t.Name)
}

func (t *HTML) ViewResources(resources []resource.ResourceFile, view resource.View) (interface{}, error) {
	return viewHTMLTemplates(t.Name, resources, view)
}

func (t *MessageHTML) ViewResources(resources []resource.ResourceFile, view resource.View) (interface{}, error) {
	return viewHTMLTemplates(t.Name, resources, view)
}

func (t *HTML) UpdateResource(_ context.Context, _ []resource.ResourceFile, resrc *resource.ResourceFile, data []byte) (*resource.ResourceFile, error) {
	return nil, fmt.Errorf("HTML resource cannot be updated. Use MessageHTML resource instead.")
}

func (t *MessageHTML) UpdateResource(ctx context.Context, _ []resource.ResourceFile, resrc *resource.ResourceFile, data []byte) (*resource.ResourceFile, error) {
	if isAllowed, err := isTemplateUpdateAllowed(ctx); !isAllowed || err != nil {
		return nil, err
	}
	return &resource.ResourceFile{
		Location: resrc.Location,
		Data:     data,
	}, nil
}

// PlainText defines a plain text template that is non-customizable
type PlainText struct {
	// Name is the name of template
	Name string
	// ComponentDependencies is the plain text component templates this template depends on.
	ComponentDependencies []*PlainText
}

// MessagePlainText defines a plain text template that is customizable
type MessagePlainText struct {
	// Name is the name of template
	Name string
}

var _ resource.Descriptor = &PlainText{}
var _ resource.Descriptor = &MessagePlainText{}

func (t *PlainText) templateResource()        {}
func (t *MessagePlainText) templateResource() {}

func (t *PlainText) MatchResource(path string) (*resource.Match, bool) {
	return matchTemplatePath(path, t.Name)
}

func (t *MessagePlainText) MatchResource(path string) (*resource.Match, bool) {
	return matchTemplatePath(path, t.Name)
}

func (t *PlainText) FindResources(fs resource.Fs) ([]resource.Location, error) {
	return readTemplates(fs, t.Name)
}

func (t *MessagePlainText) FindResources(fs resource.Fs) ([]resource.Location, error) {
	return readTemplates(fs, t.Name)
}

func (t *PlainText) ViewResources(resources []resource.ResourceFile, view resource.View) (interface{}, error) {
	return viewTextTemplates(t.Name, resources, view)
}

func (t *MessagePlainText) ViewResources(resources []resource.ResourceFile, view resource.View) (interface{}, error) {
	return viewTextTemplates(t.Name, resources, view)
}

func (t *PlainText) UpdateResource(_ context.Context, _ []resource.ResourceFile, resrc *resource.ResourceFile, data []byte) (*resource.ResourceFile, error) {
	return &resource.ResourceFile{
		Location: resrc.Location,
		Data:     data,
	}, nil
}

func (t *MessagePlainText) UpdateResource(ctx context.Context, _ []resource.ResourceFile, resrc *resource.ResourceFile, data []byte) (*resource.ResourceFile, error) {
	if isAllowed, err := isTemplateUpdateAllowed(ctx); !isAllowed || err != nil {
		return nil, err
	}
	return &resource.ResourceFile{
		Location: resrc.Location,
		Data:     data,
	}, nil
}

func RegisterHTML(name string, dependencies ...*HTML) *HTML {
	desc := &HTML{Name: name, ComponentDependencies: dependencies}
	resource.RegisterResource(desc)
	return desc
}

func RegisterMessageHTML(name string) *MessageHTML {
	desc := &MessageHTML{Name: name}
	resource.RegisterResource(desc)
	return desc
}

func RegisterPlainText(name string, dependencies ...*PlainText) *PlainText {
	desc := &PlainText{Name: name, ComponentDependencies: dependencies}
	resource.RegisterResource(desc)
	return desc
}

func RegisterMessagePlainText(name string) *MessagePlainText {
	desc := &MessagePlainText{Name: name}
	resource.RegisterResource(desc)
	return desc
}

func matchTemplatePath(path string, templateName string) (*resource.Match, bool) {
	r := fmt.Sprintf("^templates/([a-zA-Z0-9-]+)/%s$", regexp.QuoteMeta(templateName))
	matches := regexp.MustCompile(r).FindStringSubmatch(path)
	if len(matches) != 2 {
		return nil, false
	}

	languageTag := matches[1]

	isLanguageTagValid := false
	for _, localeKey := range intl.AvailableLanguages {
		if languageTag == localeKey {
			isLanguageTagValid = true
			break
		}
	}
	if !isLanguageTagValid {
		return nil, false
	}

	return &resource.Match{LanguageTag: languageTag}, true
}

func readTemplates(fs resource.Fs, templateName string) ([]resource.Location, error) {
	templatesDir, err := fs.Open("templates")
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	defer templatesDir.Close()

	langTagDirs, err := templatesDir.Readdirnames(0)
	if err != nil {
		return nil, err
	}

	var locations []resource.Location
	for _, langTag := range langTagDirs {
		p := path.Join("templates", langTag, templateName)
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

	return locations, nil
}

type languageTemplate struct {
	languageTag string
	data        []byte
}

func (t languageTemplate) GetLanguageTag() string {
	return t.languageTag
}

var templateLanguageTagRegex = regexp.MustCompile("^templates/([a-zA-Z0-9-_]+)/")

func viewTemplatesAppFile(resources []resource.ResourceFile, view resource.AppFileView) (interface{}, error) {
	// When template is viewed as AppFile,
	// the exact file is returned.
	path := view.AppFilePath()

	var found bool
	var bytes []byte
	for _, resrc := range resources {
		if resrc.Location.Fs.GetFsLevel() == resource.FsLevelApp && resrc.Location.Path == path {
			found = true
			bytes = resrc.Data
		}
	}

	if !found {
		return nil, resource.ErrResourceNotFound
	}

	return bytes, nil
}

func viewTemplatesEffectiveFile(resources []resource.ResourceFile, view resource.EffectiveFileView) (interface{}, error) {
	// When template is viewed as EffectiveFile, the most specific template is returned.
	path := view.EffectiveFilePath()

	// Compute requestedLangTag
	matches := templateLanguageTagRegex.FindStringSubmatch(path)
	if len(matches) < 2 {
		return nil, resource.ErrResourceNotFound
	}
	requestedLangTag := matches[1]

	var found bool
	var bytes []byte
	for _, resrc := range resources {
		langTag := templateLanguageTagRegex.FindStringSubmatch(resrc.Location.Path)[1]
		if langTag == requestedLangTag {
			found = true
			bytes = resrc.Data
		}
	}

	if !found {
		return nil, resource.ErrResourceNotFound
	}

	return bytes, nil
}

func viewTemplatesEffectiveResource(resources []resource.ResourceFile, view resource.EffectiveResourceView) (*languageTemplate, error) {
	preferredLanguageTags := view.PreferredLanguageTags()
	defaultLanguageTag := view.DefaultLanguageTag()

	languageTemplates := make(map[string]languageTemplate)
	add := func(langTag string, resrc resource.ResourceFile) error {
		t := languageTemplate{
			languageTag: langTag,
			data:        resrc.Data,
		}
		languageTemplates[langTag] = t
		return nil
	}
	extractLanguageTag := func(resrc resource.ResourceFile) string {
		langTag := templateLanguageTagRegex.FindStringSubmatch(resrc.Location.Path)[1]
		return langTag
	}

	err := intlresource.Prepare(resources, view, extractLanguageTag, add)
	if err != nil {
		return nil, err
	}

	var items []intlresource.LanguageItem
	for _, i := range languageTemplates {
		items = append(items, i)
	}

	matched, err := intlresource.Match(preferredLanguageTags, defaultLanguageTag, items)
	if errors.Is(err, intlresource.ErrNoLanguageMatch) {
		if len(items) > 0 {
			// Use first item in case of no match, to ensure resolution always succeed
			matched = items[0]
		} else {
			// If no configured translation for a template, fail the resolution process
			return nil, ErrNotFound
		}
	} else if err != nil {
		return nil, err
	}

	tagger := matched.(languageTemplate)
	return &tagger, nil
}

func viewHTMLTemplates(name string, resources []resource.ResourceFile, rawView resource.View) (interface{}, error) {

	switch view := rawView.(type) {
	case resource.AppFileView:
		bytes, err := viewTemplatesAppFile(resources, view)
		if err != nil {
			return nil, err
		}
		return bytes, nil
	case resource.EffectiveFileView:
		bytes, err := viewTemplatesEffectiveFile(resources, view)
		if err != nil {
			return nil, err
		}
		return bytes, nil
	case resource.EffectiveResourceView:
		templatesEffectiveResource, err := viewTemplatesEffectiveResource(resources, view)
		if err != nil {
			return nil, err
		}
		tpl := htmltemplate.New(name)
		funcMap := MakeTemplateFuncMap(tpl)
		tpl.Funcs(funcMap)
		_, err = tpl.Parse(string(templatesEffectiveResource.data))
		if err != nil {
			return nil, fmt.Errorf("invalid HTML template: %w", err)
		}
		return &HTMLTemplateEffectiveResource{
			Data:        templatesEffectiveResource.data,
			LanguageTag: templatesEffectiveResource.languageTag,
			Template:    tpl,
		}, nil
	case resource.ValidateResourceView:
		for _, resrc := range resources {
			tpl := htmltemplate.New(name)
			funcMap := MakeTemplateFuncMap(tpl)
			tpl.Funcs(funcMap)
			template, err := tpl.Parse(string(resrc.Data))
			if err != nil {
				return nil, fmt.Errorf("invalid HTML template: %w", err)
			}
			err = templateValidator.ValidateHTMLTemplate(template)
			if err != nil {
				return nil, fmt.Errorf("invalid HTML template: %w", err)
			}
		}
		return nil, nil
	default:
		return nil, fmt.Errorf("unsupported view: %T", view)
	}

}

func viewTextTemplates(name string, resources []resource.ResourceFile, rawView resource.View) (interface{}, error) {

	switch view := rawView.(type) {
	case resource.AppFileView:
		bytes, err := viewTemplatesAppFile(resources, view)
		if err != nil {
			return nil, err
		}
		return bytes, nil
	case resource.EffectiveFileView:
		bytes, err := viewTemplatesEffectiveFile(resources, view)
		if err != nil {
			return nil, err
		}
		return bytes, nil
	case resource.EffectiveResourceView:
		templatesEffectiveResource, err := viewTemplatesEffectiveResource(resources, view)
		if err != nil {
			return nil, err
		}
		tpl := texttemplate.New(name)
		funcMap := MakeTemplateFuncMap(tpl)
		tpl.Funcs(funcMap)
		_, err = tpl.Parse(string(templatesEffectiveResource.data))
		if err != nil {
			return nil, fmt.Errorf("invalid HTML template: %w", err)
		}
		return &TextTemplateEffectiveResource{
			Data:        templatesEffectiveResource.data,
			LanguageTag: templatesEffectiveResource.languageTag,
			Template:    tpl,
		}, nil
	case resource.ValidateResourceView:
		for _, resrc := range resources {
			tpl := texttemplate.New(name)
			funcMap := MakeTemplateFuncMap(tpl)
			tpl.Funcs(funcMap)
			template, err := tpl.Parse(string(resrc.Data))
			if err != nil {
				return nil, fmt.Errorf("invalid text template: %w", err)
			}
			err = templateValidator.ValidateTextTemplate(template)
			if err != nil {
				return nil, fmt.Errorf("invalid text template: %w", err)
			}
		}
		return nil, nil
	default:
		return nil, fmt.Errorf("unsupported view: %T", view)
	}

}
