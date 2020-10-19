package template

import (
	"errors"
	"fmt"
	htmltemplate "html/template"
	"os"
	"path"
	"regexp"
	texttemplate "text/template"

	"github.com/authgear/authgear-server/pkg/util/resource"
)

const ResourceArgPreferredLanguageTag = "preferred_language_tag"
const ResourceArgDefaultLanguageTag = "default_language_tag"
const LanguageTagDefault = "__default__"

type Resource interface {
	templateResource()
}

// HTML defines a HTML template
type HTML struct {
	// Name is the name of template
	Name string
	// ComponentDependencies is the HTML component templates this template depends on.
	ComponentDependencies []*HTML
}

func (t *HTML) templateResource() {}

func (t *HTML) ReadResource(fs resource.Fs) ([]resource.LayerFile, error) {
	return readTemplates(fs, t.Name)
}

func (t *HTML) MatchResource(path string) bool {
	return matchTemplatePath(path, t.Name)
}

func (t *HTML) Merge(layers []resource.LayerFile, args map[string]interface{}) (*resource.MergedFile, error) {
	return mergeTemplates(layers, args)
}

func (t *HTML) Parse(merged *resource.MergedFile) (interface{}, error) {
	tpl := htmltemplate.New(t.Name)
	tpl.Funcs(templateFuncMap)
	_, err := tpl.Parse(string(merged.Data))
	if err != nil {
		return nil, fmt.Errorf("invalid HTML template: %w", err)
	}
	return tpl, nil
}

// PlainText defines a plain text template
type PlainText struct {
	// Name is the name of template
	Name string
	// ComponentDependencies is the plain text component templates this template depends on.
	ComponentDependencies []*PlainText
}

func (t *PlainText) templateResource() {}

func (t *PlainText) ReadResource(fs resource.Fs) ([]resource.LayerFile, error) {
	return readTemplates(fs, t.Name)
}

func (t *PlainText) MatchResource(path string) bool {
	return matchTemplatePath(path, t.Name)
}

func (t *PlainText) Merge(layers []resource.LayerFile, args map[string]interface{}) (*resource.MergedFile, error) {
	return mergeTemplates(layers, args)
}

func (t *PlainText) Parse(merged *resource.MergedFile) (interface{}, error) {
	tpl := texttemplate.New(t.Name)
	tpl.Funcs(templateFuncMap)
	_, err := tpl.Parse(string(merged.Data))
	if err != nil {
		return nil, fmt.Errorf("invalid text template: %w", err)
	}
	return tpl, nil
}

func RegisterHTML(name string, dependencies ...*HTML) *HTML {
	desc := &HTML{Name: name, ComponentDependencies: dependencies}
	resource.RegisterResource(desc)
	return desc
}

func RegisterPlainText(name string, dependencies ...*PlainText) *PlainText {
	desc := &PlainText{Name: name, ComponentDependencies: dependencies}
	resource.RegisterResource(desc)
	return desc
}

func matchTemplatePath(path string, templateName string) bool {
	r := fmt.Sprintf("^templates/([a-zA-Z0-9-]+|%s)/%s$", LanguageTagDefault, regexp.QuoteMeta(templateName))
	return regexp.MustCompile(r).MatchString(path)
}

func readTemplates(fs resource.Fs, templateName string) ([]resource.LayerFile, error) {
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

	var files []resource.LayerFile
	for _, langTag := range langTagDirs {
		p := path.Join("templates", langTag, templateName)
		data, err := resource.ReadFile(fs, p)
		if os.IsNotExist(err) {
			continue
		} else if err != nil {
			return nil, err
		}
		files = append(files, resource.LayerFile{
			Path: p,
			Data: data,
		})
	}

	return files, nil
}

type languageTemplate struct {
	languageTag string
	data        []byte
}

func (t languageTemplate) GetLanguageTag() string {
	return t.languageTag
}

var templateLanguageTagRegex = regexp.MustCompile("^templates/([a-zA-Z0-9-_]+)/")

func mergeTemplates(layers []resource.LayerFile, args map[string]interface{}) (*resource.MergedFile, error) {
	preferredLanguageTags, _ := args[ResourceArgPreferredLanguageTag].([]string)
	defaultLanguageTag, _ := args[ResourceArgDefaultLanguageTag].(string)

	languageTemplates := make(map[string]languageTemplate)
	for _, file := range layers {
		langTag := templateLanguageTagRegex.FindStringSubmatch(file.Path)[1]
		t := languageTemplate{
			languageTag: langTag,
			data:        file.Data,
		}
		if t.languageTag == LanguageTagDefault {
			t.languageTag = defaultLanguageTag
		}
		languageTemplates[langTag] = t
	}

	if _, ok := languageTemplates[defaultLanguageTag]; !ok {
		languageTemplates[defaultLanguageTag] = languageTemplates[LanguageTagDefault]
	}
	delete(languageTemplates, LanguageTagDefault)

	var items []LanguageItem
	for _, i := range languageTemplates {
		items = append(items, i)
	}

	matched, err := MatchLanguage(preferredLanguageTags, defaultLanguageTag, items)
	if errors.Is(err, ErrNoLanguageMatch) {
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
	return &resource.MergedFile{Data: tagger.data}, nil
}
