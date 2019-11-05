package template

import (
	"github.com/skygeario/skygear-server/pkg/core/config"
)

// T is an template.
type T struct {
	// Type is the type of the template.
	Type config.TemplateItemType
	// Default is the default content of the template.
	Default string
}
