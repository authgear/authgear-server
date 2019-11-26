package welcemail

import (
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/template"
)

const (
	TemplateItemTypeWelcomeEmailTXT  config.TemplateItemType = "welcome_email.txt"
	TemplateItemTypeWelcomeEmailHTML config.TemplateItemType = "welcome_email.html"
)

var TemplateWelcomeEmailTXT = template.Spec{
	Type: TemplateItemTypeWelcomeEmailTXT,
	Default: `Hello {{ .email }},

Welcome to Skygear.

Thanks.`,
}

var TemplateWelcomeEmailHTML = template.Spec{
	Type:   TemplateItemTypeWelcomeEmailHTML,
	IsHTML: true,
}
