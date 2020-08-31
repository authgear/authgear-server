package welcomemessage

import (
	"github.com/authgear/authgear-server/pkg/util/template"
)

const (
	TemplateItemTypeWelcomeEmailTXT  string = "welcome_email.txt"
	TemplateItemTypeWelcomeEmailHTML string = "welcome_email.html"
)

var TemplateWelcomeEmailTXT = template.Register(template.T{
	Type: TemplateItemTypeWelcomeEmailTXT,
})

var TemplateWelcomeEmailHTML = template.Register(template.T{
	Type:   TemplateItemTypeWelcomeEmailHTML,
	IsHTML: true,
})
