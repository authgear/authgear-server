package template

import (
	"github.com/authgear/authgear-server/pkg/auth/config"
)

// Spec is an template spec.
type Spec struct {
	// Type is the type of the template.
	Type config.TemplateItemType `json:"type"`
	// IsHTML indicates whether the template content is HTML
	IsHTML bool `json:"is_html"`
	// Defines is a list of defines to be parsed after the main template is parsed.
	Defines []string `json:"-"`
	// Translation expresses that this template depends on another template to provide translation.
	Translation config.TemplateItemType `json:"-"`
	// Components is a list of components that this template depends on.
	// Defines and Translation of components are ignored.
	Components []config.TemplateItemType `json:"-"`
}
