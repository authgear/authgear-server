package welcomemessage

import (
	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/template"
)

const (
	TemplateItemTypeWelcomeEmailTXT  config.TemplateItemType = "welcome_email.txt"
	TemplateItemTypeWelcomeEmailHTML config.TemplateItemType = "welcome_email.html"
)

var TemplateWelcomeEmailTXT = template.Spec{
	Type: TemplateItemTypeWelcomeEmailTXT,
}

var TemplateWelcomeEmailHTML = template.Spec{
	Type:   TemplateItemTypeWelcomeEmailHTML,
	IsHTML: true,
}
