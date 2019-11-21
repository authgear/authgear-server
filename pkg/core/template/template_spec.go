package template

import (
	"github.com/skygeario/skygear-server/pkg/core/config"
)

// Spec is an template spec.
type Spec struct {
	// Type is the type of the template.
	Type config.TemplateItemType
	// Default is the default content of the template.
	Default string
	// IsKeyed indicates whether the template content may vary according to some key
	IsKeyed bool
	// IsHTML indicates whether the template content is HTML
	IsHTML bool
}
