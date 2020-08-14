package template

// T defines a template.
type T struct {
	// Type is the type of the template.
	Type string
	// IsHTML indicates which template package to use.
	// If it is true, html/template is used.
	// Otherwise, text/template is used.
	IsHTML bool
	// Defines is a list of additional templates to be parsed after the main template is parsed.
	Defines []string
	// TranslationTemplateType is the type of the template that provides translation.
	TranslationTemplateType string
	// ComponentTemplateTypes is the type of the template this template depends on.
	ComponentTemplateTypes []string
}

// Reference tells us how to resolve a template.
type Reference struct {
	// Type is the type of the template.
	Type string
	// LanguageTag indicates the language of the template content.
	LanguageTag string
	// URI indicates the location of template content.
	URI string
}

type Translation struct {
	LanguageTag string
	Value       string
}

func (t Translation) GetLanguageTag() string {
	return t.LanguageTag
}

type Resolved struct {
	T                 T
	Content           string
	Translations      map[string]Translation
	ComponentContents []string
}
