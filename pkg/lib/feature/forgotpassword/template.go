package forgotpassword

import (
	"github.com/authgear/authgear-server/pkg/util/template"
)

const (
	TemplateItemTypeForgotPasswordEmailTXT  string = "forgot_password_email.txt"
	TemplateItemTypeForgotPasswordEmailHTML string = "forgot_password_email.html"
	TemplateItemTypeForgotPasswordSMSTXT    string = "forgot_password_sms.txt"
)

var TemplateForgotPasswordEmailTXT = template.T{
	Type: TemplateItemTypeForgotPasswordEmailTXT,
}

var TemplateForgotPasswordEmailHTML = template.T{
	Type:   TemplateItemTypeForgotPasswordEmailHTML,
	IsHTML: true,
}

var TemplateForgotPasswordSMSTXT = template.T{
	Type: TemplateItemTypeForgotPasswordSMSTXT,
}
