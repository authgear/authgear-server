package forgotpassword

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/template"
)

const (
	TemplateItemTypeForgotPasswordEmailTXT  config.TemplateItemType = "forgot_password_email.txt"
	TemplateItemTypeForgotPasswordEmailHTML config.TemplateItemType = "forgot_password_email.html"
	TemplateItemTypeForgotPasswordSMSTXT    config.TemplateItemType = "forgot_password_sms.txt"
)

var TemplateForgotPasswordEmailTXT = template.Spec{
	Type: TemplateItemTypeForgotPasswordEmailTXT,
}

var TemplateForgotPasswordEmailHTML = template.Spec{
	Type:   TemplateItemTypeForgotPasswordEmailHTML,
	IsHTML: true,
}

var TemplateForgotPasswordSMSTXT = template.Spec{
	Type: TemplateItemTypeForgotPasswordSMSTXT,
}
