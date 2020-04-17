package forgotpassword

import (
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/template"
)

const (
	TemplateItemTypeForgotPasswordEmailTXT  config.TemplateItemType = "forgot_password_email.txt"
	TemplateItemTypeForgotPasswordEmailHTML config.TemplateItemType = "forgot_password_email.html"
	TemplateItemTypeForgotPasswordSMSTXT    config.TemplateItemType = "forgot_password_sms.txt"
)

var TemplateForgotPasswordEmailTXT = template.Spec{
	Type: TemplateItemTypeForgotPasswordEmailTXT,
	Default: `Dear {{ .email }},

You received this email because someone tries to reset your account password on {{ .appname }}. To reset your account password, click this link:

{{ .link }}

If you did not request to reset your account password, Please ignore this email.

Thanks.`,
}

var TemplateForgotPasswordEmailHTML = template.Spec{
	Type:   TemplateItemTypeForgotPasswordEmailHTML,
	IsHTML: true,
	Default: `<!DOCTYPE html>
<html>
<body>
<p>Dear {{ .email }},</p>
<p>You received this email because someone tries to reset your account password on {{ .appname }}. To reset your account password, click this link:</p>
<p><a href="{{ .link }}">{{ .link }}</a></p>
<p>If you did not request to reset your account password, Please ignore this email.</p>
<p>Thanks.</p>
</body>
</html>
`,
}

var TemplateForgotPasswordSMSTXT = template.Spec{
	Type: TemplateItemTypeForgotPasswordSMSTXT,
	Default: `Visit this link to reset your password on {{ .appname }}
{{ .link }}
`,
}
